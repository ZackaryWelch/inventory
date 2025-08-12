//go:generate mockgen -source=category_repository.go -destination=../../mocks/mock_category_repository.go -package=mocks

package repositories

import (
	"context"

	"github.com/nishiki/backend-go/domain/entities"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *entities.Category) error
	GetByID(ctx context.Context, id entities.CategoryID) (*entities.Category, error)
	GetByName(ctx context.Context, name entities.CategoryName) (*entities.Category, error)
	List(ctx context.Context, limit, offset int) ([]*entities.Category, error)
	Update(ctx context.Context, category *entities.Category) error
	Delete(ctx context.Context, id entities.CategoryID) error
	Exists(ctx context.Context, id entities.CategoryID) (bool, error)
}
