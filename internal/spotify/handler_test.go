package spotify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sidpromo/spotify-setlistfm/internal/auth"
	"github.com/sidpromo/spotify-setlistfm/internal/database"
)

func TestAuthHandler_Login_Redirects(t *testing.T) {
	store := NewTokenStore()
	userRepo := database.NewInMemoryUserRepository()
	jwtSvc := auth.NewJWTService("test-secret", 15*time.Minute, 7*24*time.Hour)
	client := NewClient("http://unused", http.DefaultClient)

	ah := NewAuthHandler(AuthConfig{
		ClientID:    "cid",
		RedirectURI: "http://localhost:8080/v1/auth/spotify/callback",
	}, store, userRepo, jwtSvc, client, http.DefaultClient)

	mux := http.NewServeMux()
	RegisterHandlers(mux, ah)

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/spotify/login", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected 307, got %d", w.Code)
	}
	loc := w.Header().Get("Location")
	if loc == "" {
		t.Fatal("expected Location header")
	}
}

func TestAuthHandler_Callback_MissingCode(t *testing.T) {
	store := NewTokenStore()
	userRepo := database.NewInMemoryUserRepository()
	jwtSvc := auth.NewJWTService("test-secret", 15*time.Minute, 7*24*time.Hour)
	client := NewClient("http://unused", http.DefaultClient)

	ah := NewAuthHandler(AuthConfig{}, store, userRepo, jwtSvc, client, http.DefaultClient)

	mux := http.NewServeMux()
	RegisterHandlers(mux, ah)

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/spotify/callback", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAuthHandler_Status_Authenticated(t *testing.T) {
	store := NewTokenStore()
	store.Save("user-123", &Token{AccessToken: "x", ExpiresAt: time.Now().Add(time.Hour)})
	userRepo := database.NewInMemoryUserRepository()
	jwtSvc := auth.NewJWTService("test-secret", 15*time.Minute, 7*24*time.Hour)
	client := NewClient("http://unused", http.DefaultClient)

	ah := NewAuthHandler(AuthConfig{}, store, userRepo, jwtSvc, client, http.DefaultClient)

	mux := http.NewServeMux()
	RegisterHandlers(mux, ah)

	// Simulate authenticated request (JWT middleware would set this)
	req := httptest.NewRequest(http.MethodGet, "/v1/auth/spotify/status", nil)
	ctx := context.WithValue(req.Context(), auth.UserIDContextKey(), "user-123")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var resp map[string]bool
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp["authenticated"] {
		t.Error("expected authenticated true")
	}
}

func TestAuthHandler_Status_NotAuthenticated(t *testing.T) {
	store := NewTokenStore()
	userRepo := database.NewInMemoryUserRepository()
	jwtSvc := auth.NewJWTService("test-secret", 15*time.Minute, 7*24*time.Hour)
	client := NewClient("http://unused", http.DefaultClient)

	ah := NewAuthHandler(AuthConfig{}, store, userRepo, jwtSvc, client, http.DefaultClient)

	mux := http.NewServeMux()
	RegisterHandlers(mux, ah)

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/spotify/status", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var resp map[string]bool
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["authenticated"] {
		t.Error("expected authenticated false")
	}
}
