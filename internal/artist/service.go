package artist

import (
	"context"
	"sync"
	"time"
)

// Service handles artist search with scoring and ranking.
type Service struct {
	client *SetlistFMClient
}

// NewService creates a new artist service.
func NewService(client *SetlistFMClient) *Service {
	return &Service{client: client}
}

// Search searches for artists and returns ranked results.
// Enrichment/scoring only runs on page 1.
func (s *Service) Search(ctx context.Context, query string, page int) (*ArtistSearchResult, error) {
	if query == "" {
		return nil, ErrEmptyQuery
	}
	if len(query) > 200 {
		return nil, ErrQueryTooLong
	}
	if page < 1 {
		return nil, ErrInvalidPage
	}

	result, err := s.client.SearchArtists(ctx, query, page)
	if err != nil {
		return nil, err
	}

	if page > 1 || len(result.Artists) == 0 {
		return result, nil
	}

	// Enrich and score top 5 candidates on page 1
	limit := 5
	if len(result.Artists) < limit {
		limit = len(result.Artists)
	}
	candidates := s.enrich(ctx, result.Artists[:limit])
	now := time.Now()
	ranked := rankCandidates(query, candidates, now)

	// Append any remaining (non-enriched) artists after the ranked ones
	if len(result.Artists) > limit {
		ranked = append(ranked, result.Artists[limit:]...)
	}
	result.Artists = ranked
	return result, nil
}

func (s *Service) enrich(ctx context.Context, artists []Artist) []scoredCandidate {
	candidates := make([]scoredCandidate, len(artists))
	var wg sync.WaitGroup

	for i, a := range artists {
		wg.Add(1)
		go func(idx int, art Artist) {
			defer wg.Done()
			data, _ := s.client.GetEnrichment(ctx, art.MBID)
			candidates[idx] = scoredCandidate{
				artist:       art,
				hasSetlists:  data.HasSetlists,
				setlistCount: data.SetlistCount,
				lastEvent:    data.LastEvent,
			}
		}(i, a)
	}
	wg.Wait()
	return candidates
}
