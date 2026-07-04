package setlist

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// SetlistFMClient calls the setlist.fm API for setlist data.
type SetlistFMClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewSetlistFMClient creates a new setlist.fm client.
func NewSetlistFMClient(baseURL, apiKey string, httpClient *http.Client) *SetlistFMClient {
	return &SetlistFMClient{baseURL: baseURL, apiKey: apiKey, httpClient: httpClient}
}

// sfm response types
type sfmSetlistsResponse struct {
	Setlist      []sfmSetlist `json:"setlist"`
	Total        int          `json:"total"`
	Page         int          `json:"page"`
	ItemsPerPage int          `json:"itemsPerPage"`
}

type sfmSetlist struct {
	ID        string    `json:"id"`
	EventDate string    `json:"eventDate"`
	Artist    sfmArtist `json:"artist"`
	Venue     sfmVenue  `json:"venue"`
	Tour      *sfmTour  `json:"tour"`
	Sets      sfmSets   `json:"sets"`
}

type sfmArtist struct {
	MBID string `json:"mbid"`
	Name string `json:"name"`
}

type sfmVenue struct {
	Name string  `json:"name"`
	City sfmCity `json:"city"`
}

type sfmCity struct {
	Name string `json:"name"`
}

type sfmTour struct {
	Name string `json:"name"`
}

type sfmSets struct {
	Set []sfmSet `json:"set"`
}

type sfmSet struct {
	Name string    `json:"name"`
	Song []sfmSong `json:"song"`
}

type sfmSong struct {
	Name  string     `json:"name"`
	Cover *sfmArtist `json:"cover"`
}

// FetchSetlists fetches a page of setlists for an artist MBID.
func (c *SetlistFMClient) FetchSetlists(ctx context.Context, mbid string, page int) (*sfmSetlistsResponse, error) {
	u := fmt.Sprintf("%s/1.0/artist/%s/setlists?p=%d", c.baseURL, url.PathEscape(mbid), page)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil) // #nosec G704 -- baseURL is from config, not user input
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req) // #nosec G704
	if err != nil {
		if ctx.Err() != nil {
			return nil, ErrTimeout
		}
		return nil, ErrProviderDown
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusOK:
		// continue
	case resp.StatusCode == http.StatusTooManyRequests:
		return nil, ErrRateLimit
	case resp.StatusCode >= 500:
		return nil, ErrProviderDown
	default:
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var result sfmSetlistsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, ErrProviderDown
	}
	return &result, nil
}

// parseSetlist converts an sfm setlist to our domain model.
func parseSetlist(s sfmSetlist) (*Setlist, error) {
	t, err := time.Parse("02-01-2006", s.EventDate)
	if err != nil {
		return nil, err
	}

	var songs []Song
	for _, set := range s.Sets.Set {
		for _, song := range set.Song {
			if song.Name == "" {
				continue
			}
			songs = append(songs, Song{
				Name:    song.Name,
				IsCover: song.Cover != nil,
			})
		}
	}

	sl := &Setlist{
		ID:        s.ID,
		EventDate: t,
		Venue:     s.Venue.Name,
		City:      s.Venue.City.Name,
		Songs:     songs,
	}
	if s.Tour != nil {
		sl.TourName = s.Tour.Name
	}
	return sl, nil
}
