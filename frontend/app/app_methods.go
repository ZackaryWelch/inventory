package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	"net/http"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
)

func (app *App) makeAuthenticatedRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody *http.Request
	var err error

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody, err = http.NewRequest(method, app.config.BackendURL+endpoint, bytes.NewBuffer(jsonBody))
		if err != nil {
			return nil, err
		}
		reqBody.Header.Set("Content-Type", "application/json")
	} else {
		reqBody, err = http.NewRequest(method, app.config.BackendURL+endpoint, nil)
		if err != nil {
			return nil, err
		}
	}

	// Get access token from auth service
	accessToken, err := app.authService.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	reqBody.Header.Set("Authorization", "Bearer "+accessToken)

	return app.httpClient.Do(reqBody)
}

// fetchCurrentUser gets the current user from the backend
func (app *App) fetchCurrentUser() error {
	resp, err := app.makeAuthenticatedRequest("GET", "/auth/me", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get current user: %d", resp.StatusCode)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return err
	}

	app.currentUser = &user
	return nil
}

// fetchGroups gets the user's groups from the backend
func (app *App) fetchGroups() error {
	resp, err := app.makeAuthenticatedRequest("GET", "/groups", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get groups: %d", resp.StatusCode)
	}

	var groups []Group
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return err
	}

	app.groups = groups
	return nil
}

// fetchCollections gets the user's collections from the backend
func (app *App) fetchCollections() error {
	if app.currentUser == nil {
		return fmt.Errorf("no current user")
	}

	resp, err := app.makeAuthenticatedRequest("GET", "/accounts/"+app.currentUser.ID+"/collections", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get collections: %d", resp.StatusCode)
	}

	var collections []Collection
	if err := json.NewDecoder(resp.Body).Decode(&collections); err != nil {
		return err
	}

	app.collections = collections
	return nil
}

// CreateMainUI creates the main application UI (exported for web builds)
func (app *App) CreateMainUI(b *core.Body) {
	app.createMainUI(b)
}

// createMainUI creates the main application UI
func (app *App) createMainUI(b *core.Body) {
	b.Styler(StyleMainBackground)

	// Create main container
	app.mainContainer = core.NewFrame(b)
	app.mainContainer.Styler(StyleMainContainer)

	if app.currentView == ViewCallback {
		app.showCallbackView()
	} else if !app.isSignedIn {
		app.showLoginView()
	} else {
		app.showDashboardView()
	}
}

// showLoginView displays the login screen
func (app *App) showLoginView() {
	app.mainContainer.DeleteChildren()
	app.currentView = ViewLogin

	// Login container with full screen layout
	loginContainer := core.NewFrame(app.mainContainer)
	loginContainer.Styler(StyleLoginContainer) // flex items-center justify-center h-screen

	// Login content center
	loginContent := core.NewFrame(loginContainer)
	loginContent.Styler(StyleLoginContent) // flex flex-col items-center justify-center

	// Logo placeholder (would be LogoVerticalPrimary in frontend)
	logo := core.NewFrame(loginContent)
	logo.Styler(StyleLoginLogo) // w-32 h-26 mb-20

	// App title
	title := core.NewText(loginContent).SetText("Nishiki Inventory")
	title.Styler(StyleAppTitle)

	// Subtitle
	subtitle := core.NewText(loginContent).SetText("Inventory Management System")
	subtitle.Styler(StyleSubtitle)

	// Login button
	loginBtn := core.NewButton(loginContent).SetText("Sign In with Authentik")
	loginBtn.Styler(StyleButtonPrimary)
	loginBtn.Styler(StyleButtonLg) // Use large size for login
	loginBtn.OnClick(func(e events.Event) {
		app.handleLogin()
	})

	app.mainContainer.Update()
}

// showCallbackView displays the authentication callback loading screen
func (app *App) showCallbackView() {
	app.mainContainer.DeleteChildren()

	// Callback container with loading screen pattern
	callbackContainer := core.NewFrame(app.mainContainer)
	callbackContainer.Styler(StyleLoadingScreen) // min-h-screen flex items-center justify-center

	// Loading content center
	loadingContent := core.NewFrame(callbackContainer)
	loadingContent.Styler(StyleTextCenter) // text-center

	// Loading spinner
	spinner := core.NewFrame(loadingContent)
	spinner.Styler(StyleLoadingSpinner) // animate-spin rounded-full h-12 w-12 border-b-2 mx-auto mb-4

	// Loading title
	title := core.NewText(loadingContent).SetText("Completing Sign In...")
	title.Styler(StyleAppTitle)

	// Loading message
	message := core.NewText(loadingContent).SetText("Please wait while we authenticate you with Authentik.")
	message.Styler(StyleSubtitle)

	app.mainContainer.Update()
}

