package prediction

import "context"

// Service predicts setlists from recent concert data.
type Service struct {
	cfg Config
}

// NewService creates a new prediction service.
func NewService(cfg Config) *Service {
	return &Service{cfg: cfg}
}

// Predict generates a predicted setlist from input setlists.
func (s *Service) Predict(_ context.Context, input PredictionInput) (*PredictedSetlist, error) {
	if len(input.Setlists) == 0 {
		return nil, ErrNotEnoughData
	}

	// Convert to simplified format
	simplified := make([]simplifiedSetlist, len(input.Setlists))
	lengths := make([]int, len(input.Setlists))
	for i, sl := range input.Setlists {
		songs := make([]string, len(sl.Songs))
		for j, song := range sl.Songs {
			songs[j] = song.Name
		}
		simplified[i] = simplifiedSetlist{songs: songs}
		lengths[i] = len(songs)
	}

	targetLen := medianLength(lengths)
	if targetLen == 0 {
		return nil, ErrNotEnoughData
	}

	si := scoringInput{setlists: simplified, decay: s.cfg.DecayFactor}
	stats := scoreSongs(si)
	totalW := totalWeight(len(simplified), s.cfg.DecayFactor)

	songs := orderSongs(stats, totalW, targetLen, s.cfg)

	return &PredictedSetlist{
		Artist:       input.Artist,
		MBID:         input.MBID,
		TourName:     input.TourName,
		Songs:        songs,
		BasedOnCount: len(input.Setlists),
	}, nil
}
