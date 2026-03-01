package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/user/micro-dp/domain"
)

type EventService struct {
	queue domain.EventQueue
}

func NewEventService(queue domain.EventQueue) *EventService {
	return &EventService{queue: queue}
}

func (s *EventService) Ingest(ctx context.Context, eventID, eventName string, properties map[string]any, eventTime time.Time) error {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant id not found in context")
	}

	dup, err := s.queue.CheckDuplicate(ctx, tenantID, eventID)
	if err != nil {
		return fmt.Errorf("check duplicate: %w", err)
	}
	if dup {
		return domain.ErrEventDuplicate
	}

	if err := s.queue.MarkProcessed(ctx, tenantID, eventID); err != nil {
		return err
	}

	propsJSON := "{}"
	if properties != nil {
		b, err := json.Marshal(properties)
		if err != nil {
			return fmt.Errorf("marshal properties: %w", err)
		}
		propsJSON = string(b)
	}

	msg := &domain.EventQueueMessage{
		EventID:    eventID,
		TenantID:   tenantID,
		EventName:  eventName,
		Properties: propsJSON,
		EventTime:  eventTime,
		ReceivedAt: time.Now().UTC(),
	}

	if err := s.queue.Enqueue(ctx, msg); err != nil {
		return err
	}

	// Best-effort counter increment (don't fail the ingest on counter error)
	_ = s.queue.IncrementCount(ctx, tenantID, eventName)

	return nil
}

func (s *EventService) Summary(ctx context.Context) (map[string]int64, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}
	return s.queue.GetCounts(ctx, tenantID)
}
