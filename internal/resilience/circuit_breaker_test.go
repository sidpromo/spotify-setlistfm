package resilience

import (
	"errors"
	"testing"

	"github.com/sony/gobreaker"
)

func TestCircuitBreaker_TripsAfterFailures(t *testing.T) {
	cfg := CircuitBreakerConfig{MaxFailures: 3, CooldownSeconds: 30}
	cb := NewCircuitBreaker("test-service", cfg)

	// Simulate 3 failures
	for i := range 3 {
		_, _ = cb.Execute(func() (any, error) {
			return nil, errors.New("service down")
		})
		_ = i
	}

	// Next call should fail immediately (circuit open)
	_, err := cb.Execute(func() (any, error) {
		t.Fatal("should not be called when circuit is open")
		return nil, nil
	})

	if err == nil {
		t.Fatal("expected error when circuit is open")
	}
	if !errors.Is(err, gobreaker.ErrOpenState) {
		t.Fatalf("expected ErrOpenState, got %v", err)
	}
}

func TestCircuitBreaker_ClosedWhenSuccessful(t *testing.T) {
	cfg := CircuitBreakerConfig{MaxFailures: 3, CooldownSeconds: 30}
	cb := NewCircuitBreaker("test-service", cfg)

	// Successful calls keep circuit closed
	for range 10 {
		result, err := cb.Execute(func() (any, error) {
			return "ok", nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "ok" {
			t.Fatalf("expected 'ok', got %v", result)
		}
	}

	// Circuit should still be closed
	if cb.State() != gobreaker.StateClosed {
		t.Errorf("expected closed, got %s", cb.State())
	}
}

func TestCircuitBreaker_FailuresResetOnSuccess(t *testing.T) {
	cfg := CircuitBreakerConfig{MaxFailures: 3, CooldownSeconds: 30}
	cb := NewCircuitBreaker("test-service", cfg)

	// 2 failures (not enough to trip)
	for range 2 {
		cb.Execute(func() (any, error) {
			return nil, errors.New("fail")
		})
	}

	// 1 success resets consecutive failures
	cb.Execute(func() (any, error) {
		return "ok", nil
	})

	// 2 more failures (still not enough to trip — counter was reset)
	for range 2 {
		cb.Execute(func() (any, error) {
			return nil, errors.New("fail")
		})
	}

	// Circuit should still be closed
	if cb.State() != gobreaker.StateClosed {
		t.Errorf("expected closed after reset, got %s", cb.State())
	}
}

func TestExecuteGeneric_Success(t *testing.T) {
	cfg := DefaultCircuitBreakerConfig()
	cb := NewCircuitBreaker("test", cfg)

	result, err := Execute(cb, func() (string, error) {
		return "hello", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestExecuteGeneric_CircuitOpen(t *testing.T) {
	cfg := CircuitBreakerConfig{MaxFailures: 1, CooldownSeconds: 60}
	cb := NewCircuitBreaker("test", cfg)

	// Trip the circuit
	Execute(cb, func() (string, error) {
		return "", errors.New("fail")
	})

	// Should get circuit open error
	_, err := Execute(cb, func() (string, error) {
		return "should not reach", nil
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
}
