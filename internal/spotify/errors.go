package spotify

import "errors"

var (
	ErrNotAuthenticated = errors.New("user not authenticated with Spotify")
	ErrTokenExpired     = errors.New("Spotify token expired, re-authentication required")
	ErrNoSongsMatched   = errors.New("no songs could be matched on Spotify")
	ErrProviderDown     = errors.New("Spotify service unavailable")
	ErrRateLimit        = errors.New("Spotify rate limit exceeded")
)
