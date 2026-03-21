package domain

import "context"

type AggregationMessage struct {
	TenantID string `json:"tenant_id"`
	Date     string `json:"date"`
}

type AggregationQueue interface {
	Enqueue(ctx context.Context, msg AggregationMessage) error
	Dequeue(ctx context.Context) (*AggregationMessage, error)
	MarkProcessed(ctx context.Context, tenantID, date string) error
	EnqueueDLQ(ctx context.Context, msg AggregationMessage, reason string) error
}
