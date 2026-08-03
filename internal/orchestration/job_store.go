package orchestration

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
)

// InMemoryJobStore is an in-memory implementation of JobRepository.
type InMemoryJobStore struct {
	mu   sync.RWMutex
	jobs map[string]*Job
}

// NewInMemoryJobStore creates a new in-memory job store.
func NewInMemoryJobStore() *InMemoryJobStore {
	return &InMemoryJobStore{jobs: make(map[string]*Job)}
}

// Create stores a new job.
func (s *InMemoryJobStore) Create(_ context.Context, job *Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.ID] = job
	return nil
}

// Get retrieves a job by ID.
func (s *InMemoryJobStore) Get(_ context.Context, id string) (*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, ok := s.jobs[id]
	if !ok {
		return nil, ErrJobNotFound
	}
	return job, nil
}

// Update updates a job in the store.
func (s *InMemoryJobStore) Update(_ context.Context, job *Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.ID] = job
	return nil
}

// ListByUser returns jobs for a user (in-memory ignores userID — returns all).
func (s *InMemoryJobStore) ListByUser(_ context.Context, _ string, limit, offset int) ([]*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var all []*Job
	for _, j := range s.jobs {
		all = append(all, j)
	}
	if offset >= len(all) {
		return nil, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], nil
}

// FindActive returns an active job (pending/processing) for a user + artist combo.
// Returns nil if no active job exists.
func (s *InMemoryJobStore) FindActive(_ context.Context, userID, artistMBID string) (*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, j := range s.jobs {
		if j.Request.UserID == userID && j.Request.ArtistMBID == artistMBID &&
			(j.Status == JobStatusPending || j.Status == JobStatusProcessing) {
			return j, nil
		}
	}
	return nil, nil
}

// GenerateJobID creates a random job ID.
func GenerateJobID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "j_" + hex.EncodeToString(b)
}
