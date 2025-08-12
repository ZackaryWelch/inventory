package usecases

import (
	"context"
	"fmt"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/repositories"
	"github.com/nishiki/backend-go/domain/services"
)

type CreateObjectRequest struct {
	ContainerID entities.ContainerID
	Name        string
	ObjectType  entities.ObjectType
	Properties  map[string]interface{}
	Tags        []string
	UserID      entities.UserID
	UserToken   string
}

type CreateObjectResponse struct {
	Object *entities.Object
}

type CreateObjectUseCase struct {
	containerRepo  repositories.ContainerRepository
	collectionRepo repositories.CollectionRepository
	authService    services.AuthService
}

func NewCreateObjectUseCase(containerRepo repositories.ContainerRepository, collectionRepo repositories.CollectionRepository, authService services.AuthService) *CreateObjectUseCase {
	return &CreateObjectUseCase{
		containerRepo:  containerRepo,
		collectionRepo: collectionRepo,
		authService:    authService,
	}
}

func (uc *CreateObjectUseCase) Execute(ctx context.Context, req CreateObjectRequest) (*CreateObjectResponse, error) {
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

	// Create object name value object
	objectName, err := entities.NewObjectName(req.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid object name: %w", err)
	}

	// Create new object
	object, err := entities.NewObject(entities.ObjectProps{
		Name:       objectName,
		ObjectType: req.ObjectType,
		Properties: req.Properties,
		Tags:       req.Tags,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create object entity: %w", err)
	}

	// Add object to container
	if err := container.AddObject(*object); err != nil {
		return nil, fmt.Errorf("failed to add object to container: %w", err)
	}

	// Save updated container
	if err := uc.containerRepo.Update(ctx, container); err != nil {
		return nil, fmt.Errorf("failed to save container: %w", err)
	}

	return &CreateObjectResponse{
		Object: object,
	}, nil
}