# CLAUDE.md

This file provides guidance to Claude Code when working with the Nishiki inventory management system.

## Project Overview

Nishiki is a full-stack inventory management system built entirely in Go:
- **Backend**: RESTful API with Clean Architecture and DDD principles
- **Frontend**: Gio UI framework (gioui.org) compiled to WebAssembly and native desktop
- **Auth**: OAuth2/OIDC via Authentik with group-based access control
- **Database**: MongoDB with embedded document structure

## Quick Start

### Backend Development
```bash
# Run locally (requires MongoDB and Authentik)
go run main.go

# Run with Docker
docker compose up --build

# Generate mocks (required before running tests)
go generate ./domain/...

# Run tests
go test ./...

# Format code
gofmt -w .

# Lint
golangci-lint run
```

### Frontend Development
```bash
cd frontend

# Build for web (WebAssembly) — outputs to gio-web/
go run cmd/web/main.go

# Serve locally
go run cmd/serve/main.go

# Check compilation (faster than full build)
go vet ./...

# Run tests
go test ./...
```

## Architecture

### Backend (Clean Architecture)

**Domain Layer** (`domain/`)
- Entities: User, Group, Collection, Container, Object, Category
- Repository interfaces for data access contracts
- Use cases for business logic orchestration
- Service interfaces for external dependencies

**Application Layer** (`app/`)
- HTTP controllers (6 total: Auth, User, Group, Collection, Container, Object)
- MCP server (`app/mcp/`) — Streamable HTTP (:3002) and SSE (:3003) transports
- Middleware (authentication, logging, CORS, panic recovery)
- HTTP utilities (`httputil/` - JSON helpers, context management, response writer wrapper)
- Configuration management (TOML with Viper)
- Dependency injection container

**Infrastructure Layer** (`external/`)
- MongoDB repositories with transaction support
- Authentik OIDC service integration

**Mocks** (`mocks/`)
- Generated via `go generate ./domain/...`
- Uses mockgen from go.uber.org/mock

### Frontend (Gio + WebAssembly)

See `frontend/CLAUDE.md` for detailed architecture, build constraints, and Gio UI patterns.

**Application** (`frontend/app/`) — `GioApp` struct with immediate-mode Gio rendering
**UI** (`frontend/ui/`) — Theme (colors, Material Design) and custom widgets (Card, Button, Dialog)
**API Clients** (`frontend/pkg/api/`) — Type-safe clients for all backend endpoints
**Build** (`frontend/cmd/`) — `web/` (WASM build), `gio-webmain/` (WASM entry), `serve/` (dev server), `desktop/` (native)

## Data Model

### Hierarchy
```
User → Groups → Collections → Containers → Objects
```

### Entities

**Collection**: Stores objects of specific types (food, books, videogames, music, boardgames, general)
- Has containers for organization
- Belongs to groups for shared access
- Metadata: name, location, object_type

**Container**: Physical or logical storage within collections
- Hierarchical (can have parent containers)
- Types: room, bookshelf, shelf, binder, cabinet, general
- Properties: capacity, dimensions, location

**Object**: Individual inventory items
- Flexible properties based on collection type
- Support for tags, expiration dates, quantities
- Type-specific fields (e.g., books have author/ISBN, food has brand/expiration)

**Group**: Shared access control
- Multiple users can collaborate
- Manages permissions for collections
- Member management and invitations

## Common Use Cases

### Creating a Collection with Objects

1. **Create a Collection**
   - `POST /accounts/{id}/collections`
   - Specify object type and location

2. **Create Containers** (optional)
   - `POST /accounts/{id}/collections/{id}/containers`
   - Organize objects hierarchically

3. **Add Objects**
   - `POST /accounts/{id}/objects`
   - Include container_id and type-specific properties

### Bulk Import

1. **Upload CSV/JSON**
   - `POST /accounts/{id}/import` (creates new collection)
   - `POST /accounts/{id}/collections/{id}/import` (adds to existing)

2. **Review Preview**
   - System parses data and shows preview
   - Displays errors for validation issues

3. **Configure Distribution**
   - Choose automatic or manual container distribution
   - Set container capacity and organization preferences

4. **Execute Import**
   - Creates containers and objects
   - Returns progress and summary

### Group Collaboration

1. **Create/Join Group**
   - `POST /groups` or `POST /groups/join`
   - Share invitation codes

2. **Assign Collections to Group**
   - Set `group_id` when creating/editing collections
   - All group members get access

3. **Manage Members**
   - `GET /groups/{id}/users`
   - `POST /groups/{id}/invite`

## Configuration

