package artist

// Artist represents an internal artist model.
type Artist struct {
	MBID           string `json:"mbid"`
	Name           string `json:"name"`
	SortName       string `json:"sortName,omitempty"`
	Disambiguation string `json:"disambiguation,omitempty"`
	URL            string `json:"url,omitempty"`
}

// ArtistSearchResult is the response for an artist search.
type ArtistSearchResult struct {
	Artists      []Artist `json:"artists"`
	Total        int      `json:"total"`
	ItemsPerPage int      `json:"itemsPerPage"`
	Page         int      `json:"page"`
}
