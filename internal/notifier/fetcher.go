package notifier

import (
	"context"
	"log"
	"strings"
	"time"
)

type Forecast struct {
	ZoneID   string
	IssuedAt time.Time
}

type ForecastFetcher func(ctx context.Context, centerID string) ([]Forecast, error)

func FetchLatestStub(ctx context.Context) ([]Forecast, error) {
	fixed := time.Date(2025, 11, 8, 19, 0, 0, 0, time.UTC)
	return []Forecast{{ZoneID: "NWAC_164", IssuedAt: fixed}}, nil
}

// MakeFetchFromSubscriptions returns a fetcher that inspects current subscriptions,
// fetches forecasts per center, and returns the latest publish time per subscribed zone.
// It expects zone IDs in the form "CENTER_ZONEID" (e.g., "NWAC_164") so it can infer the center.
func MakeFetchFromSubscriptions(repo Repository, api interface {
	FetchForecasts(centerID string) ([]ForecastSource, error)
}) ForecastFetcher {
	return func(ctx context.Context, centerID string) ([]Forecast, error) {
		// Get all zones with subscriptions
		zones, err := repo.ListSubscribedZones(ctx)
		if err != nil {
			return nil, err
		}
		if len(zones) == 0 {
			return nil, nil
		}

		// Filter zones to only those for the requested center
		zoneSet := make(map[string]struct{})
		for _, z := range zones {
			parts := strings.SplitN(z, "_", 2)
			if len(parts) > 0 && parts[0] == centerID {
				zoneSet[z] = struct{}{}
			}
		}
		if len(zoneSet) == 0 {
			return nil, nil
		}

		sources, err := api.FetchForecasts(centerID)
		if err != nil {
			log.Printf("fetch center %s failed: %v", centerID, err)
			return nil, err
		}

		// Track max PublishedAt per zone
		latest := map[string]time.Time{}
		for _, src := range sources {
			for _, z := range src.ZoneIDs() {
				if _, want := zoneSet[z]; !want {
					continue
				}
				t := src.PublishedAt()
				if t.After(latest[z]) {
					latest[z] = t
				}
			}
		}

		var out []Forecast
		for z, t := range latest {
			out = append(out, Forecast{ZoneID: z, IssuedAt: t})
		}
		return out, nil
	}
}

// ForecastSource is a small adapter interface so we can consume forecasts
// from the clients package without importing it directly here.
type ForecastSource interface {
	ZoneIDs() []string
	PublishedAt() time.Time
}
