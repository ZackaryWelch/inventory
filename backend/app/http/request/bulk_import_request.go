package request

import (
	"fmt"

	"github.com/nishiki/backend-go/domain/entities"
)

type BulkImportRequest struct {
	ContainerID string                   `json:"container_id" binding:"required"`
	Format      string                   `json:"format" binding:"required"` // "csv" or "json"
	Data        []map[string]interface{} `json:"data" binding:"required"`
	ObjectType  string                   `json:"object_type" binding:"required"`
	DefaultTags []string                 `json:"default_tags,omitempty"`
}

type BulkImportCollectionRequest struct {
	CollectionID      string                   `json:"collection_id,omitempty"` // Optional - collection_id comes from URL path
	TargetContainerID *string                  `json:"target_container_id,omitempty"`
	DistributionMode  string                   `json:"distribution_mode,omitempty"` // "automatic", "manual", "target", "location"
	Format            string                   `json:"format" binding:"required"`   // "csv" or "json"
	Data              []map[string]interface{} `json:"data" binding:"required"`
	DefaultTags       []string                 `json:"default_tags,omitempty"`
	LocationColumn    string                   `json:"location_column,omitempty"` // column name for container mapping (default: "location")
	NameColumn        string                   `json:"name_column,omitempty"`     // column name override for object name
	InferSchema       bool                     `json:"infer_schema,omitempty"`    // run type inference and save schema
}

func (r *BulkImportRequest) Validate() error {
	if r.ContainerID == "" {
		return fmt.Errorf("container_id is required")
	}

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
		case "automatic", "manual", "target", "location":
			// Valid modes
		default:
			return fmt.Errorf("distribution_mode must be 'automatic', 'manual', 'target', or 'location'")
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
	return entities.ContainerIDFromString(r.ContainerID)
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
