package services_test

import (
	"errors"
	"reflect"
	"sort"
	"testing"
	"time"

	"example.com/avalanche/internal/models"
	"example.com/avalanche/internal/services"
)

type mockForecastClient struct {
	data map[string][]models.Forecast
	err  error
}

func (m *mockForecastClient) FetchForecasts(centerID string) ([]models.Forecast, error) {
	if m.err != nil {
		return nil, m.err
	}
	if forecasts, ok := m.data[centerID]; ok {
		return forecasts, nil
	}
	return nil, nil
}

func TestSortForecastsByPublishTime(t *testing.T) {
	now := time.Now().UTC()
	older := now.Add(-time.Hour)

	in := []models.Forecast{
		{PublishedTime: older},
		{PublishedTime: now},
	}

	services.SortForecastsByPublishTime(in)

	if !in[0].PublishedTime.After(in[1].PublishedTime) {
		t.Errorf("expected newest forecast first, got: %+v", in)
	}
}

func TestFillMissingDangerRatings(t *testing.T) {
	svc := services.NewForecast(nil)
	zoneForecasts := []models.ZoneForecast{
		{ZoneID: "z1", ZoneName: "Test Zone"},
	}

	svc.FillMissingDangerRatings(zoneForecasts)

	zf := zoneForecasts[0]
	if zf.TodayDanger == nil || zf.TodayDanger.Message != "No forecast available today" {
		t.Error("expected default today danger message")
	}
	if zf.FutureDanger == nil || zf.FutureDanger.Message != "No forecast available tomorrow" {
		t.Error("expected default tomorrow danger message")
	}
}

func TestProcessForecasts_SortsAndFills(t *testing.T) {
	now := time.Now().UTC()
	bottom := "Some risk at upper elevations"
	mockForecasts := []models.Forecast{
		{
			PublishedTime: now,
			StartDate:     now.Add(-2 * time.Hour),
			EndDate:       now.Add(24 * time.Hour),
			AvalancheCenter: models.AvalancheCenter{
				Name: "IPAC",
			},
			ForecastZone: []models.Zone{
				{ZoneID: "z1", Name: "Zone One"},
			},
			Danger: []models.DangerRating{
				{ValidDay: "current", Upper: 2, Middle: 2, Lower: 1},
			},
			BottomLine: bottom,
			Status:     "published",
		},
	}

	svc := services.NewForecast(nil)
	result := svc.ProcessForecasts(mockForecasts)

	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}

	z := result[0]
	if z.ZoneName != "Zone One" {
		t.Errorf("expected Zone One, got %s", z.ZoneName)
	}
	if z.TodayDanger == nil || z.TodayDanger.Upper != 2 {
		t.Errorf("expected danger rating set, got %+v", z.TodayDanger)
	}
	if z.FutureDanger == nil || z.FutureDanger.Message != "No forecast available tomorrow" {
		t.Errorf("expected default tomorrow danger message")
	}
}

func TestGetForecastsForCenters_Success(t *testing.T) {
	now := time.Now().UTC()
	client := &mockForecastClient{
		data: map[string][]models.Forecast{
			"caic": {{
				PublishedTime: now,
				StartDate:     now.Add(-2 * time.Hour),
				EndDate:       now.Add(24 * time.Hour),
				AvalancheCenter: models.AvalancheCenter{
					Name: "IPAC",
				},
				ForecastZone: []models.Zone{
					{ZoneID: "z1", Name: "Zone One"},
				},
				Danger: []models.DangerRating{
					{ValidDay: "current", Upper: 2},
				},
				Status: "published",
			}},
		},
	}

	svc := services.NewForecast(client)

	results, err := svc.GetForecastsForCenters([]string{"caic"}, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected results, got none")
	}
	if results[0].Center != "IPAC" {
		t.Errorf("expected center IPAC, got %s", results[0].Center)
	}
}

func TestGetForecastsForCenters_Error(t *testing.T) {
	client := &mockForecastClient{err: errors.New("fetch failed")}
	svc := services.NewForecast(client)

	_, err := svc.GetForecastsForCenters([]string{"caic"}, time.Now())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSortZoneForecasts(t *testing.T) {
	in := []models.ZoneForecast{
		{Center: "B", ZoneName: "Z2"},
		{Center: "A", ZoneName: "Z3"},
		{Center: "A", ZoneName: "Z1"},
	}

	services.SortZoneForecasts(in)

	expected := []string{"A-Z1", "A-Z3", "B-Z2"}
	got := make([]string, len(in))
	for i, z := range in {
		got[i] = z.Center + "-" + z.ZoneName
	}

	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestFlattenZoneMap(t *testing.T) {
	z1 := &models.ZoneForecast{ZoneID: "1"}
	z2 := &models.ZoneForecast{ZoneID: "2"}
	m := map[string]*models.ZoneForecast{"1": z1, "2": z2}

	res := services.FlattenZoneMap(m)
	sort.Slice(res, func(i, j int) bool { return res[i].ZoneID < res[j].ZoneID })
	if len(res) != 2 {
		t.Errorf("expected 2 zones, got %d", len(res))
	}
	if res[0].ZoneID != "1" || res[1].ZoneID != "2" {
		t.Errorf("unexpected flatten result: %+v", res)
	}
}
