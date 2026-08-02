package orchestration

import "errors"

var (
	ErrJobNotFound   = errors.New("job not found")
	ErrMissingArtist = errors.New("artistMbid and artistName are required")
)
