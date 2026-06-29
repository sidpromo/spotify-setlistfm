package prediction

import "github.com/sidpromo/spotify-setlistfm/internal/setlist"

// PredictedSong is a song with its predicted position.
type PredictedSong struct {
	Name     string  `json:"name"`
	Position int     `json:"position"`
	Score    float64 `json:"-"`
}

// PredictedSetlist is the output of the prediction algorithm.
type PredictedSetlist struct {
	Artist       string          `json:"artist"`
	MBID         string          `json:"mbid"`
	TourName     string          `json:"tourName,omitempty"`
	Songs        []PredictedSong `json:"songs"`
	BasedOnCount int             `json:"basedOnCount"`
}

// PredictionInput is the input to the prediction service.
type PredictionInput struct {
	Artist   string
	MBID     string
	TourName string
	Setlists []setlist.Setlist
}
