# CLAUDE.md - Nishiki Frontend

Frontend for the Nishiki inventory management system, built with Go + Gio (gioui.org v0.9.0).
Targets both WebAssembly (browser) and native desktop from the same codebase.

## Quick Start

```bash
# Build for web (WebAssembly) — outputs to gio-web/
go run cmd/web/main.go

# Serve the WASM build locally
go run cmd/serve/main.go

# Build native desktop binary
go build ./cmd/desktop/

# Run tests
go test ./...

# Format code
gofmt -w .
```

## Configuration

### File: `app/config/config.toml` (embedded in WASM) / `config.toml` (desktop)

```toml
port = "3000"
backend_url = "http://localhost:3001"
auth_url = "https://your-authentik-server.com"
client_id = "your-client-id"
# redirect_url auto-generated as http://localhost:{port}/auth/callback
```

**WASM** (`app/config_wasm.go`): config baked in at build time via `//go:embed`.
**Desktop** (`app/config_desktop.go`): loaded from filesystem via Viper; env overrides prefixed `NISHIKI_`.

## Architecture

### Application Structure

```
app/
├── gio_app.go                # GioApp struct, event loop, view router
├── auth_service.go           # OAuth2 PKCE — WASM (syscall/js localStorage)
├── auth_service_desktop.go   # OAuth2 PKCE — desktop (system browser + local HTTP)
├── auth_utils.go             # Shared PKCE crypto helpers
├── config_wasm.go            # Config — WASM (embedded toml)
├── config_desktop.go         # Config — desktop (filesystem)
├── js_helpers.go             # Browser URL helpers — WASM (syscall/js)
├── js_helpers_desktop.go     # Browser URL stubs — desktop (no-ops)
├── import_handler.go         # File picker — WASM (DOM FileReader)
├── import_handler_desktop.go # File reader — desktop (os.ReadFile)
├── import_data.go            # CSV/JSON parsing, executeImport (cross-platform)
├── login_wasm.go             # handleLogin() — WASM
├── login_desktop.go          # handleLogin() — desktop (DesktopLogin flow)
├── login_view_simple.go
├── dashboard_view.go
├── groups_view.go
├── collections_view.go
├── collection_detail_view.go
├── collection_detail_dialogs.go
├── import_dialog.go
├── other_views.go            # Profile view, handleLogout
└── config/
    └── config.toml           # Embedded for WASM builds

ui/
├── theme/                    # Gio Material Design theme
│   ├── colors.go
│   └── theme.go
└── widgets/                  # Custom Gio widgets
    ├── button.go
    ├── card.go
    └── dialog.go

pkg/
├── api/                      # Type-safe API clients
│   ├── auth/
│   ├── collections/
│   ├── containers/
│   ├── groups/
│   ├── objects/
│   └── common/
└── types/                    # Shared domain types

cmd/
├── web/                      # WASM build tool (outputs to gio-web/)
├── gio-webmain/              # WASM entry point (js && wasm)
├── serve/                    # Development web server
└── desktop/                  # Desktop native entry point
```

### Build Constraints

Only files that use `syscall/js` or `//go:embed` carry build tags.
All Gio UI code is cross-platform (no build tag required).

| File pattern | Build tag |
|---|---|
| `js_helpers.go`, `auth_service.go`, `import_handler.go` | `js && wasm` |
| `config_wasm.go` | `js && wasm` |
| `*_desktop.go`, `login_desktop.go` | `!js \|\| !wasm` |
| All view files, `gio_app.go`, `import_data.go` | none (cross-platform) |

## Authentication Flow (OAuth2 PKCE)

### WASM
1. User clicks "Sign In" → `InitiateLogin()` stores state/verifier in `localStorage`, redirects browser to Authentik
2. Authentik redirects to `/auth/callback` with `?code=...`
3. App detects callback URL on startup → `HandleCallback()` exchanges code via backend proxy (`/auth/token`)
4. Token stored in `localStorage`; all API calls include `Authorization: Bearer {token}`

### Desktop
1. User clicks "Sign In" → `DesktopLogin()` starts local HTTP server on configured callback port
2. System browser opens to Authentik auth URL
3. Authentik redirects to `http://localhost:{port}/auth/callback`
4. Local server captures code, exchanges via backend proxy, stores token in memory
5. App shows dashboard immediately; no page reload required

## Gio UI Patterns

### Immediate-mode rendering
Gio rebuilds the entire UI each frame. All widget state must be stored explicitly in `GioApp` or `WidgetState`.

```go
// Widget state lives in WidgetState struct (gio_app.go)
type WidgetState struct {
    loginButton widget.Clickable
    // ...
}

// Handle clicks in the render function
if ga.widgetState.loginButton.Clicked(gtx) {
    ga.handleLogin()
}
```

### Async operations
Use the `ops` channel to communicate results from goroutines back to the UI:

```go
go func() {
    data, err := ga.apiClient.Fetch()
    ga.ops <- Operation{Type: "data_loaded", Data: data, Err: err}
}()

// In handleOperation (called from the event loop):
case "data_loaded":
    ga.data = op.Data.([]Item)
    ga.window.Invalidate()
```

### Adding a new view
1. Add a `ViewXxx ViewID` constant to `gio_app.go`
2. Add a `case ViewXxx: return ga.renderXxxView(gtx)` to `render()`
3. Create `xxx_view.go` with `func (ga *GioApp) renderXxxView(gtx layout.Context) layout.Dimensions`

## Dependencies

- **gioui.org v0.9.0**: UI framework (immediate-mode, cross-platform)
- **golang.org/x/oauth2**: OAuth2 client
- **github.com/spf13/viper**: Configuration (desktop builds)

## Troubleshooting

**`xkbcommon-x11` not found** (Linux desktop build):
- Install: `sudo dnf install libxkbcommon-x11-devel` (Fedora) or `sudo apt install libxkbcommon-x11-dev` (Debian)

**Config embed path not found** (WASM build):
- Ensure `app/config/config.toml` exists

**CORS errors** (WASM auth):
- Verify backend is running; check `/auth/token` and `/auth/oidc-config` endpoints

**Desktop login: "failed to start OAuth callback server"**:
- Another process is using the configured port; change `port` in config or kill the conflicting process
