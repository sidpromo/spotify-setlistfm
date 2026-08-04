package orchestration

import "errors"

var (
	ErrJobNotFound   = errors.New("job not found")
	ErrMissingArtist = errors.New("artistMbid and artistName are required")
	ErrSystemBusy    = errors.New("system is at capacity, please try again later")
)
