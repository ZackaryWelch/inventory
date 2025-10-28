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

func TestGetCollectionsUseCase_Execute(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
	mockAuthService := mocks.NewMockAuthService(mockCtrl)

	useCase := NewGetCollectionsUseCase(mockCollectionRepo, mockAuthService)

	t.Run("success - get all user collections", func(t *testing.T) {
		userID := entities.NewUserID()

		// Create test collections
		collections := []*entities.Collection{}
		for i := 0; i < 3; i++ {
			collectionName, _ := entities.NewCollectionName("Collection " + string(rune(i+'A')))
			collection := entities.ReconstructCollection(
				entities.NewCollectionID(),
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
			collections = append(collections, collection)
		}

		req := GetCollectionsRequest{
			UserID:       userID,
			CollectionID: nil,
			UserToken:    "test-token",
		}

		// Mock expectations
		mockCollectionRepo.EXPECT().
			GetByUserID(gomock.Any(), userID).
			Return(collections, nil).
			Times(1)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Collections, 3)
	})

	t.Run("success - get single collection owned by user", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		// Create test collection
		collectionName, _ := entities.NewCollectionName("My Collection")
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

		req := GetCollectionsRequest{
			UserID:       userID,
			CollectionID: &collectionID,
			UserToken:    "test-token",
		}

		// Mock expectations
		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(collection, nil).
			Times(1)

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
		ownerID := entities.NewUserID() // Different owner

		// Create test group
		groupName, _ := entities.NewGroupName("Test Group")
		testGroup := entities.ReconstructGroup(
			groupID,
			groupName,
			time.Now(),
			time.Now(),
		)

		// Create collection owned by different user but in user's group
		collectionName, _ := entities.NewCollectionName("Shared Collection")
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

		req := GetCollectionsRequest{
			UserID:       userID,
			CollectionID: &collectionID,
			UserToken:    "test-token",
		}

		// Mock expectations
		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(collection, nil).
			Times(1)

		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), "test-token", userID.String()).
			Return([]*entities.Group{testGroup}, nil).
			Times(1)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Collections, 1)
		assert.Equal(t, collectionID, resp.Collections[0].ID())
	})

	t.Run("success - get empty collection list", func(t *testing.T) {
		userID := entities.NewUserID()

		req := GetCollectionsRequest{
			UserID:       userID,
			CollectionID: nil,
			UserToken:    "test-token",
		}

		// Mock expectations
		mockCollectionRepo.EXPECT().
			GetByUserID(gomock.Any(), userID).
			Return([]*entities.Collection{}, nil).
			Times(1)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Collections, 0)
	})

	t.Run("error - collection not found", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		req := GetCollectionsRequest{
			UserID:       userID,
			CollectionID: &collectionID,
			UserToken:    "test-token",
		}

		// Mock repository returns error
		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(nil, errors.New("collection not found")).
			Times(1)

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "collection not found")
	})

	t.Run("error - access denied to collection", func(t *testing.T) {
		userID := entities.NewUserID()
		ownerID := entities.NewUserID() // Different owner
		collectionID := entities.NewCollectionID()

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

		req := GetCollectionsRequest{
			UserID:       userID,
			CollectionID: &collectionID,
			UserToken:    "test-token",
		}

		// Mock expectations
		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(collection, nil).
			Times(1)

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

		// Create collection in a group
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

		req := GetCollectionsRequest{
			UserID:       userID,
			CollectionID: &collectionID,
			UserToken:    "test-token",
		}

		// Mock expectations
		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(collection, nil).
			Times(1)

		// User is not in any groups
		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), "test-token", userID.String()).
			Return([]*entities.Group{}, nil).
			Times(1)

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

		// Create collection in a group
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

		req := GetCollectionsRequest{
			UserID:       userID,
			CollectionID: &collectionID,
			UserToken:    "test-token",
		}

		// Mock expectations
		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(collection, nil).
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

	t.Run("error - repository failure getting all collections", func(t *testing.T) {
		userID := entities.NewUserID()

		req := GetCollectionsRequest{
			UserID:       userID,
			CollectionID: nil,
			UserToken:    "test-token",
		}

		// Mock repository returns error
		mockCollectionRepo.EXPECT().
			GetByUserID(gomock.Any(), userID).
			Return(nil, errors.New("database connection failed")).
			Times(1)

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to get collections")
	})
}
