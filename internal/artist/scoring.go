package artist

import (
	"sort"
	"strings"
	"time"
)

// scoredCandidate holds an artist with enrichment data for scoring.
type scoredCandidate struct {
	artist       Artist
	hasSetlists  bool
	setlistCount int
	lastEvent    time.Time
}

func scoreNameMatch(query, name string) int {
	q := strings.ToLower(query)
	n := strings.ToLower(name)
	switch {
	case q == n:
		return 100
	case strings.HasPrefix(n, q):
		return 50
	case strings.Contains(n, q):
		return 20
	default:
		return 0
	}
}

func scoreSetlistData(hasSetlists bool, count int) int {
	if !hasSetlists {
		return 0
	}
	score := 30
	if count > 100 {
		score += 10
	} else if count > 20 {
		score += 5
	}
	return score
}

func scoreRecency(lastEvent, now time.Time) int {
	if lastEvent.IsZero() {
		return 0
	}
	months := now.Sub(lastEvent).Hours() / (24 * 30)
	switch {
	case months <= 3:
		return 60
	case months <= 12:
		return 40
	default:
		return 0
	}
}

func rankCandidates(query string, candidates []scoredCandidate, now time.Time) []Artist {
	type scored struct {
		artist Artist
		score  int
	}
	results := make([]scored, 0, len(candidates))
	for _, c := range candidates {
		s := scoreNameMatch(query, c.artist.Name) +
			scoreSetlistData(c.hasSetlists, c.setlistCount) +
			scoreRecency(c.lastEvent, now)
		results = append(results, scored{artist: c.artist, score: s})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})
	artists := make([]Artist, len(results))
	for i, r := range results {
		artists[i] = r.artist
	}
	return artists
}
