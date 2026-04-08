package usecases

import (
	"context"
	"fmt"

	"github.com/nishiki/backend/domain/entities"
	"github.com/nishiki/backend/domain/repositories"
	"github.com/nishiki/backend/domain/services"
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
	// Resolve user groups (cached) for access check
	userGroups, err := uc.authService.GetUserGroups(ctx, req.UserToken, req.UserID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}

	groupIDs := make([]entities.GroupID, len(userGroups))
	for i, g := range userGroups {
		groupIDs[i] = g.ID()
	}

	// Single aggregation: fetch containers + validate access via $lookup on collection
	containers, err := uc.containerRepo.GetByCollectionIDWithAccess(ctx, req.CollectionID, req.UserID, groupIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get containers for collection: %w", err)
	}

	return &GetContainersByCollectionResponse{
		Containers: containers,
	}, nil
}
