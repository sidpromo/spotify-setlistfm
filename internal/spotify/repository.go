package spotify

import "context"

// TokenRepository defines the persistence interface for Spotify tokens.
// Implementations: PostgresTokenRepository (production), InMemoryTokenStore (tests).
type TokenRepository interface {
	Save(ctx context.Context, userID string, token *Token) error
	Get(ctx context.Context, userID string) (*Token, error)
	Delete(ctx context.Context, userID string) error
}
