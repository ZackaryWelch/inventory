//go:build !js || !wasm

package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/oauth2"
)

// AuthService handles authentication for desktop builds using the system browser
// and a local HTTP callback server.
type AuthService struct {
	config      *oauth2.Config
	logger      *slog.Logger
	redirectURL string

	mu    sync.RWMutex
	token *oauth2.Token
}

// TokenStorage is unused on desktop; tokens live in AuthService directly.
type TokenStorage struct{}

// NewAuthService creates a desktop authentication service.
func NewAuthService(config *Config, logger *slog.Logger) *AuthService {
	oauth2Config := &oauth2.Config{
		ClientID:    config.ClientID,
		RedirectURL: config.RedirectURL,
		Scopes:      []string{"openid", "profile", "email", "groups"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.AuthURL + "/application/o/authorize/",
			TokenURL: config.BackendURL + "/auth/token",
		},
	}
	return &AuthService{
		config:      oauth2Config,
		redirectURL: config.RedirectURL,
		logger:      logger,
	}
}

// DesktopLogin runs the full OAuth PKCE flow: opens the system browser, starts
// a local HTTP server to receive the callback, and exchanges the code for a token.
func (as *AuthService) DesktopLogin() (*oauth2.Token, error) {
	codeVerifier := generateRandomString(128)
	codeChallenge := generateCodeChallenge(codeVerifier)
	state := generateRandomString(32)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	server := &http.Server{Handler: mux}

	mux.HandleFunc("/auth/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			errCh <- fmt.Errorf("OAuth state mismatch")
			http.Error(w, "Authentication failed: state mismatch", http.StatusBadRequest)
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no authorization code in callback")
			http.Error(w, "Authentication failed: no code", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html><body><h2>Authentication successful!</h2>
<p>You can close this window and return to Nishiki.</p></body></html>`)
		codeCh <- code
	})

	// Derive host:port from the configured redirect URL
	listenAddr := as.redirectListenAddr()
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to start OAuth callback server on %s: %w", listenAddr, err)
	}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			as.logger.Error("OAuth callback server error", "error", err)
		}
	}()
	defer server.Close()

	authURL := as.config.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
	if err := openBrowser(authURL); err != nil {
		return nil, fmt.Errorf("failed to open browser: %w", err)
	}

	var code string
	select {
	case code = <-codeCh:
	case err = <-errCh:
		return nil, err
	}

	token, err := as.config.Exchange(context.Background(), code,
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
	)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	as.mu.Lock()
	as.token = token
	as.mu.Unlock()

	return token, nil
}

// InitiateLogin is a no-op on desktop; the full flow is in DesktopLogin.
func (as *AuthService) InitiateLogin() error { return nil }

// HandleCallback is not used on desktop; the callback is captured by DesktopLogin.
func (as *AuthService) HandleCallback() (*oauth2.Token, error) {
	return nil, fmt.Errorf("desktop: callback is handled by the local HTTP server in DesktopLogin")
}

func (as *AuthService) GetStoredToken() (*oauth2.Token, error) {
	as.mu.RLock()
	defer as.mu.RUnlock()
	if as.token == nil {
		return nil, fmt.Errorf("no token stored")
	}
	return as.token, nil
}

func (as *AuthService) IsTokenValid() bool {
	token, err := as.GetStoredToken()
	return err == nil && token.Valid()
}

func (as *AuthService) RefreshToken() (*oauth2.Token, error) {
	as.mu.RLock()
	token := as.token
	as.mu.RUnlock()

	if token == nil || token.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}
	newToken, err := as.config.TokenSource(context.Background(), token).Token()
	if err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}
	as.mu.Lock()
	as.token = newToken
	as.mu.Unlock()
	return newToken, nil
}

func (as *AuthService) GetAccessToken() (string, error) {
	token, err := as.GetStoredToken()
	if err != nil {
		return "", err
	}
	if !token.Valid() {
		refreshed, err := as.RefreshToken()
		if err != nil {
			return "", fmt.Errorf("token expired and refresh failed: %w", err)
		}
		return refreshed.AccessToken, nil
	}
	return token.AccessToken, nil
}

func (as *AuthService) ClearToken() {
	as.mu.Lock()
	as.token = nil
	as.mu.Unlock()
}

func (as *AuthService) Logout() error {
	as.ClearToken()
	base := as.config.Endpoint.AuthURL
	suffix := "/application/o/authorize/"
	logoutBase := base[:len(base)-len(suffix)]
	logoutURL := logoutBase + "/application/o/nishiki/end-session/"
	return openBrowser(logoutURL)
}

// redirectListenAddr returns "host:port" from the configured redirect URL.
func (as *AuthService) redirectListenAddr() string {
	// redirectURL is e.g. "http://localhost:3000/auth/callback"
	// Extract host:port portion.
	url := as.redirectURL
	// Strip scheme
	for _, prefix := range []string{"https://", "http://"} {
		if len(url) > len(prefix) && url[:len(prefix)] == prefix {
			url = url[len(prefix):]
			break
		}
	}
	// Take up to the first slash
	if i := strings.IndexByte(url, '/'); i >= 0 {
		url = url[:i]
	}
	return url
}

// openBrowser opens the system default browser to the given URL.
func openBrowser(url string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd, args = "open", []string{url}
	case "windows":
		cmd, args = "rundll32", []string{"url.dll,FileProtocolHandler", url}
	default:
		cmd, args = "xdg-open", []string{url}
	}
	return exec.Command(cmd, args...).Start()
}
