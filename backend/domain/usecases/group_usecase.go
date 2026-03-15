package usecases

import (
	"context"
	"fmt"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/services"
)

type GroupUseCase struct {
	authService services.AuthService
}

func NewGroupUseCase(authService services.AuthService) *GroupUseCase {
	return &GroupUseCase{authService: authService}
}

// --- Update ---

type UpdateGroupRequest struct {
	GroupID   entities.GroupID
	Name      string
	UserToken string
}

type UpdateGroupResponse struct {
	Group *entities.Group
}

func (uc *GroupUseCase) UpdateGroup(ctx context.Context, req UpdateGroupRequest) (*UpdateGroupResponse, error) {
	groupName, err := entities.NewGroupName(req.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid group name: %w", err)
	}

	group, err := uc.authService.UpdateGroup(ctx, req.UserToken, req.GroupID.String(), groupName.String())
	if err != nil {
		return nil, fmt.Errorf("failed to update group: %w", err)
	}

	return &UpdateGroupResponse{Group: group}, nil
}

// --- Delete ---

type DeleteGroupRequest struct {
	GroupID   entities.GroupID
	UserToken string
}

func (uc *GroupUseCase) DeleteGroup(ctx context.Context, req DeleteGroupRequest) error {
	if err := uc.authService.DeleteGroup(ctx, req.UserToken, req.GroupID.String()); err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}
	return nil
}

// --- Member management ---

type GroupMemberRequest struct {
	GroupID   entities.GroupID
	UserID    string
	UserToken string
}

func (uc *GroupUseCase) AddMember(ctx context.Context, req GroupMemberRequest) error {
	if err := uc.authService.AddUserToGroup(ctx, req.UserToken, req.GroupID.String(), req.UserID); err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}
	return nil
}

func (uc *GroupUseCase) RemoveMember(ctx context.Context, req GroupMemberRequest) error {
	if err := uc.authService.RemoveUserFromGroup(ctx, req.UserToken, req.GroupID.String(), req.UserID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}
	return nil
}
