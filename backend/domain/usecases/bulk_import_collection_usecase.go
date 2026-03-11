package usecases

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/repositories"
	"github.com/nishiki/backend-go/domain/services"
)

type BulkImportCollectionRequest struct {
	UserID            entities.UserID
	CollectionID      entities.CollectionID
	TargetContainerID *entities.ContainerID // Optional: specific container to import to
	DistributionMode  string                // "automatic", "manual", "target", "location"
	Data              []map[string]interface{}
	DefaultTags       []string
	UserToken         string
	LocationColumn    string // column name for container mapping (default: "location")
	NameColumn        string // column name override for object name
	InferSchema       bool   // run type inference and save schema to collection
}

type BulkImportCollectionResponse struct {
	Imported          int                      `json:"imported"`
	Failed            int                      `json:"failed"`
	Total             int                      `json:"total"`
	Errors            []string                 `json:"errors,omitempty"`
	CapacityWarnings  []CapacityWarning        `json:"capacity_warnings,omitempty"`
	Assignments       map[string]int           `json:"assignments,omitempty"` // containerID -> count
	ContainersCreated int                      `json:"containers_created,omitempty"`
	InferredSchema    *entities.PropertySchema `json:"inferred_schema,omitempty"`
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
	typeInference  *services.TypeInferenceService
}

func NewBulkImportCollectionUseCase(collectionRepo repositories.CollectionRepository, containerRepo repositories.ContainerRepository, authService services.AuthService) *BulkImportCollectionUseCase {
	return &BulkImportCollectionUseCase{
		collectionRepo: collectionRepo,
		containerRepo:  containerRepo,
		authService:    authService,
		typeInference:  services.NewTypeInferenceService(),
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

	// Run type inference if requested
	var inferredSchema *entities.PropertySchema
	if req.InferSchema && len(req.Data) > 0 {
		// Collect headers from first row
		headers := make([]string, 0, len(req.Data[0]))
		for k := range req.Data[0] {
			headers = append(headers, k)
		}
		inferredSchema = uc.typeInference.InferSchema(headers, req.Data)
		if inferredSchema != nil {
			// Coerce all row values according to inferred schema
			for i, row := range req.Data {
				req.Data[i] = uc.typeInference.CoerceRow(row, inferredSchema)
			}
			// Save schema to collection
			collection.UpdatePropertySchema(inferredSchema)
			if err := uc.collectionRepo.Update(ctx, collection); err != nil {
				return nil, fmt.Errorf("failed to save inferred schema: %w", err)
			}
		}
	}

	// Handle location-based distribution mode before the standard switch
	if req.DistributionMode == "location" {
		return uc.executeLocationDistribution(ctx, req, collection, inferredSchema)
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
		return uc.executeAutomaticDistribution(ctx, req, collection, autoDistData, inferredSchema)

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
		InferredSchema:   inferredSchema,
	}, nil
}

func (uc *BulkImportCollectionUseCase) executeAutomaticDistribution(ctx context.Context, req BulkImportCollectionRequest, collection *entities.Collection, autoDistData *automaticDistribution, inferredSchema *entities.PropertySchema) (*BulkImportCollectionResponse, error) {
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
		InferredSchema:   inferredSchema,
	}, nil
}

type automaticDistribution struct {
	plan         *DistributionPlan
	containerMap map[string]*entities.Container
}

