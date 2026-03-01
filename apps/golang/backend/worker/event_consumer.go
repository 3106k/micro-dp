package worker

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/observability"
	"github.com/user/micro-dp/usecase"
)

type EventConsumer struct {
	queue         domain.EventQueue
	writer        *ParquetWriter
	metrics       *observability.EventMetrics
	metering      *usecase.MeteringService
	batchSize     int
	flushInterval time.Duration

	mu        sync.Mutex
	buffer    []*domain.EventQueueMessage
	lastFlush time.Time
}

func NewEventConsumer(queue domain.EventQueue, writer *ParquetWriter, metrics *observability.EventMetrics, metering *usecase.MeteringService) *EventConsumer {
	return &EventConsumer{
		queue:         queue,
		writer:        writer,
		metrics:       metrics,
		metering:      metering,
		batchSize:     1000,
		flushInterval: 30 * time.Second,
		buffer:        make([]*domain.EventQueueMessage, 0, 1000),
		lastFlush:     time.Now(),
	}
}

func (c *EventConsumer) Run(ctx context.Context) {
	log.Println("event consumer started")

	flushTicker := time.NewTicker(c.flushInterval)
	defer flushTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.flush(context.Background())
			log.Println("event consumer stopped")
			return
		case <-flushTicker.C:
			c.mu.Lock()
			if len(c.buffer) > 0 && time.Since(c.lastFlush) >= c.flushInterval {
				c.mu.Unlock()
				c.flush(ctx)
			} else {
				c.mu.Unlock()
			}
		default:
			msg, err := c.queue.Dequeue(ctx)
			if err != nil {
				if ctx.Err() != nil {
					c.flush(context.Background())
					log.Println("event consumer stopped")
					return
				}
				log.Printf("event dequeue error: %v", err)
				continue
			}
			if msg == nil {
				continue
			}

			c.mu.Lock()
			c.buffer = append(c.buffer, msg)
			shouldFlush := len(c.buffer) >= c.batchSize
			c.mu.Unlock()

			if shouldFlush {
				c.flush(ctx)
			}
		}
	}
}

func (c *EventConsumer) flush(ctx context.Context) {
	c.mu.Lock()
	if len(c.buffer) == 0 {
		c.mu.Unlock()
		return
	}
	batch := c.buffer
	c.buffer = make([]*domain.EventQueueMessage, 0, c.batchSize)
	c.lastFlush = time.Now()
	c.mu.Unlock()

	// Group by tenant for per-tenant Parquet files
	byTenant := make(map[string][]*domain.EventQueueMessage)
	for _, msg := range batch {
		byTenant[msg.TenantID] = append(byTenant[msg.TenantID], msg)
	}

	for tenantID, events := range byTenant {
		start := time.Now()

		if err := c.writer.WriteBatch(ctx, events); err != nil {
			log.Printf("write batch error tenant=%s count=%d: %v", tenantID, len(events), err)
			c.metrics.FailedTotal.Add(ctx, int64(len(events)))
			for _, msg := range events {
				if dlqErr := c.queue.EnqueueDLQ(ctx, msg, err.Error()); dlqErr != nil {
					log.Printf("enqueue dlq error event=%s: %v", msg.EventID, dlqErr)
				}
			}
		} else {
			log.Printf("flushed batch tenant=%s count=%d", tenantID, len(events))
			c.metrics.ProcessedTotal.Add(ctx, int64(len(events)))
			if c.metering != nil {
				c.metering.RecordEventsBestEffort(ctx, tenantID, len(events))
			}
		}

		c.metrics.BatchSize.Record(ctx, int64(len(events)))
		c.metrics.BatchDuration.Record(ctx, time.Since(start).Seconds())
	}
}
