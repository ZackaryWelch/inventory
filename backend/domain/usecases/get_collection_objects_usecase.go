package usecases

import (
	"context"
	"fmt"
	"strings"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/repositories"
	"github.com/nishiki/backend-go/domain/services"
)

type GetCollectionObjectsRequest struct {
	CollectionID    entities.CollectionID
	UserID          entities.UserID
	UserToken       string
	Query           string                // name contains (case-insensitive)
	Tags            []string              // all listed tags must be present
	ContainerID     *entities.ContainerID // only objects in this container
	PropertyFilters map[string]string     // property key → substring match (case-insensitive)
}

type GetCollectionObjectsResponse struct {
	Objects []entities.Object
}

type GetCollectionObjectsUseCase struct {
	collectionRepo repositories.CollectionRepository
	containerRepo  repositories.ContainerRepository
	authService    services.AuthService
}

func NewGetCollectionObjectsUseCase(collectionRepo repositories.CollectionRepository, containerRepo repositories.ContainerRepository, authService services.AuthService) *GetCollectionObjectsUseCase {
	return &GetCollectionObjectsUseCase{
		collectionRepo: collectionRepo,
		containerRepo:  containerRepo,
		authService:    authService,
	}
}

func (uc *GetCollectionObjectsUseCase) Execute(ctx context.Context, req GetCollectionObjectsRequest) (*GetCollectionObjectsResponse, error) {
	// Check user access to collection
	userGroups, err := uc.authService.GetUserGroups(ctx, req.UserToken, req.UserID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}

	collection, err := uc.collectionRepo.GetByID(ctx, req.CollectionID)
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

	// Collect objects — optionally restricted to a single container.
	var allObjects []entities.Object
	if req.ContainerID != nil {
		c, err := collection.GetContainer(*req.ContainerID)
		if err != nil {
			return nil, fmt.Errorf("container not found: %w", err)
		}
		allObjects = c.Objects()
	} else {
		allObjects = collection.GetAllObjects()
	}

	// Apply in-memory filters.
	filtered := allObjects[:0:0]
	query := strings.ToLower(req.Query)
	for _, obj := range allObjects {
		if query != "" && !strings.Contains(strings.ToLower(obj.Name().String()), query) {
			continue
		}
		if !hasAllTags(obj, req.Tags) {
			continue
		}
		if !matchesPropertyFilters(obj, req.PropertyFilters) {
			continue
		}
		filtered = append(filtered, obj)
	}

	return &GetCollectionObjectsResponse{
		Objects: filtered,
	}, nil
}

// hasAllTags returns true if obj has every tag in required (empty required → always true).
func hasAllTags(obj entities.Object, required []string) bool {
	for _, t := range required {
		if !obj.HasTag(t) {
			return false
		}
	}
	return true
}

// matchesPropertyFilters returns true if every filter key/value matches a property on obj
// (case-insensitive substring match on the string representation of the value).
func matchesPropertyFilters(obj entities.Object, filters map[string]string) bool {
	if len(filters) == 0 {
		return true
	}
	props := obj.Properties()
	for k, v := range filters {
		propVal, ok := props[k]
		if !ok || propVal == nil {
			return false
		}
		if !strings.Contains(strings.ToLower(fmt.Sprintf("%v", propVal)), strings.ToLower(v)) {
			return false
		}
	}
	return true
}
