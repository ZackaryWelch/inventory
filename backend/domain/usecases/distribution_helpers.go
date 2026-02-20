package usecases

import (
	"context"
	"fmt"
	"sort"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/repositories"
)

// DistributionPlan represents the plan for distributing objects
type DistributionPlan struct {
	Assignments       []ObjectAssignment
	CapacityWarnings  []CapacityWarning
	TotalObjects      int
	AssignedObjects   int
	UnassignedObjects int
}

// ObjectAssignment represents an object assigned to a container
type ObjectAssignment struct {
	ObjectIndex   int // Index in the input objects array
	ContainerID   entities.ContainerID
	ContainerName string
	EstimatedSize float64
}

// ContainerWithCapacity tracks capacity for a container during distribution
type ContainerWithCapacity struct {
	Container     *entities.Container
	UsedCapacity  float64
	TotalCapacity float64
	Children      []*ContainerWithCapacity
}

// DistributeObjects distributes objects across containers in a collection
func DistributeObjects(
	ctx context.Context,
	containerRepo repositories.ContainerRepository,
	collectionID entities.CollectionID,
	objects []map[string]interface{},
	objectType entities.ObjectType,
) (*DistributionPlan, error) {

	// Get all containers in the collection
	containers, err := containerRepo.GetByCollectionID(ctx, collectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get containers: %w", err)
	}

	plan := &DistributionPlan{
		Assignments:      make([]ObjectAssignment, 0),
		CapacityWarnings: make([]CapacityWarning, 0),
		TotalObjects:     len(objects),
	}

	if len(containers) == 0 {
		// No containers available
		plan.UnassignedObjects = len(objects)
		return plan, nil
	}

	// Build container hierarchy
	hierarchy, err := BuildContainerHierarchy(ctx, containers)
	if err != nil {
		return nil, fmt.Errorf("failed to build hierarchy: %w", err)
	}

	// Find all leaf containers (shelves, binders, cabinets)
	leafContainers := FindLeafContainers(hierarchy)

	if len(leafContainers) == 0 {
		// No leaf containers, use all containers as targets
		for _, container := range containers {
			leafContainers = append(leafContainers, &ContainerWithCapacity{
				Container:     container,
				UsedCapacity:  container.CalculateUsedCapacity(),
				TotalCapacity: CalculateContainerCapacity(container),
			})
		}
	}

	// Distribute objects to containers
	assignments := DistributeToContainers(objects, leafContainers, objectType)
	plan.Assignments = assignments
	plan.AssignedObjects = len(assignments)
	plan.UnassignedObjects = len(objects) - len(assignments)

	// Generate capacity warnings
	plan.CapacityWarnings = GenerateCapacityWarnings(leafContainers)

	return plan, nil
}

// EstimateObjectSize estimates the size of an object for capacity planning
func EstimateObjectSize(obj map[string]interface{}, objectType entities.ObjectType) float64 {
	// For books, estimate based on page count
	if objectType == entities.ObjectTypeBook {
		if pages, ok := obj["pages"].(float64); ok {
			return EstimateBookWidth(int(pages))
		}
		if pages, ok := obj["page_count"].(float64); ok {
			return EstimateBookWidth(int(pages))
		}
		// Default book size if no page count
		return 1.0
	}

	// For other object types, use quantity if available
	if qty, ok := obj["quantity"].(float64); ok {
		return qty
	}

	// Default size: 1 unit
	return 1.0
}

// EstimateBookWidth estimates book width in inches based on page count
// Ported from Python generate_organization.py
func EstimateBookWidth(pages int) float64 {
	switch {
	case pages <= 100:
		return 0.5
	case pages <= 200:
		return 0.75
	case pages <= 300:
		return 1.0
	case pages <= 500:
		return 1.25
	case pages <= 750:
		return 1.5
	default:
		return 2.0
	}
}

// CalculateContainerCapacity calculates total capacity based on dimensions
func CalculateContainerCapacity(container *entities.Container) float64 {
	// If capacity is explicitly set, use it
	if container.Capacity() != nil {
		return *container.Capacity()
	}

	// Otherwise, calculate from dimensions
	width := container.Width()
	depth := container.Depth()
	rows := container.Rows()

	if width != nil && depth != nil && rows != nil {
		// For shelves: capacity = width * rows
		// (assuming books are spine-out, so depth doesn't matter for count)
		return *width * float64(*rows)
	}

	// Default capacity if no dimensions
	return 100.0
}

