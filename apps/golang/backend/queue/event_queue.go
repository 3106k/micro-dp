package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/user/micro-dp/domain"
)

const (
	keyPrefix    = "micro-dp:events:"
	ingestKey    = keyPrefix + "ingest"
	dlqKey       = keyPrefix + "dlq"
	seenPrefix   = keyPrefix + "seen:"
	seenTTL      = 24 * time.Hour
	dequeueWait  = 5 * time.Second
)

type EventQueueImpl struct {
	rdb *redis.Client
}

func NewEventQueue(client *ValkeyClient) *EventQueueImpl {
	return &EventQueueImpl{rdb: client.Client()}
}

func (q *EventQueueImpl) CheckDuplicate(ctx context.Context, tenantID, eventID string) (bool, error) {
	key := seenPrefix + tenantID + ":" + eventID
	exists, err := q.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("check duplicate: %w", err)
	}
	return exists > 0, nil
}

func (q *EventQueueImpl) MarkProcessed(ctx context.Context, tenantID, eventID string) error {
	key := seenPrefix + tenantID + ":" + eventID
	ok, err := q.rdb.SetArgs(ctx, key, "1", redis.SetArgs{
		Mode: "NX",
		TTL:  seenTTL,
	}).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("mark processed: %w", err)
	}
	if ok != "OK" {
		return domain.ErrEventDuplicate
	}
	return nil
}

func (q *EventQueueImpl) Enqueue(ctx context.Context, msg *domain.EventQueueMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	return q.rdb.LPush(ctx, ingestKey, data).Err()
}

func (q *EventQueueImpl) Dequeue(ctx context.Context) (*domain.EventQueueMessage, error) {
	result, err := q.rdb.BRPop(ctx, dequeueWait, ingestKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("dequeue: %w", err)
	}

	var msg domain.EventQueueMessage
	if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
		return nil, fmt.Errorf("unmarshal event: %w", err)
	}
	return &msg, nil
}

func (q *EventQueueImpl) EnqueueDLQ(ctx context.Context, msg *domain.EventQueueMessage, reason string) error {
	wrapper := struct {
		Event  *domain.EventQueueMessage `json:"event"`
		Reason string                    `json:"reason"`
		Time   time.Time                 `json:"time"`
	}{
		Event:  msg,
		Reason: reason,
		Time:   time.Now().UTC(),
	}
	data, err := json.Marshal(wrapper)
	if err != nil {
		return fmt.Errorf("marshal dlq: %w", err)
	}
	return q.rdb.LPush(ctx, dlqKey, data).Err()
}
