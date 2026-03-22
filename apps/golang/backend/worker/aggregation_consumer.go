package worker

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/observability"
)

type AggregationConsumer struct {
	queue   domain.AggregationQueue
	writer  *AggregationWriter
	metrics *observability.AggregationMetrics
}

func NewAggregationConsumer(
	queue domain.AggregationQueue,
	writer *AggregationWriter,
	metrics *observability.AggregationMetrics,
) *AggregationConsumer {
	return &AggregationConsumer{
		queue:   queue,
		writer:  writer,
		metrics: metrics,
	}
}

func (c *AggregationConsumer) Run(ctx context.Context) {
	log.Println("aggregation consumer started")

	for {
		select {
		case <-ctx.Done():
			log.Println("aggregation consumer stopped")
			return
		default:
			msg, err := c.queue.Dequeue(ctx)
			if err != nil {
				if ctx.Err() != nil {
					log.Println("aggregation consumer stopped")
					return
				}
				log.Printf("aggregation dequeue error: %v", err)
				continue
			}
			if msg == nil {
				continue
			}

			c.processMessage(ctx, msg)
		}
	}
}

func (c *AggregationConsumer) processMessage(ctx context.Context, msg *domain.AggregationMessage) {
	start := time.Now()

	// Idempotency check
	if err := c.queue.MarkProcessed(ctx, msg.TenantID, msg.Date); err != nil {
		if errors.Is(err, domain.ErrAggregationAlreadyProcessed) {
			log.Printf("aggregation: skipping duplicate tenant=%s date=%s", msg.TenantID, msg.Date)
			c.metrics.DuplicateTotal.Add(ctx, 1)
			return
		}
		log.Printf("aggregation: mark processed error tenant=%s date=%s: %v", msg.TenantID, msg.Date, err)
		c.enqueueDLQ(ctx, msg, err.Error())
		return
	}

	if err := c.writer.AggregateEvents(ctx, msg.TenantID, msg.Date); err != nil {
		log.Printf("aggregation: tenant=%s date=%s error: %v", msg.TenantID, msg.Date, err)
		c.metrics.FailedTotal.Add(ctx, 1)
		c.enqueueDLQ(ctx, msg, err.Error())
		return
	}

	duration := time.Since(start)
	c.metrics.ProcessedTotal.Add(ctx, 1)
	c.metrics.Duration.Record(ctx, duration.Seconds())
	log.Printf("aggregation: tenant=%s date=%s completed in %s", msg.TenantID, msg.Date, duration)
}

func (c *AggregationConsumer) enqueueDLQ(ctx context.Context, msg *domain.AggregationMessage, reason string) {
	if err := c.queue.EnqueueDLQ(ctx, *msg, reason); err != nil {
		log.Printf("aggregation: enqueue dlq error tenant=%s date=%s: %v", msg.TenantID, msg.Date, err)
	}
}
