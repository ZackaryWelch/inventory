package request

import (
	"fmt"

	"github.com/nishiki/backend-go/domain/entities"
)

type BulkImportRequest struct {
	Format      string                   `json:"format" binding:"required"` // "csv" or "json"
	Data        []map[string]interface{} `json:"data" binding:"required"`
	ObjectType  string                   `json:"object_type" binding:"required"`
	DefaultTags []string                 `json:"default_tags,omitempty"`
}

type BulkImportCollectionRequest struct {
	CollectionID      string                   `json:"collection_id,omitempty"` // Optional - collection_id comes from URL path
	TargetContainerID *string                  `json:"target_container_id,omitempty"`
	DistributionMode  string                   `json:"distribution_mode,omitempty"` // "automatic", "manual", "target"
	Format            string                   `json:"format" binding:"required"`   // "csv" or "json"
	Data              []map[string]interface{} `json:"data" binding:"required"`
	DefaultTags       []string                 `json:"default_tags,omitempty"`
}

func (r *BulkImportRequest) Validate() error {
	if r.Format != "csv" && r.Format != "json" {
		return fmt.Errorf("format must be 'csv' or 'json'")
	}

	if len(r.Data) == 0 {
		return fmt.Errorf("data is required and cannot be empty")
	}

	// Validate object type
	objectType := entities.ObjectType(r.ObjectType)
	switch objectType {
	case entities.ObjectTypeFood, entities.ObjectTypeBook, entities.ObjectTypeVideoGame,
		entities.ObjectTypeMusic, entities.ObjectTypeBoardGame, entities.ObjectTypeGeneral:
		// Valid object type
	default:
		return fmt.Errorf("invalid object_type: %s", r.ObjectType)
	}

	return nil
}

func (r *BulkImportCollectionRequest) Validate() error {
	if r.Format != "csv" && r.Format != "json" {
		return fmt.Errorf("format must be 'csv' or 'json'")
	}

	if len(r.Data) == 0 {
		return fmt.Errorf("data is required and cannot be empty")
	}

	// Note: CollectionID comes from URL path, not validated here

	// Validate distribution mode if provided
	if r.DistributionMode != "" {
		switch r.DistributionMode {
		case "automatic", "manual", "target":
			// Valid modes
		default:
			return fmt.Errorf("distribution_mode must be 'automatic', 'manual', or 'target'")
		}
	}

	// If distribution mode is "target", require target_container_id
	if r.DistributionMode == "target" && r.TargetContainerID == nil {
		return fmt.Errorf("target_container_id is required when distribution_mode is 'target'")
	}

	return nil
}

func (r *BulkImportRequest) GetObjectType() entities.ObjectType {
	return entities.ObjectType(r.ObjectType)
}

func (r *BulkImportRequest) GetContainerID() (entities.ContainerID, error) {
	// This will be called after getting the container ID from the path
	return entities.ContainerID{}, fmt.Errorf("container ID should be extracted from path, not request body")
}

func (r *BulkImportCollectionRequest) GetCollectionID() (entities.CollectionID, error) {
	return entities.CollectionIDFromString(r.CollectionID)
}

func (r *BulkImportCollectionRequest) GetTargetContainerID() (*entities.ContainerID, error) {
	if r.TargetContainerID == nil {
		return nil, nil
	}
	containerID, err := entities.ContainerIDFromString(*r.TargetContainerID)
	if err != nil {
		return nil, err
	}
	return &containerID, nil
}
