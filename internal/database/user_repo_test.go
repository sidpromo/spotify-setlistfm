package database

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPostgresUserRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresUserRepository(db)
	now := time.Now()
	user := &User{
		ID:          "user-1",
		SpotifyID:   "spotify_abc123",
		Email:       "test@example.com",
		DisplayName: "Test User",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mock.ExpectExec("INSERT INTO users").
		WithArgs(user.ID, user.SpotifyID, user.Email, user.DisplayName, user.CreatedAt, user.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPostgresUserRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresUserRepository(db)
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "spotify_id", "email", "display_name", "created_at", "updated_at"}).
		AddRow("user-1", "spotify_abc", "test@example.com", "Test User", now, now)

	mock.ExpectQuery("SELECT .+ FROM users WHERE id").
		WithArgs("user-1").
		WillReturnRows(rows)

	user, err := repo.GetByID(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.SpotifyID != "spotify_abc" {
		t.Errorf("expected 'spotify_abc', got %q", user.SpotifyID)
	}
}

func TestPostgresUserRepository_GetBySpotifyID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresUserRepository(db)
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "spotify_id", "email", "display_name", "created_at", "updated_at"}).
		AddRow("user-1", "spotify_abc", "test@example.com", "Test", now, now)

	mock.ExpectQuery("SELECT .+ FROM users WHERE spotify_id").
		WithArgs("spotify_abc").
		WillReturnRows(rows)

	user, err := repo.GetBySpotifyID(context.Background(), "spotify_abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != "user-1" {
		t.Errorf("expected 'user-1', got %q", user.ID)
	}
}

func TestPostgresUserRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresUserRepository(db)
	mock.ExpectQuery("SELECT .+ FROM users WHERE id").
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetByID(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