// showDashboardView displays the main dashboard
func (app *App) showDashboardView() {
	app.mainContainer.DeleteChildren()
	app.currentView = ViewDashboard

	// Header
	header := core.NewFrame(app.mainContainer)
	header.Styler(StyleHeaderRow)

	// Header title
	headerTitle := core.NewText(header).SetText("Dashboard")
	headerTitle.Styler(StyleSectionTitle)

	// User menu button
	userBtn := core.NewButton(header)
	if app.currentUser != nil {
		userBtn.SetText(app.currentUser.Username)
	} else {
		userBtn.SetText("User")
	}
	userBtn.Styler(StyleUserButton)

	// Main content area
	content := core.NewFrame(app.mainContainer)
	content.Styler(StyleContentColumn)

	// Navigation buttons
	navContainer := core.NewFrame(content)
	navContainer.Styler(StyleNavContainer)

	// Groups button
	groupsBtn := app.createNavButton(navContainer, "Groups", icons.Group, func() {
		app.showEnhancedGroupsView()
	})

	// Collections button
	collectionsBtn := app.createNavButton(navContainer, "Collections", icons.FolderOpen, func() {
		app.showEnhancedCollectionsView()
	})

	// Profile button
	profileBtn := app.createNavButton(navContainer, "Profile", icons.Person, func() {
		app.showProfileView()
	})

	// Search button
	searchBtn := app.createNavButton(navContainer, "Search", icons.Search, func() {
		app.showGlobalSearchView()
	})

	_ = groupsBtn
	_ = collectionsBtn
	_ = profileBtn
	_ = searchBtn

	// Stats section
	statsContainer := core.NewFrame(content)
	statsContainer.Styler(StyleStatsContainer)

	statsTitle := core.NewText(statsContainer).SetText("Quick Stats")
	statsTitle.Styler(StyleStatsTitle)

	statsGrid := core.NewFrame(statsContainer)
	statsGrid.Styler(StyleStatsGrid)

	// Groups count
	app.createStatCard(statsGrid, "Groups", fmt.Sprintf("%d", len(app.groups)), ColorPrimary)

	// Collections count
	app.createStatCard(statsGrid, "Collections", fmt.Sprintf("%d", len(app.collections)), ColorAccent)

	app.mainContainer.Update()
}

// createNavButton creates a navigation button
func (app *App) createNavButton(parent core.Widget, text string, icon icons.Icon, onClick func()) *core.Button {
	btn := core.NewButton(parent).SetText(text).SetIcon(icon)
	btn.Styler(StyleNavButton)
	btn.OnClick(func(e events.Event) {
		onClick()
	})
	return btn
}

// createStatCard creates a statistics card
func (app *App) createStatCard(parent core.Widget, label, value string, cardColor color.RGBA) *core.Frame {
	card := core.NewFrame(parent)
	card.Styler(StyleStatCard(cardColor))

	valueText := core.NewText(card).SetText(value)
	valueText.Styler(StyleStatValue)

	labelText := core.NewText(card).SetText(label)
	labelText.Styler(StyleStatLabel)

	return card
}

// showGroupsView displays the groups management view
func (app *App) showGroupsView() {
	app.mainContainer.DeleteChildren()
	app.currentView = ViewGroups

	// Header with back button
	header := app.createHeader("Groups", true)

	// Refresh groups data
	if err := app.fetchGroups(); err != nil {
		app.logger.Error("Error fetching groups", "error", err)
	}

	// Main content
	content := core.NewFrame(app.mainContainer)
	content.Styler(StyleContentColumn)

	// Create group button
	createBtn := core.NewButton(content).SetText("Create Group").SetIcon(icons.Add)
	createBtn.Styler(StyleButtonPrimary)
	createBtn.Styler(StyleButtonMd)
	createBtn.Styler(StyleCreateButton)

	// Groups list
	if len(app.groups) == 0 {
		emptyText := core.NewText(content).SetText("No groups found. Create your first group!")
		emptyText.Styler(StyleEmptyText)
	} else {
		for _, group := range app.groups {
			app.createGroupCard(content, group)
		}
	}

	_ = header
	app.mainContainer.Update()
}

