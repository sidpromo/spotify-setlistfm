package orchestration

import "context"

// JobRepository defines the persistence interface for jobs.
// Implementations: PostgresJobRepository (production), InMemoryJobStore (tests).
type JobRepository interface {
	Create(ctx context.Context, job *Job) error
	Get(ctx context.Context, id string) (*Job, error)
	Update(ctx context.Context, job *Job) error
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]*Job, error)
}
