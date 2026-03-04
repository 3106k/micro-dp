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
	jobRunPrefix      = "micro-dp:job_runs:"
	jobRunIngestKey   = jobRunPrefix + "ingest"
	jobRunDLQKey      = jobRunPrefix + "dlq"
	jobRunSeenPrefix  = jobRunPrefix + "seen:"
	jobRunSeenTTL     = 24 * time.Hour
	jobRunDequeueWait = 5 * time.Second
)

type JobRunQueueImpl struct {
	rdb *redis.Client
}

func NewJobRunQueue(client *ValkeyClient) *JobRunQueueImpl {
	return &JobRunQueueImpl{rdb: client.Client()}
}

func (q *JobRunQueueImpl) Enqueue(ctx context.Context, msg *domain.JobRunMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal job run message: %w", err)
	}
	return q.rdb.LPush(ctx, jobRunIngestKey, data).Err()
}

func (q *JobRunQueueImpl) Dequeue(ctx context.Context) (*domain.JobRunMessage, error) {
	result, err := q.rdb.BRPop(ctx, jobRunDequeueWait, jobRunIngestKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("dequeue job run: %w", err)
	}

	var msg domain.JobRunMessage
	if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
		return nil, fmt.Errorf("unmarshal job run message: %w", err)
	}
	return &msg, nil
}

func (q *JobRunQueueImpl) MarkProcessed(ctx context.Context, jobRunID string) error {
	key := jobRunSeenPrefix + jobRunID
	ok, err := q.rdb.SetArgs(ctx, key, "1", redis.SetArgs{
		Mode: "NX",
		TTL:  jobRunSeenTTL,
	}).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("mark job run processed: %w", err)
	}
	if ok != "OK" {
		return domain.ErrJobRunAlreadyProcessed
	}
	return nil
}

func (q *JobRunQueueImpl) EnqueueDLQ(ctx context.Context, msg *domain.JobRunMessage, reason string) error {
	wrapper := struct {
		JobRun *domain.JobRunMessage `json:"job_run"`
		Reason string                `json:"reason"`
		Time   time.Time             `json:"time"`
	}{
		JobRun: msg,
		Reason: reason,
		Time:   time.Now().UTC(),
	}
	data, err := json.Marshal(wrapper)
	if err != nil {
		return fmt.Errorf("marshal job run dlq: %w", err)
	}
	return q.rdb.LPush(ctx, jobRunDLQKey, data).Err()
}
