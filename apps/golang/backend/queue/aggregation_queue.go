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
	aggKeyPrefix  = "micro-dp:aggregations:"
	aggIngestKey  = aggKeyPrefix + "ingest"
	aggDLQKey     = aggKeyPrefix + "dlq"
	aggSeenPrefix = aggKeyPrefix + "seen:"
	aggSeenTTL    = 24 * time.Hour
	aggDequeueWait = 5 * time.Second
)

type AggregationQueueImpl struct {
	rdb *redis.Client
}

func NewAggregationQueue(client *ValkeyClient) *AggregationQueueImpl {
	return &AggregationQueueImpl{rdb: client.Client()}
}

func (q *AggregationQueueImpl) Enqueue(ctx context.Context, msg domain.AggregationMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal aggregation message: %w", err)
	}
	return q.rdb.LPush(ctx, aggIngestKey, data).Err()
}

func (q *AggregationQueueImpl) Dequeue(ctx context.Context) (*domain.AggregationMessage, error) {
	result, err := q.rdb.BRPop(ctx, aggDequeueWait, aggIngestKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("dequeue aggregation: %w", err)
	}

	var msg domain.AggregationMessage
	if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
		return nil, fmt.Errorf("unmarshal aggregation message: %w", err)
	}
	return &msg, nil
}

func (q *AggregationQueueImpl) MarkProcessed(ctx context.Context, tenantID, date string) error {
	key := aggSeenPrefix + tenantID + ":" + date
	ok, err := q.rdb.SetArgs(ctx, key, "1", redis.SetArgs{
		Mode: "NX",
		TTL:  aggSeenTTL,
	}).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("mark aggregation processed: %w", err)
	}
	if ok != "OK" {
		return domain.ErrAggregationAlreadyProcessed
	}
	return nil
}

func (q *AggregationQueueImpl) ResetProcessed(ctx context.Context, tenantID, date string) error {
	key := aggSeenPrefix + tenantID + ":" + date
	return q.rdb.Del(ctx, key).Err()
}

func (q *AggregationQueueImpl) IsProcessed(ctx context.Context, tenantID, date string) (bool, error) {
	key := aggSeenPrefix + tenantID + ":" + date
	n, err := q.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("check aggregation processed: %w", err)
	}
	return n > 0, nil
}

func (q *AggregationQueueImpl) EnqueueDLQ(ctx context.Context, msg domain.AggregationMessage, reason string) error {
	wrapper := struct {
		Message domain.AggregationMessage `json:"message"`
		Reason  string                    `json:"reason"`
		Time    time.Time                 `json:"time"`
	}{
		Message: msg,
		Reason:  reason,
		Time:    time.Now().UTC(),
	}
	data, err := json.Marshal(wrapper)
	if err != nil {
		return fmt.Errorf("marshal aggregation dlq: %w", err)
	}
	return q.rdb.LPush(ctx, aggDLQKey, data).Err()
}
