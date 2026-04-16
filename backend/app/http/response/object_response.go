package response

import (
	"time"

	"github.com/nishiki/backend/domain/entities"
)

// TypedValueResponse mirrors entities.TypedValue for the HTTP/frontend layer.
type TypedValueResponse struct {
	Type     string `json:"type"`
	Val      any    `json:"val"`
	Approx   bool   `json:"approx,omitempty"`
	Currency string `json:"cur,omitempty"`
}

type ObjectResponse struct {
	ID          string                        `json:"id"`
	ContainerID string                        `json:"container_id,omitempty"`
	Name        string                        `json:"name"`
	Description string                        `json:"description"`
	ObjectType  string                        `json:"object_type"`
	Location    string                        `json:"location,omitempty"`
	Quantity    *float64                      `json:"quantity,omitempty"`
	Unit        string                        `json:"unit,omitempty"`
	Properties  map[string]TypedValueResponse `json:"properties"`
	Tags        []string                      `json:"tags"`
	ImageURL    string                        `json:"image_url,omitempty"`
	ExpiresAt   *time.Time                    `json:"expires_at,omitempty"`
	CreatedAt   time.Time                     `json:"created_at"`
	UpdatedAt   time.Time                     `json:"updated_at"`
}

type ObjectListResponse struct {
	Objects []ObjectResponse `json:"objects"`
	Total   int              `json:"total"`
}

func NewObjectResponse(object entities.Object, containerID string) ObjectResponse {
	rawProps := object.Properties()
	props := make(map[string]TypedValueResponse, len(rawProps))
	for k, tv := range rawProps {
		props[k] = TypedValueResponse{
			Type:     string(tv.Type),
			Val:      tv.Val,
			Approx:   tv.Approx,
			Currency: tv.Currency,
		}
	}
	return ObjectResponse{
		ID:          object.ID().String(),
		ContainerID: containerID,
		Name:        object.Name().String(),
		Description: object.Description().String(),
		ObjectType:  object.ObjectType().String(),
		Location:    object.Location(),
		Quantity:    object.Quantity(),
		Unit:        object.Unit(),
		Properties:  props,
		Tags:        object.Tags(),
		ImageURL:    object.ImageURL(),
		ExpiresAt:   object.ExpiresAt(),
		CreatedAt:   object.CreatedAt(),
		UpdatedAt:   object.UpdatedAt(),
	}
}

type CreateObjectResponse struct {
	Object ObjectResponse `json:"object"`
}

type UpdateObjectResponse struct {
	Object ObjectResponse `json:"object"`
}

type DeleteObjectResponse struct {
	Success bool `json:"success"`
}

type BulkImportResponse struct {
	Imported int      `json:"imported"`
	Failed   int      `json:"failed"`
	Total    int      `json:"total"`
	Errors   []string `json:"errors,omitempty"`
}
