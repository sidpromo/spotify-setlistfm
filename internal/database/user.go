package database

import (
	"context"
	"time"
)

// User represents a user in the system.
// Users are identified by their Spotify account — no passwords stored.
type User struct {
	ID          string
	SpotifyID   string
	Email       string
	DisplayName string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UserRepository defines the persistence interface for users.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetBySpotifyID(ctx context.Context, spotifyID string) (*User, error)
}
