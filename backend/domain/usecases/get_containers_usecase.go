package usecases

import (
	"context"
	"fmt"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/repositories"
	"github.com/nishiki/backend-go/domain/services"
)

type GetContainersRequest struct {
	GroupID   entities.GroupID
	UserID    entities.UserID
	UserToken string
}

type GetContainersResponse struct {
	Containers []*entities.Container
}

type GetContainersUseCase struct {
	containerRepo repositories.ContainerRepository
	authService   services.AuthService
}

func NewGetContainersUseCase(containerRepo repositories.ContainerRepository, authService services.AuthService) *GetContainersUseCase {
	return &GetContainersUseCase{
		containerRepo: containerRepo,
		authService:   authService,
	}
}

func (uc *GetContainersUseCase) Execute(ctx context.Context, req GetContainersRequest) (*GetContainersResponse, error) {
	// Check if user is a member of the group by fetching user's groups from Authentik
	userGroups, err := uc.authService.GetUserGroups(ctx, req.UserToken, req.UserID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}

	// Check if the requested group is in the user's groups
	isMember := false
	for _, group := range userGroups {
		if group.ID().Equals(req.GroupID) {
			isMember = true
			break
		}
	}

	if !isMember {
		return nil, fmt.Errorf("user is not a member of the group")
	}

	// Get containers for group
	containers, err := uc.containerRepo.GetByGroupID(ctx, req.GroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get containers for group: %w", err)
	}

	return &GetContainersResponse{
		Containers: containers,
	}, nil
}
