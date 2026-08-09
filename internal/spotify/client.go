package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

// Client interacts with the Spotify Web API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Spotify API client.
func NewClient(baseURL string, httpClient *http.Client) *Client {
	return &Client{baseURL: baseURL, httpClient: httpClient}
}

// SearchTrack searches for a track on Spotify.
func (c *Client) SearchTrack(ctx context.Context, accessToken, query string) (*Track, error) {
	u := fmt.Sprintf("%s/v1/search?q=%s&type=track&limit=1", c.baseURL, url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, ErrProviderDown
	}
	defer resp.Body.Close()

	if err := c.checkStatus(resp); err != nil {
		return nil, err
	}

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, ErrProviderDown
	}

	if len(result.Tracks.Items) == 0 {
		return nil, nil // not found
	}

	item := result.Tracks.Items[0]
	artistName := ""
	if len(item.Artists) > 0 {
		artistName = item.Artists[0].Name
	}
	return &Track{
		SpotifyID: item.ID,
		Name:      item.Name,
		Artist:    artistName,
		URI:       item.URI,
	}, nil
}

// GetCurrentUser gets the authenticated user's profile.
func (c *Client) GetCurrentUser(ctx context.Context, accessToken string) (*SpotifyUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, ErrProviderDown
	}
	defer resp.Body.Close()

	if err := c.checkStatus(resp); err != nil {
		return nil, err
	}

	var user spotifyMeResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, ErrProviderDown
	}
	return &SpotifyUser{ID: user.ID, DisplayName: user.DisplayName}, nil
}

// CreatePlaylist creates a new playlist for the user.
func (c *Client) CreatePlaylist(ctx context.Context, accessToken, userID, name, description string) (string, string, error) {
	u := fmt.Sprintf("%s/v1/me/playlists", c.baseURL)
	body := fmt.Sprintf(`{"name":%q,"description":%q,"public":false}`, name, description)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(body))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", ErrProviderDown
	}
	defer resp.Body.Close()

	if err := c.checkStatus(resp); err != nil {
		return "", "", err
	}

	var result createPlaylistResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", ErrProviderDown
	}
	return result.ID, result.ExternalURLs.Spotify, nil
}

// AddTracks adds tracks to a playlist.
func (c *Client) AddTracks(ctx context.Context, accessToken, playlistID string, uris []string) error {
	u := fmt.Sprintf("%s/v1/playlists/%s/items", c.baseURL, url.PathEscape(playlistID))
	urisJSON, _ := json.Marshal(uris)
	body := fmt.Sprintf(`{"uris":%s}`, urisJSON)

	slog.Debug("adding tracks to playlist", "playlistID", playlistID, "trackCount", len(uris), "url", u)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ErrProviderDown
	}
	defer resp.Body.Close()

	return c.checkStatus(resp)
}

func (c *Client) checkStatus(resp *http.Response) error {
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return nil
	case resp.StatusCode == http.StatusTooManyRequests:
		return ErrRateLimit
	case resp.StatusCode == http.StatusUnauthorized:
		return ErrTokenExpired
	default:
		body, _ := io.ReadAll(resp.Body)
		slog.Error("spotify API error", "status", resp.StatusCode, "body", string(body), "url", resp.Request.URL.String())
		return ErrProviderDown
	}
}

// Spotify API response types
type searchResponse struct {
	Tracks struct {
		Items []spotifyTrack `json:"items"`
	} `json:"tracks"`
}

type spotifyTrack struct {
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	URI     string          `json:"uri"`
	Artists []spotifyArtist `json:"artists"`
}

type spotifyArtist struct {
	Name string `json:"name"`
}

type spotifyMeResponse struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type createPlaylistResponse struct {
	ID           string `json:"id"`
	ExternalURLs struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
}
