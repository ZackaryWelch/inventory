package request

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/nishiki/backend/domain/entities"
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
		return errors.New("name must be between 1 and 255 characters")
	}
	return nil
}

func (r *UpdateGroupRequest) Validate() error {
	if len(r.Name) < 1 || len(r.Name) > 255 {
		return errors.New("name must be between 1 and 255 characters")
	}
	return nil
}

func (r *JoinGroupRequest) Validate() error {
	if len(r.InvitationHash) == 0 {
		return errors.New("invitation hash is required")
	}
	return nil
}

// GetGroupMemberIDFromPath reads the {user_id} segment from group member routes
// (e.g. /groups/{id}/users/{user_id}).
func GetGroupMemberIDFromPath(r *http.Request) (string, error) {
	id := r.PathValue("user_id")
	if id == "" {
		return "", errors.New("missing user_id in path")
	}
	return id, nil
}

func GetGroupIDFromPath(r *http.Request) (entities.GroupID, error) {
	idStr := r.PathValue("id")
	if idStr == "" {
		return entities.GroupID{}, errors.New("missing group ID in path")
	}

	groupID, err := entities.GroupIDFromString(idStr)
	if err != nil {
		return entities.GroupID{}, fmt.Errorf("invalid group ID: %w", err)
	}

	return groupID, nil
}
