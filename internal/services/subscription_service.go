package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"example.com/avalanche/internal/db"
	"example.com/avalanche/internal/domain"
	"example.com/avalanche/internal/models"
	"example.com/avalanche/internal/notifier"
)

// SubscriptionService handles the business logic for managing avalanche forecast subscriptions.
type SubscriptionService struct {
	subRepo    *db.SubscriptionRepository
	centerRepo *db.CenterRepository
	forecast   *ForecastService
	emailer    notifier.EmailSender
}

// NewSubscriptionService creates a new subscription service with all required dependencies.
func NewSubscriptionService(
	subRepo *db.SubscriptionRepository,
	centerRepo *db.CenterRepository,
	forecast *ForecastService,
	emailer notifier.EmailSender,
) *SubscriptionService {
	return &SubscriptionService{
		subRepo:    subRepo,
		centerRepo: centerRepo,
		forecast:   forecast,
		emailer:    emailer,
	}
}

type CreateSubscriptionRequest struct {
	Email  *domain.Email
	ZoneID *domain.ZoneID
}

// Create creates a new subscription and sends a welcome email asynchronously.
// It returns the created subscription or an error if the operation fails.
func (s *SubscriptionService) Create(ctx context.Context, req CreateSubscriptionRequest) (*models.Subscription, error) {
	sub := &models.Subscription{
		Email:  req.Email.String(),
		ZoneID: req.ZoneID.String(),
	}

	if err := s.subRepo.Create(sub); err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	go s.sendWelcomeEmail(context.Background(), sub, req.ZoneID)

	return sub, nil
}

// Delete removes a subscription for the given email and zone ID.
func (s *SubscriptionService) Delete(ctx context.Context, email *domain.Email, zoneID *domain.ZoneID) error {
	if err := s.subRepo.Delete(email.String(), zoneID.String()); err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}
	return nil
}

// GetByEmail retrieves all subscriptions for a given email address.
func (s *SubscriptionService) GetByEmail(ctx context.Context, email *domain.Email) ([]models.Subscription, error) {
	subs, err := s.subRepo.GetByEmail(email.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions by email: %w", err)
	}
	return subs, nil
}

// GetByZone retrieves all subscriptions for a given zone ID.
func (s *SubscriptionService) GetByZone(ctx context.Context, zoneID *domain.ZoneID) ([]models.Subscription, error) {
	subs, err := s.subRepo.GetByZone(zoneID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions by zone: %w", err)
	}
	return subs, nil
}

// sendWelcomeEmail sends a welcome email with the latest forecast to a new subscriber.
func (s *SubscriptionService) sendWelcomeEmail(ctx context.Context, sub *models.Subscription, zoneID *domain.ZoneID) {
	centerID := zoneID.Center()

	forecasts, err := s.forecast.GetForecastsForCenters([]string{centerID}, time.Now().UTC())
	if err != nil {
		log.Printf("[SubscriptionService] failed to fetch forecasts for welcome email (center=%s): %v", centerID, err)
		return
	}

	if len(forecasts) == 0 {
		log.Printf("[SubscriptionService] no forecasts available for welcome email (center=%s)", centerID)
		return
	}

	center, err := s.getCenterInfo(centerID)
	if err != nil {
		log.Printf("[SubscriptionService] failed to fetch center info for welcome email (center=%s): %v", centerID, err)
		center = &models.AvalancheCenter{
			ID:   centerID,
			Name: centerID,
			URL:  "https://avalanche.org",
		}
	}

	if zoneID.IsCenterLevel() {
		s.sendCenterWelcomeEmail(ctx, sub, forecasts, center)
	} else {
		s.sendZoneWelcomeEmail(ctx, sub, forecasts, center, zoneID)
	}
}

// sendZoneWelcomeEmail sends a welcome email for a specific zone subscription.
func (s *SubscriptionService) sendZoneWelcomeEmail(
	ctx context.Context,
	sub *models.Subscription,
	forecasts []models.ZoneForecast,
	center *models.AvalancheCenter,
	zoneID *domain.ZoneID,
) {
	targetZoneID := zoneID.String()

	log.Printf("[SubscriptionService] looking for zone %s, found %d forecasts:", targetZoneID, len(forecasts))
	for _, f := range forecasts {
		log.Printf("[SubscriptionService]   - %s (%s)", f.ZoneID, f.ZoneName)
	}

	for _, forecast := range forecasts {
		if forecast.ZoneID == targetZoneID {
			emailData := notifier.EmailData{
				ZoneID:     forecast.ZoneID,
				ZoneName:   forecast.ZoneName,
				IssuedAt:   time.Now().UTC(),
				Today:      forecast.TodayDanger,
				Tomorrow:   forecast.FutureDanger,
				CenterLink: center.URL,
			}

			if err := s.emailer.SendForecastEmail(ctx, sub.Email, emailData); err != nil {
				log.Printf("[SubscriptionService] failed to send zone welcome email (zone=%s, email=%s): %v",
					targetZoneID, sub.Email, err)
			} else {
				log.Printf("[SubscriptionService] sent zone welcome email (zone=%s, email=%s)",
					targetZoneID, sub.Email)
			}
			return
		}
	}

	log.Printf("[SubscriptionService] no forecast found for zone %s in welcome email", targetZoneID)
}

// sendCenterWelcomeEmail sends a welcome email with all zones for a center-level subscription.
func (s *SubscriptionService) sendCenterWelcomeEmail(
	ctx context.Context,
	sub *models.Subscription,
	forecasts []models.ZoneForecast,
	center *models.AvalancheCenter,
) {
	summaries := notifier.BuildZoneSummaries(forecasts)

	if err := s.emailer.SendCenterForecastEmail(ctx, sub.Email, center.Name, center.URL, summaries); err != nil {
		log.Printf("[SubscriptionService] failed to send center welcome email (center=%s, email=%s): %v",
			center.ID, sub.Email, err)
	} else {
		log.Printf("[SubscriptionService] sent center welcome email (center=%s, email=%s, zones=%d)",
			center.ID, sub.Email, len(summaries))
	}
}

// getCenterInfo retrieves avalanche center information from the database.
func (s *SubscriptionService) getCenterInfo(centerID string) (*models.AvalancheCenter, error) {
	centers, err := s.centerRepo.GetActiveCenters()
	if err != nil {
		return nil, err
	}

	for _, c := range centers {
		if c.ID == centerID {
			return &c, nil
		}
	}

	return nil, fmt.Errorf("center %s not found", centerID)
}
