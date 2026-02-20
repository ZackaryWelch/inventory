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

func TestDeleteCollectionUseCase_Execute(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)

	useCase := NewDeleteCollectionUseCase(mockCollectionRepo)

	t.Run("success - delete empty collection", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		// Create collection without containers
		collectionName, _ := entities.NewCollectionName("Test Collection")
		existingCollection := entities.ReconstructCollection(
			collectionID,
			userID,
			nil,
			collectionName,
			nil,
			entities.ObjectTypeGeneral,
			[]entities.Container{}, // Empty
			[]string{},
			"",
			time.Now(),
			time.Now(),
		)

		req := DeleteCollectionRequest{
			CollectionID: collectionID,
			UserID:       userID,
		}

		// Mock expectations
		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(existingCollection, nil).
			Times(1)

		mockCollectionRepo.EXPECT().
			Delete(gomock.Any(), collectionID).
			Return(nil).
			Times(1)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)
	})

	t.Run("error - collection not found", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		req := DeleteCollectionRequest{
			CollectionID: collectionID,
			UserID:       userID,
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

		req := DeleteCollectionRequest{
			CollectionID: collectionID,
			UserID:       differentUserID, // Different user trying to delete
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

	t.Run("error - cannot delete collection with containers", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		// Create container
		containerName, _ := entities.NewContainerName("Test Container")
		container, _ := entities.NewContainer(entities.ContainerProps{
			CollectionID: collectionID,
			Name:         containerName,
		})

		// Create collection with container
		collectionName, _ := entities.NewCollectionName("Test Collection")
		existingCollection := entities.ReconstructCollection(
			collectionID,
			userID,
			nil,
			collectionName,
			nil,
			entities.ObjectTypeGeneral,
			[]entities.Container{*container}, // Has container
			[]string{},
			"",
			time.Now(),
			time.Now(),
		)

		req := DeleteCollectionRequest{
			CollectionID: collectionID,
			UserID:       userID,
		}

		// Mock expectations
		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(existingCollection, nil).
			Times(1)

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "cannot delete collection with containers")
	})

	t.Run("error - repository delete failure", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		// Create empty collection
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

		req := DeleteCollectionRequest{
			CollectionID: collectionID,
			UserID:       userID,
		}

		// Mock expectations
		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(existingCollection, nil).
			Times(1)

		mockCollectionRepo.EXPECT().
			Delete(gomock.Any(), collectionID).
			Return(errors.New("database connection failed")).
			Times(1)

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to delete collection")
	})
}
