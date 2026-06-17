package setlist

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func makeSetlistJSON(id, date, tour string, songCount int) string {
	songs := ""
	for i := range songCount {
		if i > 0 {
			songs += ","
		}
		songs += fmt.Sprintf(`{"name":"Song%d"}`, i+1)
	}
	tourJSON := ""
	if tour != "" {
		tourJSON = fmt.Sprintf(`,"tour":{"name":"%s"}`, tour)
	}
	return fmt.Sprintf(`{"id":"%s","eventDate":"%s","artist":{"mbid":"aaa-bbb","name":"TestArtist"},"venue":{"name":"V","city":{"name":"C"}}%s,"sets":{"set":[{"song":[%s]}]}}`, id, date, tourJSON, songs)
}

func TestService_GetRecentSetlists_Success(t *testing.T) {
	now := time.Now()
	recentDate := now.AddDate(0, -1, 0).Format("02-01-2006")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"setlist":[%s,%s],"total":2,"page":1,"itemsPerPage":20}`,
			makeSetlistJSON("s1", recentDate, "Tour X", 8),
			makeSetlistJSON("s2", recentDate, "Tour X", 6),
		)))
	}))
	defer srv.Close()

	client := NewSetlistFMClient(srv.URL, "key", srv.Client())
	svc := NewService(client, DefaultConfig())

	result, err := svc.GetRecentSetlists(context.Background(), "65f4f0c5-ef9e-490c-aee3-909e7ae6b2ab")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Setlists) != 2 {
		t.Fatalf("expected 2, got %d", len(result.Setlists))
	}
	if result.TourName != "Tour X" {
		t.Errorf("expected 'Tour X', got %q", result.TourName)
	}
	if result.Artist != "TestArtist" {
		t.Errorf("expected 'TestArtist', got %q", result.Artist)
	}
}

func TestService_GetRecentSetlists_FiltersIncomplete(t *testing.T) {
	now := time.Now()
	recentDate := now.AddDate(0, -1, 0).Format("02-01-2006")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"setlist":[%s,%s,%s],"total":3,"page":1,"itemsPerPage":20}`,
			makeSetlistJSON("s1", recentDate, "", 8),
			makeSetlistJSON("s2", recentDate, "", 2), // too few songs
			makeSetlistJSON("s3", recentDate, "", 6),
		)))
	}))
	defer srv.Close()

	client := NewSetlistFMClient(srv.URL, "key", srv.Client())
	svc := NewService(client, DefaultConfig())

	result, err := svc.GetRecentSetlists(context.Background(), "65f4f0c5-ef9e-490c-aee3-909e7ae6b2ab")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Setlists) != 2 {
		t.Fatalf("expected 2 (filtered incomplete), got %d", len(result.Setlists))
	}
}

func TestService_GetRecentSetlists_NoRecentData(t *testing.T) {
	oldDate := time.Now().AddDate(-2, 0, 0).Format("02-01-2006")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"setlist":[%s],"total":1,"page":1,"itemsPerPage":20}`,
			makeSetlistJSON("s1", oldDate, "", 8),
		)))
	}))
	defer srv.Close()

	client := NewSetlistFMClient(srv.URL, "key", srv.Client())
	svc := NewService(client, DefaultConfig())

	_, err := svc.GetRecentSetlists(context.Background(), "65f4f0c5-ef9e-490c-aee3-909e7ae6b2ab")
	if err != ErrNoRecentSetlists {
		t.Fatalf("expected ErrNoRecentSetlists, got %v", err)
	}
}

func TestService_GetRecentSetlists_FallbackWindow(t *testing.T) {
	now := time.Now()
	// 4 months ago — outside 3-month primary, inside 6-month fallback
	fallbackDate := now.AddDate(0, -4, 0).Format("02-01-2006")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"setlist":[%s,%s,%s],"total":3,"page":1,"itemsPerPage":20}`,
			makeSetlistJSON("s1", fallbackDate, "", 8),
			makeSetlistJSON("s2", fallbackDate, "", 6),
			makeSetlistJSON("s3", fallbackDate, "", 7),
		)))
	}))
	defer srv.Close()

	client := NewSetlistFMClient(srv.URL, "key", srv.Client())
	svc := NewService(client, DefaultConfig())

	result, err := svc.GetRecentSetlists(context.Background(), "65f4f0c5-ef9e-490c-aee3-909e7ae6b2ab")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should use fallback since primary has <3 results
	if len(result.Setlists) != 3 {
		t.Fatalf("expected 3 (fallback window), got %d", len(result.Setlists))
	}
}
