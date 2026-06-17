package spotify

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestService_CreatePlaylist_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v1/me":
			w.Write([]byte(`{"id":"user1","display_name":"Test"}`))
		case r.URL.Path == "/v1/search":
			w.Write([]byte(`{"tracks":{"items":[{"id":"t1","name":"Song","uri":"spotify:track:t1","artists":[{"name":"A"}]}]}}`))
		case r.URL.Path == "/v1/users/user1/playlists":
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id":"pl1","external_urls":{"spotify":"https://open.spotify.com/playlist/pl1"}}`))
		case r.URL.Path == "/v1/playlists/pl1/tracks":
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{}`))
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	store := NewTokenStore()
	store.Save("sess1", &Token{AccessToken: "tok", ExpiresAt: time.Now().Add(time.Hour)})

	svc := NewService(client, store)
	result, err := svc.CreatePlaylist(context.Background(), "sess1", PlaylistInput{
		ArtistName: "Metallica",
		TourName:   "M72",
		Songs:      []string{"Enter Sandman", "Fuel"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TracksAdded != 2 {
		t.Errorf("expected 2 tracks added, got %d", result.TracksAdded)
	}
	if result.Name != "Metallica - M72 (Predicted Setlist)" {
		t.Errorf("unexpected name: %q", result.Name)
	}
}

func TestService_CreatePlaylist_NotAuthenticated(t *testing.T) {
	store := NewTokenStore()
	svc := NewService(nil, store)

	_, err := svc.CreatePlaylist(context.Background(), "unknown", PlaylistInput{Songs: []string{"X"}})
	if err != ErrNotAuthenticated {
		t.Fatalf("expected ErrNotAuthenticated, got %v", err)
	}
}

func TestService_CreatePlaylist_TokenExpired(t *testing.T) {
	store := NewTokenStore()
	store.Save("sess1", &Token{AccessToken: "old", ExpiresAt: time.Now().Add(-time.Hour)})

	svc := NewService(nil, store)
	_, err := svc.CreatePlaylist(context.Background(), "sess1", PlaylistInput{Songs: []string{"X"}})
	if err != ErrTokenExpired {
		t.Fatalf("expected ErrTokenExpired, got %v", err)
	}
}

func TestService_CreatePlaylist_NoSongsMatched(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v1/me":
			w.Write([]byte(`{"id":"user1","display_name":"Test"}`))
		case r.URL.Path == "/v1/search":
			w.Write([]byte(`{"tracks":{"items":[]}}`)) // no results
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	store := NewTokenStore()
	store.Save("sess1", &Token{AccessToken: "tok", ExpiresAt: time.Now().Add(time.Hour)})

	svc := NewService(client, store)
	_, err := svc.CreatePlaylist(context.Background(), "sess1", PlaylistInput{
		ArtistName: "Unknown",
		Songs:      []string{"Nonexistent Song"},
	})
	if err != ErrNoSongsMatched {
		t.Fatalf("expected ErrNoSongsMatched, got %v", err)
	}
}

func TestService_CreatePlaylist_PartialMatch(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v1/me":
			w.Write([]byte(`{"id":"user1","display_name":"Test"}`))
		case r.URL.Path == "/v1/search":
			callCount++
			// First 2 calls (strict+relaxed for song 1) return nothing
			// Next 2 calls (strict+relaxed for song 2) return a result
			if callCount <= 2 {
				w.Write([]byte(`{"tracks":{"items":[]}}`))
			} else {
				w.Write([]byte(`{"tracks":{"items":[{"id":"t1","name":"Found","uri":"spotify:track:t1","artists":[{"name":"A"}]}]}}`))
			}
		case r.URL.Path == "/v1/users/user1/playlists":
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id":"pl1","external_urls":{"spotify":"https://example.com/pl1"}}`))
		case r.URL.Path == "/v1/playlists/pl1/tracks":
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{}`))
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	store := NewTokenStore()
	store.Save("sess1", &Token{AccessToken: "tok", ExpiresAt: time.Now().Add(time.Hour)})

	svc := NewService(client, store)
	result, err := svc.CreatePlaylist(context.Background(), "sess1", PlaylistInput{
		ArtistName: "Artist",
		Songs:      []string{"Missing Song", "Found Song"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TracksAdded != 1 {
		t.Errorf("expected 1 track, got %d", result.TracksAdded)
	}
	if len(result.NotFound) != 1 || result.NotFound[0] != "Missing Song" {
		t.Errorf("expected NotFound ['Missing Song'], got %v", result.NotFound)
	}
}

func TestPlaylistName(t *testing.T) {
	if got := playlistName("Band", "Tour"); got != "Band - Tour (Predicted Setlist)" {
		t.Errorf("unexpected: %q", got)
	}
	if got := playlistName("Band", ""); got != "Band - Predicted Setlist" {
		t.Errorf("unexpected: %q", got)
	}
}
