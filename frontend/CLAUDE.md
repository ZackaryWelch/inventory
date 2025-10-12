# CLAUDE.md - Nishiki Frontend

This file provides guidance to Claude Code when working with the Nishiki Frontend codebase.

## Project Overview

The Nishiki Frontend is a cross-platform inventory management application built with Go and Cogent Core, compiling to both native desktop applications and WebAssembly for web deployment. It implements proper OAuth2/OIDC authentication with Authentik and provides a modern UI for managing inventory collections.

## Architecture

### Build Targets

The frontend supports multiple build targets using Go build constraints:

1. **Desktop Application** (`//go:build !js || !wasm`)
   - Native GUI using Cogent Core
   - Direct filesystem access for configuration
   - Local storage for user data

2. **WebAssembly Application** (`//go:build js && wasm`)
   - Runs in browser environment
   - Embedded configuration using `//go:embed`
   - Browser localStorage for token storage
   - JavaScript interop via `syscall/js`

### Configuration System

#### Configuration Structure
```go
type Config struct {
    BackendURL  string `mapstructure:"backend_url"`
    AuthURL     string `mapstructure:"auth_url"`
    ClientID    string `mapstructure:"client_id"`
    RedirectURL string `mapstructure:"redirect_url"`
    Port        string `mapstructure:"port"`
}
```

#### Build-Specific Configuration Loading

**Desktop Builds** (`app/config_desktop.go`):
- Loads `config.toml` from filesystem using Viper
- Search paths: `"."` and `"./config"`
- Supports environment variable overrides with `NISHIKI_` prefix

**WebAssembly Builds** (`app/config_wasm.go`):
- Uses embedded `config.toml` via `//go:embed config/config.toml`
- Configuration is baked into the WASM binary at build time
- Still supports environment variable overrides

**Shared Configuration Package** (`config/config.go`):
- Provides base `LoadConfig()` function for non-WebAssembly contexts
- Used by serve command and other utilities
- No WebAssembly dependencies to avoid build constraint conflicts

#### Auto-Generated Settings

- **Redirect URL**: Automatically generated as `http://localhost:{port}/auth/callback` if not explicitly set
- **Port Fallback**: Defaults to "8080" if not configured

## Authentication Flow

### OAuth2/OIDC Architecture

The frontend implements a **hybrid authentication architecture** that avoids CORS issues:

1. **Authorization** → Authentik directly
2. **Token Exchange** → Backend proxy (avoids CORS)
3. **API Calls** → Backend with Bearer tokens

### Authentication Components

#### AuthService (`app/auth_service.go`)
- Handles OAuth2 PKCE flow with Authentik
- Manages token storage in browser localStorage
- Implements token refresh and validation
- Uses structured logging with slog

#### Key Methods:
- `InitiateLogin()` - Redirects to Authentik with PKCE challenge
- `HandleCallback()` - Processes OAuth callback and exchanges code via backend
- `GetAccessToken()` - Retrieves valid token with automatic refresh
- `IsTokenValid()` - Checks token expiration
- `Logout()` - Clears tokens and redirects to Authentik logout

### Authentication Endpoints

#### Frontend OAuth2 Configuration
```go
oauth2Config := &oauth2.Config{
    ClientID:    config.ClientID,
    RedirectURL: config.RedirectURL,
    Scopes:      []string{"openid", "profile", "email", "groups"},
    Endpoint: oauth2.Endpoint{
        AuthURL:  config.AuthURL + "/application/o/authorize/",
        TokenURL: config.BackendURL + "/auth/token", // Backend proxy!
    },
}
```

#### Backend Proxy Endpoints
The backend provides CORS-safe proxy endpoints:
- **`/auth/oidc-config`** - OIDC discovery (proxies to Authentik)
- **`/auth/token`** - Token exchange (proxies to Authentik with proper CORS headers)
- **`/auth/me`** - Current user information

### Token Storage and Management

**WebAssembly localStorage Operations:**
```go
// Store token
localStorage := js.Global().Get("localStorage")
localStorage.Call("setItem", "access_token", tokenJSON)

// Retrieve token
value := localStorage.Call("getItem", "access_token")
```

