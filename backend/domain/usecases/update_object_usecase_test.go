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

func TestUpdateObjectUseCase_Execute(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockContainerRepo := mocks.NewMockContainerRepository(mockCtrl)
	mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
	mockAuthService := mocks.NewMockAuthService(mockCtrl)

	useCase := NewUpdateObjectUseCase(mockContainerRepo, mockCollectionRepo, mockAuthService)

	t.Run("success - update object name", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()
		containerID := entities.NewContainerID()
		objectID := entities.NewObjectID()

		// Create test object
		objectName, _ := entities.NewObjectName("Old Name")
		objectDesc := entities.NewObjectDescription("Test Object Description")
		testObject := entities.ReconstructObject(
			objectID,
			objectName,
			objectDesc,
			entities.ObjectTypeGeneral,
			nil,
			"",
			map[string]interface{}{},
			[]string{},
			nil,
			time.Now(),
			time.Now(),
		)

		// Create test container with object
		containerName, _ := entities.NewContainerName("Test Container")
		container := entities.ReconstructContainer(
			containerID,
			collectionID,
			containerName,
			entities.ContainerTypeGeneral,
			nil,
			nil,
			nil,
			[]entities.Object{*testObject},
			"",
			nil,
			nil,
			nil,
			nil,
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

		newName := "New Object Name"
		req := UpdateObjectRequest{
			ContainerID: containerID,
			ObjectID:    objectID,
			Name:        &newName,
			Properties:  nil,
			Tags:        nil,
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
				// Verify object name was updated
				obj, err := c.GetObject(objectID)
				require.NoError(t, err)
				assert.Equal(t, "New Object Name", obj.Name().String())
				return nil
			}).
			Times(1)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "New Object Name", resp.Object.Name().String())
	})

	t.Run("success - update object properties and tags", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()
		containerID := entities.NewContainerID()
		objectID := entities.NewObjectID()

		// Create test object
		objectName, _ := entities.NewObjectName("Test Object")
		objectDesc := entities.NewObjectDescription("Test Object Description")
		testObject := entities.ReconstructObject(
			objectID,
			objectName,
			objectDesc,
			entities.ObjectTypeGeneral,
			nil,
			"",
			map[string]interface{}{"old": "value"},
			[]string{"old-tag"},
			nil,
			time.Now(),
			time.Now(),
		)

		// Create test container with object
		containerName, _ := entities.NewContainerName("Test Container")
		container := entities.ReconstructContainer(
			containerID,
			collectionID,
			containerName,
			entities.ContainerTypeGeneral,
			nil,
			nil,
			nil,
			[]entities.Object{*testObject},
			"",
			nil,
			nil,
			nil,
			nil,
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

		req := UpdateObjectRequest{
			ContainerID: containerID,
			ObjectID:    objectID,
			Name:        nil,
			Properties:  map[string]interface{}{"new": "property", "count": 42},
			Tags:        []string{"new-tag", "updated"},
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
			Return(nil).
			Times(1)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("error - container not found", func(t *testing.T) {
		userID := entities.NewUserID()
		containerID := entities.NewContainerID()
		objectID := entities.NewObjectID()

		newName := "New Name"
		req := UpdateObjectRequest{
			ContainerID: containerID,
			ObjectID:    objectID,
			Name:        &newName,
			Properties:  nil,
			Tags:        nil,
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
			entities.ContainerTypeGeneral,
			nil,
			nil,
			nil,
			[]entities.Object{}, // Empty
			"",
			nil,
			nil,
			nil,
			nil,
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

		newName := "New Name"
		req := UpdateObjectRequest{
			ContainerID: containerID,
			ObjectID:    objectID,
			Name:        &newName,
			Properties:  nil,
			Tags:        nil,
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
		assert.Contains(t, err.Error(), "object not found in container")
	})

	t.Run("error - access denied", func(t *testing.T) {
		userID := entities.NewUserID()
		ownerID := entities.NewUserID() // Different owner
		collectionID := entities.NewCollectionID()
		containerID := entities.NewContainerID()
		objectID := entities.NewObjectID()

		// Create test object
		objectName, _ := entities.NewObjectName("Test Object")
		objectDesc := entities.NewObjectDescription("Test Object Description")
		testObject := entities.ReconstructObject(
			objectID,
			objectName,
			objectDesc,
			entities.ObjectTypeGeneral,
			nil,
			"",
			map[string]interface{}{},
			[]string{},
			nil,
			time.Now(),
			time.Now(),
		)

		// Create test container
		containerName, _ := entities.NewContainerName("Test Container")
		container := entities.ReconstructContainer(
			containerID,
			collectionID,
			containerName,
			entities.ContainerTypeGeneral,
			nil,
			nil,
			nil,
			[]entities.Object{*testObject},
			"",
			nil,
			nil,
			nil,
			nil,
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

		newName := "New Name"
		req := UpdateObjectRequest{
			ContainerID: containerID,
			ObjectID:    objectID,
			Name:        &newName,
			Properties:  nil,
			Tags:        nil,
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

	t.Run("error - invalid object name", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()
		containerID := entities.NewContainerID()
		objectID := entities.NewObjectID()

		// Create test object
		objectName, _ := entities.NewObjectName("Test Object")
		objectDesc := entities.NewObjectDescription("Test Object Description")
		testObject := entities.ReconstructObject(
			objectID,
			objectName,
			objectDesc,
			entities.ObjectTypeGeneral,
			nil,
			"",
			map[string]interface{}{},
			[]string{},
			nil,
			time.Now(),
			time.Now(),
		)

		// Create test container with object
		containerName, _ := entities.NewContainerName("Test Container")
		container := entities.ReconstructContainer(
			containerID,
			collectionID,
			containerName,
			entities.ContainerTypeGeneral,
			nil,
			nil,
			nil,
			[]entities.Object{*testObject},
			"",
			nil,
			nil,
			nil,
			nil,
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

		emptyName := "" // Invalid
		req := UpdateObjectRequest{
			ContainerID: containerID,
			ObjectID:    objectID,
			Name:        &emptyName,
			Properties:  nil,
			Tags:        nil,
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
		assert.Contains(t, err.Error(), "invalid object name")
	})
}
