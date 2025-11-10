package services

import (
	"sort"
	"time"

	"example.com/avalanche/internal/models"
)

// ForecastClient defines the behavior required to fetch avalanche forecasts
// for a given avalanche center. Implementations of this interface can use
// different data sources (e.g., HTTP APIs, caches, or databases).
type ForecastClient interface {
	FetchForecasts(centerID string) ([]models.Forecast, error)
}

// ForecastService provides methods to fetch, process, and organize avalanche
// forecasts across multiple avalanche centers.
type ForecastService struct {
	client ForecastClient
}

// NewForecast returns a new ForecastService configured with the given ForecastClient.
// The client is used to retrieve raw forecast data from one or more avalanche centers.
func NewForecast(client ForecastClient) *ForecastService {
	return &ForecastService{client: client}
}

// GetForecastsForCenters retrieves forecasts from all provided avalanche center IDs,
// filters them by the specified target date, and returns a processed list of
// zone-level forecasts. Results are sorted by avalanche center and zone name.
func (s *ForecastService) GetForecastsForCenters(centerIDs []string, targetDate time.Time) ([]models.ZoneForecast, error) {
	var allForecasts []models.Forecast

	for _, id := range centerIDs {
		forecasts, err := s.client.FetchForecasts(id)
		if err != nil {
			return nil, err
		}
		for i := range forecasts {
			if forecasts[i].AvalancheCenter.ID == "" {
				forecasts[i].AvalancheCenter.ID = id
			}
		}
		allForecasts = append(allForecasts, forecasts...)
	}

	filtered := s.filterForecastsForDate(allForecasts, targetDate)
	result := s.ProcessForecasts(filtered)
	return result, nil
}

// ProcessForecasts transforms a list of raw forecasts into zone forecasts.
// It sorts input forecasts by publish time, groups them by zone, fills in
// missing danger ratings, and sorts the final output by avalanche center and zone.
func (s *ForecastService) ProcessForecasts(forecasts []models.Forecast) []models.ZoneForecast {
	SortForecastsByPublishTime(forecasts)
	zoneMap := s.BuildZoneForecasts(forecasts)
	result := FlattenZoneMap(zoneMap)
	s.FillMissingDangerRatings(result)
	SortZoneForecasts(result)
	return result
}

// BuildZoneForecasts constructs a map of zone forecasts keyed by zone ID.
// Each zone forecast includes metadata, issued times, and available danger ratings.
// Only the most recent forecast for each zone is retained.
// Zone IDs are prefixed with the center ID (e.g., "10" becomes "NWAC_10").
func (s *ForecastService) BuildZoneForecasts(forecasts []models.Forecast) map[string]*models.ZoneForecast {
	zoneMap := make(map[string]*models.ZoneForecast)

	for _, f := range forecasts {
		centerID := f.AvalancheCenter.ID

		for _, z := range f.ForecastZone {
			fullZoneID := centerID + "_" + z.ZoneID

			if _, exists := zoneMap[fullZoneID]; exists {
				continue
			}

			zf := &models.ZoneForecast{
				ZoneID:     fullZoneID,
				ZoneName:   z.Name,
				Center:     f.AvalancheCenter.Name,
				IssuedTime: f.PublishedTime.Format(time.RFC3339),
				StartDate:  f.StartDate.Format(time.RFC3339),
				EndDate:    f.EndDate.Format(time.RFC3339),
				BottomLine: bottomLineOrDefault(f),
			}

			for _, d := range f.Danger {
				switch d.ValidDay {
				case "current":
					zf.TodayDanger = &d
				case "tomorrow":
					zf.FutureDanger = &d
				}
			}

			zoneMap[fullZoneID] = zf
		}
	}

	return zoneMap
}

// FillMissingDangerRatings ensures each zone forecast has both today and tomorrow
// danger ratings. If either is missing, a placeholder message is added.
func (s *ForecastService) FillMissingDangerRatings(zoneForecasts []models.ZoneForecast) {
	for i := range zoneForecasts {
		zf := &zoneForecasts[i]
		if zf.TodayDanger == nil {
			zf.TodayDanger = &models.DangerRating{
				ValidDay: "current",
				Message:  "No forecast available today",
			}
		}
		if zf.FutureDanger == nil {
			zf.FutureDanger = &models.DangerRating{
				ValidDay: "tomorrow",
				Message:  "No forecast available tomorrow",
			}
		}
	}
}

// filterForecastsForDate filters forecasts to those that are valid for the given target date.
// Forecasts must have a "published" status and overlap the target date range.
func (s *ForecastService) filterForecastsForDate(all []models.Forecast, targetDate time.Time) []models.Forecast {
	targetDate = time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, time.UTC)

	filtered := make([]models.Forecast, 0)
	for _, f := range all {
		start := time.Date(f.StartDate.Year(), f.StartDate.Month(), f.StartDate.Day(), 0, 0, 0, 0, time.UTC)
		end := time.Date(f.EndDate.Year(), f.EndDate.Month(), f.EndDate.Day(), 0, 0, 0, 0, time.UTC)

		if f.Status == "published" && !targetDate.Before(start) && !targetDate.After(end) {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// SortForecastsByPublishTime sorts forecasts in descending order of publish time,
// ensuring that the newest forecasts appear first.
func SortForecastsByPublishTime(forecasts []models.Forecast) {
	sort.Slice(forecasts, func(i, j int) bool {
		return forecasts[i].PublishedTime.After(forecasts[j].PublishedTime)
	})
}

// SortZoneForecasts sorts zone forecasts first by avalanche center name,
// then alphabetically by zone name.
func SortZoneForecasts(zoneForecasts []models.ZoneForecast) {
	sort.SliceStable(zoneForecasts, func(i, j int) bool {
		if zoneForecasts[i].Center == zoneForecasts[j].Center {
			return zoneForecasts[i].ZoneName < zoneForecasts[j].ZoneName
		}
		return zoneForecasts[i].Center < zoneForecasts[j].Center
	})
}

// FlattenZoneMap converts a map of zone forecasts into a slice.
// The returned slice preserves no particular order.
func FlattenZoneMap(zoneMap map[string]*models.ZoneForecast) []models.ZoneForecast {
	result := make([]models.ZoneForecast, 0, len(zoneMap))
	for _, zf := range zoneMap {
		result = append(result, *zf)
	}
	return result
}

// bottomLineOrDefault returns the bottom line summary from a forecast,
// or a fallback message if none is available.
func bottomLineOrDefault(f models.Forecast) string {
	if f.BottomLine != "" {
		return f.BottomLine
	}
	return "No bottom line available"
}