**Token Lifecycle:**
1. User clicks "Sign In" → Redirect to Authentik
2. Authentik callback → Extract authorization code
3. Exchange code via backend proxy → Receive JWT token
4. Store token in localStorage
5. Use token for authenticated API calls
6. Automatic refresh when token expires

## Build System

### Build Commands

**WebAssembly Build:**
```bash
./bin/web  # Compiles Go to WASM and generates web assets
```

**Desktop Build:**
```bash
go build -o bin/desktop cmd/desktop/main.go
```

**Serve Command:**
```bash
go build -o bin/serve cmd/serve/main.go
./bin/serve  # Serves the WebAssembly app with config-based port
```

### Build Outputs

**WebAssembly Build** (`./bin/web`):
- Outputs to `web/` directory
- Generates `app.wasm`, `app.js`, `index.html`
- Embeds configuration from `app/config/config.toml`
- Uses Cogent Core's web build system

**Serve Command**:
- Loads configuration from filesystem
- Serves static files from `web/` directory
- Uses port from `config.toml` (defaults to 8080)
- Provides fallback routing for SPA behavior

## Project Structure

```
frontend/
├── app/                          # Main application package
│   ├── config/                   # Embedded config for WASM
│   │   └── config.toml          # Configuration file
│   ├── auth_service.go          # OAuth2/OIDC authentication
│   ├── app.go                   # Main app struct and initialization
│   ├── app_methods.go           # UI methods and event handlers
│   ├── config_desktop.go        # Desktop config loading
│   ├── config_wasm.go           # WebAssembly config loading
│   ├── styles.go                # UI styling definitions
│   └── ui_management.go         # UI state management
├── cmd/                         # Build commands
│   ├── desktop/main.go          # Desktop application entry
│   ├── serve/main.go            # Web server for WASM
│   ├── web/main.go              # Web build tool
│   └── webmain/main.go          # WebAssembly entry point
├── config/                      # Shared configuration package
│   └── config.go                # Non-WASM config loading
├── web/                         # Generated web assets
│   ├── app.wasm                 # Compiled WebAssembly
│   ├── app.js                   # Cogent Core JavaScript
│   ├── index.html               # Main HTML page
│   └── 404.html                 # SPA fallback page
├── config.toml                  # Main configuration file
└── bin/                         # Compiled binaries
    ├── desktop                  # Desktop application
    ├── serve                    # Web server
    └── web                      # Web build tool
```

## Configuration Examples

### Development Configuration (`config.toml`)
```toml
# Server Configuration
port = "3000"

# Backend API URL
backend_url = "http://localhost:3001"

# Authentik OIDC Configuration
auth_url = "https://192.168.0.125:30141"
client_id = "VVyph7MnbGDqtiPq8vfgl51ECIO2GcgZ12skA4VR"
# redirect_url is auto-generated as http://localhost:{port}/auth/callback
```

### Environment Variable Overrides
```bash
export NISHIKI_PORT="8080"
export NISHIKI_BACKEND_URL="https://api.example.com"
export NISHIKI_AUTH_URL="https://auth.example.com"
export NISHIKI_CLIENT_ID="production-client-id"
```

## Development Workflow

### Local Development
1. **Configure Authentik** - Set up OAuth2 provider with correct redirect URI
2. **Update config.toml** - Set backend URL, auth URL, client ID, and port
3. **Build WebAssembly** - Run `./bin/web` to compile and generate assets
4. **Start Server** - Run `./bin/serve` to serve the application
5. **Development Loop** - Modify Go code → rebuild → refresh browser

### Authentication Testing
1. **Verify Backend** - Ensure backend is running with matching configuration
2. **Check Authentik** - Confirm OAuth2 application is configured correctly
3. **Test Flow** - Login → callback → token exchange → API calls → logout
4. **Debug Logging** - Check browser console for structured JSON logs

## Key Features

### UI Components
- **Login View** - Authentik sign-in button and branding
- **Dashboard View** - Main navigation and statistics
- **Collections View** - Inventory collection management
- **Groups View** - User group management
- **Profile View** - User information and logout

