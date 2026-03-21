# MCP Feature Audit Guide

A reusable guide for auditing any MCP server implementation against the current spec (2025-06-18) and Go SDK v1.4.0. Use this to identify missing features and prioritize improvements.

## How to Use This Document

1. Read through each feature section
2. Check your server's code against the "How to check" steps
3. Mark each feature as implemented, partially implemented, or missing
4. Use the priority guidance to decide what to implement first

---

## MCP Protocol Version

The Go MCP SDK v1.4.0 implements protocol version **2025-06-18**. Key files in the SDK:

```
mcp/
├── protocol.go      # All protocol types (params, results, capabilities)
├── server.go        # Server implementation, ServerOptions, session management
├── content.go       # Content types (Text, Image, Audio, ResourceLink, etc.)
├── tool.go          # Tool registration, AddTool with generics
├── resource.go      # Resource and ResourceTemplate registration
├── prompt.go        # Prompt registration
├── logging.go       # LoggingHandler (slog.Handler for MCP)
├── features.go      # Internal feature set management
├── session.go       # ServerSession — per-client session with Elicit, CreateMessage
├── streamable.go    # Streamable HTTP transport
├── sse.go           # SSE transport
└── event.go         # Event/notification types
```

---

## Feature Checklist

### 1. Server Instructions

**Spec:** The server can provide an `instructions` string during initialization.
**Purpose:** Gives the LLM client a high-level description of what the server does and how to use it.
**SDK location:** `ServerOptions.Instructions` field.

**How to check:** Look for `NewServer` call — is `ServerOptions` non-nil with `Instructions` set?

```go
mcp.NewServer(impl, &mcp.ServerOptions{
    Instructions: "Description of your server and how to use it",
})
```

**Priority:** Trivial to add, immediate value. Do this first.

---

### 2. Tool Annotations

**Spec:** Each `Tool` can have an `Annotations` field of type `ToolAnnotations`.
**Purpose:** Hints for the client about tool behavior. Clients use these for auto-approval decisions.

**Fields:**
| Field | Type | Default | Meaning |
|-------|------|---------|---------|
| `ReadOnlyHint` | bool | false | Tool doesn't modify anything |
| `DestructiveHint` | *bool | true | Tool may destroy data (only meaningful when not read-only) |
| `IdempotentHint` | bool | false | Calling repeatedly with same args has no additional effect |
| `OpenWorldHint` | *bool | true | Tool interacts with external systems |
| `Title` | string | "" | Human-readable display name |

**How to check:** Grep for `Annotations` in tool registration code. If absent, all tools appear as potentially destructive and open-world to the client.

```go
mcp.AddTool(s, &mcp.Tool{
    Name: "my_tool",
    Annotations: &mcp.ToolAnnotations{
        ReadOnlyHint:    true,
        DestructiveHint: boolPtr(false),
        IdempotentHint:  true,
        OpenWorldHint:   boolPtr(false),
    },
}, handler)
```

**Priority:** Small effort, high UX impact. Auto-approval in Claude Desktop depends on these.

---

### 3. Structured Output (OutputSchema)

**Spec:** Tools can declare an `OutputSchema` (JSON Schema) describing the structure of `StructuredContent` in `CallToolResult`.
**Purpose:** Clients can parse tool results programmatically instead of extracting from text content.

**SDK mechanism:** The `AddTool` generic function uses the `Out` type parameter:
```go
// Signature: AddTool[In, Out any](s *Server, tool *Tool, handler ToolHandlerFor[In, Out])
// Where ToolHandlerFor = func(ctx, req, In) (*CallToolResult, Out, error)

// Without structured output (Out = any):
func handler(ctx context.Context, req *mcp.CallToolRequest, input MyInput) (*mcp.CallToolResult, any, error)

// With structured output:
func handler(ctx context.Context, req *mcp.CallToolRequest, input MyInput) (*mcp.CallToolResult, MyOutput, error)
```

When `Out` is a concrete type (not `any`):
- The SDK infers `OutputSchema` from the type via reflection
- The handler's second return value is marshaled into `StructuredContent`
- If `Content` is nil, it's auto-populated with JSON text of the output

**How to check:** Look at tool handler signatures. If they all return `(*mcp.CallToolResult, any, error)` or build results manually with `TextContent`, structured output is not being used.

**Priority:** Medium effort, high value for reliable tool result parsing.

---

### 4. Resource Subscriptions

**Spec:** Clients can subscribe to resource URIs. When the resource changes, the server sends `notifications/resources/updated`.
**Purpose:** Real-time data freshness. After a mutation, subscribed clients get notified to re-read resources.

**SDK mechanism:**
```go
server := mcp.NewServer(impl, &mcp.ServerOptions{
    SubscribeHandler: func(ctx context.Context, req *mcp.SubscribeRequest) error {
        // req.Params.URI — the resource URI being subscribed to
        return nil
    },
    UnsubscribeHandler: func(ctx context.Context, req *mcp.UnsubscribeRequest) error {
        return nil
    },
})

// After data changes:
server.ResourceUpdated("your://resource/uri")
```

