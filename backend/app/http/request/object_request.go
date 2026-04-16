package request

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/nishiki/backend/domain/entities"
)

type CreateObjectRequest struct {
	ContainerID string         `json:"container_id,omitempty"`
	Name        string         `json:"name" binding:"required,min=1,max=255"`
	Description string         `json:"description,omitempty"`
	ObjectType  string         `json:"object_type" binding:"required"`
	Location    string         `json:"location,omitempty"`
	Quantity    *float64       `json:"quantity,omitempty"`
	Unit        string         `json:"unit,omitempty"`
	Properties  map[string]any `json:"properties,omitempty"`
	Tags        []string       `json:"tags,omitempty"`
	ExpiresAt   *time.Time     `json:"expires_at,omitempty"`
}

type UpdateObjectRequest struct {
	ContainerID string         `json:"container_id,omitempty"`
	Name        *string        `json:"name,omitempty"`
	Description *string        `json:"description,omitempty"`
	Location    *string        `json:"location,omitempty"`
	Quantity    *float64       `json:"quantity,omitempty"`
	Unit        *string        `json:"unit,omitempty"`
	Properties  map[string]any `json:"properties,omitempty"`
	Tags        []string       `json:"tags,omitempty"`
	ExpiresAt   *time.Time     `json:"expires_at,omitempty"`
}

func (r *CreateObjectRequest) Validate() error {
	if len(r.Name) < 1 || len(r.Name) > 255 {
		return errors.New("name must be between 1 and 255 characters")
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

func (r *UpdateObjectRequest) Validate() error {
	if r.Name != nil && (len(*r.Name) < 1 || len(*r.Name) > 255) {
		return errors.New("name must be between 1 and 255 characters")
	}
	return nil
}

func (r *UpdateObjectRequest) GetContainerID() (*entities.ContainerID, error) {
	if r.ContainerID == "" {
		return nil, nil
	}
	cid, err := entities.ContainerIDFromString(r.ContainerID)
	if err != nil {
		return nil, err
	}
	return &cid, nil
}

func (r *CreateObjectRequest) GetContainerID() (*entities.ContainerID, error) {
	if r.ContainerID == "" {
		return nil, nil
	}
	cid, err := entities.ContainerIDFromString(r.ContainerID)
	if err != nil {
		return nil, err
	}
	return &cid, nil
}

func (r *CreateObjectRequest) GetObjectType() entities.ObjectType {
	return entities.ObjectType(r.ObjectType)
}

func GetObjectIDFromPath(r *http.Request) (entities.ObjectID, error) {
	idStr := r.PathValue("object_id")
	if idStr == "" {
		idStr = r.PathValue("id")
	}
	if idStr == "" {
		return entities.ObjectID{}, errors.New("missing object ID in path")
	}

	objectID, err := entities.ObjectIDFromHex(idStr)
	if err != nil {
		return entities.ObjectID{}, fmt.Errorf("invalid object ID: %w", err)
	}

	return objectID, nil
}
