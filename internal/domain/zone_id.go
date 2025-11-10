package domain

import (
	"errors"
	"strings"
)

// ZoneID represents a validated zone identifier in the format "CENTER_ZONE" or just "CENTER".
// Examples: "NWAC_164" (specific zone), "NWAC" (center-level subscription)
type ZoneID struct {
	center string
	zone   string
}

// ParseZoneID parses and validates a raw zone identifier.
func ParseZoneID(raw string) (*ZoneID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("zone ID cannot be empty")
	}
	parts := strings.SplitN(raw, "_", 2)
	center := strings.ToUpper(parts[0])
	if center == "" {
		return nil, errors.New("center ID cannot be empty")
	}
	z := &ZoneID{center: center}
	if len(parts) > 1 && parts[1] != "" {
		z.zone = parts[1]
	}
	return z, nil
}

// Center returns the avalanche center ID (e.g., "NWAC", "IPAC").
func (z *ZoneID) Center() string { return z.center }

// Zone returns the specific zone identifier (empty if center-level).
func (z *ZoneID) Zone() string { return z.zone }

// IsCenterLevel indicates a center-level subscription.
func (z *ZoneID) IsCenterLevel() bool { return z.zone == "" }

// IsSpecificZone indicates a zone-level subscription.
func (z *ZoneID) IsSpecificZone() bool { return z.zone != "" }

// String returns canonical string form.
func (z *ZoneID) String() string {
	if z.IsCenterLevel() {
		return z.center
	}
	return z.center + "_" + z.zone
}

// Equals compares two ZoneID values.
func (z *ZoneID) Equals(other *ZoneID) bool {
	if other == nil {
		return false
	}
	return z.center == other.center && z.zone == other.zone
}
