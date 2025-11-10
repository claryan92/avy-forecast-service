package db

import (
	"time"

	"example.com/avalanche/internal/models"
	"gorm.io/gorm"
)

type SubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(sub *models.Subscription) error {
	return r.db.Create(sub).Error
}

func (r *SubscriptionRepository) Delete(email, zoneID string) error {
	return r.db.Where("email = ? AND zone_id = ?", email, zoneID).Delete(&models.Subscription{}).Error
}

func (r *SubscriptionRepository) GetByZone(zoneID string) ([]models.Subscription, error) {
	var subs []models.Subscription
	err := r.db.Where("zone_id = ?", zoneID).Find(&subs).Error
	return subs, err
}

func (r *SubscriptionRepository) GetByEmail(email string) ([]models.Subscription, error) {
	var subs []models.Subscription
	err := r.db.Where("email = ?", email).Find(&subs).Error
	return subs, err
}

func (r *SubscriptionRepository) UpdateLastNotified(subID int, t time.Time) error {
	return r.db.Model(&models.Subscription{}).Where("id = ?", subID).Update("last_notified", t).Error
}
