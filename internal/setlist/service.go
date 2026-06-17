package setlist

import (
	"context"
	"regexp"
	"time"
)

// ServiceConfig holds configurable setlist service parameters.
type ServiceConfig struct {
	RecencyMonths  int
	FallbackMonths int
	MinSongs       int
	MaxFetch       int
}

// DefaultConfig returns default service configuration.
func DefaultConfig() ServiceConfig {
	return ServiceConfig{
		RecencyMonths:  3,
		FallbackMonths: 6,
		MinSongs:       5,
		MaxFetch:       10,
	}
}

// Service handles fetching and filtering setlists.
type Service struct {
	client *SetlistFMClient
	cfg    ServiceConfig
}

// NewService creates a new setlist service.
func NewService(client *SetlistFMClient, cfg ServiceConfig) *Service {
	return &Service{client: client, cfg: cfg}
}

var mbidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// GetRecentSetlists fetches, filters and returns recent setlists for an artist.
func (s *Service) GetRecentSetlists(ctx context.Context, mbid string) (*SetlistResult, error) {
	if !mbidRegex.MatchString(mbid) {
		return nil, ErrInvalidMBID
	}

	now := time.Now()
	primaryCutoff := now.AddDate(0, -s.cfg.RecencyMonths, 0)
	fallbackCutoff := now.AddDate(0, -s.cfg.FallbackMonths, 0)

	var allSetlists []Setlist
	var artistName string

	for page := 1; ; page++ {
		resp, err := s.client.FetchSetlists(ctx, mbid, page)
		if err != nil {
			return nil, err
		}

		if artistName == "" && len(resp.Setlist) > 0 {
			artistName = resp.Setlist[0].Artist.Name
		}

		pastFallback := false
		for _, sfm := range resp.Setlist {
			sl, err := parseSetlist(sfm)
			if err != nil {
				continue
			}
			if sl.EventDate.Before(fallbackCutoff) {
				pastFallback = true
				break
			}
			allSetlists = append(allSetlists, *sl)
		}

		if pastFallback || len(allSetlists) >= s.cfg.MaxFetch || resp.Page*resp.ItemsPerPage >= resp.Total {
			break
		}
	}

	// Filter incomplete setlists
	filtered := filterIncomplete(allSetlists, s.cfg.MinSongs)

	// Apply recency: prefer primary window, fall back if needed
	primary := filterByDate(filtered, primaryCutoff)
	if len(primary) >= 3 {
		filtered = primary
	}
	// else keep fallback window results

	if len(filtered) == 0 {
		return nil, ErrNoRecentSetlists
	}

	// Cap at MaxFetch
	if len(filtered) > s.cfg.MaxFetch {
		filtered = filtered[:s.cfg.MaxFetch]
	}

	tourName := detectTour(filtered)

	return &SetlistResult{
		Artist:   artistName,
		MBID:     mbid,
		TourName: tourName,
		Setlists: filtered,
	}, nil
}

func filterIncomplete(setlists []Setlist, minSongs int) []Setlist {
	var result []Setlist
	for _, s := range setlists {
		if len(s.Songs) >= minSongs {
			result = append(result, s)
		}
	}
	return result
}

func filterByDate(setlists []Setlist, cutoff time.Time) []Setlist {
	var result []Setlist
	for _, s := range setlists {
		if !s.EventDate.Before(cutoff) {
			result = append(result, s)
		}
	}
	return result
}

func detectTour(setlists []Setlist) string {
	if len(setlists) == 0 {
		return ""
	}
	name := setlists[0].TourName
	if name == "" {
		return ""
	}
	for _, s := range setlists[1:] {
		if s.TourName != name {
			return ""
		}
	}
	return name
}
