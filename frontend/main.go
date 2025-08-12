package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	"net/http"
	"time"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

// Config holds application configuration
type Config struct {
	BackendURL   string `mapstructure:"backend_url"`
	AuthURL      string `mapstructure:"auth_url"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
}

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
	oauth2Config   *oauth2.Config
	currentUser    *User
	token          *oauth2.Token
	groups         []Group
	collections    []Collection
	httpClient     *http.Client
	currentView    string
	isSignedIn     bool
	mainContainer  *core.Frame
	currentOverlay *core.Frame
	dialogState    *DialogState
	searchFilter   *SearchFilter
}

// loadConfig loads configuration from config file and environment variables
func loadConfig() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set defaults
	viper.SetDefault("backend_url", "http://localhost:3001")
	viper.SetDefault("auth_url", "https://authentik.local")
	viper.SetDefault("redirect_url", "http://localhost:8080/auth/callback")

	// Environment variable bindings
	viper.SetEnvPrefix("NISHIKI")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Config file not found, using defaults: %v\n", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("Error unmarshaling config: %v\n", err)
	}

	return &config
}

// NewApp creates a new application instance
func NewApp() *App {
	config := loadConfig()

	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.AuthURL + "/application/o/authorize/",
			TokenURL: config.AuthURL + "/application/o/token/",
		},
	}

	app := &App{
		config:       config,
		oauth2Config: oauth2Config,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		currentView:  "login",
		isSignedIn:   false,
	}

	// Initialize dialog state
	app.dialogState = &DialogState{}

	// Initialize search filter
	app.searchFilter = &SearchFilter{
		SortBy:        "name",
		SortDirection: "asc",
	}

	return app
}

// makeAuthenticatedRequest makes an HTTP request with authentication
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

	if app.token != nil {
		reqBody.Header.Set("Authorization", "Bearer "+app.token.AccessToken)
	}

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

// createMainUI creates the main application UI
func (app *App) createMainUI(b *core.Body) {
	b.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Background = colors.Uniform(ColorGrayLightest) // var(--color-gray-lightest)
	})

	// Create main container
	app.mainContainer = core.NewFrame(b)
	app.mainContainer.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Grow.Set(1, 1)
	})

	if !app.isSignedIn {
		app.showLoginView()
	} else {
		app.showDashboardView()
	}
}

// showLoginView displays the login screen
func (app *App) showLoginView() {
	app.mainContainer.DeleteChildren()
	app.currentView = "login"

	// Login container
	loginContainer := core.NewFrame(app.mainContainer)
	loginContainer.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Align.Items = styles.Center
		s.Justify.Content = styles.Center
		s.Grow.Set(1, 1)
		s.Gap.Set(units.Dp(32))
		s.Padding.Set(units.Dp(24))
	})

	// App title
	title := core.NewText(loginContainer).SetText("Nishiki Inventory")
	title.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(32)
		s.Font.Weight = styles.WeightBold
		s.Color = colors.Uniform(ColorPrimary) // var(--color-primary)
	})

	// Subtitle
	subtitle := core.NewText(loginContainer).SetText("Inventory Management System")
	subtitle.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(16)
		s.Color = colors.Uniform(ColorGrayDark) // var(--color-gray-dark)
	})

	// Login button
	loginBtn := core.NewButton(loginContainer).SetText("Sign In with Authentik")
	app.styleButtonPrimary(loginBtn)
	loginBtn.OnClick(func(e events.Event) {
		app.handleLogin()
	})

	app.mainContainer.Update()
}

// showDashboardView displays the main dashboard
func (app *App) showDashboardView() {
	app.mainContainer.DeleteChildren()
	app.currentView = "dashboard"

	// Header
	header := core.NewFrame(app.mainContainer)
	header.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Justify.Content = styles.SpaceBetween
		s.Background = colors.Uniform(ColorWhite)
		s.Padding.Set(units.Dp(16))
		s.Border.Style.Bottom = styles.BorderSolid
		s.Border.Width.Bottom = units.Dp(1)
		s.Border.Color.Bottom = colors.Uniform(ColorGrayLight) // var(--color-gray-light)
	})

	// Header title
	headerTitle := core.NewText(header).SetText("Dashboard")
	headerTitle.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(20)
		s.Font.Weight = styles.WeightSemiBold
		s.Color = colors.Uniform(ColorBlack) // var(--color-black)
	})

	// User menu button
	userBtn := core.NewButton(header)
	if app.currentUser != nil {
		userBtn.SetText(app.currentUser.Username)
	} else {
		userBtn.SetText("User")
	}
	userBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(color.RGBA{R: 240, G: 240, B: 240, A: 255})
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(8), units.Dp(16))
	})

	// Main content area
	content := core.NewFrame(app.mainContainer)
	content.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Grow.Set(1, 1)
		s.Padding.Set(units.Dp(16))
		s.Gap.Set(units.Dp(16))
	})

	// Navigation buttons
	navContainer := core.NewFrame(content)
	navContainer.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(units.Dp(12))
		s.Wrap = true
	})

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
	statsContainer.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(units.Dp(12))
		s.Background = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(16))
	})

	statsTitle := core.NewText(statsContainer).SetText("Quick Stats")
	statsTitle.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(18)
		s.Font.Weight = styles.WeightSemiBold
	})

	statsGrid := core.NewFrame(statsContainer)
	statsGrid.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(units.Dp(16))
		s.Wrap = true
	})

	// Groups count
	app.createStatCard(statsGrid, "Groups", fmt.Sprintf("%d", len(app.groups)), ColorPrimary)

	// Collections count
	app.createStatCard(statsGrid, "Collections", fmt.Sprintf("%d", len(app.collections)), ColorAccent)

	app.mainContainer.Update()
}

// createNavButton creates a navigation button
func (app *App) createNavButton(parent core.Widget, text string, icon icons.Icon, onClick func()) *core.Button {
	btn := core.NewButton(parent).SetText(text).SetIcon(icon)
	btn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(16))
		s.Gap.Set(units.Dp(8))
		s.Min.X.Set(120, units.UnitDp)
		s.Border.Style.Set(styles.BorderSolid)
		s.Border.Width.Set(units.Dp(1))
		s.Border.Color.Set(colors.Uniform(ColorGrayLight))
	})
	btn.OnClick(func(e events.Event) {
		onClick()
	})
	return btn
}

// createStatCard creates a statistics card
func (app *App) createStatCard(parent core.Widget, label, value string, cardColor color.RGBA) *core.Frame {
	card := core.NewFrame(parent)
	card.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Align.Items = styles.Center
		s.Background = colors.Uniform(cardColor)
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(units.Dp(16))
		s.Gap.Set(units.Dp(4))
		s.Min.X.Set(100, units.UnitDp)
	})

	valueText := core.NewText(card).SetText(value)
	valueText.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(24)
		s.Font.Weight = styles.WeightBold
		s.Color = colors.Uniform(ColorWhite)
	})

	labelText := core.NewText(card).SetText(label)
	labelText.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(14)
		s.Color = colors.Uniform(ColorWhite)
	})

	return card
}

// showGroupsView displays the groups management view
func (app *App) showGroupsView() {
	app.mainContainer.DeleteChildren()
	app.currentView = "groups"

	// Header with back button
	header := app.createHeader("Groups", true)

	// Refresh groups data
	if err := app.fetchGroups(); err != nil {
		fmt.Printf("Error fetching groups: %v\n", err)
	}

	// Main content
	content := core.NewFrame(app.mainContainer)
	content.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Grow.Set(1, 1)
		s.Padding.Set(units.Dp(16))
		s.Gap.Set(units.Dp(16))
	})

	// Create group button
	createBtn := core.NewButton(content).SetText("Create Group").SetIcon(icons.Add)
	app.styleButtonPrimary(createBtn)
	createBtn.Styler(func(s *styles.Style) {
		s.Align.Self = styles.End
	})

	// Groups list
	if len(app.groups) == 0 {
		emptyText := core.NewText(content).SetText("No groups found. Create your first group!")
		emptyText.Styler(func(s *styles.Style) {
			s.Color = colors.Uniform(ColorGrayDark) // var(--color-gray-dark)
			s.Align.Self = styles.Center
			s.Margin.Top = units.Dp(32)
		})
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
	app.currentView = "collections"

	// Header with back button
	header := app.createHeader("Collections", true)

	// Refresh collections data
	if err := app.fetchCollections(); err != nil {
		fmt.Printf("Error fetching collections: %v\n", err)
	}

	// Main content
	content := core.NewFrame(app.mainContainer)
	content.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Grow.Set(1, 1)
		s.Padding.Set(units.Dp(16))
		s.Gap.Set(units.Dp(16))
	})

	// Create collection button
	createBtn := core.NewButton(content).SetText("Create Collection").SetIcon(icons.Add)
	app.styleButtonPrimary(createBtn)
	createBtn.Styler(func(s *styles.Style) {
		s.Align.Self = styles.End
	})

	// Collections list
	if len(app.collections) == 0 {
		emptyText := core.NewText(content).SetText("No collections found. Create your first collection!")
		emptyText.Styler(func(s *styles.Style) {
			s.Color = colors.Uniform(ColorGrayDark) // var(--color-gray-dark)
			s.Align.Self = styles.Center
			s.Margin.Top = units.Dp(32)
		})
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
	app.currentView = "profile"

	// Header with back button
	header := app.createHeader("Profile", true)

	// Main content
	content := core.NewFrame(app.mainContainer)
	content.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Grow.Set(1, 1)
		s.Padding.Set(units.Dp(16))
		s.Gap.Set(units.Dp(16))
	})

	if app.currentUser != nil {
		// User info card
		userCard := core.NewFrame(content)
		userCard.Styler(func(s *styles.Style) {
			s.Direction = styles.Column
			s.Background = colors.Uniform(ColorWhite)
			s.Border.Radius = styles.BorderRadiusLarge
			s.Padding.Set(units.Dp(16))
			s.Gap.Set(units.Dp(12))
		})

		// Username
		usernameLabel := core.NewText(userCard).SetText("Username:")
		usernameLabel.Styler(func(s *styles.Style) {
			s.Font.Weight = styles.WeightSemiBold
			s.Color = colors.Uniform(ColorGrayDark)
		})
		username := core.NewText(userCard).SetText(app.currentUser.Username)

		// Email
		emailLabel := core.NewText(userCard).SetText("Email:")
		emailLabel.Styler(func(s *styles.Style) {
			s.Font.Weight = styles.WeightSemiBold
			s.Color = colors.Uniform(ColorGrayDark)
		})
		email := core.NewText(userCard).SetText(app.currentUser.Email)

		// Name (if available)
		if app.currentUser.Name != "" {
			nameLabel := core.NewText(userCard).SetText("Name:")
			nameLabel.Styler(func(s *styles.Style) {
				s.Font.Weight = styles.WeightSemiBold
				s.Color = colors.Uniform(ColorGrayDark)
			})
			name := core.NewText(userCard).SetText(app.currentUser.Name)
			_ = name
		}

		_ = username
		_ = email
	}

	// Logout button
	logoutBtn := core.NewButton(content).SetText("Sign Out").SetIcon(icons.Logout)
	app.styleButtonDanger(logoutBtn)
	logoutBtn.Styler(func(s *styles.Style) {
		s.Align.Self = styles.Start
		s.Margin.Top = units.Dp(16)
	})
	logoutBtn.OnClick(func(e events.Event) {
		app.handleLogout()
	})

	_ = header
	app.mainContainer.Update()
}

// createHeader creates a header with optional back button
func (app *App) createHeader(title string, showBack bool) *core.Frame {
	header := core.NewFrame(app.mainContainer)
	header.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Justify.Content = styles.SpaceBetween
		s.Background = colors.Uniform(ColorWhite)
		s.Padding.Set(units.Dp(16))
		s.Border.Style.Bottom = styles.BorderSolid
		s.Border.Width.Bottom = units.Dp(1)
		s.Border.Color.Bottom = colors.Uniform(ColorGrayLight) // var(--color-gray-light)
	})

	// Left side
	leftContainer := core.NewFrame(header)
	leftContainer.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Gap.Set(units.Dp(12))
	})

	if showBack {
		backBtn := core.NewButton(leftContainer).SetIcon(icons.ArrowBack)
		backBtn.Styler(func(s *styles.Style) {
			s.Background = colors.Uniform(color.RGBA{R: 240, G: 240, B: 240, A: 255})
			s.Border.Radius = styles.BorderRadiusFull
			s.Padding.Set(units.Dp(8))
		})
		backBtn.OnClick(func(e events.Event) {
			app.showDashboardView()
		})
	}

	// Header title
	headerTitle := core.NewText(leftContainer).SetText(title)
	headerTitle.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(20)
		s.Font.Weight = styles.WeightSemiBold
		s.Color = colors.Uniform(ColorBlack) // var(--color-black)
	})

	return header
}

// createGroupCard creates a card for displaying group information
func (app *App) createGroupCard(parent core.Widget, group Group) *core.Frame {
	card := core.NewFrame(parent)
	app.styleCard(card)
	card.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Justify.Content = styles.SpaceBetween
		s.Padding.Set(units.Dp(16))
		s.Border.Style.Set(styles.BorderSolid)
		s.Border.Width.Set(units.Dp(1))
		s.Border.Color.Set(colors.Uniform(ColorGrayLight))
	})

	// Group info
	infoContainer := core.NewFrame(card)
	infoContainer.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(units.Dp(4))
		s.Grow.Set(1, 0)
	})

	groupName := core.NewText(infoContainer).SetText(group.Name)
	groupName.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(16)
		s.Font.Weight = styles.WeightSemiBold
	})

	if group.Description != "" {
		groupDesc := core.NewText(infoContainer).SetText(group.Description)
		groupDesc.Styler(func(s *styles.Style) {
			s.Font.Size = units.Dp(14)
			s.Color = colors.Uniform(ColorGrayDark)
		})
	}

	membersText := core.NewText(infoContainer).SetText(fmt.Sprintf("%d members", len(group.Members)))
	membersText.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(12)
		s.Color = colors.Uniform(ColorGrayDark)
	})

	// View button
	viewBtn := core.NewButton(card).SetText("View").SetIcon(icons.ArrowForward)
	viewBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorPrimary)
		s.Color = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(units.Dp(8), units.Dp(12))
		s.Gap.Set(units.Dp(4))
	})

	return card
}

// createCollectionCard creates a card for displaying collection information
func (app *App) createCollectionCard(parent core.Widget, collection Collection) *core.Frame {
	card := core.NewFrame(parent)
	app.styleCard(card)
	card.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Justify.Content = styles.SpaceBetween
		s.Padding.Set(units.Dp(16))
		s.Border.Style.Set(styles.BorderSolid)
		s.Border.Width.Set(units.Dp(1))
		s.Border.Color.Set(colors.Uniform(ColorGrayLight))
	})

	// Collection info
	infoContainer := core.NewFrame(card)
	infoContainer.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(units.Dp(4))
		s.Grow.Set(1, 0)
	})

	collectionName := core.NewText(infoContainer).SetText(collection.Name)
	collectionName.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(16)
		s.Font.Weight = styles.WeightSemiBold
	})

	if collection.Description != "" {
		collectionDesc := core.NewText(infoContainer).SetText(collection.Description)
		collectionDesc.Styler(func(s *styles.Style) {
			s.Font.Size = units.Dp(14)
			s.Color = colors.Uniform(ColorGrayDark)
		})
	}

	objectTypeText := core.NewText(infoContainer).SetText("Type: " + collection.ObjectType)
	objectTypeText.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(12)
		s.Color = colors.Uniform(ColorGrayDark)
	})

	// View button
	viewBtn := core.NewButton(card).SetText("View").SetIcon(icons.ArrowForward)
	viewBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorAccent) // var(--color-accent)
		s.Color = colors.Uniform(ColorBlack)
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(units.Dp(8), units.Dp(12))
		s.Gap.Set(units.Dp(4))
	})

	return card
}

// handleLogin initiates the OAuth2 login flow
func (app *App) handleLogin() {
	// In a real implementation, this would open a browser window
	// For now, we'll simulate a successful login
	fmt.Println("Login initiated...")

	// Create a mock token (in real implementation, this would come from OAuth2 flow)
	app.token = &oauth2.Token{
		AccessToken: "mock_access_token",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(time.Hour),
	}

	// Fetch user data
	if err := app.fetchCurrentUser(); err != nil {
		fmt.Printf("Error fetching user: %v\n", err)
		return
	}

	// Fetch initial data
	app.fetchGroups()
	app.fetchCollections()

	app.isSignedIn = true
	app.showDashboardView()
}

// handleLogout signs the user out
func (app *App) handleLogout() {
	app.token = nil
	app.currentUser = nil
	app.groups = nil
	app.collections = nil
	app.isSignedIn = false
	app.showLoginView()
}

func main() {
	// Create the app
	app := NewApp()

	// Create and run the UI
	core.TheApp.SetName("Nishiki Inventory")
	core.AppAbout = "A cross-platform inventory management application built with Cogent Core"

	core.NewBody("Nishiki Inventory").AddAppBar(func(tb *core.Toolbar) {
		// App bar customization if needed
	}).SetFunc(func(b *core.Body) {
		app.createMainUI(b)
	}).NewWindow().Run()
}
