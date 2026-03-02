package observability

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

type TransformMetrics struct {
	ProcessedTotal metric.Int64Counter
	FailedTotal    metric.Int64Counter
	DuplicateTotal metric.Int64Counter
	RowsTotal      metric.Int64Counter
	Duration       metric.Float64Histogram
}

func NewTransformMetrics() *TransformMetrics {
	meter := otel.Meter("micro-dp-transforms")

	processed, _ := meter.Int64Counter("transforms_processed_total",
		metric.WithDescription("Total transforms processed by worker"))
	failed, _ := meter.Int64Counter("transforms_failed_total",
		metric.WithDescription("Total transforms failed and sent to DLQ"))
	duplicate, _ := meter.Int64Counter("transforms_duplicate_total",
		metric.WithDescription("Total duplicate transforms skipped"))
	rows, _ := meter.Int64Counter("transforms_rows_total",
		metric.WithDescription("Total rows output by transforms"))
	duration, _ := meter.Float64Histogram("transforms_processing_duration_seconds",
		metric.WithDescription("Time to process a transform"))

	return &TransformMetrics{
		ProcessedTotal: processed,
		FailedTotal:    failed,
		DuplicateTotal: duplicate,
		RowsTotal:      rows,
		Duration:       duration,
	}
}
