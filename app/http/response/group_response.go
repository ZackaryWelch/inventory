package response

import (
	"time"

	"github.com/nishiki/backend-go/domain/entities"
)

type GroupResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GroupListResponse []GroupResponse

func NewGroupResponse(group *entities.Group) GroupResponse {
	return GroupResponse{
		ID:        group.ID().String(),
		Name:      group.Name().String(),
		CreatedAt: group.CreatedAt(),
		UpdatedAt: group.UpdatedAt(),
	}
}

func NewGroupListResponse(groups []*entities.Group) GroupListResponse {
	groupResponses := make([]GroupResponse, len(groups))
	for i, group := range groups {
		groupResponses[i] = NewGroupResponse(group)
	}

	return GroupListResponse(groupResponses)
}
