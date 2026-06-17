package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ebedber/setlist-spotify/internal/artist"
	"github.com/ebedber/setlist-spotify/internal/config"
	"github.com/ebedber/setlist-spotify/internal/middleware"
	"github.com/ebedber/setlist-spotify/internal/orchestration"
	"github.com/ebedber/setlist-spotify/internal/prediction"
	"github.com/ebedber/setlist-spotify/internal/setlist"
	"github.com/ebedber/setlist-spotify/internal/spotify"
)

func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	return mux
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	cfg := config.Load()

	timeout, _ := strconv.Atoi(cfg.HTTPTimeoutSec)
	if timeout == 0 {
		timeout = 10
	}
	httpClient := &http.Client{Timeout: time.Duration(timeout) * time.Second}

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
	jobStore := orchestration.NewJobStore()
	orchSvc := orchestration.NewService(setlistSvc, predictionSvc, spotifySvc, jobStore)

	// Wire routes
	mux := newMux()
	artist.RegisterHandlers(mux, artistSvc)
	setlist.RegisterHandlers(mux, setlistSvc)
	spotify.RegisterHandlers(mux, authHandler)
	orchestration.RegisterHandlers(mux, orchSvc)

	handler := middleware.CORS(cfg.CORSAllowedOrigin)(mux)

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: handler,
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
