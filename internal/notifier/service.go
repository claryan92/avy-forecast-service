package notifier

import (
	"context"
	"log"
	"strings"
	"time"
)

type Service struct {
	repo     Repository
	sender   EmailSender
	interval time.Duration
	fetchFn  ForecastFetcher
}

func NewService(repo Repository, sender EmailSender, interval time.Duration, fetchFn ForecastFetcher) *Service {
	return &Service{repo: repo, sender: sender, interval: interval, fetchFn: fetchFn}
}

func (s *Service) Run(ctx context.Context) error {
	s.checkAndNotify(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkAndNotify(ctx)
		case <-ctx.Done():
			return nil
		}
	}
}

func (s *Service) checkAndNotify(ctx context.Context) {
	centers, err := s.repo.ListSubscribedCenters(ctx)
	if err != nil {
		log.Printf("failed to list centers: %v", err)
		return
	}
	for _, centerID := range centers {
		forecasts, err := s.fetchFn(ctx, centerID)
		if err != nil {
			log.Printf("fetch failed for center %s: %v", centerID, err)
			continue
		}
		centerMeta, err := s.repo.GetCenterByID(ctx, centerID)
		if err != nil {
			log.Printf("center lookup failed for %s: %v", centerID, err)
			continue
		}
		centerURL := centerMeta.URL
		if strings.TrimSpace(centerURL) == "" {
			log.Printf("center %s has no base_url configured", centerID)
			continue
		}
		for _, f := range forecasts {
			lastIssued, err := s.repo.GetLastIssued(ctx, f.ZoneID)
			if err != nil {
				log.Printf("cache read failed for %s: %v", f.ZoneID, err)
				continue
			}
			if !lastIssued.IsZero() && !f.IssuedAt.After(lastIssued) {
				log.Printf("no new forecast for %s", f.ZoneID)
				continue
			}
			subs, err := s.repo.GetSubscriptionsForZone(ctx, f.ZoneID)
			if err != nil {
				log.Printf("subs read failed for %s: %v", f.ZoneID, err)
				continue
			}
			for _, sub := range subs {
				data := EmailData{ZoneID: f.ZoneID, IssuedAt: f.IssuedAt, CenterLink: centerURL}
				if err := s.sender.SendForecastEmail(ctx, sub.Email, data); err != nil {
					log.Printf("send failed to %s: %v", sub.Email, err)
					continue
				}
				_ = s.repo.UpdateLastNotified(ctx, sub.ID, time.Now())
			}
			if err := s.repo.UpsertLastIssued(ctx, f.ZoneID, f.IssuedAt); err != nil {
				log.Printf("cache update failed for %s: %v", f.ZoneID, err)
			}
		}
	}
}
