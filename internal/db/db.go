package db

import (
	"example.com/avalanche/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewConnection(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

type CenterRepository struct {
	db *gorm.DB
}

func NewCenterRepository(db *gorm.DB) *CenterRepository {
	return &CenterRepository{db: db}
}

func (r *CenterRepository) GetActiveCenters() ([]models.AvalancheCenter, error) {
	var centers []models.AvalancheCenter
	if err := r.db.Where("active = ?", true).Find(&centers).Error; err != nil {
		return nil, err
	}
	return centers, nil
}
