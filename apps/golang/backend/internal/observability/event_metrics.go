package observability

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

type EventMetrics struct {
	ReceivedTotal   metric.Int64Counter
	EnqueuedTotal   metric.Int64Counter
	DuplicateTotal  metric.Int64Counter
	ProcessedTotal  metric.Int64Counter
	FailedTotal     metric.Int64Counter
	BatchSize       metric.Int64Histogram
	BatchDuration   metric.Float64Histogram
}

func NewEventMetrics() *EventMetrics {
	meter := otel.Meter("micro-dp-events")

	received, _ := meter.Int64Counter("events_received_total",
		metric.WithDescription("Total events received by API"))
	enqueued, _ := meter.Int64Counter("events_enqueued_total",
		metric.WithDescription("Total events enqueued successfully"))
	duplicate, _ := meter.Int64Counter("events_duplicate_total",
		metric.WithDescription("Total duplicate events skipped"))
	processed, _ := meter.Int64Counter("events_processed_total",
		metric.WithDescription("Total events processed by worker"))
	failed, _ := meter.Int64Counter("events_failed_total",
		metric.WithDescription("Total events failed and sent to DLQ"))
	batchSize, _ := meter.Int64Histogram("events_batch_size",
		metric.WithDescription("Number of events per batch"))
	batchDuration, _ := meter.Float64Histogram("events_batch_duration_seconds",
		metric.WithDescription("Time to process a batch"))

	return &EventMetrics{
		ReceivedTotal:  received,
		EnqueuedTotal:  enqueued,
		DuplicateTotal: duplicate,
		ProcessedTotal: processed,
		FailedTotal:    failed,
		BatchSize:      batchSize,
		BatchDuration:  batchDuration,
	}
}
