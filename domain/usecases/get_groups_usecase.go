package usecases

import (
	"context"
	"fmt"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/services"
)

type GetGroupsRequest struct {
	UserID    entities.UserID
	UserToken string
}

type GetGroupsResponse struct {
	Groups []*entities.Group
}

type GetGroupsUseCase struct {
	authService services.AuthService
}

func NewGetGroupsUseCase(authService services.AuthService) *GetGroupsUseCase {
	return &GetGroupsUseCase{
		authService: authService,
	}
}

func (uc *GetGroupsUseCase) Execute(ctx context.Context, req GetGroupsRequest) (*GetGroupsResponse, error) {
	// Get groups for user from Authentik
	groups, err := uc.authService.GetUserGroups(ctx, req.UserToken, req.UserID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get groups for user: %w", err)
	}

	return &GetGroupsResponse{
		Groups: groups,
	}, nil
}
