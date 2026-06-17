package orchestration

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

// JobStore is an in-memory job store.
type JobStore struct {
	mu   sync.RWMutex
	jobs map[string]*Job
}

// NewJobStore creates a new in-memory job store.
func NewJobStore() *JobStore {
	return &JobStore{jobs: make(map[string]*Job)}
}

// Create stores a new job.
func (s *JobStore) Create(job *Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.ID] = job
	return nil
}

// Get retrieves a job by ID.
func (s *JobStore) Get(id string) (*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, ok := s.jobs[id]
	if !ok {
		return nil, ErrJobNotFound
	}
	return job, nil
}

// Update updates a job in the store.
func (s *JobStore) Update(job *Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.ID] = job
	return nil
}

// GenerateJobID creates a random job ID.
func GenerateJobID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "j_" + hex.EncodeToString(b)
}
