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

		container := NewTestContainer(CtrID(containerID), CtrCollectionID(collectionID))
		collection := NewTestCollection(ColID(collectionID), ColUserID(userID))

		req := CreateObjectRequest{
			ContainerID: &containerID,
			Name:        "Test Object",
			ObjectType:  entities.ObjectTypeGeneral,
			Properties:  Props("description", "Test description"),
			Tags:        []string{"test"},
			UserID:      userID,
			UserToken:   "test-token",
		}

		mockContainerRepo.EXPECT().GetByID(gomock.Any(), containerID).Return(container, nil)
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", userID.String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(collection, nil)
		mockContainerRepo.EXPECT().AddObject(gomock.Any(), containerID, gomock.Any()).Return(nil)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "Test Object", resp.Object.Name().String())
		assert.Equal(t, entities.ObjectTypeGeneral, resp.Object.ObjectType())
	})

	t.Run("success - create object in group collection", func(t *testing.T) {
		userID := entities.NewUserID()
		groupID := entities.NewGroupID()
		collectionID := entities.NewCollectionID()
		containerID := entities.NewContainerID()
		ownerID := entities.NewUserID()

		testGroup := NewTestGroup(GrpID(groupID))
		container := NewTestContainer(CtrID(containerID), CtrCollectionID(collectionID))
		collection := NewTestCollection(ColID(collectionID), ColUserID(ownerID), ColGroupID(&groupID))

		req := CreateObjectRequest{
			ContainerID: &containerID,
			Name:        "Group Object",
			ObjectType:  entities.ObjectTypeGeneral,
			Properties:  map[string]entities.TypedValue{},
			Tags:        []string{},
			UserID:      userID,
			UserToken:   "test-token",
		}

		mockContainerRepo.EXPECT().GetByID(gomock.Any(), containerID).Return(container, nil)
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", userID.String()).Return([]*entities.Group{testGroup}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(collection, nil)
		mockContainerRepo.EXPECT().AddObject(gomock.Any(), containerID, gomock.Any()).Return(nil)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.NotNil(t, resp.Object)
	})

	t.Run("error - container not found", func(t *testing.T) {
		userID := entities.NewUserID()
		containerID := entities.NewContainerID()

		req := CreateObjectRequest{
			ContainerID: &containerID,
			Name:        "Test Object",
			ObjectType:  entities.ObjectTypeGeneral,
			Properties:  map[string]entities.TypedValue{},
			Tags:        []string{},
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

		container := NewTestContainer(CtrID(containerID), CtrCollectionID(collectionID))

		req := CreateObjectRequest{
			ContainerID: &containerID,
			Name:        "Test Object",
			ObjectType:  entities.ObjectTypeGeneral,
			Properties:  map[string]entities.TypedValue{},
			Tags:        []string{},
			UserID:      userID,
			UserToken:   "test-token",
		}

		mockContainerRepo.EXPECT().GetByID(gomock.Any(), containerID).Return(container, nil)
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

		container := NewTestContainer(CtrID(containerID), CtrCollectionID(collectionID))
		collection := NewTestCollection(ColID(collectionID), ColUserID(ownerID))

		req := CreateObjectRequest{
			ContainerID: &containerID,
			Name:        "Test Object",
			ObjectType:  entities.ObjectTypeGeneral,
			Properties:  map[string]entities.TypedValue{},
			Tags:        []string{},
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

	t.Run("error - invalid object name", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()
		containerID := entities.NewContainerID()

		container := NewTestContainer(CtrID(containerID), CtrCollectionID(collectionID))
		collection := NewTestCollection(ColID(collectionID), ColUserID(userID))

		req := CreateObjectRequest{
			ContainerID: &containerID,
			Name:        "",
			ObjectType:  entities.ObjectTypeGeneral,
			Properties:  map[string]entities.TypedValue{},
			Tags:        []string{},
			UserID:      userID,
			UserToken:   "test-token",
		}

		mockContainerRepo.EXPECT().GetByID(gomock.Any(), containerID).Return(container, nil)
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", userID.String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(collection, nil)

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid object name")
	})

	t.Run("error - auth service failure", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()
		containerID := entities.NewContainerID()

		container := NewTestContainer(CtrID(containerID), CtrCollectionID(collectionID))
		collection := NewTestCollection(ColID(collectionID), ColUserID(userID))

		req := CreateObjectRequest{
			ContainerID: &containerID,
			Name:        "Test Object",
			ObjectType:  entities.ObjectTypeGeneral,
			Properties:  map[string]entities.TypedValue{},
			Tags:        []string{},
			UserID:      userID,
			UserToken:   "test-token",
		}

		mockContainerRepo.EXPECT().GetByID(gomock.Any(), containerID).Return(container, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(collection, nil)
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

		container := NewTestContainer(CtrID(containerID), CtrCollectionID(collectionID))
		collection := NewTestCollection(ColID(collectionID), ColUserID(userID))

		req := CreateObjectRequest{
			ContainerID: &containerID,
			Name:        "Test Object",
			ObjectType:  entities.ObjectTypeGeneral,
			Properties:  map[string]entities.TypedValue{},
			Tags:        []string{},
			UserID:      userID,
			UserToken:   "test-token",
		}

		mockContainerRepo.EXPECT().GetByID(gomock.Any(), containerID).Return(container, nil)
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", userID.String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(collection, nil)
		mockContainerRepo.EXPECT().AddObject(gomock.Any(), containerID, gomock.Any()).Return(errors.New("database connection failed"))

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to add object to container")
	})
}
