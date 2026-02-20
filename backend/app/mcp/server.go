package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RunMCPServer creates and runs the MCP server on stdio transport.
// The container logger must already be redirected to stderr before calling this.
func RunMCPServer(mctx *MCPContext) {
	notifier := mctx.Notifier
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "nishiki",
		Version: "1.0.0",
	}, &mcp.ServerOptions{
		InitializedHandler: func(ctx context.Context, req *mcp.InitializedRequest) {
			notifier.SetSession(req.Session)
			notifier.StartConnectionMonitor(ctx, mctx)
		},
	})

	registerResources(server, mctx)
	registerTools(server, mctx)
	registerPrompts(server)

	log.Println("[nishiki] MCP server starting on stdio")
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal("[nishiki] MCP server failed:", err)
	}
}

// jsonResult marshals a result to JSON and returns it as MCP text content.
func jsonResult(result any) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, nil
}

// errorResult returns an MCP tool result marked as an error.
func errorResult(err error) (*mcp.CallToolResult, error) {
	r := &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: err.Error()},
		},
	}
	r.SetError(err)
	return r, nil
}

// jsonResourceResult marshals a result and wraps it in an MCP ReadResourceResult.
func jsonResourceResult(uri string, result any) (*mcp.ReadResourceResult, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      uri,
			MIMEType: "application/json",
			Text:     string(data),
		}},
	}, nil
}
