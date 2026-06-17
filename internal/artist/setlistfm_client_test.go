package artist

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchArtists_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1.0/search/artists" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("artistName") != "Metallica" {
			t.Errorf("unexpected artistName: %s", r.URL.Query().Get("artistName"))
		}
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("missing api key header")
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("missing accept header")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"artist":[{"mbid":"abc","name":"Metallica","sortName":"Metallica","disambiguation":"","url":"http://example.com"}],"total":1,"page":1,"itemsPerPage":20}`))
	}))
	defer srv.Close()

	client := NewSetlistFMClient(srv.URL, "test-key", srv.Client())
	result, err := client.SearchArtists(context.Background(), "Metallica", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Artists) != 1 {
		t.Fatalf("expected 1 artist, got %d", len(result.Artists))
	}
	if result.Artists[0].MBID != "abc" {
		t.Errorf("expected mbid 'abc', got %q", result.Artists[0].MBID)
	}
}

func TestSearchArtists_RateLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	client := NewSetlistFMClient(srv.URL, "test-key", srv.Client())
	_, err := client.SearchArtists(context.Background(), "Metallica", 1)
	if err != ErrRateLimit {
		t.Fatalf("expected ErrRateLimit, got %v", err)
	}
}

func TestSearchArtists_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := NewSetlistFMClient(srv.URL, "test-key", srv.Client())
	_, err := client.SearchArtists(context.Background(), "Metallica", 1)
	if err != ErrProviderDown {
		t.Fatalf("expected ErrProviderDown, got %v", err)
	}
}

func TestSearchArtists_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{broken`))
	}))
	defer srv.Close()

	client := NewSetlistFMClient(srv.URL, "test-key", srv.Client())
	_, err := client.SearchArtists(context.Background(), "Metallica", 1)
	if err != ErrProviderDown {
		t.Fatalf("expected ErrProviderDown, got %v", err)
	}
}

func TestGetEnrichment_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1.0/artist/abc/setlists" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"setlist":[{"eventDate":"15-05-2026"}],"total":42,"page":1,"itemsPerPage":20}`))
	}))
	defer srv.Close()

	client := NewSetlistFMClient(srv.URL, "test-key", srv.Client())
	data, err := client.GetEnrichment(context.Background(), "abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !data.HasSetlists {
		t.Error("expected HasSetlists true")
	}
	if data.SetlistCount != 42 {
		t.Errorf("expected count 42, got %d", data.SetlistCount)
	}
	if data.LastEvent.IsZero() {
		t.Error("expected non-zero last event")
	}
}

func TestGetEnrichment_Failure_NotFatal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := NewSetlistFMClient(srv.URL, "test-key", srv.Client())
	data, err := client.GetEnrichment(context.Background(), "abc")
	if err != nil {
		t.Fatalf("enrichment failure should not error: %v", err)
	}
	if data.HasSetlists {
		t.Error("expected HasSetlists false on failure")
	}
}
