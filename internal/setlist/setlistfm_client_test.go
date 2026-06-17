package setlist

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchSetlists_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1.0/artist/65f4f0c5-ef9e-490c-aee3-909e7ae6b2ab/setlists" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("x-api-key") != "key" {
			t.Error("missing api key")
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Error("missing accept header")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"setlist":[{"id":"s1","eventDate":"15-05-2026","artist":{"mbid":"65f4f0c5-ef9e-490c-aee3-909e7ae6b2ab","name":"Metallica"},"venue":{"name":"Wembley","city":{"name":"London"}},"tour":{"name":"M72"},"sets":{"set":[{"song":[{"name":"Song1"},{"name":"Song2"}]}]}}],"total":1,"page":1,"itemsPerPage":20}`))
	}))
	defer srv.Close()

	client := NewSetlistFMClient(srv.URL, "key", srv.Client())
	resp, err := client.FetchSetlists(context.Background(), "65f4f0c5-ef9e-490c-aee3-909e7ae6b2ab", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Setlist) != 1 {
		t.Fatalf("expected 1 setlist, got %d", len(resp.Setlist))
	}
	if resp.Setlist[0].Artist.Name != "Metallica" {
		t.Errorf("expected Metallica, got %q", resp.Setlist[0].Artist.Name)
	}
}

func TestFetchSetlists_RateLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	client := NewSetlistFMClient(srv.URL, "key", srv.Client())
	_, err := client.FetchSetlists(context.Background(), "abc", 1)
	if err != ErrRateLimit {
		t.Fatalf("expected ErrRateLimit, got %v", err)
	}
}

func TestParseSetlist(t *testing.T) {
	s := sfmSetlist{
		ID:        "s1",
		EventDate: "15-05-2026",
		Venue:     sfmVenue{Name: "Arena", City: sfmCity{Name: "Berlin"}},
		Tour:      &sfmTour{Name: "World Tour"},
		Sets: sfmSets{Set: []sfmSet{
			{Song: []sfmSong{{Name: "Song1"}, {Name: "Song2"}}},
			{Name: "Encore", Song: []sfmSong{{Name: "Song3", Cover: &sfmArtist{Name: "Other"}}}},
		}},
	}

	sl, err := parseSetlist(s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sl.TourName != "World Tour" {
		t.Errorf("expected 'World Tour', got %q", sl.TourName)
	}
	if len(sl.Songs) != 3 {
		t.Fatalf("expected 3 songs, got %d", len(sl.Songs))
	}
	if !sl.Songs[2].IsCover {
		t.Error("expected Song3 to be a cover")
	}
}

func TestParseSetlist_EmptySongs(t *testing.T) {
	s := sfmSetlist{
		ID:        "s1",
		EventDate: "15-05-2026",
		Sets: sfmSets{Set: []sfmSet{
			{Song: []sfmSong{{Name: ""}, {Name: "Real Song"}}},
		}},
	}

	sl, err := parseSetlist(s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sl.Songs) != 1 {
		t.Fatalf("expected 1 song (skip empty), got %d", len(sl.Songs))
	}
}

func TestFilterIncomplete(t *testing.T) {
	setlists := []Setlist{
		{ID: "a", Songs: make([]Song, 10)},
		{ID: "b", Songs: make([]Song, 3)},
		{ID: "c", Songs: make([]Song, 5)},
	}
	result := filterIncomplete(setlists, 5)
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
}

func TestFilterByDate(t *testing.T) {
	now := time.Now()
	setlists := []Setlist{
		{ID: "a", EventDate: now.AddDate(0, -1, 0)},
		{ID: "b", EventDate: now.AddDate(0, -5, 0)},
		{ID: "c", EventDate: now.AddDate(-1, 0, 0)},
	}
	cutoff := now.AddDate(0, -3, 0)
	result := filterByDate(setlists, cutoff)
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
	if result[0].ID != "a" {
		t.Errorf("expected 'a', got %q", result[0].ID)
	}
}

func TestDetectTour(t *testing.T) {
	tests := []struct {
		name     string
		setlists []Setlist
		expected string
	}{
		{"consistent", []Setlist{{TourName: "X"}, {TourName: "X"}}, "X"},
		{"mixed", []Setlist{{TourName: "X"}, {TourName: "Y"}}, ""},
		{"empty tour", []Setlist{{TourName: ""}, {TourName: ""}}, ""},
		{"empty list", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectTour(tt.setlists)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestService_InvalidMBID(t *testing.T) {
	svc := NewService(nil, DefaultConfig())
	_, err := svc.GetRecentSetlists(context.Background(), "not-a-uuid")
	if err != ErrInvalidMBID {
		t.Fatalf("expected ErrInvalidMBID, got %v", err)
	}
}
