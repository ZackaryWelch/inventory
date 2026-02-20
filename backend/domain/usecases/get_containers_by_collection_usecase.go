package usecases

import (
	"context"
	"fmt"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/repositories"
	"github.com/nishiki/backend-go/domain/services"
)

type GetContainersByCollectionRequest struct {
	CollectionID entities.CollectionID
	UserID       entities.UserID
	UserToken    string
}

type GetContainersByCollectionResponse struct {
	Containers []*entities.Container
}

type GetContainersByCollectionUseCase struct {
	containerRepo  repositories.ContainerRepository
	collectionRepo repositories.CollectionRepository
	authService    services.AuthService
}

func NewGetContainersByCollectionUseCase(containerRepo repositories.ContainerRepository, collectionRepo repositories.CollectionRepository, authService services.AuthService) *GetContainersByCollectionUseCase {
	return &GetContainersByCollectionUseCase{
		containerRepo:  containerRepo,
		collectionRepo: collectionRepo,
		authService:    authService,
	}
}

func (uc *GetContainersByCollectionUseCase) Execute(ctx context.Context, req GetContainersByCollectionRequest) (*GetContainersByCollectionResponse, error) {
	// Get the collection to verify access
	collection, err := uc.collectionRepo.GetByID(ctx, req.CollectionID)
	if err != nil {
		return nil, fmt.Errorf("collection not found: %w", err)
	}

	// Check if user is a member of the group by fetching user's groups from Authentik
	userGroups, err := uc.authService.GetUserGroups(ctx, req.UserToken, req.UserID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
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

	// Get containers for collection
	containers, err := uc.containerRepo.GetByCollectionID(ctx, req.CollectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get containers for collection: %w", err)
	}

	return &GetContainersByCollectionResponse{
		Containers: containers,
	}, nil
}
