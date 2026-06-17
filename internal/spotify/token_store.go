package spotify

import (
	"sync"
	"time"
)

// Token holds OAuth2 token data.
type Token struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// IsExpired returns true if the token has expired.
func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// TokenStore is an in-memory store for Spotify tokens.
type TokenStore struct {
	mu     sync.RWMutex
	tokens map[string]*Token
}

// NewTokenStore creates a new in-memory token store.
func NewTokenStore() *TokenStore {
	return &TokenStore{tokens: make(map[string]*Token)}
}

// Save stores a token for a session.
func (s *TokenStore) Save(sessionID string, token *Token) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[sessionID] = token
	return nil
}

// Get retrieves a token for a session.
func (s *TokenStore) Get(sessionID string) (*Token, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tokens[sessionID]
	if !ok {
		return nil, ErrNotAuthenticated
	}
	return t, nil
}

// Delete removes a session token.
func (s *TokenStore) Delete(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, sessionID)
	return nil
}
