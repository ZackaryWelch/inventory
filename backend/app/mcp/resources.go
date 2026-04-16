package mcpserver

import (
	"context"
	"log/slog"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nishiki/backend/app/http/response"
	"github.com/nishiki/backend/domain/entities"
	"github.com/nishiki/backend/domain/usecases"
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
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		user, _, err := MCPUserFromContext(ctx)
		if err != nil {
			return nil, err
		}
		result := map[string]any{
			"id":       user.ID().String(),
			"username": user.Username().String(),
			"email":    user.EmailAddress().String(),
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
		user, token, err := MCPUserFromContext(ctx)
		if err != nil {
			return nil, err
		}
		resp, err := mctx.getGroupsUC().Execute(ctx, usecases.GetGroupsRequest{
			UserID:    user.ID(),
			UserToken: token,
		})
		if err != nil {
			slog.Error("failed to get groups", "err", err)
			return nil, err
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
		user, token, err := MCPUserFromContext(ctx)
		if err != nil {
			return nil, err
		}
		resp, err := mctx.getCollectionsUC().Execute(ctx, usecases.GetCollectionsRequest{
			UserID:    user.ID(),
			UserToken: token,
		})
		if err != nil {
			slog.Error("failed to get collections", "err", err)
			return nil, err
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
		user, token, err := MCPUserFromContext(ctx)
		if err != nil {
			return nil, err
		}
		resp, err := mctx.getAllContainersUC().Execute(ctx, usecases.GetAllContainersRequest{
			UserID:    user.ID(),
			UserToken: token,
		})
		if err != nil {
			slog.Error("failed to get all containers", "err", err)
			return nil, err
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
		_, token, err := MCPUserFromContext(ctx)
		if err != nil {
			return nil, err
		}
		id := extractID(req.Params.URI, "nishiki://groups/")
		group, err := mctx.Container.AuthService.GetGroupByID(ctx, token, id)
		if err != nil {
			slog.Error("failed to get group", "group_id", id, "err", err)
			return nil, err
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
		_, token, err := MCPUserFromContext(ctx)
		if err != nil {
			return nil, err
		}
		id := extractID(req.Params.URI, "nishiki://groups/")
		users, err := mctx.Container.AuthService.GetGroupUsers(ctx, token, id)
		if err != nil {
			slog.Error("failed to get group users", "group_id", id, "err", err)
			return nil, err
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
		user, token, err := MCPUserFromContext(ctx)
		if err != nil {
			return nil, err
		}
		id := extractID(req.Params.URI, "nishiki://groups/")
		groupID, err := entities.GroupIDFromString(id)
		if err != nil {
			slog.Error("invalid group ID", "group_id", id, "err", err)
			return nil, ErrInvalidFormat.With(map[string]any{"field": "group_id", "value": id}).Wrap(err)
		}
		resp, err := mctx.getContainersUC().Execute(ctx, usecases.GetContainersRequest{
			GroupID:   groupID,
			UserID:    user.ID(),
			UserToken: token,
		})
		if err != nil {
			slog.Error("failed to get group containers", "group_id", id, "err", err)
			return nil, err
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
		user, token, err := MCPUserFromContext(ctx)
		if err != nil {
			return nil, err
		}
		id := extractID(req.Params.URI, "nishiki://collections/")
		collectionID, err := entities.CollectionIDFromString(id)
		if err != nil {
			slog.Error("invalid collection ID", "collection_id", id, "err", err)
			return nil, ErrInvalidFormat.With(map[string]any{"field": "collection_id", "value": id}).Wrap(err)
		}
		resp, err := mctx.getCollectionsUC().Execute(ctx, usecases.GetCollectionsRequest{
			UserID:       user.ID(),
			CollectionID: &collectionID,
			UserToken:    token,
		})
		if err != nil {
			slog.Error("failed to get collection", "collection_id", id, "err", err)
			return nil, err
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
		user, token, err := MCPUserFromContext(ctx)
		if err != nil {
			return nil, err
		}
		id := extractID(req.Params.URI, "nishiki://collections/")
		collectionID, err := entities.CollectionIDFromString(id)
		if err != nil {
			slog.Error("invalid collection ID", "collection_id", id, "err", err)
			return nil, ErrInvalidFormat.With(map[string]any{"field": "collection_id", "value": id}).Wrap(err)
		}
		resp, err := mctx.getContainersByCollectionUC().Execute(ctx, usecases.GetContainersByCollectionRequest{
			CollectionID: collectionID,
			UserID:       user.ID(),
			UserToken:    token,
		})
		if err != nil {
			slog.Error("failed to get collection containers", "collection_id", id, "err", err)
			return nil, err
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
		user, token, err := MCPUserFromContext(ctx)
		if err != nil {
			return nil, err
		}
		id := extractID(req.Params.URI, "nishiki://collections/")
		collectionID, err := entities.CollectionIDFromString(id)
		if err != nil {
			slog.Error("invalid collection ID", "collection_id", id, "err", err)
			return nil, ErrInvalidFormat.With(map[string]any{"field": "collection_id", "value": id}).Wrap(err)
		}
		resp, err := mctx.getCollectionObjectsUC().Execute(ctx, usecases.GetCollectionObjectsRequest{
			CollectionID: collectionID,
			UserID:       user.ID(),
			UserToken:    token,
		})
		if err != nil {
			slog.Error("failed to get collection objects", "collection_id", id, "err", err)
			return nil, err
		}
		objectResponses := make([]response.ObjectResponse, len(resp.Objects))
		for i, item := range resp.Objects {
			objectResponses[i] = response.NewObjectResponse(item.Object, item.ContainerID.String())
		}
		return jsonResourceResult(req.Params.URI, response.ObjectListResponse{Objects: objectResponses, Total: len(objectResponses)})
	})

	// nishiki://containers/{id}
	s.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "nishiki://containers/{id}",
		Name:        "container",
		Description: "A specific container by ID",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		user, token, err := MCPUserFromContext(ctx)
		if err != nil {
			return nil, err
		}
		id := extractID(req.Params.URI, "nishiki://containers/")
		containerID, err := entities.ContainerIDFromString(id)
		if err != nil {
			slog.Error("invalid container ID", "container_id", id, "err", err)
			return nil, ErrInvalidFormat.With(map[string]any{"field": "container_id", "value": id}).Wrap(err)
		}
		resp, err := mctx.getContainerByIDUC().Execute(ctx, usecases.GetContainerByIDRequest{
			ContainerID: containerID,
			UserID:      user.ID(),
			UserToken:   token,
		})
		if err != nil {
			slog.Error("failed to get container", "container_id", id, "err", err)
			return nil, err
		}
		return jsonResourceResult(req.Params.URI, response.NewContainerResponse(resp.Container))
	})
}

// extractID parses the first path segment after the given URI prefix.
// e.g., extractID("nishiki://groups/abc/users", "nishiki://groups/") → "abc"
func extractID(uri, prefix string) string {
	rest := strings.TrimPrefix(uri, prefix)
	if before, _, ok := strings.Cut(rest, "/"); ok {
		return before
	}
	return rest
}
