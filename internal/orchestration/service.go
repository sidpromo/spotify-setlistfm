package orchestration

import (
	"context"
	"log/slog"
	"time"

	"github.com/sidpromo/spotify-setlistfm/internal/prediction"
	"github.com/sidpromo/spotify-setlistfm/internal/setlist"
	"github.com/sidpromo/spotify-setlistfm/internal/spotify"
)

// SetlistService is the interface the orchestration needs from the setlist module.
type SetlistService interface {
	GetRecentSetlists(ctx context.Context, mbid string) (*setlist.SetlistResult, error)
}

// PredictionService is the interface the orchestration needs from the prediction module.
type PredictionService interface {
	Predict(ctx context.Context, input prediction.PredictionInput) (*prediction.PredictedSetlist, error)
}

// SpotifyService is the interface the orchestration needs from the spotify module.
type SpotifyService interface {
	CreatePlaylist(ctx context.Context, sessionID string, input spotify.PlaylistInput) (*spotify.PlaylistResult, error)
}

// Service coordinates the full pipeline.
type Service struct {
	setlistSvc    SetlistService
	predictionSvc PredictionService
	spotifySvc    SpotifyService
	jobStore      *JobStore
}

// NewService creates a new orchestration service.
func NewService(ss SetlistService, ps PredictionService, sp SpotifyService, js *JobStore) *Service {
	return &Service{
		setlistSvc:    ss,
		predictionSvc: ps,
		spotifySvc:    sp,
		jobStore:      js,
	}
}

// CreateJob validates input and starts an async pipeline.
func (s *Service) CreateJob(req JobRequest) (*Job, error) {
	if req.ArtistMBID == "" || req.ArtistName == "" {
		return nil, ErrMissingArtist
	}

	now := time.Now()
	job := &Job{
		ID:        GenerateJobID(),
		Status:    JobStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
		Request:   req,
	}

	s.jobStore.Create(job)

	go s.runPipeline(job, req.SessionID)

	return job, nil
}

// GetJob retrieves a job by ID.
func (s *Service) GetJob(id string) (*Job, error) {
	return s.jobStore.Get(id)
}

func (s *Service) runPipeline(job *Job, sessionID string) {
	ctx := context.Background()

	s.updateStatus(job, JobStatusProcessing)

	// Step 1: Fetch setlists
	setlistResult, err := s.setlistSvc.GetRecentSetlists(ctx, job.Request.ArtistMBID)
	if err != nil {
		s.failJob(job, err.Error())
		return
	}

	// Step 2: Predict
	pred, err := s.predictionSvc.Predict(ctx, prediction.PredictionInput{
		Artist:   job.Request.ArtistName,
		MBID:     job.Request.ArtistMBID,
		TourName: setlistResult.TourName,
		Setlists: setlistResult.Setlists,
	})
	if err != nil {
		s.failJob(job, err.Error())
		return
	}

	// Step 3: Create Spotify playlist
	songs := make([]string, len(pred.Songs))
	for i, song := range pred.Songs {
		songs[i] = song.Name
	}

	playlistResult, err := s.spotifySvc.CreatePlaylist(ctx, sessionID, spotify.PlaylistInput{
		ArtistName: job.Request.ArtistName,
		TourName:   pred.TourName,
		Songs:      songs,
	})
	if err != nil {
		s.failJob(job, err.Error())
		return
	}

	// Step 4: Complete
	job.Status = JobStatusCompleted
	job.UpdatedAt = time.Now()
	job.Result = &JobResult{
		PlaylistID:   playlistResult.PlaylistID,
		PlaylistURL:  playlistResult.PlaylistURL,
		PlaylistName: playlistResult.Name,
		TracksAdded:  playlistResult.TracksAdded,
		TracksTotal:  playlistResult.TracksTotal,
		NotFound:     playlistResult.NotFound,
		TourName:     pred.TourName,
		BasedOnCount: pred.BasedOnCount,
	}
	s.jobStore.Update(job)
	slog.Info("job completed", "jobId", job.ID, "playlist", playlistResult.PlaylistURL)
}

func (s *Service) updateStatus(job *Job, status JobStatus) {
	job.Status = status
	job.UpdatedAt = time.Now()
	s.jobStore.Update(job)
}

func (s *Service) failJob(job *Job, errMsg string) {
	job.Status = JobStatusFailed
	job.UpdatedAt = time.Now()
	job.Error = errMsg
	s.jobStore.Update(job)
	slog.Error("job failed", "jobId", job.ID, "error", errMsg)
}
