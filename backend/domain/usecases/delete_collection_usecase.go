package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/nishiki/backend/domain/entities"
	"github.com/nishiki/backend/domain/repositories"
)

type DeleteCollectionRequest struct {
	CollectionID entities.CollectionID
	UserID       entities.UserID
	Force        bool
}

type DeleteCollectionResponse struct {
	Success           bool
	ContainersDeleted int64
}

type DeleteCollectionUseCase struct {
	collectionRepo repositories.CollectionRepository
	containerRepo  repositories.ContainerRepository
}

func NewDeleteCollectionUseCase(collectionRepo repositories.CollectionRepository, containerRepo repositories.ContainerRepository) *DeleteCollectionUseCase {
	return &DeleteCollectionUseCase{
		collectionRepo: collectionRepo,
		containerRepo:  containerRepo,
	}
}

func (uc *DeleteCollectionUseCase) Execute(ctx context.Context, req DeleteCollectionRequest) (*DeleteCollectionResponse, error) {
	// Get collection to validate ownership
	collection, err := uc.collectionRepo.GetByID(ctx, req.CollectionID)
	if err != nil {
		return nil, errors.New("collection not found")
	}

	// Validate access - only owner can delete
	if !collection.UserID().Equals(req.UserID) {
		return nil, errors.New("access denied: only collection owner can delete")
	}

	// If collection has containers, require force flag
	if collection.ContainerCount() > 0 && !req.Force {
		return nil, errors.New("collection has containers: use force to delete collection and all its containers and objects")
	}

	// Cascade-delete containers (and their embedded objects)
	var containersDeleted int64
	if collection.ContainerCount() > 0 {
		containersDeleted, err = uc.containerRepo.DeleteByCollectionID(ctx, req.CollectionID)
		if err != nil {
			return nil, fmt.Errorf("failed to delete containers: %w", err)
		}
	}

	// Delete collection
	if err := uc.collectionRepo.Delete(ctx, req.CollectionID); err != nil {
		return nil, fmt.Errorf("failed to delete collection: %w", err)
	}

	return &DeleteCollectionResponse{
		Success:           true,
		ContainersDeleted: containersDeleted,
	}, nil
}
