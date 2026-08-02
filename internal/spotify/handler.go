package spotify

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sidpromo/spotify-setlistfm/internal/auth"
	"github.com/sidpromo/spotify-setlistfm/internal/database"
)

// AuthConfig holds Spotify OAuth2 configuration.
type AuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// AuthHandler handles Spotify OAuth2 endpoints.
type AuthHandler struct {
	cfg        AuthConfig
	tokenStore *TokenStore
	userRepo   database.UserRepository
	jwtSvc     *auth.JWTService
	client     *Client
	httpClient *http.Client
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(cfg AuthConfig, tokenStore *TokenStore, userRepo database.UserRepository, jwtSvc *auth.JWTService, client *Client, httpClient *http.Client) *AuthHandler {
	return &AuthHandler{
		cfg:        cfg,
		tokenStore: tokenStore,
		userRepo:   userRepo,
		jwtSvc:     jwtSvc,
		client:     client,
		httpClient: httpClient,
	}
}

// RegisterHandlers registers Spotify auth and status routes.
func RegisterHandlers(mux *http.ServeMux, ah *AuthHandler) {
	mux.HandleFunc("GET /v1/auth/spotify/login", ah.Login)
	mux.HandleFunc("GET /v1/auth/spotify/callback", ah.Callback)
	mux.HandleFunc("GET /v1/auth/spotify/status", ah.Status)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	state := generateState()

	params := url.Values{
		"client_id":     {h.cfg.ClientID},
		"response_type": {"code"},
		"redirect_uri":  {h.cfg.RedirectURI},
		"scope":         {"playlist-modify-private playlist-modify-public user-read-email"},
		"state":         {state},
	}
	authURL := "https://accounts.spotify.com/authorize?" + params.Encode()
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	// Step 1: Exchange code for Spotify token
	spotifyToken, err := h.exchangeCode(code)
	if err != nil {
		slog.Error("token exchange failed", "error", err)
		http.Error(w, "failed to exchange token", http.StatusBadGateway)
		return
	}

	// Step 2: Get Spotify user profile
	spotifyUser, err := h.client.GetCurrentUser(r.Context(), spotifyToken.AccessToken)
	if err != nil {
		slog.Error("failed to get spotify profile", "error", err)
		http.Error(w, "failed to get user profile", http.StatusBadGateway)
		return
	}

	// Step 3: Find or create user in our DB
	user, err := h.findOrCreateUser(r, spotifyUser)
	if err != nil {
		slog.Error("failed to find/create user", "error", err)
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	// Step 4: Store Spotify token linked to our user ID
	_ = h.tokenStore.Save(user.ID, spotifyToken) // #nosec G104

	// Step 5: Generate JWT
	accessToken, err := h.jwtSvc.GenerateAccessToken(user.ID)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}
	refreshToken, err := h.jwtSvc.GenerateRefreshToken(user.ID)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{ // #nosec G104
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"userId":       user.ID,
		"displayName":  user.DisplayName,
	})
}

func (h *AuthHandler) Status(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	authenticated := false
	if userID != "" {
		_, err := h.tokenStore.Get(userID)
		authenticated = err == nil
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"authenticated": authenticated}) // #nosec G104
}

func (h *AuthHandler) findOrCreateUser(r *http.Request, spotifyUser *SpotifyUser) (*database.User, error) {
	// Try to find existing user
	existing, err := h.userRepo.GetBySpotifyID(r.Context(), spotifyUser.ID)
	if err == nil {
		return existing, nil
	}

	// Create new user
	now := time.Now()
	newUser := &database.User{
		ID:          generateUUID(),
		SpotifyID:   spotifyUser.ID,
		Email:       "", // Spotify may not provide email depending on scope
		DisplayName: spotifyUser.DisplayName,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := h.userRepo.Create(r.Context(), newUser); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	slog.Info("new user created", "userId", newUser.ID, "spotifyId", newUser.SpotifyID)
	return newUser, nil
}

func (h *AuthHandler) exchangeCode(code string) (*Token, error) {
	data := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {h.cfg.RedirectURI},
	}

	req, _ := http.NewRequest(http.MethodPost, "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(h.cfg.ClientID, h.cfg.ClientSecret)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed: %s", body)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	return &Token{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}, nil
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	// Set version 4 and variant bits
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
