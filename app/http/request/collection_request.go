package request

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/nishiki/backend-go/domain/entities"
)

type CreateCollectionRequest struct {
	GroupID    *string  `json:"group_id,omitempty"`
	Name       string   `json:"name" binding:"required,min=1,max=255"`
	ObjectType string   `json:"object_type" binding:"required"`
	Tags       []string `json:"tags,omitempty"`
	Location   string   `json:"location,omitempty"`
}

type UpdateCollectionRequest struct {
	Name     string   `json:"name" binding:"required,min=1,max=255"`
	Tags     []string `json:"tags,omitempty"`
	Location string   `json:"location,omitempty"`
}

func (r *CreateCollectionRequest) Validate() error {
	if len(r.Name) < 1 || len(r.Name) > 255 {
		return fmt.Errorf("name must be between 1 and 255 characters")
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

func (r *UpdateCollectionRequest) Validate() error {
	if len(r.Name) < 1 || len(r.Name) > 255 {
		return fmt.Errorf("name must be between 1 and 255 characters")
	}
	return nil
}

func (r *CreateCollectionRequest) GetGroupID() (*entities.GroupID, error) {
	if r.GroupID == nil || *r.GroupID == "" {
		return nil, nil
	}

	groupID, err := entities.GroupIDFromString(*r.GroupID)
	if err != nil {
		return nil, err
	}
	return &groupID, nil
}

func (r *CreateCollectionRequest) GetObjectType() entities.ObjectType {
	return entities.ObjectType(r.ObjectType)
}

func GetCollectionIDFromPath(c *gin.Context) (entities.CollectionID, error) {
	idStr := c.Param("collection_id")
	if idStr == "" {
		return entities.CollectionID{}, fmt.Errorf("missing collection ID in path")
	}

	collectionID, err := entities.CollectionIDFromString(idStr)
	if err != nil {
		return entities.CollectionID{}, fmt.Errorf("invalid collection ID: %w", err)
	}

	return collectionID, nil
}
