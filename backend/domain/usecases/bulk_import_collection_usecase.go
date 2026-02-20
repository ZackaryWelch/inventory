package usecases

import (
	"context"
	"fmt"
	"log"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/repositories"
	"github.com/nishiki/backend-go/domain/services"
)

type BulkImportCollectionRequest struct {
	UserID            entities.UserID
	CollectionID      entities.CollectionID
	TargetContainerID *entities.ContainerID // Optional: specific container to import to
	DistributionMode  string                // "automatic", "manual", "target"
	Data              []map[string]interface{}
	DefaultTags       []string
	UserToken         string
}

type BulkImportCollectionResponse struct {
	Imported         int               `json:"imported"`
	Failed           int               `json:"failed"`
	Total            int               `json:"total"`
	Errors           []string          `json:"errors,omitempty"`
	CapacityWarnings []CapacityWarning `json:"capacity_warnings,omitempty"`
	Assignments      map[string]int    `json:"assignments,omitempty"` // containerID -> count
}

type CapacityWarning struct {
	ContainerID   string  `json:"container_id"`
	ContainerName string  `json:"container_name"`
	UsedCapacity  float64 `json:"used_capacity"`
	TotalCapacity float64 `json:"total_capacity"`
	Utilization   float64 `json:"utilization"`
	Severity      string  `json:"severity"`
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

	// Determine target container(s) based on distribution mode
	var targetContainers []*entities.Container

	switch req.DistributionMode {
	case "target":
		// Import to specific container
		if req.TargetContainerID == nil {
			return nil, fmt.Errorf("target container ID required for target distribution mode")
		}
		container, err := uc.containerRepo.GetByID(ctx, *req.TargetContainerID)
		if err != nil {
			return nil, fmt.Errorf("target container not found: %w", err)
		}
		// Verify container belongs to this collection
		if !container.CollectionID().Equals(req.CollectionID) {
			return nil, fmt.Errorf("target container does not belong to this collection")
		}
		targetContainers = append(targetContainers, container)

	case "automatic":
		// Use distribution helpers for automatic distribution
		distributionPlan, err := DistributeObjects(ctx, uc.containerRepo, req.CollectionID, req.Data, collection.ObjectType())
		if err != nil {
			return nil, fmt.Errorf("failed to create distribution plan: %w", err)
		}

		if distributionPlan.AssignedObjects == 0 {
			return nil, fmt.Errorf("no containers available for automatic distribution")
		}

		// Get containers for assignment
		containerMap := make(map[string]*entities.Container)
		log.Printf("[AutoDist] Building containerMap from %d assignments", len(distributionPlan.Assignments))
		for _, assignment := range distributionPlan.Assignments {
			if _, exists := containerMap[assignment.ContainerID.String()]; !exists {
				container, err := uc.containerRepo.GetByID(ctx, assignment.ContainerID)
				if err != nil {
					return nil, fmt.Errorf("failed to get container %s: %w", assignment.ContainerID.String(), err)
				}
				log.Printf("[AutoDist] Fetched container %s with %d existing objects", container.ID().String(), len(container.Objects()))
				containerMap[assignment.ContainerID.String()] = container
			}
		}
		log.Printf("[AutoDist] ContainerMap built with %d unique containers", len(containerMap))

		// Store distribution plan for later use
		// We'll use it after creating objects
		autoDistData := &automaticDistribution{
			plan:         distributionPlan,
			containerMap: containerMap,
		}

		// Process objects with automatic distribution
		return uc.executeAutomaticDistribution(ctx, req, collection, autoDistData)

	default:
		// Use first available container or create default
		containers := collection.Containers()
		if len(containers) > 0 {
			targetContainers = append(targetContainers, &containers[0])
		} else {
			// Create a default container for bulk import
			containerName, err := entities.NewContainerName("Default Container")
			if err != nil {
				return nil, fmt.Errorf("failed to create container name: %w", err)
			}

			newContainer, err := entities.NewContainer(entities.ContainerProps{
				CollectionID:  req.CollectionID,
				Name:          containerName,
				ContainerType: entities.ContainerTypeGeneral,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create default container: %w", err)
			}

			// Add container to collection
			if err := collection.AddContainer(*newContainer); err != nil {
				return nil, fmt.Errorf("failed to add container to collection: %w", err)
			}

			targetContainers = append(targetContainers, newContainer)
		}
	}

	// Use first target container for simple distribution
	targetContainer := targetContainers[0]

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

	// Save the updated container with objects
	if err := uc.containerRepo.Update(ctx, targetContainer); err != nil {
		return nil, fmt.Errorf("failed to save container with imported objects: %w", err)
	}

	// If a new container was created (default case), also update the collection
	if len(collection.Containers()) > 0 && collection.Containers()[len(collection.Containers())-1].ID().Equals(targetContainer.ID()) {
		if err := uc.collectionRepo.Update(ctx, collection); err != nil {
			return nil, fmt.Errorf("failed to save collection: %w", err)
		}
	}

	total := imported + failed

	// Build assignments map
	assignments := make(map[string]int)
	assignments[targetContainer.ID().String()] = imported

	return &BulkImportCollectionResponse{
		Imported:         imported,
		Failed:           failed,
		Total:            total,
		Errors:           errors,
		CapacityWarnings: []CapacityWarning{}, // TODO: Calculate capacity warnings
		Assignments:      assignments,
	}, nil
}

func (uc *BulkImportCollectionUseCase) executeAutomaticDistribution(ctx context.Context, req BulkImportCollectionRequest, collection *entities.Collection, autoDistData *automaticDistribution) (*BulkImportCollectionResponse, error) {
	plan := autoDistData.plan
	containerMap := autoDistData.containerMap

	imported := 0
	failed := 0
	var errors []string
	assignments := make(map[string]int)

	// Process each assignment from the distribution plan
	for _, assignment := range plan.Assignments {
		// Get the object data for this assignment
		if assignment.ObjectIndex >= len(req.Data) {
			errors = append(errors, fmt.Sprintf("invalid object index: %d", assignment.ObjectIndex))
			failed++
			continue
		}

		item := req.Data[assignment.ObjectIndex]

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

		// Get the target container for this assignment
		container, exists := containerMap[assignment.ContainerID.String()]
		if !exists {
			errors = append(errors, fmt.Sprintf("container %s not found for object '%s'", assignment.ContainerID.String(), name))
			failed++
			continue
		}

		// Add object to container
		if err := container.AddObject(*newObject); err != nil {
			errors = append(errors, fmt.Sprintf("failed to add object '%s' to container: %v", name, err))
			failed++
			continue
		}
		log.Printf("[AutoDist] Added object '%s' to container %s (now has %d objects)", name, container.ID().String(), len(container.Objects()))

		imported++

		// Track assignments
		containerIDStr := assignment.ContainerID.String()
		assignments[containerIDStr]++
	}

	// Update all affected containers
	log.Printf("[AutoDist] Updating %d containers with new objects", len(containerMap))
	for _, container := range containerMap {
		log.Printf("[AutoDist] Updating container %s with %d total objects", container.ID().String(), len(container.Objects()))
		if err := uc.containerRepo.Update(ctx, container); err != nil {
			return nil, fmt.Errorf("failed to update container %s: %w", container.ID().String(), err)
		}
		log.Printf("[AutoDist] Successfully updated container %s", container.ID().String())
	}
	log.Printf("[AutoDist] All containers updated successfully")

	total := imported + failed

	// Convert capacity warnings from distribution plan
	capacityWarnings := make([]CapacityWarning, len(plan.CapacityWarnings))
	for i, warning := range plan.CapacityWarnings {
		capacityWarnings[i] = CapacityWarning{
			ContainerID:   warning.ContainerID,
			ContainerName: warning.ContainerName,
			UsedCapacity:  warning.UsedCapacity,
			TotalCapacity: warning.TotalCapacity,
			Utilization:   warning.Utilization,
			Severity:      warning.Severity,
		}
	}

	return &BulkImportCollectionResponse{
		Imported:         imported,
		Failed:           failed,
		Total:            total,
		Errors:           errors,
		CapacityWarnings: capacityWarnings,
		Assignments:      assignments,
	}, nil
}

type automaticDistribution struct {
	plan         *DistributionPlan
	containerMap map[string]*entities.Container
}
