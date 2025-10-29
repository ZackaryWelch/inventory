//go:build js && wasm

package app

import (
	"fmt"
	"log/slog"
	"syscall/js"

	"cogentcore.org/core/core"
	"github.com/nishiki/frontend/config"
	authAPI "github.com/nishiki/frontend/pkg/api/auth"
	collectionsAPI "github.com/nishiki/frontend/pkg/api/collections"
	apiCommon "github.com/nishiki/frontend/pkg/api/common"
	containersAPI "github.com/nishiki/frontend/pkg/api/containers"
	groupsAPI "github.com/nishiki/frontend/pkg/api/groups"
	"github.com/nishiki/frontend/pkg/types"
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

// Type aliases for convenience
type (
	User             = types.User
	AuthInfoResponse = types.AuthInfoResponse
	ClaimsInfo       = types.ClaimsInfo
	Group            = types.Group
	Collection       = types.Collection
)

// App holds the main application state
type App struct {
	config        *Config
	authService   *AuthService
	currentUser   *User
	groups        []Group
	collections   []Collection
	currentView   string
	isSignedIn    bool
	body          *core.Body // Reference to the body for dialogs
	mainContainer *core.Frame
	bottomMenu    *core.Frame // Reference to the bottom menu
	dialogState   *DialogState
	searchFilter  *SearchFilter
	logger        *slog.Logger
	// API clients
	apiClient         *apiCommon.Client
	authClient        *authAPI.Client
	groupsClient      *groupsAPI.Client
	collectionsClient *collectionsAPI.Client
	containersClient  *containersAPI.Client
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

	// Initialize API clients
	apiClient := apiCommon.NewClient(config.BackendURL, authService)
	authClient := authAPI.NewClient(apiClient, config.ClientID)
	groupsClient := groupsAPI.NewClient(apiClient)
	collectionsClient := collectionsAPI.NewClient(apiClient)
	containersClient := containersAPI.NewClient(apiClient)

	app := &App{
		config:            config,
		authService:       authService,
		currentView:       ViewLogin,
		isSignedIn:        false,
		logger:            logger,
		apiClient:         apiClient,
		authClient:        authClient,
		groupsClient:      groupsClient,
		collectionsClient: collectionsClient,
		containersClient:  containersClient,
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
