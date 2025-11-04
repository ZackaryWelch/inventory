# CLAUDE.md - Nishiki Frontend

This file provides frontend-specific guidance for the Nishiki inventory management system.

## Overview

Cross-platform UI application built with Go + Cogent Core v0.3.12, compiled to:
- **WebAssembly**: Browser-based deployment
- **Desktop**: Native applications (future)

**Key Technologies:**
- Cogent Core v0.3.12 for UI
- OAuth2 PKCE for authentication (public client, no secrets)
- WebAssembly for browser deployment
- Type-safe API clients for backend communication

## Quick Start

```bash
# Build for web
./bin/web

# Serve locally (uses port from config.toml)
./bin/serve

# Run tests
go test ./...

# Format code
gofmt -w .
```

## Configuration

### File: `app/config/config.toml`

```toml
port = "3000"
backend_url = "http://localhost:3001"
auth_url = "https://your-authentik-server.com"
client_id = "your-client-id"
# redirect_url auto-generated as http://localhost:{port}/auth/callback
```

### Build-Specific Behavior

**Desktop** (`app/config_desktop.go`):
- Loads from filesystem via Viper
- Search paths: `"."` and `"./config"`
- Environment overrides: `NISHIKI_PORT`, `NISHIKI_BACKEND_URL`, etc.

**WebAssembly** (`app/config_wasm.go`):
- Embedded at build time: `//go:embed config/config.toml`
- Configuration baked into WASM binary
- Still supports environment variable overrides

## Architecture

### Application Structure

```
app/
├── app.go                 # Application initialization
├── auth_service.go        # OAuth2 PKCE flow
├── collections_ui.go      # Collection management
├── containers_ui.go       # Container tree view
├── objects_ui.go          # Object CRUD
├── ui_management.go       # Groups and navigation
├── ui_helpers.go          # Dialog/form helpers
└── config/
    └── config.toml        # Embedded config for WASM

ui/
├── styles/                # Centralized styling
│   ├── tokens.go          # Design tokens (colors, spacing, fonts)
│   ├── components.go      # Component styles
│   ├── layouts.go         # Layout styles
│   └── utilities.go       # Utility styles
├── components/            # Reusable UI components
└── layouts/               # Layout components

pkg/
├── api/                   # Type-safe API clients
│   ├── auth/
│   ├── collections/
│   ├── containers/
│   ├── objects/
│   └── common/
└── types/                 # Shared domain types

cmd/
├── web/                   # WASM build tool
├── webmain/               # WASM entry point
└── serve/                 # Development server
```

## Authentication Flow (OAuth2 PKCE)

1. **User clicks "Sign In"**
   - Frontend generates PKCE code verifier and challenge
   - Redirects to Authentik with challenge

2. **User authenticates at Authentik**
   - Authentik redirects back with authorization code

3. **Token Exchange via Backend Proxy**
   - Frontend sends code + verifier to backend `/auth/token`
   - Backend exchanges with Authentik (avoids CORS)
   - Returns JWT token to frontend

4. **Token Storage**
   - Stored in browser localStorage
   - Included in all API calls as `Authorization: Bearer {token}`

5. **Automatic Refresh**
   - `GetAccessToken()` checks expiration
   - Refreshes token if needed before API calls

## Styling Conventions

### Core Principles

1. **Never use inline styling** - Always use style functions
2. **Centralized definitions** - All styles in `ui/styles/`
3. **Semantic naming** - `StyleButtonPrimary`, `StyleFormLabel`, etc.
4. **Consistent patterns** - Reuse style functions across components

### Common Patterns

#### Dialog Creation

```go
app.showDialog(DialogConfig{
    Title: "Create Item",
    SubmitButtonText: "Create",
    SubmitButtonStyle: appstyles.StyleButtonPrimary,
    ContentBuilder: func(dialog core.Widget, closeDialog func()) {
        // Use helper functions for all fields
        nameField = createTextField(dialog, "Item name")
        descField = createTextField(dialog, "Description (optional)")

        // Use section headers for organization
        createSectionHeader(dialog, "Additional Details")

        // Use style functions for containers
        propsContainer := core.NewFrame(dialog)
        propsContainer.Styler(appstyles.StylePropertiesContainer)
    },
    OnSubmit: func() {
        // Handle form submission
    },
})
```

#### Form Field Helpers

```go
// Text fields
nameField = createTextField(dialog, "Field label")

// Search fields (includes min width)
searchField = createSearchField(parent, "Search...")

// Section headers
createSectionHeader(dialog, "Section Title")
```

#### Container Styling

```go
// Type selection button containers
typeContainer.Styler(appstyles.StyleTypeButtonContainer)

// Properties containers (with background and padding)
propsContainer.Styler(appstyles.StylePropertiesContainer)

// Group labels with margin
groupLabel.Styler(appstyles.StyleGroupLabelWithMargin)

// Group dropdowns
groupDropdown.Styler(appstyles.StyleGroupDropdownButtonGrow)
```

#### Button Styling

```go
// Primary action buttons
btn.Styler(appstyles.StyleButtonPrimary)

// Accent buttons (important secondary actions)
btn.Styler(appstyles.StyleButtonAccent)

// Danger buttons (destructive actions)
btn.Styler(appstyles.StyleButtonDanger)

// Cancel buttons
btn.Styler(appstyles.StyleButtonCancel)

// Filter buttons
filterBtn.Styler(appstyles.StyleFilterButton)
```

