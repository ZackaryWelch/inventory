//go:build js && wasm

package app

import (
	"fmt"
	"syscall/js"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"

	"github.com/nishiki/frontend/ui/components"
	"github.com/nishiki/frontend/ui/layouts"
	appstyles "github.com/nishiki/frontend/ui/styles"
)

// fetchCurrentUser gets the current user from the backend
func (app *App) fetchCurrentUser() error {
	authInfo, err := app.authClient.GetCurrentUser()
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	app.currentUser = &authInfo.User
	app.logger.Debug("Fetched current user", "user_id", authInfo.User.ID, "name", authInfo.User.Name)
	return nil
}

// fetchGroups gets the user's groups from the backend
func (app *App) fetchGroups() error {
	groups, err := app.groupsClient.List()
	if err != nil {
		return fmt.Errorf("failed to get groups: %w", err)
	}

	app.groups = groups
	app.logger.Debug("Fetched groups", "count", len(groups))
	return nil
}

// fetchCollections gets the user's collections from the backend
func (app *App) fetchCollections() error {
	if app.currentUser == nil {
		return fmt.Errorf("no current user")
	}

	collections, err := app.collectionsClient.List(app.currentUser.ID)
	if err != nil {
		return fmt.Errorf("failed to get collections: %w", err)
	}

	app.collections = collections
	app.logger.Debug("Fetched collections", "count", len(collections))
	return nil
}

// CreateMainUI creates the main application UI (exported for web builds)
func (app *App) CreateMainUI(b *core.Body) {
	app.createMainUI(b)
}

// createMainUI creates the main application UI
func (app *App) createMainUI(b *core.Body) {
	b.Styler(func(s *styles.Style) {
		appstyles.StyleMainBackground(s)
		s.Direction = styles.Column // Ensure column layout
	})

	// Store body reference for creating overlays
	app.body = b

	// Create main container - this will grow to fill space
	app.mainContainer = core.NewFrame(b)
	app.mainContainer.Styler(func(s *styles.Style) {
		appstyles.StyleMainContainer(s)
		s.Grow.Set(1, 1)                   // Grow to fill available space
		s.Overflow.Y = styles.OverflowAuto // Allow scrolling if content is tall
	})

	// Mark UI as ready for snackbars and other UI operations
	app.uiReady = true
	app.logger.Debug("UI initialization complete, ready for user interactions")

	if app.currentView == ViewCallback {
		app.showCallbackView()
	} else if !app.isSignedIn {
		app.showLoginView()
	} else {
		app.showDashboardView()
	}
}

// updateBottomMenu updates or creates the bottom menu at body level
func (app *App) updateBottomMenu(activeView string) {
	// Safety check: ensure body is initialized
	if app.body == nil {
		app.logger.Error("Cannot create bottom menu: body is nil")
		return
	}

	// Remove existing bottom menu if it exists
	if app.bottomMenu != nil {
		app.bottomMenu.Delete()
	}

	// Create bottom menu at body level (after mainContainer)
	app.bottomMenu = layouts.CreateDefaultBottomMenu(app.body, activeView, app.handleNavigation)
}

// showLoginView displays the login screen matching React LoginPage.tsx
func (app *App) showLoginView() {
	app.currentView = ViewLogin

	// If UI not initialized yet, just set state and return
	if app.mainContainer == nil {
		app.logger.Debug("UI not initialized, deferring showLoginView")
		return
	}

	app.mainContainer.DeleteChildren()

	// Override mainContainer styling for login - className="flex items-center justify-center h-screen"
	// This replaces the default StyleMainContainer which has padding/column that breaks centering
	app.mainContainer.Styler(appstyles.StyleLoginContainer)

	// Logo placeholder - className="w-32 h-26 mb-20"
	// TODO: Add actual LogoVerticalPrimary SVG when available
	logo := core.NewFrame(app.mainContainer)
	logo.Styler(appstyles.StyleLoginLogo)

	// Logo text placeholder (remove when SVG logo is added)
	logoText := core.NewText(logo).SetText("NISHIKI")
	logoText.Styler(appstyles.StyleAppTitle) // Reuses existing app title style

	// Login button - matching React AuthentikLoginButton.tsx
	// className="w-full flex items-center justify-center px-4 py-3 border border-transparent rounded-md shadow-sm text-base font-medium text-white bg-blue-600"
	loginBtn := core.NewButton(app.mainContainer)
	loginBtn.SetText("Sign in with Authentik")
	loginBtn.SetIcon(icons.Login)
	loginBtn.Styler(appstyles.StyleButtonLogin)
	loginBtn.OnClick(func(e events.Event) {
		app.handleLogin()
	})

	// Subtitle text - className="mt-4 text-center" > "text-sm text-gray-600"
	subtitle := core.NewText(app.mainContainer).SetText("Secure authentication powered by Authentik")
	subtitle.Styler(appstyles.StyleLoginSubtitle)

	app.mainContainer.Update()
}

