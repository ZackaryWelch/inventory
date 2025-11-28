//go:build js && wasm

package app

import (
	"fmt"
	"image/color"
	"log/slog"
	"strings"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/nishiki/backend-go/app/http/response"
	"github.com/nishiki/frontend/config"
	authAPI "github.com/nishiki/frontend/pkg/api/auth"
	collectionsAPI "github.com/nishiki/frontend/pkg/api/collections"
	apiCommon "github.com/nishiki/frontend/pkg/api/common"
	containersAPI "github.com/nishiki/frontend/pkg/api/containers"
	groupsAPI "github.com/nishiki/frontend/pkg/api/groups"
	objectsAPI "github.com/nishiki/frontend/pkg/api/objects"
	"github.com/nishiki/frontend/ui/theme"
	"github.com/nishiki/frontend/ui/widgets"
)

// Type aliases
type Config = config.Config

// Type aliases for backend response types
type (
	User             = response.UserResponse
	AuthInfoResponse = response.AuthInfoResponse
	ClaimsInfo       = response.ClaimsInfo
	Group            = response.GroupResponse
	Collection       = response.CollectionResponse
	Container        = response.ContainerResponse
	Object           = response.ObjectResponse
)

// consoleWriter writes logs to browser console
type consoleWriter struct{}

func (cw consoleWriter) Write(p []byte) (n int, err error) {
	// Remove the newline at the end if present for cleaner console output
	msg := string(p)
	if len(msg) > 0 && msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}
	fmt.Print(string(p))
	return len(p), nil
}

// GioApp holds the Gio-based application state
type GioApp struct {
	config             *config.Config
	authService        *AuthService
	currentUser        *User
	groups             []Group
	collections        []Collection
	containers         []Container
	objects            []Object
	selectedCollection *Collection
	selectedGroup      *Group
	selectedContainer  *Container
	selectedObject     *Object
	currentView        ViewID
	isSignedIn         bool
	logger             *slog.Logger

	// Dialog state
	showGroupDialog      bool
	groupDialogMode      string // "create" or "edit"
	showDeleteConfirm    bool
	deleteGroupID        string
	showCollectionDialog bool
	collectionDialogMode string // "create" or "edit"
	showDeleteCollection bool
	deleteCollectionID   string
	showContainerDialog  bool
	containerDialogMode  string // "create" or "edit"
	showDeleteContainer  bool
	deleteContainerID    string
	showObjectDialog     bool
	objectDialogMode     string // "create" or "edit"
	showDeleteObject     bool
	deleteObjectID       string
	selectedObjectType    string
	selectedContainerType string
	selectedGroupID       *string
	selectedContainerID   *string

	// Import state
	showImportPreview bool
	importData        *ImportData
	importFilename    string

	// Gio-specific fields
	window *app.Window
	theme  *theme.NishikiTheme
	ops    chan Operation

	// API clients
	apiClient         *apiCommon.Client
	authClient        *authAPI.Client
	groupsClient      *groupsAPI.Client
	collectionsClient *collectionsAPI.Client
	containersClient  *containersAPI.Client
	objectsClient     *objectsAPI.Client

	// Widget state
	widgetState *WidgetState
}

// WidgetState holds all widget state for the application
type WidgetState struct {
	// Login view
	loginButton widget.Clickable

	// Dashboard navigation buttons
	groupsButton      widget.Clickable
	collectionsButton widget.Clickable
	profileButton     widget.Clickable
	searchButton      widget.Clickable

	// Profile view
	logoutButton widget.Clickable

	// Bottom menu buttons
	menuDashboard    widget.Clickable
	menuGroups       widget.Clickable
	menuCollections  widget.Clickable
	menuProfile      widget.Clickable

	// Groups view
	groupsCreateButton widget.Clickable
	groupsSearchField  widget.Editor
	groupsList         widget.List
	groupItems         []GroupItemState

	// Groups dialog widgets
	groupNameEditor        widget.Editor
	groupDescriptionEditor widget.Editor
	groupDialogSubmit      widget.Clickable
	groupDialogCancel      widget.Clickable

	// Collections view
	collectionsCreateButton widget.Clickable
	collectionsSearchField  widget.Editor
	collectionsList         widget.List
	collectionItems         []CollectionItemState

	// Collections dialog widgets
	collectionNameEditor     widget.Editor
	collectionLocationEditor widget.Editor
	collectionTagsEditor     widget.Editor
	collectionTypeButtons    map[string]*widget.Clickable
	collectionGroupButtons   map[string]*widget.Clickable
	collectionDialogSubmit   widget.Clickable
	collectionDialogCancel   widget.Clickable

	// Container/Object view
	containersSearchField widget.Editor
	containersList        widget.List
	containerItems        []ContainerItemState
	objectsSearchField    widget.Editor
	objectsList           widget.List
	objectItems           []ObjectItemState
	backToCollections     widget.Clickable
	createContainerButton widget.Clickable
	createObjectButton    widget.Clickable
	importButton          widget.Clickable
	importExecuteButton   widget.Clickable
	importCancelButton    widget.Clickable

	// Container dialog widgets
	containerNameEditor     widget.Editor
	containerLocationEditor widget.Editor
	containerTypeButtons    map[string]*widget.Clickable
	containerDialogSubmit   widget.Clickable
	containerDialogCancel   widget.Clickable

	// Object dialog widgets
	objectNameEditor        widget.Editor
	objectDescriptionEditor widget.Editor
	objectQuantityEditor    widget.Editor
	objectUnitEditor        widget.Editor
	objectDialogSubmit      widget.Clickable
	objectDialogCancel      widget.Clickable

	// Dialog instances
	collectionDialog *widgets.Dialog
	groupDialog      *widgets.Dialog
	deleteDialog     *widgets.Dialog
	containerDialog  *widgets.Dialog
	objectDialog     *widgets.Dialog
}

