package usecases

import (
	"context"
	"fmt"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/repositories"
)

type DeleteCollectionRequest struct {
	CollectionID entities.CollectionID
	UserID       entities.UserID
}

type DeleteCollectionResponse struct {
	Success bool
}

type DeleteCollectionUseCase struct {
	collectionRepo repositories.CollectionRepository
}

func NewDeleteCollectionUseCase(collectionRepo repositories.CollectionRepository) *DeleteCollectionUseCase {
	return &DeleteCollectionUseCase{
		collectionRepo: collectionRepo,
	}
}

func (uc *DeleteCollectionUseCase) Execute(ctx context.Context, req DeleteCollectionRequest) (*DeleteCollectionResponse, error) {
	// Get collection to validate ownership
	collection, err := uc.collectionRepo.GetByID(ctx, req.CollectionID)
	if err != nil {
		return nil, fmt.Errorf("collection not found")
	}

	// Validate access - only owner can delete
	if !collection.UserID().Equals(req.UserID) {
		return nil, fmt.Errorf("access denied: only collection owner can delete")
	}

	// Business rule: Cannot delete collection with containers
	if collection.ContainerCount() > 0 {
		return nil, fmt.Errorf("cannot delete collection with containers: remove all containers first")
	}

	// Delete collection
	if err := uc.collectionRepo.Delete(ctx, req.CollectionID); err != nil {
		return nil, fmt.Errorf("failed to delete collection: %w", err)
	}

	return &DeleteCollectionResponse{Success: true}, nil
}