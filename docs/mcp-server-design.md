# MCP Server Design: Nishiki Inventory

## Goal

Replace the Cogent Core/WebAssembly frontend with an MCP server as the primary interface. Inventory management is fundamentally CRUD — creating, finding, organizing, and updating items. This is a natural fit for conversational AI with no visualization requirements.

## Architecture

```
User (natural language)
    ↕
Claude (AI reasoning layer)
    ↕ stdio / MCP protocol
nishiki-mcp (adapter layer)
    ↕ HTTP/JSON + Bearer token
nishiki-backend (existing server, unchanged)
    ↕
MongoDB / Authentik
```

**Key principle**: The MCP server is a thin adapter. All business logic, auth, and persistence stays in the existing backend. The MCP server holds a JWT token and proxies requests.

The MCP server lives in `mcp/` at the root of this repository and communicates with the backend via HTTP.

## Why MCP Over a GUI

Inventory management through a GUI means: navigate to the right collection, open the right container, find the item, click edit, change a field, save. Through MCP:

- **"Move the drill from the garage shelf to the workshop cabinet"** — one sentence, multiple API calls orchestrated automatically
- **"What's expiring this week?"** — query across all collections, filter by date, summarize
- **"Add everything from this receipt to the kitchen pantry"** — bulk import with natural language item descriptions
- **"How many books do I have?"** — aggregate across all collections and containers
- **"Share the workshop inventory with the homelab group"** — find collection, find group, update

The AI handles hierarchy navigation (collection → container → object) that makes GUIs tedious for deep structures. Unlike go-budget (which benefits from chart rendering), Nishiki has **zero visualization requirements** — every interaction is better as text.

## Authentication

The backend uses OIDC via Authentik with JWT Bearer tokens. The MCP server authenticates as a user by holding a pre-configured token.

### Approach: Token Passthrough

User authenticates via Authentik in a browser (one-time), stores the resulting token in MCP config. The MCP server uses the token for all requests.

```json
{
  "mcpServers": {
    "nishiki": {
      "command": "/path/to/nishiki-mcp",
      "args": ["--api-url", "http://localhost:3001"],
      "env": {
        "NISHIKI_TOKEN": "eyJ..."
      }
    }
  }
}
```

**User ID resolution**: Many endpoints require `{user_id}` in the path. The MCP server calls `GET /auth/me` on startup to retrieve the user ID and caches it for all subsequent requests. The user never manages their account ID directly.

**Future improvement**: OAuth2 device authorization flow for automatic token acquisition and refresh.

## MCP Surface Design

### Resources (Read-Only State Inspection)

| Resource URI | Backend endpoint | Description |
|---|---|---|
| `nishiki://health` | `GET /health` | Backend health status |
| `nishiki://me` | `GET /auth/me` | Current user info |
| `nishiki://groups` | `GET /groups` | All groups the user belongs to |
| `nishiki://groups/{id}` | `GET /groups/{id}` | Group details |
| `nishiki://groups/{id}/users` | `GET /groups/{id}/users` | Group members |
| `nishiki://groups/{id}/containers` | `GET /groups/{id}/containers` | Containers in a group |
| `nishiki://collections` | `GET /accounts/{user_id}/collections` | All user collections |
| `nishiki://collections/{id}` | `GET /accounts/{user_id}/collections/{id}` | Collection details |
| `nishiki://collections/{id}/containers` | `GET /accounts/{user_id}/collections/{id}/containers` | Containers in a collection |
| `nishiki://collections/{id}/objects` | `GET /accounts/{user_id}/collections/{id}/objects` | All objects in a collection |
| `nishiki://containers` | `GET /containers` | All containers across all groups |
| `nishiki://containers/{id}` | `GET /containers/{id}` | Container details |

### Tools (State-Modifying Actions)

**Collections**

| Tool | Method + Endpoint | Notes |
|---|---|---|
| `create_collection` | `POST /accounts/{id}/collections` | name, object_type, location, group_id |
| `update_collection` | `PUT /accounts/{id}/collections/{id}` | |
| `delete_collection` | `DELETE /accounts/{id}/collections/{id}` | |

**Containers**