// GroupItemState holds widget state for a single group list item
type GroupItemState struct {
	clickable    widget.Clickable
	editButton   widget.Clickable
	deleteButton widget.Clickable
}

// CollectionItemState holds widget state for a single collection list item
type CollectionItemState struct {
	clickable    widget.Clickable
	viewButton   widget.Clickable
	editButton   widget.Clickable
	deleteButton widget.Clickable
}

// ContainerItemState holds widget state for a single container list item
type ContainerItemState struct {
	clickable    widget.Clickable
	editButton   widget.Clickable
	deleteButton widget.Clickable
}

// ObjectItemState holds widget state for a single object list item
type ObjectItemState struct {
	clickable    widget.Clickable
	editButton   widget.Clickable
	deleteButton widget.Clickable
}

// ViewID represents different views in the application
type ViewID int

const (
	ViewLoginGio ViewID = iota
	ViewCallbackGio
	ViewDashboardGio
	ViewGroupsGio
	ViewCollectionsGio
	ViewCollectionDetailGio
	ViewProfileGio
	ViewSearchGio
)

// Operation represents an async operation result
type Operation struct {
	Type string
	Data interface{}
	Err  error
}

// NewGioApp creates a new Gio-based application instance
func NewGioApp() *GioApp {
	cfg := LoadConfig()

	// Create logger (console output for WebAssembly)
	logger := slog.New(slog.NewJSONHandler(consoleWriter{}, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Create authentication service
	authService := NewAuthService(cfg, logger)

	// Initialize API clients
	apiClient := apiCommon.NewClient(cfg.BackendURL, authService)
	authClient := authAPI.NewClient(apiClient, cfg.ClientID)
	groupsClient := groupsAPI.NewClient(apiClient)
	collectionsClient := collectionsAPI.NewClient(apiClient)
	containersClient := containersAPI.NewClient(apiClient)
	objectsClient := objectsAPI.NewClient(apiClient)

	// Create Gio window
	w := new(app.Window)
	w.Option(app.Title("Nishiki - Inventory Management"))

	// Create Nishiki theme
	th := theme.NewTheme()

	// Initialize widget state with button maps and dialogs
	widgetState := &WidgetState{
		collectionTypeButtons:  make(map[string]*widget.Clickable),
		collectionGroupButtons: make(map[string]*widget.Clickable),
		containerTypeButtons:   make(map[string]*widget.Clickable),
		collectionDialog:       widgets.NewDialog(),
		groupDialog:            widgets.NewDialog(),
		deleteDialog:           widgets.NewDialog(),
		containerDialog:        widgets.NewDialog(),
		objectDialog:           widgets.NewDialog(),
	}

	gioApp := &GioApp{
		config:            cfg,
		authService:       authService,
		currentView:       ViewLoginGio,
		isSignedIn:        false,
		logger:            logger,
		window:            w,
		theme:             th,
		ops:               make(chan Operation, 10),
		apiClient:         apiClient,
		authClient:        authClient,
		groupsClient:      groupsClient,
		collectionsClient: collectionsClient,
		containersClient:  containersClient,
		objectsClient:     objectsClient,
		widgetState:       widgetState,
	}

	// Check authentication state on startup
	gioApp.initializeAuthState()

	return gioApp
}

// Run starts the Gio application event loop
func (ga *GioApp) Run() error {
	var ops op.Ops

	// Start with an initial invalidate to get the first frame
	go func() {
		ga.window.Invalidate()
	}()

	for {
		// Process any pending operations before handling window events
		select {
		case op := <-ga.ops:
			ga.handleOperation(op)
			ga.window.Invalidate()
		default:
			// No operations pending, continue to window event
		}

		// Handle window events
		e := ga.window.Event()
		switch e := e.(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			ga.render(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

// render renders the current view
func (ga *GioApp) render(gtx layout.Context) layout.Dimensions {
	// Paint background
	ga.paintBackground(gtx, theme.ColorWhite)

	// Use a stack to layer dialogs on top of views
	return layout.Stack{}.Layout(gtx,
		// Base view layer
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			switch ga.currentView {
			case ViewLoginGio:
				return ga.renderLoginViewSimple(gtx)
			case ViewCallbackGio:
				return ga.renderCallbackView(gtx)
			case ViewDashboardGio:
				return ga.renderDashboardView(gtx)
			case ViewGroupsGio:
				return ga.renderGroupsView(gtx)
			case ViewCollectionsGio:
				return ga.renderCollectionsView(gtx)
			case ViewCollectionDetailGio:
				return ga.renderCollectionDetailView(gtx)
			case ViewProfileGio:
				return ga.renderProfileView(gtx)
			default:
				return ga.renderLoginViewSimple(gtx)
			}
		}),

		// Dialog layer (rendered on top)
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			// Render group dialogs if in groups view
			if ga.currentView == ViewGroupsGio {
				if ga.showGroupDialog {
					return ga.renderGroupDialog(gtx)
				}
				if ga.showDeleteConfirm {
					return ga.renderDeleteConfirmDialog(gtx)
				}
			}
			// Render collection dialogs if in collections view
			if ga.currentView == ViewCollectionsGio {
				if ga.showCollectionDialog {
					return ga.renderCollectionDialog(gtx)
				}
				if ga.showDeleteCollection {
					return ga.renderDeleteCollectionDialog(gtx)
				}
			}
			return layout.Dimensions{}
		}),
	)
}

// paintBackground paints a solid color background
func (ga *GioApp) paintBackground(gtx layout.Context, bgColor color.NRGBA) {
	defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
	paint.ColorOp{Color: bgColor}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
}

// initializeAuthState checks authentication state on app startup
func (ga *GioApp) initializeAuthState() {
	ga.logger.Info("Initializing auth state on startup")

	// Check if we're on the callback URL
	if ga.isCallbackURL() {
		ga.logger.Info("Detected callback URL")
		// First check if we already have a valid token (e.g., user refreshed the callback page)
		if ga.authService.IsTokenValid() {
			ga.logger.Info("Valid token already exists, skipping callback and redirecting to dashboard")
			ga.isSignedIn = true
			ga.currentView = ViewDashboardGio
			// Redirect away from callback URL
			go func() {
				ga.redirectToPath("/")
				ga.loadUserData()
				ga.window.Invalidate()
			}()
			return
		}
		// No valid token, proceed with OAuth callback
		ga.logger.Info("No valid token, handling OAuth callback")
		ga.currentView = ViewCallbackGio
		ga.handleAuthCallback()
		return
	}

	// Check if we have a valid stored token
	token, err := ga.authService.GetStoredToken()
	if err != nil {
		ga.logger.Info("No stored token found", "error", err)
		ga.isSignedIn = false
		ga.currentView = ViewLoginGio
		return
	}

	ga.logger.Info("Stored token found", "valid", token.Valid(), "expiry", token.Expiry)

	if ga.authService.IsTokenValid() {
		ga.logger.Info("Valid token found, signing in user automatically")
		ga.isSignedIn = true
		ga.currentView = ViewDashboardGio
		// Load user data asynchronously
		go func() {
			ga.loadUserData()
			ga.window.Invalidate() // Trigger re-render after data loads
		}()
	} else {
		ga.logger.Info("Token expired, attempting refresh")
		// Try to refresh the token
		if _, err := ga.authService.RefreshToken(); err != nil {
			ga.logger.Warn("Token refresh failed, showing login", "error", err)
			ga.authService.ClearToken()
			ga.isSignedIn = false
			ga.currentView = ViewLoginGio
		} else {
			ga.logger.Info("Token refreshed successfully, signing in user")
			ga.isSignedIn = true
			ga.currentView = ViewDashboardGio
			go func() {
				ga.loadUserData()
				ga.window.Invalidate()
			}()
		}
	}
}

// isCallbackURL checks if the current URL is the OAuth callback URL
func (ga *GioApp) isCallbackURL() bool {
	path := getCurrentPath()
	ga.logger.Debug("Checking callback URL", "path", path)
	// Check if path contains /auth/callback
	return strings.Contains(path, "/auth/callback")
}

// handleLogin initiates OAuth login flow
func (ga *GioApp) handleLogin() {
	ga.logger.Info("Initiating login")
	if err := ga.authService.InitiateLogin(); err != nil {
		ga.logger.Error("Failed to initiate login", "error", err)
	}
}

// handleAuthCallback processes the OAuth callback
func (ga *GioApp) handleAuthCallback() {
	ga.logger.Info("Starting auth callback handler")
	ga.currentView = ViewCallbackGio

	go func() {
		ga.logger.Debug("Exchanging authorization code for token")
		token, err := ga.authService.HandleCallback()
		if err != nil {
			ga.logger.Error("Authentication callback failed", "error", err)
			ga.isSignedIn = false
			ga.currentView = ViewLoginGio
			ga.window.Invalidate()
			return
		}

		// Authentication successful
		ga.logger.Info("Authentication successful", "expires", token.Expiry)
		ga.isSignedIn = true

		// Load user data
		ga.loadUserData()

		// Show dashboard
		ga.logger.Info("Showing dashboard after successful authentication")
		ga.currentView = ViewDashboardGio
		ga.window.Invalidate()
	}()
}

// renderCallbackView renders a loading message during OAuth callback
func (ga *GioApp) renderCallbackView(gtx layout.Context) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		label := material.H5(ga.theme.Theme, "Authenticating...")
		label.Color = theme.ColorTextSecondary
		return label.Layout(gtx)
	})
}

