package response

import (
	"time"

	"github.com/nishiki/backend-go/domain/entities"
)

type PropertyDefinitionResponse struct {
	Key          string `json:"key"`
	DisplayName  string `json:"display_name"`
	Type         string `json:"type"`
	Required     bool   `json:"required"`
	CurrencyCode string `json:"currency_code,omitempty"`
}

type PropertySchemaResponse struct {
	Definitions []PropertyDefinitionResponse `json:"definitions"`
}

type CollectionResponse struct {
	ID             string                  `json:"id"`
	UserID         string                  `json:"user_id"`
	GroupID        *string                 `json:"group_id,omitempty"`
	Name           string                  `json:"name"`
	CategoryID     *string                 `json:"category_id,omitempty"`
	ObjectType     string                  `json:"object_type"`
	Containers     []ContainerResponse     `json:"containers"`
	Tags           []string                `json:"tags"`
	Location       string                  `json:"location"`
	PropertySchema *PropertySchemaResponse `json:"property_schema,omitempty"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
}

type CollectionListResponse []CollectionResponse

type CollectionSummaryResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	ObjectType  string    `json:"object_type"`
	ObjectCount int       `json:"object_count"`
	Tags        []string  `json:"tags"`
	Location    string    `json:"location"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CollectionSummaryListResponse struct {
	Collections []CollectionSummaryResponse `json:"collections"`
	Total       int                         `json:"total"`
}

func NewPropertySchemaResponse(schema *entities.PropertySchema) *PropertySchemaResponse {
	if schema == nil {
		return nil
	}
	defs := make([]PropertyDefinitionResponse, len(schema.Definitions))
	for i, d := range schema.Definitions {
		defs[i] = PropertyDefinitionResponse{
			Key:          d.Key,
			DisplayName:  d.DisplayName,
			Type:         string(d.Type),
			Required:     d.Required,
			CurrencyCode: d.CurrencyCode,
		}
	}
	return &PropertySchemaResponse{Definitions: defs}
}

func NewCollectionResponse(collection *entities.Collection) CollectionResponse {
	containers := make([]ContainerResponse, len(collection.Containers()))
	for i, container := range collection.Containers() {
		containers[i] = NewContainerResponse(&container)
	}

	response := CollectionResponse{
		ID:             collection.ID().String(),
		UserID:         collection.UserID().String(),
		Name:           collection.Name().String(),
		ObjectType:     collection.ObjectType().String(),
		Containers:     containers,
		Tags:           collection.Tags(),
		Location:       collection.Location(),
		PropertySchema: NewPropertySchemaResponse(collection.PropertySchema()),
		CreatedAt:      collection.CreatedAt(),
		UpdatedAt:      collection.UpdatedAt(),
	}

	if collection.GroupID() != nil {
		groupIDStr := collection.GroupID().String()
		response.GroupID = &groupIDStr
	}

	if collection.CategoryID() != nil {
		categoryIDStr := collection.CategoryID().String()
		response.CategoryID = &categoryIDStr
	}

	return response
}

func NewCollectionListResponse(collections []*entities.Collection) CollectionListResponse {
	collectionResponses := make([]CollectionResponse, len(collections))
	for i, collection := range collections {
		collectionResponses[i] = NewCollectionResponse(collection)
	}

	return CollectionListResponse(collectionResponses)
}

func NewCollectionSummaryResponse(collection *entities.Collection) CollectionSummaryResponse {
	return CollectionSummaryResponse{
		ID:          collection.ID().String(),
		Name:        collection.Name().String(),
		ObjectType:  collection.ObjectType().String(),
		ObjectCount: collection.TotalObjectCount(),
		Tags:        collection.Tags(),
		Location:    collection.Location(),
		CreatedAt:   collection.CreatedAt(),
		UpdatedAt:   collection.UpdatedAt(),
	}
}

func NewCollectionSummaryListResponse(collections []*entities.Collection) CollectionSummaryListResponse {
	summaryResponses := make([]CollectionSummaryResponse, len(collections))
	for i, collection := range collections {
		summaryResponses[i] = NewCollectionSummaryResponse(collection)
	}

	return CollectionSummaryListResponse{
		Collections: summaryResponses,
		Total:       len(collections),
	}
}
