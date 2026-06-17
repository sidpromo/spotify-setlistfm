package artist

import (
	"testing"
	"time"
)

func TestScoreNameMatch(t *testing.T) {
	tests := []struct {
		query    string
		name     string
		expected int
	}{
		{"metallica", "Metallica", 100},
		{"METALLICA", "Metallica", 100},
		{"metal", "Metallica", 50},
		{"tall", "Metallica", 20},
		{"xyz", "Metallica", 0},
	}
	for _, tt := range tests {
		got := scoreNameMatch(tt.query, tt.name)
		if got != tt.expected {
			t.Errorf("scoreNameMatch(%q, %q) = %d, want %d", tt.query, tt.name, got, tt.expected)
		}
	}
}

func TestScoreSetlistData(t *testing.T) {
	tests := []struct {
		hasSetlists bool
		count       int
		expected    int
	}{
		{false, 0, 0},
		{true, 3, 30},
		{true, 25, 35},
		{true, 150, 40},
	}
	for _, tt := range tests {
		got := scoreSetlistData(tt.hasSetlists, tt.count)
		if got != tt.expected {
			t.Errorf("scoreSetlistData(%v, %d) = %d, want %d", tt.hasSetlists, tt.count, got, tt.expected)
		}
	}
}

func TestScoreRecency(t *testing.T) {
	now := time.Now()
	tests := []struct {
		lastEvent time.Time
		expected  int
	}{
		{now.AddDate(0, -1, 0), 60},  // 1 month ago → within 3 months
		{now.AddDate(0, -2, 0), 60},  // 2 months ago → within 3 months
		{now.AddDate(0, -6, 0), 40},  // 6 months ago → within 12 months
		{now.AddDate(-2, 0, 0), 0},   // 2 years ago → no recency
		{time.Time{}, 0},             // zero time → no recency
	}
	for _, tt := range tests {
		got := scoreRecency(tt.lastEvent, now)
		if got != tt.expected {
			t.Errorf("scoreRecency(%v) = %d, want %d", tt.lastEvent, got, tt.expected)
		}
	}
}

func TestRankArtists(t *testing.T) {
	now := time.Now()
	candidates := []scoredCandidate{
		{artist: Artist{Name: "Genesis", MBID: "1"}, hasSetlists: true, setlistCount: 500, lastEvent: now.AddDate(0, -8, 0)},
		{artist: Artist{Name: "Genesis", MBID: "2"}, hasSetlists: true, setlistCount: 3, lastEvent: now.AddDate(0, -1, 0)},
		{artist: Artist{Name: "Genesis Project", MBID: "3"}, hasSetlists: false},
	}

	ranked := rankCandidates("genesis", candidates, now)

	if len(ranked) != 3 {
		t.Fatalf("expected 3 results, got %d", len(ranked))
	}
	// The one touring now (MBID "2") should be first
	if ranked[0].MBID != "2" {
		t.Errorf("expected MBID '2' first, got %q", ranked[0].MBID)
	}
	// The one with lots of setlists but not recent (MBID "1") should be second
	if ranked[1].MBID != "1" {
		t.Errorf("expected MBID '1' second, got %q", ranked[1].MBID)
	}
}
