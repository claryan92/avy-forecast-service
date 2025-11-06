package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"example.com/avalanche/internal/models"
)

type ForecastService interface {
	GetForecastsForCenters(centerIDs []string, targetDate time.Time) ([]models.ZoneForecast, error)
}

type CenterRepository interface {
	GetActiveCenters() ([]models.AvalancheCenter, error)
}

type ForecastHandler struct {
	service ForecastService
	repo    CenterRepository
}

func NewForecastHandlerWithRepo(s ForecastService, r CenterRepository) *ForecastHandler {
	return &ForecastHandler{service: s, repo: r}
}

func (h *ForecastHandler) GetForecast(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	centersStr := r.URL.Query().Get("centers")

	var targetDate time.Time
	var err error
	if dateStr == "" {
		targetDate = time.Now().UTC()
	} else {
		targetDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			http.Error(w, "invalid date format (use YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
	}

	var centerIDs []string
	if centersStr != "" {
		parts := strings.Split(centersStr, ",")
		for _, p := range parts {
			if s := strings.TrimSpace(p); s != "" {
				centerIDs = append(centerIDs, s)
			}
		}
	} else {
		centers, err := h.repo.GetActiveCenters()
		if err != nil {
			http.Error(w, "failed to load centers: "+err.Error(), http.StatusInternalServerError)
			return
		}
		for _, c := range centers {
			centerIDs = append(centerIDs, c.ID)
		}
	}

	results, err := h.service.GetForecastsForCenters(centerIDs, targetDate)
	if err != nil {
		http.Error(w, "error fetching forecasts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(results)
}
