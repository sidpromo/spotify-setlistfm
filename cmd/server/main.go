package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/sidpromo/spotify-setlistfm/internal/artist"
	"github.com/sidpromo/spotify-setlistfm/internal/config"
	"github.com/sidpromo/spotify-setlistfm/internal/database"
	"github.com/sidpromo/spotify-setlistfm/internal/middleware"
	"github.com/sidpromo/spotify-setlistfm/internal/orchestration"
	"github.com/sidpromo/spotify-setlistfm/internal/prediction"
	"github.com/sidpromo/spotify-setlistfm/internal/setlist"
	"github.com/sidpromo/spotify-setlistfm/internal/spotify"
)

var db *sql.DB

func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	return mux
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	if db != nil {
		if err := db.PingContext(r.Context()); err != nil {
			status = "degraded"
			slog.Error("health check: database unreachable", "error", err)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": status}) // #nosec G104
}

func main() {
	cfg := config.Load()

	timeout, _ := strconv.Atoi(cfg.HTTPTimeoutSec)
	if timeout == 0 {
		timeout = 10
	}
	httpClient := &http.Client{Timeout: time.Duration(timeout) * time.Second}

	// Database (optional — falls back to in-memory if DATABASE_URL not set)
	var jobRepo orchestration.JobRepository
	if cfg.DatabaseURL != "" {
		var err error
		db, err = database.Connect(cfg.DatabaseURL)
		if err != nil {
			slog.Error("failed to connect to database", "error", err)
			os.Exit(1)
		}
		defer db.Close()

		if err := database.RunMigrations(db, cfg.MigrationsPath); err != nil {
			slog.Error("failed to run migrations", "error", err)
			os.Exit(1)
		}

		// TODO: Use PostgresJobRepository as jobRepo once adapter is wired
		// For now, still using in-memory until orchestration.Job ↔ database.JobRow adapter is built
		jobRepo = orchestration.NewInMemoryJobStore()
		slog.Info("database connected, using PostgreSQL")
	} else {
		jobRepo = orchestration.NewInMemoryJobStore()
		slog.Warn("DATABASE_URL not set, using in-memory storage")
	}

	// Artist module
	artistClient := artist.NewSetlistFMClient(cfg.SetlistFMBaseURL, cfg.SetlistFMAPIKey, httpClient)
	artistSvc := artist.NewService(artistClient)

	// Setlist module
	setlistClient := setlist.NewSetlistFMClient(cfg.SetlistFMBaseURL, cfg.SetlistFMAPIKey, httpClient)
	setlistSvc := setlist.NewService(setlistClient, setlist.DefaultConfig())

	// Prediction module
	predictionSvc := prediction.NewService(prediction.DefaultConfig())

	// Spotify module
	tokenStore := spotify.NewTokenStore()
	spotifyClient := spotify.NewClient("https://api.spotify.com", httpClient)
	spotifySvc := spotify.NewService(spotifyClient, tokenStore)
	authHandler := spotify.NewAuthHandler(spotify.AuthConfig{
		ClientID:     cfg.SpotifyClientID,
		ClientSecret: cfg.SpotifyClientSecret,
		RedirectURI:  cfg.SpotifyRedirectURI,
	}, tokenStore, httpClient)

	// Orchestration module
	orchSvc := orchestration.NewService(setlistSvc, predictionSvc, spotifySvc, jobRepo)

	// Wire routes
	mux := newMux()
	artist.RegisterHandlers(mux, artistSvc)
	setlist.RegisterHandlers(mux, setlistSvc)
	spotify.RegisterHandlers(mux, authHandler)
	orchestration.RegisterHandlers(mux, orchSvc)

	handler := middleware.CORS(cfg.CORSAllowedOrigin)(mux)

	srv := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}
