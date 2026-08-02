package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/sidpromo/spotify-setlistfm/internal/artist"
)

const artistSearchTTL = 1 * time.Hour

// CachedArtistService wraps the artist service with Redis caching.
type CachedArtistService struct {
	inner *artist.Service
	redis *Redis
}

// NewCachedArtistService creates a caching wrapper around the artist service.
func NewCachedArtistService(inner *artist.Service, redis *Redis) *CachedArtistService {
	return &CachedArtistService{inner: inner, redis: redis}
}

// Search checks cache first, falls back to the real service on miss.
func (s *CachedArtistService) Search(ctx context.Context, query string, page int) (*artist.ArtistSearchResult, error) {
	// Only cache page 1 (enriched/scored results)
	if page == 1 && s.redis != nil {
		key := artistSearchKey(query)
		var cached artist.ArtistSearchResult
		if s.redis.GetJSON(ctx, key, &cached) {
			return &cached, nil
		}
	}

	// Cache miss — call real service
	result, err := s.inner.Search(ctx, query, page)
	if err != nil {
		return nil, err
	}

	// Store in cache (only page 1)
	if page == 1 && s.redis != nil {
		key := artistSearchKey(query)
		s.redis.SetJSON(ctx, key, result, artistSearchTTL)
	}

	return result, nil
}

func artistSearchKey(query string) string {
	return fmt.Sprintf("artist:search:%s", query)
}
