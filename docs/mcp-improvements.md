# Nishiki MCP Server — Missing Features & Implementation Plan

Current state: Go MCP SDK v1.4.0, protocol version 2025-06-18.
Server package: `backend/app/mcp/`

## What We Have

| Feature | Status |
|---------|--------|
| Resources (5 static + 7 templates) | Done |
| Tools (20 with annotations) | Done |
| Prompts (7) | Done |
| DB health monitor (log-only) | Done |
| Server Instructions | Done |
| Tool Annotations (all 20 tools) | Done |
| Completion Handler (resource IDs) | Done |
| Resource Subscriptions | Done |

## What Was Implemented

### 1. Server Instructions — Done

**File:** `server.go`

Added `Instructions` string to `ServerOptions` in `NewMCPServer`. Clients get immediate context about the server's purpose during initialization.

### 2. Tool Annotations — Done

**File:** `tools.go`

Added `ToolAnnotations` to all 20 tools using four reusable annotation patterns:

| Pattern | Tools | ReadOnly | Destructive | Idempotent | OpenWorld |
|---------|-------|----------|-------------|------------|-----------|
| `readOnlyAnnotations` | `search_objects`, `export_collection`, `get_collection_schema` | **true** | false | true | false |
| `createAnnotations` | `create_collection`, `create_container`, `create_object`, `create_group`, `join_group`, `bulk_import`, `smart_import` | false | false | false | false |
| `updateAnnotations` | `update_collection`, `update_container`, `update_object`, `update_group`, `add_group_member`, `update_collection_schema` | false | false | true | false |
| `deleteAnnotations` | `delete_collection`, `delete_container`, `delete_object`, `delete_group`, `remove_group_member` | false | *default (true)* | true | false |

Helper `boolPtr(b bool) *bool` used for pointer fields (`DestructiveHint`, `OpenWorldHint`).

### 3. Completion Handler — Done

**File:** `server.go`

`CompletionHandler` in `ServerOptions` provides autocomplete for resource template `{id}` parameters. Queries the database for matching collection, container, and group IDs filtered by the user's typed prefix. Capped at 100 results per MCP spec.

### 4. Resource Subscriptions — Done

**Files:** `server.go`, `context.go`, `tools.go`

- `SubscribeHandler` / `UnsubscribeHandler` added to `ServerOptions` (log subscriptions)
- `MCPContext.Server` field stores the `*mcp.Server` reference (set after `NewMCPServer` returns)
- `MCPContext.notifyResourceUpdated(ctx, uris...)` helper broadcasts `ResourceUpdatedNotificationParams` to subscribed sessions
- Every mutation tool calls `notifyResourceUpdated` after success with relevant URIs:
  - Collection CRUD → `nishiki://collections`, `nishiki://collections/{id}`
  - Container CRUD → `nishiki://containers`, `nishiki://collections/{id}/containers`, `nishiki://containers/{id}`
  - Object CRUD → `nishiki://containers/{id}` (container holding the object)
  - Group CRUD → `nishiki://groups`, `nishiki://groups/{id}`, `nishiki://groups/{id}/users`
  - Import tools → `nishiki://collections/{id}/objects`, `nishiki://containers`
  - Schema update → `nishiki://collections/{id}`

## What's Left

### 5. Structured Output / OutputSchema (medium effort, high value)

**File:** `tools.go`
**What:** Define typed output structs for tools. The SDK auto-populates `StructuredContent` alongside `Content`.
**Why:** Claude can parse tool results reliably instead of guessing JSON structure from text.

Currently we do:
```go
func(ctx context.Context, req *mcp.CallToolRequest, input SearchInput) (*mcp.CallToolResult, any, error) {
    // ... logic ...
    return jsonResult(results)  // returns TextContent with JSON string
}
```

With structured output:
```go
type SearchOutput struct {
    Objects []response.ObjectResponse `json:"objects"`
    Total   int                       `json:"total"`
}

mcp.AddTool(s, &mcp.Tool{
    Name: "search_objects",
    // OutputSchema is auto-inferred from SearchOutput
}, func(ctx context.Context, req *mcp.CallToolRequest, input SearchInput) (*mcp.CallToolResult, SearchOutput, error) {
    // ... logic ...
    return nil, SearchOutput{Objects: results, Total: len(results)}, nil
})
```

The `Out` type parameter on `AddTool` (second return value from handler) drives this. When `Out` is not `any`, the SDK infers `OutputSchema` and marshals the value into `StructuredContent`.

**Note:** Error cases still return `(*CallToolResult, zero-Out, nil)` using `errorResult()`. Need to verify SDK behavior when both `CallToolResult` and `Out` are provided.

### 6. Content Annotations (small effort)

**What:** `Annotations` on `TextContent` / `ImageContent` with `Priority` (0-1) and `Audience` (user/assistant).
**Why:** Helps Claude prioritize which parts of responses matter most.

```go
&mcp.TextContent{
    Text: expiringSoonJSON,
    Annotations: &mcp.Annotations{
        Priority: 0.9,
        Audience: []mcp.Role{"user"},
    },
}
```

Best candidates: search results, expiration warnings, import summaries.

### 7. MCP Logging (small effort)

**File:** `server.go`, handler setup
**What:** Send server-side log messages to the client via MCP's logging notification.
**Why:** Visible in Claude Desktop for debugging. Replaces our log-only MCPNotifier approach.

The SDK provides `mcp.NewLoggingHandler(session, opts)` which implements `slog.Handler`. It sends structured log entries to the connected client at whatever log level the client requests.

This requires per-session setup since the handler needs a `*ServerSession`. Integration point is during session initialization.

### 8. Elicitation (medium effort, interactive workflows)

**What:** Server asks the user a question via Claude Desktop and gets the answer back.
**Why:** Useful for ambiguous operations — e.g., during bulk import, ask which collection to target.
**Caveat:** Requires the client to declare `elicitation` capability. Claude Desktop supports this.

```go
result, err := session.Elicit(ctx, &mcp.ElicitParams{
    Message: "Multiple food collections found. Which one should receive the imported items?",
    RequestedSchema: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "collection": map[string]any{
                "type": "string",
                "enum": []string{"Pantry", "Fridge", "Freezer"},
                "description": "Target collection",
            },
        },
        "required": []string{"collection"},
    },
})
```

### 9. Sampling (large effort, high value)

**What:** Server asks Claude to do LLM inference.
**Why:** Smart import categorization, natural-language object matching, receipt parsing on the server side.

```go
result, err := session.CreateMessage(ctx, &mcp.CreateMessageParams{
    Messages: []*mcp.SamplingMessage{{
        Role:    "user",
        Content: &mcp.TextContent{Text: "Classify this grocery item into a category: " + itemName},
    }},
    MaxTokens: 50,
})
```

This reverses the typical flow — the server requests completions from the client's LLM. Powerful but needs careful design to avoid runaway token usage.

## Suggested Implementation Order (remaining)

1. **Structured Output** — 2-3 hours, reliable result parsing
2. **Content Annotations** — 30 minutes, better prioritization
3. **MCP Logging** — 1 hour, debugging visibility
4. **Elicitation** — 2-3 hours, interactive workflows
5. **Sampling** — 4+ hours, smart categorization
