package usecases

import (
	"context"
	"fmt"
	"strings"

	"github.com/nishiki/backend/domain/entities"
	"github.com/nishiki/backend/domain/repositories"
	"github.com/nishiki/backend/domain/services"
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

type ObjectWithContainerID struct {
	Object      entities.Object
	ContainerID entities.ContainerID
}

type GetCollectionObjectsResponse struct {
	Objects []ObjectWithContainerID
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
	// Resolve user groups (cached) for access check
	userGroups, err := uc.authService.GetUserGroups(ctx, req.UserToken, req.UserID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}

	groupIDs := make([]entities.GroupID, len(userGroups))
	for i, g := range userGroups {
		groupIDs[i] = g.ID()
	}

	// Single aggregation: fetch containers + validate access via $lookup on collection
	var containers []*entities.Container
	if req.ContainerID != nil {
		c, err := uc.containerRepo.GetByID(ctx, *req.ContainerID)
		if err != nil {
			return nil, fmt.Errorf("container not found: %w", err)
		}
		containers = []*entities.Container{c}
	} else {
		containers, err = uc.containerRepo.GetByCollectionIDWithAccess(ctx, req.CollectionID, req.UserID, groupIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get containers: %w", err)
		}
	}

	// Collect objects from containers
	var allObjects []ObjectWithContainerID
	for _, c := range containers {
		for _, obj := range c.Objects() {
			allObjects = append(allObjects, ObjectWithContainerID{Object: obj, ContainerID: c.ID()})
		}
	}

	// Apply in-memory filters.
	var filtered []ObjectWithContainerID
	query := strings.ToLower(req.Query)
	for _, item := range allObjects {
		if query != "" && !strings.Contains(strings.ToLower(item.Object.Name().String()), query) {
			continue
		}
		if !hasAllTags(item.Object, req.Tags) {
			continue
		}
		if !matchesPropertyFilters(item.Object, req.PropertyFilters) {
			continue
		}
		filtered = append(filtered, item)
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
		tv, ok := props[k]
		if !ok || tv.Val == nil {
			return false
		}
		if !strings.Contains(strings.ToLower(fmt.Sprintf("%v", tv.Val)), strings.ToLower(v)) {
			return false
		}
	}
	return true
}
