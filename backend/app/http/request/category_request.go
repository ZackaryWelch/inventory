package request

import (
	"fmt"
	"net/http"

	"github.com/nishiki/backend-go/domain/entities"
)

type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Color       string `json:"color,omitempty"`
}

type UpdateCategoryRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	Color       *string `json:"color,omitempty"`
}

func (r *CreateCategoryRequest) Validate() error {
	if len(r.Name) < 1 || len(r.Name) > 255 {
		return fmt.Errorf("name must be between 1 and 255 characters")
	}
	return nil
}

func (r *UpdateCategoryRequest) Validate() error {
	if r.Name != nil && (len(*r.Name) < 1 || len(*r.Name) > 255) {
		return fmt.Errorf("name must be between 1 and 255 characters")
	}
	return nil
}

func GetCategoryIDFromPath(r *http.Request) (entities.CategoryID, error) {
	idStr := r.PathValue("id")
	if idStr == "" {
		return entities.CategoryID{}, fmt.Errorf("missing category ID in path")
	}

	categoryID, err := entities.CategoryIDFromHex(idStr)
	if err != nil {
		return entities.CategoryID{}, fmt.Errorf("invalid category ID: %w", err)
	}

	return categoryID, nil
}
