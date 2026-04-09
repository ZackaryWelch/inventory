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
			UserID: userID, Name: "My Food Collection", ObjectType: entities.ObjectTypeFood,
			Tags: []string{"pantry", "kitchen"}, Location: "Kitchen", UserToken: "test-token",
		}

		mockCollectionRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "My Food Collection", resp.Collection.Name().String())
		assert.Equal(t, entities.ObjectTypeFood, resp.Collection.ObjectType())
		assert.Equal(t, userID, resp.Collection.UserID())
		assert.Nil(t, resp.Collection.GroupID())
	})

	t.Run("success - create collection with group", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
		mockAuthService := mocks.NewMockAuthService(mockCtrl)
		useCase := NewCreateCollectionUseCase(mockCollectionRepo, mockAuthService)

		userID := entities.NewUserID()
		groupID := entities.NewGroupID()
		testGroup := NewTestGroup(GrpID(groupID))

		req := CreateCollectionRequest{
			UserID: userID, GroupID: &groupID, Name: "Shared Collection",
			ObjectType: entities.ObjectTypeGeneral, Tags: []string{"shared"}, Location: "Office", UserToken: "test-token",
		}

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", userID.String()).Return([]*entities.Group{testGroup}, nil)
		mockCollectionRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, collection *entities.Collection) error {
			require.NotNil(t, collection.GroupID())
			return nil
		})

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.NotNil(t, resp.Collection)
	})

	t.Run("error - invalid collection name", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
		mockAuthService := mocks.NewMockAuthService(mockCtrl)
		useCase := NewCreateCollectionUseCase(mockCollectionRepo, mockAuthService)

		resp, err := useCase.Execute(context.Background(), CreateCollectionRequest{
			UserID: entities.NewUserID(), Name: "", ObjectType: entities.ObjectTypeGeneral, UserToken: "test-token",
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid collection name")
	})

	t.Run("error - user not member of group", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
		mockAuthService := mocks.NewMockAuthService(mockCtrl)
		useCase := NewCreateCollectionUseCase(mockCollectionRepo, mockAuthService)

		userID := entities.NewUserID()
		groupID, _ := entities.GroupIDFromString("group-123")
		differentGroupID, _ := entities.GroupIDFromString("group-456")

		differentGroup := NewTestGroup(GrpID(differentGroupID), GrpName("Different Group"))

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", userID.String()).Return([]*entities.Group{differentGroup}, nil)

		resp, err := useCase.Execute(context.Background(), CreateCollectionRequest{
			UserID: userID, GroupID: &groupID, Name: "Shared Collection",
			ObjectType: entities.ObjectTypeGeneral, UserToken: "test-token",
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "user is not a member of the group")
	})

	t.Run("error - auth service failure", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
		mockAuthService := mocks.NewMockAuthService(mockCtrl)
		useCase := NewCreateCollectionUseCase(mockCollectionRepo, mockAuthService)

		userID := entities.NewUserID()
		groupID := entities.NewGroupID()

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", userID.String()).Return(nil, errors.New("auth service unavailable"))

		resp, err := useCase.Execute(context.Background(), CreateCollectionRequest{
			UserID: userID, GroupID: &groupID, Name: "Shared Collection",
			ObjectType: entities.ObjectTypeGeneral, UserToken: "test-token",
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to get user groups")
	})

	t.Run("error - repository failure", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
		mockAuthService := mocks.NewMockAuthService(mockCtrl)
		useCase := NewCreateCollectionUseCase(mockCollectionRepo, mockAuthService)

		mockCollectionRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("database connection failed"))

		resp, err := useCase.Execute(context.Background(), CreateCollectionRequest{
			UserID: entities.NewUserID(), Name: "My Collection",
			ObjectType: entities.ObjectTypeGeneral, UserToken: "test-token",
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to save collection")
	})

	t.Run("success - create collection with all object types", func(t *testing.T) {
		userID := entities.NewUserID()
		for _, objType := range entities.AllObjectTypes {
			t.Run(string(objType), func(t *testing.T) {
				mockCtrl := gomock.NewController(t)
				defer mockCtrl.Finish()

				mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
				mockAuthService := mocks.NewMockAuthService(mockCtrl)
				useCase := NewCreateCollectionUseCase(mockCollectionRepo, mockAuthService)

				mockCollectionRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

				resp, err := useCase.Execute(context.Background(), CreateCollectionRequest{
					UserID: userID, Name: fake.ProductName(), ObjectType: objType, UserToken: "test-token",
				})

				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, objType, resp.Collection.ObjectType())
			})
		}
	})
}