### State Management
- **View Constants** - Typed view states (ViewLogin, ViewCallback, etc.)
- **Authentication State** - Persistent across page reloads via localStorage
- **Error Handling** - Structured error logging and user feedback
- **Loading States** - Callback processing and async operations

### Cross-Platform Features
- **Responsive UI** - Works on desktop and web
- **Native Performance** - Desktop app with full system integration
- **Web Compatibility** - Browser-based deployment with PWA support
- **Shared Codebase** - Single Go codebase for multiple platforms

## Security Considerations

### Authentication Security
- **PKCE Flow** - Proof Key for Code Exchange for public OAuth2 clients
- **State Verification** - CSRF protection via state parameter validation
- **Token Storage** - Secure browser localStorage with automatic cleanup
- **No Client Secrets** - Frontend is a public OAuth2 client (no secrets)

### CORS Handling
- **Backend Proxy** - All Authentik calls go through backend to avoid CORS
- **Proper Headers** - Backend sets appropriate CORS headers
- **Token Exchange** - Secure server-to-server communication for token exchange

### Build Security
- **Embedded Config** - Configuration baked into WASM binary (no runtime config exposure)
- **Environment Overrides** - Sensitive values can be set via environment variables
- **No Hardcoded Secrets** - All sensitive values externalized to configuration

## Troubleshooting

### Common Issues

**Build Errors:**
- `syscall/js` import conflicts → Use build constraints properly
- Config embed path errors → Ensure `config/config.toml` exists in `app/` directory

**Authentication Errors:**
- CORS errors → Verify backend proxy endpoints are working
- Invalid redirect URI → Check Authentik application configuration matches `redirect_url`
- Token exchange failures → Verify backend can reach Authentik

**Runtime Issues:**
- Config not found → Check file paths and build constraints
- localStorage errors → Verify WebAssembly browser compatibility
- UI not updating → Check view state management and logging

### Debugging Tips
- **Enable Debug Logging** - Set log level to debug for detailed authentication flow
- **Check Browser Console** - WebAssembly logs appear as JSON in browser console
- **Verify Configuration** - Check loaded config values in startup logs
- **Test Backend Directly** - Verify backend `/auth/oidc-config` and `/auth/token` endpoints

## Dependencies

### Core Dependencies
- **Cogent Core** - Cross-platform UI framework
- **golang.org/x/oauth2** - OAuth2 client implementation
- **github.com/spf13/viper** - Configuration management
- **log/slog** - Structured logging

### Build Tools
- **Go 1.21+** - Required for WebAssembly and build constraints
- **Cogent Core CLI** - Web build system and asset generation

### Runtime Dependencies
- **Modern Browser** - WebAssembly support required for web deployment
- **Authentik Server** - OIDC provider for authentication
- **Backend API** - Nishiki backend for data and auth proxy

## Styling Architecture (Cogent Core v0.3.12)

### Design System Integration

The frontend implements a comprehensive styling system that exactly matches the nishiki-frontend React application design tokens and Tailwind CSS configuration. All styling is centralized in `app/styles.go` following a consistent architectural pattern.

### Cogent Core v0.3.12 API Requirements

#### Key API Changes from Previous Versions:
- **Border Radius**: Must use `sides.NewValues(units.Dp(X))` instead of `units.Dp(X)`
- **Import Requirements**: Must import `"cogentcore.org/core/styles/sides"`
- **Position Styling**: Removed - no `s.Position` field available
- **Border Constants**: Use custom `units.Dp()` values instead of style constants

#### Critical Styling Patterns:
```go
// Correct border radius usage
s.Border.Radius = sides.NewValues(units.Dp(10)) // rounded (10px)
s.Border.Radius = sides.NewValues(units.Dp(9999)) // rounded-full

// Required imports
import (
    "cogentcore.org/core/styles/sides"
    "cogentcore.org/core/styles/units"
)
```

### Styling Architecture Principles

#### 1. Centralized Style Functions
All styling is defined in `app/styles.go` using semantic naming:
- **Layout Functions**: `StyleContent*`, `StyleHeader*`, `StyleContainer*`
- **Component Functions**: `StyleButton*`, `StyleCard*`, `StyleForm*`
- **Text Functions**: `StyleText*`, `StyleTitle*`, `StyleLabel*`
- **State Functions**: `StyleHover*`, `StyleActive*`, `StyleDisabled*`

