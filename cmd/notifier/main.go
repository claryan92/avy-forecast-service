package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"example.com/avalanche/internal/clients"
	"example.com/avalanche/internal/models"
	"example.com/avalanche/internal/notifier"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const defaultAvyAPIBaseURL = "https://api.avalanche.org/v2/public"

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	emailClient := notifier.NewSendGridEmailClient()

	pollIntervalStr := os.Getenv("NOTIFIER_POLL_INTERVAL")
	if pollIntervalStr == "" {
		pollIntervalStr = "12h"
	}
	pollInterval, err := time.ParseDuration(pollIntervalStr)
	if err != nil {
		log.Fatalf("invalid NOTIFIER_POLL_INTERVAL: %v", err)
	}

	repo := notifier.NewGormRepository(db)

	baseURL := os.Getenv("AVY_API_BASE_URL")
	if baseURL == "" {
		baseURL = defaultAvyAPIBaseURL
	}
	httpClient := &clients.HTTPClient{Client: &http.Client{Timeout: 10 * time.Second}}
	apiClient := clients.NewAvalancheAPIClient(baseURL, httpClient)

	fetcher := notifier.MakeFetchFromSubscriptions(repo, forecastSourceAdapter{apiClient})
	service := notifier.NewService(repo, emailClient, pollInterval, fetcher)

	ctx := context.Background()
	log.Printf("email notifier started; polling every %s", pollInterval)
	if err := service.Run(ctx); err != nil {
		log.Fatalf("notifier stopped: %v", err)
	}
}

type forecastSourceAdapter struct{ c *clients.AvalancheAPIClient }

func (a forecastSourceAdapter) FetchForecasts(centerID string) ([]notifier.ForecastSource, error) {
	forecasts, err := a.c.FetchForecasts(centerID)
	if err != nil {
		return nil, err
	}
	out := make([]notifier.ForecastSource, 0, len(forecasts))
	for i := range forecasts {
		f := forecasts[i]
		out = append(out, forecastSource{F: &f})
	}
	return out, nil
}

type forecastSource struct{ F *models.Forecast }

func (fs forecastSource) ZoneIDs() []string {
	ids := make([]string, 0, len(fs.F.ForecastZone))
	for _, z := range fs.F.ForecastZone {
		ids = append(ids, z.ZoneID)
	}
	return ids
}

func (fs forecastSource) PublishedAt() time.Time { return fs.F.PublishedTime }
