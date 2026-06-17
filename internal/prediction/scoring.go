package prediction

import (
	"math"
	"sort"
)

// Config holds prediction algorithm parameters.
type Config struct {
	DecayFactor      float64
	OpenerThreshold  float64
	CloserThreshold  float64
}

// DefaultConfig returns default prediction configuration.
func DefaultConfig() Config {
	return Config{
		DecayFactor:     0.85,
		OpenerThreshold: 0.70,
		CloserThreshold: 0.70,
	}
}

// songStats tracks a song's appearances and positions across setlists.
type songStats struct {
	name       string
	weights    float64 // sum of weights where song appeared
	positions  []weightedPos
	openerCount float64 // weighted count as opener
	closerCount float64 // weighted count as closer
}

type weightedPos struct {
	position int
	weight   float64
}

// scoreSongs computes weighted frequency scores for all songs.
func scoreSongs(input scoringInput) map[string]*songStats {
	stats := make(map[string]*songStats)

	for i, sl := range input.setlists {
		w := math.Pow(input.decay, float64(i))
		setLen := len(sl.songs)
		for pos, songName := range sl.songs {
			s, ok := stats[songName]
			if !ok {
				s = &songStats{name: songName}
				stats[songName] = s
			}
			s.weights += w
			s.positions = append(s.positions, weightedPos{position: pos, weight: w})
			if pos == 0 {
				s.openerCount += w
			}
			if pos == setLen-1 {
				s.closerCount += w
			}
		}
	}

	return stats
}

// scoringInput is the internal representation for the scoring algorithm.
type scoringInput struct {
	setlists []simplifiedSetlist
	decay    float64
}

type simplifiedSetlist struct {
	songs []string
}

// totalWeight returns the sum of all setlist weights.
func totalWeight(n int, decay float64) float64 {
	total := 0.0
	for i := range n {
		total += math.Pow(decay, float64(i))
	}
	return total
}

// predictPosition calculates the weighted median position for a song.
func predictPosition(positions []weightedPos) int {
	if len(positions) == 0 {
		return 0
	}
	sort.Slice(positions, func(i, j int) bool {
		return positions[i].position < positions[j].position
	})

	totalW := 0.0
	for _, p := range positions {
		totalW += p.weight
	}

	half := totalW / 2
	cumulative := 0.0
	for _, p := range positions {
		cumulative += p.weight
		if cumulative >= half {
			return p.position
		}
	}
	return positions[len(positions)-1].position
}

// medianLength returns the median setlist length.
func medianLength(lengths []int) int {
	if len(lengths) == 0 {
		return 0
	}
	sorted := make([]int, len(lengths))
	copy(sorted, lengths)
	sort.Ints(sorted)
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}
