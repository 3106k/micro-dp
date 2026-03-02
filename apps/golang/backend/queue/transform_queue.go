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
	transformPrefix      = "micro-dp:transforms:"
	transformIngestKey   = transformPrefix + "ingest"
	transformDLQKey      = transformPrefix + "dlq"
	transformSeenPrefix  = transformPrefix + "seen:"
	transformSeenTTL     = 24 * time.Hour
	transformDequeueWait = 5 * time.Second
)

type TransformQueueImpl struct {
	rdb *redis.Client
}

func NewTransformQueue(client *ValkeyClient) *TransformQueueImpl {
	return &TransformQueueImpl{rdb: client.Client()}
}

func (q *TransformQueueImpl) Enqueue(ctx context.Context, msg *domain.TransformJobMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal transform job: %w", err)
	}
	return q.rdb.LPush(ctx, transformIngestKey, data).Err()
}

func (q *TransformQueueImpl) Dequeue(ctx context.Context) (*domain.TransformJobMessage, error) {
	result, err := q.rdb.BRPop(ctx, transformDequeueWait, transformIngestKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("dequeue transform: %w", err)
	}

	var msg domain.TransformJobMessage
	if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
		return nil, fmt.Errorf("unmarshal transform job: %w", err)
	}
	return &msg, nil
}

func (q *TransformQueueImpl) MarkProcessed(ctx context.Context, jobRunID string) error {
	key := transformSeenPrefix + jobRunID
	ok, err := q.rdb.SetArgs(ctx, key, "1", redis.SetArgs{
		Mode: "NX",
		TTL:  transformSeenTTL,
	}).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("mark transform processed: %w", err)
	}
	if ok != "OK" {
		return domain.ErrTransformAlreadyProcessed
	}
	return nil
}

func (q *TransformQueueImpl) EnqueueDLQ(ctx context.Context, msg *domain.TransformJobMessage, reason string) error {
	wrapper := struct {
		Transform *domain.TransformJobMessage `json:"transform"`
		Reason    string                      `json:"reason"`
		Time      time.Time                   `json:"time"`
	}{
		Transform: msg,
		Reason:    reason,
		Time:      time.Now().UTC(),
	}
	data, err := json.Marshal(wrapper)
	if err != nil {
		return fmt.Errorf("marshal transform dlq: %w", err)
	}
	return q.rdb.LPush(ctx, transformDLQKey, data).Err()
}