// loadUserData fetches initial user data after authentication
func (ga *GioApp) loadUserData() {
	// Fetch user data first
	ga.logger.Debug("Fetching user data from backend")
	if err := ga.fetchCurrentUser(); err != nil {
		ga.logger.Error("Error fetching user after login", "error", err)
		return
	}
	// Groups and collections will be fetched after user is loaded (see handleOperation)
}

// fetchCurrentUser gets the current user from the backend
func (ga *GioApp) fetchCurrentUser() error {
	go func() {
		authInfo, err := ga.authClient.GetCurrentUser()
		if err != nil {
			ga.logger.Error("Failed to fetch current user", "error", err)
			ga.ops <- Operation{Type: "user_load_failed", Err: err}
			return
		}
		ga.logger.Info("Current user fetched", "user_id", authInfo.User.ID, "name", authInfo.User.Name)
		ga.ops <- Operation{Type: "user_loaded", Data: &authInfo.User}
	}()
	return nil
}

// fetchGroups gets the user's groups from the backend
func (ga *GioApp) fetchGroups() {
	go func() {
		groups, err := ga.groupsClient.List()
		if err != nil {
			ga.logger.Error("Failed to fetch groups", "error", err)
			ga.ops <- Operation{Type: "groups_load_failed", Err: err}
			return
		}
		ga.logger.Debug("Groups fetched", "count", len(groups))
		ga.ops <- Operation{Type: "groups_loaded", Data: groups}
	}()
}