// executeLocationDistribution creates containers from unique Location column values
// and assigns each object to its matching container.
func (uc *BulkImportCollectionUseCase) executeLocationDistribution(
	ctx context.Context,
	req BulkImportCollectionRequest,
	collection *entities.Collection,
	inferredSchema *entities.PropertySchema,
) (*BulkImportCollectionResponse, error) {
	// Determine the location column name (default: "location")
	locationCol := req.LocationColumn
	if locationCol == "" {
		locationCol = "location"
	}

	// Determine the name column (default: auto-detect)
	nameCol := req.NameColumn

	// Collect unique location values from the data
	uniqueLocations := make(map[string]struct{})
	for _, row := range req.Data {
		if v, ok := row[locationCol]; ok {
			if loc := strings.TrimSpace(fmt.Sprintf("%v", v)); loc != "" {
				uniqueLocations[loc] = struct{}{}
			}
		}
	}

	// Fetch existing containers for this collection
	existingContainers, err := uc.containerRepo.GetByCollectionID(ctx, req.CollectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing containers: %w", err)
	}

	// Build a case-insensitive name → container map from existing containers
	locationToContainer := make(map[string]*entities.Container)
	for _, c := range existingContainers {
		locationToContainer[strings.ToLower(c.Name().String())] = c
	}

	// Create containers for new location values
	containersCreated := 0
	for loc := range uniqueLocations {
		lowerLoc := strings.ToLower(loc)
		if _, exists := locationToContainer[lowerLoc]; exists {
			continue
		}
		containerName, err := entities.NewContainerName(loc)
		if err != nil {
			continue // skip invalid names
		}
		newContainer, err := entities.NewContainer(entities.ContainerProps{
			CollectionID:  req.CollectionID,
			Name:          containerName,
			ContainerType: entities.ContainerTypeGeneral,
		})
		if err != nil {
			continue
		}
		if err := uc.containerRepo.Create(ctx, newContainer); err != nil {
			return nil, fmt.Errorf("failed to create container '%s': %w", loc, err)
		}
		if err := collection.AddContainer(*newContainer); err != nil {
			return nil, fmt.Errorf("failed to register container '%s' on collection: %w", loc, err)
		}
		locationToContainer[lowerLoc] = newContainer
		containersCreated++
	}

	// Update collection if new containers were added
	if containersCreated > 0 {
		if err := uc.collectionRepo.Update(ctx, collection); err != nil {
			return nil, fmt.Errorf("failed to update collection with new containers: %w", err)
		}
	}

	// Ensure a default container exists for objects with no location
	const defaultContainerName = "Default"
	defaultKey := strings.ToLower(defaultContainerName)
	if _, exists := locationToContainer[defaultKey]; !exists {
		containerName, _ := entities.NewContainerName(defaultContainerName)
		defaultContainer, err := entities.NewContainer(entities.ContainerProps{
			CollectionID:  req.CollectionID,
			Name:          containerName,
			ContainerType: entities.ContainerTypeGeneral,
		})
		if err == nil {
			if createErr := uc.containerRepo.Create(ctx, defaultContainer); createErr == nil {
				_ = collection.AddContainer(*defaultContainer)
				locationToContainer[defaultKey] = defaultContainer
				containersCreated++
				_ = uc.collectionRepo.Update(ctx, collection)
			}
		}
	}

	// Import objects into their containers
	imported := 0
	failed := 0
	var errors []string
	assignments := make(map[string]int)
	// Track which containers were modified for bulk save
	dirtyContainers := make(map[string]*entities.Container)

	objectType := collection.ObjectType()

	for _, item := range req.Data {
		// Resolve name
		name, ok := resolveNameField(item, nameCol)
		if !ok {
			errors = append(errors, "missing required field: name")
			failed++
			continue
		}

		// Resolve target container from location column
		locValue := ""
		if v, ok := item[locationCol]; ok {
			locValue = strings.TrimSpace(fmt.Sprintf("%v", v))
		}
		containerKey := strings.ToLower(locValue)
		if containerKey == "" {
			containerKey = defaultKey
		}
		container, exists := locationToContainer[containerKey]
		if !exists {
			container = locationToContainer[defaultKey]
		}

		// Build properties (exclude name, tags, location column)
		properties := make(map[string]interface{})
		for key, value := range item {
			lk := strings.ToLower(key)
			if lk == strings.ToLower(nameCol) || lk == "name" || lk == "title" || lk == "item" || lk == "tags" || strings.ToLower(key) == strings.ToLower(locationCol) {
				continue
			}
			properties[key] = value
		}

		// Combine default tags with item-specific tags
		tags := append([]string(nil), req.DefaultTags...)
		if itemTags, ok := item["tags"].([]interface{}); ok {
			for _, tag := range itemTags {
				if tagStr, ok := tag.(string); ok {
					tags = append(tags, tagStr)
				}
			}
		}

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

		if err := container.AddObject(*newObject); err != nil {
			errors = append(errors, fmt.Sprintf("failed to add object '%s' to container: %v", name, err))
			failed++
			continue
		}

		dirtyContainers[container.ID().String()] = container
		assignments[container.ID().String()]++
		imported++
	}

	// Persist all modified containers
	for _, c := range dirtyContainers {
		if err := uc.containerRepo.Update(ctx, c); err != nil {
			return nil, fmt.Errorf("failed to save container %s: %w", c.ID().String(), err)
		}
	}

	total := imported + failed
	return &BulkImportCollectionResponse{
		Imported:          imported,
		Failed:            failed,
		Total:             total,
		Errors:            errors,
		CapacityWarnings:  []CapacityWarning{},
		Assignments:       assignments,
		ContainersCreated: containersCreated,
		InferredSchema:    inferredSchema,
	}, nil
}

// resolveNameField finds the object name from a data row using explicit nameCol or auto-detection.
func resolveNameField(item map[string]interface{}, nameCol string) (string, bool) {
	if nameCol != "" {
		if v, ok := item[nameCol]; ok {
			name := strings.TrimSpace(fmt.Sprintf("%v", v))
			if name != "" {
				return name, true
			}
		}
	}
	// Auto-detect common name columns
	for _, candidate := range []string{"name", "Name", "title", "Title", "item", "Item"} {
		if v, ok := item[candidate]; ok {
			name := strings.TrimSpace(fmt.Sprintf("%v", v))
			if name != "" {
				return name, true
			}
		}
	}
	return "", false
}
