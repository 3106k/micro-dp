package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/user/micro-dp/domain"
)

type AggregationBackfillService struct {
	queue   domain.AggregationQueue
	tenants domain.TenantRepository
}

func NewAggregationBackfillService(queue domain.AggregationQueue, tenants domain.TenantRepository) *AggregationBackfillService {
	return &AggregationBackfillService{queue: queue, tenants: tenants}
}

type BackfillResult struct {
	Enqueued int
	Skipped  int
}

func (s *AggregationBackfillService) TriggerBackfill(ctx context.Context, tenantID string, startDate, endDate time.Time, force bool) (*BackfillResult, error) {
	var tenantIDs []string
	if tenantID != "" {
		tenantIDs = []string{tenantID}
	} else {
		tenants, err := s.tenants.ListAll(ctx)
		if err != nil {
			return nil, fmt.Errorf("list tenants: %w", err)
		}
		for _, t := range tenants {
			tenantIDs = append(tenantIDs, t.ID)
		}
	}

	result := &BackfillResult{}
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		date := d.Format("2006-01-02")
		for _, tid := range tenantIDs {
			if force {
				if err := s.queue.ResetProcessed(ctx, tid, date); err != nil {
					log.Printf("reset processed error tenant=%s date=%s: %v", tid, date, err)
				}
			} else {
				processed, err := s.queue.IsProcessed(ctx, tid, date)
				if err != nil {
					log.Printf("check processed error tenant=%s date=%s: %v", tid, date, err)
				}
				if processed {
					result.Skipped++
					continue
				}
			}

			msg := domain.AggregationMessage{TenantID: tid, Date: date}
			if err := s.queue.Enqueue(ctx, msg); err != nil {
				return nil, fmt.Errorf("enqueue backfill tenant=%s date=%s: %w", tid, date, err)
			}
			result.Enqueued++
		}
	}

	return result, nil
}
