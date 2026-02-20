package request

import (
	"fmt"
	"net/http"

	"github.com/nishiki/backend-go/domain/entities"
)

type CreateContainerRequest struct {
	CollectionID      string   `json:"collection_id" binding:"required"`
	Name              string   `json:"name" binding:"required,min=1,max=255"`
	Type              string   `json:"type"` // room, bookshelf, shelf, binder, cabinet, general
	ParentContainerID *string  `json:"parent_container_id,omitempty"`
	GroupID           *string  `json:"group_id,omitempty"`
	Location          string   `json:"location,omitempty"`
	Width             *float64 `json:"width,omitempty"`
	Depth             *float64 `json:"depth,omitempty"`
	Rows              *int     `json:"rows,omitempty"`
	Capacity          *float64 `json:"capacity,omitempty"`
}

type UpdateContainerRequest struct {
	Name              string   `json:"name" binding:"required,min=1,max=255"`
	Type              string   `json:"type,omitempty"`
	ParentContainerID *string  `json:"parent_container_id,omitempty"`
	GroupID           *string  `json:"group_id,omitempty"`
	Location          string   `json:"location,omitempty"`
	Width             *float64 `json:"width,omitempty"`
	Depth             *float64 `json:"depth,omitempty"`
	Rows              *int     `json:"rows,omitempty"`
	Capacity          *float64 `json:"capacity,omitempty"`
}

func (r *CreateContainerRequest) Validate() error {
	if len(r.Name) < 1 || len(r.Name) > 255 {
		return fmt.Errorf("name must be between 1 and 255 characters")
	}
	if r.CollectionID == "" {
		return fmt.Errorf("collection_id is required")
	}
	// Validate container type if provided
	if r.Type != "" && !entities.IsValidContainerType(r.Type) {
		return fmt.Errorf("invalid container type: %s", r.Type)
	}
	// Validate dimensions if provided
	if r.Width != nil && *r.Width < 0 {
		return fmt.Errorf("width must be non-negative")
	}
	if r.Depth != nil && *r.Depth < 0 {
		return fmt.Errorf("depth must be non-negative")
	}
	if r.Rows != nil && *r.Rows < 0 {
		return fmt.Errorf("rows must be non-negative")
	}
	if r.Capacity != nil && *r.Capacity < 0 {
		return fmt.Errorf("capacity must be non-negative")
	}
	return nil
}

func (r *UpdateContainerRequest) Validate() error {
	if len(r.Name) < 1 || len(r.Name) > 255 {
		return fmt.Errorf("name must be between 1 and 255 characters")
	}
	// Validate container type if provided
	if r.Type != "" && !entities.IsValidContainerType(r.Type) {
		return fmt.Errorf("invalid container type: %s", r.Type)
	}
	// Validate dimensions if provided
	if r.Width != nil && *r.Width < 0 {
		return fmt.Errorf("width must be non-negative")
	}
	if r.Depth != nil && *r.Depth < 0 {
		return fmt.Errorf("depth must be non-negative")
	}
	if r.Rows != nil && *r.Rows < 0 {
		return fmt.Errorf("rows must be non-negative")
	}
	if r.Capacity != nil && *r.Capacity < 0 {
		return fmt.Errorf("capacity must be non-negative")
	}
	return nil
}

func (r *CreateContainerRequest) GetCollectionID() (entities.CollectionID, error) {
	return entities.CollectionIDFromString(r.CollectionID)
}

func GetContainerIDFromPath(r *http.Request) (entities.ContainerID, error) {
	idStr := r.PathValue("container_id")
	if idStr == "" {
		return entities.ContainerID{}, fmt.Errorf("missing container ID in path")
	}

	containerID, err := entities.ContainerIDFromString(idStr)
	if err != nil {
		return entities.ContainerID{}, fmt.Errorf("invalid container ID: %w", err)
	}

	return containerID, nil
}
