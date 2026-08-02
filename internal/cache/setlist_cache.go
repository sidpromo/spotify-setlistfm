package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/sidpromo/spotify-setlistfm/internal/setlist"
)

const setlistTTL = 24 * time.Hour

// CachedSetlistService wraps the setlist service with Redis caching.
type CachedSetlistService struct {
	inner *setlist.Service
	redis *Redis
}

// NewCachedSetlistService creates a caching wrapper around the setlist service.
func NewCachedSetlistService(inner *setlist.Service, redis *Redis) *CachedSetlistService {
	return &CachedSetlistService{inner: inner, redis: redis}
}

// GetRecentSetlists checks cache first, falls back to the real service on miss.
func (s *CachedSetlistService) GetRecentSetlists(ctx context.Context, mbid string) (*setlist.SetlistResult, error) {
	if s.redis != nil {
		key := setlistKey(mbid)
		var cached setlist.SetlistResult
		if s.redis.GetJSON(ctx, key, &cached) {
			return &cached, nil
		}
	}

	// Cache miss — call real service
	result, err := s.inner.GetRecentSetlists(ctx, mbid)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if s.redis != nil {
		key := setlistKey(mbid)
		s.redis.SetJSON(ctx, key, result, setlistTTL)
	}

	return result, nil
}

func setlistKey(mbid string) string {
	return fmt.Sprintf("setlist:recent:%s", mbid)
}