| Tool | Method + Endpoint | Notes |
|---|---|---|
| `create_container` | `POST /accounts/{id}/collections/{id}/containers` | name, type, parent_container_id, capacity, dimensions |
| `update_container` | `PUT /containers/{id}` | Also handles reparenting and group assignment |
| `delete_container` | — | **MISSING: no DELETE endpoint** — see gap analysis |

**Objects**

| Tool | Method + Endpoint | Notes |
|---|---|---|
| `create_object` | `POST /accounts/{id}/objects` | container_id in body |
| `update_object` | `PUT /accounts/{id}/objects/{id}` | **BUG: handler reads container_id from path but URL has no {container_id}** — see gap analysis |
| `delete_object` | `DELETE /accounts/{id}/objects/{id}?container_id={id}` | Requires container_id as query param |
| `bulk_import` | `POST /accounts/{id}/collections/{id}/import` | Use collection-scoped endpoint; `/accounts/{id}/import` is broken — see gap analysis |

**Groups**

| Tool | Method + Endpoint | Notes |
|---|---|---|
| `create_group` | `POST /groups` | |
| `join_group` | `POST /groups/join` | **UNIMPLEMENTED: always returns 501** |
| `update_group` | — | **MISSING: no PUT endpoint** — see gap analysis |
| `delete_group` | — | **MISSING: no DELETE endpoint** — see gap analysis |

### Notifications

| Notification | Trigger | Use |
|---|---|---|
| `items/expiring_soon` | Daily scan for items expiring within 7 days | Proactive alerts for perishables |
| `capacity/high` | Container utilization exceeds 90% | Space management |
| `connection/changed` | Backend becomes unreachable or recovers | Connectivity awareness |

Implementation: the MCP server runs periodic checks (configurable interval, default 1 hour) by reading collections and scanning for expiration dates and capacity thresholds.

### Prompts (Workflow Templates)

**inventory_summary**
- Read all collections and containers
- Count objects per collection and total across all
- Report capacity utilization for containers with limits
- List items expiring soon
- Summarize by object type (food, books, videogames, etc.)

**add_receipt**
- Accept a natural language description of purchased items
- Identify or ask for the target collection and container
- Parse items with quantities, units, and properties
- Bulk import via `POST /accounts/{id}/collections/{id}/import`
- Report what was added

**find_item**
- Accept a natural language description
- Search across all collections, containers, and objects by name, tags, and properties
- Report full location path (collection → container) and quantity

**expiration_check**
- Scan all objects with `expires_at` set
- Group by timeframe: expired, this week, this month, this quarter
- Suggest actions for expired or soon-to-expire items

**reorganize**
- Read container hierarchy and object distribution
- Identify containers that are over/under capacity
- Suggest moves to balance utilization
- Execute moves with user approval

## Backend API CRUD Gap Analysis

This section documents all missing, broken, and incomplete backend API endpoints that affect MCP tool implementation. Gaps must be addressed in the backend before the corresponding MCP tools can be built.

### Summary Table

| Entity | C | R | U | D | Notes |
|--------|---|---|---|---|-------|
| Collection | ✓ | ✓ | ✓ | ✓ | Full CRUD |
| Container | ✓ | ✓ | ✓ | ❌ | No delete endpoint or use case |
| Object | ✓ | ⚠️ | ❌ | ✓ | No get-single; update handler broken; container-scoped GET unfiltered |
| Group | ✓ | ✓ | ❌ | ❌ | No update/delete; join returns 501 |
| Category | ❌ | ❌ | ❌ | ❌ | Entity/repo/DTOs exist, no controller or routes |
| User | — | ✓ | — | — | Managed by Authentik |

### Container: Missing Delete

No `DELETE /containers/{id}` route exists. The `ContainerController` has no `DeleteContainer` method and there is no `DeleteContainerUseCase` in the domain.

**MCP impact**: The `delete_container` tool cannot be implemented.

**Backend work needed**:
1. `DeleteContainerUseCase` in `domain/usecases/`
2. `DeleteContainer` method on `ContainerController`
3. Routes: `DELETE /containers/{container_id}` and `DELETE /accounts/{id}/collections/{collection_id}/containers/{container_id}`
4. `Delete` method on `ContainerRepository` and MongoDB implementation