The SDK maintains an internal map of `uri -> set[*ServerSession]` and delivers notifications only to subscribed sessions.

**How to check:** Look for `SubscribeHandler` in `ServerOptions`. If absent, clients cannot subscribe.

**Priority:** Medium effort. Most valuable when your tools mutate data that resources expose.

---

### 5. Completion Handler (Autocomplete)

**Spec:** Clients can request completions for resource template parameters and prompt arguments.
**Purpose:** Better UX — users get suggestions when filling in IDs, names, etc.

**SDK mechanism:**
```go
server := mcp.NewServer(impl, &mcp.ServerOptions{
    CompletionHandler: func(ctx context.Context, req *mcp.CompleteRequest) (*mcp.CompleteResult, error) {
        ref := req.Params.Ref        // { Type: "ref/resource" | "ref/prompt", URI/Name }
        arg := req.Params.Argument   // { Name: "id", Value: "partial..." }

        // Return matching completions
        return &mcp.CompleteResult{
            Completion: mcp.CompletionResult{
                Values:  []string{"suggestion1", "suggestion2"},
                HasMore: false,
            },
        }, nil
    },
})
```

**How to check:** Look for `CompletionHandler` in `ServerOptions`.

**Priority:** Medium effort, good UX improvement for servers with ID-based resource templates.

---

### 6. MCP Logging

**Spec:** Servers can send `notifications/message` log entries to clients. Clients set a minimum log level via `logging/setLevel`.
**Purpose:** Server-side logs visible in the client UI for debugging.

**SDK mechanism:** `mcp.NewLoggingHandler` implements `slog.Handler`:
```go
// Per-session setup (needs *ServerSession):
handler := mcp.NewLoggingHandler(session, &mcp.LoggingHandlerOptions{
    LoggerName:  "my-server",
    MinInterval: 100 * time.Millisecond, // rate limiting
})
logger := slog.New(handler)
logger.Info("something happened", "key", "value")
```

**Log levels (MCP):** debug, info, notice, warning, error, critical, alert, emergency

The handler respects the client's requested log level — it won't send messages below the threshold.

**Capability:** Logging is enabled by default in `ServerOptions` (the default capabilities include `{"logging": {}}`). Setting `Capabilities` to `&mcp.ServerCapabilities{}` disables it.

**How to check:** Look for `NewLoggingHandler` usage. If only using standard `slog` or `log`, MCP logging is not being used.

**Priority:** Small effort, useful for debugging.

---

### 7. Elicitation

**Spec:** Server sends `elicitation/create` to ask the user a question. Client shows a form and returns the answer.
**Purpose:** Interactive server-driven workflows — disambiguation, confirmation, user input during operations.

**SDK mechanism:**
```go
// On a *ServerSession:
result, err := session.Elicit(ctx, &mcp.ElicitParams{
    Message: "Which option do you prefer?",
    RequestedSchema: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "choice": map[string]any{
                "type":        "string",
                "enum":        []string{"option_a", "option_b"},
                "description": "Pick one",
            },
        },
        "required": []string{"choice"},
    },
})
// result.Action: "accept" | "decline" | "cancel"
// result.Content: map with user's answers
```

**Supported schema types for form fields:** string, boolean, number, integer, enum (string with enum array)

**Capability requirement:** Client must declare `elicitation` capability. Check via `session.InitializeParams().Capabilities.Elicitation`.

**How to check:** Search for `Elicit(` calls. If absent, server never asks the user questions.

**Priority:** Medium effort. High value for import/migration workflows.

---

### 8. Sampling (Server-Initiated LLM Requests)

**Spec:** Server sends `sampling/createMessage` to request an LLM completion from the client.
**Purpose:** The server can use the client's LLM for classification, summarization, NL parsing, etc.

**SDK mechanism:**
```go
// Basic sampling:
result, err := session.CreateMessage(ctx, &mcp.CreateMessageParams{
    Messages: []*mcp.SamplingMessage{{
        Role:    "user",
        Content: &mcp.TextContent{Text: "Classify this item: flour 2kg"},
    }},
    MaxTokens: 100,
    // Optional:
    SystemPrompt:   "You are a grocery categorizer. Respond with a single category.",
    Temperature:    0.0,
    IncludeContext: "none", // "none" | "thisServer" | "allServers"
})
// result.Content — the LLM's response
// result.Model — which model was used

// Sampling with tools (parallel tool calls):
result, err := session.CreateMessageWithTools(ctx, &mcp.CreateMessageWithToolsParams{
    Messages: []*mcp.SamplingMessageV2{...},
    MaxTokens: 500,
    Tools: []mcp.Tool{...},       // tools the LLM can use
    ToolChoice: &mcp.ToolChoice{Mode: "auto"},
})
```

**Capability requirement:** Client must declare `sampling` capability. Sub-capabilities:
- `sampling.context` — supports `IncludeContext` values other than "none"
- `sampling.tools` — supports tools and toolChoice in sampling requests

**How to check:** Search for `CreateMessage(` or `CreateMessageWithTools(` calls.

