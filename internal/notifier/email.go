package notifier

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"
	"time"

	"example.com/avalanche/internal/models"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailData struct {
	ZoneID     string
	ZoneName   string
	IssuedAt   time.Time
	Today      *models.DangerRating
	Tomorrow   *models.DangerRating
	CenterLink string
}

type EmailSender interface {
	SendForecastEmail(ctx context.Context, recipient string, data EmailData) error
	SendCenterForecastEmail(ctx context.Context, recipient string, centerName string, centerLink string, zones []ZoneSummary) error
}

// ZoneSummary represents a simplified view of a zone for aggregated emails.
type ZoneSummary struct {
	ZoneID      string
	ZoneName    string
	TodayStr    string
	TomorrowStr string
}

type SendGridEmailClient struct {
	sg   *sendgrid.Client
	from string
	tmpl *template.Template
}

func NewSendGridEmailClient() *SendGridEmailClient {
	key := os.Getenv("SENDGRID_API_KEY")
	if key == "" {
		panic("SENDGRID_API_KEY is required")
	}
	from := os.Getenv("EMAIL_FROM")
	if from == "" {
		from = "alerts@example.com"
	}

	tmpl, err := template.ParseFiles("internal/email/templates/forecast.html")
	if err != nil {
		tmpl = template.Must(template.New("forecast").Parse(defaultTemplate))
	}

	return &SendGridEmailClient{
		sg:   sendgrid.NewSendClient(key),
		from: from,
		tmpl: tmpl,
	}
}

func (c *SendGridEmailClient) SendForecastEmail(ctx context.Context, recipient string, data EmailData) error {
	var buf bytes.Buffer
	_ = c.tmpl.Execute(&buf, map[string]any{
		"ZoneID":     data.ZoneID,
		"ZoneName":   data.ZoneName,
		"IssuedAt":   data.IssuedAt.Format("Mon Jan 2 15:04 2006 MST"),
		"Today":      data.Today,
		"Tomorrow":   data.Tomorrow,
		"CenterLink": data.CenterLink,
	})

	label := data.ZoneID
	if strings.TrimSpace(data.ZoneName) != "" {
		label = data.ZoneName
	}
	subject := fmt.Sprintf("New Avalanche Forecast for %s", label)
	from := mail.NewEmail("Avy Notifier", c.from)
	to := mail.NewEmail("Subscriber", recipient)
	m := mail.NewSingleEmail(from, subject, to, "A new avalanche forecast is available.", buf.String())

	resp, err := c.sg.Send(m)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("sendgrid send failed: %d %s", resp.StatusCode, resp.Body)
	}
	log.Printf("sent forecast email to %s for zone %s", recipient, data.ZoneID)
	return nil
}

const defaultTemplate = `<div style="font-family:Arial,sans-serif;">
	<h2>New Avalanche Forecast</h2>
	<p>A new forecast has been issued for zone <b>{{if .ZoneName}}{{.ZoneName}} ({{.ZoneID}}){{else}}{{.ZoneID}}{{end}}</b> at <b>{{.IssuedAt}}</b>.</p>
	<p>Check the latest details on <a href="{{.CenterLink}}">Visit Center Website</a>.</p>
</div>`

// BuildZoneSummaries derives zone summaries from processed zone forecasts.
func BuildZoneSummaries(forecasts []models.ZoneForecast) []ZoneSummary {
	out := make([]ZoneSummary, 0, len(forecasts))
	for _, f := range forecasts {
		todayStr := formatDanger(f.TodayDanger)
		tomorrowStr := formatDanger(f.FutureDanger)
		out = append(out, ZoneSummary{
			ZoneID:      f.ZoneID,
			ZoneName:    f.ZoneName,
			TodayStr:    todayStr,
			TomorrowStr: tomorrowStr,
		})
	}
	return out
}

// formatDanger converts a DangerRating to a display string.
func formatDanger(d *models.DangerRating) string {
	if d == nil {
		return "No rating"
	}
	parts := []string{}
	if d.Upper != 0 {
		parts = append(parts, fmt.Sprintf("U:%d", d.Upper))
	}
	if d.Middle != 0 {
		parts = append(parts, fmt.Sprintf("M:%d", d.Middle))
	}
	if d.Lower != 0 {
		parts = append(parts, fmt.Sprintf("L:%d", d.Lower))
	}
	if len(parts) == 0 && d.Message != "" {
		return d.Message
	}
	if len(parts) == 0 {
		return "No rating"
	}
	return strings.Join(parts, "/")
}

// SendCenterForecastEmail sends a single aggregated email listing all zones for a center using the HTML template.
func (c *SendGridEmailClient) SendCenterForecastEmail(ctx context.Context, recipient string, centerName string, centerLink string, zones []ZoneSummary) error {
	if len(zones) == 0 {
		return nil
	}

	// Load and execute the center forecast template
	tmpl, err := template.ParseFiles("internal/email/templates/center_forecast.html")
	if err != nil {
		// Fallback to inline HTML if template not found
		log.Printf("center template not found, using fallback: %v", err)
		return c.sendCenterForecastFallback(ctx, recipient, centerName, centerLink, zones)
	}

	var buf bytes.Buffer
	data := map[string]any{
		"CenterName": centerName,
		"CenterLink": centerLink,
		"ZoneCount":  len(zones),
		"Zones":      zones,
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("template execute failed: %w", err)
	}

	from := mail.NewEmail("Avy Notifier", c.from)
	to := mail.NewEmail("Subscriber", recipient)
	subject := fmt.Sprintf("%s Avalanche Center Forecast Summary", centerName)
	msg := mail.NewSingleEmail(from, subject, to, "Your avalanche center forecast summary is available.", buf.String())
	resp, err := c.sg.Send(msg)
	if err != nil {
		return fmt.Errorf("sendgrid send failed: %w", err)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("sendgrid send failed: %d %s", resp.StatusCode, resp.Body)
	}
	log.Printf("sent aggregated center forecast email to %s for center %s", recipient, centerName)
	return nil
}

// sendCenterForecastFallback sends a simple HTML email if the template is unavailable.
func (c *SendGridEmailClient) sendCenterForecastFallback(ctx context.Context, recipient string, centerName string, centerLink string, zones []ZoneSummary) error {
	var b strings.Builder
	fmt.Fprintf(&b, "<h2>Latest Avalanche Forecasts - %s</h2>", centerName)
	fmt.Fprintf(&b, "<p>%d zones have current forecasts.</p><ul>", len(zones))
	for _, z := range zones {
		fmt.Fprintf(&b, "<li><strong>%s</strong> (%s) - Today: %s | Tomorrow: %s</li>", z.ZoneName, z.ZoneID, z.TodayStr, z.TomorrowStr)
	}
	fmt.Fprintf(&b, "</ul><p>For details visit <a href=\"%s\">%s</a>.</p>", centerLink, centerName)

	from := mail.NewEmail("Avy Notifier", c.from)
	to := mail.NewEmail("Subscriber", recipient)
	subject := fmt.Sprintf("%s Avalanche Center Forecast Summary", centerName)
	msg := mail.NewSingleEmail(from, subject, to, "Your avalanche center forecast summary is available.", b.String())
	resp, err := c.sg.Send(msg)
	if err != nil {
		return fmt.Errorf("sendgrid send failed: %w", err)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("sendgrid send failed: %d %s", resp.StatusCode, resp.Body)
	}
	log.Printf("sent aggregated center forecast email (fallback) to %s for center %s", recipient, centerName)
	return nil
}
