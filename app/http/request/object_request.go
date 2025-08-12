package request

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/nishiki/backend-go/domain/entities"
)

type CreateObjectRequest struct {
	ContainerID string                 `json:"container_id" binding:"required"`
	Name        string                 `json:"name" binding:"required,min=1,max=255"`
	ObjectType  string                 `json:"object_type" binding:"required"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
}

type UpdateObjectRequest struct {
	Name       *string                `json:"name,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Tags       []string               `json:"tags,omitempty"`
}

func (r *CreateObjectRequest) Validate() error {
	if len(r.Name) < 1 || len(r.Name) > 255 {
		return fmt.Errorf("name must be between 1 and 255 characters")
	}
	if r.ContainerID == "" {
		return fmt.Errorf("container_id is required")
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
		return fmt.Errorf("name must be between 1 and 255 characters")
	}
	return nil
}

func (r *CreateObjectRequest) GetContainerID() (entities.ContainerID, error) {
	return entities.ContainerIDFromString(r.ContainerID)
}

func (r *CreateObjectRequest) GetObjectType() entities.ObjectType {
	return entities.ObjectType(r.ObjectType)
}

func GetObjectIDFromPath(c *gin.Context) (entities.ObjectID, error) {
	idStr := c.Param("object_id")
	if idStr == "" {
		idStr = c.Param("id")
	}
	if idStr == "" {
		return entities.ObjectID{}, fmt.Errorf("missing object ID in path")
	}

	objectID, err := entities.ObjectIDFromHex(idStr)
	if err != nil {
		return entities.ObjectID{}, fmt.Errorf("invalid object ID: %w", err)
	}

	return objectID, nil
}

