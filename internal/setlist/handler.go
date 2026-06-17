package setlist

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

// Handler handles setlist HTTP endpoints.
type Handler struct {
	service *Service
}

// RegisterHandlers registers setlist routes on the mux.
func RegisterHandlers(mux *http.ServeMux, service *Service) {
	h := &Handler{service: service}
	mux.HandleFunc("GET /v1/artists/{mbid}/setlists", h.GetSetlists)
}

func (h *Handler) GetSetlists(w http.ResponseWriter, r *http.Request) {
	mbid := r.PathValue("mbid")

	result, err := h.service.GetRecentSetlists(r.Context(), mbid)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidMBID):
		writeErr(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrNoRecentSetlists):
		writeErr(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrRateLimit):
		writeErr(w, http.StatusTooManyRequests, "rate limit exceeded")
	case errors.Is(err, ErrTimeout):
		writeErr(w, http.StatusGatewayTimeout, "upstream timeout")
	default:
		slog.Error("setlist fetch failed", "error", err)
		writeErr(w, http.StatusBadGateway, "service unavailable")
	}
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
