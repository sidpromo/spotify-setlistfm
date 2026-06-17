package artist

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTestHandler(t *testing.T) (*Handler, *httptest.Server) {
	t.Helper()
	sfmServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/1.0/search/artists":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"artist":[{"mbid":"abc","name":"Metallica","sortName":"Metallica"}],"total":1,"page":1,"itemsPerPage":20}`))
		case r.URL.Path == "/1.0/artist/abc/setlists":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"setlist":[{"eventDate":"15-05-2026"}],"total":10,"page":1,"itemsPerPage":20}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	client := NewSetlistFMClient(sfmServer.URL, "test-key", sfmServer.Client())
	service := NewService(client)
	handler := NewHandler(service)
	return handler, sfmServer
}

func TestHandler_Search_Success(t *testing.T) {
	h, srv := setupTestHandler(t)
	defer srv.Close()

	req := httptest.NewRequest(http.MethodGet, "/v1/artists/search?q=Metallica", nil)
	w := httptest.NewRecorder()
	h.Search(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result ArtistSearchResult
	json.NewDecoder(w.Body).Decode(&result)
	if len(result.Artists) != 1 {
		t.Fatalf("expected 1 artist, got %d", len(result.Artists))
	}
}

func TestHandler_Search_MissingQuery(t *testing.T) {
	h, srv := setupTestHandler(t)
	defer srv.Close()

	req := httptest.NewRequest(http.MethodGet, "/v1/artists/search", nil)
	w := httptest.NewRecorder()
	h.Search(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Search_InvalidPage(t *testing.T) {
	h, srv := setupTestHandler(t)
	defer srv.Close()

	req := httptest.NewRequest(http.MethodGet, "/v1/artists/search?q=test&page=0", nil)
	w := httptest.NewRecorder()
	h.Search(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Search_InvalidPageNotNumber(t *testing.T) {
	h, srv := setupTestHandler(t)
	defer srv.Close()

	req := httptest.NewRequest(http.MethodGet, "/v1/artists/search?q=test&page=abc", nil)
	w := httptest.NewRecorder()
	h.Search(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Search_ProviderError(t *testing.T) {
	sfmServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer sfmServer.Close()

	client := NewSetlistFMClient(sfmServer.URL, "test-key", sfmServer.Client())
	service := NewService(client)
	h := NewHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/v1/artists/search?q=test", nil)
	w := httptest.NewRecorder()
	h.Search(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", w.Code)
	}
}
