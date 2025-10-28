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

func TestCreateCollectionUseCase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("success - create collection without group", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
		mockAuthService := mocks.NewMockAuthService(mockCtrl)

		useCase := NewCreateCollectionUseCase(mockCollectionRepo, mockAuthService)
		userID := entities.NewUserID()

		req := CreateCollectionRequest{
			UserID:     userID,
			GroupID:    nil,
			Name:       "My Food Collection",
			ObjectType: entities.ObjectTypeFood,
			Tags:       []string{"pantry", "kitchen"},
			Location:   "Kitchen",
			UserToken:  "test-token",
		}

		// Expect repository to save collection
		mockCollectionRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.NotNil(t, resp.Collection)
		assert.Equal(t, "My Food Collection", resp.Collection.Name().String())
		assert.Equal(t, entities.ObjectTypeFood, resp.Collection.ObjectType())
		assert.Equal(t, userID, resp.Collection.UserID())
		assert.Nil(t, resp.Collection.GroupID())
	})

	t.Run("success - create collection with group", func(t *testing.T) {
		userID := entities.NewUserID()
		groupID := entities.NewGroupID()

		// Create test group
		groupName, _ := entities.NewGroupName("Test Group")
		testGroup, _ := entities.NewGroup(entities.GroupProps{
			Name: groupName,
		})
		testGroup = entities.ReconstructGroup(groupID, groupName, time.Now(), time.Now())

		req := CreateCollectionRequest{
			UserID:     userID,
			GroupID:    &groupID,
			Name:       "Shared Collection",
			ObjectType: entities.ObjectTypeGeneral,
			Tags:       []string{"shared"},
			Location:   "Office",
			UserToken:  "test-token",
		}

		// Expect auth service to return user's groups
		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), "test-token", userID.String()).
			Return([]*entities.Group{testGroup}, nil).
			Times(1)

		// Expect repository to save collection
		mockCollectionRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, collection *entities.Collection) error {
				// Verify the group ID was set
				require.NotNil(t, collection.GroupID())
				return nil
			}).
			Times(1)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.NotNil(t, resp.Collection)
	})

	t.Run("error - invalid collection name", func(t *testing.T) {
		userID := entities.NewUserID()

		req := CreateCollectionRequest{
			UserID:     userID,
			GroupID:    nil,
			Name:       "", // Empty name is invalid
			ObjectType: entities.ObjectTypeGeneral,
			Tags:       []string{},
			Location:   "",
			UserToken:  "test-token",
		}

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid collection name")
	})

	t.Run("error - user not member of group", func(t *testing.T) {
		userID := entities.NewUserID()
		groupID := entities.NewGroupID()
		differentGroupID := entities.NewGroupID()

		// Create test group with different ID
		groupName, _ := entities.NewGroupName("Different Group")
		testGroup := entities.ReconstructGroup(
			differentGroupID,
			groupName,
			time.Now(),
			time.Now(),
		)

		req := CreateCollectionRequest{
			UserID:     userID,
			GroupID:    &groupID, // Requesting access to groupID
			Name:       "Shared Collection",
			ObjectType: entities.ObjectTypeGeneral,
			Tags:       []string{},
			Location:   "",
			UserToken:  "test-token",
		}

		// Auth service returns different group (user is not member of requested group)
		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), "test-token", userID.String()).
			Return([]*entities.Group{testGroup}, nil).
			Times(1)

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "user is not a member of the group")
	})

	t.Run("error - auth service failure", func(t *testing.T) {
		userID := entities.NewUserID()
		groupID := entities.NewGroupID()

		req := CreateCollectionRequest{
			UserID:     userID,
			GroupID:    &groupID,
			Name:       "Shared Collection",
			ObjectType: entities.ObjectTypeGeneral,
			Tags:       []string{},
			Location:   "",
			UserToken:  "test-token",
		}

		// Auth service returns error
		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), "test-token", userID.String()).
			Return(nil, errors.New("auth service unavailable")).
			Times(1)

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to get user groups")
	})

	t.Run("error - repository failure", func(t *testing.T) {
		userID := entities.NewUserID()

		req := CreateCollectionRequest{
			UserID:     userID,
			GroupID:    nil,
			Name:       "My Collection",
			ObjectType: entities.ObjectTypeGeneral,
			Tags:       []string{},
			Location:   "",
			UserToken:  "test-token",
		}

		// Repository returns error
		mockCollectionRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(errors.New("database connection failed")).
			Times(1)

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to save collection")
	})

	t.Run("success - create collection with all object types", func(t *testing.T) {
		userID := entities.NewUserID()
		objectTypes := []entities.ObjectType{
			entities.ObjectTypeFood,
			entities.ObjectTypeBook,
			entities.ObjectTypeVideoGame,
			entities.ObjectTypeMusic,
			entities.ObjectTypeBoardGame,
			entities.ObjectTypeGeneral,
		}

		for _, objType := range objectTypes {
			t.Run(string(objType), func(t *testing.T) {
				req := CreateCollectionRequest{
					UserID:     userID,
					GroupID:    nil,
					Name:       fake.ProductName(),
					ObjectType: objType,
					Tags:       []string{},
					Location:   "",
					UserToken:  "test-token",
				}

				mockCollectionRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)

				resp, err := useCase.Execute(context.Background(), req)

				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, objType, resp.Collection.ObjectType())
			})
		}
	})
}
