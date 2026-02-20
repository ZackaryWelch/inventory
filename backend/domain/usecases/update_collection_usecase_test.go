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

func TestUpdateCollectionUseCase_Execute(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
	mockAuthService := mocks.NewMockAuthService(mockCtrl)

	useCase := NewUpdateCollectionUseCase(mockCollectionRepo, mockAuthService)

	t.Run("success - update collection name", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		// Create existing collection
		collectionName, _ := entities.NewCollectionName("Old Name")
		existingCollection := entities.ReconstructCollection(
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

		newName := "New Collection Name"
		req := UpdateCollectionRequest{
			CollectionID: collectionID,
			UserID:       userID,
			Name:         &newName,
			Tags:         []string{},
			Location:     nil,
			UserToken:    "test-token",
		}

		// Mock expectations
		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(existingCollection, nil).
			Times(1)

		mockCollectionRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, collection *entities.Collection) error {
				assert.Equal(t, "New Collection Name", collection.Name().String())
				return nil
			}).
			Times(1)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "New Collection Name", resp.Collection.Name().String())
	})

	t.Run("success - update collection location and tags", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		// Create existing collection
		collectionName, _ := entities.NewCollectionName("Test Collection")
		existingCollection := entities.ReconstructCollection(
			collectionID,
			userID,
			nil,
			collectionName,
			nil,
			entities.ObjectTypeGeneral,
			[]entities.Container{},
			[]string{"old-tag"},
			"Old Location",
			time.Now(),
			time.Now(),
		)

		newLocation := "New Location"
		req := UpdateCollectionRequest{
			CollectionID: collectionID,
			UserID:       userID,
			Name:         nil,
			Tags:         []string{"new-tag", "updated"},
			Location:     &newLocation,
			UserToken:    "test-token",
		}

		// Mock expectations
		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(existingCollection, nil).
			Times(1)

		mockCollectionRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, collection *entities.Collection) error {
				assert.Equal(t, "New Location", collection.Location())
				return nil
			}).
			Times(1)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "New Location", resp.Collection.Location())
	})

	t.Run("success - update only tags", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		// Create existing collection
		collectionName, _ := entities.NewCollectionName("Test Collection")
		existingCollection := entities.ReconstructCollection(
			collectionID,
			userID,
			nil,
			collectionName,
			nil,
			entities.ObjectTypeGeneral,
			[]entities.Container{},
			[]string{"old-tag"},
			"Location",
			time.Now(),
			time.Now(),
		)

		req := UpdateCollectionRequest{
			CollectionID: collectionID,
			UserID:       userID,
			Name:         nil,
			Tags:         []string{"new-tag"},
			Location:     nil,
			UserToken:    "test-token",
		}

		// Mock expectations
		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(existingCollection, nil).
			Times(1)

		mockCollectionRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("error - collection not found", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		newName := "New Name"
		req := UpdateCollectionRequest{
			CollectionID: collectionID,
			UserID:       userID,
			Name:         &newName,
			Tags:         []string{},
			Location:     nil,
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

	t.Run("error - access denied (not owner)", func(t *testing.T) {
		ownerID := entities.NewUserID()
		differentUserID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		// Create collection owned by different user
		collectionName, _ := entities.NewCollectionName("Test Collection")
		existingCollection := entities.ReconstructCollection(
			collectionID,
			ownerID, // Different owner
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
		req := UpdateCollectionRequest{
			CollectionID: collectionID,
			UserID:       differentUserID, // Different user trying to update
			Name:         &newName,
			Tags:         []string{},
			Location:     nil,
			UserToken:    "test-token",
		}

		// Mock expectations
		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(existingCollection, nil).
			Times(1)

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "access denied")
	})

	t.Run("error - invalid collection name", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		// Create existing collection
		collectionName, _ := entities.NewCollectionName("Old Name")
		existingCollection := entities.ReconstructCollection(
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

		emptyName := "" // Invalid empty name
		req := UpdateCollectionRequest{
			CollectionID: collectionID,
			UserID:       userID,
			Name:         &emptyName,
			Tags:         []string{},
			Location:     nil,
			UserToken:    "test-token",
		}

		// Mock expectations
		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(existingCollection, nil).
			Times(1)

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid collection name")
	})

	t.Run("error - repository update failure", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		// Create existing collection
		collectionName, _ := entities.NewCollectionName("Test Collection")
		existingCollection := entities.ReconstructCollection(
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
		req := UpdateCollectionRequest{
			CollectionID: collectionID,
			UserID:       userID,
			Name:         &newName,
			Tags:         []string{},
			Location:     nil,
			UserToken:    "test-token",
		}

		// Mock expectations
		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(existingCollection, nil).
			Times(1)

		mockCollectionRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			Return(errors.New("database connection failed")).
			Times(1)

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to update collection")
	})
}
