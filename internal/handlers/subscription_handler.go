package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"example.com/avalanche/internal/domain"
	"example.com/avalanche/internal/models"
	"example.com/avalanche/internal/services"
)

// SubscriptionHandler handles HTTP requests for subscription management.
type SubscriptionHandler struct {
	service *services.SubscriptionService
}

// NewSubscriptionHandler creates a new subscription handler with the given service.
func NewSubscriptionHandler(service *services.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{service: service}
}

// POST /api/subscriptions
func (h *SubscriptionHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email  string `json:"email"`
		ZoneID string `json:"zone_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	email, err := domain.NewEmail(req.Email)
	if err != nil {
		http.Error(w, "invalid email: "+err.Error(), http.StatusBadRequest)
		return
	}

	zoneID, err := domain.ParseZoneID(req.ZoneID)
	if err != nil {
		http.Error(w, "invalid zone_id: "+err.Error(), http.StatusBadRequest)
		return
	}

	sub, err := h.service.Create(r.Context(), services.CreateSubscriptionRequest{
		Email:  email,
		ZoneID: zoneID,
	})
	if err != nil {
		log.Printf("[SubscriptionHandler] failed to create subscription: %v", err)
		http.Error(w, "failed to create subscription", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sub)
}

// DELETE /api/subscriptions
func (h *SubscriptionHandler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	emailStr := r.URL.Query().Get("email")
	zoneIDStr := r.URL.Query().Get("zone_id")

	if emailStr == "" || zoneIDStr == "" {
		http.Error(w, "email and zone_id query parameters are required", http.StatusBadRequest)
		return
	}

	email, err := domain.NewEmail(emailStr)
	if err != nil {
		http.Error(w, "invalid email: "+err.Error(), http.StatusBadRequest)
		return
	}

	zoneID, err := domain.ParseZoneID(zoneIDStr)
	if err != nil {
		http.Error(w, "invalid zone_id: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Delete subscription via service
	if err := h.service.Delete(r.Context(), email, zoneID); err != nil {
		log.Printf("[SubscriptionHandler] failed to delete subscription: %v", err)
		http.Error(w, "failed to delete subscription", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET /api/subscriptions?email=EMAIL or /api/subscriptions?zone_id=ZONE_ID
// Consolidated GET handler that supports fetching by either email or zone_id.
// If both are provided, email takes precedence.
func (h *SubscriptionHandler) GetSubscriptions(w http.ResponseWriter, r *http.Request) {
	emailStr := r.URL.Query().Get("email")
	zoneIDStr := r.URL.Query().Get("zone_id")

	if emailStr == "" && zoneIDStr == "" {
		http.Error(w, "either email or zone_id query parameter is required", http.StatusBadRequest)
		return
	}

	var (
		subs []models.Subscription
		err  error
	)

	if emailStr != "" {
		email, parseErr := domain.NewEmail(emailStr)
		if parseErr != nil {
			http.Error(w, "invalid email: "+parseErr.Error(), http.StatusBadRequest)
			return
		}
		subs, err = h.service.GetByEmail(r.Context(), email)
	} else {
		zoneID, parseErr := domain.ParseZoneID(zoneIDStr)
		if parseErr != nil {
			http.Error(w, "invalid zone_id: "+parseErr.Error(), http.StatusBadRequest)
			return
		}
		subs, err = h.service.GetByZone(r.Context(), zoneID)
	}

	if err != nil {
		log.Printf("[SubscriptionHandler] failed to get subscriptions: %v", err)
		http.Error(w, "failed to load subscriptions", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(subs)
}