#### 2. Design Token Mapping
Colors, typography, and spacing exactly match nishiki-frontend:
```go
// Colors from nishiki-frontend globals.css
ColorPrimary = #6ab3ab        // --color-primary
ColorAccent = #fcd884         // --color-accent
ColorDanger = #cd5a5a         // --color-danger

// Typography scale from Tailwind config
text-xs = 12px, text-sm = 14px, text-base = 16px
text-lg = 18px, text-xl = 20px, text-2xl = 24px

// Border radius from Tailwind config
rounded = 10px (0.625rem), rounded-full = 9999px
```

#### 3. Mobile-First Responsive Design
Following nishiki-frontend MobileLayout component:
```go
// Main container spacing
s.Padding.Set(units.Dp(48), units.Dp(0), units.Dp(64), units.Dp(0)) // pt-12 pb-16
s.Min.Y.Set(100, units.UnitVh) // min-h-screen

// Content spacing  
s.Padding.Set(units.Dp(24), units.Dp(16)) // pt-6 px-4
s.Gap.Set(units.Dp(8)) // gap-2
```

### Styling Guidelines

#### 1. No Inline Styling
❌ **Never use inline styling:**
```go
// WRONG - Inline styling
btn.Styler(func(s *styles.Style) {
    s.Background = colors.Uniform(ColorPrimary)
    s.Color = colors.Uniform(ColorWhite)
})
```

✅ **Always use style functions:**
```go
// CORRECT - Centralized styling
btn.Styler(StyleButtonPrimary)
```

#### 2. Semantic Naming Convention
- `Style[Component][Variant]` - e.g., `StyleButtonPrimary`, `StyleCardHeader`
- `Style[Layout][Purpose]` - e.g., `StyleContentColumn`, `StyleActionsRow`
- `Style[Text][Size]` - e.g., `StyleTextTitle`, `StyleTextSmall`

#### 3. Component Consistency
All components must use consistent styling patterns:
- **Buttons**: Standardized padding, border radius, colors
- **Cards**: Uniform background, radius, spacing
- **Forms**: Consistent field sizing, validation states
- **Text**: Proper typography hierarchy

### Style Function Categories

#### Layout Functions
- `StyleMainContainer` - Root container with mobile spacing
- `StyleContentColumn` - Main content area
- `StyleHeaderRow` - Fixed header layout
- `StyleActionsRow` - Button/action containers
- `StyleGridContainer` - Grid layout containers

#### Component Functions
- `StyleButtonPrimary/Danger/Accent/Cancel` - Button variants
- `StyleCard/CardHeader/CardContent` - Card components
- `StyleFormField/FormLabel/FormInput` - Form elements
- `StyleDialog/DialogTitle/DialogActions` - Modal dialogs

#### Text Functions
- `StyleAppTitle` - Main application title (24px, bold)
- `StyleSectionTitle` - Section headers (20px, semibold)
- `StyleCardTitle` - Card titles (18px, semibold)
- `StyleDescriptionText` - Secondary text (14px, gray)
- `StyleSmallText` - Helper text (12px, gray)

### Performance Considerations

#### 1. Style Function Reuse
- Minimize unique style functions (target ~60 total)
- Group similar patterns into reusable functions
- Use parameterized functions for variations

#### 2. Cognitive Load Reduction
- Clear, predictable naming patterns
- Logical grouping in `styles.go`
- Comprehensive documentation

### Development Workflow

#### 1. Adding New Components
1. Check existing style functions first
2. If new styling needed, add to `styles.go`
3. Follow naming conventions
4. Match nishiki-frontend design tokens
5. Test responsive behavior

#### 2. Styling Maintenance
1. All style changes go through `styles.go`
2. Verify cross-component consistency
3. Test mobile and desktop layouts
4. Ensure accessibility compliance

This architecture ensures visual consistency with the React frontend while leveraging Cogent Core's powerful styling system efficiently.