// fetchCollections gets the user's collections from the backend
func (ga *GioApp) fetchCollections() {
	go func() {
		if ga.currentUser == nil {
			ga.logger.Error("Cannot fetch collections: no current user")
			return
		}

		collections, err := ga.collectionsClient.List(ga.currentUser.ID)
		if err != nil {
			ga.logger.Error("Failed to fetch collections", "error", err)
			ga.ops <- Operation{Type: "collections_load_failed", Err: err}
			return
		}
		ga.logger.Debug("Collections fetched", "count", len(collections))
		ga.ops <- Operation{Type: "collections_loaded", Data: collections}
	}()
}

// handleOperation handles async operations
func (ga *GioApp) handleOperation(op Operation) {
	ga.logger.Debug("Handling operation", "type", op.Type)

	if op.Err != nil {
		ga.logger.Error("Operation failed", "type", op.Type, "error", op.Err)
		return
	}

	// Handle different operation types
	switch op.Type {
	case "user_loaded":
		if user, ok := op.Data.(*User); ok {
			ga.currentUser = user
			ga.logger.Info("User loaded in state", "user_id", user.ID, "name", user.Name)
			// Now that user is loaded, fetch groups and collections
			ga.fetchGroups()
			ga.fetchCollections()
		}
	case "user_load_failed":
		ga.logger.Error("User load failed on refresh, clearing auth", "error", op.Err)
		// Token might be invalid on backend side, clear it and show login
		// Just clear the token, don't redirect to logout URL
		ga.authService.ClearToken()
		ga.isSignedIn = false
		ga.currentUser = nil
		ga.currentView = ViewLoginGio
	case "groups_loaded":
		if groups, ok := op.Data.([]Group); ok {
			ga.groups = groups
			ga.logger.Info("Groups loaded in state", "count", len(groups))
		}
	case "collections_loaded":
		if collections, ok := op.Data.([]Collection); ok {
			ga.collections = collections
			ga.logger.Info("Collections loaded in state", "count", len(collections))
		}
	}
}
