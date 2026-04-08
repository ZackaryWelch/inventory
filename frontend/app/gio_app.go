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

	"github.com/nishiki/backend/app/http/response"

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
	User               = response.UserResponse
	AuthInfoResponse   = response.AuthInfoResponse
	ClaimsInfo         = response.ClaimsInfo
	Group              = response.GroupResponse
	Collection         = response.CollectionResponse
	Container          = response.ContainerResponse
	Object             = response.ObjectResponse
	PropertySchema     = response.PropertySchemaResponse
	PropertyDefinition = response.PropertyDefinitionResponse
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

	// Login error message shown on the login screen after auth failures
	loginErrorMsg string

	// Dialog state
	showGroupDialog           bool
	groupDialogMode           string // "create" or "edit"
	showDeleteConfirm         bool
	deleteGroupID             string
	showCollectionDialog      bool
	collectionDialogMode      string // "create" or "edit"
	showDeleteCollection      bool
	deleteCollectionID        string
	showDeleteCollectionError bool
	deleteCollectionErrorMsg  string
	showContainerDialog       bool
	containerDialogMode       string // "create" or "edit"
	showDeleteContainer       bool
	deleteContainerID         string
	showObjectDialog          bool
	objectDialogMode          string // "create" or "edit"
	showDeleteObject          bool
	deleteObjectID            string
	selectedObjectType        string
	selectedContainerType     string
	selectedGroupID           *string
	selectedContainerID       *string
	selectedParentContainerID *string // nil = no parent (root), pointer to "" = explicitly clearing parent

	// Group members dialog state
	showMembersDialog bool
	groupMembersOf    *Group
	groupMembers      []User
	knownUsers        []User

	// Join group dialog state
	showJoinGroupDialog bool

	// Schema editor state
	showSchemaDialog bool

	// Import state
	showImportPreview    bool
	importData           *ImportData
	importFilename       string
	importNameColumn     string
	importLocationColumn *string // nil = no location column (automatic distribution)
	importRunning        bool
	importResult         *importResult

	// Container display mode in collection detail
	showContainersPanel bool   // whether container column is visible
	containerViewMode   string // "split" (side-by-side) or "grouped" (objects grouped by container)

	// Grouped-text filter state (property key → selected value; empty = "All")
	activeGroupedTextFilters map[string]string

	// Render caches — invalidated when underlying data changes (see invalidateObjectCaches)
	cachedGroupedTextValues map[string][]string // collectGroupedTextValues result
	cachedGroupedTextValid  bool
	cachedPropertyDefMap    map[string]*PropertyDefinition // key → def lookup
	cachedPropertyDefValid  bool

	// Filtered list caches — invalidated when search/filter inputs change
	cachedFilteredObjects     []Object
	cachedFilteredObjIndices  []int
	cachedObjSearchQuery      string
	cachedObjFilters          map[string]string // snapshot of activeGroupedTextFilters
	cachedObjDataLen          int               // len(ga.objects) when cache was built
	cachedFilteredContainers  []Container
	cachedFilteredContIndices []int
	cachedContSearchQuery     string
	cachedContDataLen         int

	// Gio-specific fields
	window *app.Window
	theme  *theme.NishikiTheme
	ops    chan func()

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
	menuDashboard   widget.Clickable
	menuGroups      widget.Clickable
	menuCollections widget.Clickable
	menuProfile     widget.Clickable

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
	collectionNameEditor         widget.Editor
	collectionLocationEditor     widget.Editor
	collectionTagsEditor         widget.Editor
	collectionTypeButtons        map[string]*widget.Clickable
	collectionGroupButtons       map[string]*widget.Clickable
	collectionDialogSubmit       widget.Clickable
	collectionDialogCancel       widget.Clickable
	collectionErrorDialogDismiss widget.Clickable
	collectionErrorDialog        *widgets.Dialog

	// Container/Object view
	toggleContainersButton widget.Clickable
	containerViewSplitBtn  widget.Clickable
	containerViewGroupBtn  widget.Clickable
	containersSearchField  widget.Editor
	containersList         widget.List
	containerItems         []ContainerItemState
	objectsSearchField     widget.Editor
	objectsList            widget.List
	objectItems            []ObjectItemState
	backToCollections      widget.Clickable
	createContainerButton  widget.Clickable
	createObjectButton     widget.Clickable
	importButton           widget.Clickable
	importExecuteButton    widget.Clickable
	importCancelButton     widget.Clickable
	importDialogList       widget.List
	importPreviewList      widget.List

	// Import column mapping
	importNameColumnButtons     map[string]*widget.Clickable
	importLocationColumnButtons map[string]*widget.Clickable
	importInferSchemaCheck      widget.Bool

	// Grouped-text filter chips (key = "propKey||value")
	groupedTextFilterButtons map[string]*widget.Clickable

	// Containers page
	containersPageButton widget.Clickable
	containersBackButton widget.Clickable
	containerDetailList  widget.List

	// Container dialog widgets
	containerNameEditor     widget.Editor
	containerLocationEditor widget.Editor
	containerTypeButtons    map[string]*widget.Clickable
	parentContainerButtons  map[string]*widget.Clickable
	containerDialogSubmit   widget.Clickable
	containerDialogCancel   widget.Clickable

	// Object dialog widgets
	objectNameEditor        widget.Editor
	objectDescriptionEditor widget.Editor
	objectQuantityEditor    widget.Editor
	objectUnitEditor        widget.Editor
	objectDialogSubmit      widget.Clickable
	objectDialogCancel      widget.Clickable
	objectContainerButtons  map[string]*widget.Clickable
	objectSchemaList        widget.List
	objectPropertyEditors   map[string]*widget.Editor
	objectPropertyBools     map[string]*widget.Bool

	// Group members dialog
	membersDialog       *widgets.Dialog
	membersDialogClose  widget.Clickable
	membersAddButton    widget.Clickable
	memberUserIDEditor  widget.Editor
	memberSearchEditor  widget.Editor
	knownUserClickables map[string]*widget.Clickable
	memberItems         []MemberItemState
	membersList         widget.List
	knownUsersList      widget.List

	// Join group dialog
	joinGroupButton widget.Clickable
	joinGroupDialog *widgets.Dialog
	joinGroupClose  widget.Clickable
	joinHashEditor  widget.Editor

	// Schema editor dialog
	editSchemaButton   widget.Clickable
	schemaDialog       *widgets.Dialog
	schemaDialogSubmit widget.Clickable
	schemaDialogCancel widget.Clickable
	schemaAddRowButton widget.Clickable
	schemaRows         []SchemaRowState
	schemaList         widget.List

	// Dialog instances
	collectionDialog *widgets.Dialog
	groupDialog      *widgets.Dialog
	deleteDialog     *widgets.Dialog
	containerDialog  *widgets.Dialog
	objectDialog     *widgets.Dialog
}

