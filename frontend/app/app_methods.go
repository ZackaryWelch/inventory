package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"

	"github.com/nishiki/frontend/ui/components"
	"github.com/nishiki/frontend/ui/layouts"
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

	// Centered layout for login screen
	loginContainer := layouts.CenteredLayout(app.mainContainer)

	// Login content column
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

	// Login button using component library
	components.Button(loginContent, components.ButtonProps{
		Text:    "Sign In with Authentik",
		Variant: components.ButtonPrimary,
		Size:    components.ButtonSizeLarge,
		OnClick: func(e events.Event) {
			app.handleLogin()
		},
	})

	app.mainContainer.Update()
}

// showCallbackView displays the authentication callback loading screen
func (app *App) showCallbackView() {
	app.mainContainer.DeleteChildren()

	// Centered layout for loading screen
	callbackContainer := layouts.CenteredLayout(app.mainContainer)

	// Loading content center
	loadingContent := core.NewFrame(callbackContainer)
	loadingContent.Styler(StyleTextCenter) // text-center

	// Loading spinner using component library
	components.LoadingSpinner(loadingContent)

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

	// Header with user menu button
	username := "User"
	if app.currentUser != nil {
		username = app.currentUser.Username
	}
	layouts.Header(app.mainContainer, layouts.HeaderProps{
		Title: "Dashboard",
		RightItems: []layouts.HeaderItem{
			{
				Text: username,
				OnClick: func() {
					app.showProfileView()
				},
			},
		},
	})

	// Main content area
	content := layouts.ContentColumn(app.mainContainer)

	// Navigation buttons
	navContainer := core.NewFrame(content)
	navContainer.Styler(StyleNavContainer)

	// Groups button
	components.Button(navContainer, components.ButtonProps{
		Text:    "Groups",
		Icon:    icons.Group,
		Variant: components.ButtonPrimary,
		Size:    components.ButtonSizeMedium,
		OnClick: func(e events.Event) {
			app.showEnhancedGroupsView()
		},
	})

	// Collections button
	components.Button(navContainer, components.ButtonProps{
		Text:    "Collections",
		Icon:    icons.FolderOpen,
		Variant: components.ButtonPrimary,
		Size:    components.ButtonSizeMedium,
		OnClick: func(e events.Event) {
			app.showEnhancedCollectionsView()
		},
	})

	// Profile button
	components.Button(navContainer, components.ButtonProps{
		Text:    "Profile",
		Icon:    icons.Person,
		Variant: components.ButtonPrimary,
		Size:    components.ButtonSizeMedium,
		OnClick: func(e events.Event) {
			app.showProfileView()
		},
	})

	// Search button
	components.Button(navContainer, components.ButtonProps{
		Text:    "Search",
		Icon:    icons.Search,
		Variant: components.ButtonPrimary,
		Size:    components.ButtonSizeMedium,
		OnClick: func(e events.Event) {
			app.showGlobalSearchView()
		},
	})

	// Stats section
	statsContainer := core.NewFrame(content)
	statsContainer.Styler(StyleStatsContainer)

	statsTitle := core.NewText(statsContainer).SetText("Quick Stats")
	statsTitle.Styler(StyleStatsTitle)

	statsGrid := core.NewFrame(statsContainer)
	statsGrid.Styler(StyleStatsGrid)

	// Groups count card
	groupsCard := components.Card(statsGrid, components.CardProps{})
	groupsCard.Styler(StyleStatCard(ColorPrimary))
	groupsValue := core.NewText(groupsCard).SetText(fmt.Sprintf("%d", len(app.groups)))
	groupsValue.Styler(StyleStatValue)
	groupsLabel := core.NewText(groupsCard).SetText("Groups")
	groupsLabel.Styler(StyleStatLabel)

	// Collections count card
	collectionsCard := components.Card(statsGrid, components.CardProps{})
	collectionsCard.Styler(StyleStatCard(ColorAccent))
	collectionsValue := core.NewText(collectionsCard).SetText(fmt.Sprintf("%d", len(app.collections)))
	collectionsValue.Styler(StyleStatValue)
	collectionsLabel := core.NewText(collectionsCard).SetText("Collections")
	collectionsLabel.Styler(StyleStatLabel)

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

	// Header with back button using layout component
	layouts.SimpleHeader(app.mainContainer, "Groups", true, func() {
		app.showDashboardView()
	})

	// Refresh groups data
	if err := app.fetchGroups(); err != nil {
		app.logger.Error("Error fetching groups", "error", err)
	}

	// Main content
	content := layouts.ContentColumn(app.mainContainer)

	// Create group button using component library
	components.Button(content, components.ButtonProps{
		Text:    "Create Group",
		Icon:    icons.Add,
		Variant: components.ButtonPrimary,
		Size:    components.ButtonSizeMedium,
		OnClick: func(e events.Event) {
			// TODO: Show create group dialog
		},
	})

	// Groups list
	if len(app.groups) == 0 {
		components.EmptyState(content, "No groups found. Create your first group!")
	} else {
		for _, group := range app.groups {
			app.createGroupCard(content, group)
		}
	}

	app.mainContainer.Update()
}

