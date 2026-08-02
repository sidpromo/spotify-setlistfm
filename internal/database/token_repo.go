package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// TokenRow represents a Spotify token row in PostgreSQL.
type TokenRow struct {
	UserID       string
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	UpdatedAt    time.Time
}

// PostgresTokenRepository stores Spotify tokens in PostgreSQL.
type PostgresTokenRepository struct {
	db *sql.DB
}

// NewPostgresTokenRepository creates a new Postgres-backed token repository.
func NewPostgresTokenRepository(db *sql.DB) *PostgresTokenRepository {
	return &PostgresTokenRepository{db: db}
}

// Save upserts a Spotify token for a user.
func (r *PostgresTokenRepository) Save(ctx context.Context, token *TokenRow) error {
	query := `
		INSERT INTO spotify_tokens (user_id, access_token, refresh_token, expires_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE SET
			access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			expires_at = EXCLUDED.expires_at,
			updated_at = EXCLUDED.updated_at`

	_, err := r.db.ExecContext(ctx, query,
		token.UserID,
		token.AccessToken,
		token.RefreshToken,
		token.ExpiresAt,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("save token: %w", err)
	}
	return nil
}

// Get retrieves a Spotify token for a user.
func (r *PostgresTokenRepository) Get(ctx context.Context, userID string) (*TokenRow, error) {
	query := `SELECT user_id, access_token, refresh_token, expires_at, updated_at FROM spotify_tokens WHERE user_id = $1`

	row := r.db.QueryRowContext(ctx, query, userID)
	token := &TokenRow{}
	err := row.Scan(
		&token.UserID,
		&token.AccessToken,
		&token.RefreshToken,
		&token.ExpiresAt,
		&token.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("token not found for user: %s", userID)
	}
	if err != nil {
		return nil, fmt.Errorf("query token: %w", err)
	}
	return token, nil
}

// Delete removes a Spotify token for a user.
func (r *PostgresTokenRepository) Delete(ctx context.Context, userID string) error {
	query := `DELETE FROM spotify_tokens WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("delete token: %w", err)
	}
	return nil
}
