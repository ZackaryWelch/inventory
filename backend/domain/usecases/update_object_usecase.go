package usecases

import (
	"context"
	"fmt"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/repositories"
	"github.com/nishiki/backend-go/domain/services"
)

type UpdateObjectRequest struct {
	ContainerID entities.ContainerID
	ObjectID    entities.ObjectID
	Name        *string
	Properties  map[string]interface{}
	Tags        []string
	UserID      entities.UserID
	UserToken   string
}

type UpdateObjectResponse struct {
	Object *entities.Object
}

type UpdateObjectUseCase struct {
	containerRepo  repositories.ContainerRepository
	collectionRepo repositories.CollectionRepository
	authService    services.AuthService
}

func NewUpdateObjectUseCase(containerRepo repositories.ContainerRepository, collectionRepo repositories.CollectionRepository, authService services.AuthService) *UpdateObjectUseCase {
	return &UpdateObjectUseCase{
		containerRepo:  containerRepo,
		collectionRepo: collectionRepo,
		authService:    authService,
	}
}

func (uc *UpdateObjectUseCase) Execute(ctx context.Context, req UpdateObjectRequest) (*UpdateObjectResponse, error) {
	// Get container
	container, err := uc.containerRepo.GetByID(ctx, req.ContainerID)
	if err != nil {
		return nil, fmt.Errorf("container not found: %w", err)
	}

	// Check user access to collection
	userGroups, err := uc.authService.GetUserGroups(ctx, req.UserToken, req.UserID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}

	collection, err := uc.collectionRepo.GetByID(ctx, container.CollectionID())
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

	// Get existing object
	existingObject, err := container.GetObject(req.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("object not found in container: %w", err)
	}

	// Create updated object
	updatedObject := *existingObject

	// Update name if provided
	if req.Name != nil {
		objectName, err := entities.NewObjectName(*req.Name)
		if err != nil {
			return nil, fmt.Errorf("invalid object name: %w", err)
		}
		if err := updatedObject.UpdateName(objectName); err != nil {
			return nil, fmt.Errorf("failed to update object name: %w", err)
		}
	}

	// Update properties if provided
	if req.Properties != nil {
		if err := updatedObject.UpdateProperties(req.Properties); err != nil {
			return nil, fmt.Errorf("failed to update object properties: %w", err)
		}
	}

	// Update tags if provided
	if req.Tags != nil {
		if err := updatedObject.UpdateTags(req.Tags); err != nil {
			return nil, fmt.Errorf("failed to update object tags: %w", err)
		}
	}

	// Update object in container
	if err := container.UpdateObject(req.ObjectID, updatedObject); err != nil {
		return nil, fmt.Errorf("failed to update object in container: %w", err)
	}

	// Save updated container
	if err := uc.containerRepo.Update(ctx, container); err != nil {
		return nil, fmt.Errorf("failed to save container: %w", err)
	}

	return &UpdateObjectResponse{
		Object: &updatedObject,
	}, nil
}
