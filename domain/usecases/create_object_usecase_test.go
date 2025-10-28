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

func TestCreateObjectUseCase_Execute(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockContainerRepo := mocks.NewMockContainerRepository(mockCtrl)
	mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
	mockAuthService := mocks.NewMockAuthService(mockCtrl)

	useCase := NewCreateObjectUseCase(mockContainerRepo, mockCollectionRepo, mockAuthService)

	t.Run("success - create object as collection owner", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()
		containerID := entities.NewContainerID()

		// Create test container
		containerName, _ := entities.NewContainerName("Test Container")
		container, _ := entities.NewContainer(entities.ContainerProps{
			CollectionID: collectionID,
			Name:         containerName,
		})
		container = entities.ReconstructContainer(
			containerID,
			collectionID,
			containerName,
			nil, // No category
			[]entities.Object{}, // Empty objects
			"", // Empty location
			time.Now(),
			time.Now(),
		)

		// Create test collection owned by user
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

		req := CreateObjectRequest{
			ContainerID: containerID,
			Name:        "Test Object",
			ObjectType:  entities.ObjectTypeGeneral,
			Properties:  map[string]interface{}{"description": "Test description"},
			Tags:        []string{"test"},
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
				// Verify object was added
				objects := c.Objects()
				require.Len(t, objects, 1)
				assert.Equal(t, "Test Object", objects[0].Name().String())
				return nil
			}).
			Times(1)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.NotNil(t, resp.Object)
		assert.Equal(t, "Test Object", resp.Object.Name().String())
		assert.Equal(t, entities.ObjectTypeGeneral, resp.Object.ObjectType())
	})

	t.Run("success - create object in group collection", func(t *testing.T) {
		userID := entities.NewUserID()
		groupID := entities.NewGroupID()
		collectionID := entities.NewCollectionID()
		containerID := entities.NewContainerID()
		ownerID := entities.NewUserID() // Different owner

		// Create test group
		groupName, _ := entities.NewGroupName("Test Group")
		testGroup := entities.ReconstructGroup(
			groupID,
			groupName,
			time.Now(),
			time.Now(),
		)

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

		// Create collection owned by different user but in user's group
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

		req := CreateObjectRequest{
			ContainerID: containerID,
			Name:        "Group Object",
			ObjectType:  entities.ObjectTypeGeneral,
			Properties:  map[string]interface{}{},
			Tags:        []string{},
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
		assert.NotNil(t, resp.Object)
	})

	t.Run("error - container not found", func(t *testing.T) {
		userID := entities.NewUserID()
		containerID := entities.NewContainerID()

		req := CreateObjectRequest{
			ContainerID: containerID,
			Name:        "Test Object",
			ObjectType:  entities.ObjectTypeGeneral,
			Properties:  map[string]interface{}{},
			Tags:        []string{},
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

		req := CreateObjectRequest{
			ContainerID: containerID,
			Name:        "Test Object",
			ObjectType:  entities.ObjectTypeGeneral,
			Properties:  map[string]interface{}{},
			Tags:        []string{},
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

		// Create collection owned by different user, no group
		collectionName, _ := entities.NewCollectionName("Private Collection")
		collection := entities.ReconstructCollection(
			collectionID,
			ownerID,
			nil, // No group
			collectionName,
			nil,
			entities.ObjectTypeGeneral,
			[]entities.Container{},
			[]string{},
			"",
			time.Now(),
			time.Now(),
		)

		req := CreateObjectRequest{
			ContainerID: containerID,
			Name:        "Test Object",
			ObjectType:  entities.ObjectTypeGeneral,
			Properties:  map[string]interface{}{},
			Tags:        []string{},
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

		req := CreateObjectRequest{
			ContainerID: containerID,
			Name:        "", // Empty name is invalid
			ObjectType:  entities.ObjectTypeGeneral,
			Properties:  map[string]interface{}{},
			Tags:        []string{},
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

	t.Run("error - auth service failure", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()
		containerID := entities.NewContainerID()

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

		req := CreateObjectRequest{
			ContainerID: containerID,
			Name:        "Test Object",
			ObjectType:  entities.ObjectTypeGeneral,
			Properties:  map[string]interface{}{},
			Tags:        []string{},
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

		req := CreateObjectRequest{
			ContainerID: containerID,
			Name:        "Test Object",
			ObjectType:  entities.ObjectTypeGeneral,
			Properties:  map[string]interface{}{},
			Tags:        []string{},
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
