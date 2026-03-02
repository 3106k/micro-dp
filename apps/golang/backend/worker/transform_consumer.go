package worker

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/observability"
	"github.com/user/micro-dp/usecase"
)

type TransformConsumer struct {
	queue    domain.TransformJobQueue
	writer   *TransformWriter
	metrics  *observability.TransformMetrics
	metering *usecase.MeteringService
	jobRuns  domain.JobRunRepository
}

func NewTransformConsumer(
	queue domain.TransformJobQueue,
	writer *TransformWriter,
	metrics *observability.TransformMetrics,
	metering *usecase.MeteringService,
	jobRuns domain.JobRunRepository,
) *TransformConsumer {
	return &TransformConsumer{
		queue:    queue,
		writer:   writer,
		metrics:  metrics,
		metering: metering,
		jobRuns:  jobRuns,
	}
}

func (c *TransformConsumer) Run(ctx context.Context) {
	log.Println("transform consumer started")

	for {
		select {
		case <-ctx.Done():
			log.Println("transform consumer stopped")
			return
		default:
			msg, err := c.queue.Dequeue(ctx)
			if err != nil {
				if ctx.Err() != nil {
					log.Println("transform consumer stopped")
					return
				}
				log.Printf("transform dequeue error: %v", err)
				continue
			}
			if msg == nil {
				continue
			}

			c.processMessage(ctx, msg)
		}
	}
}

func (c *TransformConsumer) processMessage(ctx context.Context, msg *domain.TransformJobMessage) {
	start := time.Now()

	// Idempotency check
	if err := c.queue.MarkProcessed(ctx, msg.JobRunID); err != nil {
		if errors.Is(err, domain.ErrTransformAlreadyProcessed) {
			log.Printf("transform: skipping duplicate job_run_id=%s", msg.JobRunID)
			c.metrics.DuplicateTotal.Add(ctx, 1)
			return
		}
		log.Printf("transform: mark processed error job_run_id=%s: %v", msg.JobRunID, err)
		c.enqueueDLQ(ctx, msg, err.Error())
		return
	}

	// Update status to running
	if err := c.jobRuns.UpdateStatus(ctx, msg.TenantID, msg.JobRunID, domain.StatusRunning); err != nil {
		log.Printf("transform: update status to running error job_run_id=%s: %v", msg.JobRunID, err)
	}

	// Execute transform
	result, err := c.writer.Execute(ctx, msg)
	if err != nil {
		log.Printf("transform: execute error job_run_id=%s: %v", msg.JobRunID, err)
		c.metrics.FailedTotal.Add(ctx, 1)
		if updateErr := c.jobRuns.UpdateStatus(ctx, msg.TenantID, msg.JobRunID, domain.StatusFailed); updateErr != nil {
			log.Printf("transform: update status to failed error job_run_id=%s: %v", msg.JobRunID, updateErr)
		}
		c.enqueueDLQ(ctx, msg, err.Error())
		c.metrics.Duration.Record(ctx, time.Since(start).Seconds())
		return
	}

	// Success
	if updateErr := c.jobRuns.UpdateStatus(ctx, msg.TenantID, msg.JobRunID, domain.StatusSuccess); updateErr != nil {
		log.Printf("transform: update status to success error job_run_id=%s: %v", msg.JobRunID, updateErr)
	}

	c.metrics.ProcessedTotal.Add(ctx, 1)
	c.metrics.RowsTotal.Add(ctx, result.RowCount)
	c.metrics.Duration.Record(ctx, time.Since(start).Seconds())

	log.Printf("transform: completed job_run_id=%s rows=%d output=%s",
		msg.JobRunID, result.RowCount, result.OutputKey)
}

func (c *TransformConsumer) enqueueDLQ(ctx context.Context, msg *domain.TransformJobMessage, reason string) {
	if err := c.queue.EnqueueDLQ(ctx, msg, reason); err != nil {
		log.Printf("transform: enqueue dlq error job_run_id=%s: %v", msg.JobRunID, err)
	}
}
