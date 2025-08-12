package usecases

import (
	"context"
	"fmt"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/repositories"
	"github.com/nishiki/backend-go/domain/services"
)

type DeleteObjectRequest struct {
	ContainerID entities.ContainerID
	ObjectID    entities.ObjectID
	UserID      entities.UserID
	UserToken   string
}

type DeleteObjectResponse struct {
	Success bool
}

type DeleteObjectUseCase struct {
	containerRepo  repositories.ContainerRepository
	collectionRepo repositories.CollectionRepository
	authService    services.AuthService
}

func NewDeleteObjectUseCase(containerRepo repositories.ContainerRepository, collectionRepo repositories.CollectionRepository, authService services.AuthService) *DeleteObjectUseCase {
	return &DeleteObjectUseCase{
		containerRepo:  containerRepo,
		collectionRepo: collectionRepo,
		authService:    authService,
	}
}

func (uc *DeleteObjectUseCase) Execute(ctx context.Context, req DeleteObjectRequest) (*DeleteObjectResponse, error) {
	// Get container
	container, err := uc.containerRepo.GetByID(ctx, req.ContainerID)
	if err != nil {
		return nil, fmt.Errorf("container not found: %w", err)
	}

	// Check user access to collection
	userGroups, err := uc.authService.GetUserGroups(ctx, req.UserToken, req.UserID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}

	collection, err := uc.collectionRepo.GetByID(ctx, container.CollectionID())
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

	// Remove object from container
	if err := container.RemoveObject(req.ObjectID); err != nil {
		return nil, fmt.Errorf("failed to remove object from container: %w", err)
	}

	// Save updated container
	if err := uc.containerRepo.Update(ctx, container); err != nil {
		return nil, fmt.Errorf("failed to save container: %w", err)
	}

	return &DeleteObjectResponse{
		Success: true,
	}, nil
}