package models

import (
	"testing"
	"time"
)

func TestSubscription_GetCenterID_ParseValueObject(t *testing.T) {
	s := &Subscription{ZoneID: "NWAC_10"}
	if got := s.GetCenterID(); got != "NWAC" {
		t.Fatalf("expected NWAC, got %s", got)
	}
}

func TestSubscription_LevelChecks(t *testing.T) {
	c := &Subscription{ZoneID: "NWAC"}
	z := &Subscription{ZoneID: "NWAC_10"}
	if !c.IsCenterLevel() || c.IsSpecificZone() {
		t.Fatalf("expected center-level true for NWAC")
	}
	if z.IsCenterLevel() || !z.IsSpecificZone() {
		t.Fatalf("expected zone-level true for NWAC_10")
	}
}

func TestSubscription_IsDueForNotification(t *testing.T) {
	s := &Subscription{ZoneID: "NWAC_10"}
	if !s.IsDueForNotification(time.Minute) {
		t.Fatalf("never-notified should be due")
	}
	now := time.Now()
	s.MarkNotified(now)
	if s.IsDueForNotification(time.Hour) {
		t.Fatalf("should not be due within interval")
	}
	past := now.Add(-2 * time.Hour)
	s.MarkNotified(past)
	if !s.IsDueForNotification(time.Hour) {
		t.Fatalf("should be due after interval elapsed")
	}
}
