package spotify

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
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
	httpClient *http.Client
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(cfg AuthConfig, tokenStore *TokenStore, httpClient *http.Client) *AuthHandler {
	return &AuthHandler{cfg: cfg, tokenStore: tokenStore, httpClient: httpClient}
}

// RegisterHandlers registers Spotify auth and status routes.
func RegisterHandlers(mux *http.ServeMux, ah *AuthHandler) {
	mux.HandleFunc("GET /v1/auth/spotify/login", ah.Login)
	mux.HandleFunc("GET /v1/auth/spotify/callback", ah.Callback)
	mux.HandleFunc("GET /v1/auth/spotify/status", ah.Status)
}

const sessionCookieName = "spotify_session"

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	sessionID := generateSessionID()
	// TODO: Add Secure: true when serving over HTTPS in production, add SameSite: http.SameSiteLaxMode
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})

	params := url.Values{
		"client_id":     {h.cfg.ClientID},
		"response_type": {"code"},
		"redirect_uri":  {h.cfg.RedirectURI},
		"scope":         {"playlist-modify-private playlist-modify-public"},
		"state":         {sessionID},
	}
	authURL := "https://accounts.spotify.com/authorize?" + params.Encode()
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" || state == "" {
		http.Error(w, "missing code or state", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	token, err := h.exchangeCode(code)
	if err != nil {
		http.Error(w, "failed to exchange token", http.StatusBadGateway)
		return
	}

	_ = h.tokenStore.Save(state, token) // #nosec G104

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "authenticated"}) // #nosec G104
}

func (h *AuthHandler) Status(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"authenticated": false}) // #nosec G104
		return
	}

	_, err = h.tokenStore.Get(cookie.Value)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"authenticated": err == nil}) // #nosec G104
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

func generateSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
