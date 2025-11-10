package domain_test

import (
	"testing"

	"example.com/avalanche/internal/domain"
)

func TestParseZoneID(t *testing.T) {
	cases := []struct {
		in       string
		ok       bool
		center   string
		zone     string
		full     string
		centerLv bool
	}{
		{"NWAC_10", true, "NWAC", "10", "NWAC_10", false},
		{"NWAC", true, "NWAC", "", "NWAC", true},
		{"", false, "", "", "", false},
	}
	for _, c := range cases {
		z, err := domain.ParseZoneID(c.in)
		if c.ok && err != nil {
			t.Fatalf("expected ok for %q: %v", c.in, err)
		}
		if !c.ok && err == nil {
			continue
		}
		if c.ok {
			if z.Center() != c.center {
				if z.Center() != c.center {
					t.Fatalf("expected center %q got %q", c.center, z.Center())
				}
			}
			if z.IsCenterLevel() != c.centerLv {
				if z.IsCenterLevel() != c.centerLv {
					t.Fatalf("center level mismatch for %q", c.in)
				}
			}
		}
	}
}
