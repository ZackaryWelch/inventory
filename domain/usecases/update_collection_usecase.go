package usecases

import (
	"context"
	"fmt"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/repositories"
	"github.com/nishiki/backend-go/domain/services"
)

type UpdateCollectionRequest struct {
	CollectionID entities.CollectionID
	UserID       entities.UserID
	Name         *string
	Tags         []string
	Location     *string
	UserToken    string
}

type UpdateCollectionResponse struct {
	Collection *entities.Collection
}

type UpdateCollectionUseCase struct {
	collectionRepo repositories.CollectionRepository
	authService    services.AuthService
}

func NewUpdateCollectionUseCase(collectionRepo repositories.CollectionRepository, authService services.AuthService) *UpdateCollectionUseCase {
	return &UpdateCollectionUseCase{
		collectionRepo: collectionRepo,
		authService:    authService,
	}
}

func (uc *UpdateCollectionUseCase) Execute(ctx context.Context, req UpdateCollectionRequest) (*UpdateCollectionResponse, error) {
	// Get collection
	collection, err := uc.collectionRepo.GetByID(ctx, req.CollectionID)
	if err != nil {
		return nil, fmt.Errorf("collection not found")
	}

	// Validate access - only owner can update
	if !collection.UserID().Equals(req.UserID) {
		return nil, fmt.Errorf("access denied: only collection owner can update")
	}

	// Update name if provided
	if req.Name != nil {
		collectionName, err := entities.NewCollectionName(*req.Name)
		if err != nil {
			return nil, fmt.Errorf("invalid collection name: %w", err)
		}
		if err := collection.UpdateName(collectionName); err != nil {
			return nil, fmt.Errorf("failed to update name: %w", err)
		}
	}

	// Update location if provided (location is stored as string in collection)
	if req.Location != nil {
		// Update location through reconstruction since we don't have direct setter
		collection = entities.ReconstructCollection(
			collection.ID(),
			collection.UserID(),
			collection.GroupID(),
			collection.Name(),
			collection.CategoryID(),
			collection.ObjectType(),
			collection.Containers(),
			req.Tags,      // Use new tags
			*req.Location, // Use new location
			collection.CreatedAt(),
			collection.UpdatedAt(),
		)
	} else if len(req.Tags) > 0 {
		// Update only tags
		collection = entities.ReconstructCollection(
			collection.ID(),
			collection.UserID(),
			collection.GroupID(),
			collection.Name(),
			collection.CategoryID(),
			collection.ObjectType(),
			collection.Containers(),
			req.Tags,
			collection.Location(),
			collection.CreatedAt(),
			collection.UpdatedAt(),
		)
	}

	// Save updated collection
	if err := uc.collectionRepo.Update(ctx, collection); err != nil {
		return nil, fmt.Errorf("failed to update collection: %w", err)
	}

	return &UpdateCollectionResponse{Collection: collection}, nil
}