### Backend (`app.toml`)

```toml
[server]
port = 3001
debug = true

[database]
uri = "mongodb://localhost:27017"  # or individual host/port/auth fields
database = "nishiki"

[auth]
authentik_url = "https://your-authentik-server.com"
api_token = "your-api-token"

# Multiple OAuth clients (one per frontend deployment)
[[auth.clients]]
provider_name = "nishiki"
client_id = "your-client-id"
client_secret = "your-client-secret"
redirect_url = "http://localhost:3000/auth/callback"

[logging]
level = "info"
```

**Environment Variable Overrides:**
- Prefix with `NISHIKI_` (e.g., `NISHIKI_SERVER_PORT=3001`)
- Use underscores for nesting (e.g., `NISHIKI_DATABASE_URI=...`)

### Frontend (`frontend/app/config/config.toml`)

```toml
port = "3000"
backend_url = "http://localhost:3001"
auth_url = "https://your-authentik-server.com"
client_id = "your-client-id"
# redirect_url auto-generated as http://localhost:{port}/auth/callback
```

**Build-Specific Behavior:**
- **Desktop**: Loads from filesystem
- **WebAssembly**: Embedded at build time via `//go:embed`

## API Endpoints

### Core Resources
- **Auth**: `/auth/me`, `/auth/token`, `/auth/oidc-config`
- **Groups**: `/groups`, `/groups/{id}`, `/groups/join`
- **Collections**: `/accounts/{id}/collections` (CRUD)
- **Containers**: `/accounts/{id}/collections/{id}/containers` (CRUD)
- **Objects**: `/accounts/{id}/objects` (CRUD)
- **Import**: `/accounts/{id}/import`, `/accounts/{id}/collections/{id}/import`
- **Categories**: `/categories` (CRUD)

### Authentication Flow
1. Frontend redirects to Authentik with PKCE challenge
2. User authenticates
3. Callback exchanges code via backend proxy (`/auth/token`)
4. Frontend stores JWT token in localStorage
5. All API calls include `Authorization: Bearer {token}` header

## Testing

### Backend
```bash
# Generate mocks first (if not already done)
go generate ./domain/...

# All tests
go test ./...

# With coverage
go test -cover ./...

# Integration tests
go test ./test/integration/...

# Specific package
go test ./domain/usecases
```

### Frontend
```bash
cd frontend
go test ./...
```

## Deployment

### Backend
- Docker multi-stage build
- Requires MongoDB v5.0+ and Authentik OIDC provider
- Environment variables override config file
- Health check at `/health`

### Frontend
- Build: `cd frontend && go run cmd/web/main.go`
- Output: `frontend/gio-web/` directory
- Serve via nginx, Apache, or `go run cmd/serve/main.go`

## Technology Stack

- **Language**: Go 1.24+
- **Backend**: net/http with Go 1.22+ routing patterns, MongoDB, Authentik (OIDC), Viper (config), slog (logging)
- **Frontend**: Gio v0.9.0 (UI), OAuth2 (auth), WebAssembly + native desktop
- **Testing**: Testcontainers, go-mock

## Backend HTTP Layer

The backend uses Go's standard library `net/http` with Go 1.22+ enhanced routing patterns.

### Package Structure (`app/http/`)
- `controllers/` - HTTP handlers with signature `(w http.ResponseWriter, r *http.Request)`
- `middleware/` - Middleware using `func(http.Handler) http.Handler` pattern
- `httputil/` - Helpers for JSON, context, response writing, and middleware chaining
- `request/` - Request DTOs and path parameter extraction via `r.PathValue()`
- `response/` - Response DTOs
- `routes/` - Route registration with `http.ServeMux`

### Routing Patterns
```go
// Go 1.22+ method-based routing
mux.HandleFunc("GET /users/{id}", handler)
mux.HandleFunc("POST /accounts/{id}/collections", handler)

// Path parameters
userID := r.PathValue("id")
collectionID := r.PathValue("collection_id")
```

### Middleware Pattern
```go
func MyMiddleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // before
            next.ServeHTTP(w, r)
            // after
        })
    }
}
```

### Context for Auth Data
```go
// Setting (in middleware)
r = httputil.SetContextValue(r, httputil.AuthUserKey, user)

// Getting (in handlers)
user, ok := middleware.GetCurrentUser(r)
token, ok := middleware.GetCurrentToken(r)
```

### JSON Helpers
```go
// Response
httputil.JSON(w, http.StatusOK, data)
httputil.Error(w, http.StatusBadRequest, "error message")

// Request parsing
httputil.DecodeJSON(r, &requestStruct)
```
