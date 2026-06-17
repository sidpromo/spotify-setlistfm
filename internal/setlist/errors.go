package setlist

import "errors"

var (
	ErrInvalidMBID     = errors.New("invalid MBID format")
	ErrNoRecentSetlists = errors.New("no recent setlists found for this artist")
	ErrProviderDown    = errors.New("setlist.fm service unavailable")
	ErrRateLimit       = errors.New("rate limit exceeded")
	ErrTimeout         = errors.New("upstream timeout")
)