// showCollectionsView displays the collections management view
func (app *App) showCollectionsView() {
	app.mainContainer.DeleteChildren()
	app.currentView = ViewCollections

	// Header with back button using layout component
	layouts.SimpleHeader(app.mainContainer, "Collections", true, func() {
		app.showDashboardView()
	})

	// Refresh collections data
	if err := app.fetchCollections(); err != nil {
		app.logger.Error("Error fetching collections", "error", err)
	}

	// Main content
	content := layouts.ContentColumn(app.mainContainer)

	// Create collection button using component library
	components.Button(content, components.ButtonProps{
		Text:    "Create Collection",
		Icon:    icons.Add,
		Variant: components.ButtonPrimary,
		Size:    components.ButtonSizeMedium,
		OnClick: func(e events.Event) {
			// TODO: Show create collection dialog
		},
	})

	// Collections list
	if len(app.collections) == 0 {
		components.EmptyState(content, "No collections found. Create your first collection!")
	} else {
		for _, collection := range app.collections {
			app.createCollectionCard(content, collection)
		}
	}

	app.mainContainer.Update()
}

// showProfileView displays the user profile view
func (app *App) showProfileView() {
	app.mainContainer.DeleteChildren()
	app.currentView = ViewProfile

	// Header with back button using layout component
	layouts.SimpleHeader(app.mainContainer, "Profile", true, func() {
		app.showDashboardView()
	})

	// Main content
	content := layouts.ContentColumn(app.mainContainer)

	if app.currentUser != nil {
		// User info card using component library
		userCard := components.Card(content, components.CardProps{})

		// Username
		usernameLabel := core.NewText(userCard).SetText("Username:")
		usernameLabel.Styler(StyleUserFieldLabel)
		core.NewText(userCard).SetText(app.currentUser.Username)

		// Email
		emailLabel := core.NewText(userCard).SetText("Email:")
		emailLabel.Styler(StyleUserFieldLabel)
		core.NewText(userCard).SetText(app.currentUser.Email)

		// Name (if available)
		if app.currentUser.Name != "" {
			nameLabel := core.NewText(userCard).SetText("Name:")
			nameLabel.Styler(StyleUserFieldLabel)
			core.NewText(userCard).SetText(app.currentUser.Name)
		}
	}

	// Logout button using component library
	components.Button(content, components.ButtonProps{
		Text:    "Sign Out",
		Icon:    icons.Logout,
		Variant: components.ButtonDanger,
		Size:    components.ButtonSizeMedium,
		OnClick: func(e events.Event) {
			app.handleLogout()
		},
	})

	// Developer tools section
	devSection := core.NewFrame(content)
	devSection.Styler(StyleDevSection)

	devTitle := core.NewText(devSection).SetText("Developer Tools")
	devTitle.Styler(StyleDevTitle)

	// Clear cache button using component library
	components.Button(devSection, components.ButtonProps{
		Text:    "Clear Cache & Reload",
		Icon:    icons.Refresh,
		Variant: components.ButtonPrimary,
		Size:    components.ButtonSizeMedium,
		OnClick: func(e events.Event) {
			// TODO: Implement cache clearing
		},
	})

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
