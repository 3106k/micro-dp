package observability

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

type UploadMetrics struct {
	ProcessedTotal metric.Int64Counter
	FailedTotal    metric.Int64Counter
	FilesConverted metric.Int64Counter
	RowsTotal      metric.Int64Counter
	DuplicateTotal metric.Int64Counter
	Duration       metric.Float64Histogram
}

func NewUploadMetrics() *UploadMetrics {
	meter := otel.Meter("micro-dp-uploads")

	processed, _ := meter.Int64Counter("uploads_processed_total",
		metric.WithDescription("Total uploads processed by worker"))
	failed, _ := meter.Int64Counter("uploads_failed_total",
		metric.WithDescription("Total uploads failed and sent to DLQ"))
	files, _ := meter.Int64Counter("uploads_files_converted_total",
		metric.WithDescription("Total CSV files converted to Parquet"))
	rows, _ := meter.Int64Counter("uploads_rows_total",
		metric.WithDescription("Total rows imported from CSV files"))
	duplicate, _ := meter.Int64Counter("uploads_duplicate_total",
		metric.WithDescription("Total duplicate uploads skipped"))
	duration, _ := meter.Float64Histogram("uploads_processing_duration_seconds",
		metric.WithDescription("Time to process an upload"))

	return &UploadMetrics{
		ProcessedTotal: processed,
		FailedTotal:    failed,
		FilesConverted: files,
		RowsTotal:      rows,
		DuplicateTotal: duplicate,
		Duration:       duration,
	}
}