### Object: No Get-Single

There is no `GET /accounts/{id}/objects/{object_id}` endpoint. Objects can only be retrieved as a collection via `GET /accounts/{id}/collections/{collection_id}/objects`. The MCP server cannot efficiently fetch or verify a single object by ID.

**MCP impact**: Tools that need to display or confirm a specific object must fetch the entire collection and filter client-side.

**Backend work needed**: `GET /accounts/{id}/objects/{object_id}` route and a `GetObjectByIDUseCase`.

### Object: UpdateObject Handler Bug

Route: `PUT /accounts/{id}/objects/{object_id}`

The `UpdateObject` handler (`object_controller.go`) calls `request.GetContainerIDFromPath(r)`, which reads `r.PathValue("container_id")`. This path parameter is **not present in the route pattern** — the URL is `/accounts/{id}/objects/{object_id}` with no `{container_id}` segment. Every call to this endpoint returns HTTP 400 with "missing container ID in path".

**MCP impact**: The `update_object` tool is completely non-functional.

**Fix options**:
- Remove the `GetContainerIDFromPath` call and let the use case resolve the container from the object ID
- Accept `container_id` as a query parameter, consistent with `DeleteObject`

### Object: Container-Scoped GET Unfiltered

Route: `GET /accounts/{id}/collections/{collection_id}/containers/{container_id}/objects`

The handler checks for a `container_id` path value but explicitly ignores it, returning all objects in the collection regardless:

```go
// TODO: Create a GetContainerObjects use case
ctrl.logger.Warn("Container-specific object retrieval not fully implemented yet", ...)
objects = resp.Objects  // returns full collection, not filtered by container
```

**MCP impact**: Fetching objects by container returns incorrect data — Claude would reason about all collection objects when asked about a specific container.

**Backend work needed**: `GetContainerObjectsUseCase` that retrieves objects filtered to a specific container ID.

### Object: BulkImport Endpoint Broken

Route: `POST /accounts/{id}/import`

The `BulkImport` handler calls `request.GetContainerIDFromPath(r)`, but the route has no `{container_id}` path segment. Every call returns HTTP 400. The collection-scoped variant (`POST /accounts/{id}/collections/{collection_id}/import`) works correctly and should be preferred.

**MCP impact**: The global bulk import endpoint is non-functional. The `bulk_import` tool must use the collection-scoped endpoint instead.

**Fix options**: Accept `container_id` from the request body, or deprecate in favor of the collection-scoped endpoint.

### Group: No Update or Delete

No `PUT /groups/{id}` or `DELETE /groups/{id}` routes exist. There are no corresponding use cases.

**MCP impact**: Groups cannot be renamed or deleted through the MCP interface.

**Backend work needed**:
1. `UpdateGroupUseCase` and `DeleteGroupUseCase` in `domain/usecases/`
2. `UpdateGroup` and `DeleteGroup` methods on `GroupController`
3. Routes: `PUT /groups/{id}`, `DELETE /groups/{id}`
4. Corresponding Authentik API calls in `AuthService`

### Group: Join Not Implemented

Route: `POST /groups/join`

The `JoinGroup` handler returns HTTP 501 with "group join not implemented yet". Request parsing and validation exist but the invite hash lookup and Authentik group membership call are not implemented.

**MCP impact**: Users cannot join groups through the MCP interface.

**Backend work needed**: Implement the invite hash system (MongoDB or Authentik metadata) and the Authentik group membership call in `AuthService`.

### Category: No API

The `Category` entity (`domain/entities/category.go`), repository interface (`domain/repositories/category_repository.go`), MongoDB repository (`external/repositories/category_mongo_repository.go`), response DTOs (`app/http/response/category_response.go`), and frontend API client (`frontend/pkg/api/categories/client.go`) all exist. However, there is no `CategoryController`, no category use cases, and no routes registered for categories.

**MCP impact**: Category resources and tools cannot be implemented.

**Backend work needed**: Category use cases, `CategoryController`, and route registration (`GET /categories`, `POST /categories`, `PUT /categories/{id}`, `DELETE /categories/{id}`).

