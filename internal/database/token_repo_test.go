package database

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPostgresTokenRepository_Save(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresTokenRepository(db)
	token := &TokenRow{
		UserID:       "user-1",
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_456",
		ExpiresAt:    time.Now().Add(time.Hour),
	}

	mock.ExpectExec("INSERT INTO spotify_tokens").
		WithArgs(token.UserID, token.AccessToken, token.RefreshToken, token.ExpiresAt, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Save(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPostgresTokenRepository_Get(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresTokenRepository(db)
	now := time.Now()

	rows := sqlmock.NewRows([]string{"user_id", "access_token", "refresh_token", "expires_at", "updated_at"}).
		AddRow("user-1", "at_123", "rt_456", now.Add(time.Hour), now)

	mock.ExpectQuery("SELECT .+ FROM spotify_tokens WHERE user_id").
		WithArgs("user-1").
		WillReturnRows(rows)

	token, err := repo.Get(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "at_123" {
		t.Errorf("expected 'at_123', got %q", token.AccessToken)
	}
	if token.RefreshToken != "rt_456" {
		t.Errorf("expected 'rt_456', got %q", token.RefreshToken)
	}
}

func TestPostgresTokenRepository_GetNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresTokenRepository(db)
	mock.ExpectQuery("SELECT .+ FROM spotify_tokens WHERE user_id").
		WithArgs("unknown").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.Get(context.Background(), "unknown")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPostgresTokenRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresTokenRepository(db)
	mock.ExpectExec("DELETE FROM spotify_tokens WHERE user_id").
		WithArgs("user-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Delete(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
