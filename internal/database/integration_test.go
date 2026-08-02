//go:build integration

package database_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/sidpromo/spotify-setlistfm/internal/database"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	ctx := context.Background()

	_, filename, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(filename), "..", "..", "migrations")

	pgContainer, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	t.Cleanup(func() { pgContainer.Terminate(ctx) })

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	db, err := database.Connect(connStr)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := database.RunMigrations(db, migrationsPath); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return db
}

func TestIntegration_UserRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := database.NewPostgresUserRepository(db)
	ctx := context.Background()

	// Create user
	user := &database.User{
		ID:          "u_integration_1",
		SpotifyID:   "spotify_int_test_1",
		Email:       "integration@test.com",
		DisplayName: "Integration Test",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Get by ID
	got, err := repo.GetByID(ctx, "u_integration_1")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Email != "integration@test.com" {
		t.Errorf("expected email 'integration@test.com', got %q", got.Email)
	}

	// Get by Spotify ID
	got, err = repo.GetBySpotifyID(ctx, "spotify_int_test_1")
	if err != nil {
		t.Fatalf("GetBySpotifyID failed: %v", err)
	}
	if got.ID != "u_integration_1" {
		t.Errorf("expected ID 'u_integration_1', got %q", got.ID)
	}

	// Not found
	_, err = repo.GetByID(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent user")
	}
}

func TestIntegration_JobRepository(t *testing.T) {
	db := setupTestDB(t)
	userRepo := database.NewPostgresUserRepository(db)
	jobRepo := database.NewPostgresJobRepository(db)
	ctx := context.Background()

	// Create user first (FK constraint)
	user := &database.User{
		ID:           "u_job_test",
		Email:        "jobtest@test.com",
		SpotifyID:   "spotify_test_user",
		DisplayName:  "Job Test",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	userRepo.Create(ctx, user)

	// Create job
	now := time.Now()
	job := &database.JobRow{
		ID:         "j_integration_1",
		UserID:     "u_job_test",
		Status:     "pending",
		ArtistMBID: "mbid-123",
		ArtistName: "Metallica",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := jobRepo.Create(ctx, job); err != nil {
		t.Fatalf("Create job failed: %v", err)
	}

	// Get job
	got, err := jobRepo.Get(ctx, "j_integration_1")
	if err != nil {
		t.Fatalf("Get job failed: %v", err)
	}
	if got.Status != "pending" {
		t.Errorf("expected 'pending', got %q", got.Status)
	}

	// Update job
	got.Status = "completed"
	got.Result = database.MarshalResult(map[string]string{"playlistUrl": "http://example.com"})
	if err := jobRepo.Update(ctx, got); err != nil {
		t.Fatalf("Update job failed: %v", err)
	}

	// Verify update
	got, _ = jobRepo.Get(ctx, "j_integration_1")
	if got.Status != "completed" {
		t.Errorf("expected 'completed', got %q", got.Status)
	}

	// List by user
	jobs, err := jobRepo.ListByUser(ctx, "u_job_test", 10, 0)
	if err != nil {
		t.Fatalf("ListByUser failed: %v", err)
	}
	if len(jobs) != 1 {
		t.Errorf("expected 1 job, got %d", len(jobs))
	}
}

func TestIntegration_TokenRepository(t *testing.T) {
	db := setupTestDB(t)
	userRepo := database.NewPostgresUserRepository(db)
	tokenRepo := database.NewPostgresTokenRepository(db)
	ctx := context.Background()

	// Create user first
	user := &database.User{
		ID:           "u_token_test",
		Email:        "tokentest@test.com",
		SpotifyID:   "spotify_token_user",
		DisplayName:  "Token Test",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	userRepo.Create(ctx, user)

	// Save token
	token := &database.TokenRow{
		UserID:       "u_token_test",
		AccessToken:  "access_123",
		RefreshToken: "refresh_456",
		ExpiresAt:    time.Now().Add(time.Hour),
	}
	if err := tokenRepo.Save(ctx, token); err != nil {
		t.Fatalf("Save token failed: %v", err)
	}

	// Get token
	got, err := tokenRepo.Get(ctx, "u_token_test")
	if err != nil {
		t.Fatalf("Get token failed: %v", err)
	}
	if got.AccessToken != "access_123" {
		t.Errorf("expected 'access_123', got %q", got.AccessToken)
	}

	// Upsert (save again with new values)
	token.AccessToken = "new_access"
	if err := tokenRepo.Save(ctx, token); err != nil {
		t.Fatalf("Upsert token failed: %v", err)
	}
	got, _ = tokenRepo.Get(ctx, "u_token_test")
	if got.AccessToken != "new_access" {
		t.Errorf("expected 'new_access', got %q", got.AccessToken)
	}

	// Delete
	if err := tokenRepo.Delete(ctx, "u_token_test"); err != nil {
		t.Fatalf("Delete token failed: %v", err)
	}
	_, err = tokenRepo.Get(ctx, "u_token_test")
	if err == nil {
		t.Error("expected error after delete")
	}
}