**Priority:** Large effort but very powerful. Design carefully to avoid runaway token costs.

---

### 9. Content Annotations

**Spec:** All content types (`TextContent`, `ImageContent`, `AudioContent`, `ResourceLink`, `EmbeddedResource`) support `Annotations`.
**Purpose:** Metadata hints for the client about content importance and intended audience.

**Fields:**
```go
type Annotations struct {
    Audience     []Role  // "user", "assistant" — who this content is for
    LastModified string  // ISO 8601 timestamp
    Priority     float64 // 0.0 (least important) to 1.0 (most important)
}
```

**Usage:**
```go
&mcp.TextContent{
    Text: data,
    Annotations: &mcp.Annotations{
        Priority: 0.9,
        Audience: []mcp.Role{"user"},
    },
}
```

**How to check:** Grep for `Annotations:` in content construction. If all content is plain `TextContent{Text: ...}`, annotations aren't used.

**Priority:** Small effort, marginal value. Most useful for mixed-audience responses.

---

### 10. Content Types Beyond Text

**Spec types:**
- `TextContent` — plain text
- `ImageContent` — base64 image data with MIME type
- `AudioContent` — base64 audio data with MIME type
- `ResourceLink` — a clickable link to a resource URI
- `EmbeddedResource` — inline resource content

**How to check:** If all tool results and resources only use `TextContent`, you're missing opportunities. For example:
- Tool results could include `ResourceLink` to point to related resources
- Export tools could return `EmbeddedResource` with CSV/JSON content

---

### 11. Progress Notifications

**Spec:** Long-running tool calls can send progress updates via `notifications/progress`.
**Purpose:** Client shows progress bars or status messages during slow operations.

**SDK mechanism:** The request's `Meta` contains a progress token. Use the session to send updates:
```go
// In a tool handler:
token := req.GetProgressToken()
if token != nil {
    session.SendProgress(ctx, token, 50, 100) // 50 of 100
}
```

**How to check:** Look for `SendProgress` or `ProgressToken` usage.

**Priority:** Low effort per tool, useful for bulk operations (import, export).

---

### 12. KeepAlive / Ping

**Spec:** Regular ping requests to detect dead sessions.
**SDK mechanism:**
```go
mcp.NewServer(impl, &mcp.ServerOptions{
    KeepAlive: 30 * time.Second, // ping interval
})
```

If the peer fails to respond, the session is automatically closed.

**How to check:** Look for `KeepAlive` in `ServerOptions`.

**Priority:** Low effort, prevents ghost sessions.

---

## Audit Procedure

1. **Find your `NewServer` call** — check `ServerOptions` for:
   - [ ] `Instructions` set?
   - [ ] `CompletionHandler` set?
   - [ ] `SubscribeHandler` / `UnsubscribeHandler` set?
   - [ ] `KeepAlive` set?
   - [ ] `Capabilities` explicitly configured?

2. **Check each tool registration** for:
   - [ ] `Annotations` with appropriate hints?
   - [ ] Typed `Out` parameter (structured output)?
   - [ ] Progress token handling for long operations?

3. **Check each resource/tool handler** for:
   - [ ] Content annotations on important results?
   - [ ] `ResourceLink` or `EmbeddedResource` where appropriate?
   - [ ] `ResourceUpdated` calls after mutations?

4. **Check session usage** for:
   - [ ] `Elicit` calls for ambiguous operations?
   - [ ] `CreateMessage` for classification/parsing tasks?
   - [ ] `NewLoggingHandler` for debug visibility?

---

## SDK Reference Locations

All paths relative to `$(go env GOMODCACHE)/github.com/modelcontextprotocol/go-sdk@v1.4.0/mcp/`:

| Concept | File | Key Types/Functions |
|---------|------|-------------------|
| Server setup | `server.go` | `NewServer`, `ServerOptions`, `ServerSession` |
| Protocol types | `protocol.go` | `Tool`, `ToolAnnotations`, `Annotations`, `ServerCapabilities`, `ClientCapabilities`, `ElicitParams`, `CreateMessageParams` |
| Content types | `content.go` | `TextContent`, `ImageContent`, `AudioContent`, `ResourceLink`, `EmbeddedResource`, `ToolUseContent`, `ToolResultContent` |
| Tool registration | `tool.go` | `AddTool`, `ToolHandlerFor` |
| Resource registration | `resource.go` | `AddResource`, `AddResourceTemplate` |
| Prompt registration | `prompt.go` | `AddPrompt` |
| Logging | `logging.go` | `NewLoggingHandler`, `LoggingHandlerOptions` |
| Transports | `streamable.go`, `sse.go` | `NewStreamableHTTPHandler`, `NewSSEHandler` |

---

## Claude Desktop Client Capabilities

When connecting to your MCP server, Claude Desktop declares these client capabilities (as of early 2025):

- **Roots** — provides workspace directory roots
- **Sampling** — supports server-initiated LLM requests (with context and tools sub-capabilities)
- **Elicitation** — supports form-based user input requests

To check what a specific client declares, inspect `session.InitializeParams().Capabilities` after initialization.
