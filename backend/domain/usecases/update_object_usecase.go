package usecases

import (
	"context"
	"fmt"

	"github.com/nishiki/backend/domain/entities"
	"github.com/nishiki/backend/domain/repositories"
	"github.com/nishiki/backend/domain/services"
)

type UpdateObjectRequest struct {
	ContainerID *entities.ContainerID // nil = keep current container
	ObjectID    entities.ObjectID
	Name        *string
	Description *string
	Location    *string
	Quantity    *float64
	Unit        *string
	Properties  map[string]interface{}
	Tags        []string
	UserID      entities.UserID
	UserToken   string
}

type UpdateObjectResponse struct {
	Object      *entities.Object
	ContainerID entities.ContainerID
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
	// Find the container that currently holds the object
	currentContainer, err := uc.containerRepo.FindByObjectID(ctx, req.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("object not found: %w", err)
	}

	// Check user access to collection
	userGroups, err := uc.authService.GetUserGroups(ctx, req.UserToken, req.UserID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}

	collection, err := uc.collectionRepo.GetByID(ctx, currentContainer.CollectionID())
	if err != nil {
		return nil, fmt.Errorf("collection not found: %w", err)
	}

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

	// Get existing object from current container
	existingObject, err := currentContainer.GetObject(req.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("object not found in container: %w", err)
	}

	// Apply field updates
	updatedObject := *existingObject

	if req.Name != nil {
		objectName, err := entities.NewObjectName(*req.Name)
		if err != nil {
			return nil, fmt.Errorf("invalid object name: %w", err)
		}
		if err := updatedObject.UpdateName(objectName); err != nil {
			return nil, fmt.Errorf("failed to update object name: %w", err)
		}
	}

	if req.Description != nil {
		desc := entities.NewObjectDescription(*req.Description)
		if err := updatedObject.UpdateDescription(desc); err != nil {
			return nil, fmt.Errorf("failed to update object description: %w", err)
		}
	}

	if req.Location != nil {
		if err := updatedObject.UpdateLocation(*req.Location); err != nil {
			return nil, fmt.Errorf("failed to update object location: %w", err)
		}
	}

	if req.Quantity != nil {
		if err := updatedObject.UpdateQuantity(req.Quantity); err != nil {
			return nil, fmt.Errorf("failed to update object quantity: %w", err)
		}
	}

	if req.Unit != nil {
		if err := updatedObject.UpdateUnit(*req.Unit); err != nil {
			return nil, fmt.Errorf("failed to update object unit: %w", err)
		}
	}

	if req.Properties != nil {
		if err := updatedObject.UpdateProperties(req.Properties); err != nil {
			return nil, fmt.Errorf("failed to update object properties: %w", err)
		}
	}

	if req.Tags != nil {
		if err := updatedObject.UpdateTags(req.Tags); err != nil {
			return nil, fmt.Errorf("failed to update object tags: %w", err)
		}
	}

	// Determine target container
	targetContainer := currentContainer
	if req.ContainerID != nil && !req.ContainerID.Equals(currentContainer.ID()) {
		// Moving to a different container
		targetContainer, err = uc.containerRepo.GetByID(ctx, *req.ContainerID)
		if err != nil {
			return nil, fmt.Errorf("target container not found: %w", err)
		}

		// Remove from old container
		if err := currentContainer.RemoveObject(req.ObjectID); err != nil {
			return nil, fmt.Errorf("failed to remove object from current container: %w", err)
		}
		if err := uc.containerRepo.Update(ctx, currentContainer); err != nil {
			return nil, fmt.Errorf("failed to save old container: %w", err)
		}

		// Add to new container
		if err := targetContainer.AddObject(updatedObject); err != nil {
			return nil, fmt.Errorf("failed to add object to target container: %w", err)
		}
	} else {
		// Update in place
		if err := currentContainer.UpdateObject(req.ObjectID, updatedObject); err != nil {
			return nil, fmt.Errorf("failed to update object in container: %w", err)
		}
	}

	// Save target container
	if err := uc.containerRepo.Update(ctx, targetContainer); err != nil {
		return nil, fmt.Errorf("failed to save container: %w", err)
	}

	return &UpdateObjectResponse{
		Object:      &updatedObject,
		ContainerID: targetContainer.ID(),
	}, nil
}
