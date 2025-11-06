package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"example.com/avalanche/internal/handlers"
	"example.com/avalanche/internal/models"
)

type mockService struct {
	forecasts []models.ZoneForecast
	err       error
}

func (m *mockService) GetForecastsForCenters(centerIDs []string, targetDate time.Time) ([]models.ZoneForecast, error) {
	return m.forecasts, m.err
}

type mockRepo struct {
	centers []models.AvalancheCenter
	err     error
}

func (m *mockRepo) GetActiveCenters() ([]models.AvalancheCenter, error) {
	return m.centers, m.err
}

func TestForecastHandler_ReturnsJSON(t *testing.T) {
	ms := &mockService{
		forecasts: []models.ZoneForecast{
			{
				ZoneID:   "kootenai",
				ZoneName: "East Cabinet Mountains",
				Center:   "IPAC",
			},
		},
	}
	mr := &mockRepo{
		centers: []models.AvalancheCenter{
			{ID: "IPAC", Name: "IPAC", URL: "https://api.avalanche.org/v2/public"},
		},
	}

	h := handlers.NewForecastHandlerWithRepo(ms, mr)

	req := httptest.NewRequest(http.MethodGet, "/api/forecast", nil)
	rec := httptest.NewRecorder()

	h.GetForecast(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	var out []models.ZoneForecast
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(out) != 1 {
		t.Fatalf("expected 1 forecast, got %d", len(out))
	}
	if out[0].ZoneID != "kootenai" {
		t.Errorf("expected kootenai, got %s", out[0].ZoneID)
	}
}

func TestForecastHandler_BadDate(t *testing.T) {
	ms := &mockService{}
	mr := &mockRepo{
		centers: []models.AvalancheCenter{
			{ID: "IPAC", Name: "IPAC"},
		},
	}

	h := handlers.NewForecastHandlerWithRepo(ms, mr)

	req := httptest.NewRequest(http.MethodGet, "/api/forecast?date=2025-13-40", nil)
	rec := httptest.NewRecorder()

	h.GetForecast(rec, req)

	if rec.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad date, got %d", rec.Result().StatusCode)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "invalid date format") {
		t.Errorf("expected invalid date message, got %s", body)
	}
}
