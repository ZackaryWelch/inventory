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

func TestCreateContainerUseCase_Execute(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockContainerRepo := mocks.NewMockContainerRepository(mockCtrl)
	mockAuthService := mocks.NewMockAuthService(mockCtrl)

	useCase := NewCreateContainerUseCase(mockContainerRepo, mockAuthService)

	t.Run("success - create container", func(t *testing.T) {
		// Create test data
		userID := entities.NewUserID()
		groupID := entities.NewGroupID()
		containerName := fake.Word()

		// Create test group that user belongs to
		groupName, _ := entities.NewGroupName(fake.Company())
		userGroup := entities.ReconstructGroup(
			groupID,
			groupName,
			[]entities.ContainerID{},
			[]entities.UserID{userID},
			time.Now(),
			time.Now(),
		)

		// Setup mock expectations
		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), gomock.Any(), userID.String()).
			Return([]*entities.Group{userGroup}, nil).
			Times(1)

		mockContainerRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

		// Execute use case
		req := CreateContainerRequest{
			GroupID: groupID,
			Name:    containerName,
			UserID:  userID,
		}

		resp, err := useCase.Execute(context.Background(), req)

		// Assert results
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Container)
		assert.Equal(t, groupID.String(), resp.Container.GroupID().String())
		assert.Equal(t, containerName, resp.Container.Name().String())
		assert.Empty(t, resp.Container.Foods())
	})

	t.Run("error - user not member of group", func(t *testing.T) {
		userID := entities.NewUserID()
		groupID := entities.NewGroupID()
		containerName := fake.Word()

		// Create different group that user belongs to (not the requested one)
		differentGroupID := entities.NewGroupID()
		groupName, _ := entities.NewGroupName(fake.Company())
		userGroup := entities.ReconstructGroup(
			differentGroupID,
			groupName,
			[]entities.ContainerID{},
			[]entities.UserID{userID},
			time.Now(),
			time.Now(),
		)

		// Setup mock expectations
		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), gomock.Any(), userID.String()).
			Return([]*entities.Group{userGroup}, nil).
			Times(1)

		// Execute use case
		req := CreateContainerRequest{
			GroupID: groupID,
			Name:    containerName,
			UserID:  userID,
		}

		resp, err := useCase.Execute(context.Background(), req)

		// Assert error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user is not a member of the group")
		assert.Nil(t, resp)
	})

	t.Run("error - auth service failure", func(t *testing.T) {
		userID := entities.NewUserID()
		groupID := entities.NewGroupID()
		containerName := fake.Word()

		// Setup mock expectations for auth service error
		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), gomock.Any(), userID.String()).
			Return(nil, errors.New("auth service error")).
			Times(1)

		// Execute use case
		req := CreateContainerRequest{
			GroupID: groupID,
			Name:    containerName,
			UserID:  userID,
		}

		resp, err := useCase.Execute(context.Background(), req)

		// Assert error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get user groups")
		assert.Nil(t, resp)
	})

	t.Run("error - invalid container name", func(t *testing.T) {
		userID := entities.NewUserID()
		groupID := entities.NewGroupID()
		invalidName := "" // Empty name should be invalid

		// Create test group that user belongs to
		groupName, _ := entities.NewGroupName(fake.Company())
		userGroup := entities.ReconstructGroup(
			groupID,
			groupName,
			[]entities.ContainerID{},
			[]entities.UserID{userID},
			time.Now(),
			time.Now(),
		)

		// Setup mock expectations
		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), gomock.Any(), userID.String()).
			Return([]*entities.Group{userGroup}, nil).
			Times(1)

		// Execute use case
		req := CreateContainerRequest{
			GroupID: groupID,
			Name:    invalidName,
			UserID:  userID,
		}

		resp, err := useCase.Execute(context.Background(), req)

		// Assert error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid container name")
		assert.Nil(t, resp)
	})

	t.Run("error - repository save failure", func(t *testing.T) {
		userID := entities.NewUserID()
		groupID := entities.NewGroupID()
		containerName := fake.Word()

		// Create test group that user belongs to
		groupName, _ := entities.NewGroupName(fake.Company())
		userGroup := entities.ReconstructGroup(
			groupID,
			groupName,
			[]entities.ContainerID{},
			[]entities.UserID{userID},
			time.Now(),
			time.Now(),
		)

		// Setup mock expectations
		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), gomock.Any(), userID.String()).
			Return([]*entities.Group{userGroup}, nil).
			Times(1)

		mockContainerRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(errors.New("database error")).
			Times(1)

		// Execute use case
		req := CreateContainerRequest{
			GroupID: groupID,
			Name:    containerName,
			UserID:  userID,
		}

		resp, err := useCase.Execute(context.Background(), req)

		// Assert error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save container")
		assert.Nil(t, resp)
	})
}
