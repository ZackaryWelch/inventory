package types

import "time"

// Object represents an inventory item within a container
type Object struct {
	ID          string                 `json:"id"`
	ContainerID string                 `json:"container_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Tags        []string               `json:"tags"`
	Quantity    float64                `json:"quantity"`
	Unit        string                 `json:"unit"`
	Properties  map[string]interface{} `json:"properties,omitempty"` // Flexible properties for different object types
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"` // For food items
}

// CreateObjectRequest represents the request to create a new object
type CreateObjectRequest struct {
	ContainerID string                 `json:"container_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Tags        []string               `json:"tags"`
	Quantity    float64                `json:"quantity"`
	Unit        string                 `json:"unit"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
}

// UpdateObjectRequest represents the request to update an object
type UpdateObjectRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Tags        []string               `json:"tags"`
	Quantity    float64                `json:"quantity"`
	Unit        string                 `json:"unit"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
}

// BulkImportRequest represents a bulk import of objects
type BulkImportRequest struct {
	Objects []CreateObjectRequest `json:"objects"`
}
