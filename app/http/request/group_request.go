package request

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/nishiki/backend-go/domain/entities"
)

type CreateGroupRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Description string `json:"description,omitempty"`
}

type UpdateGroupRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Description string `json:"description,omitempty"`
}

type JoinGroupRequest struct {
	InvitationHash string `json:"invitationHash" binding:"required"`
}

func (r *CreateGroupRequest) Validate() error {
	if len(r.Name) < 1 || len(r.Name) > 255 {
		return fmt.Errorf("name must be between 1 and 255 characters")
	}
	return nil
}

func (r *UpdateGroupRequest) Validate() error {
	if len(r.Name) < 1 || len(r.Name) > 255 {
		return fmt.Errorf("name must be between 1 and 255 characters")
	}
	return nil
}

func (r *JoinGroupRequest) Validate() error {
	if len(r.InvitationHash) == 0 {
		return fmt.Errorf("invitation hash is required")
	}
	return nil
}

func GetGroupIDFromPath(c *gin.Context) (entities.GroupID, error) {
	idStr := c.Param("id")
	if idStr == "" {
		return entities.GroupID{}, fmt.Errorf("missing group ID in path")
	}

	groupID, err := entities.GroupIDFromString(idStr)
	if err != nil {
		return entities.GroupID{}, fmt.Errorf("invalid group ID: %w", err)
	}

	return groupID, nil
}
