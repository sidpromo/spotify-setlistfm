package prediction

import (
	"testing"
)

func TestOrderSongs_OpenerLock(t *testing.T) {
	cfg := DefaultConfig()
	// "Opener" appears in position 0 in all setlists (100% > 70% threshold)
	input := scoringInput{
		setlists: []simplifiedSetlist{
			{songs: []string{"Opener", "B", "C"}},
			{songs: []string{"Opener", "B", "C"}},
			{songs: []string{"Opener", "B", "C"}},
		},
		decay: cfg.DecayFactor,
	}
	stats := scoreSongs(input)
	totalW := totalWeight(3, cfg.DecayFactor)

	result := orderSongs(stats, totalW, 3, cfg)
	if result[0].Name != "Opener" {
		t.Errorf("expected Opener first, got %q", result[0].Name)
	}
	if result[0].Position != 1 {
		t.Errorf("expected position 1, got %d", result[0].Position)
	}
}

func TestOrderSongs_CloserLock(t *testing.T) {
	cfg := DefaultConfig()
	input := scoringInput{
		setlists: []simplifiedSetlist{
			{songs: []string{"A", "B", "Closer"}},
			{songs: []string{"A", "B", "Closer"}},
			{songs: []string{"A", "B", "Closer"}},
		},
		decay: cfg.DecayFactor,
	}
	stats := scoreSongs(input)
	totalW := totalWeight(3, cfg.DecayFactor)

	result := orderSongs(stats, totalW, 3, cfg)
	if result[len(result)-1].Name != "Closer" {
		t.Errorf("expected Closer last, got %q", result[len(result)-1].Name)
	}
}

func TestOrderSongs_TopN(t *testing.T) {
	cfg := DefaultConfig()
	input := scoringInput{
		setlists: []simplifiedSetlist{
			{songs: []string{"A", "B", "C", "D", "E"}},
			{songs: []string{"A", "B", "C", "D", "E"}},
		},
		decay: cfg.DecayFactor,
	}
	stats := scoreSongs(input)
	totalW := totalWeight(2, cfg.DecayFactor)

	result := orderSongs(stats, totalW, 3, cfg)
	if len(result) != 3 {
		t.Fatalf("expected 3 songs, got %d", len(result))
	}
}
