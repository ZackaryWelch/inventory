package mcpserver

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nishiki/backend-go/app/http/response"
	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/usecases"
)

func registerResources(s *mcp.Server, mctx *MCPContext) {
	// --- Static resources ---

	// nishiki://health
	s.AddResource(&mcp.Resource{
		URI:         "nishiki://health",
		Name:        "health",
		Description: "Server health status",
		MIMEType:    "application/json",
	}, func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return jsonResourceResult(req.Params.URI, map[string]string{"status": "ok"})
	})

	// nishiki://me
	s.AddResource(&mcp.Resource{
		URI:         "nishiki://me",
		Name:        "me",
		Description: "Current authenticated user",
		MIMEType:    "application/json",
	}, func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		u := mctx.User
		result := map[string]any{
			"id":       u.ID().String(),
			"username": u.Username().String(),
			"email":    u.EmailAddress().String(),
		}
		return jsonResourceResult(req.Params.URI, result)
	})

	// nishiki://groups
	s.AddResource(&mcp.Resource{
		URI:         "nishiki://groups",
		Name:        "groups",
		Description: "Groups the current user belongs to",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		resp, err := mctx.getGroupsUC().Execute(ctx, usecases.GetGroupsRequest{
			UserID:    mctx.userID(),
			UserToken: mctx.Token,
		})
		if err != nil {
			return nil, fmt.Errorf("get groups: %w", err)
		}
		return jsonResourceResult(req.Params.URI, response.NewGroupListResponse(resp.Groups))
	})

	// nishiki://collections
	s.AddResource(&mcp.Resource{
		URI:         "nishiki://collections",
		Name:        "collections",
		Description: "All collections owned by or shared with the current user",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		resp, err := mctx.getCollectionsUC().Execute(ctx, usecases.GetCollectionsRequest{
			UserID:    mctx.userID(),
			UserToken: mctx.Token,
		})
		if err != nil {
			return nil, fmt.Errorf("get collections: %w", err)
		}
		return jsonResourceResult(req.Params.URI, response.NewCollectionListResponse(resp.Collections))
	})

	// nishiki://containers
	s.AddResource(&mcp.Resource{
		URI:         "nishiki://containers",
		Name:        "containers",
		Description: "All containers across all groups the current user belongs to",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		resp, err := mctx.getAllContainersUC().Execute(ctx, usecases.GetAllContainersRequest{
			UserID:    mctx.userID(),
			UserToken: mctx.Token,
		})
		if err != nil {
			return nil, fmt.Errorf("get all containers: %w", err)
		}
		return jsonResourceResult(req.Params.URI, response.NewContainerListResponse(resp.Containers))
	})

	// --- Parameterized resource templates ---

	// nishiki://groups/{id}
	s.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "nishiki://groups/{id}",
		Name:        "group",
		Description: "A specific group by ID",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		id := extractID(req.Params.URI, "nishiki://groups/")
		group, err := mctx.Container.AuthService.GetGroupByID(ctx, mctx.Token, id)
		if err != nil {
			return nil, fmt.Errorf("get group: %w", err)
		}
		return jsonResourceResult(req.Params.URI, response.NewGroupResponse(group))
	})

	// nishiki://groups/{id}/users
	s.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "nishiki://groups/{id}/users",
		Name:        "group-users",
		Description: "Members of a specific group",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		id := extractID(req.Params.URI, "nishiki://groups/")
		users, err := mctx.Container.AuthService.GetGroupUsers(ctx, mctx.Token, id)
		if err != nil {
			return nil, fmt.Errorf("get group users: %w", err)
		}
		type userInfo struct {
			ID       string `json:"id"`
			Username string `json:"username"`
			Email    string `json:"email"`
		}
		result := make([]userInfo, len(users))
		for i, u := range users {
			result[i] = userInfo{
				ID:       u.ID().String(),
				Username: u.Username().String(),
				Email:    u.EmailAddress().String(),
			}
		}
		return jsonResourceResult(req.Params.URI, result)
	})

	// nishiki://groups/{id}/containers
	s.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "nishiki://groups/{id}/containers",
		Name:        "group-containers",
		Description: "Containers belonging to a specific group",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		id := extractID(req.Params.URI, "nishiki://groups/")
		groupID, err := entities.GroupIDFromString(id)
		if err != nil {
			return nil, fmt.Errorf("invalid group ID: %w", err)
		}
		resp, err := mctx.getContainersUC().Execute(ctx, usecases.GetContainersRequest{
			GroupID:   groupID,
			UserID:    mctx.userID(),
			UserToken: mctx.Token,
		})
		if err != nil {
			return nil, fmt.Errorf("get group containers: %w", err)
		}
		return jsonResourceResult(req.Params.URI, response.NewContainerListResponse(resp.Containers))
	})

	// nishiki://collections/{id}
	s.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "nishiki://collections/{id}",
		Name:        "collection",
		Description: "A specific collection by ID",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		id := extractID(req.Params.URI, "nishiki://collections/")
		collectionID, err := entities.CollectionIDFromString(id)
		if err != nil {
			return nil, fmt.Errorf("invalid collection ID: %w", err)
		}
		resp, err := mctx.getCollectionsUC().Execute(ctx, usecases.GetCollectionsRequest{
			UserID:       mctx.userID(),
			CollectionID: &collectionID,
			UserToken:    mctx.Token,
		})
		if err != nil {
			return nil, fmt.Errorf("get collection: %w", err)
		}
		if len(resp.Collections) == 0 {
			return nil, mcp.ResourceNotFoundError(req.Params.URI)
		}
		return jsonResourceResult(req.Params.URI, response.NewCollectionResponse(resp.Collections[0]))
	})

	// nishiki://collections/{id}/containers
	s.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "nishiki://collections/{id}/containers",
		Name:        "collection-containers",
		Description: "Containers within a specific collection",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		id := extractID(req.Params.URI, "nishiki://collections/")
		collectionID, err := entities.CollectionIDFromString(id)
		if err != nil {
			return nil, fmt.Errorf("invalid collection ID: %w", err)
		}
		resp, err := mctx.getContainersByCollectionUC().Execute(ctx, usecases.GetContainersByCollectionRequest{
			CollectionID: collectionID,
			UserID:       mctx.userID(),
			UserToken:    mctx.Token,
		})
		if err != nil {
			return nil, fmt.Errorf("get collection containers: %w", err)
		}
		return jsonResourceResult(req.Params.URI, response.NewContainerListResponse(resp.Containers))
	})

	// nishiki://collections/{id}/objects
	s.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "nishiki://collections/{id}/objects",
		Name:        "collection-objects",
		Description: "Objects within a specific collection (across all containers)",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		id := extractID(req.Params.URI, "nishiki://collections/")
		collectionID, err := entities.CollectionIDFromString(id)
		if err != nil {
			return nil, fmt.Errorf("invalid collection ID: %w", err)
		}
		resp, err := mctx.getCollectionObjectsUC().Execute(ctx, usecases.GetCollectionObjectsRequest{
			CollectionID: collectionID,
			UserID:       mctx.userID(),
			UserToken:    mctx.Token,
		})
		if err != nil {
			return nil, fmt.Errorf("get collection objects: %w", err)
		}
		return jsonResourceResult(req.Params.URI, response.NewObjectListResponse(resp.Objects))
	})

	// nishiki://containers/{id}
	s.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "nishiki://containers/{id}",
		Name:        "container",
		Description: "A specific container by ID",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		id := extractID(req.Params.URI, "nishiki://containers/")
		containerID, err := entities.ContainerIDFromString(id)
		if err != nil {
			return nil, fmt.Errorf("invalid container ID: %w", err)
		}
		resp, err := mctx.getContainerByIDUC().Execute(ctx, usecases.GetContainerByIDRequest{
			ContainerID: containerID,
			UserID:      mctx.userID(),
			UserToken:   mctx.Token,
		})
		if err != nil {
			return nil, fmt.Errorf("get container: %w", err)
		}
		return jsonResourceResult(req.Params.URI, response.NewContainerResponse(resp.Container))
	})
}

// extractID parses the first path segment after the given URI prefix.
// e.g., extractID("nishiki://groups/abc/users", "nishiki://groups/") → "abc"
func extractID(uri, prefix string) string {
	rest := strings.TrimPrefix(uri, prefix)
	if idx := strings.Index(rest, "/"); idx >= 0 {
		return rest[:idx]
	}
	return rest
}
