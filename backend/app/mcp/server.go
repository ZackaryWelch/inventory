package mcpserver

import (
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewMCPServer creates a configured MCP server with all resources, tools, and prompts registered.
func NewMCPServer(mctx *MCPContext) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "nishiki",
		Version: "1.0.0",
	}, nil)

	registerResources(server, mctx)
	registerTools(server, mctx)
	registerPrompts(server)

	return server
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

// textResult returns an MCP tool result with plain text content.
func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}
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
