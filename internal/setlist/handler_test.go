package setlist

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandler_GetSetlists_Success(t *testing.T) {
	recentDate := time.Now().AddDate(0, -1, 0).Format("02-01-2006")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"setlist":[%s],"total":1,"page":1,"itemsPerPage":20}`,
			makeSetlistJSON("s1", recentDate, "Tour", 8),
		)))
	}))
	defer srv.Close()

	client := NewSetlistFMClient(srv.URL, "key", srv.Client())
	svc := NewService(client, DefaultConfig())

	mux := http.NewServeMux()
	RegisterHandlers(mux, svc)

	req := httptest.NewRequest(http.MethodGet, "/v1/artists/65f4f0c5-ef9e-490c-aee3-909e7ae6b2ab/setlists", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_GetSetlists_InvalidMBID(t *testing.T) {
	client := NewSetlistFMClient("http://unused", "key", http.DefaultClient)
	svc := NewService(client, DefaultConfig())

	mux := http.NewServeMux()
	RegisterHandlers(mux, svc)

	req := httptest.NewRequest(http.MethodGet, "/v1/artists/not-valid/setlists", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
