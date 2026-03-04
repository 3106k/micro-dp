package observability

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

type JobRunMetrics struct {
	DispatchedTotal metric.Int64Counter
	ProcessedTotal  metric.Int64Counter
	FailedTotal     metric.Int64Counter
	DuplicateTotal  metric.Int64Counter
	Duration        metric.Float64Histogram
}

func NewJobRunMetrics() *JobRunMetrics {
	meter := otel.Meter("micro-dp-job-runs")

	dispatched, _ := meter.Int64Counter("job_runs_dispatched_total",
		metric.WithDescription("Total job runs enqueued by poller"))
	processed, _ := meter.Int64Counter("job_runs_processed_total",
		metric.WithDescription("Total job runs processed by consumer"))
	failed, _ := meter.Int64Counter("job_runs_failed_total",
		metric.WithDescription("Total job runs failed and sent to DLQ"))
	duplicate, _ := meter.Int64Counter("job_runs_duplicate_total",
		metric.WithDescription("Total duplicate job runs skipped"))
	duration, _ := meter.Float64Histogram("job_runs_processing_duration_seconds",
		metric.WithDescription("Time to process a job run"))

	return &JobRunMetrics{
		DispatchedTotal: dispatched,
		ProcessedTotal:  processed,
		FailedTotal:     failed,
		DuplicateTotal:  duplicate,
		Duration:        duration,
	}
}
