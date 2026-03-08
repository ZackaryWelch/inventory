# Nishiki

Nishiki is a personal inventory management system designed as a self-hosted alternative to Libib and similar collection management services. It tracks what you own, where it is, and lets you organize, search, and bulk-import across multiple collection types (books, food, video games, board games, music, etc.).

The primary interface is an MCP server that lets Claude manage your inventory through natural language. A Gio-based GUI is also in development for visual organization tasks.

## Features

- **Multi-type collections** — books, food, video games, board games, music, and general items
- **Hierarchical organization** — collections → containers → objects, with container capacity tracking
- **Bulk import** — CSV/JSON import with automatic container distribution
- **Expiration tracking** — for food and other perishables, with proactive MCP alerts
- **Group sharing** — share collections across users via Authentik groups
- **MCP server** — full inventory management via Claude (natural language interface)
- **Self-hosted** — no subscription required; runs on your own infrastructure

## Architecture

```
Claude (AI)
    ↕ MCP (stdio or SSE)
nishiki backend --mcp       ← MCP server (embedded in backend binary)
    ↕ HTTP
nishiki backend             ← RESTful API (net/http, Go 1.26)
    ↕
MongoDB         Authentik (OIDC)
```

The backend binary serves two modes:
- `./backend` — HTTP REST API on port 3001
- `./backend --mcp` — MCP server via stdio (for claude-desktop direct use)

The MCP proxy service (docker-compose `nishiki-mcp`) wraps the stdio MCP as an SSE HTTP server on port 3002 for claude-desktop and other SSE-capable MCP clients.

## Quick Start

### Prerequisites

- Docker and Docker Compose
- MongoDB (or use external instance)
- Authentik server for authentication

### Run with Docker

```bash
# Copy and configure
cp app.toml.example app.toml
# Edit app.toml with your Authentik URL, MongoDB URI, and OAuth client config

# Start HTTP backend only
docker compose up --build nishiki-backend

# Start HTTP backend + MCP proxy (for claude-desktop)
NISHIKI_TOKEN=eyJ... docker compose up --build
```

### Run locally (development)

```bash
cd backend

# Generate mocks (required before tests)
go generate ./domain/...

# Run
go run main.go

# Test
go test ./...

# Lint
golangci-lint run
```

## Configuration

### `app.toml`

```toml
[server]
port = 3001
debug = true

[database]
uri = "mongodb://localhost:27017"
database = "nishiki"

[auth]
authentik_url = "https://your-authentik-server.com"
api_token = "your-api-token"

[[auth.clients]]
provider_name = "nishiki"
client_id = "your-client-id"
client_secret = "your-client-secret"
redirect_url = "http://localhost:3000/auth/callback"

[logging]
level = "info"
```

All fields can be overridden with `NISHIKI_` prefixed environment variables (e.g. `NISHIKI_SERVER_PORT=3001`, `NISHIKI_DATABASE_URI=mongodb://...`).

## MCP Server

The MCP server is embedded in the backend binary and exposes resources, tools, and prompts for Claude to manage your inventory.

### claude-desktop configuration

**Option A — SSE via docker (recommended for always-on use)**

```json
{
  "mcpServers": {
    "nishiki": {
      "type": "sse",
      "url": "http://localhost:3002/sse"
    }
  }
}
```

Requires `docker compose up` with `NISHIKI_TOKEN` set.

**Option B — stdio (local binary)**

```json
{
  "mcpServers": {
    "nishiki": {
      "command": "/path/to/backend",
      "args": ["--mcp"],
      "env": {
        "NISHIKI_TOKEN": "eyJ...",
        "NISHIKI_API_URL": "http://localhost:3001"
      }
    }
  }
}
```

### Getting a token

Authenticate via the frontend or Authentik directly, then copy the JWT from browser localStorage (`nishiki_token`).

### MCP capabilities

**Resources** (read-only state):
- `nishiki://me`, `nishiki://groups`, `nishiki://collections`, `nishiki://collections/{id}/objects`
- `nishiki://containers`, `nishiki://containers/{id}`, and more

**Tools** (state-modifying):
- Collections: `create_collection`, `update_collection`, `delete_collection`
- Containers: `create_container`, `update_container`
- Objects: `create_object`, `update_object`, `delete_object`, `bulk_import`
- Groups: `create_group`

**Prompts** (workflow templates):
- `inventory_summary` — full overview with capacity and expiration status
- `add_receipt` — parse purchased items and bulk-add to a collection
- `find_item` — locate an item across all collections
- `expiration_check` — scan for expired and soon-to-expire items
- `reorganize` — suggest container reorganization based on capacity

## API

### Endpoints

| Resource | Endpoints |
|---|---|
| Auth | `GET /auth/me`, `POST /auth/token`, `GET /auth/oidc-config` |
| Groups | `GET /groups`, `POST /groups`, `GET /groups/{id}`, `GET /groups/{id}/users` |
| Collections | `GET/POST /accounts/{id}/collections`, `GET/PUT/DELETE /accounts/{id}/collections/{id}` |
| Containers | `GET/POST /accounts/{id}/collections/{id}/containers`, `GET/PUT /containers/{id}` |
| Objects | `GET /accounts/{id}/collections/{id}/objects`, `POST /accounts/{id}/objects`, `PUT/DELETE /accounts/{id}/objects/{id}` |
| Import | `POST /accounts/{id}/collections/{id}/import` |
| Categories | `GET /categories`, `POST /categories`, `PUT/DELETE /categories/{id}` |
| Health | `GET /health` |

### OpenAPI

OpenAPI spec is available in `backend/documents/`.

## Ecosystem Integration

Nishiki is designed to work alongside other self-hosted services:

| System | Role |
|---|---|
| **Nishiki** | What you own, where it is |
| **Grocy** (via MCP) | Recipe planning, meal plans, shopping lists (food side) |
| **Libib** (via MCP) | Import existing Libib library data |

Claude bridges these services through multiple MCP connections — no direct API integration needed.

See `docs/recipe-planning.md` for notes on the recipe/meal planning architecture.

## Frontend

A Gio-based GUI is in development for visual organization tasks. The MCP server is the primary interface for day-to-day use.

See `frontend/GIO_MIGRATION_PLAN.md` for the current frontend plan.

```bash
cd frontend

# Build for web (WebAssembly)
./bin/web

# Serve locally
./bin/serve

# Run tests
go test ./...
```

## Technology Stack

- **Language**: Go 1.26
- **Backend**: `net/http` with Go 1.22+ routing patterns
- **Database**: MongoDB
- **Auth**: Authentik (OIDC/OAuth2), JWT validation
- **Config**: TOML + Viper with `NISHIKI_` env var overrides
- **Frontend**: Gio UI (migrating from Cogent Core), compiled to WebAssembly
- **MCP**: `github.com/modelcontextprotocol/go-sdk`
- **Testing**: Testcontainers, go-mock (`go.uber.org/mock`)

## License

[License information]
