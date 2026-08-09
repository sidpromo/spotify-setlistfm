package spotify

import (
	"context"
	"database/sql"
	"time"
)

// PersistentTokenStore stores Spotify tokens in PostgreSQL.
// Implements the same interface as the in-memory TokenStore.
type PersistentTokenStore struct {
	db *sql.DB
}

// NewPersistentTokenStore creates a Postgres-backed token store.
func NewPersistentTokenStore(db *sql.DB) *PersistentTokenStore {
	return &PersistentTokenStore{db: db}
}

// Save upserts a token for a user.
func (s *PersistentTokenStore) Save(userID string, token *Token) error {
	query := `
		INSERT INTO spotify_tokens (user_id, access_token, refresh_token, expires_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE SET
			access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			expires_at = EXCLUDED.expires_at,
			updated_at = EXCLUDED.updated_at`

	_, err := s.db.ExecContext(context.Background(), query,
		userID,
		token.AccessToken,
		token.RefreshToken,
		token.ExpiresAt,
		time.Now(),
	)
	return err
}

// Get retrieves a token for a user.
func (s *PersistentTokenStore) Get(userID string) (*Token, error) {
	query := `SELECT access_token, refresh_token, expires_at FROM spotify_tokens WHERE user_id = $1`

	var token Token
	err := s.db.QueryRowContext(context.Background(), query, userID).Scan(
		&token.AccessToken,
		&token.RefreshToken,
		&token.ExpiresAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotAuthenticated
	}
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// Delete removes a token for a user.
func (s *PersistentTokenStore) Delete(userID string) error {
	_, err := s.db.ExecContext(context.Background(), `DELETE FROM spotify_tokens WHERE user_id = $1`, userID)
	return err
}
