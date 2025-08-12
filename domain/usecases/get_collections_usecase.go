package usecases

import (
	"context"
	"fmt"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/repositories"
	"github.com/nishiki/backend-go/domain/services"
)

type GetCollectionsRequest struct {
	UserID       entities.UserID
	CollectionID *entities.CollectionID // Optional - for single collection
	UserToken    string
}

type GetCollectionsResponse struct {
	Collections []*entities.Collection
}

type GetCollectionsUseCase struct {
	collectionRepo repositories.CollectionRepository
	authService    services.AuthService
}

func NewGetCollectionsUseCase(collectionRepo repositories.CollectionRepository, authService services.AuthService) *GetCollectionsUseCase {
	return &GetCollectionsUseCase{
		collectionRepo: collectionRepo,
		authService:    authService,
	}
}

func (uc *GetCollectionsUseCase) Execute(ctx context.Context, req GetCollectionsRequest) (*GetCollectionsResponse, error) {
	if req.CollectionID != nil {
		// Get single collection with access validation
		collection, err := uc.collectionRepo.GetByID(ctx, *req.CollectionID)
		if err != nil {
			return nil, fmt.Errorf("collection not found")
		}

		// Validate access
		if err := uc.validateCollectionAccess(ctx, collection, req.UserID, req.UserToken); err != nil {
			return nil, err
		}

		return &GetCollectionsResponse{Collections: []*entities.Collection{collection}}, nil
	}

	// Get all collections for user
	collections, err := uc.collectionRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get collections: %w", err)
	}

	return &GetCollectionsResponse{Collections: collections}, nil
}

func (uc *GetCollectionsUseCase) validateCollectionAccess(ctx context.Context, collection *entities.Collection, userID entities.UserID, userToken string) error {
	// User owns collection
	if collection.UserID().Equals(userID) {
		return nil
	}

	// Collection belongs to group - check membership
	if collection.GroupID() != nil {
		userGroups, err := uc.authService.GetUserGroups(ctx, userToken, userID.String())
		if err != nil {
			return fmt.Errorf("failed to get user groups: %w", err)
		}

		for _, group := range userGroups {
			if group.ID().Equals(*collection.GroupID()) {
				return nil
			}
		}
	}

	return fmt.Errorf("access denied")
}