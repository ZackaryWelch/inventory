//go:build js && wasm

package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"syscall/js"
	"time"

	"golang.org/x/oauth2"
)

// AuthService handles authentication using OAuth2/OIDC with Authentik
type AuthService struct {
	config       *oauth2.Config
	state        string
	redirectURL  string
	codeVerifier string
	logger       *slog.Logger
	backendURL   string
}

// TokenStorage handles storing and retrieving tokens from localStorage
type TokenStorage struct{}

// NewAuthService creates a new authentication service
func NewAuthService(config *Config, logger *slog.Logger) *AuthService {
	oauth2Config := &oauth2.Config{
		ClientID:    config.ClientID,
		RedirectURL: config.RedirectURL,
		Scopes:      []string{"openid", "profile", "email", "groups"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.AuthURL + "/application/o/authorize/",
			TokenURL: config.BackendURL + "/auth/token", // Use backend proxy for token exchange
		},
	}

	return &AuthService{
		config:      oauth2Config,
		redirectURL: config.RedirectURL,
		state:       generateRandomString(32),
		logger:      logger,
	}
}

// InitiateLogin redirects the user to Authentik for authentication
func (as *AuthService) InitiateLogin() error {
	// Generate PKCE code verifier and challenge
	as.codeVerifier = generateRandomString(128)
	codeChallenge := generateCodeChallenge(as.codeVerifier)

	// Build authorization URL with PKCE
	authURL := as.config.AuthCodeURL(as.state,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	// Debug: Log the auth URL being generated
	as.logger.Debug("Generated OAuth2 authorization URL", "url", authURL)

	// Store state and code verifier in localStorage for callback verification
	as.storeInLocalStorage("auth_state", as.state)
	as.storeInLocalStorage("code_verifier", as.codeVerifier)

	// Redirect to Authentik
	return as.redirectTo(authURL)
}

// HandleCallback processes the OAuth2 callback and exchanges code for token via backend
func (as *AuthService) HandleCallback() (*oauth2.Token, error) {
	// Get current URL to extract code and state
	currentURL := js.Global().Get("window").Get("location").Get("href").String()
	parsedURL, err := url.Parse(currentURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse callback URL: %w", err)
	}

	// Extract code and state from URL parameters
	code := parsedURL.Query().Get("code")
	returnedState := parsedURL.Query().Get("state")

	if code == "" {
		return nil, fmt.Errorf("no authorization code in callback")
	}

	as.logger.Debug("Extracted callback parameters", "code", code[:8]+"...", "state", returnedState[:8]+"...")

	// Verify state parameter
	storedState, err := as.getFromLocalStorage("auth_state")
	if err != nil || storedState != returnedState {
		return nil, fmt.Errorf("invalid state parameter")
	}

	// Get stored code verifier
	codeVerifier, err := as.getFromLocalStorage("code_verifier")
	if err != nil {
		return nil, fmt.Errorf("code verifier not found")
	}

	as.logger.Debug("Exchanging authorization code for token via backend")

	// Exchange authorization code for token via backend proxy
	token, err := as.config.Exchange(
		context.Background(),
		code,
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	as.logger.Info("Token exchange successful", "expires", token.Expiry)

	// Store token in localStorage
	if err := as.storeToken(token); err != nil {
		return nil, fmt.Errorf("failed to store token: %w", err)
	}

	// Clean up temporary storage
	as.removeFromLocalStorage("auth_state")
	as.removeFromLocalStorage("code_verifier")

	return token, nil
}

// GetStoredToken retrieves the stored token from localStorage
func (as *AuthService) GetStoredToken() (*oauth2.Token, error) {
	tokenJSON, err := as.getFromLocalStorage("access_token")
	if err != nil {
		return nil, err
	}

	if tokenJSON == "" {
		return nil, fmt.Errorf("no token found")
	}

	var token oauth2.Token
	if err := json.Unmarshal([]byte(tokenJSON), &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	return &token, nil
}

// IsTokenValid checks if the current token is valid and not expired
func (as *AuthService) IsTokenValid() bool {
	token, err := as.GetStoredToken()
	if err != nil {
		return false
	}

	return token.Valid()
}

// RefreshToken attempts to refresh the current access token
func (as *AuthService) RefreshToken() (*oauth2.Token, error) {
	token, err := as.GetStoredToken()
	if err != nil {
		return nil, err
	}

	if token.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	// Create token source for automatic refresh
	tokenSource := as.config.TokenSource(context.Background(), token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Store the refreshed token
	if err := as.storeToken(newToken); err != nil {
		return nil, fmt.Errorf("failed to store refreshed token: %w", err)
	}

	return newToken, nil
}

// Logout clears all stored authentication data
func (as *AuthService) Logout() error {
	// Clear all auth-related data from localStorage
	as.removeFromLocalStorage("access_token")
	as.removeFromLocalStorage("auth_state")
	as.removeFromLocalStorage("code_verifier")

	// Optionally redirect to Authentik logout endpoint
	logoutURL := fmt.Sprintf("%s/application/o/nishiki/end-session/", as.config.Endpoint.AuthURL[:len(as.config.Endpoint.AuthURL)-len("/application/o/authorize/")])
	return as.redirectTo(logoutURL)
}

// GetAccessToken returns the current access token string
func (as *AuthService) GetAccessToken() (string, error) {
	token, err := as.GetStoredToken()
	if err != nil {
		return "", err
	}

	if !token.Valid() {
		// Try to refresh the token
		refreshedToken, err := as.RefreshToken()
		if err != nil {
			return "", fmt.Errorf("token expired and refresh failed: %w", err)
		}
		return refreshedToken.AccessToken, nil
	}

	return token.AccessToken, nil
}

// Helper methods for localStorage operations

func (as *AuthService) storeToken(token *oauth2.Token) error {
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return err
	}
	as.storeInLocalStorage("access_token", string(tokenJSON))
	return nil
}

func (as *AuthService) storeInLocalStorage(key, value string) {
	localStorage := js.Global().Get("localStorage")
	localStorage.Call("setItem", key, value)
}

func (as *AuthService) getFromLocalStorage(key string) (string, error) {
	localStorage := js.Global().Get("localStorage")
	value := localStorage.Call("getItem", key)
	if value.IsNull() {
		return "", fmt.Errorf("key %s not found in localStorage", key)
	}
	return value.String(), nil
}

func (as *AuthService) removeFromLocalStorage(key string) {
	localStorage := js.Global().Get("localStorage")
	localStorage.Call("removeItem", key)
}

func (as *AuthService) redirectTo(url string) error {
	js.Global().Get("window").Get("location").Set("href", url)
	return nil
}

// Utility functions

func generateRandomString(length int) string {
	// Generate cryptographically secure random string
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// Log the error and fallback to timestamp-based if crypto/rand fails in WebAssembly
		// Note: We use fmt here since this is a utility function without access to logger
		fmt.Printf("Warning: crypto/rand failed in WebAssembly, using fallback: %v\n", err)
		return fmt.Sprintf("%d_%d", time.Now().UnixNano(), length)
	}

	// Encode as base64url (RFC 4648 Section 5)
	return base64.RawURLEncoding.EncodeToString(bytes)
}

func generateCodeChallenge(verifier string) string {
	// Generate SHA256 hash of the code verifier
	hash := sha256.Sum256([]byte(verifier))
	// Encode as base64url without padding
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
