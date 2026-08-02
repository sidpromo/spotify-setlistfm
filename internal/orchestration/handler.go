package orchestration

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sidpromo/spotify-setlistfm/internal/auth"
)

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
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		writeErr(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req struct {
		ArtistMBID string `json:"artistMbid"`
		ArtistName string `json:"artistName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body")
		return
	}

	job, err := h.service.CreateJob(JobRequest{
		ArtistMBID: req.ArtistMBID,
		ArtistName: req.ArtistName,
		UserID:     userID,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrMissingArtist):
			writeErr(w, http.StatusBadRequest, err.Error())
		default:
			writeErr(w, http.StatusInternalServerError, "failed to create job")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{ // #nosec G104
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
	_ = json.NewEncoder(w).Encode(job) // #nosec G104
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg}) // #nosec G104
}
