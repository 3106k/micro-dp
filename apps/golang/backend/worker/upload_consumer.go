package worker

import (
	"context"
	"errors"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/observability"
	"github.com/user/micro-dp/usecase"
)

type UploadConsumer struct {
	queue    domain.UploadJobQueue
	writer   *CSVImportWriter
	metrics  *observability.UploadMetrics
	metering *usecase.MeteringService
}

func NewUploadConsumer(queue domain.UploadJobQueue, writer *CSVImportWriter, metrics *observability.UploadMetrics, metering *usecase.MeteringService) *UploadConsumer {
	return &UploadConsumer{
		queue:    queue,
		writer:   writer,
		metrics:  metrics,
		metering: metering,
	}
}

func (c *UploadConsumer) Run(ctx context.Context) {
	log.Println("upload consumer started")

	for {
		select {
		case <-ctx.Done():
			log.Println("upload consumer stopped")
			return
		default:
			msg, err := c.queue.Dequeue(ctx)
			if err != nil {
				if ctx.Err() != nil {
					log.Println("upload consumer stopped")
					return
				}
				log.Printf("upload dequeue error: %v", err)
				continue
			}
			if msg == nil {
				continue
			}

			c.processMessage(ctx, msg)
		}
	}
}

func (c *UploadConsumer) processMessage(ctx context.Context, msg *domain.UploadJobMessage) {
	start := time.Now()

	// Idempotency check
	if err := c.queue.MarkProcessed(ctx, msg.UploadID); err != nil {
		if errors.Is(err, domain.ErrUploadAlreadyProcessed) {
			log.Printf("csv import: skipping duplicate upload_id=%s", msg.UploadID)
			c.metrics.DuplicateTotal.Add(ctx, 1)
			return
		}
		log.Printf("csv import: mark processed error upload_id=%s: %v", msg.UploadID, err)
		c.enqueueDLQ(ctx, msg, err.Error())
		return
	}

	var totalRows int64
	var filesConverted int64
	var lastErr error

	for _, file := range msg.Files {
		if !isCSV(file.FileName) {
			log.Printf("csv import: skipping non-csv file=%s upload_id=%s", file.FileName, msg.UploadID)
			continue
		}

		result, err := c.writer.ProcessFile(ctx, msg.TenantID, file)
		if err != nil {
			log.Printf("csv import: process file error file=%s upload_id=%s: %v", file.FileName, msg.UploadID, err)
			lastErr = err
			continue
		}

		filesConverted++
		totalRows += result.RowCount
		log.Printf("csv import: converted file=%s rows=%d output=%s upload_id=%s",
			file.FileName, result.RowCount, result.OutputKey, msg.UploadID)
	}

	if lastErr != nil {
		c.metrics.FailedTotal.Add(ctx, 1)
		c.enqueueDLQ(ctx, msg, lastErr.Error())
	} else {
		c.metrics.ProcessedTotal.Add(ctx, 1)
		if c.metering != nil {
			var totalStorageBytes int64
			for _, f := range msg.Files {
				totalStorageBytes += f.SizeBytes
			}
			c.metering.RecordUploadBestEffort(ctx, msg.TenantID, int(totalRows), totalStorageBytes)
			if err := c.metering.RecordUploadCount(ctx, msg.TenantID); err != nil {
				log.Printf("metering record upload count error: %v", err)
			}
		}
	}

	c.metrics.FilesConverted.Add(ctx, filesConverted)
	c.metrics.RowsTotal.Add(ctx, totalRows)
	c.metrics.Duration.Record(ctx, time.Since(start).Seconds())
}

func (c *UploadConsumer) enqueueDLQ(ctx context.Context, msg *domain.UploadJobMessage, reason string) {
	if err := c.queue.EnqueueDLQ(ctx, msg, reason); err != nil {
		log.Printf("csv import: enqueue dlq error upload_id=%s: %v", msg.UploadID, err)
	}
}

func isCSV(filename string) bool {
	return strings.EqualFold(filepath.Ext(filename), ".csv")
}
