package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/nishiki/backend/domain/entities"
	"github.com/nishiki/backend/mocks"
)

func TestGetAllContainersUseCase_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockContainerRepo := mocks.NewMockContainerRepository(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)

	useCase := NewGetAllContainersUseCase(mockContainerRepo, mockAuthService)

	ctx := context.Background()
	userID, _ := entities.UserIDFromString("test-user-123")
	userToken := "test-jwt-token"

	groupID1, _ := entities.GroupIDFromString("group-1")
	groupID2, _ := entities.GroupIDFromString("group-2")
	group1 := NewTestGroup(GrpID(groupID1), GrpName("Test Group 1"))
	group2 := NewTestGroup(GrpID(groupID2), GrpName("Test Group 2"))

	containerID1, _ := entities.ContainerIDFromString("container-1")
	containerID2, _ := entities.ContainerIDFromString("container-2")
	containerID3, _ := entities.ContainerIDFromString("container-3")
	container1 := NewTestContainer(CtrID(containerID1), CtrName("Test Container 1"), CtrGroupID(&groupID1))
	container2 := NewTestContainer(CtrID(containerID2), CtrName("Test Container 2"), CtrGroupID(&groupID1))
	container3 := NewTestContainer(CtrID(containerID3), CtrName("Test Container 3"), CtrGroupID(&groupID2))

	t.Run("Success - Returns containers from all user groups", func(t *testing.T) {
		mockAuthService.EXPECT().GetUserGroups(ctx, userToken, userID.String()).Return([]*entities.Group{group1, group2}, nil)
		mockContainerRepo.EXPECT().GetByGroupID(ctx, groupID1).Return([]*entities.Container{container1, container2}, nil)
		mockContainerRepo.EXPECT().GetByGroupID(ctx, groupID2).Return([]*entities.Container{container3}, nil)

		resp, err := useCase.Execute(ctx, GetAllContainersRequest{UserID: userID, UserToken: userToken})

		require.NoError(t, err)
		assert.Len(t, resp.Containers, 3)

		containerIDs := make(map[string]bool)
		for _, c := range resp.Containers {
			containerIDs[c.ID().String()] = true
		}
		for _, id := range []string{containerID1.String(), containerID2.String(), containerID3.String()} {
			assert.True(t, containerIDs[id], "Expected container ID %s not found", id)
		}
	})

	t.Run("Success - No groups means no containers", func(t *testing.T) {
		mockAuthService.EXPECT().GetUserGroups(ctx, userToken, userID.String()).Return([]*entities.Group{}, nil)

		resp, err := useCase.Execute(ctx, GetAllContainersRequest{UserID: userID, UserToken: userToken})

		require.NoError(t, err)
		assert.Empty(t, resp.Containers)
	})

	t.Run("Success - Group with no containers", func(t *testing.T) {
		mockAuthService.EXPECT().GetUserGroups(ctx, userToken, userID.String()).Return([]*entities.Group{group1}, nil)
		mockContainerRepo.EXPECT().GetByGroupID(ctx, groupID1).Return([]*entities.Container{}, nil)

		resp, err := useCase.Execute(ctx, GetAllContainersRequest{UserID: userID, UserToken: userToken})

		require.NoError(t, err)
		assert.Empty(t, resp.Containers)
	})

	t.Run("Error - Auth service fails to get user groups", func(t *testing.T) {
		mockAuthService.EXPECT().GetUserGroups(ctx, userToken, userID.String()).Return(nil, errors.New("failed to get groups"))

		resp, err := useCase.Execute(ctx, GetAllContainersRequest{UserID: userID, UserToken: userToken})

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to get user groups")
	})

	t.Run("Partial Success - Container repo error for one group", func(t *testing.T) {
		mockAuthService.EXPECT().GetUserGroups(ctx, userToken, userID.String()).Return([]*entities.Group{group1, group2}, nil)
		mockContainerRepo.EXPECT().GetByGroupID(ctx, groupID1).Return([]*entities.Container{container1, container2}, nil)
		mockContainerRepo.EXPECT().GetByGroupID(ctx, groupID2).Return(nil, errors.New("database connection failed"))

		resp, err := useCase.Execute(ctx, GetAllContainersRequest{UserID: userID, UserToken: userToken})

		require.NoError(t, err)
		assert.Len(t, resp.Containers, 2)
		for _, c := range resp.Containers {
			assert.True(t, c.GroupID() != nil && c.GroupID().Equals(groupID1))
		}
	})
}
