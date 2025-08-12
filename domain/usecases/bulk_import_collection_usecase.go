package usecases

import (
	"context"
	"fmt"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/repositories"
	"github.com/nishiki/backend-go/domain/services"
)

type BulkImportCollectionRequest struct {
	UserID       entities.UserID
	CollectionID entities.CollectionID
	Data         []map[string]interface{}
	DefaultTags  []string
	UserToken    string
}

type BulkImportCollectionResponse struct {
	Imported int      `json:"imported"`
	Failed   int      `json:"failed"`
	Total    int      `json:"total"`
	Errors   []string `json:"errors,omitempty"`
}

type BulkImportCollectionUseCase struct {
	collectionRepo repositories.CollectionRepository
	containerRepo  repositories.ContainerRepository
	authService    services.AuthService
}

func NewBulkImportCollectionUseCase(collectionRepo repositories.CollectionRepository, containerRepo repositories.ContainerRepository, authService services.AuthService) *BulkImportCollectionUseCase {
	return &BulkImportCollectionUseCase{
		collectionRepo: collectionRepo,
		containerRepo:  containerRepo,
		authService:    authService,
	}
}

func (uc *BulkImportCollectionUseCase) Execute(ctx context.Context, req BulkImportCollectionRequest) (*BulkImportCollectionResponse, error) {
	// Verify user access to the collection
	userGroups, err := uc.authService.GetUserGroups(ctx, req.UserToken, req.UserID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}

	collection, err := uc.collectionRepo.GetByID(ctx, req.CollectionID)
	if err != nil {
		return nil, fmt.Errorf("collection not found: %w", err)
	}

	// Check access: user is owner OR user is member of collection's group
	hasAccess := false
	if collection.UserID().Equals(req.UserID) {
		hasAccess = true
	} else if collection.GroupID() != nil {
		for _, group := range userGroups {
			if group.Name().String() == collection.GroupID().String() {
				hasAccess = true
				break
			}
		}
	}

	if !hasAccess {
		return nil, fmt.Errorf("access denied")
	}

	// Get or create a default container for the collection
	containers := collection.Containers()

	var targetContainer *entities.Container
	if len(containers) > 0 {
		// Use the first available container
		targetContainer = &containers[0]
	} else {
		// Create a default container for bulk import
		containerName, err := entities.NewContainerName("Default Container")
		if err != nil {
			return nil, fmt.Errorf("failed to create container name: %w", err)
		}

		newContainer, err := entities.NewContainer(entities.ContainerProps{
			CollectionID: req.CollectionID,
			Name:         containerName,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create default container: %w", err)
		}

		// Add container to collection
		if err := collection.AddContainer(*newContainer); err != nil {
			return nil, fmt.Errorf("failed to add container to collection: %w", err)
		}

		targetContainer = newContainer
	}

	// Process the bulk import data
	imported := 0
	failed := 0
	var errors []string

	for _, item := range req.Data {
		// Extract name
		nameValue, ok := item["name"]
		if !ok {
			errors = append(errors, "missing required field: name")
			failed++
			continue
		}

		name, ok := nameValue.(string)
		if !ok || name == "" {
			errors = append(errors, "invalid name: must be a non-empty string")
			failed++
			continue
		}

		// Use the collection's object type
		objectType := collection.ObjectType()

		// Extract properties (all fields except name and tags)
		properties := make(map[string]interface{})
		for key, value := range item {
			if key != "name" && key != "tags" {
				properties[key] = value
			}
		}

		// Combine default tags with any item-specific tags
		tags := append([]string(nil), req.DefaultTags...)
		if itemTags, ok := item["tags"].([]interface{}); ok {
			for _, tag := range itemTags {
				if tagStr, ok := tag.(string); ok {
					tags = append(tags, tagStr)
				}
			}
		}

		// Create the object
		objectName, err := entities.NewObjectName(name)
		if err != nil {
			errors = append(errors, fmt.Sprintf("invalid object name '%s': %v", name, err))
			failed++
			continue
		}

		newObject, err := entities.NewObject(entities.ObjectProps{
			Name:       objectName,
			ObjectType: objectType,
			Properties: properties,
			Tags:       tags,
		})
		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to create object '%s': %v", name, err))
			failed++
			continue
		}

		// Add object to container
		if err := targetContainer.AddObject(*newObject); err != nil {
			errors = append(errors, fmt.Sprintf("failed to add object '%s' to container: %v", name, err))
			failed++
			continue
		}

		imported++
	}

	// Save the updated collection (which contains the container with objects)
	if err := uc.collectionRepo.Update(ctx, collection); err != nil {
		return nil, fmt.Errorf("failed to save collection with imported objects: %w", err)
	}

	total := imported + failed

	return &BulkImportCollectionResponse{
		Imported: imported,
		Failed:   failed,
		Total:    total,
		Errors:   errors,
	}, nil
}
