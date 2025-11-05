package usecases

import (
	"context"
	"fmt"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/repositories"
	"github.com/nishiki/backend-go/domain/services"
)

type BulkImportObjectsRequest struct {
	ContainerID entities.ContainerID
	Objects     []ObjectImportData
	UserID      entities.UserID
	UserToken   string
}

type ObjectImportData struct {
	Name       string
	ObjectType entities.ObjectType
	Properties map[string]interface{}
	Tags       []string
}

type BulkImportObjectsResponse struct {
	Imported int      `json:"imported"`
	Failed   int      `json:"failed"`
	Total    int      `json:"total"`
	Errors   []string `json:"errors,omitempty"`
}

type BulkImportObjectsUseCase struct {
	containerRepo  repositories.ContainerRepository
	collectionRepo repositories.CollectionRepository
	authService    services.AuthService
}

func NewBulkImportObjectsUseCase(containerRepo repositories.ContainerRepository, collectionRepo repositories.CollectionRepository, authService services.AuthService) *BulkImportObjectsUseCase {
	return &BulkImportObjectsUseCase{
		containerRepo:  containerRepo,
		collectionRepo: collectionRepo,
		authService:    authService,
	}
}

func (uc *BulkImportObjectsUseCase) Execute(ctx context.Context, req BulkImportObjectsRequest) (*BulkImportObjectsResponse, error) {
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

	response := &BulkImportObjectsResponse{
		Total: len(req.Objects),
	}

	// Process each object
	for i, objectData := range req.Objects {
		// Create object name value object
		objectName, err := entities.NewObjectName(objectData.Name)
		if err != nil {
			response.Failed++
			response.Errors = append(response.Errors, fmt.Sprintf("Item %d: invalid name: %s", i+1, err.Error()))
			continue
		}

		// Create object
		object, err := entities.NewObject(entities.ObjectProps{
			Name:       objectName,
			ObjectType: objectData.ObjectType,
			Properties: objectData.Properties,
			Tags:       objectData.Tags,
		})
		if err != nil {
			response.Failed++
			response.Errors = append(response.Errors, fmt.Sprintf("Item %d: %s", i+1, err.Error()))
			continue
		}

		// Add to container
		if err := container.AddObject(*object); err != nil {
			response.Failed++
			response.Errors = append(response.Errors, fmt.Sprintf("Item %d: failed to add to container: %s", i+1, err.Error()))
			continue
		}

		response.Imported++
	}

	// Save updated container if any objects were imported
	if response.Imported > 0 {
		if err := uc.containerRepo.Update(ctx, container); err != nil {
			return nil, fmt.Errorf("failed to save container: %w", err)
		}
	}

	return response, nil
}
