package usecases

import (
	"context"
	"fmt"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/repositories"
	"github.com/nishiki/backend-go/domain/services"
)

type CreateCollectionRequest struct {
	UserID     entities.UserID
	GroupID    *entities.GroupID
	Name       string
	ObjectType entities.ObjectType
	Tags       []string
	Location   string
	UserToken  string
}

type CreateCollectionResponse struct {
	Collection *entities.Collection
}

type CreateCollectionUseCase struct {
	collectionRepo repositories.CollectionRepository
	authService    services.AuthService
}

func NewCreateCollectionUseCase(collectionRepo repositories.CollectionRepository, authService services.AuthService) *CreateCollectionUseCase {
	return &CreateCollectionUseCase{
		collectionRepo: collectionRepo,
		authService:    authService,
	}
}

func (uc *CreateCollectionUseCase) Execute(ctx context.Context, req CreateCollectionRequest) (*CreateCollectionResponse, error) {
	// If GroupID is provided, verify user is member of the group
	if req.GroupID != nil {
		userGroups, err := uc.authService.GetUserGroups(ctx, req.UserToken, req.UserID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to get user groups: %w", err)
		}

		isMember := false
		for _, group := range userGroups {
			if group.ID().Equals(*req.GroupID) {
				isMember = true
				break
			}
		}

		if !isMember {
			return nil, fmt.Errorf("user is not a member of the group")
		}
	}

	// Create collection name value object
	collectionName, err := entities.NewCollectionName(req.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid collection name: %w", err)
	}

	// Create new collection
	collection, err := entities.NewCollection(entities.CollectionProps{
		UserID:     req.UserID,
		GroupID:    req.GroupID,
		Name:       collectionName,
		ObjectType: req.ObjectType,
		Tags:       req.Tags,
		Location:   req.Location,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create collection entity: %w", err)
	}

	// Save collection to repository
	if err := uc.collectionRepo.Create(ctx, collection); err != nil {
		return nil, fmt.Errorf("failed to save collection: %w", err)
	}

	return &CreateCollectionResponse{
		Collection: collection,
	}, nil
}
