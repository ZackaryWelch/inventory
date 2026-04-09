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

func TestGetCollectionsUseCase_Execute(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
	mockAuthService := mocks.NewMockAuthService(mockCtrl)

	useCase := NewGetCollectionsUseCase(mockCollectionRepo, mockAuthService)

	t.Run("success - get all user collections", func(t *testing.T) {
		userID := entities.NewUserID()

		collections := []*entities.Collection{
			NewTestCollection(ColUserID(userID), ColName("Collection A")),
			NewTestCollection(ColUserID(userID), ColName("Collection B")),
			NewTestCollection(ColUserID(userID), ColName("Collection C")),
		}

		req := GetCollectionsRequest{UserID: userID, UserToken: "test-token"}

		mockCollectionRepo.EXPECT().GetByUserIDSummary(gomock.Any(), userID).Return(collections, nil)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Collections, 3)
	})

	t.Run("success - get single collection owned by user", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		collection := NewTestCollection(ColID(collectionID), ColUserID(userID), ColName("My Collection"))

		req := GetCollectionsRequest{UserID: userID, CollectionID: &collectionID, UserToken: "test-token"}

		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(collection, nil)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Collections, 1)
		assert.Equal(t, collectionID, resp.Collections[0].ID())
	})

	t.Run("success - get single collection in user's group", func(t *testing.T) {
		userID := entities.NewUserID()
		groupID := entities.NewGroupID()
		collectionID := entities.NewCollectionID()
		ownerID := entities.NewUserID()

		testGroup := NewTestGroup(GrpID(groupID))
		collection := NewTestCollection(ColID(collectionID), ColUserID(ownerID), ColGroupID(&groupID), ColName("Shared Collection"))

		req := GetCollectionsRequest{UserID: userID, CollectionID: &collectionID, UserToken: "test-token"}

		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(collection, nil)
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", userID.String()).Return([]*entities.Group{testGroup}, nil)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Collections, 1)
		assert.Equal(t, collectionID, resp.Collections[0].ID())
	})

	t.Run("success - get empty collection list", func(t *testing.T) {
		userID := entities.NewUserID()

		req := GetCollectionsRequest{UserID: userID, UserToken: "test-token"}

		mockCollectionRepo.EXPECT().GetByUserIDSummary(gomock.Any(), userID).Return([]*entities.Collection{}, nil)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Collections, 0)
	})

	t.Run("error - collection not found", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		req := GetCollectionsRequest{UserID: userID, CollectionID: &collectionID, UserToken: "test-token"}

		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(nil, errors.New("collection not found"))

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "collection not found")
	})

	t.Run("error - access denied to collection", func(t *testing.T) {
		userID := entities.NewUserID()
		ownerID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		collection := NewTestCollection(ColID(collectionID), ColUserID(ownerID), ColName("Private Collection"))

		req := GetCollectionsRequest{UserID: userID, CollectionID: &collectionID, UserToken: "test-token"}

		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(collection, nil)

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "access denied")
	})

	t.Run("error - access denied to group collection (not in group)", func(t *testing.T) {
		userID := entities.NewUserID()
		ownerID := entities.NewUserID()
		groupID := entities.NewGroupID()
		collectionID := entities.NewCollectionID()

		collection := NewTestCollection(ColID(collectionID), ColUserID(ownerID), ColGroupID(&groupID))

		req := GetCollectionsRequest{UserID: userID, CollectionID: &collectionID, UserToken: "test-token"}

		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(collection, nil)
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", userID.String()).Return([]*entities.Group{}, nil)

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "access denied")
	})

	t.Run("error - auth service failure", func(t *testing.T) {
		userID := entities.NewUserID()
		ownerID := entities.NewUserID()
		groupID := entities.NewGroupID()
		collectionID := entities.NewCollectionID()

		collection := NewTestCollection(ColID(collectionID), ColUserID(ownerID), ColGroupID(&groupID))

		req := GetCollectionsRequest{UserID: userID, CollectionID: &collectionID, UserToken: "test-token"}

		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(collection, nil)
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", userID.String()).Return(nil, errors.New("auth service unavailable"))

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to get user groups")
	})

	t.Run("error - repository failure getting all collections", func(t *testing.T) {
		userID := entities.NewUserID()

		req := GetCollectionsRequest{UserID: userID, UserToken: "test-token"}

		mockCollectionRepo.EXPECT().GetByUserIDSummary(gomock.Any(), userID).Return(nil, errors.New("database connection failed"))

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to get collections")
	})
}
