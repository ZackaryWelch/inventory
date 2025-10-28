package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/mocks"
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

		// Create test object
		objectName, _ := entities.NewObjectName("Test Object")
		testObject := entities.ReconstructObject(
			objectID,
			objectName,
			entities.ObjectTypeGeneral,
			map[string]interface{}{},
			[]string{},
			time.Now(),
		)

		// Create test container with object
		containerName, _ := entities.NewContainerName("Test Container")
		container := entities.ReconstructContainer(
			containerID,
			collectionID,
			containerName,
			nil,
			[]entities.Object{*testObject},
			"",
			time.Now(),
			time.Now(),
		)

		// Create test collection
		collectionName, _ := entities.NewCollectionName("Test Collection")
		collection := entities.ReconstructCollection(
			collectionID,
			userID,
			nil,
			collectionName,
			nil,
			entities.ObjectTypeGeneral,
			[]entities.Container{},
			[]string{},
			"",
			time.Now(),
			time.Now(),
		)

		req := DeleteObjectRequest{
			ContainerID: containerID,
			ObjectID:    objectID,
			UserID:      userID,
			UserToken:   "test-token",
		}

		// Mock expectations
		mockContainerRepo.EXPECT().
			GetByID(gomock.Any(), containerID).
			Return(container, nil).
			Times(1)

		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), "test-token", userID.String()).
			Return([]*entities.Group{}, nil).
			Times(1)

		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(collection, nil).
			Times(1)

		mockContainerRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, c *entities.Container) error {
				// Verify object was removed
				_, err := c.GetObject(objectID)
				assert.Error(t, err) // Object should not be found
				return nil
			}).
			Times(1)

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
		ownerID := entities.NewUserID() // Different owner

		// Create test group
		groupName, _ := entities.NewGroupName("Test Group")
		testGroup := entities.ReconstructGroup(
			groupID,
			groupName,
			time.Now(),
			time.Now(),
		)

		// Create test object
		objectName, _ := entities.NewObjectName("Test Object")
		testObject := entities.ReconstructObject(
			objectID,
			objectName,
			entities.ObjectTypeGeneral,
			map[string]interface{}{},
			[]string{},
			time.Now(),
		)

		// Create test container with object
		containerName, _ := entities.NewContainerName("Test Container")
		container := entities.ReconstructContainer(
			containerID,
			collectionID,
			containerName,
			nil,
			[]entities.Object{*testObject},
			"",
			time.Now(),
			time.Now(),
		)

		// Create collection in group
		collectionName, _ := entities.NewCollectionName("Group Collection")
		collection := entities.ReconstructCollection(
			collectionID,
			ownerID,
			&groupID,
			collectionName,
			nil,
			entities.ObjectTypeGeneral,
			[]entities.Container{},
			[]string{},
			"",
			time.Now(),
			time.Now(),
		)

		req := DeleteObjectRequest{
			ContainerID: containerID,
			ObjectID:    objectID,
			UserID:      userID,
			UserToken:   "test-token",
		}

		// Mock expectations
		mockContainerRepo.EXPECT().
			GetByID(gomock.Any(), containerID).
			Return(container, nil).
			Times(1)

		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), "test-token", userID.String()).
			Return([]*entities.Group{testGroup}, nil).
			Times(1)

		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(collection, nil).
			Times(1)

		mockContainerRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

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
			ContainerID: containerID,
			ObjectID:    objectID,
			UserID:      userID,
			UserToken:   "test-token",
		}

		// Mock repository returns error
		mockContainerRepo.EXPECT().
			GetByID(gomock.Any(), containerID).
			Return(nil, errors.New("container not found")).
			Times(1)

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

		// Create test container
		containerName, _ := entities.NewContainerName("Test Container")
		container := entities.ReconstructContainer(
			containerID,
			collectionID,
			containerName,
			nil,
			[]entities.Object{},
			"",
			time.Now(),
			time.Now(),
		)

		req := DeleteObjectRequest{
			ContainerID: containerID,
			ObjectID:    objectID,
			UserID:      userID,
			UserToken:   "test-token",
		}

		// Mock expectations
		mockContainerRepo.EXPECT().
			GetByID(gomock.Any(), containerID).
			Return(container, nil).
			Times(1)

		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), "test-token", userID.String()).
			Return([]*entities.Group{}, nil).
			Times(1)

		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(nil, errors.New("collection not found")).
			Times(1)

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "collection not found")
	})

	t.Run("error - access denied", func(t *testing.T) {
		userID := entities.NewUserID()
		ownerID := entities.NewUserID() // Different owner
		collectionID := entities.NewCollectionID()
		containerID := entities.NewContainerID()
		objectID := entities.NewObjectID()

		// Create test object
		objectName, _ := entities.NewObjectName("Test Object")
		testObject := entities.ReconstructObject(
			objectID,
			objectName,
			entities.ObjectTypeGeneral,
			map[string]interface{}{},
			[]string{},
			time.Now(),
		)

		// Create test container with object
		containerName, _ := entities.NewContainerName("Test Container")
		container := entities.ReconstructContainer(
			containerID,
			collectionID,
			containerName,
			nil,
			[]entities.Object{*testObject},
			"",
			time.Now(),
			time.Now(),
		)

		// Create collection owned by different user, no group
		collectionName, _ := entities.NewCollectionName("Private Collection")
		collection := entities.ReconstructCollection(
			collectionID,
			ownerID,
			nil,
			collectionName,
			nil,
			entities.ObjectTypeGeneral,
			[]entities.Container{},
			[]string{},
			"",
			time.Now(),
			time.Now(),
		)

		req := DeleteObjectRequest{
			ContainerID: containerID,
			ObjectID:    objectID,
			UserID:      userID,
			UserToken:   "test-token",
		}

		// Mock expectations
		mockContainerRepo.EXPECT().
			GetByID(gomock.Any(), containerID).
			Return(container, nil).
			Times(1)

		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), "test-token", userID.String()).
			Return([]*entities.Group{}, nil).
			Times(1)

		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(collection, nil).
			Times(1)

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "access denied")
	})

	t.Run("error - object not found in container", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()
		containerID := entities.NewContainerID()
		objectID := entities.NewObjectID()

		// Create test container without the object
		containerName, _ := entities.NewContainerName("Test Container")
		container := entities.ReconstructContainer(
			containerID,
			collectionID,
			containerName,
			nil,
			[]entities.Object{}, // Empty
			"",
			time.Now(),
			time.Now(),
		)

		// Create test collection
		collectionName, _ := entities.NewCollectionName("Test Collection")
		collection := entities.ReconstructCollection(
			collectionID,
			userID,
			nil,
			collectionName,
			nil,
			entities.ObjectTypeGeneral,
			[]entities.Container{},
			[]string{},
			"",
			time.Now(),
			time.Now(),
		)

		req := DeleteObjectRequest{
			ContainerID: containerID,
			ObjectID:    objectID,
			UserID:      userID,
			UserToken:   "test-token",
		}

		// Mock expectations
		mockContainerRepo.EXPECT().
			GetByID(gomock.Any(), containerID).
			Return(container, nil).
			Times(1)

		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), "test-token", userID.String()).
			Return([]*entities.Group{}, nil).
			Times(1)

		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(collection, nil).
			Times(1)

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to remove object from container")
	})

	t.Run("error - auth service failure", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()
		containerID := entities.NewContainerID()
		objectID := entities.NewObjectID()

		// Create test container
		containerName, _ := entities.NewContainerName("Test Container")
		container := entities.ReconstructContainer(
			containerID,
			collectionID,
			containerName,
			nil,
			[]entities.Object{},
			"",
			time.Now(),
			time.Now(),
		)

		req := DeleteObjectRequest{
			ContainerID: containerID,
			ObjectID:    objectID,
			UserID:      userID,
			UserToken:   "test-token",
		}

		// Mock expectations
		mockContainerRepo.EXPECT().
			GetByID(gomock.Any(), containerID).
			Return(container, nil).
			Times(1)

		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), "test-token", userID.String()).
			Return(nil, errors.New("auth service unavailable")).
			Times(1)

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

		// Create test object
		objectName, _ := entities.NewObjectName("Test Object")
		testObject := entities.ReconstructObject(
			objectID,
			objectName,
			entities.ObjectTypeGeneral,
			map[string]interface{}{},
			[]string{},
			time.Now(),
		)

		// Create test container with object
		containerName, _ := entities.NewContainerName("Test Container")
		container := entities.ReconstructContainer(
			containerID,
			collectionID,
			containerName,
			nil,
			[]entities.Object{*testObject},
			"",
			time.Now(),
			time.Now(),
		)

		// Create test collection
		collectionName, _ := entities.NewCollectionName("Test Collection")
		collection := entities.ReconstructCollection(
			collectionID,
			userID,
			nil,
			collectionName,
			nil,
			entities.ObjectTypeGeneral,
			[]entities.Container{},
			[]string{},
			"",
			time.Now(),
			time.Now(),
		)

		req := DeleteObjectRequest{
			ContainerID: containerID,
			ObjectID:    objectID,
			UserID:      userID,
			UserToken:   "test-token",
		}

		// Mock expectations
		mockContainerRepo.EXPECT().
			GetByID(gomock.Any(), containerID).
			Return(container, nil).
			Times(1)

		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), "test-token", userID.String()).
			Return([]*entities.Group{}, nil).
			Times(1)

		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(collection, nil).
			Times(1)

		mockContainerRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			Return(errors.New("database connection failed")).
			Times(1)

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to save container")
	})
}
