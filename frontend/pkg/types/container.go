package types

import "time"

// ContainerType represents the type of physical container
type ContainerType string

const (
	ContainerTypeRoom      ContainerType = "room"
	ContainerTypeBookshelf ContainerType = "bookshelf"
	ContainerTypeShelf     ContainerType = "shelf"
	ContainerTypeBinder    ContainerType = "binder"
	ContainerTypeCabinet   ContainerType = "cabinet"
	ContainerTypeGeneral   ContainerType = "general"
)

// Container represents a storage container within a collection
type Container struct {
	ID                  string        `json:"id"`
	CollectionID        string        `json:"collection_id"`
	Name                string        `json:"name"`
	Type                string        `json:"type"`
	ParentContainerID   *string       `json:"parent_container_id,omitempty"`
	CategoryID          *string       `json:"category_id,omitempty"`
	Objects             []Object      `json:"objects"`
	Location            string        `json:"location"`
	Width               *float64      `json:"width,omitempty"`
	Depth               *float64      `json:"depth,omitempty"`
	Rows                *int          `json:"rows,omitempty"`
	Capacity            *float64      `json:"capacity,omitempty"`
	UsedCapacity        *float64      `json:"used_capacity,omitempty"`
	CapacityUtilization *float64      `json:"capacity_utilization,omitempty"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`
}

// IsLeafContainer returns true if this container type cannot have children
func (c *Container) IsLeafContainer() bool {
	return c.Type == string(ContainerTypeShelf) ||
		c.Type == string(ContainerTypeBinder) ||
		c.Type == string(ContainerTypeCabinet)
}

// CanHaveChildren returns true if this container type can have child containers
func (c *Container) CanHaveChildren() bool {
	return c.Type == string(ContainerTypeRoom) ||
		c.Type == string(ContainerTypeBookshelf)
}

// CreateContainerRequest represents the request to create a new container
type CreateContainerRequest struct {
	CollectionID      string   `json:"collection_id" binding:"required"`
	Name              string   `json:"name" binding:"required"`
	Type              string   `json:"type,omitempty"`
	ParentContainerID *string  `json:"parent_container_id,omitempty"`
	Location          string   `json:"location,omitempty"`
	Width             *float64 `json:"width,omitempty"`
	Depth             *float64 `json:"depth,omitempty"`
	Rows              *int     `json:"rows,omitempty"`
	Capacity          *float64 `json:"capacity,omitempty"`
}

// UpdateContainerRequest represents the request to update a container
type UpdateContainerRequest struct {
	Name              string   `json:"name" binding:"required"`
	Type              string   `json:"type,omitempty"`
	ParentContainerID *string  `json:"parent_container_id,omitempty"`
	Location          string   `json:"location,omitempty"`
	Width             *float64 `json:"width,omitempty"`
	Depth             *float64 `json:"depth,omitempty"`
	Rows              *int     `json:"rows,omitempty"`
	Capacity          *float64 `json:"capacity,omitempty"`
}
