package database

import (
	"context"
	"fmt"
	"sync"
)

// InMemoryUserRepository is an in-memory implementation of UserRepository for tests.
type InMemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]*User // keyed by ID
}

// NewInMemoryUserRepository creates a new in-memory user repository.
func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{users: make(map[string]*User)}
}

func (r *InMemoryUserRepository) Create(_ context.Context, user *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.ID] = user
	return nil
}

func (r *InMemoryUserRepository) GetByID(_ context.Context, id string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return u, nil
}

func (r *InMemoryUserRepository) GetBySpotifyID(_ context.Context, spotifyID string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.SpotifyID == spotifyID {
			return u, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}
