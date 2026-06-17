package httpx

import (
	"net/http"
	"time"
)

// NewClient creates an HTTP client with the given timeout.
func NewClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}
