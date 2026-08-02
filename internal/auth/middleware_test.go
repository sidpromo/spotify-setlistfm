package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMiddleware_ValidToken(t *testing.T) {
	svc := NewJWTService("test-secret", 15*time.Minute, 7*24*time.Hour)
	token, _ := svc.GenerateAccessToken("user-123")

	var gotUserID string
	handler := Middleware(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if gotUserID != "user-123" {
		t.Errorf("expected 'user-123', got %q", gotUserID)
	}
}

func TestMiddleware_MissingHeader(t *testing.T) {
	svc := NewJWTService("test-secret", 15*time.Minute, 7*24*time.Hour)

	handler := Middleware(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestMiddleware_InvalidFormat(t *testing.T) {
	svc := NewJWTService("test-secret", 15*time.Minute, 7*24*time.Hour)

	handler := Middleware(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "NotBearer token123")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestMiddleware_ExpiredToken(t *testing.T) {
	svc := NewJWTService("test-secret", -1*time.Second, 7*24*time.Hour)
	token, _ := svc.GenerateAccessToken("user-123")

	handler := Middleware(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestMiddleware_WrongSecret(t *testing.T) {
	svc1 := NewJWTService("secret-1", 15*time.Minute, 7*24*time.Hour)
	svc2 := NewJWTService("secret-2", 15*time.Minute, 7*24*time.Hour)
	token, _ := svc1.GenerateAccessToken("user-123")

	handler := Middleware(svc2)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestUserIDFromContext_Empty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	got := UserIDFromContext(req.Context())
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
