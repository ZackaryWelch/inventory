package request

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/nishiki/backend-go/domain/entities"
)

type CreateContainerRequest struct {
	CollectionID string `json:"collection_id" binding:"required"`
	Name         string `json:"name" binding:"required,min=1,max=255"`
}

type UpdateContainerRequest struct {
	Name string `json:"name" binding:"required,min=1,max=255"`
}

func (r *CreateContainerRequest) Validate() error {
	if len(r.Name) < 1 || len(r.Name) > 255 {
		return fmt.Errorf("name must be between 1 and 255 characters")
	}
	if r.CollectionID == "" {
		return fmt.Errorf("collection_id is required")
	}
	return nil
}

func (r *UpdateContainerRequest) Validate() error {
	if len(r.Name) < 1 || len(r.Name) > 255 {
		return fmt.Errorf("name must be between 1 and 255 characters")
	}
	return nil
}

func (r *CreateContainerRequest) GetCollectionID() (entities.CollectionID, error) {
	return entities.CollectionIDFromString(r.CollectionID)
}

func GetContainerIDFromPath(c *gin.Context) (entities.ContainerID, error) {
	idStr := c.Param("container_id")
	if idStr == "" {
		return entities.ContainerID{}, fmt.Errorf("missing container ID in path")
	}

	containerID, err := entities.ContainerIDFromString(idStr)
	if err != nil {
		return entities.ContainerID{}, fmt.Errorf("invalid container ID: %w", err)
	}

	return containerID, nil
}
