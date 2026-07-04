package artist

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
)

// Handler handles artist HTTP endpoints.
type Handler struct {
	service *Service
}

// NewHandler creates a new artist handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterHandlers registers artist routes on the mux.
func RegisterHandlers(mux *http.ServeMux, service *Service) {
	h := NewHandler(service)
	mux.HandleFunc("GET /v1/artists/search", h.Search)
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	pageStr := r.URL.Query().Get("page")

	page := 1
	if pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "page must be a number")
			return
		}
	}

	result, err := h.service.Search(r.Context(), query, page)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result) // #nosec G104
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrEmptyQuery), errors.Is(err, ErrQueryTooLong), errors.Is(err, ErrInvalidPage):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrRateLimit):
		writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
	case errors.Is(err, ErrTimeout):
		writeError(w, http.StatusGatewayTimeout, "upstream timeout")
	default:
		slog.Error("artist search failed", "error", err)
		writeError(w, http.StatusBadGateway, "service unavailable")
	}
}

// TODO: handle encode errors (low priority — if response write fails, connection is dead)
func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg}) // #nosec G104
}
