package mcpserver

import (
	"context"
	"errors"

	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nishiki/backend/app/container"
	"github.com/nishiki/backend/domain/entities"
	"github.com/nishiki/backend/domain/usecases"
)

// MCPContext holds shared state for the MCP server (no per-request auth).
type MCPContext struct {
	Container *container.Container
	Notifier  *MCPNotifier
	Server    *mcp.Server // set after NewMCPServer returns; used for resource notifications
}

// mcpAuthKey is the context key for per-request auth data.
type mcpAuthKey struct{}

type mcpAuth struct {
	User  *entities.User
	Token string
}

// WithMCPUser stores the authenticated user and token in the context.
func WithMCPUser(ctx context.Context, user *entities.User, token string) context.Context {
	return context.WithValue(ctx, mcpAuthKey{}, &mcpAuth{User: user, Token: token})
}

// MCPUserFromContext extracts the authenticated user and token from the context.
func MCPUserFromContext(ctx context.Context) (*entities.User, string, error) {
	auth, ok := ctx.Value(mcpAuthKey{}).(*mcpAuth)
	if !ok || auth == nil {
		return nil, "", errors.New("unauthorized: no MCP auth in context")
	}
	return auth.User, auth.Token, nil
}

// Use case factories — constructed on demand so no state is shared across calls.

func (c *MCPContext) getCollectionsUC() *usecases.GetCollectionsUseCase {
	return usecases.NewGetCollectionsUseCase(c.Container.CollectionRepo, c.Container.AuthService)
}

func (c *MCPContext) createCollectionUC() *usecases.CreateCollectionUseCase {
	return usecases.NewCreateCollectionUseCase(c.Container.CollectionRepo, c.Container.AuthService)
}

func (c *MCPContext) updateCollectionUC() *usecases.UpdateCollectionUseCase {
	return usecases.NewUpdateCollectionUseCase(c.Container.CollectionRepo, c.Container.AuthService)
}

func (c *MCPContext) deleteCollectionUC() *usecases.DeleteCollectionUseCase {
	return usecases.NewDeleteCollectionUseCase(c.Container.CollectionRepo, c.Container.ContainerRepo)
}

func (c *MCPContext) getContainersByCollectionUC() *usecases.GetContainersByCollectionUseCase {
	return usecases.NewGetContainersByCollectionUseCase(c.Container.ContainerRepo, c.Container.CollectionRepo, c.Container.AuthService)
}

func (c *MCPContext) getCollectionObjectsUC() *usecases.GetCollectionObjectsUseCase {
	return usecases.NewGetCollectionObjectsUseCase(c.Container.CollectionRepo, c.Container.ContainerRepo, c.Container.AuthService)
}

func (c *MCPContext) getAllContainersUC() *usecases.GetAllContainersUseCase {
	return usecases.NewGetAllContainersUseCase(c.Container.ContainerRepo, c.Container.AuthService)
}

func (c *MCPContext) getContainerByIDUC() *usecases.GetContainerByIDUseCase {
	return usecases.NewGetContainerByIDUseCase(c.Container.ContainerRepo, c.Container.CollectionRepo, c.Container.AuthService)
}

func (c *MCPContext) getContainersUC() *usecases.GetContainersUseCase {
	return usecases.NewGetContainersUseCase(c.Container.ContainerRepo, c.Container.AuthService)
}

func (c *MCPContext) createContainerUC() *usecases.CreateContainerUseCase {
	return usecases.NewCreateContainerUseCase(c.Container.ContainerRepo, c.Container.CollectionRepo, c.Container.AuthService)
}

func (c *MCPContext) updateContainerUC() *usecases.UpdateContainerUseCase {
	return usecases.NewUpdateContainerUseCase(c.Container.ContainerRepo, c.Container.CollectionRepo, c.Container.AuthService)
}

func (c *MCPContext) deleteContainerUC() *usecases.DeleteContainerUseCase {
	return usecases.NewDeleteContainerUseCase(c.Container.ContainerRepo, c.Container.CollectionRepo, c.Container.AuthService)
}

func (c *MCPContext) createObjectUC() *usecases.CreateObjectUseCase {
	return usecases.NewCreateObjectUseCase(c.Container.ContainerRepo, c.Container.CollectionRepo, c.Container.AuthService)
}

func (c *MCPContext) updateObjectUC() *usecases.UpdateObjectUseCase {
	return usecases.NewUpdateObjectUseCase(c.Container.ContainerRepo, c.Container.CollectionRepo, c.Container.AuthService)
}

func (c *MCPContext) deleteObjectUC() *usecases.DeleteObjectUseCase {
	return usecases.NewDeleteObjectUseCase(c.Container.ContainerRepo, c.Container.CollectionRepo, c.Container.AuthService)
}

func (c *MCPContext) getGroupsUC() *usecases.GetGroupsUseCase {
	return usecases.NewGetGroupsUseCase(c.Container.AuthService)
}

func (c *MCPContext) createGroupUC() *usecases.CreateGroupUseCase {
	return usecases.NewCreateGroupUseCase(c.Container.AuthService)
}

func (c *MCPContext) groupUC() *usecases.GroupUseCase {
	return usecases.NewGroupUseCase(c.Container.AuthService)
}

func (c *MCPContext) bulkImportCollectionUC() *usecases.BulkImportCollectionUseCase {
	return usecases.NewBulkImportCollectionUseCase(c.Container.CollectionRepo, c.Container.ContainerRepo, c.Container.AuthService, c.Container.GetConfig().Import.ReservedColumns, c.Container.ImageSearchService)
}

func (c *MCPContext) updatePropertySchemaUC() *usecases.UpdatePropertySchemaUseCase {
	return usecases.NewUpdatePropertySchemaUseCase(c.Container.CollectionRepo)
}

func (c *MCPContext) exportCollectionUC() *usecases.ExportCollectionUseCase {
	return usecases.NewExportCollectionUseCase(c.Container.CollectionRepo, c.Container.AuthService)
}

// notifyResourceUpdated sends a resource-changed notification to subscribed clients.
// It is a no-op if the server is not yet set.
func (c *MCPContext) notifyResourceUpdated(ctx context.Context, uris ...string) {
	if c.Server == nil {
		return
	}
	for _, uri := range uris {
		if err := c.Server.ResourceUpdated(ctx, &mcp.ResourceUpdatedNotificationParams{URI: uri}); err != nil {
			slog.Warn("MCP: failed to send resource update notification", "uri", uri, "error", err)
		}
	}
}
