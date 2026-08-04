package resilience

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/sony/gobreaker"
)

// CircuitBreakerConfig holds circuit breaker settings.
type CircuitBreakerConfig struct {
	MaxFailures     uint32
	CooldownSeconds int
}

// DefaultCircuitBreakerConfig returns sensible defaults.
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxFailures:     5,
		CooldownSeconds: 30,
	}
}

// NewCircuitBreaker creates a named circuit breaker.
func NewCircuitBreaker(name string, cfg CircuitBreakerConfig) *gobreaker.CircuitBreaker {
	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: 1, // allow 1 request in half-open state
		Interval:    0, // don't clear counts periodically (use consecutive failures)
		Timeout:     time.Duration(cfg.CooldownSeconds) * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= cfg.MaxFailures
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			slog.Warn("circuit breaker state change",
				"name", name,
				"from", from.String(),
				"to", to.String(),
			)
		},
	}
	return gobreaker.NewCircuitBreaker(settings)
}

// Execute runs a function through the circuit breaker.
// Returns ErrCircuitOpen if the circuit is open.
func Execute[T any](cb *gobreaker.CircuitBreaker, fn func() (T, error)) (T, error) {
	result, err := cb.Execute(func() (any, error) {
		return fn()
	})
	if err != nil {
		var zero T
		if err == gobreaker.ErrOpenState || err == gobreaker.ErrTooManyRequests {
			return zero, fmt.Errorf("%s: %w", cb.Name(), ErrCircuitOpen)
		}
		return zero, err
	}
	return result.(T), nil
}

// ErrCircuitOpen indicates the circuit breaker is open (service unavailable).
var ErrCircuitOpen = fmt.Errorf("circuit open: service unavailable")
