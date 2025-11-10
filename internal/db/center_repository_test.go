package db_test

import (
	"testing"

	"example.com/avalanche/internal/db"
	"example.com/avalanche/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCenterRepository_GetActiveCenters_MapsCenterID(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := gdb.AutoMigrate(&models.AvalancheCenter{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	seed := models.AvalancheCenter{ID: "IPAC", Name: "Intermountain", URL: "https://api.avalanche.org/v2/public", Active: true}
	if err := gdb.Create(&seed).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	repo := db.NewCenterRepository(gdb)
	centers, err := repo.GetActiveCenters()
	if err != nil {
		t.Fatalf("GetActiveCenters error: %v", err)
	}
	if len(centers) != 1 {
		t.Fatalf("expected 1 center, got %d", len(centers))
	}
	if centers[0].ID != "IPAC" {
		t.Fatalf("expected ID 'IPAC' from center_id, got %q", centers[0].ID)
	}
}
