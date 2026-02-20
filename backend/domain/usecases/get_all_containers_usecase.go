package usecases

import (
	"context"
	"fmt"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/repositories"
	"github.com/nishiki/backend-go/domain/services"
)

type GetAllContainersRequest struct {
	UserID    entities.UserID
	UserToken string
}

type GetAllContainersResponse struct {
	Containers []*entities.Container
}

type GetAllContainersUseCase struct {
	containerRepo repositories.ContainerRepository
	authService   services.AuthService
}

func NewGetAllContainersUseCase(containerRepo repositories.ContainerRepository, authService services.AuthService) *GetAllContainersUseCase {
	return &GetAllContainersUseCase{
		containerRepo: containerRepo,
		authService:   authService,
	}
}

func (uc *GetAllContainersUseCase) Execute(ctx context.Context, req GetAllContainersRequest) (*GetAllContainersResponse, error) {
	// Get all user's groups from Authentik (filtered by nishiki role)
	userGroups, err := uc.authService.GetUserGroups(ctx, req.UserToken, req.UserID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}

	allContainers := make([]*entities.Container, 0)

	// Get containers for each group the user is a member of
	for _, group := range userGroups {
		containers, err := uc.containerRepo.GetByGroupID(ctx, group.ID())
		if err != nil {
			// Log the error but continue with other groups
			continue
		}

		allContainers = append(allContainers, containers...)
	}

	return &GetAllContainersResponse{
		Containers: allContainers,
	}, nil
}
