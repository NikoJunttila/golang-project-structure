package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/nikojunttila/community/internal/auth"
	"github.com/nikojunttila/community/internal/db"
	oath "github.com/nikojunttila/community/internal/oauth"
	userService "github.com/nikojunttila/community/internal/services/user"
)

// GoogleUserInfo represents the user data from Google
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

// StateStore interface for managing OAuth states
type StateStore interface {
	Set(state string, expiry time.Time) error
	Validate(state string) bool
	Remove(state string) error
}

// Simple in-memory state store (replace with Redis/DB in production)
type MemoryStateStore struct {
	states map[string]time.Time
}

func NewMemoryStateStore() *MemoryStateStore {
	return &MemoryStateStore{
		states: make(map[string]time.Time),
	}
}

func (m *MemoryStateStore) Set(state string, expiry time.Time) error {
	m.states[state] = expiry
	return nil
}

func (m *MemoryStateStore) Validate(state string) bool {
	expiry, exists := m.states[state]
	if !exists {
		return false
	}
	if time.Now().After(expiry) {
		delete(m.states, state)
		return false
	}
	return true
}

func (m *MemoryStateStore) Remove(state string) error {
	delete(m.states, state)
	return nil
}

// Global state store instance (initialize this properly in your app)
var stateStore StateStore = NewMemoryStateStore()

// generateState creates a cryptographically secure random state
func generateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetGoogleLogin redirects user to Google's OAuth2 login page
func GetGoogleLogin(w http.ResponseWriter, r *http.Request) {
	// Generate secure random state
	state, err := generateState()
	if err != nil {
		log.Warn().Msg("google login failed to generate state")
		http.Error(w, "Failed to generate state", http.StatusInternalServerError)
		return
	}

	// Store state with 10-minute expiry
	expiry := time.Now().Add(10 * time.Minute)
	if err := stateStore.Set(state, expiry); err != nil {
		log.Warn().Msg("google login failed to store state")
		http.Error(w, "Failed to store state", http.StatusInternalServerError)
		return
	}

	// Get OAuth config and generate auth URL
	cfg := oath.GoogleConfig()
	authURL := cfg.AuthCodeURL(state)

	// Set security headers
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// GetGoogleCallBack handles the OAuth callback from Google
func GetGoogleCallBack(w http.ResponseWriter, r *http.Request) {
	// Validate state parameter
	state := r.URL.Query().Get("state")
	if state == "" {
		log.Warn().Msg("google callback missing state param")
		http.Error(w, "Missing state parameter", http.StatusBadRequest)
		return
	}

	if !stateStore.Validate(state) {
		log.Warn().Msg("google callback invalid or expired state")
		http.Error(w, "Invalid or expired state", http.StatusBadRequest)
		return
	}

	// Remove used state
	stateStore.Remove(state)

	// Check for error parameter
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		errorDesc := r.URL.Query().Get("error_description")
		log.Error().Msgf("OAuth error: %s - %s", errParam, errorDesc)
		http.Error(w, fmt.Sprintf("OAuth error: %s - %s", errParam, errorDesc), http.StatusBadRequest)
		return
	}

	// Get authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		log.Error().Msg("google callback Missing auth code")
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	cfg := oath.GoogleConfig()
	token, err := cfg.Exchange(r.Context(), code)
	if err != nil {
		log.Error().Msg("google callback failed to exhance code for token")
		http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
		return
	}

	// Fetch user info using the token
	userInfo, err := fetchGoogleUserInfo(token.AccessToken)
	if err != nil {
		log.Error().Msg("google callback failed to fetch user info")
		http.Error(w, "Failed to fetch user information", http.StatusInternalServerError)
		return
	}

	// Validate that email is verified
	if !userInfo.VerifiedEmail {
		http.Error(w, "Email not verified", http.StatusBadRequest)
		return
	}
	exists, err := userService.CheckUserExists(r.Context(), userInfo.Email)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Internal server error", err)
		return
	}
	var user db.User
	if exists {
		user, err = db.Get().GetUserByEmail(r.Context(), userInfo.Email)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Internal server error", err)
		}
	} else {
		user, err = userService.CreateUser(r.Context(), "", userService.CreateUserParams{
			Email:     userInfo.Email,
			Name:      userInfo.Name,
			AvatarUrl: userInfo.Picture,
		}, userService.OauthCreate{
			IsOAuth:       true,
			EmailVerified: true,
			Provider:      "google",
			ProviderID:    "google",
		})
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
	}
	jwtToken := auth.MakeToken(user.LookupID)

	sendAuthDataToMainWindow(w, userInfo, jwtToken)
}

// New function to send both user data and JWT token to main window
func sendAuthDataToMainWindow(w http.ResponseWriter, userInfo *GoogleUserInfo, jwtToken string) {
	// Create response data that includes both user info and token
	authData := map[string]any{
		"user":    userInfo,
		"token":   jwtToken,
		"success": true,
	}

	// Sanitize data for JSON embedding
	authJSON, err := json.Marshal(authData)
	if err != nil {
		http.Error(w, "Failed to encode auth data", http.StatusInternalServerError)
		return
	}

	// Escape JSON for safe embedding in JavaScript
	escapedJSON := string(authJSON)

	// Set security headers
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'unsafe-inline'")

	// Get allowed origin from config
	allowedOrigin := "http://localhost:5173" // TODO: Move to config

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <title>OAuth Success</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
</head>
<body>
    <div id="status">
        <p>Authentication successful! Redirecting...</p>
        <p><small>If this window doesn't close automatically, you can close it manually.</small></p>
    </div>
    <script>
        (function() {
            try {
                if (window.opener) {
                    window.opener.postMessage(%s, %q);
                    window.close();
                } else {
                    document.getElementById('status').innerHTML = 
                        '<p>Authentication successful!</p><p>Please close this window and return to the application.</p>';
                }
            } catch (e) {
                console.error('Failed to send message to parent:', e);
                document.getElementById('status').innerHTML = 
                    '<p>Authentication successful!</p><p>Please close this window and return to the application.</p>';
            }
        })();
    </script>
</body>
</html>`, escapedJSON, allowedOrigin)

	fmt.Fprint(w, html)
}

// fetchGoogleUserInfo retrieves user information from Google API
func fetchGoogleUserInfo(accessToken string) (*GoogleUserInfo, error) {
	// Use userinfo endpoint with proper error handling
	apiURL := "https://www.googleapis.com/oauth2/v2/userinfo"
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	// Set authorization header (preferred over query parameter)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}
