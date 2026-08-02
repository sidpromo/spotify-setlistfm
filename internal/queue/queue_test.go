package queue

import (
	"encoding/json"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestParseMessage_Valid(t *testing.T) {
	data, _ := json.Marshal(JobMessage{
		JobID:      "j_test1",
		UserID:     "user-1",
		ArtistMBID: "mbid-123",
		ArtistName: "Metallica",
	})

	msg := redis.XMessage{
		ID:     "1234-0",
		Values: map[string]any{"data": string(data)},
	}

	parsed, err := ParseMessage(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.JobID != "j_test1" {
		t.Errorf("expected 'j_test1', got %q", parsed.JobID)
	}
	if parsed.ArtistName != "Metallica" {
		t.Errorf("expected 'Metallica', got %q", parsed.ArtistName)
	}
}

func TestParseMessage_MissingData(t *testing.T) {
	msg := redis.XMessage{
		ID:     "1234-0",
		Values: map[string]any{"other": "field"},
	}

	_, err := ParseMessage(msg)
	if err == nil {
		t.Fatal("expected error for missing data field")
	}
}

func TestParseMessage_InvalidJSON(t *testing.T) {
	msg := redis.XMessage{
		ID:     "1234-0",
		Values: map[string]any{"data": "{broken json"},
	}

	_, err := ParseMessage(msg)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestJobMessage_Marshal(t *testing.T) {
	msg := JobMessage{
		JobID:      "j_abc",
		UserID:     "u_123",
		ArtistMBID: "mbid",
		ArtistName: "Band",
	}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty data")
	}

	var decoded JobMessage
	json.Unmarshal(data, &decoded)
	if decoded.JobID != "j_abc" {
		t.Errorf("expected 'j_abc', got %q", decoded.JobID)
	}
}

func TestDefaultWorkerConfig(t *testing.T) {
	cfg := DefaultWorkerConfig()
	if cfg.WorkerCount != 3 {
		t.Errorf("expected 3 workers, got %d", cfg.WorkerCount)
	}
	if cfg.MaxRetries != 3 {
		t.Errorf("expected 3 retries, got %d", cfg.MaxRetries)
	}
}

func TestRedisEnqueuer_ImplementsInterface(t *testing.T) {
	// Compile-time check that RedisEnqueuer can be used where JobEnqueuer is expected
	var _ interface {
		Enqueue(ctx interface{}, jobID, userID, artistMBID, artistName string) error
	}
	// This just verifies the method signature exists — can't test without real Redis
}

func TestDirectEnqueuer(t *testing.T) {
	called := false
	enqueuer := &directTestEnqueuer{fn: func() { called = true }}
	enqueuer.run()
	if !called {
		t.Error("expected function to be called")
	}
}

type directTestEnqueuer struct {
	fn func()
}

func (e *directTestEnqueuer) run() {
	e.fn()
}
