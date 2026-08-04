package resilience

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sidpromo/spotify-setlistfm/internal/auth"
)

// RateLimiterConfig holds rate limit settings.
type RateLimiterConfig struct {
	RequestsPerMinute int
}

// RateLimiter implements token bucket rate limiting backed by Redis.
type RateLimiter struct {
	client *redis.Client
	cfg    RateLimiterConfig
	prefix string
}

// NewRateLimiter creates a new Redis-backed rate limiter.
func NewRateLimiter(client *redis.Client, prefix string, cfg RateLimiterConfig) *RateLimiter {
	return &RateLimiter{client: client, cfg: cfg, prefix: prefix}
}

// Allow checks if the request is within rate limits.
// Uses a simple sliding window counter in Redis.
func (rl *RateLimiter) Allow(ctx context.Context, identifier string) (bool, error) {
	key := fmt.Sprintf("ratelimit:%s:%s", rl.prefix, identifier)
	now := time.Now().Unix()
	windowStart := now - 60 // 1 minute window

	pipe := rl.client.Pipeline()

	// Remove old entries outside the window
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

	// Count current entries in window
	countCmd := pipe.ZCard(ctx, key)

	// Add current request
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: fmt.Sprintf("%d", now)})

	// Set expiry so keys don't accumulate forever
	pipe.Expire(ctx, key, 2*time.Minute)

	_, err := pipe.Exec(ctx)
	if err != nil {
		slog.Warn("rate limiter redis error, allowing request", "error", err)
		return true, nil // graceful degradation: if Redis is down, allow
	}

	count := countCmd.Val()
	return count < int64(rl.cfg.RequestsPerMinute), nil
}

// Middleware returns an HTTP middleware that rate-limits requests.
// Uses user ID from JWT context if available, otherwise client IP.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identifier := identifyClient(r)

		allowed, _ := rl.Allow(r.Context(), identifier)
		if !allowed {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]string{ // #nosec G104
				"error": "rate limit exceeded, try again later",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// identifyClient returns user ID (if authenticated) or IP address.
func identifyClient(r *http.Request) string {
	if userID := auth.UserIDFromContext(r.Context()); userID != "" {
		return "user:" + userID
	}
	// Fall back to IP
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	// Strip port
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return "ip:" + ip
}
