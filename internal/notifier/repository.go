package notifier

import (
	"context"
	"errors"
	"strings"
	"time"

	"example.com/avalanche/internal/models"
	"gorm.io/gorm"
)

// SubscriptionReader provides read-only access to subscription data.
type SubscriptionReader interface {
	ListSubscribedCenters(ctx context.Context) ([]string, error)
	ListSubscribedZones(ctx context.Context) ([]string, error)
	GetSubscriptionsForZone(ctx context.Context, zoneID string) ([]models.Subscription, error)
}

// SubscriptionWriter provides write access to subscription data.
type SubscriptionWriter interface {
	UpdateLastNotified(ctx context.Context, subID uint, t time.Time) error
}

// ForecastCache provides access to forecast cache data for tracking last issued times.
type ForecastCache interface {
	GetLastIssued(ctx context.Context, zoneID string) (time.Time, error)
	UpsertLastIssued(ctx context.Context, zoneID string, issuedAt time.Time) error
}

// Repository combines all data access interfaces for the notifier service.
type Repository interface {
	SubscriptionReader
	SubscriptionWriter
	ForecastCache
	CenterLookup
}

// CenterLookup provides access to avalanche center metadata.
type CenterLookup interface {
	GetCenterByID(ctx context.Context, centerID string) (*models.AvalancheCenter, error)
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) GetSubscriptionsForZone(ctx context.Context, zoneID string) ([]models.Subscription, error) {
	var subs []models.Subscription
	if err := r.db.WithContext(ctx).Where("zone_id = ?", zoneID).Find(&subs).Error; err != nil {
		return nil, err
	}
	return subs, nil
}

func (r *GormRepository) UpdateLastNotified(ctx context.Context, subID uint, t time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.Subscription{}).
		Where("id = ?", subID).
		Update("last_notified", t).Error
}

func (r *GormRepository) GetLastIssued(ctx context.Context, zoneID string) (time.Time, error) {
	var fc models.ForecastCache
	err := r.db.WithContext(ctx).First(&fc, "zone_id = ?", zoneID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return time.Time{}, nil
	}
	return fc.LastIssued, err
}

func (r *GormRepository) UpsertLastIssued(ctx context.Context, zoneID string, issuedAt time.Time) error {
	fc := models.ForecastCache{ZoneID: zoneID, LastIssued: issuedAt}
	return r.db.WithContext(ctx).Save(&fc).Error
}

// GetCenterByID returns center metadata for a given center ID.
func (r *GormRepository) GetCenterByID(ctx context.Context, centerID string) (*models.AvalancheCenter, error) {
	var center models.AvalancheCenter
	res := r.db.WithContext(ctx).Where("LOWER(center_id) = LOWER(?)", centerID).Find(&center)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &center, nil
}

func (r *GormRepository) ListSubscribedZones(ctx context.Context) ([]string, error) {
	var zones []string
	if err := r.db.WithContext(ctx).Model(&models.Subscription{}).Distinct().Pluck("zone_id", &zones).Error; err != nil {
		return nil, err
	}
	return zones, nil
}

// ListSubscribedCenters returns all unique center IDs (prefix before underscore in zone_id) with at least one subscription.
func (r *GormRepository) ListSubscribedCenters(ctx context.Context) ([]string, error) {
	var zoneIDs []string
	if err := r.db.WithContext(ctx).Model(&models.Subscription{}).Distinct().Pluck("zone_id", &zoneIDs).Error; err != nil {
		return nil, err
	}
	centerSet := make(map[string]struct{})
	for _, z := range zoneIDs {

		if idx := strings.Index(z, "_"); idx > 0 {
			centerSet[z[:idx]] = struct{}{}
		} else if z != "" {
			// For center-level subscriptions (e.g., "NWAC"), use the whole ID as the center
			centerSet[z] = struct{}{}
		}
	}
	centers := make([]string, 0, len(centerSet))
	for c := range centerSet {
		centers = append(centers, c)
	}
	return centers, nil
}
