package types

import "time"

// Container represents a storage container within a collection
type Container struct {
	ID           string    `json:"id"`
	CollectionID string    `json:"collection_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Space        *Space    `json:"space,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Objects      []Object  `json:"objects,omitempty"`
}

// Space represents storage space information for a container
type Space struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Depth  float64 `json:"depth"`
	Unit   string  `json:"unit"` // cm, inch, etc.
}

// CreateContainerRequest represents the request to create a new container
type CreateContainerRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Space       *Space  `json:"space,omitempty"`
}

// UpdateContainerRequest represents the request to update a container
type UpdateContainerRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Space       *Space  `json:"space,omitempty"`
}
