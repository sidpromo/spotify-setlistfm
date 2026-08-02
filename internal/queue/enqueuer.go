package queue

import "context"

// RedisEnqueuer implements orchestration.JobEnqueuer using Redis Streams.
type RedisEnqueuer struct {
	queue *Queue
}

// NewRedisEnqueuer creates an enqueuer backed by Redis Streams.
func NewRedisEnqueuer(queue *Queue) *RedisEnqueuer {
	return &RedisEnqueuer{queue: queue}
}

// Enqueue pushes a job to the Redis stream for worker processing.
func (e *RedisEnqueuer) Enqueue(ctx context.Context, jobID, userID, artistMBID, artistName string) error {
	_, err := e.queue.Enqueue(ctx, JobMessage{
		JobID:      jobID,
		UserID:     userID,
		ArtistMBID: artistMBID,
		ArtistName: artistName,
	})
	return err
}
