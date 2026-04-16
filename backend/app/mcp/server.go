package mcpserver

import (
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"log/slog"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nishiki/backend/domain/usecases"
)

// NewMCPServer creates a configured MCP server with all resources, tools, and prompts registered.
func NewMCPServer(mctx *MCPContext) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "nishiki",
		Version: "1.0.0",
	}, &mcp.ServerOptions{
		Instructions: "Nishiki inventory management system. " +
			"Use resources to browse collections, containers, and objects. " +
			"Use tools to create, update, delete, and search inventory. " +
			"Collections belong to groups for shared access.",
		CompletionHandler:  completionHandler(mctx),
		SubscribeHandler:   subscribeHandler(),
		UnsubscribeHandler: unsubscribeHandler(),
	})

	registerResources(server, mctx)
	registerTools(server, mctx)
	registerPrompts(server)

	mctx.Server = server
	return server
}

// completionHandler returns a handler that provides autocomplete for resource template parameters.
func completionHandler(mctx *MCPContext) func(context.Context, *mcp.CompleteRequest) (*mcp.CompleteResult, error) {
	return func(ctx context.Context, req *mcp.CompleteRequest) (*mcp.CompleteResult, error) {
		ref := req.Params.Ref
		arg := req.Params.Argument

		if ref.Type == "ref/resource" && arg.Name == "id" {
			return completeResourceID(ctx, mctx, ref.URI, arg.Value)
		}
		return &mcp.CompleteResult{}, nil
	}
}

// completeResourceID returns matching IDs for resource template parameters.
func completeResourceID(ctx context.Context, mctx *MCPContext, uri, prefix string) (*mcp.CompleteResult, error) {
	user, token, err := MCPUserFromContext(ctx)
	if err != nil {
		return &mcp.CompleteResult{}, nil
	}

	var values []string
	switch {
	case strings.HasPrefix(uri, "nishiki://collections"):
		resp, err := mctx.getCollectionsUC().Execute(ctx, usecases.GetCollectionsRequest{
			UserID: user.ID(), UserToken: token,
		})
		if err != nil {
			return &mcp.CompleteResult{}, nil
		}
		for _, c := range resp.Collections {
			id := c.ID().String()
			if strings.HasPrefix(id, prefix) {
				values = append(values, id)
			}
		}

	case strings.HasPrefix(uri, "nishiki://containers"):
		resp, err := mctx.getAllContainersUC().Execute(ctx, usecases.GetAllContainersRequest{
			UserID: user.ID(), UserToken: token,
		})
		if err != nil {
			return &mcp.CompleteResult{}, nil
		}
		for _, c := range resp.Containers {
			id := c.ID().String()
			if strings.HasPrefix(id, prefix) {
				values = append(values, id)
			}
		}

	case strings.HasPrefix(uri, "nishiki://groups"):
		resp, err := mctx.getGroupsUC().Execute(ctx, usecases.GetGroupsRequest{
			UserID: user.ID(), UserToken: token,
		})
		if err != nil {
			return &mcp.CompleteResult{}, nil
		}
		for _, g := range resp.Groups {
			id := g.ID().String()
			if strings.HasPrefix(id, prefix) {
				values = append(values, id)
			}
		}
	}

	// Cap at 100 results per MCP spec recommendation
	hasMore := len(values) > 100
	if hasMore {
		values = values[:100]
	}

	return &mcp.CompleteResult{
		Completion: mcp.CompletionResultDetails{
			Values:  values,
			HasMore: hasMore,
			Total:   len(values),
		},
	}, nil
}

func subscribeHandler() func(context.Context, *mcp.SubscribeRequest) error {
	return func(_ context.Context, req *mcp.SubscribeRequest) error {
		slog.Info("MCP: client subscribed", "uri", req.Params.URI)
		return nil
	}
}

func unsubscribeHandler() func(context.Context, *mcp.UnsubscribeRequest) error {
	return func(_ context.Context, req *mcp.UnsubscribeRequest) error {
		slog.Info("MCP: client unsubscribed", "uri", req.Params.URI)
		return nil
	}
}

// jsonResult marshals a result to JSON and returns it as MCP text content.
func jsonResult(result any) (*mcp.CallToolResult, error) {
	data, err := json.Marshal(result, jsontext.Multiline(true))
	if err != nil {
		return nil, err
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
	data, err := json.Marshal(result, jsontext.Multiline(true))
	if err != nil {
		return nil, err
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      uri,
			MIMEType: "application/json",
			Text:     string(data),
		}},
	}, nil
}