### Creating New Style Functions

When you need new styling:

1. Check if existing style function can be reused
2. If not, add to appropriate file in `ui/styles/`:
   - `tokens.go` - New colors, spacing, fonts
   - `components.go` - Component-specific styles
   - `layouts.go` - Layout styles
   - `utilities.go` - Utility and helper styles
3. Use semantic naming: `Style{Component}{Variant}`
4. Document the pattern in comments

## API Client Usage

### Making API Calls

```go
// Collections
collection, err := app.collectionsClient.Create(types.CreateCollectionRequest{
    Name:       "My Collection",
    ObjectType: "book",
})

// Containers
container, err := app.containersClient.Create(collectionID, types.CreateContainerRequest{
    Name: "Bookshelf",
    Type: "shelf",
})

// Objects
object, err := app.objectsClient.Create(types.CreateObjectRequest{
    Name:        "The Hobbit",
    ContainerID: containerID,
    Properties: map[string]interface{}{
        "author": "J.R.R. Tolkien",
        "isbn":   "978-0547928227",
    },
})
```

### Error Handling

```go
result, err := app.apiClient.SomeMethod(params)
if err != nil {
    app.logger.Error("Operation failed", "error", err)
    core.ErrorSnackbar(app.body, err, "Operation Failed")
    return
}
```

## UI Helper Functions

### Available Helpers (from `ui_helpers.go`)

```go
// Text fields
createTextField(parent, "placeholder")

// Search fields
createSearchField(parent, "Search...")

// Section headers
createSectionHeader(parent, "Section Title")

// Flex rows
createFlexRow(parent, gap, justifyAlign)

// Dialog system
showDialog(DialogConfig{...})
```

## Common Use Cases

### Adding a New Dialog

1. **Define fields** as variables outside dialog
2. **Use `showDialog()`** with `DialogConfig`
3. **Use helper functions** for all fields
4. **Apply style functions** for containers
5. **Implement `OnSubmit`** handler

Example:
```go
func (app *App) showCreateItemDialog() {
    var nameField *core.TextField

    app.showDialog(DialogConfig{
        Title: "Create Item",
        SubmitButtonText: "Create",
        SubmitButtonStyle: appstyles.StyleButtonPrimary,
        ContentBuilder: func(dialog core.Widget, closeDialog func()) {
            nameField = createTextField(dialog, "Item name")
        },
        OnSubmit: func() {
            app.handleCreateItem(nameField.Text())
        },
    })
}
```

### Adding a New View

1. **Create view function**: `func (app *App) showMyView()`
2. **Clear container**: `app.mainContainer.DeleteChildren()`
3. **Set current view**: `app.currentView = "my_view"`
4. **Build UI** using helpers and style functions
5. **Update display**: `app.mainContainer.Update()`

### Async Operations

```go
go func() {
    data, err := app.apiClient.FetchData()
    if err != nil {
        // Handle error
        return
    }

    // Update UI on main thread
    app.mainContainer.AsyncLock()
    defer app.mainContainer.AsyncUnlock()

    // Update UI elements
    container.DeleteChildren()
    for _, item := range data {
        app.createItemCard(container, item)
    }
    container.Update()
}()
```

## Build System

### Commands

```bash
# Build for web (compiles to WASM)
./bin/web

# Serve locally
./bin/serve

# Output location
# - frontend/web/app.wasm
# - frontend/web/app.js
# - frontend/web/index.html
```

### Build Constraints

Use build tags to separate platform-specific code:

```go
//go:build js && wasm

package app

import "syscall/js"  // Only for WebAssembly

// WASM-specific code here
```

```go
//go:build !js || !wasm

package app

// Desktop-specific code here
```

## Troubleshooting

### Build Issues

**Error**: `syscall/js` import conflicts
- **Fix**: Check build constraints are properly set

**Error**: Config embed path not found
- **Fix**: Ensure `config/config.toml` exists in `app/` directory

### Authentication Issues

**CORS Errors**
- **Fix**: Verify backend is running and proxy endpoints work
- Check `/auth/token` and `/auth/oidc-config` endpoints

**Invalid Redirect URI**
- **Fix**: Ensure Authentik OAuth app redirect URI matches config
- Default: `http://localhost:{port}/auth/callback`

### Runtime Issues

**UI Not Updating**
- **Fix**: Call `.Update()` on parent widget after changes
- For async: Use `AsyncLock()` / `AsyncUnlock()`

**Browser Console Errors**
- Check browser console for WASM logs (appear as JSON)
- Enable debug logging for more details

## Dependencies

- **Cogent Core v0.3.12**: UI framework
- **golang.org/x/oauth2**: OAuth2 client
- **github.com/spf13/viper**: Configuration (desktop builds)

## Cogent Core v0.3.12 API Notes

### Key API Patterns

```go
// Border radius (requires sides import)
s.Border.Radius = sides.NewValues(units.Dp(10))

// Required imports
import "cogentcore.org/core/styles/sides"

// Text wrapping
s.Text.WhiteSpace = text.WrapAsNeeded  // Allow wrapping
s.Text.WhiteSpace = text.WrapNever     // No wrapping
```

### Important: No Position Styling

- `s.Position` field is not available in v0.3.12
- Use layout containers and flex properties instead
