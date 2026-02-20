package usecases

import (
	"context"
	"fmt"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/repositories"
	"github.com/nishiki/backend-go/domain/services"
)

type DeleteContainerRequest struct {
	ContainerID entities.ContainerID
	UserID      entities.UserID
	UserToken   string
}

type DeleteContainerResponse struct {
	Success bool
}

type DeleteContainerUseCase struct {
	containerRepo  repositories.ContainerRepository
	collectionRepo repositories.CollectionRepository
	authService    services.AuthService
}

func NewDeleteContainerUseCase(containerRepo repositories.ContainerRepository, collectionRepo repositories.CollectionRepository, authService services.AuthService) *DeleteContainerUseCase {
	return &DeleteContainerUseCase{
		containerRepo:  containerRepo,
		collectionRepo: collectionRepo,
		authService:    authService,
	}
}

func (uc *DeleteContainerUseCase) Execute(ctx context.Context, req DeleteContainerRequest) (*DeleteContainerResponse, error) {
	// Get container
	container, err := uc.containerRepo.GetByID(ctx, req.ContainerID)
	if err != nil {
		return nil, fmt.Errorf("container not found")
	}

	// Verify user access via collection ownership or group membership
	userGroups, err := uc.authService.GetUserGroups(ctx, req.UserToken, req.UserID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}

	collection, err := uc.collectionRepo.GetByID(ctx, container.CollectionID())
	if err != nil {
		return nil, fmt.Errorf("collection not found")
	}

	hasAccess := collection.UserID().Equals(req.UserID)
	if !hasAccess && collection.GroupID() != nil {
		for _, group := range userGroups {
			if group.ID().Equals(*collection.GroupID()) {
				hasAccess = true
				break
			}
		}
	}

	if !hasAccess {
		return nil, fmt.Errorf("access denied: user does not have access to this container")
	}

	// Business rule: cannot delete a container that has child containers
	children, err := uc.containerRepo.GetChildContainers(ctx, req.ContainerID)
	if err != nil {
		return nil, fmt.Errorf("failed to check child containers: %w", err)
	}
	if len(children) > 0 {
		return nil, fmt.Errorf("cannot delete container with child containers: remove all children first")
	}

	// Remove container reference from collection
	if err := collection.RemoveContainer(req.ContainerID); err != nil {
		return nil, fmt.Errorf("failed to remove container from collection: %w", err)
	}

	if err := uc.collectionRepo.Update(ctx, collection); err != nil {
		return nil, fmt.Errorf("failed to update collection: %w", err)
	}

	// Delete the container document
	if err := uc.containerRepo.Delete(ctx, req.ContainerID); err != nil {
		return nil, fmt.Errorf("failed to delete container: %w", err)
	}

	return &DeleteContainerResponse{Success: true}, nil
}
