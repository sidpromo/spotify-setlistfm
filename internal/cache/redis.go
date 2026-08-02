package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis wraps go-redis with graceful degradation.
// If Redis is unavailable, operations return cache misses instead of errors.
type Redis struct {
	client *redis.Client
}

// Connect creates a new Redis connection.
func Connect(redisURL string) (*Redis, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	slog.Info("redis connected")
	return &Redis{client: client}, nil
}

// Close closes the Redis connection.
func (r *Redis) Close() error {
	return r.client.Close()
}

// Ping checks Redis connectivity.
func (r *Redis) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Get retrieves a cached value. Returns nil if key doesn't exist or Redis is down.
func (r *Redis) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // cache miss
	}
	if err != nil {
		slog.Warn("redis get failed, treating as miss", "key", key, "error", err)
		return nil, nil // graceful degradation
	}
	return val, nil
}

// Set stores a value with TTL. Silently fails if Redis is down.
func (r *Redis) Set(ctx context.Context, key string, value []byte, ttl time.Duration) {
	if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
		slog.Warn("redis set failed", "key", key, "error", err)
	}
}

// Delete removes a key. Silently fails if Redis is down.
func (r *Redis) Delete(ctx context.Context, key string) {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		slog.Warn("redis delete failed", "key", key, "error", err)
	}
}

// GetJSON retrieves and unmarshals a cached JSON value.
func (r *Redis) GetJSON(ctx context.Context, key string, dest any) bool {
	data, _ := r.Get(ctx, key)
	if data == nil {
		return false // miss
	}
	if err := json.Unmarshal(data, dest); err != nil {
		slog.Warn("redis unmarshal failed", "key", key, "error", err)
		return false
	}
	return true // hit
}

// SetJSON marshals and caches a value as JSON.
func (r *Redis) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) {
	data, err := json.Marshal(value)
	if err != nil {
		slog.Warn("redis marshal failed", "key", key, "error", err)
		return
	}
	r.Set(ctx, key, data, ttl)
}
