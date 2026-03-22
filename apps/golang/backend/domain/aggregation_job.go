package domain

import (
	"context"
	"errors"
)

var ErrAggregationAlreadyProcessed = errors.New("aggregation already processed")

type AggregationMessage struct {
	TenantID string `json:"tenant_id"`
	Date     string `json:"date"`
}

type AggregationQueue interface {
	Enqueue(ctx context.Context, msg AggregationMessage) error
	Dequeue(ctx context.Context) (*AggregationMessage, error)
	MarkProcessed(ctx context.Context, tenantID, date string) error
	ResetProcessed(ctx context.Context, tenantID, date string) error
	IsProcessed(ctx context.Context, tenantID, date string) (bool, error)
	EnqueueDLQ(ctx context.Context, msg AggregationMessage, reason string) error
}
