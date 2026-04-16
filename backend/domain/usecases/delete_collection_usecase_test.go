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

func TestDeleteCollectionUseCase_Execute(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
	mockContainerRepo := mocks.NewMockContainerRepository(mockCtrl)

	useCase := NewDeleteCollectionUseCase(mockCollectionRepo, mockContainerRepo)

	t.Run("success - delete empty collection", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		collection := NewTestCollection(ColID(collectionID), ColUserID(userID))

		req := DeleteCollectionRequest{CollectionID: collectionID, UserID: userID}

		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(collection, nil)
		mockCollectionRepo.EXPECT().Delete(gomock.Any(), collectionID).Return(nil)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)
		assert.Equal(t, int64(0), resp.ContainersDeleted)
	})

	t.Run("success - force delete collection with containers", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		ctr := NewTestContainer(CtrCollectionID(collectionID))
		collection := NewTestCollection(ColID(collectionID), ColUserID(userID), ColContainers(*ctr))

		req := DeleteCollectionRequest{CollectionID: collectionID, UserID: userID, Force: true}

		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(collection, nil)
		mockContainerRepo.EXPECT().DeleteByCollectionID(gomock.Any(), collectionID).Return(int64(1), nil)
		mockCollectionRepo.EXPECT().Delete(gomock.Any(), collectionID).Return(nil)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)
		assert.Equal(t, int64(1), resp.ContainersDeleted)
	})

	t.Run("error - collection has containers without force", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		ctr := NewTestContainer(CtrCollectionID(collectionID))
		collection := NewTestCollection(ColID(collectionID), ColUserID(userID), ColContainers(*ctr))

		req := DeleteCollectionRequest{CollectionID: collectionID, UserID: userID}

		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(collection, nil)

		resp, err := useCase.Execute(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "collection has containers")
	})

	t.Run("error - collection not found", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		req := DeleteCollectionRequest{CollectionID: collectionID, UserID: userID}

		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(nil, errors.New("collection not found"))

		resp, err := useCase.Execute(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "collection not found")
	})

	t.Run("error - access denied (not owner)", func(t *testing.T) {
		ownerID := entities.NewUserID()
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		collection := NewTestCollection(ColID(collectionID), ColUserID(ownerID))

		req := DeleteCollectionRequest{CollectionID: collectionID, UserID: userID}

		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(collection, nil)

		resp, err := useCase.Execute(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "access denied")
	})

	t.Run("error - repository delete failure", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		collection := NewTestCollection(ColID(collectionID), ColUserID(userID))

		req := DeleteCollectionRequest{CollectionID: collectionID, UserID: userID}

		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(collection, nil)
		mockCollectionRepo.EXPECT().Delete(gomock.Any(), collectionID).Return(errors.New("database connection failed"))

		resp, err := useCase.Execute(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to delete collection")
	})
}
