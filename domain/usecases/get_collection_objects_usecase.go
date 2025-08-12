package usecases

import (
	"context"
	"fmt"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/repositories"
	"github.com/nishiki/backend-go/domain/services"
)

type GetCollectionObjectsRequest struct {
	CollectionID entities.CollectionID
	UserID       entities.UserID
	UserToken    string
}

type GetCollectionObjectsResponse struct {
	Objects []entities.Object
}

type GetCollectionObjectsUseCase struct {
	collectionRepo repositories.CollectionRepository
	containerRepo  repositories.ContainerRepository
	authService    services.AuthService
}

func NewGetCollectionObjectsUseCase(collectionRepo repositories.CollectionRepository, containerRepo repositories.ContainerRepository, authService services.AuthService) *GetCollectionObjectsUseCase {
	return &GetCollectionObjectsUseCase{
		collectionRepo: collectionRepo,
		containerRepo:  containerRepo,
		authService:    authService,
	}
}

func (uc *GetCollectionObjectsUseCase) Execute(ctx context.Context, req GetCollectionObjectsRequest) (*GetCollectionObjectsResponse, error) {
	// Check user access to collection
	userGroups, err := uc.authService.GetUserGroups(ctx, req.UserToken, req.UserID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}

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

	// Get all objects from all containers in the collection
	allObjects := collection.GetAllObjects()

	return &GetCollectionObjectsResponse{
		Objects: allObjects,
	}, nil
}