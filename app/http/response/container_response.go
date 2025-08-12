package response

import (
	"time"

	"github.com/nishiki/backend-go/domain/entities"
)

type ContainerResponse struct {
	ID           string           `json:"id"`
	CollectionID string           `json:"collection_id"`
	Name         string           `json:"name"`
	CategoryID   *string          `json:"category_id,omitempty"`
	Objects      []ObjectResponse `json:"objects"`
	Location     string           `json:"location"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

type ContainerListResponse struct {
	Containers []ContainerResponse `json:"containers"`
	Total      int                 `json:"total"`
}

func NewContainerResponse(container *entities.Container) ContainerResponse {
	objects := make([]ObjectResponse, len(container.Objects()))
	for i, object := range container.Objects() {
		objects[i] = NewObjectResponse(object)
	}

	var categoryID *string
	if container.CategoryID() != nil {
		id := container.CategoryID().String()
		categoryID = &id
	}

	return ContainerResponse{
		ID:           container.ID().String(),
		CollectionID: container.CollectionID().String(),
		Name:         container.Name().String(),
		CategoryID:   categoryID,
		Objects:      objects,
		Location:     container.Location(),
		CreatedAt:    container.CreatedAt(),
		UpdatedAt:    container.UpdatedAt(),
	}
}

func NewContainerListResponse(containers []*entities.Container) ContainerListResponse {
	containerResponses := make([]ContainerResponse, len(containers))
	for i, container := range containers {
		containerResponses[i] = NewContainerResponse(container)
	}

	return ContainerListResponse{
		Containers: containerResponses,
		Total:      len(containers),
	}
}