// showCallbackView displays the authentication callback loading screen
func (app *App) showCallbackView() {
	// If UI not initialized yet, just return (view already set by caller)
	if app.mainContainer == nil {
		app.logger.Debug("UI not initialized, deferring showCallbackView")
		return
	}

	app.mainContainer.DeleteChildren()

	// Centered layout for loading screen
	callbackContainer := layouts.CenteredLayout(app.mainContainer)

	// Loading content center
	loadingContent := core.NewFrame(callbackContainer)
	loadingContent.Styler(appstyles.StyleTextCenter) // text-center

	// Loading spinner using component library
	components.LoadingSpinner(loadingContent)

	// Loading title
	title := core.NewText(loadingContent).SetText("Completing Sign In...")
	title.Styler(appstyles.StyleAppTitle)

	// Loading message
	message := core.NewText(loadingContent).SetText("Please wait while we authenticate you with Authentik.")
	message.Styler(appstyles.StyleSubtitle)

	app.mainContainer.Update()
}

// showDashboardView displays the main dashboard
func (app *App) showDashboardView() {
	app.currentView = ViewDashboard

	// If UI not initialized yet, just set state and return
	if app.mainContainer == nil {
		app.logger.Debug("UI not initialized, deferring showDashboardView")
		return
	}

	app.mainContainer.DeleteChildren()

	// Restore mainContainer styling for mobile layout (pt-12 pb-16 min-h-screen)
	app.mainContainer.Styler(appstyles.StyleMainContainer)

	// Header with user menu button
	username := "User"
	if app.currentUser != nil {
		username = app.currentUser.Name
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
	navContainer.Styler(appstyles.StyleNavContainer)

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
	statsContainer.Styler(appstyles.StyleStatsContainer)

	statsTitle := core.NewText(statsContainer).SetText("Quick Stats")
	statsTitle.Styler(appstyles.StyleStatsTitle)

	statsGrid := core.NewFrame(statsContainer)
	statsGrid.Styler(appstyles.StyleStatsGrid)

	// Groups count card
	groupsCard := components.Card(statsGrid, components.CardProps{})
	groupsCard.Styler(appstyles.StyleStatCard(appstyles.ColorPrimary))
	groupsValue := core.NewText(groupsCard).SetText(fmt.Sprintf("%d", len(app.groups)))
	groupsValue.Styler(appstyles.StyleStatValue)
	groupsLabel := core.NewText(groupsCard).SetText("Groups")
	groupsLabel.Styler(appstyles.StyleStatLabel)

	// Collections count card
	collectionsCard := components.Card(statsGrid, components.CardProps{})
	collectionsCard.Styler(appstyles.StyleStatCard(appstyles.ColorAccent))
	collectionsValue := core.NewText(collectionsCard).SetText(fmt.Sprintf("%d", len(app.collections)))
	collectionsValue.Styler(appstyles.StyleStatValue)
	collectionsLabel := core.NewText(collectionsCard).SetText("Collections")
	collectionsLabel.Styler(appstyles.StyleStatLabel)

	// Bottom navigation bar
	app.updateBottomMenu("dashboard")

	app.body.Update()
}

// showProfileView displays the user profile view
func (app *App) showProfileView() {
	app.currentView = ViewProfile

	// If UI not initialized yet, just set state and return
	if app.mainContainer == nil {
		app.logger.Debug("UI not initialized, deferring showProfileView")
		return
	}

	app.mainContainer.DeleteChildren()

	// Page title - using helper function
	layouts.PageTitle(app.mainContainer, "Profile")

	// Main content - using existing layout function
	content := layouts.ContentColumn(app.mainContainer)

	if app.currentUser != nil {
		// User info card using component library
		userCard := components.Card(content, components.CardProps{})
		userCard.Styler(appstyles.StyleProfileCard) // Add proper card layout

		// Name
		nameLabel := core.NewText(userCard).SetText("Name:")
		nameLabel.Styler(appstyles.StyleUserFieldLabel)
		nameValue := core.NewText(userCard).SetText(app.currentUser.Name)
		nameValue.Styler(func(s *styles.Style) {
			s.Color = colors.Uniform(appstyles.ColorBlack) // Ensure text is visible
			s.Font.Size = units.Dp(appstyles.FontSizeBase)
		})

		// Email
		emailLabel := core.NewText(userCard).SetText("Email:")
		emailLabel.Styler(appstyles.StyleUserFieldLabel)
		emailValue := core.NewText(userCard).SetText(app.currentUser.Email)
		emailValue.Styler(func(s *styles.Style) {
			s.Color = colors.Uniform(appstyles.ColorBlack) // Ensure text is visible
			s.Font.Size = units.Dp(appstyles.FontSizeBase)
		})
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
	devSection.Styler(appstyles.StyleDevSection)

	devTitle := core.NewText(devSection).SetText("Developer Tools")
	devTitle.Styler(appstyles.StyleDevTitle)

	// Clear cache button using component library
	components.Button(devSection, components.ButtonProps{
		Text:    "Clear Cache & Reload",
		Icon:    icons.Refresh,
		Variant: components.ButtonPrimary,
		Size:    components.ButtonSizeMedium,
		OnClick: func(e events.Event) {
			app.handleClearCacheAndReload()
		},
	})

	// Bottom navigation bar
	app.updateBottomMenu("profile")

	app.body.Update()
}

// createCollectionCard creates a card for displaying collection information
// Matches nishiki-frontend ContainerCard pattern: Card className="flex justify-between gap-2"
func (app *App) createCollectionCard(parent core.Widget, collection Collection) *core.Frame {
	card := core.NewFrame(parent)
	card.Styler(appstyles.StyleCardFlexBetween) // Card + flex justify-between gap-2

	// Link content area (flex grow gap-4 items-center pl-4 py-2)
	contentArea := core.NewFrame(card)
	contentArea.Styler(appstyles.StyleCardContentGrow) // flex grow gap-4 items-center pl-4 py-2
	contentArea.OnClick(func(e events.Event) {
		app.showCollectionDetailView(collection)
	})

	// Icon circle (flex items-center justify-center bg-accent rounded-full w-11 h-11)
	iconCircle := core.NewFrame(contentArea)
	iconCircle.Styler(appstyles.StyleIconCircleAccent) // bg-accent rounded-full w-11 h-11

	icon := core.NewIcon(iconCircle).SetIcon(app.getIcon(collection.ObjectType))
	icon.Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(appstyles.ColorBlack) // color="black" for accent background
		s.Font.Size = units.Dp(24)                     // size={6} in frontend (24px)
	})

	// Collection name (leading-5)
	collectionName := core.NewText(contentArea).SetText(collection.Name)
	collectionName.Styler(func(s *styles.Style) {
		s.Text.LineHeight = 20 // leading-5 (20px)
	})

	// Dropdown menu button (w-12)
	menuBtn := core.NewButton(card).SetIcon(icons.MoreVert)
	menuBtn.Styler(func(s *styles.Style) {
		s.Min.X.Set(48, units.UnitDp) // w-12
		// variant="ghost" - transparent background
		s.Background = nil
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

// handleClearCacheAndReload clears browser cache and performs a hard reload
// This forces the browser to fetch the latest app.wasm instead of using cached version
func (app *App) handleClearCacheAndReload() {
	app.logger.Info("Clearing cache and reloading...")

	// Clear localStorage to remove cached tokens
	localStorage := js.Global().Get("localStorage")
	if !localStorage.IsUndefined() {
		localStorage.Call("clear")
		app.logger.Info("Cleared localStorage")
	}

	// Clear sessionStorage
	sessionStorage := js.Global().Get("sessionStorage")
	if !sessionStorage.IsUndefined() {
		sessionStorage.Call("clear")
		app.logger.Info("Cleared sessionStorage")
	}

	// Perform hard reload to bypass HTTP cache and fetch latest app.wasm
	// location.reload() in modern browsers already bypasses cache
	app.logger.Info("Performing page reload to fetch latest app.wasm")
	js.Global().Get("location").Call("reload")
}

// handleNavigation handles bottom menu navigation
func (app *App) handleNavigation(view string) {
	switch view {
	case "dashboard":
		app.showDashboardView()
	case "groups":
		app.showEnhancedGroupsView()
	case "collections":
		app.showEnhancedCollectionsView()
	case "search":
		app.showGlobalSearchView()
	case "profile":
		app.showProfileView()
	}
}
