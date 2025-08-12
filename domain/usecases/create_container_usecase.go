package usecases

import (
	"context"
	"fmt"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/repositories"
	"github.com/nishiki/backend-go/domain/services"
)

type CreateContainerRequest struct {
	CollectionID entities.CollectionID
	Name         string
	UserID       entities.UserID
	UserToken    string
}

type CreateContainerResponse struct {
	Container *entities.Container
}

type CreateContainerUseCase struct {
	containerRepo   repositories.ContainerRepository
	collectionRepo  repositories.CollectionRepository
	authService     services.AuthService
}

func NewCreateContainerUseCase(containerRepo repositories.ContainerRepository, collectionRepo repositories.CollectionRepository, authService services.AuthService) *CreateContainerUseCase {
	return &CreateContainerUseCase{
		containerRepo:  containerRepo,
		collectionRepo: collectionRepo,
		authService:    authService,
	}
}

func (uc *CreateContainerUseCase) Execute(ctx context.Context, req CreateContainerRequest) (*CreateContainerResponse, error) {
	// Check if user is a member of the group by fetching user's groups from Authentik
	userGroups, err := uc.authService.GetUserGroups(ctx, req.UserToken, req.UserID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}

	// Check if the requested collection exists and user has access to it
	collection, err := uc.collectionRepo.GetByID(ctx, req.CollectionID)
	if err != nil {
		return nil, fmt.Errorf("collection not found: %w", err)
	}

	// Check access: user is owner OR user is member of collection's group
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
		return nil, fmt.Errorf("access denied: user does not have access to this collection")
	}

	// Create container name value object
	containerName, err := entities.NewContainerName(req.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid container name: %w", err)
	}

	// Create new container
	container, err := entities.NewContainer(entities.ContainerProps{
		CollectionID: req.CollectionID,
		Name:         containerName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create container entity: %w", err)
	}

	// Save container to repository
	if err := uc.containerRepo.Create(ctx, container); err != nil {
		return nil, fmt.Errorf("failed to save container: %w", err)
	}

	// Note: We no longer need to update the group entity since groups are managed in Authentik
	// Containers are linked to groups via the GroupID field

	return &CreateContainerResponse{
		Container: container,
	}, nil
}
