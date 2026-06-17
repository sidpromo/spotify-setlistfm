package spotify

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_SearchTrack_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/search" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer token123" {
			t.Error("missing or wrong auth header")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"tracks":{"items":[{"id":"t1","name":"Enter Sandman","uri":"spotify:track:t1","artists":[{"name":"Metallica"}]}]}}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	track, err := client.SearchTrack(context.Background(), "token123", "track:Enter Sandman artist:Metallica")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if track == nil {
		t.Fatal("expected track, got nil")
	}
	if track.URI != "spotify:track:t1" {
		t.Errorf("expected URI 'spotify:track:t1', got %q", track.URI)
	}
}

func TestClient_SearchTrack_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"tracks":{"items":[]}}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	track, err := client.SearchTrack(context.Background(), "token", "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if track != nil {
		t.Errorf("expected nil track, got %+v", track)
	}
}

func TestClient_SearchTrack_RateLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	_, err := client.SearchTrack(context.Background(), "token", "query")
	if err != ErrRateLimit {
		t.Fatalf("expected ErrRateLimit, got %v", err)
	}
}

func TestClient_SearchTrack_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	_, err := client.SearchTrack(context.Background(), "bad-token", "query")
	if err != ErrTokenExpired {
		t.Fatalf("expected ErrTokenExpired, got %v", err)
	}
}

func TestClient_CreatePlaylist_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/users/user1/playlists" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"pl1","external_urls":{"spotify":"https://open.spotify.com/playlist/pl1"}}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	id, url, err := client.CreatePlaylist(context.Background(), "token", "user1", "My Playlist", "desc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "pl1" {
		t.Errorf("expected 'pl1', got %q", id)
	}
	if url != "https://open.spotify.com/playlist/pl1" {
		t.Errorf("unexpected url: %q", url)
	}
}

func TestClient_AddTracks_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/playlists/pl1/tracks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"snapshot_id":"snap1"}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	err := client.AddTracks(context.Background(), "token", "pl1", []string{"spotify:track:t1", "spotify:track:t2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetCurrentUser_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"user1","display_name":"Test User"}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	user, err := client.GetCurrentUser(context.Background(), "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != "user1" {
		t.Errorf("expected 'user1', got %q", user.ID)
	}
}
