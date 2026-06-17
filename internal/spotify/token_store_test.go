package spotify

import (
	"testing"
	"time"
)

func TestTokenStore_SaveAndGet(t *testing.T) {
	store := NewTokenStore()
	token := &Token{AccessToken: "abc", RefreshToken: "def", ExpiresAt: time.Now().Add(time.Hour)}

	if err := store.Save("sess1", token); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := store.Get("sess1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.AccessToken != "abc" {
		t.Errorf("expected 'abc', got %q", got.AccessToken)
	}
}

func TestTokenStore_GetNonExistent(t *testing.T) {
	store := NewTokenStore()
	_, err := store.Get("unknown")
	if err != ErrNotAuthenticated {
		t.Fatalf("expected ErrNotAuthenticated, got %v", err)
	}
}

func TestTokenStore_Delete(t *testing.T) {
	store := NewTokenStore()
	store.Save("sess1", &Token{AccessToken: "x"})
	store.Delete("sess1")

	_, err := store.Get("sess1")
	if err != ErrNotAuthenticated {
		t.Fatalf("expected ErrNotAuthenticated after delete, got %v", err)
	}
}

func TestToken_IsExpired(t *testing.T) {
	expired := &Token{ExpiresAt: time.Now().Add(-time.Hour)}
	if !expired.IsExpired() {
		t.Error("expected expired")
	}

	valid := &Token{ExpiresAt: time.Now().Add(time.Hour)}
	if valid.IsExpired() {
		t.Error("expected valid")
	}
}
