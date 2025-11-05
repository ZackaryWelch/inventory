package response

import (
	"time"

	"github.com/nishiki/backend-go/domain/entities"
)

type ObjectResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	ObjectType  string                 `json:"object_type"`
	Quantity    *float64               `json:"quantity,omitempty"`
	Unit        string                 `json:"unit,omitempty"`
	Properties  map[string]interface{} `json:"properties"`
	Tags        []string               `json:"tags"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type ObjectListResponse struct {
	Objects []ObjectResponse `json:"objects"`
	Total   int              `json:"total"`
}

func NewObjectResponse(object entities.Object) ObjectResponse {
	return ObjectResponse{
		ID:          object.ID().String(),
		Name:        object.Name().String(),
		Description: object.Description().String(),
		ObjectType:  object.ObjectType().String(),
		Quantity:    object.Quantity(),
		Unit:        object.Unit(),
		Properties:  object.Properties(),
		Tags:        object.Tags(),
		ExpiresAt:   object.ExpiresAt(),
		CreatedAt:   object.CreatedAt(),
		UpdatedAt:   object.UpdatedAt(),
	}
}

func NewObjectListResponse(objects []entities.Object) ObjectListResponse {
	objectResponses := make([]ObjectResponse, len(objects))
	for i, object := range objects {
		objectResponses[i] = NewObjectResponse(object)
	}

	return ObjectListResponse{
		Objects: objectResponses,
		Total:   len(objects),
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
