package database

import (
	"context"
	"database/sql"
	"fmt"
)

// PostgresUserRepository implements UserRepository using PostgreSQL.
type PostgresUserRepository struct {
	db *sql.DB
}

// NewPostgresUserRepository creates a new Postgres-backed user repository.
func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// Create inserts a new user (created on first Spotify login).
func (r *PostgresUserRepository) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (id, spotify_id, email, display_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.SpotifyID,
		user.Email,
		user.DisplayName,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

// GetByID retrieves a user by internal ID.
func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*User, error) {
	query := `SELECT id, spotify_id, email, display_name, created_at, updated_at FROM users WHERE id = $1`
	return r.scanUser(r.db.QueryRowContext(ctx, query, id))
}

// GetBySpotifyID retrieves a user by their Spotify ID (used on login).
func (r *PostgresUserRepository) GetBySpotifyID(ctx context.Context, spotifyID string) (*User, error) {
	query := `SELECT id, spotify_id, email, display_name, created_at, updated_at FROM users WHERE spotify_id = $1`
	return r.scanUser(r.db.QueryRowContext(ctx, query, spotifyID))
}

func (r *PostgresUserRepository) scanUser(row *sql.Row) (*User, error) {
	user := &User{}
	err := row.Scan(
		&user.ID,
		&user.SpotifyID,
		&user.Email,
		&user.DisplayName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query user: %w", err)
	}
	return user, nil
}
