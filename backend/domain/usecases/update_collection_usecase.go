package usecases

import (
	"context"
	"fmt"

	"github.com/nishiki/backend/domain/entities"
	"github.com/nishiki/backend/domain/repositories"
	"github.com/nishiki/backend/domain/services"
)

type UpdateCollectionRequest struct {
	CollectionID entities.CollectionID
	UserID       entities.UserID
	Name         *string
	ObjectType   *string
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
		return nil, errors.New("collection not found")
	}

	// Validate access - only owner can update
	if !collection.UserID().Equals(req.UserID) {
		return nil, errors.New("access denied: only collection owner can update")
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

	// Determine effective values for reconstruction
	objectType := collection.ObjectType()
	if req.ObjectType != nil {
		objectType = entities.ObjectType(*req.ObjectType)
	}

	location := collection.Location()
	if req.Location != nil {
		location = *req.Location
	}

	tags := collection.Tags()
	if len(req.Tags) > 0 {
		tags = req.Tags
	}

	// Reconstruct with updated fields
	if req.ObjectType != nil || req.Location != nil || len(req.Tags) > 0 {
		collection = entities.ReconstructCollection(
			collection.ID(),
			collection.UserID(),
			collection.GroupID(),
			collection.Name(),
			collection.CategoryID(),
			objectType,
			collection.Containers(),
			tags,
			location,
			collection.PropertySchema(),
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
