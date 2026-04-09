package usecases

import (
	"context"
	"errors"
	"testing"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/nishiki/backend/domain/entities"
	"github.com/nishiki/backend/mocks"
)

func TestGetGroupsUseCase_Execute(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAuthService := mocks.NewMockAuthService(mockCtrl)

	useCase := NewGetGroupsUseCase(mockAuthService)

	t.Run("success - multiple groups", func(t *testing.T) {
		t.Parallel()

		userID := entities.NewUserID()
		group1 := NewTestGroup(GrpName(fake.Company()))
		group2 := NewTestGroup(GrpName(fake.Company()))

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), gomock.Any(), userID.String()).Return([]*entities.Group{group1, group2}, nil)

		resp, err := useCase.Execute(context.Background(), GetGroupsRequest{UserID: userID})

		require.NoError(t, err)
		assert.Len(t, resp.Groups, 2)
		assert.Equal(t, group1.ID().String(), resp.Groups[0].ID().String())
		assert.Equal(t, group2.ID().String(), resp.Groups[1].ID().String())
	})

	t.Run("success - no groups", func(t *testing.T) {
		t.Parallel()

		userID := entities.NewUserID()

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), gomock.Any(), userID.String()).Return([]*entities.Group{}, nil)

		resp, err := useCase.Execute(context.Background(), GetGroupsRequest{UserID: userID})

		require.NoError(t, err)
		assert.Empty(t, resp.Groups)
	})

	t.Run("error - auth service failure", func(t *testing.T) {
		t.Parallel()

		userID := entities.NewUserID()

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), gomock.Any(), userID.String()).Return(nil, errors.New("auth service error"))

		resp, err := useCase.Execute(context.Background(), GetGroupsRequest{UserID: userID})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get groups for user")
		assert.Nil(t, resp)
	})

	t.Run("error - invalid user ID", func(t *testing.T) {
		t.Parallel()

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), gomock.Any(), "").Return(nil, errors.New("invalid user ID"))

		resp, err := useCase.Execute(context.Background(), GetGroupsRequest{UserID: entities.UserID{}})

		require.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("success - groups with containers and users", func(t *testing.T) {
		t.Parallel()

		userID := entities.NewUserID()
		group := NewTestGroup(GrpName(fake.Company()))

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), gomock.Any(), userID.String()).Return([]*entities.Group{group}, nil)

		resp, err := useCase.Execute(context.Background(), GetGroupsRequest{UserID: userID})

		require.NoError(t, err)
		assert.Len(t, resp.Groups, 1)
		assert.Equal(t, group.ID().String(), resp.Groups[0].ID().String())
		assert.Equal(t, group.Name().String(), resp.Groups[0].Name().String())
	})
}
