package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	StreamName    = "jobs:pipeline"
	GroupName     = "workers"
	DefaultMaxLen = 10000 // max stream length (old entries trimmed)
)

// JobMessage is the payload pushed to the queue.
type JobMessage struct {
	JobID    string `json:"jobId"`
	UserID   string `json:"userId"`
	ArtistMBID string `json:"artistMbid"`
	ArtistName string `json:"artistName"`
}

// Queue handles publishing and consuming from Redis Streams.
type Queue struct {
	client *redis.Client
}

// New creates a new Redis Streams queue and ensures the consumer group exists.
func New(client *redis.Client) (*Queue, error) {
	q := &Queue{client: client}

	// Create consumer group (idempotent — ignore error if already exists)
	ctx := context.Background()
	err := client.XGroupCreateMkStream(ctx, StreamName, GroupName, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return nil, fmt.Errorf("create consumer group: %w", err)
	}

	slog.Info("job queue ready", "stream", StreamName, "group", GroupName)
	return q, nil
}

// Enqueue pushes a job message to the stream.
func (q *Queue) Enqueue(ctx context.Context, msg JobMessage) (string, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("marshal job message: %w", err)
	}

	id, err := q.client.XAdd(ctx, &redis.XAddArgs{
		Stream: StreamName,
		MaxLen: DefaultMaxLen,
		Approx: true,
		Values: map[string]any{"data": string(data)},
	}).Result()
	if err != nil {
		return "", fmt.Errorf("enqueue job: %w", err)
	}

	slog.Debug("job enqueued", "jobId", msg.JobID, "streamId", id)
	return id, nil
}

// Read reads pending messages for a consumer. Blocks until a message is available or timeout.
func (q *Queue) Read(ctx context.Context, consumerName string, count int64, blockTimeout time.Duration) ([]redis.XStream, error) {
	streams, err := q.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    GroupName,
		Consumer: consumerName,
		Streams:  []string{StreamName, ">"},
		Count:    count,
		Block:    blockTimeout,
	}).Result()
	if err == redis.Nil {
		return nil, nil // timeout, no messages
	}
	if err != nil {
		return nil, fmt.Errorf("read from stream: %w", err)
	}
	return streams, nil
}

// Ack acknowledges a message (marks it as processed).
func (q *Queue) Ack(ctx context.Context, messageID string) error {
	return q.client.XAck(ctx, StreamName, GroupName, messageID).Err()
}

// ClaimStale reclaims messages that have been pending longer than maxIdle.
// Used by workers to pick up jobs abandoned by crashed workers.
func (q *Queue) ClaimStale(ctx context.Context, consumerName string, maxIdle time.Duration, count int64) ([]redis.XMessage, error) {
	// First, find pending messages older than maxIdle
	pending, err := q.client.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: StreamName,
		Group:  GroupName,
		Start:  "-",
		End:    "+",
		Count:  count,
		Idle:   maxIdle,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("check pending: %w", err)
	}

	if len(pending) == 0 {
		return nil, nil
	}

	// Collect message IDs to claim
	ids := make([]string, len(pending))
	for i, p := range pending {
		ids[i] = p.ID
	}

	// Claim them for this consumer
	messages, err := q.client.XClaim(ctx, &redis.XClaimArgs{
		Stream:   StreamName,
		Group:    GroupName,
		Consumer: consumerName,
		MinIdle:  maxIdle,
		Messages: ids,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("claim stale messages: %w", err)
	}

	return messages, nil
}

// ParseMessage extracts a JobMessage from a Redis stream message.
func ParseMessage(msg redis.XMessage) (*JobMessage, error) {
	data, ok := msg.Values["data"].(string)
	if !ok {
		return nil, fmt.Errorf("message missing 'data' field")
	}
	var job JobMessage
	if err := json.Unmarshal([]byte(data), &job); err != nil {
		return nil, fmt.Errorf("unmarshal job message: %w", err)
	}
	return &job, nil
}

// Len returns the current number of messages in the stream.
func (q *Queue) Len(ctx context.Context) (int64, error) {
	return q.client.XLen(ctx, StreamName).Result()
}
