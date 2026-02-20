package mcpserver

import (
	"github.com/nishiki/backend-go/app/container"
	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/usecases"
)

// MCPContext holds all shared state for the MCP server.
type MCPContext struct {
	Container *container.Container
	User      *entities.User
	Token     string
	Notifier  *MCPNotifier
}

func (c *MCPContext) userID() entities.UserID {
	return c.User.ID()
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
	return usecases.NewDeleteCollectionUseCase(c.Container.CollectionRepo)
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

func (c *MCPContext) bulkImportCollectionUC() *usecases.BulkImportCollectionUseCase {
	return usecases.NewBulkImportCollectionUseCase(c.Container.CollectionRepo, c.Container.ContainerRepo, c.Container.AuthService)
}
