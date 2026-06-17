package prediction

import (
	"testing"
)

func TestScoreSongs_AllSetlists(t *testing.T) {
	input := scoringInput{
		setlists: []simplifiedSetlist{
			{songs: []string{"A", "B", "C"}},
			{songs: []string{"A", "B", "C"}},
			{songs: []string{"A", "B", "C"}},
		},
		decay: 0.85,
	}
	stats := scoreSongs(input)
	totalW := totalWeight(3, 0.85)

	scoreA := stats["A"].weights / totalW
	if scoreA < 0.99 {
		t.Errorf("song in all setlists should score ~1.0, got %f", scoreA)
	}
}

func TestScoreSongs_RecencyMatters(t *testing.T) {
	input := scoringInput{
		setlists: []simplifiedSetlist{
			{songs: []string{"New"}},        // most recent, weight 1.0
			{songs: []string{"Old"}},        // weight 0.85
			{songs: []string{"Old"}},        // weight 0.72
		},
		decay: 0.85,
	}
	stats := scoreSongs(input)
	totalW := totalWeight(3, 0.85)

	scoreNew := stats["New"].weights / totalW
	scoreOld := stats["Old"].weights / totalW

	// "New" played once recently vs "Old" played twice but older
	// New weight = 1.0, Old weight = 0.85 + 0.72 = 1.57
	if scoreNew >= scoreOld {
		t.Errorf("Old (played 2x) should score higher than New (played 1x): new=%f old=%f", scoreNew, scoreOld)
	}
}

func TestTotalWeight(t *testing.T) {
	w := totalWeight(3, 0.85)
	expected := 1.0 + 0.85 + 0.85*0.85 // 2.5725
	if diff := w - expected; diff > 0.001 || diff < -0.001 {
		t.Errorf("expected %f, got %f", expected, w)
	}
}

func TestMedianLength_Odd(t *testing.T) {
	got := medianLength([]int{18, 20, 16})
	if got != 18 {
		t.Errorf("expected 18, got %d", got)
	}
}

func TestMedianLength_Even(t *testing.T) {
	got := medianLength([]int{18, 20, 16, 22})
	if got != 19 { // (18+20)/2
		t.Errorf("expected 19, got %d", got)
	}
}

func TestMedianLength_Empty(t *testing.T) {
	got := medianLength(nil)
	if got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

func TestPredictPosition(t *testing.T) {
	positions := []weightedPos{
		{position: 0, weight: 1.0},
		{position: 0, weight: 0.85},
		{position: 1, weight: 0.72},
	}
	got := predictPosition(positions)
	if got != 0 {
		t.Errorf("expected position 0, got %d", got)
	}
}
