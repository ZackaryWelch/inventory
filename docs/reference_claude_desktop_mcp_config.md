---
name: Claude Desktop MCP Configuration
description: How to configure MCP servers in Claude Desktop — transport options, config format, native vs mcp-remote
type: reference
---

## Claude Desktop MCP Config (`claude_desktop_config.json`)

The config file only supports **stdio** transport (local processes via `command`/`args`).
It does **not** support `"url"` or `"type": "http"` — that's Claude Code only.

For remote/HTTP MCP servers, use `npx mcp-remote` as the command.

### mcp-remote Transport Flags

```
--transport http-first   (default) tries Streamable HTTP, falls back to SSE on 404
--transport sse-first    tries SSE first, falls back to HTTP on 405
--transport http-only    Streamable HTTP only
--transport sse-only     SSE only
```

Other flags: `--header "Authorization: Bearer TOKEN"`, `--auth-timeout 60`

### Example Config

```json
{
  "mcpServers": {
    "my-server": {
      "command": "npx",
      "args": ["-y", "mcp-remote", "http://localhost:3003/", "--transport", "sse-only"]
    }
  }
}
```

### Claude Desktop Connectors UI

Settings > Connectors > "Add custom connector" supports remote MCP natively via URL.
Requires HTTPS with CA-signed certs — won't work for `http://localhost`.
Available on Pro, Max, Team, Enterprise plans.

## Claude Code MCP Config

Claude Code supports `"type": "http"` natively:

```json
{
  "mcpServers": {
    "my-server": {
      "type": "http",
      "url": "https://example.com/mcp",
      "headers": { "Authorization": "Bearer TOKEN" }
    }
  }
}
```

CLI: `claude mcp add --transport http my-server https://example.com/mcp`

## Go MCP SDK Handler Paths

- `mcp.NewStreamableHTTPHandler()` and `mcp.NewSSEHandler()` are `http.Handler` implementations
- They handle requests at whatever path they're mounted on (no built-in `/mcp` or `/sse` sub-paths)
- SSE: GET to mounted path creates session (returns `text/event-stream` with endpoint event), POST with `?sessionid=<id>` sends messages
- Streamable HTTP: POST to mounted path; use `StreamableHTTPOptions{Stateless: true}` to skip session ID validation

## Nishiki Port Mapping

| Host Port | Container Port | Transport |
|-----------|---------------|-----------|
| 3001      | 3001          | REST API  |
| 7073      | 3002          | Streamable HTTP |
| 7072      | 3003          | SSE       |