// showCollectionsView displays the collections management view
func (app *App) showCollectionsView() {
	app.mainContainer.DeleteChildren()
	app.currentView = ViewCollections

	// Header with back button
	header := app.createHeader("Collections", true)

	// Refresh collections data
	if err := app.fetchCollections(); err != nil {
		app.logger.Error("Error fetching collections", "error", err)
	}

	// Main content
	content := core.NewFrame(app.mainContainer)
	content.Styler(StyleContentColumn)

	// Create collection button
	createBtn := core.NewButton(content).SetText("Create Collection").SetIcon(icons.Add)
	createBtn.Styler(StyleButtonPrimary)
	createBtn.Styler(StyleButtonMd)
	createBtn.Styler(StyleCreateButton)

	// Collections list
	if len(app.collections) == 0 {
		emptyText := core.NewText(content).SetText("No collections found. Create your first collection!")
		emptyText.Styler(StyleEmptyText)
	} else {
		for _, collection := range app.collections {
			app.createCollectionCard(content, collection)
		}
	}

	_ = header
	app.mainContainer.Update()
}

// showProfileView displays the user profile view
func (app *App) showProfileView() {
	app.mainContainer.DeleteChildren()
	app.currentView = ViewProfile

	// Header with back button
	header := app.createHeader("Profile", true)

	// Main content
	content := core.NewFrame(app.mainContainer)
	content.Styler(StyleContentColumn)

	if app.currentUser != nil {
		// User info card
		userCard := core.NewFrame(content)
		userCard.Styler(StyleCard)

		// Username
		usernameLabel := core.NewText(userCard).SetText("Username:")
		usernameLabel.Styler(StyleUserFieldLabel)
		username := core.NewText(userCard).SetText(app.currentUser.Username)

		// Email
		emailLabel := core.NewText(userCard).SetText("Email:")
		emailLabel.Styler(StyleUserFieldLabel)
		email := core.NewText(userCard).SetText(app.currentUser.Email)

		// Name (if available)
		if app.currentUser.Name != "" {
			nameLabel := core.NewText(userCard).SetText("Name:")
			nameLabel.Styler(StyleUserFieldLabel)
			name := core.NewText(userCard).SetText(app.currentUser.Name)
			_ = name
		}

		_ = username
		_ = email
	}

	// Logout button
	logoutBtn := core.NewButton(content).SetText("Sign Out").SetIcon(icons.Logout)
	logoutBtn.Styler(StyleButtonDanger)
	logoutBtn.Styler(StyleButtonMd)
	logoutBtn.Styler(StyleLogoutButton)
	logoutBtn.OnClick(func(e events.Event) {
		app.handleLogout()
	})

	// Developer tools section
	devSection := core.NewFrame(content)
	devSection.Styler(StyleDevSection)

	devTitle := core.NewText(devSection).SetText("Developer Tools")
	devTitle.Styler(StyleDevTitle)

	// Clear cache button
	clearCacheBtn := core.NewButton(devSection).SetText("Clear Cache & Reload").SetIcon(icons.Refresh)
	clearCacheBtn.Styler(StyleClearCacheButton)
	/*clearCacheBtn.OnClick(func(e events.Event) {
		app.clearCacheAndReload()
	})*/

	_ = header
	app.mainContainer.Update()
}

// createHeader creates a header with optional back button
func (app *App) createHeader(title string, showBack bool) *core.Frame {
	header := core.NewFrame(app.mainContainer)
	header.Styler(StyleHeaderRow)

	// Left side
	leftContainer := core.NewFrame(header)
	leftContainer.Styler(StyleHeaderLeftContainer)

	if showBack {
		backBtn := core.NewButton(leftContainer).SetIcon(icons.ArrowBack)
		backBtn.Styler(StyleBackButton)
		backBtn.OnClick(func(e events.Event) {
			app.showDashboardView()
		})
	}

	// Header title
	headerTitle := core.NewText(leftContainer).SetText(title)
	headerTitle.Styler(StyleSectionTitle)

	return header
}

