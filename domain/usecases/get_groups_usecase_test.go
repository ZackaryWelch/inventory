package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/mocks"
)

func TestGetGroupsUseCase_Execute(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockAuthService := mocks.NewMockAuthService(mockCtrl)

	useCase := NewGetGroupsUseCase(mockAuthService)

	t.Run("success - multiple groups", func(t *testing.T) {
		t.Parallel()

		// Create test user
		userID := entities.NewUserID()

		// Create test groups
		groupName1, _ := entities.NewGroupName(fake.Company())
		groupDesc1 := entities.NewGroupDescription("Test Group 1 Description")
		groupName2, _ := entities.NewGroupName(fake.Company())
		groupDesc2 := entities.NewGroupDescription("Test Group 2 Description")

		group1 := entities.ReconstructGroup(
			entities.NewGroupID(),
			groupName1,
			groupDesc1,
			time.Now(),
			time.Now(),
		)

		group2 := entities.ReconstructGroup(
			entities.NewGroupID(),
			groupName2,
			groupDesc2,
			time.Now(),
			time.Now(),
		)

		expectedGroups := []*entities.Group{group1, group2}

		// Setup mock expectations
		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), gomock.Any(), userID.String()).
			Return(expectedGroups, nil)

		// Execute use case
		req := GetGroupsRequest{
			UserID: userID,
		}

		resp, err := useCase.Execute(context.Background(), req)

		// Assert results
		require.NoError(t, err)
		assert.Len(t, resp.Groups, 2)
		assert.Equal(t, group1.ID().String(), resp.Groups[0].ID().String())
		assert.Equal(t, group2.ID().String(), resp.Groups[1].ID().String())
		assert.Equal(t, group1.Name().String(), resp.Groups[0].Name().String())
		assert.Equal(t, group2.Name().String(), resp.Groups[1].Name().String())
	})

	t.Run("success - no groups", func(t *testing.T) {
		t.Parallel()

		userID := entities.NewUserID()

		// Setup mock expectations for empty result
		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), gomock.Any(), userID.String()).
			Return([]*entities.Group{}, nil)

		// Execute use case
		req := GetGroupsRequest{
			UserID: userID,
		}

		resp, err := useCase.Execute(context.Background(), req)

		// Assert results
		require.NoError(t, err)
		assert.Empty(t, resp.Groups)
	})

	t.Run("error - auth service failure", func(t *testing.T) {
		t.Parallel()

		userID := entities.NewUserID()
		authError := errors.New("auth service error")

		// Setup mock expectations for error
		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), gomock.Any(), userID.String()).
			Return(nil, authError)

		// Execute use case
		req := GetGroupsRequest{
			UserID: userID,
		}

		resp, err := useCase.Execute(context.Background(), req)

		// Assert error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get groups for user")
		assert.Nil(t, resp)
	})

	t.Run("error - invalid user ID", func(t *testing.T) {
		t.Parallel()

		// Create request with empty user ID
		req := GetGroupsRequest{
			UserID: entities.UserID{}, // Empty/invalid user ID
		}

		// Setup mock to expect call with empty string and return error
		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), gomock.Any(), "").
			Return(nil, errors.New("invalid user ID"))

		resp, err := useCase.Execute(context.Background(), req)

		// Assert error
		require.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("success - groups with containers and users", func(t *testing.T) {
		t.Parallel()

		userID := entities.NewUserID()

		// Create test group
		groupName, _ := entities.NewGroupName(fake.Company())
		groupDesc := entities.NewGroupDescription("Test Group Description")
		group := entities.ReconstructGroup(
			entities.NewGroupID(),
			groupName,
			groupDesc,
			time.Now(),
			time.Now(),
		)

		expectedGroups := []*entities.Group{group}

		// Setup mock expectations
		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), gomock.Any(), userID.String()).
			Return(expectedGroups, nil)

		// Execute use case
		req := GetGroupsRequest{
			UserID: userID,
		}

		resp, err := useCase.Execute(context.Background(), req)

		// Assert results
		require.NoError(t, err)
		assert.Len(t, resp.Groups, 1)

		returnedGroup := resp.Groups[0]
		assert.Equal(t, group.ID().String(), returnedGroup.ID().String())
		assert.Equal(t, group.Name().String(), returnedGroup.Name().String())
	})
}
