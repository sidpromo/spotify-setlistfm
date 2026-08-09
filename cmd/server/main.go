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
	"github.com/sidpromo/spotify-setlistfm/internal/auth"
	"github.com/sidpromo/spotify-setlistfm/internal/cache"
	"github.com/sidpromo/spotify-setlistfm/internal/config"
	"github.com/sidpromo/spotify-setlistfm/internal/database"
	"github.com/sidpromo/spotify-setlistfm/internal/middleware"
	"github.com/sidpromo/spotify-setlistfm/internal/orchestration"
	"github.com/sidpromo/spotify-setlistfm/internal/prediction"
	"github.com/sidpromo/spotify-setlistfm/internal/setlist"
	"github.com/sidpromo/spotify-setlistfm/internal/spotify"
)

var db *sql.DB
var redisClient *cache.Redis

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
	if redisClient != nil {
		if err := redisClient.Ping(r.Context()); err != nil {
			if status == "ok" {
				status = "degraded"
			}
			slog.Error("health check: redis unreachable", "error", err)
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

	// JWT service
	jwtSvc := auth.NewJWTService(cfg.JWTSecret, 15*time.Minute, 7*24*time.Hour)

	// Database (optional — falls back to in-memory if DATABASE_URL not set)
	var jobRepo orchestration.JobRepository
	var userRepo database.UserRepository

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

		jobRepo = orchestration.NewInMemoryJobStore() // TODO: wire PostgresJobRepository adapter
		userRepo = database.NewPostgresUserRepository(db)
		slog.Info("database connected, using PostgreSQL")
	} else {
		jobRepo = orchestration.NewInMemoryJobStore()
		userRepo = database.NewInMemoryUserRepository()
		slog.Warn("DATABASE_URL not set, using in-memory storage")
	}

	// Artist module
	artistClient := artist.NewSetlistFMClient(cfg.SetlistFMBaseURL, cfg.SetlistFMAPIKey, httpClient)
	artistSvc := artist.NewService(artistClient)

	// Setlist module
	setlistClient := setlist.NewSetlistFMClient(cfg.SetlistFMBaseURL, cfg.SetlistFMAPIKey, httpClient)
	setlistSvc := setlist.NewService(setlistClient, setlist.DefaultConfig())

	// Redis cache (optional — falls back to no caching if REDIS_URL not reachable)
	var redisCache *cache.Redis
	if cfg.RedisURL != "" {
		var err error
		redisCache, err = cache.Connect(cfg.RedisURL)
		if err != nil {
			slog.Warn("redis unavailable, running without cache", "error", err)
		} else {
			redisClient = redisCache
			defer redisCache.Close()
		}
	}

	// Wrap services with cache if available
	cachedArtistSvc := cache.NewCachedArtistService(artistSvc, redisCache)
	cachedSetlistSvc := cache.NewCachedSetlistService(setlistSvc, redisCache)

	// Prediction module
	predictionSvc := prediction.NewService(prediction.DefaultConfig())

	// Spotify module
	var tokenStore spotify.TokenStorer
	if db != nil {
		tokenStore = spotify.NewPersistentTokenStore(db)
		slog.Info("using persistent token store (PostgreSQL)")
	} else {
		tokenStore = spotify.NewTokenStore()
		slog.Warn("using in-memory token store (tokens lost on restart)")
	}
	spotifyClient := spotify.NewClient("https://api.spotify.com", httpClient)
	spotifySvc := spotify.NewService(spotifyClient, tokenStore)
	authHandler := spotify.NewAuthHandler(spotify.AuthConfig{
		ClientID:     cfg.SpotifyClientID,
		ClientSecret: cfg.SpotifyClientSecret,
		RedirectURI:  cfg.SpotifyRedirectURI,
		FrontendURL:  cfg.FrontendURL,
	}, tokenStore, userRepo, jwtSvc, spotifyClient, httpClient)

	// Orchestration module (uses cached setlist service)
	// When Redis queue is available, jobs go through the worker pool.
	// Without it, falls back to direct goroutine execution.
	orchSvc := orchestration.NewService(cachedSetlistSvc, predictionSvc, spotifySvc, jobRepo, nil)

	// Wire routes
	mux := newMux()
	artist.RegisterHandlers(mux, cachedArtistSvc)
	setlist.RegisterHandlers(mux, setlistSvc)
	spotify.RegisterHandlers(mux, authHandler)

	// Protected routes (require JWT)
	authMw := auth.Middleware(jwtSvc)
	protectedMux := http.NewServeMux()
	orchestration.RegisterHandlers(protectedMux, orchSvc)
	mux.Handle("/v1/playlists", authMw(protectedMux))
	mux.Handle("/v1/playlists/", authMw(protectedMux))

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