## Interaction Patterns

### Adding Items

**Natural language**: "Add 2 liters of milk to the fridge, expires next Tuesday"

Flow:
1. Fetch `nishiki://collections` to find a food collection
2. Fetch `nishiki://collections/{id}/containers` to resolve "fridge"
3. Call `create_object` with container_id, quantity=2, unit=liter, expires_at=next Tuesday

### Finding Items

**Natural language**: "Where did I put the spare HDMI cables?"

Flow:
1. Fetch `nishiki://containers` to get all containers
2. For each container, fetch objects
3. Filter by name and tags matching "HDMI cables"
4. Report full path: Collection → Container, with quantity

### Bulk Operations

**Natural language**: "I just got back from Costco. I bought: 2 packs of chicken breast, a bag of rice, 3 cans of tomatoes, laundry detergent, and paper towels"

Flow:
1. Identify collections by object type (food vs. household)
2. Ask to confirm target collection/container if ambiguous
3. Call `bulk_import` via `POST /accounts/{id}/collections/{id}/import` for each group

### Organization

**Natural language**: "The garage is getting full. What's in there and what can I move to the basement?"

Flow:
1. Fetch containers with "garage" in name or location
2. List contents with quantities and capacity utilization
3. Check basement containers for available capacity
4. Suggest moves; execute with `update_object` (requires UpdateObject bug fix)

### Sharing

**Natural language**: "Share my book collection with the family group"

Flow:
1. Fetch `nishiki://collections` to find the book collection
2. Fetch `nishiki://groups` to find the family group
3. Call `update_collection` with the group_id

### Expiration Management

The `items/expiring_soon` notification fires daily. Claude proactively reports:

> "3 items in your kitchen pantry expire this week: milk (Feb 17), yogurt (Feb 18), bread (Feb 19)."

## Implementation Notes

### Project Structure

```
mcp/
├── main.go              — MCP server entry point, stdio transport
├── auth.go              — token management, user ID resolution
├── resources.go         — resource handlers (GET proxies)
├── tools.go             — tool handlers (POST/PUT/DELETE proxies)
├── prompts.go           — prompt template definitions
├── notifications.go     — expiration scanning, capacity monitoring
└── go.mod
```

New directory at the root of the inventory repo. The MCP server works primarily with JSON and does not need to import backend domain types.

### Build

Add a `build-mcp` target:

```makefile
build-mcp:
    go build -o bin/nishiki-mcp ./mcp
```

### Configuration

All configuration via environment variables:

| Variable | Description | Default |
|---|---|---|
| `NISHIKI_API_URL` | Backend base URL | `http://localhost:3001` |
| `NISHIKI_TOKEN` | Bearer token | required |
| `NISHIKI_CHECK_INTERVAL` | Notification check interval | `1h` |

### Error Handling

- MCP tools return `isError: true` with the error message from the backend
- HTTP 401 triggers a notification suggesting the user refresh their token
- HTTP 403 is reported with context about which resource was denied
- HTTP 404 is surfaced clearly ("collection not found", "container not found")
- Calls to known-broken endpoints (UpdateObject, BulkImport via `/accounts/{id}/import`) return a descriptive error referencing the gap analysis rather than proxying a misleading 400

### Search

The backend has no dedicated search endpoint. The MCP server implements client-side search:
1. Fetch all collections for the user
2. Fetch containers for each collection
3. Fetch objects for relevant containers
4. Filter by name, tags, properties, description

Acceptable for personal inventory sizes. A server-side search endpoint can be added later if latency becomes a problem.

## Cross-Service Integration with go-budget

With both go-budget and Nishiki running as MCP servers, Claude can orchestrate across them:

- **Receipt-based expense tracking**: Parse receipt items in Nishiki, aggregate costs by category, update go-budget monthly expenses
- **Purchase planning**: "I need to restock the pantry — what will that cost and how does it affect my budget?" — read Nishiki for what's low/expired, estimate costs, run go-budget simulation
- **Shared household budgeting**: Nishiki groups map to shared expense categories in go-budget

No direct integration between services is needed — the AI bridges them through two separate MCP connections.
