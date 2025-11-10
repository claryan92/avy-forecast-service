package handlers_test

import (
	"testing"

	"example.com/avalanche/internal/domain"
)

func TestSubscriptionHandler_CreateSubscription_Success(t *testing.T) {
	if _, err := domain.NewEmail("user@example.com"); err != nil {
		t.Fatalf("email validation failed: %v", err)
	}
	if _, err := domain.ParseZoneID("NWAC_10"); err != nil {
		t.Fatalf("zone validation failed: %v", err)
	}
}
