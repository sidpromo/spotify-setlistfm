package database

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPostgresJobRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresJobRepository(db)
	now := time.Now()
	job := &JobRow{
		ID:         "j_test1",
		UserID:     "user-1",
		Status:     "pending",
		ArtistMBID: "mbid-1",
		ArtistName: "Metallica",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	mock.ExpectExec("INSERT INTO jobs").
		WithArgs(job.ID, job.UserID, job.Status, job.ArtistMBID, job.ArtistName,
			job.Result, job.Error, job.CreatedAt, job.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Create(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPostgresJobRepository_Get(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresJobRepository(db)
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "user_id", "status", "artist_mbid", "artist_name", "result", "error", "created_at", "updated_at"}).
		AddRow("j_test1", "user-1", "completed", "mbid-1", "Metallica", sql.NullString{String: `{"playlistId":"pl1"}`, Valid: true}, sql.NullString{}, now, now)

	mock.ExpectQuery("SELECT .+ FROM jobs WHERE id").
		WithArgs("j_test1").
		WillReturnRows(rows)

	job, err := repo.Get(context.Background(), "j_test1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.Status != "completed" {
		t.Errorf("expected 'completed', got %q", job.Status)
	}
	if job.ArtistName != "Metallica" {
		t.Errorf("expected 'Metallica', got %q", job.ArtistName)
	}
}

func TestPostgresJobRepository_GetNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresJobRepository(db)
	mock.ExpectQuery("SELECT .+ FROM jobs WHERE id").
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.Get(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPostgresJobRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresJobRepository(db)
	job := &JobRow{
		ID:     "j_test1",
		Status: "completed",
		Result: sql.NullString{String: `{"playlistId":"pl1"}`, Valid: true},
		Error:  sql.NullString{},
	}

	mock.ExpectExec("UPDATE jobs SET").
		WithArgs(job.Status, job.Result, job.Error, sqlmock.AnyArg(), job.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Update(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPostgresJobRepository_ListByUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresJobRepository(db)
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "user_id", "status", "artist_mbid", "artist_name", "result", "error", "created_at", "updated_at"}).
		AddRow("j_1", "user-1", "completed", "mbid-1", "Band A", sql.NullString{}, sql.NullString{}, now, now).
		AddRow("j_2", "user-1", "pending", "mbid-2", "Band B", sql.NullString{}, sql.NullString{}, now, now)

	mock.ExpectQuery("SELECT .+ FROM jobs WHERE user_id").
		WithArgs("user-1", 10, 0).
		WillReturnRows(rows)

	jobs, err := repo.ListByUser(context.Background(), "user-1", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}
	if jobs[0].ArtistName != "Band A" {
		t.Errorf("expected 'Band A', got %q", jobs[0].ArtistName)
	}
}

func TestMarshalResult(t *testing.T) {
	result := MarshalResult(map[string]string{"playlistId": "pl1"})
	if !result.Valid {
		t.Fatal("expected valid result")
	}
	if result.String == "" {
		t.Fatal("expected non-empty string")
	}

	nilResult := MarshalResult(nil)
	if nilResult.Valid {
		t.Fatal("expected invalid for nil input")
	}
}

func TestUnmarshalResult(t *testing.T) {
	s := sql.NullString{String: `{"playlistId":"pl1"}`, Valid: true}
	var result map[string]string
	err := UnmarshalResult(s, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["playlistId"] != "pl1" {
		t.Errorf("expected 'pl1', got %q", result["playlistId"])
	}

	// Test invalid (NULL)
	empty := sql.NullString{}
	var result2 map[string]string
	err = UnmarshalResult(empty, &result2)
	if err != nil {
		t.Fatalf("unexpected error for NULL: %v", err)
	}
}
