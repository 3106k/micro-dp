package worker

import (
	"context"
	"log"
	"time"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/observability"
)

type AggregationConsumer struct {
	tenants  domain.TenantRepository
	writer   *AggregationWriter
	metrics  *observability.AggregationMetrics
	interval time.Duration
}

func NewAggregationConsumer(
	tenants domain.TenantRepository,
	writer *AggregationWriter,
	metrics *observability.AggregationMetrics,
	interval time.Duration,
) *AggregationConsumer {
	return &AggregationConsumer{
		tenants:  tenants,
		writer:   writer,
		metrics:  metrics,
		interval: interval,
	}
}

func (c *AggregationConsumer) Run(ctx context.Context) {
	log.Printf("aggregation consumer started (interval=%s)", c.interval)

	// Run once on startup after a short delay
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("aggregation consumer stopped")
			return
		case <-timer.C:
			c.runAggregation(ctx)
			timer.Reset(c.interval)
		}
	}
}

func (c *AggregationConsumer) runAggregation(ctx context.Context) {
	tenants, err := c.tenants.ListAll(ctx)
	if err != nil {
		log.Printf("aggregation: list tenants error: %v", err)
		return
	}

	// Aggregate today's and yesterday's data
	today := time.Now().UTC().Format("2006-01-02")
	yesterday := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")
	dates := []string{yesterday, today}

	for _, tenant := range tenants {
		for _, date := range dates {
			start := time.Now()

			if err := c.writer.AggregateEvents(ctx, tenant.ID, date); err != nil {
				log.Printf("aggregation: tenant=%s date=%s error: %v", tenant.ID, date, err)
				c.metrics.FailedTotal.Add(ctx, 1)
				continue
			}

			duration := time.Since(start)
			c.metrics.ProcessedTotal.Add(ctx, 1)
			c.metrics.Duration.Record(ctx, duration.Seconds())
			log.Printf("aggregation: tenant=%s date=%s completed in %s", tenant.ID, date, duration)
		}
	}
}
