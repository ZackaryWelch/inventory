# Nishiki Inventory Management System

## Purpose
Full-stack inventory management system built entirely in Go. Manages collections, containers, and objects organized by user groups with OAuth2/OIDC authentication.

## Tech Stack
- **Backend**: Go 1.24, net/http (Go 1.22+ routing), MongoDB v2, Authentik OIDC, Viper config, MCP server
- **Frontend**: Go 1.25, Gio (gioui.org v0.9.0) — immediate-mode UI, targets both WebAssembly (browser) and native desktop
- **Auth**: OAuth2 PKCE via Authentik, group-based access control
- **Database**: MongoDB with embedded documents
- **MCP**: `github.com/modelcontextprotocol/go-sdk v1.4.0` — single binary serves REST (:3001), MCP Streamable HTTP (:3002), MCP SSE (:3003)

## Repository Structure
```
inventory/
├── backend/           # Go backend (module: github.com/nishiki/backend)
│   ├── app/
│   │   ├── config/    # Viper TOML config
│   │   ├── container/ # DI container
│   │   ├── http/      # Controllers, middleware, routes, request/response DTOs
│   │   └── mcp/       # MCP server (package mcpserver)
│   ├── domain/        # Clean Architecture domain layer
│   │   ├── entities/
│   │   ├── adapters/
│   │   ├── repositories/
│   │   ├── services/
│   │   ├── usecases/
│   │   └── util/
│   ├── external/      # MongoDB repos, Authentik OIDC
│   ├── mocks/         # Generated mocks (go generate ./domain/...)
│   └── main.go
├── frontend/          # Go frontend (module: github.com/nishiki/frontend)
│   ├── app/           # GioApp, views, auth, import
│   ├── ui/            # Theme, widgets
│   ├── pkg/           # Type-safe API clients, shared types
│   └── cmd/           # web (WASM build), serve, desktop, gio-webmain
├── docker-compose.yml
└── Dockerfile
```

## Data Model Hierarchy
User → Groups → Collections → Containers → Objects

## Entities
- **Collection**: Stores objects of types: food, books, videogames, music, boardgames, general
- **Container**: Hierarchical physical/logical storage (room, bookshelf, shelf, binder, cabinet, general)
- **Object**: Inventory items with type-specific fields (books→author/ISBN, food→brand/expiration)
- **Group**: Shared access control, multiple users, manages collection permissions

## API Endpoints
- Auth: `/auth/me`, `/auth/token`, `/auth/oidc-config`
- Groups: `/groups`, `/groups/{id}`, `/groups/join`
- Collections: `/accounts/{id}/collections` (CRUD)
- Containers: `/accounts/{id}/collections/{id}/containers` (CRUD)
- Objects: `/accounts/{id}/objects` (CRUD)
- Import: `/accounts/{id}/import`, `/accounts/{id}/collections/{id}/import`
- Categories: `/categories` (CRUD)

## MCP Server (backend/app/mcp/, package mcpserver)
- `context.go` — MCPContext, WithMCPUser/MCPUserFromContext for per-request auth
- `server.go` — NewMCPServer, jsonResult/errorResult/jsonResourceResult helpers
- `resources.go` — 12 resources (5 static + 7 templates)
- `tools.go` — 9 working tools + 4 stubs (join_group, update_group, delete_group stub)
- `prompts.go` — 5 prompts
- `notifications.go` — MCPNotifier with DB health polling
