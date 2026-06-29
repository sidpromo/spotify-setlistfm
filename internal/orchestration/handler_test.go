package orchestration

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sidpromo/spotify-setlistfm/internal/prediction"
	"github.com/sidpromo/spotify-setlistfm/internal/setlist"
	"github.com/sidpromo/spotify-setlistfm/internal/spotify"
)

func newTestService() *Service {
	ss := &mockSetlistService{result: &setlist.SetlistResult{
		Artist: "Band", Setlists: []setlist.Setlist{{Songs: []setlist.Song{{Name: "S1"}, {Name: "S2"}, {Name: "S3"}, {Name: "S4"}, {Name: "S5"}}, EventDate: time.Now()}},
	}}
	ps := &mockPredictionService{result: &prediction.PredictedSetlist{
		Songs: []prediction.PredictedSong{{Name: "S1"}}, BasedOnCount: 1,
	}}
	sp := &mockSpotifyService{result: &spotify.PlaylistResult{
		PlaylistID: "pl1", PlaylistURL: "http://x", Name: "n", TracksAdded: 1, TracksTotal: 1,
	}}
	return NewService(ss, ps, sp, NewJobStore())
}

func TestHandler_CreatePlaylist_Success(t *testing.T) {
	svc := newTestService()
	mux := http.NewServeMux()
	RegisterHandlers(mux, svc)

	body := `{"artistMbid":"abc","artistName":"Band"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/playlists", strings.NewReader(body))
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "sess1"})
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_CreatePlaylist_MissingFields(t *testing.T) {
	svc := newTestService()
	mux := http.NewServeMux()
	RegisterHandlers(mux, svc)

	body := `{"artistMbid":"","artistName":""}`
	req := httptest.NewRequest(http.MethodPost, "/v1/playlists", strings.NewReader(body))
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "sess1"})
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_CreatePlaylist_NoCookie(t *testing.T) {
	svc := newTestService()
	mux := http.NewServeMux()
	RegisterHandlers(mux, svc)

	body := `{"artistMbid":"abc","artistName":"Band"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/playlists", strings.NewReader(body))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestHandler_GetJob_Success(t *testing.T) {
	svc := newTestService()
	mux := http.NewServeMux()
	RegisterHandlers(mux, svc)

	// Create a job first
	job, _ := svc.CreateJob(JobRequest{ArtistMBID: "abc", ArtistName: "Band", SessionID: "s"})

	req := httptest.NewRequest(http.MethodGet, "/v1/playlists/jobs/"+job.ID, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_GetJob_NotFound(t *testing.T) {
	svc := newTestService()
	mux := http.NewServeMux()
	RegisterHandlers(mux, svc)

	req := httptest.NewRequest(http.MethodGet, "/v1/playlists/jobs/nonexistent", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
