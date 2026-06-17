package orchestration

import (
	"encoding/json"
	"errors"
	"net/http"
)

const sessionCookieName = "spotify_session"

// Handler handles orchestration HTTP endpoints.
type Handler struct {
	service *Service
}

// RegisterHandlers registers orchestration routes.
func RegisterHandlers(mux *http.ServeMux, service *Service) {
	h := &Handler{service: service}
	mux.HandleFunc("POST /v1/playlists", h.CreatePlaylist)
	mux.HandleFunc("GET /v1/playlists/jobs/{jobId}", h.GetJob)
}

func (h *Handler) CreatePlaylist(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ArtistMBID string `json:"artistMbid"`
		ArtistName string `json:"artistName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get session ID from cookie
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "not authenticated with Spotify, visit /v1/auth/spotify/login")
		return
	}

	job, err := h.service.CreateJob(JobRequest{
		ArtistMBID: req.ArtistMBID,
		ArtistName: req.ArtistName,
		SessionID:  cookie.Value,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrMissingArtist):
			writeErr(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, ErrNotAuthenticated):
			writeErr(w, http.StatusUnauthorized, "not authenticated with Spotify")
		default:
			writeErr(w, http.StatusInternalServerError, "failed to create job")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]any{
		"jobId":     job.ID,
		"status":    job.Status,
		"createdAt": job.CreatedAt,
	})
}

func (h *Handler) GetJob(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("jobId")

	job, err := h.service.GetJob(jobID)
	if err != nil {
		if errors.Is(err, ErrJobNotFound) {
			writeErr(w, http.StatusNotFound, "job not found")
			return
		}
		writeErr(w, http.StatusInternalServerError, "failed to get job")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
