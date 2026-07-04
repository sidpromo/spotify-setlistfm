package artist

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// SetlistFMClient calls the setlist.fm API.
type SetlistFMClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewSetlistFMClient creates a new client for the setlist.fm API.
func NewSetlistFMClient(baseURL, apiKey string, httpClient *http.Client) *SetlistFMClient {
	return &SetlistFMClient{baseURL: baseURL, apiKey: apiKey, httpClient: httpClient}
}

// setlistfm API response types
type sfmArtistSearchResponse struct {
	Artist       []sfmArtist `json:"artist"`
	Total        int         `json:"total"`
	Page         int         `json:"page"`
	ItemsPerPage int         `json:"itemsPerPage"`
}

type sfmArtist struct {
	MBID           string `json:"mbid"`
	Name           string `json:"name"`
	SortName       string `json:"sortName"`
	Disambiguation string `json:"disambiguation"`
	URL            string `json:"url"`
}

type sfmSetlistsResponse struct {
	Setlist      []sfmSetlist `json:"setlist"`
	Total        int          `json:"total"`
	Page         int          `json:"page"`
	ItemsPerPage int          `json:"itemsPerPage"`
}

type sfmSetlist struct {
	EventDate string `json:"eventDate"`
}

// SearchArtists searches for artists by name.
func (c *SetlistFMClient) SearchArtists(ctx context.Context, name string, page int) (*ArtistSearchResult, error) {
	u := fmt.Sprintf("%s/1.0/search/artists?artistName=%s&p=%d", c.baseURL, url.QueryEscape(name), page)

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

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var sfm sfmArtistSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&sfm); err != nil {
		return nil, ErrProviderDown
	}

	result := &ArtistSearchResult{
		Total:        sfm.Total,
		Page:         sfm.Page,
		ItemsPerPage: sfm.ItemsPerPage,
	}
	for _, a := range sfm.Artist {
		result.Artists = append(result.Artists, Artist{
			MBID:           a.MBID,
			Name:           a.Name,
			SortName:       a.SortName,
			Disambiguation: a.Disambiguation,
			URL:            a.URL,
		})
	}
	return result, nil
}

// EnrichmentData holds metadata about an artist's setlist activity.
type EnrichmentData struct {
	HasSetlists  bool
	SetlistCount int
	LastEvent    time.Time
}

// GetEnrichment fetches page 1 of setlists for an artist to get metadata.
func (c *SetlistFMClient) GetEnrichment(ctx context.Context, mbid string) (*EnrichmentData, error) {
	u := fmt.Sprintf("%s/1.0/artist/%s/setlists?p=1", c.baseURL, url.PathEscape(mbid))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &EnrichmentData{}, nil // enrichment failure is not fatal
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &EnrichmentData{}, nil
	}

	var sfm sfmSetlistsResponse
	if err := json.NewDecoder(resp.Body).Decode(&sfm); err != nil {
		return &EnrichmentData{}, nil
	}

	data := &EnrichmentData{
		HasSetlists:  sfm.Total > 0,
		SetlistCount: sfm.Total,
	}
	if len(sfm.Setlist) > 0 {
		if t, err := time.Parse("02-01-2006", sfm.Setlist[0].EventDate); err == nil {
			data.LastEvent = t
		}
	}
	return data, nil
}

func checkResponse(resp *http.Response) error {
	switch {
	case resp.StatusCode == http.StatusOK:
		return nil
	case resp.StatusCode == http.StatusTooManyRequests:
		return ErrRateLimit
	case resp.StatusCode >= 500:
		return ErrProviderDown
	default:
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
}