// GroupItemState holds widget state for a single group list item
type GroupItemState struct {
	clickable     widget.Clickable
	editButton    widget.Clickable
	deleteButton  widget.Clickable
	membersButton widget.Clickable
}

// MemberItemState holds widget state for a single group member row
type MemberItemState struct {
	removeButton widget.Clickable
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

// SchemaRowState holds widget state for a single schema definition row
type SchemaRowState struct {
	keyEditor         widget.Editor
	displayNameEditor widget.Editor
	requiredCheck     widget.Bool
	deleteButton      widget.Clickable
	typeButtons       map[string]*widget.Clickable
	selectedType      string
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
	ViewContainersGio
	ViewProfileGio
	ViewSearchGio
)

// do schedules a state mutation from a goroutine. The mutation is applied
// inside the next FrameEvent handler, before rendering, so the frame always
// sees fresh state. Invalidate wakes the blocked window.Event() call.
func (ga *GioApp) do(fn func()) {
	ga.ops <- fn
	ga.window.Invalidate()
}

// drainOps applies all pending state mutations queued via do().
// Called at the start of each frame so rendering always sees up-to-date state.
func (ga *GioApp) drainOps() {
	for {
		select {
		case fn := <-ga.ops:
			fn()
		default:
			return
		}
	}
}

// NewGioApp creates a new Gio-based application instance
func NewGioApp() *GioApp {
	cfg := LoadConfig()

	// Create logger (console output for WebAssembly)
	logger := slog.New(slog.NewJSONHandler(consoleWriter{}, &slog.HandlerOptions{
		Level: slog.LevelInfo,
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
		collectionTypeButtons:       make(map[string]*widget.Clickable),
		collectionGroupButtons:      make(map[string]*widget.Clickable),
		containerTypeButtons:        make(map[string]*widget.Clickable),
		importNameColumnButtons:     make(map[string]*widget.Clickable),
		importLocationColumnButtons: make(map[string]*widget.Clickable),
		groupedTextFilterButtons:    make(map[string]*widget.Clickable),
		importDialogList:            widget.List{List: layout.List{Axis: layout.Vertical}},
		importPreviewList:           widget.List{List: layout.List{Axis: layout.Vertical}},
		collectionDialog:            widgets.NewDialog(),
		groupDialog:                 widgets.NewDialog(),
		deleteDialog:                widgets.NewDialog(),
		collectionErrorDialog:       widgets.NewDialog(),
		containerDialog:             widgets.NewDialog(),
		objectDialog:                widgets.NewDialog(),
		schemaDialog:                widgets.NewDialog(),
		membersDialog:               widgets.NewDialog(),
		joinGroupDialog:             widgets.NewDialog(),
		knownUserClickables:         make(map[string]*widget.Clickable),
	}

	gioApp := &GioApp{
		config:            cfg,
		authService:       authService,
		currentView:       ViewLoginGio,
		isSignedIn:        false,
		logger:            logger,
		window:            w,
		theme:             th,
		ops:               make(chan func(), 10),
		apiClient:         apiClient,
		authClient:        authClient,
		groupsClient:      groupsClient,
		collectionsClient: collectionsClient,
		containersClient:  containersClient,
		objectsClient:     objectsClient,
		widgetState:       widgetState,
	}

	// Handle session expiry: any API call that can't obtain a token or receives a 401
	// will trigger this callback to clear local state and return to the login screen.
	apiClient.OnAuthError = func() {
		gioApp.do(gioApp.handleSessionExpired)
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
		e := ga.window.Event()
		switch e := e.(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			ga.drainOps()
			gtx := app.NewContext(&ops, e)
			ga.render(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

// render renders the current view
func (ga *GioApp) render(gtx layout.Context) layout.Dimensions {
	// Paint background
	ga.paintBackground(gtx, theme.ColorBackground)

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
			case ViewContainersGio:
				return ga.renderContainersPageView(gtx)
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
				if ga.showMembersDialog {
					return ga.renderMembersDialog(gtx)
				}
				if ga.showJoinGroupDialog {
					return ga.renderJoinGroupDialog(gtx)
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
				if ga.showDeleteCollectionError {
					return ga.renderDeleteCollectionErrorDialog(gtx)
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
			ga.loginErrorMsg = "Your session has expired. Please sign in again."
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
			ga.loginErrorMsg = "Sign in failed. Please try again."
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
	gtx.Constraints.Min = gtx.Constraints.Max
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
	// Groups and collections will be fetched after user is loaded (see fetchCurrentUser)
}

// fetchCurrentUser gets the current user from the backend
func (ga *GioApp) fetchCurrentUser() error {
	go func() {
		authInfo, err := ga.authClient.GetCurrentUser()
		if err != nil {
			ga.logger.Error("Failed to fetch current user", "error", err)
			ga.do(func() {
				// Token might be invalid, clear it and show login
				ga.authService.ClearToken()
				ga.isSignedIn = false
				ga.currentUser = nil
				ga.currentView = ViewLoginGio
			})
			return
		}
		ga.logger.Info("Current user fetched", "user_id", authInfo.User.ID, "name", authInfo.User.Name)
		user := authInfo.User
		ga.do(func() {
			ga.currentUser = &user
			ga.logger.Info("User loaded in state", "user_id", user.ID, "name", user.Name)
			ga.fetchGroups()
			ga.fetchCollections()
		})
	}()
	return nil
}

// fetchGroups gets the user's groups from the backend
func (ga *GioApp) fetchGroups() {
	go func() {
		groups, err := ga.groupsClient.List()
		if err != nil {
			ga.logger.Error("Failed to fetch groups", "error", err)
			return
		}
		ga.do(func() {
			ga.groups = groups
			ga.logger.Info("Groups loaded in state", "count", len(groups))
		})
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
			return
		}
		ga.do(func() {
			ga.collections = collections
			ga.logger.Info("Collections loaded in state", "count", len(collections))
		})
	}()
}
