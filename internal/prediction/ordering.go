package prediction

import "sort"

// orderSongs takes scored songs, applies opener/closer locks, and returns ordered results.
func orderSongs(stats map[string]*songStats, totalW float64, targetLen int, cfg Config) []PredictedSong {
	// Calculate scores and collect candidates
	type candidate struct {
		name  string
		score float64
		pos   int
	}

	var candidates []candidate
	for _, s := range stats {
		score := s.weights / totalW
		pos := predictPosition(s.positions)
		candidates = append(candidates, candidate{name: s.name, score: score, pos: pos})
	}

	// Sort by score descending, take top N
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})
	if len(candidates) > targetLen {
		candidates = candidates[:targetLen]
	}

	// Identify opener and closer locks
	var opener, closer string
	for _, s := range stats {
		ratio := s.openerCount / totalW
		if ratio >= cfg.OpenerThreshold {
			opener = s.name
		}
		ratio = s.closerCount / totalW
		if ratio >= cfg.CloserThreshold {
			closer = s.name
		}
	}

	// Sort by predicted position
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].pos < candidates[j].pos
	})

	// Apply locks: ensure opener is first, closer is last
	result := make([]PredictedSong, 0, len(candidates))
	for i, c := range candidates {
		result = append(result, PredictedSong{
			Name:     c.name,
			Position: i + 1,
			Score:    c.score,
		})
	}

	// Move opener to position 1 if locked
	if opener != "" {
		for i, s := range result {
			if s.Name == opener && i != 0 {
				result = append([]PredictedSong{result[i]}, append(result[:i], result[i+1:]...)...)
				break
			}
		}
	}
	// Move closer to last position if locked
	if closer != "" && closer != opener {
		for i, s := range result {
			if s.Name == closer && i != len(result)-1 {
				item := result[i]
				result = append(result[:i], result[i+1:]...)
				result = append(result, item)
				break
			}
		}
	}

	// Reassign positions after reordering
	for i := range result {
		result[i].Position = i + 1
	}

	return result
}
