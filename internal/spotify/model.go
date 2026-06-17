package spotify

// SpotifyUser represents the authenticated user.
type SpotifyUser struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

// Track represents a matched Spotify track.
type Track struct {
	SpotifyID string `json:"spotifyId"`
	Name      string `json:"name"`
	Artist    string `json:"artist"`
	URI       string `json:"uri"`
}

// PlaylistResult is the result of creating a playlist.
type PlaylistResult struct {
	PlaylistID  string   `json:"playlistId"`
	PlaylistURL string   `json:"playlistUrl"`
	Name        string   `json:"name"`
	TracksAdded int      `json:"tracksAdded"`
	TracksTotal int      `json:"tracksTotal"`
	NotFound    []string `json:"notFound,omitempty"`
}

// PlaylistInput is the input for creating a playlist.
type PlaylistInput struct {
	ArtistName string
	TourName   string
	Songs      []string
}
