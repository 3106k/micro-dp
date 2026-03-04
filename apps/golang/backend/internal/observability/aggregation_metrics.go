package observability

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

type AggregationMetrics struct {
	ProcessedTotal metric.Int64Counter
	FailedTotal    metric.Int64Counter
	Duration       metric.Float64Histogram
}

func NewAggregationMetrics() *AggregationMetrics {
	meter := otel.Meter("micro-dp-aggregation")

	processed, _ := meter.Int64Counter("aggregation_processed_total",
		metric.WithDescription("Total aggregation runs completed"))
	failed, _ := meter.Int64Counter("aggregation_failed_total",
		metric.WithDescription("Total aggregation runs failed"))
	duration, _ := meter.Float64Histogram("aggregation_duration_seconds",
		metric.WithDescription("Time to run an aggregation"))

	return &AggregationMetrics{
		ProcessedTotal: processed,
		FailedTotal:    failed,
		Duration:       duration,
	}
}
