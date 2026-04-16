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

		obj := NewTestObject(ObjID(objectID), ObjName("Old Name"))
		container := NewTestContainer(CtrID(containerID), CtrCollectionID(collectionID), CtrObjects(*obj))
		collection := NewTestCollection(ColID(collectionID), ColUserID(userID))

		newName := "New Object Name"
		req := UpdateObjectRequest{
			ContainerID: &containerID,
			ObjectID:    objectID,
			Name:        &newName,
			UserID:      userID,
			UserToken:   "test-token",
		}

		mockContainerRepo.EXPECT().FindByObjectID(gomock.Any(), objectID).Return(container, nil)
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", userID.String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(collection, nil)
		mockContainerRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, c *entities.Container) error {
			obj, err := c.GetObject(objectID)
			require.NoError(t, err)
			assert.Equal(t, "New Object Name", obj.Name().String())
			return nil
		})

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

		obj := NewTestObject(ObjID(objectID), ObjProps(Props("old", "value")), ObjTags("old-tag"))
		container := NewTestContainer(CtrID(containerID), CtrCollectionID(collectionID), CtrObjects(*obj))
		collection := NewTestCollection(ColID(collectionID), ColUserID(userID))

		req := UpdateObjectRequest{
			ContainerID: &containerID,
			ObjectID:    objectID,
			Properties:  Props("new", "property", "count", entities.TypedValue{Val: 42}),
			Tags:        []string{"new-tag", "updated"},
			UserID:      userID,
			UserToken:   "test-token",
		}

		mockContainerRepo.EXPECT().FindByObjectID(gomock.Any(), objectID).Return(container, nil)
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", userID.String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(collection, nil)
		mockContainerRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("error - object not found", func(t *testing.T) {
		userID := entities.NewUserID()
		containerID := entities.NewContainerID()
		objectID := entities.NewObjectID()

		newName := "New Name"
		req := UpdateObjectRequest{
			ContainerID: &containerID,
			ObjectID:    objectID,
			Name:        &newName,
			UserID:      userID,
			UserToken:   "test-token",
		}

		mockContainerRepo.EXPECT().FindByObjectID(gomock.Any(), objectID).Return(nil, errors.New("object not found"))

		resp, err := useCase.Execute(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "object not found")
	})

	t.Run("error - object not found in container", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()
		containerID := entities.NewContainerID()
		objectID := entities.NewObjectID()

		container := NewTestContainer(CtrID(containerID), CtrCollectionID(collectionID))
		collection := NewTestCollection(ColID(collectionID), ColUserID(userID))

		newName := "New Name"
		req := UpdateObjectRequest{
			ContainerID: &containerID,
			ObjectID:    objectID,
			Name:        &newName,
			UserID:      userID,
			UserToken:   "test-token",
		}

		mockContainerRepo.EXPECT().FindByObjectID(gomock.Any(), objectID).Return(container, nil)
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", userID.String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(collection, nil)

		resp, err := useCase.Execute(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "object not found in container")
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

		newName := "New Name"
		req := UpdateObjectRequest{
			ContainerID: &containerID,
			ObjectID:    objectID,
			Name:        &newName,
			UserID:      userID,
			UserToken:   "test-token",
		}

		mockContainerRepo.EXPECT().FindByObjectID(gomock.Any(), objectID).Return(container, nil)
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", userID.String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(collection, nil)

		resp, err := useCase.Execute(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "access denied")
	})

	t.Run("error - invalid object name", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()
		containerID := entities.NewContainerID()
		objectID := entities.NewObjectID()

		obj := NewTestObject(ObjID(objectID))
		container := NewTestContainer(CtrID(containerID), CtrCollectionID(collectionID), CtrObjects(*obj))
		collection := NewTestCollection(ColID(collectionID), ColUserID(userID))

		emptyName := ""
		req := UpdateObjectRequest{
			ContainerID: &containerID,
			ObjectID:    objectID,
			Name:        &emptyName,
			UserID:      userID,
			UserToken:   "test-token",
		}

		mockContainerRepo.EXPECT().FindByObjectID(gomock.Any(), objectID).Return(container, nil)
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", userID.String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(collection, nil)

		resp, err := useCase.Execute(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid object name")
	})
}
