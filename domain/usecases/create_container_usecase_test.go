package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/mocks"
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
		// Create test data
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()
		containerName := fake.Word()

		// Create test collection that user owns
		collectionName, _ := entities.NewCollectionName(fake.Company())
		objectType := entities.ObjectTypeGeneral
		testCollection, _ := entities.NewCollection(entities.CollectionProps{
			UserID:     userID,
			GroupID:    nil, // no group
			Name:       collectionName,
			ObjectType: objectType,
		})

		// Setup mock expectations
		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), gomock.Any(), userID.String()).
			Return([]*entities.Group{}, nil).
			Times(1)

		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(testCollection, nil).
			Times(1)

		mockContainerRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

		mockCollectionRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

		// Execute use case
		req := CreateContainerRequest{
			CollectionID: collectionID,
			Name:         containerName,
			UserID:       userID,
			UserToken:    "test-token",
		}

		resp, err := useCase.Execute(context.Background(), req)

		// Assert results
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Container)
		assert.Equal(t, collectionID.String(), resp.Container.CollectionID().String())
		assert.Equal(t, containerName, resp.Container.Name().String())
	})

	t.Run("error - user not owner of collection", func(t *testing.T) {
		userID := entities.NewUserID()
		differentUserID := entities.NewUserID()
		collectionID := entities.NewCollectionID()
		containerName := fake.Word()

		// Create test collection owned by different user
		collectionName, _ := entities.NewCollectionName(fake.Company())
		objectType := entities.ObjectTypeGeneral
		testCollection, _ := entities.NewCollection(entities.CollectionProps{
			UserID:     differentUserID, // different owner
			GroupID:    nil,
			Name:       collectionName,
			ObjectType: objectType,
		})

		// Setup mock expectations
		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), gomock.Any(), userID.String()).
			Return([]*entities.Group{}, nil).
			Times(1)

		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(testCollection, nil).
			Times(1)

		// Execute use case
		req := CreateContainerRequest{
			CollectionID: collectionID,
			Name:         containerName,
			UserID:       userID,
			UserToken:    "test-token",
		}

		resp, err := useCase.Execute(context.Background(), req)

		// Assert error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "access denied")
		assert.Nil(t, resp)
	})

	t.Run("error - auth service failure", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()
		containerName := fake.Word()

		// Setup mock expectations for auth service error
		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), gomock.Any(), userID.String()).
			Return(nil, errors.New("auth service error")).
			Times(1)

		// Execute use case
		req := CreateContainerRequest{
			CollectionID: collectionID,
			Name:         containerName,
			UserID:       userID,
			UserToken:    "test-token",
		}

		resp, err := useCase.Execute(context.Background(), req)

		// Assert error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get user groups")
		assert.Nil(t, resp)
	})

	t.Run("error - invalid container name", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()
		invalidName := "" // Empty name should be invalid

		// Create test collection that user owns
		collectionName, _ := entities.NewCollectionName(fake.Company())
		objectType := entities.ObjectTypeGeneral
		testCollection, _ := entities.NewCollection(entities.CollectionProps{
			UserID:     userID,
			GroupID:    nil,
			Name:       collectionName,
			ObjectType: objectType,
		})

		// Setup mock expectations
		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), gomock.Any(), userID.String()).
			Return([]*entities.Group{}, nil).
			Times(1)

		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(testCollection, nil).
			Times(1)

		// Execute use case
		req := CreateContainerRequest{
			CollectionID: collectionID,
			Name:         invalidName,
			UserID:       userID,
			UserToken:    "test-token",
		}

		resp, err := useCase.Execute(context.Background(), req)

		// Assert error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid container name")
		assert.Nil(t, resp)
	})

	t.Run("error - repository save failure", func(t *testing.T) {
		userID := entities.NewUserID()
		groupID := entities.NewGroupID()
		collectionID := entities.NewCollectionID()
		containerName := fake.Word()

		// Create test group that user belongs to
		groupName, _ := entities.NewGroupName(fake.Company())
		groupDesc := entities.NewGroupDescription("Test group")
		userGroup := entities.ReconstructGroup(
			groupID,
			groupName,
			groupDesc,
			time.Now(),
			time.Now(),
		)

		// Create test collection
		collectionName, _ := entities.NewCollectionName(fake.Company())
		testCollection, _ := entities.NewCollection(entities.CollectionProps{
			UserID:     userID,
			GroupID:    &groupID,
			Name:       collectionName,
			ObjectType: entities.ObjectTypeGeneral,
		})

		// Setup mock expectations
		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), gomock.Any(), userID.String()).
			Return([]*entities.Group{userGroup}, nil).
			Times(1)

		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(testCollection, nil).
			Times(1)

		mockContainerRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(errors.New("database error")).
			Times(1)

		// Execute use case
		req := CreateContainerRequest{
			CollectionID: collectionID,
			GroupID:      &groupID,
			Name:         containerName,
			UserID:       userID,
			UserToken:    "test-token",
		}

		resp, err := useCase.Execute(context.Background(), req)

		// Assert error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save container")
		assert.Nil(t, resp)
	})
}
