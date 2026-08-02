package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// JobRow represents a job row in PostgreSQL.
type JobRow struct {
	ID         string
	UserID     string
	Status     string
	ArtistMBID string
	ArtistName string
	Result     sql.NullString
	Error      sql.NullString
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// PostgresJobRepository implements JobRepository using PostgreSQL.
type PostgresJobRepository struct {
	db *sql.DB
}

// NewPostgresJobRepository creates a new Postgres-backed job repository.
func NewPostgresJobRepository(db *sql.DB) *PostgresJobRepository {
	return &PostgresJobRepository{db: db}
}

// Create inserts a new job into the database.
func (r *PostgresJobRepository) Create(ctx context.Context, job *JobRow) error {
	query := `
		INSERT INTO jobs (id, user_id, status, artist_mbid, artist_name, result, error, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.ExecContext(ctx, query,
		job.ID,
		job.UserID,
		job.Status,
		job.ArtistMBID,
		job.ArtistName,
		job.Result,
		job.Error,
		job.CreatedAt,
		job.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert job: %w", err)
	}
	return nil
}

// Get retrieves a job by ID.
func (r *PostgresJobRepository) Get(ctx context.Context, id string) (*JobRow, error) {
	query := `
		SELECT id, user_id, status, artist_mbid, artist_name, result, error, created_at, updated_at
		FROM jobs WHERE id = $1`

	row := r.db.QueryRowContext(ctx, query, id)
	job := &JobRow{}
	err := row.Scan(
		&job.ID,
		&job.UserID,
		&job.Status,
		&job.ArtistMBID,
		&job.ArtistName,
		&job.Result,
		&job.Error,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("job not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("query job: %w", err)
	}
	return job, nil
}

// Update updates a job's status, result, and error fields.
func (r *PostgresJobRepository) Update(ctx context.Context, job *JobRow) error {
	query := `
		UPDATE jobs SET status = $1, result = $2, error = $3, updated_at = $4
		WHERE id = $5`

	_, err := r.db.ExecContext(ctx, query,
		job.Status,
		job.Result,
		job.Error,
		time.Now(),
		job.ID,
	)
	if err != nil {
		return fmt.Errorf("update job: %w", err)
	}
	return nil
}

// ListByUser retrieves jobs for a user, ordered by creation date (newest first).
func (r *PostgresJobRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*JobRow, error) {
	query := `
		SELECT id, user_id, status, artist_mbid, artist_name, result, error, created_at, updated_at
		FROM jobs WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*JobRow
	for rows.Next() {
		job := &JobRow{}
		if err := rows.Scan(
			&job.ID,
			&job.UserID,
			&job.Status,
			&job.ArtistMBID,
			&job.ArtistName,
			&job.Result,
			&job.Error,
			&job.CreatedAt,
			&job.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan job row: %w", err)
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

// MarshalResult converts a result struct to JSON for storage.
func MarshalResult(v any) sql.NullString {
	if v == nil {
		return sql.NullString{}
	}
	data, err := json.Marshal(v)
	if err != nil {
		return sql.NullString{}
	}
	return sql.NullString{String: string(data), Valid: true}
}

// UnmarshalResult reads JSON from storage into a struct.
func UnmarshalResult(s sql.NullString, v any) error {
	if !s.Valid {
		return nil
	}
	return json.Unmarshal([]byte(s.String), v)
}
