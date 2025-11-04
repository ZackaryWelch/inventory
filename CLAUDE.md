# CLAUDE.md

This file provides guidance to Claude Code when working with the Nishiki inventory management system.

## Project Overview

Nishiki is a full-stack inventory management system built entirely in Go:
- **Backend**: RESTful API with Clean Architecture and DDD principles
- **Frontend**: Cogent Core framework compiled to WebAssembly for web deployment
- **Auth**: OAuth2/OIDC via Authentik with group-based access control
- **Database**: MongoDB with embedded document structure

## Quick Start

### Backend Development
```bash
# Run locally (requires MongoDB and Authentik)
go run main.go

# Run with Docker
docker compose up --build

# Run tests
go test ./...

# Format code
gofmt -w .
```

### Frontend Development
```bash
cd frontend

# Build for web (WebAssembly)
./bin/web

# Serve locally
./bin/serve

# Run tests
go test ./...
```

## Architecture

### Backend (Clean Architecture)

**Domain Layer** (`domain/`)
- Entities: User, Group, Collection, Container, Object, Category
- Repository interfaces for data access contracts
- Use cases for business logic orchestration (26 total)
- Service interfaces for external dependencies

**Application Layer** (`app/`)
- HTTP controllers (6 total: Auth, User, Group, Collection, Container, Object)
- Middleware (authentication, logging, CORS, error handling)
- Configuration management (TOML with Viper)
- Dependency injection container

**Infrastructure Layer** (`external/`)
- MongoDB repositories with transaction support
- Authentik OIDC service integration
- Mocks for testing

### Frontend (Cogent Core + WebAssembly)

**Application** (`frontend/app/`)
- `app.go` - Application initialization
- `auth_service.go` - OAuth2 PKCE flow (secure public client auth)
- `collections_ui.go` - Collection management UI
- `objects_ui.go` - Object CRUD operations
- `containers_ui.go` - Container tree view and management
- `ui_management.go` - Groups and navigation
- `ui_helpers.go` - Dialog and form helpers

**UI System** (`frontend/ui/`)
- `styles/` - Centralized styling with design tokens
  - `tokens.go` - Colors, spacing, typography constants
  - `components.go` - Component style functions
  - `layouts.go` - Layout style functions
  - `utilities.go` - Utility style functions
- `components/` - Reusable UI components (Card, Button, Badge, etc.)
- `layouts/` - Application layout components

**API Clients** (`frontend/pkg/api/`)
- Type-safe clients for all backend endpoints
- Common HTTP utilities and error handling

**Build System** (`frontend/cmd/`)
- `web/` - WebAssembly build tool
- `webmain/` - WASM entry point
- `serve/` - Development server

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

## Frontend Styling Conventions

### Centralized Styling
All styling uses helper functions from `ui/styles/`:
- Never use inline styling in component code
- Always use `appstyles.StyleXxx` functions
- Create new style functions for repeated patterns

### Common Patterns

**Dialog Creation:**
```go
app.showDialog(DialogConfig{
    Title: "Dialog Title",
    SubmitButtonText: "Submit",
    SubmitButtonStyle: appstyles.StyleButtonPrimary,
    ContentBuilder: func(dialog core.Widget, closeDialog func()) {
        nameField = createTextField(dialog, "Field label")
    },
    OnSubmit: func() {
        // Handle submission
    },
})
```

**Form Fields:**
```go
// Always use helpers, never inline TextField creation
nameField = createTextField(dialog, "Field name")
searchField = createSearchField(parent, "Search...")
header = createSectionHeader(dialog, "Section Title")
```

**Styling Containers:**
```go
// Use existing style functions
typeContainer.Styler(appstyles.StyleTypeButtonContainer)
propsContainer.Styler(appstyles.StylePropertiesContainer)
groupLabel.Styler(appstyles.StyleGroupLabelWithMargin)
```

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
- Build: `cd frontend && ./bin/web`
- Output: `frontend/web/` directory
- Serve via nginx, Apache, or `./bin/serve`
- Files: `app.wasm`, `wasm_exec.js`, `index.html`

## Technology Stack

- **Language**: Go 1.24+
- **Backend**: Gin (HTTP), MongoDB, Authentik (OIDC), Viper (config), Zap (logging)
- **Frontend**: Cogent Core v0.3.12 (UI), OAuth2 (auth), WebAssembly
- **Testing**: Testcontainers, go-mock
