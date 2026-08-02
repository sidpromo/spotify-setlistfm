package orchestration

import "time"

// JobStatus represents the state of an async job.
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

// Job represents an async playlist creation job.
type Job struct {
	ID        string     `json:"id"`
	Status    JobStatus  `json:"status"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	Request   JobRequest `json:"request"`
	Result    *JobResult `json:"result,omitempty"`
	Error     string     `json:"error,omitempty"`
}

// JobRequest is the input for creating a playlist job.
type JobRequest struct {
	ArtistMBID string `json:"artistMbid"`
	ArtistName string `json:"artistName"`
	UserID     string `json:"-"`
}

// JobResult holds the output of a completed job.
type JobResult struct {
	PlaylistID   string   `json:"playlistId"`
	PlaylistURL  string   `json:"playlistUrl"`
	PlaylistName string   `json:"playlistName"`
	TracksAdded  int      `json:"tracksAdded"`
	TracksTotal  int      `json:"tracksTotal"`
	NotFound     []string `json:"notFound,omitempty"`
	TourName     string   `json:"tourName,omitempty"`
	BasedOnCount int      `json:"basedOnCount"`
}
