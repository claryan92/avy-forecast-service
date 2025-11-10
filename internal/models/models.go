package models

import (
	"time"

	"example.com/avalanche/internal/domain"
)

// AvalancheCenter represents a specific avalanche forecasting center,
// such as IPAC (Idaho Panhandle Avalanche Center) or NWAC (Northwest Avalanche Center).
// It provides identifying information used to retrieve and label forecast data.
type AvalancheCenter struct {
	// ID is the domain identifier used by the upstream API (e.g., "IPAC", "NWAC").
	// It maps to the center_id column in the database.
	ID     string `json:"id" gorm:"column:center_id"`
	Name   string `json:"name" gorm:"column:name"`
	URL    string `json:"url,omitempty" gorm:"column:base_url"`
	Active bool   `json:"-" gorm:"column:active"`
}

// TableName ensures GORM uses the correct table name
func (AvalancheCenter) TableName() string { return "avalanche_centers" }

// Forecast describes a single avalanche forecast as published by an avalanche center.
// Each forecast includes metadata such as publication time, validity period,
// bottom-line summary, and associated danger ratings for specific zones.
type Forecast struct {
	AvalancheCenter AvalancheCenter `json:"avalanche_center"`
	ID              int             `json:"id"`
	StartDate       time.Time       `json:"start_date"`
	EndDate         time.Time       `json:"end_date"`
	PublishedTime   time.Time       `json:"published_time"`
	Status          string          `json:"status"`
	BottomLine      string          `json:"bottom_line"`
	DangerLevelText string          `json:"danger_level_text"`
	ForecastZone    []Zone          `json:"forecast_zone"`
	Danger          []DangerRating  `json:"danger"`
}

// Zone represents an individual geographic forecast zone within an avalanche center's coverage area.
// Each zone has an identifying code and display name.
type Zone struct {
	ZoneID string `json:"zone_id"`
	Name   string `json:"name"`
	URL    string `json:"url,omitempty"`
}

// DangerRating describes avalanche danger levels for a specific day and elevation range.
// ValidDay typically corresponds to "current" or "tomorrow".
// Elevation-specific ratings are provided for upper, middle, and lower elevation bands.
type DangerRating struct {
	Upper    int    `json:"upper"`
	Middle   int    `json:"middle"`
	Lower    int    `json:"lower"`
	ValidDay string `json:"valid_day"`
	Message  string `json:"message,omitempty"`
}

// ZoneForecast represents the processed and normalized forecast for a single avalanche zone.
// It merges information from one or more Forecasts into a simplified structure suitable
// for API responses or display in a frontend application.
type ZoneForecast struct {
	ZoneID       string        `json:"zone_id"`
	ZoneName     string        `json:"zone_name"`
	Center       string        `json:"center"`
	IssuedTime   string        `json:"issued_time"`
	StartDate    string        `json:"start_date"`
	EndDate      string        `json:"end_date"`
	BottomLine   string        `json:"bottom_line"`
	TodayDanger  *DangerRating `json:"today_danger,omitempty"`
	FutureDanger *DangerRating `json:"future_danger,omitempty"`
}

type Subscription struct {
	ID           uint       `json:"id,omitempty" gorm:"primaryKey"`
	ZoneID       string     `json:"zone_id" gorm:"index;not null"`
	Email        string     `json:"email" gorm:"index;not null"`
	LastNotified *time.Time `json:"last_notified,omitempty"`
	CreatedAt    time.Time  `json:"created_at,omitempty"`
	UpdatedAt    time.Time  `json:"updated_at,omitempty"`
}

// IsDueForNotification checks if enough time has passed since the last notification.
// Returns true if the subscription has never been notified or if the interval has elapsed.
func (s *Subscription) IsDueForNotification(interval time.Duration) bool {
	if s.LastNotified == nil {
		return true
	}
	return time.Since(*s.LastNotified) >= interval
}

// MarkNotified updates the last notified timestamp to the given time.
func (s *Subscription) MarkNotified(t time.Time) {
	s.LastNotified = &t
}

// GetCenterID extracts the avalanche center ID from the zone ID.
func (s *Subscription) GetCenterID() string {
	if zid, err := domain.ParseZoneID(s.ZoneID); err == nil {
		return zid.Center()
	}
	if idx := stringIndexByte(s.ZoneID, '_'); idx != -1 {
		return s.ZoneID[:idx]
	}
	return s.ZoneID
}

// IsCenterLevel returns true if this is a center-level subscription (no specific zone).
func (s *Subscription) IsCenterLevel() bool {
	return stringIndexByte(s.ZoneID, '_') == -1
}

// IsSpecificZone returns true if this subscription targets a specific zone within a center.
func (s *Subscription) IsSpecificZone() bool {
	return stringIndexByte(s.ZoneID, '_') != -1
}

func stringIndexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

// ForecastCache stores the last issued forecast time per zone for the notifier service.
type ForecastCache struct {
	ZoneID     string    `gorm:"primaryKey"`
	LastIssued time.Time `gorm:"not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// TableName overrides GORM's default pluralization for ForecastCache.
func (ForecastCache) TableName() string { return "forecast_cache" }
