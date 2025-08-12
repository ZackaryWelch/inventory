package usecases

import (
	"context"
	"fmt"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/services"
)

type CreateGroupRequest struct {
	Name      string
	CreatorID entities.UserID
	UserToken string
}

type CreateGroupResponse struct {
	Group *entities.Group
}

type CreateGroupUseCase struct {
	authService services.AuthService
}

func NewCreateGroupUseCase(authService services.AuthService) *CreateGroupUseCase {
	return &CreateGroupUseCase{
		authService: authService,
	}
}

func (uc *CreateGroupUseCase) Execute(ctx context.Context, req CreateGroupRequest) (*CreateGroupResponse, error) {
	// Validate group name
	groupName, err := entities.NewGroupName(req.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid group name: %w", err)
	}

	// Create group in Authentik with nishiki role and add creator as member
	group, err := uc.authService.CreateGroup(ctx, req.UserToken, groupName.String(), req.CreatorID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to create group in Authentik: %w", err)
	}

	return &CreateGroupResponse{
		Group: group,
	}, nil
}
