package setlist

import "time"

// Song represents a song in a setlist.
type Song struct {
	Name    string `json:"name"`
	IsCover bool   `json:"isCover,omitempty"`
}

// Setlist represents a single concert setlist.
type Setlist struct {
	ID        string    `json:"id"`
	EventDate time.Time `json:"eventDate"`
	Venue     string    `json:"venue"`
	City      string    `json:"city"`
	TourName  string    `json:"tourName,omitempty"`
	Songs     []Song    `json:"songs"`
}

// SetlistResult is the output of fetching setlists for an artist.
type SetlistResult struct {
	Artist   string    `json:"artist"`
	MBID     string    `json:"mbid"`
	TourName string    `json:"tourName,omitempty"`
	Setlists []Setlist `json:"setlists"`
}
