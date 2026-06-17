package artist

import "errors"

var (
	ErrEmptyQuery    = errors.New("query parameter 'q' is required")
	ErrQueryTooLong  = errors.New("query parameter 'q' must be 200 characters or less")
	ErrInvalidPage   = errors.New("page must be >= 1")
	ErrProviderDown  = errors.New("setlist.fm service unavailable")
	ErrRateLimit     = errors.New("rate limit exceeded")
	ErrTimeout       = errors.New("upstream timeout")
)
