package request

import (
	"fmt"
	"net/http"

	"github.com/nishiki/backend/domain/entities"
)

type PropertyDefinitionRequest struct {
	Key          string `json:"key"`
	DisplayName  string `json:"display_name"`
	Type         string `json:"type"`
	Required     bool   `json:"required"`
	CurrencyCode string `json:"currency_code,omitempty"`
}

type PropertySchemaRequest struct {
	Definitions []PropertyDefinitionRequest `json:"definitions"`
}

func (r *PropertySchemaRequest) ToEntity() *entities.PropertySchema {
	if r == nil {
		return nil
	}
	defs := make([]entities.PropertyDefinition, len(r.Definitions))
	for i, d := range r.Definitions {
		defs[i] = entities.PropertyDefinition{
			Key:          d.Key,
			DisplayName:  d.DisplayName,
			Type:         entities.PropertyType(d.Type),
			Required:     d.Required,
			CurrencyCode: d.CurrencyCode,
		}
	}
	return &entities.PropertySchema{Definitions: defs}
}

type CreateCollectionRequest struct {
	GroupID        *string                `json:"group_id,omitempty"`
	Name           string                 `json:"name" binding:"required,min=1,max=255"`
	ObjectType     string                 `json:"object_type,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	Location       string                 `json:"location,omitempty"`
	PropertySchema *PropertySchemaRequest `json:"property_schema,omitempty"`
}

type UpdateCollectionRequest struct {
	Name           string                 `json:"name" binding:"required,min=1,max=255"`
	ObjectType     string                 `json:"object_type,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	Location       string                 `json:"location,omitempty"`
	PropertySchema *PropertySchemaRequest `json:"property_schema,omitempty"`
}

// UpdatePropertySchemaRequest is used by the dedicated schema endpoint.
type UpdatePropertySchemaRequest struct {
	PropertySchema PropertySchemaRequest `json:"property_schema"`
}

func (r *CreateCollectionRequest) Validate() error {
	if len(r.Name) < 1 || len(r.Name) > 255 {
		return errors.New("name must be between 1 and 255 characters")
	}
	// ObjectType is optional; defaults to "general" if empty.
	// Custom types (e.g. "electronic_supplies") are allowed.
	return nil
}

func (r *UpdateCollectionRequest) Validate() error {
	if len(r.Name) < 1 || len(r.Name) > 255 {
		return errors.New("name must be between 1 and 255 characters")
	}
	return nil
}

func (r *CreateCollectionRequest) GetGroupID() (*entities.GroupID, error) {
	if r.GroupID == nil || *r.GroupID == "" {
		return nil, nil
	}

	groupID, err := entities.GroupIDFromString(*r.GroupID)
	if err != nil {
		return nil, err
	}
	return &groupID, nil
}

func (r *CreateCollectionRequest) GetObjectType() entities.ObjectType {
	if r.ObjectType == "" {
		return entities.ObjectTypeGeneral
	}
	return entities.ObjectType(r.ObjectType)
}

func GetCollectionIDFromPath(r *http.Request) (entities.CollectionID, error) {
	idStr := r.PathValue("collection_id")
	if idStr == "" {
		return entities.CollectionID{}, errors.New("missing collection ID in path")
	}

	collectionID, err := entities.CollectionIDFromString(idStr)
	if err != nil {
		return entities.CollectionID{}, fmt.Errorf("invalid collection ID: %w", err)
	}

	return collectionID, nil
}
