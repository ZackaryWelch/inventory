package usecases

import (
	"context"
	"fmt"

	"github.com/nishiki/backend/domain/entities"
	"github.com/nishiki/backend/domain/repositories"
)

type UpdatePropertySchemaRequest struct {
	CollectionID   entities.CollectionID
	UserID         entities.UserID
	PropertySchema *entities.PropertySchema
}

type UpdatePropertySchemaResponse struct {
	Collection *entities.Collection
}

type UpdatePropertySchemaUseCase struct {
	collectionRepo repositories.CollectionRepository
}

func NewUpdatePropertySchemaUseCase(collectionRepo repositories.CollectionRepository) *UpdatePropertySchemaUseCase {
	return &UpdatePropertySchemaUseCase{collectionRepo: collectionRepo}
}

func (uc *UpdatePropertySchemaUseCase) Execute(ctx context.Context, req UpdatePropertySchemaRequest) (*UpdatePropertySchemaResponse, error) {
	collection, err := uc.collectionRepo.GetByID(ctx, req.CollectionID)
	if err != nil {
		return nil, fmt.Errorf("collection not found")
	}

	if !collection.UserID().Equals(req.UserID) {
		return nil, fmt.Errorf("access denied: only collection owner can update schema")
	}

	collection.UpdatePropertySchema(req.PropertySchema)

	if err := uc.collectionRepo.Update(ctx, collection); err != nil {
		return nil, fmt.Errorf("failed to save collection: %w", err)
	}

	return &UpdatePropertySchemaResponse{Collection: collection}, nil
}
