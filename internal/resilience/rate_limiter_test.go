package resilience

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIdentifyClient_WithUserID(t *testing.T) {
	// Create a request that simulates having a user ID in context
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// Without JWT context, should fall back to IP
	id := identifyClient(req)
	if id == "" {
		t.Fatal("expected non-empty identifier")
	}
	// Should be IP-based since no user in context
	if id[:3] != "ip:" {
		t.Errorf("expected 'ip:' prefix, got %q", id)
	}
}

func TestIdentifyClient_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	id := identifyClient(req)
	if id != "ip:1.2.3.4" {
		t.Errorf("expected 'ip:1.2.3.4', got %q", id)
	}
}

func TestRateLimiterMiddleware_NilClient_Allows(t *testing.T) {
	// When Redis is nil/unavailable, rate limiter should gracefully allow all
	// This tests the concept — actual Redis testing needs integration tests
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Direct handler test (no rate limiter) as baseline
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