// createGroupCard creates a card for displaying group information
// Matches nishiki-frontend pattern: Card className="flex justify-between gap-2"
func (app *App) createGroupCard(parent core.Widget, group Group) *core.Frame {
	card := core.NewFrame(parent)
	card.Styler(StyleCardFlexBetween) // Card + flex justify-between gap-2

	// Link content area (flex grow flex-col gap-3 pl-4 py-2)
	contentArea := core.NewFrame(card)
	contentArea.Styler(StyleCardContentColumn) // flex grow flex-col gap-3 pl-4 py-2
	contentArea.OnClick(func(e events.Event) {
		app.showGroupDetailView(group)
	})

	// Group name (text-lg leading-6)
	groupName := core.NewText(contentArea).SetText(group.Name)
	groupName.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(18) // text-lg
		s.Text.LineHeight = 24     // leading-6
	})

	// Stats area (w-full flex justify-between items-center)
	statsArea := core.NewFrame(contentArea)
	statsArea.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Justify.Content = styles.SpaceBetween
		s.Align.Items = styles.Center
		s.Min.X.Set(100, units.UnitEw) // w-full
	})

	// Member count (matching frontend pattern)
	membersText := core.NewText(statsArea).SetText(fmt.Sprintf("%d members", len(group.Members)))
	membersText.Styler(StyleSmallText)

	// Dropdown menu button (w-12)
	menuBtn := core.NewButton(card).SetIcon(icons.MoreVert)
	menuBtn.Styler(func(s *styles.Style) {
		s.Min.X.Set(48, units.UnitDp)                                     // w-12
		s.Background = colors.Uniform(color.RGBA{R: 0, G: 0, B: 0, A: 0}) // variant="ghost"
	})

	return card
}

// createCollectionCard creates a card for displaying collection information
// Matches nishiki-frontend ContainerCard pattern: Card className="flex justify-between gap-2"
func (app *App) createCollectionCard(parent core.Widget, collection Collection) *core.Frame {
	card := core.NewFrame(parent)
	card.Styler(StyleCardFlexBetween) // Card + flex justify-between gap-2

	// Link content area (flex grow gap-4 items-center pl-4 py-2)
	contentArea := core.NewFrame(card)
	contentArea.Styler(StyleCardContentGrow) // flex grow gap-4 items-center pl-4 py-2
	contentArea.OnClick(func(e events.Event) {
		app.showCollectionDetailView(collection)
	})

	// Icon circle (flex items-center justify-center bg-accent rounded-full w-11 h-11)
	iconCircle := core.NewFrame(contentArea)
	iconCircle.Styler(StyleIconCircleAccent) // bg-accent rounded-full w-11 h-11

	icon := core.NewIcon(iconCircle).SetIcon(app.getIcon(collection.ObjectType))
	icon.Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(ColorBlack) // color="black" for accent background
		s.Font.Size = units.Dp(24)           // size={6} in frontend (24px)
	})

	// Collection name (leading-5)
	collectionName := core.NewText(contentArea).SetText(collection.Name)
	collectionName.Styler(func(s *styles.Style) {
		s.Text.LineHeight = 20 // leading-5 (20px)
	})

	// Dropdown menu button (w-12)
	menuBtn := core.NewButton(card).SetIcon(icons.MoreVert)
	menuBtn.Styler(func(s *styles.Style) {
		s.Min.X.Set(48, units.UnitDp)                                     // w-12
		s.Background = colors.Uniform(color.RGBA{R: 0, G: 0, B: 0, A: 0}) // variant="ghost"
	})

	return card
}

// handleLogin initiates the OAuth2 login flow
func (app *App) handleLogin() {
	app.logger.Info("Login initiated")

	// Initiate OAuth2 login flow through Authentik
	if err := app.authService.InitiateLogin(); err != nil {
		app.logger.Error("Error initiating login", "error", err)
		return
	}

	// The user will be redirected to Authentik, and we'll handle the callback
	// when they return to /auth/callback
}

// handleLogout signs the user out
func (app *App) handleLogout() {
	// Clear application state
	app.currentUser = nil
	app.groups = nil
	app.collections = nil
	app.isSignedIn = false

	// Logout through auth service (clears tokens and redirects to Authentik logout)
	if err := app.authService.Logout(); err != nil {
		app.logger.Error("Error during logout", "error", err)
	}

	app.showLoginView()
}
