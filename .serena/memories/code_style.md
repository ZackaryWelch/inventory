# Code Style and Conventions

## Go General
- Standard Go naming conventions (PascalCase for exported, camelCase for unexported)
- No docstrings by default unless explicitly requested
- `gofmt -w .` for formatting
- `golangci-lint run` for linting
- Tab indentation (Go standard) — IMPORTANT: use literal tabs in Edit tool old_string snippets

## Backend Architecture (Clean Architecture + DDD)
- Domain Layer: entities, repository interfaces, use cases, service interfaces
- Application Layer: HTTP controllers, middleware, DI container, config
- Infrastructure Layer: MongoDB repositories, Authentik OIDC
- Controllers have signature `(w http.ResponseWriter, r *http.Request)`
- Middleware pattern: `func(http.Handler) http.Handler`
- Go 1.22+ routing: `mux.HandleFunc("GET /users/{id}", handler)`
- Path params: `r.PathValue("id")`
- Auth context: `httputil.SetContextValue` / `middleware.GetCurrentUser`
- JSON helpers: `httputil.JSON(w, status, data)`, `httputil.Error(w, status, "msg")`, `httputil.DecodeJSON(r, &dto)`
- Container has exported fields: `ContainerRepo`, `CategoryRepo`, `CollectionRepo`, `AuthService`
- `app/http/response/` DTOs reused by MCP resources

## Frontend Architecture (Gio immediate-mode UI)
- Gio rebuilds entire UI each frame; all widget state stored in `GioApp` or `WidgetState`
- Async ops use `ops` channel: goroutine sends `Operation{Type, Data, Err}`, handled in `handleOperation()`
- Build constraints: only `syscall/js` or `//go:embed` files carry build tags
  - WASM files: `//go:build js && wasm`
  - Desktop files: `//go:build !js || !wasm` (filename pattern `*_desktop.go`)
- Adding a new view: (1) add `ViewXxx ViewID` const, (2) add case to `render()`, (3) create `xxx_view.go`

## Frontend Styling (OLD Cogent Core conventions — NO LONGER APPLICABLE)
The frontend was previously Cogent Core but is now Gio (gioui.org v0.9.0).
Use Gio patterns: `ui/theme/` for colors/theme, `ui/widgets/` for custom widgets.

## Mocks
- Generated with `go generate ./domain/...`
- Uses `go.uber.org/mock` (mockgen)
- Regenerate when domain interfaces change

## Testing
- Integration tests use Testcontainers
- Always run `go generate ./domain/...` before running tests
