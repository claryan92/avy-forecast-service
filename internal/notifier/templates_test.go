package notifier_test

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestForecastTemplate_RendersCenterLinkAndZone(t *testing.T) {
	path := filepath.Join("internal", "email", "templates", "forecast.html")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Skip("template not available in test environment:", err)
	}
	tmpl, err := template.New("forecast").Parse(string(b))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	var out strings.Builder
	data := map[string]any{
		"ZoneID":     "NWAC_10",
		"ZoneName":   "Mt Hood",
		"IssuedAt":   "Mon Jan 2 15:04 2006 MST",
		"CenterLink": "https://nwac.us/",
	}
	if err := tmpl.Execute(&out, data); err != nil {
		t.Fatalf("exec failed: %v", err)
	}
	html := out.String()
	if !strings.Contains(html, "Mt Hood") || !strings.Contains(html, "NWAC_10") {
		t.Fatalf("expected zone name and id in output: %s", html)
	}
	if !strings.Contains(html, "https://nwac.us/") {
		t.Fatalf("expected center link in output: %s", html)
	}
}

func TestCenterForecastTemplate_RendersLinkAndZones(t *testing.T) {
	path := filepath.Join("internal", "email", "templates", "center_forecast.html")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Skip("template not available in test environment:", err)
	}
	tmpl, err := template.New("center").Parse(string(b))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	var out strings.Builder
	data := map[string]any{
		"CenterName": "Northwest Avalanche Center",
		"CenterLink": "https://nwac.us/",
		"ZoneCount":  2,
		"Zones": []map[string]string{
			{"ZoneName": "Mt Hood", "ZoneID": "NWAC_10", "TodayStr": "Low", "TomorrowStr": "Moderate"},
			{"ZoneName": "Stevens Pass", "ZoneID": "NWAC_2", "TodayStr": "Low", "TomorrowStr": "Low"},
		},
	}
	if err := tmpl.Execute(&out, data); err != nil {
		t.Fatalf("exec failed: %v", err)
	}
	html := out.String()
	if !strings.Contains(html, "Visit Northwest Avalanche Center Website") {
		t.Fatalf("missing link label: %s", html)
	}
	if !strings.Contains(html, "https://nwac.us/") {
		t.Fatalf("missing center link: %s", html)
	}
	if !strings.Contains(html, "Mt Hood") || !strings.Contains(html, "Stevens Pass") {
		t.Fatalf("missing zone names: %s", html)
	}
}
