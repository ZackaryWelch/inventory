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

func TestDeleteObjectUseCase_Execute(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockContainerRepo := mocks.NewMockContainerRepository(mockCtrl)
	mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
	mockAuthService := mocks.NewMockAuthService(mockCtrl)

	useCase := NewDeleteObjectUseCase(mockContainerRepo, mockCollectionRepo, mockAuthService)

	t.Run("success - delete object from container", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()
		containerID := entities.NewContainerID()
		objectID := entities.NewObjectID()

		obj := NewTestObject(ObjID(objectID))
		container := NewTestContainer(CtrID(containerID), CtrCollectionID(collectionID), CtrObjects(*obj))
		collection := NewTestCollection(ColID(collectionID), ColUserID(userID))

		req := DeleteObjectRequest{
			ContainerID: &containerID,
			ObjectID:    objectID,
			UserID:      userID,
			UserToken:   "test-token",
		}

		mockContainerRepo.EXPECT().GetByID(gomock.Any(), containerID).Return(container, nil)
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", userID.String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(collection, nil)
		mockContainerRepo.EXPECT().RemoveObject(gomock.Any(), containerID, objectID).Return(nil)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)
	})

	t.Run("success - delete object from group collection", func(t *testing.T) {
		userID := entities.NewUserID()
		groupID := entities.NewGroupID()
		collectionID := entities.NewCollectionID()
		containerID := entities.NewContainerID()
		objectID := entities.NewObjectID()
		ownerID := entities.NewUserID()

		testGroup := NewTestGroup(GrpID(groupID))
		obj := NewTestObject(ObjID(objectID))
		container := NewTestContainer(CtrID(containerID), CtrCollectionID(collectionID), CtrObjects(*obj))
		collection := NewTestCollection(ColID(collectionID), ColUserID(ownerID), ColGroupID(&groupID))

		req := DeleteObjectRequest{
			ContainerID: &containerID,
			ObjectID:    objectID,
			UserID:      userID,
			UserToken:   "test-token",
		}

		mockContainerRepo.EXPECT().GetByID(gomock.Any(), containerID).Return(container, nil)
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", userID.String()).Return([]*entities.Group{testGroup}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(collection, nil)
		mockContainerRepo.EXPECT().RemoveObject(gomock.Any(), containerID, objectID).Return(nil)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)
	})

	t.Run("error - container not found", func(t *testing.T) {
		userID := entities.NewUserID()
		containerID := entities.NewContainerID()
		objectID := entities.NewObjectID()

		req := DeleteObjectRequest{
			ContainerID: &containerID,
			ObjectID:    objectID,
			UserID:      userID,
			UserToken:   "test-token",
		}

		mockContainerRepo.EXPECT().GetByID(gomock.Any(), containerID).Return(nil, errors.New("container not found"))

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "container not found")
	})

	t.Run("error - collection not found", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()
		containerID := entities.NewContainerID()
		objectID := entities.NewObjectID()

		container := NewTestContainer(CtrID(containerID), CtrCollectionID(collectionID))

		req := DeleteObjectRequest{
			ContainerID: &containerID,
			ObjectID:    objectID,
			UserID:      userID,
			UserToken:   "test-token",
		}

		mockContainerRepo.EXPECT().GetByID(gomock.Any(), containerID).Return(container, nil)
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", userID.String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(nil, errors.New("collection not found"))

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "collection not found")
	})

	t.Run("error - access denied", func(t *testing.T) {
		userID := entities.NewUserID()
		ownerID := entities.NewUserID()
		collectionID := entities.NewCollectionID()
		containerID := entities.NewContainerID()
		objectID := entities.NewObjectID()

		obj := NewTestObject(ObjID(objectID))
		container := NewTestContainer(CtrID(containerID), CtrCollectionID(collectionID), CtrObjects(*obj))
		collection := NewTestCollection(ColID(collectionID), ColUserID(ownerID))

		req := DeleteObjectRequest{
			ContainerID: &containerID,
			ObjectID:    objectID,
			UserID:      userID,
			UserToken:   "test-token",
		}

		mockContainerRepo.EXPECT().GetByID(gomock.Any(), containerID).Return(container, nil)
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", userID.String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(collection, nil)

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "access denied")
	})

	t.Run("success - remove non-existent object is idempotent", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()
		containerID := entities.NewContainerID()
		objectID := entities.NewObjectID()

		container := NewTestContainer(CtrID(containerID), CtrCollectionID(collectionID))
		collection := NewTestCollection(ColID(collectionID), ColUserID(userID))

		req := DeleteObjectRequest{
			ContainerID: &containerID,
			ObjectID:    objectID,
			UserID:      userID,
			UserToken:   "test-token",
		}

		mockContainerRepo.EXPECT().GetByID(gomock.Any(), containerID).Return(container, nil)
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", userID.String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(collection, nil)
		mockContainerRepo.EXPECT().RemoveObject(gomock.Any(), containerID, objectID).Return(nil)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)
	})

	t.Run("error - auth service failure", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()
		containerID := entities.NewContainerID()
		objectID := entities.NewObjectID()

		container := NewTestContainer(CtrID(containerID), CtrCollectionID(collectionID))

		req := DeleteObjectRequest{
			ContainerID: &containerID,
			ObjectID:    objectID,
			UserID:      userID,
			UserToken:   "test-token",
		}

		mockContainerRepo.EXPECT().GetByID(gomock.Any(), containerID).Return(container, nil)
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", userID.String()).Return(nil, errors.New("auth service unavailable"))

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to get user groups")
	})

	t.Run("error - repository update failure", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()
		containerID := entities.NewContainerID()
		objectID := entities.NewObjectID()

		obj := NewTestObject(ObjID(objectID))
		container := NewTestContainer(CtrID(containerID), CtrCollectionID(collectionID), CtrObjects(*obj))
		collection := NewTestCollection(ColID(collectionID), ColUserID(userID))

		req := DeleteObjectRequest{
			ContainerID: &containerID,
			ObjectID:    objectID,
			UserID:      userID,
			UserToken:   "test-token",
		}

		mockContainerRepo.EXPECT().GetByID(gomock.Any(), containerID).Return(container, nil)
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", userID.String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(collection, nil)
		mockContainerRepo.EXPECT().RemoveObject(gomock.Any(), containerID, objectID).Return(errors.New("database connection failed"))

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to remove object from container")
	})
}
