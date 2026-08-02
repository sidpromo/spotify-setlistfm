package orchestration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sidpromo/spotify-setlistfm/internal/prediction"
	"github.com/sidpromo/spotify-setlistfm/internal/setlist"
	"github.com/sidpromo/spotify-setlistfm/internal/spotify"
)

// Mock services
type mockSetlistService struct {
	result *setlist.SetlistResult
	err    error
}

func (m *mockSetlistService) GetRecentSetlists(_ context.Context, _ string) (*setlist.SetlistResult, error) {
	return m.result, m.err
}

type mockPredictionService struct {
	result *prediction.PredictedSetlist
	err    error
}

func (m *mockPredictionService) Predict(_ context.Context, _ prediction.PredictionInput) (*prediction.PredictedSetlist, error) {
	return m.result, m.err
}

type mockSpotifyService struct {
	result *spotify.PlaylistResult
	err    error
}

func (m *mockSpotifyService) CreatePlaylist(_ context.Context, _ string, _ spotify.PlaylistInput) (*spotify.PlaylistResult, error) {
	return m.result, m.err
}

func TestService_CreateJob_Success(t *testing.T) {
	ss := &mockSetlistService{result: &setlist.SetlistResult{
		Artist: "Band", TourName: "Tour",
		Setlists: []setlist.Setlist{{Songs: []setlist.Song{{Name: "S1"}, {Name: "S2"}, {Name: "S3"}, {Name: "S4"}, {Name: "S5"}}, EventDate: time.Now()}},
	}}
	ps := &mockPredictionService{result: &prediction.PredictedSetlist{
		Songs: []prediction.PredictedSong{{Name: "S1"}, {Name: "S2"}}, TourName: "Tour", BasedOnCount: 1,
	}}
	sp := &mockSpotifyService{result: &spotify.PlaylistResult{
		PlaylistID: "pl1", PlaylistURL: "http://x", Name: "Band - Tour (Predicted Setlist)", TracksAdded: 2, TracksTotal: 2,
	}}

	svc := NewService(ss, ps, sp, NewInMemoryJobStore(), nil)
	job, err := svc.CreateJob(JobRequest{ArtistMBID: "abc-def-123-456-789012345678", ArtistName: "Band", UserID: "user-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.Status != JobStatusPending {
		t.Errorf("expected pending, got %s", job.Status)
	}

	// Wait for async
	time.Sleep(50 * time.Millisecond)

	got, _ := svc.GetJob(job.ID)
	if got.Status != JobStatusCompleted {
		t.Fatalf("expected completed, got %s (error: %s)", got.Status, got.Error)
	}
	if got.Result.PlaylistURL != "http://x" {
		t.Errorf("unexpected url: %q", got.Result.PlaylistURL)
	}
}

func TestService_CreateJob_SetlistFails(t *testing.T) {
	ss := &mockSetlistService{err: errors.New("no recent setlists")}
	svc := NewService(ss, nil, nil, NewInMemoryJobStore(), nil)

	job, _ := svc.CreateJob(JobRequest{ArtistMBID: "abc", ArtistName: "Band", UserID: "user-1"})
	time.Sleep(50 * time.Millisecond)

	got, _ := svc.GetJob(job.ID)
	if got.Status != JobStatusFailed {
		t.Fatalf("expected failed, got %s", got.Status)
	}
	if got.Error != "no recent setlists" {
		t.Errorf("unexpected error: %q", got.Error)
	}
}

func TestService_CreateJob_PredictionFails(t *testing.T) {
	ss := &mockSetlistService{result: &setlist.SetlistResult{Setlists: []setlist.Setlist{{Songs: []setlist.Song{{Name: "X"}}}}}}
	ps := &mockPredictionService{err: errors.New("not enough data")}
	svc := NewService(ss, ps, nil, NewInMemoryJobStore(), nil)

	job, _ := svc.CreateJob(JobRequest{ArtistMBID: "abc", ArtistName: "Band", UserID: "user-1"})
	time.Sleep(50 * time.Millisecond)

	got, _ := svc.GetJob(job.ID)
	if got.Status != JobStatusFailed {
		t.Fatalf("expected failed, got %s", got.Status)
	}
}

func TestService_CreateJob_SpotifyFails(t *testing.T) {
	ss := &mockSetlistService{result: &setlist.SetlistResult{Setlists: []setlist.Setlist{{Songs: []setlist.Song{{Name: "X"}}}}}}
	ps := &mockPredictionService{result: &prediction.PredictedSetlist{Songs: []prediction.PredictedSong{{Name: "X"}}}}
	sp := &mockSpotifyService{err: errors.New("spotify unavailable")}
	svc := NewService(ss, ps, sp, NewInMemoryJobStore(), nil)

	job, _ := svc.CreateJob(JobRequest{ArtistMBID: "abc", ArtistName: "Band", UserID: "user-1"})
	time.Sleep(50 * time.Millisecond)

	got, _ := svc.GetJob(job.ID)
	if got.Status != JobStatusFailed {
		t.Fatalf("expected failed, got %s", got.Status)
	}
}

func TestService_CreateJob_MissingFields(t *testing.T) {
	svc := NewService(nil, nil, nil, NewInMemoryJobStore(), nil)
	_, err := svc.CreateJob(JobRequest{})
	if err != ErrMissingArtist {
		t.Fatalf("expected ErrMissingArtist, got %v", err)
	}
}
