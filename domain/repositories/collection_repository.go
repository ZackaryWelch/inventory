//go:generate mockgen -source=collection_repository.go -destination=../../mocks/mock_collection_repository.go -package=mocks

package repositories

import (
	"context"

	"github.com/nishiki/backend-go/domain/entities"
)

type CollectionRepository interface {
	Create(ctx context.Context, collection *entities.Collection) error
	GetByID(ctx context.Context, id entities.CollectionID) (*entities.Collection, error)
	Update(ctx context.Context, collection *entities.Collection) error
	Delete(ctx context.Context, id entities.CollectionID) error
	GetByUserID(ctx context.Context, userID entities.UserID) ([]*entities.Collection, error)
	GetByGroupID(ctx context.Context, groupID entities.GroupID) ([]*entities.Collection, error)
	List(ctx context.Context, limit, offset int) ([]*entities.Collection, error)
	Exists(ctx context.Context, id entities.CollectionID) (bool, error)
}