//go:build js && wasm

package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"syscall/js"
	"time"

	"cogentcore.org/core/core"
	"github.com/nishiki/frontend/config"
)

// View constants
const (
	ViewLogin       = "login"
	ViewCallback    = "callback"
	ViewDashboard   = "dashboard"
	ViewGroups      = "groups"
	ViewCollections = "collections"
	ViewProfile     = "profile"
)

// Use the shared config type
type Config = config.Config

// User represents a user in the system
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
}

// Group represents a group in the system
type Group struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Members     []User    `json:"members,omitempty"`
}

// Collection represents a collection of objects
type Collection struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ObjectType  string    `json:"object_type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// App holds the main application state
type App struct {
	config         *Config
	authService    *AuthService
	currentUser    *User
	groups         []Group
	collections    []Collection
	httpClient     *http.Client
	currentView    string
	isSignedIn     bool
	mainContainer  *core.Frame
	currentOverlay *core.Frame
	dialogState    *DialogState
	searchFilter   *SearchFilter
	logger         *slog.Logger
}


// NewApp creates a new application instance
func NewApp() *App {
	config := LoadConfig()

	// Create logger (console output for WebAssembly)
	logger := slog.New(slog.NewJSONHandler(consoleWriter{}, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Create authentication service
	authService := NewAuthService(config, logger)

	app := &App{
		config:      config,
		authService: authService,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		currentView: ViewLogin,
		isSignedIn:  false,
		logger:      logger,
	}

	// Initialize dialog state
	app.dialogState = &DialogState{}

	// Initialize search filter
	app.searchFilter = &SearchFilter{
		SortBy:        "name",
		SortDirection: "asc",
	}

	// Check authentication state on startup
	app.initializeAuthState()

	return app
}

// consoleWriter implements io.Writer to write logs to browser console in WebAssembly
type consoleWriter struct{}

func (cw consoleWriter) Write(p []byte) (n int, err error) {
	// In WebAssembly, we can write to console.log
	// Remove the newline at the end if present for cleaner console output
	msg := string(p)
	if len(msg) > 0 && msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}
	
	// Use fmt.Print for now, but this goes to console in WebAssembly
	fmt.Print(string(p))
	return len(p), nil
}

// initializeAuthState checks authentication state on app startup
func (app *App) initializeAuthState() {
	app.logger.Debug("Initializing auth state")
	
	// Check if we're on the callback URL
	if app.isCallbackURL() {
		app.logger.Info("Detected callback URL, handling auth callback")
		app.handleAuthCallback()
		return
	}

	// Check if we have a valid stored token
	if app.authService.IsTokenValid() {
		app.logger.Info("Valid token found, signing in user")
		app.isSignedIn = true
		app.currentView = ViewDashboard
		// Load user data
		go func() {
			if err := app.fetchCurrentUser(); err != nil {
				app.logger.Error("Error fetching user on startup", "error", err)
				app.isSignedIn = false
				app.currentView = ViewLogin
			}
		}()
	} else {
		app.logger.Info("No valid token found, showing login view")
		app.isSignedIn = false
		app.currentView = ViewLogin
	}
}

// isCallbackURL checks if the current URL is the OAuth callback URL
func (app *App) isCallbackURL() bool {
	currentURL := js.Global().Get("window").Get("location").Get("pathname").String()
	app.logger.Debug("Checking callback URL", "current_url", currentURL)
	isCallback := currentURL == "/auth/callback"
	app.logger.Debug("Callback URL check result", "is_callback", isCallback)
	return isCallback
}

// handleAuthCallback processes the OAuth callback
func (app *App) handleAuthCallback() {
	app.logger.Info("Starting auth callback handler")
	app.currentView = ViewCallback

	go func() {
		app.logger.Debug("Exchanging authorization code for token")
		token, err := app.authService.HandleCallback()
		if err != nil {
			app.logger.Error("Authentication callback failed", "error", err)
			app.isSignedIn = false
			app.currentView = ViewLogin
			app.showLoginView()
			return
		}

		// Authentication successful
		app.logger.Info("Authentication successful", "expires", token.Expiry)
		app.isSignedIn = true

		// Fetch user data
		app.logger.Debug("Fetching user data from backend")
		if err := app.fetchCurrentUser(); err != nil {
			app.logger.Error("Error fetching user after login", "error", err)
		}

		// Fetch initial data
		app.logger.Debug("Fetching initial data")
		app.fetchGroups()
		app.fetchCollections()

		// Show dashboard
		app.logger.Info("Showing dashboard after successful authentication")
		app.currentView = ViewDashboard
		app.showDashboardView()
	}()
}
