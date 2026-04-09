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

func TestCreateContainerUseCase_Execute(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	mockContainerRepo := mocks.NewMockContainerRepository(mockCtrl)
	mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
	mockAuthService := mocks.NewMockAuthService(mockCtrl)

	useCase := NewCreateContainerUseCase(mockContainerRepo, mockCollectionRepo, mockAuthService)

	t.Run("success - create container", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()
		containerName := fake.Word()

		collectionName, _ := entities.NewCollectionName(fake.Company())
		testCollection, _ := entities.NewCollection(entities.CollectionProps{
			UserID: userID, Name: collectionName, ObjectType: entities.ObjectTypeGeneral,
		})

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), gomock.Any(), userID.String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(testCollection, nil)
		mockContainerRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
		mockCollectionRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

		resp, err := useCase.Execute(context.Background(), CreateContainerRequest{
			CollectionID: collectionID, Name: containerName, UserID: userID, UserToken: "test-token",
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, collectionID.String(), resp.Container.CollectionID().String())
		assert.Equal(t, containerName, resp.Container.Name().String())
	})

	t.Run("error - user not owner of collection", func(t *testing.T) {
		userID := entities.NewUserID()
		differentUserID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		collectionName, _ := entities.NewCollectionName(fake.Company())
		testCollection, _ := entities.NewCollection(entities.CollectionProps{
			UserID: differentUserID, Name: collectionName, ObjectType: entities.ObjectTypeGeneral,
		})

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), gomock.Any(), userID.String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(testCollection, nil)

		resp, err := useCase.Execute(context.Background(), CreateContainerRequest{
			CollectionID: collectionID, Name: fake.Word(), UserID: userID, UserToken: "test-token",
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "access denied")
		assert.Nil(t, resp)
	})

	t.Run("error - auth service failure", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), gomock.Any(), userID.String()).Return(nil, errors.New("auth service error"))

		resp, err := useCase.Execute(context.Background(), CreateContainerRequest{
			CollectionID: collectionID, Name: fake.Word(), UserID: userID, UserToken: "test-token",
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get user groups")
		assert.Nil(t, resp)
	})

	t.Run("error - invalid container name", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		collectionName, _ := entities.NewCollectionName(fake.Company())
		testCollection, _ := entities.NewCollection(entities.CollectionProps{
			UserID: userID, Name: collectionName, ObjectType: entities.ObjectTypeGeneral,
		})

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), gomock.Any(), userID.String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(testCollection, nil)

		resp, err := useCase.Execute(context.Background(), CreateContainerRequest{
			CollectionID: collectionID, Name: "", UserID: userID, UserToken: "test-token",
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid container name")
		assert.Nil(t, resp)
	})

	t.Run("error - repository save failure", func(t *testing.T) {
		userID := entities.NewUserID()
		groupID := entities.NewGroupID()
		collectionID := entities.NewCollectionID()

		userGroup := NewTestGroup(GrpID(groupID))

		collectionName, _ := entities.NewCollectionName(fake.Company())
		testCollection, _ := entities.NewCollection(entities.CollectionProps{
			UserID: userID, GroupID: &groupID, Name: collectionName, ObjectType: entities.ObjectTypeGeneral,
		})

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), gomock.Any(), userID.String()).Return([]*entities.Group{userGroup}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(testCollection, nil)
		mockContainerRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("database error"))

		resp, err := useCase.Execute(context.Background(), CreateContainerRequest{
			CollectionID: collectionID, GroupID: &groupID, Name: fake.Word(), UserID: userID, UserToken: "test-token",
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save container")
		assert.Nil(t, resp)
	})
}
