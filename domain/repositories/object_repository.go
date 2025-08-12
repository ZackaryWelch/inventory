//go:generate mockgen -source=object_repository.go -destination=../../mocks/mock_object_repository.go -package=mocks

package repositories

import (
	"context"

	"github.com/nishiki/backend-go/domain/entities"
)

type ObjectRepository interface {
	Create(ctx context.Context, object *entities.Object) error
	GetByID(ctx context.Context, id entities.ObjectID) (*entities.Object, error)
	Update(ctx context.Context, object *entities.Object) error
	Delete(ctx context.Context, id entities.ObjectID) error
	GetByCollectionID(ctx context.Context, collectionID entities.CollectionID) ([]*entities.Object, error)
	GetByType(ctx context.Context, objectType entities.ObjectType) ([]*entities.Object, error)
	List(ctx context.Context, limit, offset int) ([]*entities.Object, error)
	Exists(ctx context.Context, id entities.ObjectID) (bool, error)
}