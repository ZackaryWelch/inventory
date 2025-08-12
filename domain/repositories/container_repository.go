//go:generate mockgen -source=container_repository.go -destination=../../mocks/mock_container_repository.go -package=mocks

package repositories

import (
	"context"

	"github.com/nishiki/backend-go/domain/entities"
)

type ContainerRepository interface {
	Create(ctx context.Context, container *entities.Container) error
	GetByID(ctx context.Context, id entities.ContainerID) (*entities.Container, error)
	Update(ctx context.Context, container *entities.Container) error
	Delete(ctx context.Context, id entities.ContainerID) error
	GetByGroupID(ctx context.Context, groupID entities.GroupID) ([]*entities.Container, error)
	List(ctx context.Context, limit, offset int) ([]*entities.Container, error)
	Exists(ctx context.Context, id entities.ContainerID) (bool, error)
	GetContainersWithExpiredFood(ctx context.Context, groupID entities.GroupID) ([]*entities.Container, error)
}
