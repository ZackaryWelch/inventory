package response

import (
	"time"

	"github.com/nishiki/backend-go/domain/entities"
)

type ContainerResponse struct {
	ID                string           `json:"id"`
	CollectionID      string           `json:"collection_id"`
	Name              string           `json:"name"`
	Type              string           `json:"type"`
	ParentContainerID *string          `json:"parent_container_id,omitempty"`
	CategoryID        *string          `json:"category_id,omitempty"`
	Objects           []ObjectResponse `json:"objects"`
	Location          string           `json:"location"`
	Width             *float64         `json:"width,omitempty"`
	Depth             *float64         `json:"depth,omitempty"`
	Rows              *int             `json:"rows,omitempty"`
	Capacity          *float64         `json:"capacity,omitempty"`
	UsedCapacity      *float64         `json:"used_capacity,omitempty"`
	CapacityUtilization *float64       `json:"capacity_utilization,omitempty"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
}

type ContainerListResponse []ContainerResponse

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

	var parentContainerID *string
	if container.ParentContainerID() != nil {
		id := container.ParentContainerID().String()
		parentContainerID = &id
	}

	// Calculate used capacity
	usedCapacity := container.CalculateUsedCapacity()

	return ContainerResponse{
		ID:                  container.ID().String(),
		CollectionID:        container.CollectionID().String(),
		Name:                container.Name().String(),
		Type:                string(container.ContainerType()),
		ParentContainerID:   parentContainerID,
		CategoryID:          categoryID,
		Objects:             objects,
		Location:            container.Location(),
		Width:               container.Width(),
		Depth:               container.Depth(),
		Rows:                container.Rows(),
		Capacity:            container.Capacity(),
		UsedCapacity:        &usedCapacity,
		CapacityUtilization: container.GetCapacityUtilization(),
		CreatedAt:           container.CreatedAt(),
		UpdatedAt:           container.UpdatedAt(),
	}
}

func NewContainerListResponse(containers []*entities.Container) ContainerListResponse {
	containerResponses := make([]ContainerResponse, len(containers))
	for i, container := range containers {
		containerResponses[i] = NewContainerResponse(container)
	}

	return ContainerListResponse(containerResponses)
}
