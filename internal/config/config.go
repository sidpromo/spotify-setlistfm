package config

import "os"

type Config struct {
	ServerPort               string
	CORSAllowedOrigin        string
	DatabaseURL              string
	RedisURL                 string
	MigrationsPath           string
	JWTSecret                string
	WorkerCount              string
	MaxQueueDepth            string
	RateLimitSearch          string
	RateLimitPlaylist        string
	CircuitBreakerThreshold  string
	CircuitBreakerCooldown   string
	SetlistFMAPIKey          string
	SetlistFMBaseURL         string
	HTTPTimeoutSec           string
	SpotifyClientID          string
	SpotifyClientSecret      string
	SpotifyRedirectURI       string
}

func Load() *Config {
	return &Config{
		ServerPort:               getEnv("SERVER_PORT", "8080"),
		CORSAllowedOrigin:       getEnv("CORS_ALLOWED_ORIGIN", "*"),
		DatabaseURL:             os.Getenv("DATABASE_URL"),
		RedisURL:                getEnv("REDIS_URL", "redis://localhost:6379"),
		MigrationsPath:          getEnv("MIGRATIONS_PATH", "migrations"),
		JWTSecret:               getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		WorkerCount:             getEnv("WORKER_COUNT", "3"),
		MaxQueueDepth:           getEnv("MAX_QUEUE_DEPTH", "100"),
		RateLimitSearch:         getEnv("RATE_LIMIT_SEARCH", "30"),
		RateLimitPlaylist:       getEnv("RATE_LIMIT_PLAYLIST", "5"),
		CircuitBreakerThreshold: getEnv("CIRCUIT_BREAKER_THRESHOLD", "5"),
		CircuitBreakerCooldown:  getEnv("CIRCUIT_BREAKER_COOLDOWN", "30"),
		SetlistFMAPIKey:         os.Getenv("SETLISTFM_API_KEY"),
		SetlistFMBaseURL:        getEnv("SETLISTFM_BASE_URL", "https://api.setlist.fm/rest"),
		HTTPTimeoutSec:          getEnv("HTTP_TIMEOUT_SECONDS", "10"),
		SpotifyClientID:         os.Getenv("SPOTIFY_CLIENT_ID"),
		SpotifyClientSecret:     os.Getenv("SPOTIFY_CLIENT_SECRET"),
		SpotifyRedirectURI:      os.Getenv("SPOTIFY_REDIRECT_URI"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
