package orchestration

import (
	"testing"
	"time"
)

func TestJobStore_CreateAndGet(t *testing.T) {
	store := NewJobStore()
	job := &Job{ID: "j_test1", Status: JobStatusPending, CreatedAt: time.Now()}

	if err := store.Create(job); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := store.Get("j_test1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status != JobStatusPending {
		t.Errorf("expected pending, got %s", got.Status)
	}
}

func TestJobStore_GetNotFound(t *testing.T) {
	store := NewJobStore()
	_, err := store.Get("unknown")
	if err != ErrJobNotFound {
		t.Fatalf("expected ErrJobNotFound, got %v", err)
	}
}

func TestJobStore_Update(t *testing.T) {
	store := NewJobStore()
	job := &Job{ID: "j_test2", Status: JobStatusPending}
	store.Create(job)

	job.Status = JobStatusCompleted
	store.Update(job)

	got, _ := store.Get("j_test2")
	if got.Status != JobStatusCompleted {
		t.Errorf("expected completed, got %s", got.Status)
	}
}

func TestGenerateJobID(t *testing.T) {
	id := GenerateJobID()
	if len(id) < 4 || id[:2] != "j_" {
		t.Errorf("unexpected job ID format: %q", id)
	}
}