// BuildContainerHierarchy builds a hierarchical tree of containers
func BuildContainerHierarchy(
	ctx context.Context,
	containers []*entities.Container,
) ([]*ContainerWithCapacity, error) {

	// Create lookup maps
	containerMap := make(map[string]*ContainerWithCapacity)
	roots := make([]*ContainerWithCapacity, 0)

	// First pass: create all container wrappers
	for _, container := range containers {
		cwc := &ContainerWithCapacity{
			Container:     container,
			UsedCapacity:  container.CalculateUsedCapacity(),
			TotalCapacity: CalculateContainerCapacity(container),
			Children:      make([]*ContainerWithCapacity, 0),
		}
		containerMap[container.ID().String()] = cwc
	}

	// Second pass: build hierarchy
	for _, container := range containers {
		cwc := containerMap[container.ID().String()]

		if container.ParentContainerID() == nil {
			// Root container (no parent)
			roots = append(roots, cwc)
		} else {
			// Has parent - add to parent's children
			parentID := container.ParentContainerID().String()
			if parent, exists := containerMap[parentID]; exists {
				parent.Children = append(parent.Children, cwc)
			}
		}
	}

	return roots, nil
}

// FindLeafContainers returns all leaf containers (containers that can hold objects directly)
func FindLeafContainers(roots []*ContainerWithCapacity) []*ContainerWithCapacity {
	leaves := make([]*ContainerWithCapacity, 0)

	var traverse func(*ContainerWithCapacity)
	traverse = func(node *ContainerWithCapacity) {
		if len(node.Children) == 0 || node.Container.IsLeafContainer() {
			leaves = append(leaves, node)
		} else {
			for _, child := range node.Children {
				traverse(child)
			}
		}
	}

	for _, root := range roots {
		traverse(root)
	}

	return leaves
}

// DistributeToContainers assigns objects to specific containers based on capacity
func DistributeToContainers(
	objects []map[string]interface{},
	containers []*ContainerWithCapacity,
	objectType entities.ObjectType,
) []ObjectAssignment {

	assignments := make([]ObjectAssignment, 0, len(objects))

	// Sort containers by available capacity (most space first)
	sortedContainers := make([]*ContainerWithCapacity, len(containers))
	copy(sortedContainers, containers)
	sort.Slice(sortedContainers, func(i, j int) bool {
		availI := sortedContainers[i].TotalCapacity - sortedContainers[i].UsedCapacity
		availJ := sortedContainers[j].TotalCapacity - sortedContainers[j].UsedCapacity
		return availI > availJ
	})

	currentContainerIdx := 0

	// Assign each object
	for i, obj := range objects {
		if currentContainerIdx >= len(sortedContainers) {
			// No more containers available
			break
		}

		objSize := EstimateObjectSize(obj, objectType)
		container := sortedContainers[currentContainerIdx]

		// Check if object fits in current container
		availableSpace := container.TotalCapacity - container.UsedCapacity
		if availableSpace < objSize {
			// Move to next container
			currentContainerIdx++
			if currentContainerIdx >= len(sortedContainers) {
				// No more containers
				break
			}
			container = sortedContainers[currentContainerIdx]
		}

		// Assign object to container
		assignments = append(assignments, ObjectAssignment{
			ObjectIndex:   i,
			ContainerID:   container.Container.ID(),
			ContainerName: container.Container.Name().String(),
			EstimatedSize: objSize,
		})

		// Update used capacity
		container.UsedCapacity += objSize
	}

	return assignments
}

// GenerateCapacityWarnings checks for containers that are over or near capacity
func GenerateCapacityWarnings(containers []*ContainerWithCapacity) []CapacityWarning {
	warnings := make([]CapacityWarning, 0)

	for _, container := range containers {
		if container.TotalCapacity == 0 {
			continue
		}

		utilization := (container.UsedCapacity / container.TotalCapacity) * 100

		if utilization >= 80 {
			severity := "warning"
			if utilization > 100 {
				severity = "critical"
			}

			warnings = append(warnings, CapacityWarning{
				ContainerID:   container.Container.ID().String(),
				ContainerName: container.Container.Name().String(),
				UsedCapacity:  container.UsedCapacity,
				TotalCapacity: container.TotalCapacity,
				Utilization:   utilization,
				Severity:      severity,
			})
		}
	}

	return warnings
}
