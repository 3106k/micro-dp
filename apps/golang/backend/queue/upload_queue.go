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
	uploadPrefix   = "micro-dp:uploads:"
	uploadIngestKey = uploadPrefix + "ingest"
	uploadDLQKey    = uploadPrefix + "dlq"
	uploadSeenPrefix = uploadPrefix + "seen:"
	uploadSeenTTL    = 24 * time.Hour
	uploadDequeueWait = 5 * time.Second
)

type UploadQueueImpl struct {
	rdb *redis.Client
}

func NewUploadQueue(client *ValkeyClient) *UploadQueueImpl {
	return &UploadQueueImpl{rdb: client.Client()}
}

func (q *UploadQueueImpl) Enqueue(ctx context.Context, msg *domain.UploadJobMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal upload job: %w", err)
	}
	return q.rdb.LPush(ctx, uploadIngestKey, data).Err()
}

func (q *UploadQueueImpl) Dequeue(ctx context.Context) (*domain.UploadJobMessage, error) {
	result, err := q.rdb.BRPop(ctx, uploadDequeueWait, uploadIngestKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("dequeue upload: %w", err)
	}

	var msg domain.UploadJobMessage
	if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
		return nil, fmt.Errorf("unmarshal upload job: %w", err)
	}
	return &msg, nil
}

func (q *UploadQueueImpl) MarkProcessed(ctx context.Context, uploadID string) error {
	key := uploadSeenPrefix + uploadID
	ok, err := q.rdb.SetArgs(ctx, key, "1", redis.SetArgs{
		Mode: "NX",
		TTL:  uploadSeenTTL,
	}).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("mark upload processed: %w", err)
	}
	if ok != "OK" {
		return domain.ErrUploadAlreadyProcessed
	}
	return nil
}

func (q *UploadQueueImpl) EnqueueDLQ(ctx context.Context, msg *domain.UploadJobMessage, reason string) error {
	wrapper := struct {
		Upload *domain.UploadJobMessage `json:"upload"`
		Reason string                   `json:"reason"`
		Time   time.Time                `json:"time"`
	}{
		Upload: msg,
		Reason: reason,
		Time:   time.Now().UTC(),
	}
	data, err := json.Marshal(wrapper)
	if err != nil {
		return fmt.Errorf("marshal upload dlq: %w", err)
	}
	return q.rdb.LPush(ctx, uploadDLQKey, data).Err()
}
