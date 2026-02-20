package openapi

import "time"

// ErrorResponse is returned by all endpoints on error.
type ErrorResponse struct {
	Error string `json:"error"`
}

// EmptyResponse is returned by DELETE endpoints on success.
type EmptyResponse struct{}

// OpenAPIObjectResponse is an OpenAPI-safe version of response.ObjectResponse.
// Properties uses map[string]string instead of map[string]interface{} to avoid
// swagno panicking on interface{} types.
type OpenAPIObjectResponse struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	ObjectType  string            `json:"object_type"`
	Quantity    *float64          `json:"quantity,omitempty"`
	Unit        string            `json:"unit,omitempty"`
	Properties  map[string]string `json:"properties"`
	Tags        []string          `json:"tags"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// OpenAPIObjectListResponse wraps a list of objects.
type OpenAPIObjectListResponse struct {
	Objects []OpenAPIObjectResponse `json:"objects"`
	Total   int                     `json:"total"`
}

// OpenAPICreateObjectResponse wraps a single created object.
type OpenAPICreateObjectResponse struct {
	Object OpenAPIObjectResponse `json:"object"`
}

// OpenAPIUpdateObjectResponse wraps a single updated object.
type OpenAPIUpdateObjectResponse struct {
	Object OpenAPIObjectResponse `json:"object"`
}

// OpenAPICreateObjectRequest is an OpenAPI-safe version of request.CreateObjectRequest.
type OpenAPICreateObjectRequest struct {
	ContainerID string            `json:"container_id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	ObjectType  string            `json:"object_type"`
	Quantity    *float64          `json:"quantity,omitempty"`
	Unit        string            `json:"unit,omitempty"`
	Properties  map[string]string `json:"properties,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
}

// OpenAPIUpdateObjectRequest is an OpenAPI-safe version of request.UpdateObjectRequest.
type OpenAPIUpdateObjectRequest struct {
	ContainerID string            `json:"container_id"`
	Name        *string           `json:"name,omitempty"`
	Description *string           `json:"description,omitempty"`
	Quantity    *float64          `json:"quantity,omitempty"`
	Unit        *string           `json:"unit,omitempty"`
	Properties  map[string]string `json:"properties,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
}

// OpenAPIBulkImportCollectionRequest is an OpenAPI-safe version of request.BulkImportCollectionRequest.
// Data uses []map[string]string instead of []map[string]interface{}.
type OpenAPIBulkImportCollectionRequest struct {
	TargetContainerID *string             `json:"target_container_id,omitempty"`
	DistributionMode  string              `json:"distribution_mode,omitempty"`
	Format            string              `json:"format"`
	Data              []map[string]string `json:"data"`
	DefaultTags       []string            `json:"default_tags,omitempty"`
}
