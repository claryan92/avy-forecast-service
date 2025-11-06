package models

import "time"

// AvalancheCenter represents a specific avalanche forecasting center,
// such as IPAC (Idaho Panhandle Avalanche Center) or NWAC (Northwest Avalanche Center).
// It provides identifying information used to retrieve and label forecast data.
type AvalancheCenter struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

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
