package domain

import (
	"context"
	"errors"
	"time"
)

var ErrEventDuplicate = errors.New("event already processed")

type EventQueueMessage struct {
	EventID    string    `json:"event_id"`
	TenantID   string    `json:"tenant_id"`
	EventName  string    `json:"event_name"`
	Properties string    `json:"properties"`
	EventTime  time.Time `json:"event_time"`
	ReceivedAt time.Time `json:"received_at"`
}

type EventQueue interface {
	CheckDuplicate(ctx context.Context, tenantID, eventID string) (bool, error)
	MarkProcessed(ctx context.Context, tenantID, eventID string) error
	Enqueue(ctx context.Context, msg *EventQueueMessage) error
	Dequeue(ctx context.Context) (*EventQueueMessage, error)
	EnqueueDLQ(ctx context.Context, msg *EventQueueMessage, reason string) error
}
