package queue

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// JobHandler is the function workers call to process a job.
type JobHandler func(ctx context.Context, msg *JobMessage) error

// WorkerConfig holds worker pool configuration.
type WorkerConfig struct {
	WorkerCount   int
	BlockTimeout  time.Duration
	ClaimInterval time.Duration
	ClaimMaxIdle  time.Duration
	MaxRetries    int
}

// DefaultWorkerConfig returns sensible defaults.
func DefaultWorkerConfig() WorkerConfig {
	return WorkerConfig{
		WorkerCount:   3,
		BlockTimeout:  5 * time.Second,
		ClaimInterval: 30 * time.Second,
		ClaimMaxIdle:  60 * time.Second,
		MaxRetries:    3,
	}
}

// WorkerPool manages a pool of workers consuming from the queue.
type WorkerPool struct {
	queue   *Queue
	handler JobHandler
	cfg     WorkerConfig
	wg      sync.WaitGroup
	cancel  context.CancelFunc
}

// NewWorkerPool creates a new worker pool.
func NewWorkerPool(queue *Queue, handler JobHandler, cfg WorkerConfig) *WorkerPool {
	return &WorkerPool{
		queue:   queue,
		handler: handler,
		cfg:     cfg,
	}
}

// Start launches all workers. Call Stop() to shut them down.
func (wp *WorkerPool) Start(ctx context.Context) {
	ctx, wp.cancel = context.WithCancel(ctx)

	for i := range wp.cfg.WorkerCount {
		wp.wg.Add(1)
		go wp.runWorker(ctx, fmt.Sprintf("worker-%d", i))
	}

	// Stale message claimer (picks up jobs from crashed workers)
	wp.wg.Add(1)
	go wp.runClaimer(ctx)

	slog.Info("worker pool started", "workers", wp.cfg.WorkerCount)
}

// Stop gracefully shuts down all workers.
func (wp *WorkerPool) Stop() {
	slog.Info("stopping worker pool")
	wp.cancel()
	wp.wg.Wait()
	slog.Info("worker pool stopped")
}

func (wp *WorkerPool) runWorker(ctx context.Context, name string) {
	defer wp.wg.Done()
	slog.Info("worker started", "name", name)

	for {
		select {
		case <-ctx.Done():
			slog.Info("worker shutting down", "name", name)
			return
		default:
		}

		streams, err := wp.queue.Read(ctx, name, 1, wp.cfg.BlockTimeout)
		if err != nil {
			if ctx.Err() != nil {
				return // shutdown
			}
			slog.Error("worker read error", "name", name, "error", err)
			time.Sleep(time.Second) // backoff on error
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				jobMsg, err := ParseMessage(msg)
				if err != nil {
					slog.Error("worker: bad message", "name", name, "id", msg.ID, "error", err)
					_ = wp.queue.Ack(ctx, msg.ID) // can't parse, dead letter
					continue
				}
				wp.handleMessage(ctx, name, msg.ID, jobMsg)
			}
		}
	}
}

func (wp *WorkerPool) handleMessage(ctx context.Context, workerName, messageID string, jobMsg *JobMessage) {
	slog.Debug("processing job", "worker", workerName, "jobId", jobMsg.JobID)

	if err := wp.handler(ctx, jobMsg); err != nil {
		slog.Error("job processing failed", "worker", workerName, "jobId", jobMsg.JobID, "error", err)
		// Don't ACK — message stays pending for retry via claimer
		return
	}

	// Success — acknowledge
	if err := wp.queue.Ack(ctx, messageID); err != nil {
		slog.Error("failed to ack message", "worker", workerName, "messageId", messageID, "error", err)
	}
	slog.Debug("job completed", "worker", workerName, "jobId", jobMsg.JobID)
}

func (wp *WorkerPool) runClaimer(ctx context.Context) {
	defer wp.wg.Done()
	ticker := time.NewTicker(wp.cfg.ClaimInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			messages, err := wp.queue.ClaimStale(ctx, "claimer", wp.cfg.ClaimMaxIdle, 10)
			if err != nil {
				slog.Error("claimer error", "error", err)
				continue
			}
			for _, msg := range messages {
				jobMsg, err := ParseMessage(msg)
				if err != nil {
					slog.Error("claimer: bad message", "id", msg.ID, "error", err)
					_ = wp.queue.Ack(ctx, msg.ID) // dead letter
					continue
				}
				slog.Warn("reclaiming stale job", "jobId", jobMsg.JobID, "messageId", msg.ID)
				wp.handleMessage(ctx, "claimer", msg.ID, jobMsg)
			}
		}
	}
}
