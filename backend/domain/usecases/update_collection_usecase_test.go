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

		existing := NewTestCollection(ColID(collectionID), ColUserID(userID), ColName("Old Name"))

		newName := "New Collection Name"
		req := UpdateCollectionRequest{
			CollectionID: collectionID, UserID: userID, Name: &newName, Tags: []string{}, UserToken: "test-token",
		}

		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(existing, nil)
		mockCollectionRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, c *entities.Collection) error {
			assert.Equal(t, "New Collection Name", c.Name().String())
			return nil
		})

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "New Collection Name", resp.Collection.Name().String())
	})

	t.Run("success - update collection location and tags", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		existing := NewTestCollection(ColID(collectionID), ColUserID(userID), ColTags("old-tag"), ColLocation("Old Location"))

		newLocation := "New Location"
		req := UpdateCollectionRequest{
			CollectionID: collectionID, UserID: userID, Tags: []string{"new-tag", "updated"}, Location: &newLocation, UserToken: "test-token",
		}

		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(existing, nil)
		mockCollectionRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, c *entities.Collection) error {
			assert.Equal(t, "New Location", c.Location())
			return nil
		})

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "New Location", resp.Collection.Location())
	})

	t.Run("success - update only tags", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		existing := NewTestCollection(ColID(collectionID), ColUserID(userID), ColTags("old-tag"), ColLocation("Location"))

		req := UpdateCollectionRequest{
			CollectionID: collectionID, UserID: userID, Tags: []string{"new-tag"}, UserToken: "test-token",
		}

		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(existing, nil)
		mockCollectionRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("error - collection not found", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		newName := "New Name"
		req := UpdateCollectionRequest{
			CollectionID: collectionID, UserID: userID, Name: &newName, Tags: []string{}, UserToken: "test-token",
		}

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

		existing := NewTestCollection(ColID(collectionID), ColUserID(ownerID))

		newName := "New Name"
		req := UpdateCollectionRequest{
			CollectionID: collectionID, UserID: userID, Name: &newName, Tags: []string{}, UserToken: "test-token",
		}

		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(existing, nil)

		resp, err := useCase.Execute(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "access denied")
	})

	t.Run("error - invalid collection name", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		existing := NewTestCollection(ColID(collectionID), ColUserID(userID), ColName("Old Name"))

		emptyName := ""
		req := UpdateCollectionRequest{
			CollectionID: collectionID, UserID: userID, Name: &emptyName, Tags: []string{}, UserToken: "test-token",
		}

		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(existing, nil)

		resp, err := useCase.Execute(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid collection name")
	})

	t.Run("error - repository update failure", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		existing := NewTestCollection(ColID(collectionID), ColUserID(userID))

		newName := "New Name"
		req := UpdateCollectionRequest{
			CollectionID: collectionID, UserID: userID, Name: &newName, Tags: []string{}, UserToken: "test-token",
		}

		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(existing, nil)
		mockCollectionRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("database connection failed"))

		resp, err := useCase.Execute(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to update collection")
	})
}
