package response

import (
	"time"

	"github.com/nishiki/backend-go/domain/entities"
)

type CategoryResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Color       string    `json:"color"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CategoryListResponse []CategoryResponse

func NewCategoryResponse(category *entities.Category) CategoryResponse {
	return CategoryResponse{
		ID:          category.ID().String(),
		Name:        category.Name().String(),
		Description: category.Description().String(),
		Icon:        category.Icon(),
		Color:       category.Color(),
		CreatedAt:   category.CreatedAt(),
		UpdatedAt:   category.UpdatedAt(),
	}
}

func NewCategoryListResponse(categories []*entities.Category) CategoryListResponse {
	categoryResponses := make([]CategoryResponse, len(categories))
	for i, category := range categories {
		categoryResponses[i] = NewCategoryResponse(category)
	}

	return CategoryListResponse(categoryResponses)
}
