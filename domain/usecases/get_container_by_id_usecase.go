package usecases

import (
	"context"
	"fmt"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/repositories"
	"github.com/nishiki/backend-go/domain/services"
)

type GetContainerByIDRequest struct {
	ContainerID entities.ContainerID
	UserID      entities.UserID
	UserToken   string
}

type GetContainerByIDResponse struct {
	Container *entities.Container
}

type GetContainerByIDUseCase struct {
	containerRepo  repositories.ContainerRepository
	collectionRepo repositories.CollectionRepository
	authService    services.AuthService
}

func NewGetContainerByIDUseCase(containerRepo repositories.ContainerRepository, collectionRepo repositories.CollectionRepository, authService services.AuthService) *GetContainerByIDUseCase {
	return &GetContainerByIDUseCase{
		containerRepo:  containerRepo,
		collectionRepo: collectionRepo,
		authService:    authService,
	}
}

func (uc *GetContainerByIDUseCase) Execute(ctx context.Context, req GetContainerByIDRequest) (*GetContainerByIDResponse, error) {
	// Get container from repository
	container, err := uc.containerRepo.GetByID(ctx, req.ContainerID)
	if err != nil {
		return nil, fmt.Errorf("container not found: %w", err)
	}

	// Check if user is a member of the group that owns the container
	userGroups, err := uc.authService.GetUserGroups(ctx, req.UserToken, req.UserID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to verify user access: %w", err)
	}

	// Get the collection to check access
	collection, err := uc.collectionRepo.GetByID(ctx, container.CollectionID())
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
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
		return nil, fmt.Errorf("access denied: user does not have access to this container's collection")
	}

	return &GetContainerByIDResponse{
		Container: container,
	}, nil
}
