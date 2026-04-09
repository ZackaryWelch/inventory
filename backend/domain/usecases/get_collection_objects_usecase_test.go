package usecases

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/nishiki/backend/domain/entities"
	"github.com/nishiki/backend/mocks"
)

func TestGetCollectionObjectsUseCase_Filters(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
	mockContainerRepo := mocks.NewMockContainerRepository(mockCtrl)
	mockAuthService := mocks.NewMockAuthService(mockCtrl)

	uc := NewGetCollectionObjectsUseCase(mockCollectionRepo, mockContainerRepo, mockAuthService)

	userID := entities.NewUserID()
	userGroups := []*entities.Group{}

	obj1 := *NewTestObject(ObjName("Apple Juice"), ObjTags("food", "beverage"), ObjProps(Props("brand", "Tropicana", "for_sale", "true")))
	obj2 := *NewTestObject(ObjName("Banana Smoothie"), ObjTags("food"), ObjProps(Props("brand", "Dole")))
	obj3 := *NewTestObject(ObjName("Code Book"), ObjTags("book"), ObjProps(Props("author", "Clean Coder")))

	collectionID := entities.NewCollectionID()
	cid1 := entities.NewContainerID()
	cid2 := entities.NewContainerID()

	c1 := NewTestContainer(CtrID(cid1), CtrCollectionID(collectionID), CtrName("Container A"), CtrObjects(obj1, obj2))
	c2 := NewTestContainer(CtrID(cid2), CtrCollectionID(collectionID), CtrName("Container B"), CtrObjects(obj3))

	collection := NewTestCollection(ColID(collectionID), ColUserID(userID), ColContainers(*c1, *c2))

	containerSlice := collection.Containers()
	containerPtrs := make([]*entities.Container, len(containerSlice))
	for i := range containerSlice {
		containerPtrs[i] = &containerSlice[i]
	}

	setupMocks := func() {
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "tok", userID.String()).Return(userGroups, nil)
		mockContainerRepo.EXPECT().GetByCollectionIDWithAccess(gomock.Any(), collection.ID(), userID, gomock.Any()).Return(containerPtrs, nil)
	}

	t.Run("no filters returns all objects", func(t *testing.T) {
		setupMocks()
		resp, err := uc.Execute(context.Background(), GetCollectionObjectsRequest{
			CollectionID: collection.ID(), UserID: userID, UserToken: "tok",
		})
		require.NoError(t, err)
		assert.Len(t, resp.Objects, 3)
	})

	t.Run("query filter matches name substring case-insensitively", func(t *testing.T) {
		setupMocks()
		resp, err := uc.Execute(context.Background(), GetCollectionObjectsRequest{
			CollectionID: collection.ID(), UserID: userID, UserToken: "tok", Query: "juice",
		})
		require.NoError(t, err)
		require.Len(t, resp.Objects, 1)
		assert.Equal(t, "Apple Juice", resp.Objects[0].Object.Name().String())
	})

	t.Run("query filter returns empty when no match", func(t *testing.T) {
		setupMocks()
		resp, err := uc.Execute(context.Background(), GetCollectionObjectsRequest{
			CollectionID: collection.ID(), UserID: userID, UserToken: "tok", Query: "zzznomatch",
		})
		require.NoError(t, err)
		assert.Empty(t, resp.Objects)
	})

	t.Run("tag filter matches single tag", func(t *testing.T) {
		setupMocks()
		resp, err := uc.Execute(context.Background(), GetCollectionObjectsRequest{
			CollectionID: collection.ID(), UserID: userID, UserToken: "tok", Tags: []string{"beverage"},
		})
		require.NoError(t, err)
		require.Len(t, resp.Objects, 1)
		assert.Equal(t, "Apple Juice", resp.Objects[0].Object.Name().String())
	})

	t.Run("tag filter requires all tags", func(t *testing.T) {
		setupMocks()
		resp, err := uc.Execute(context.Background(), GetCollectionObjectsRequest{
			CollectionID: collection.ID(), UserID: userID, UserToken: "tok", Tags: []string{"food", "beverage"},
		})
		require.NoError(t, err)
		require.Len(t, resp.Objects, 1)
		assert.Equal(t, "Apple Juice", resp.Objects[0].Object.Name().String())
	})

	t.Run("container_id filter restricts to that container", func(t *testing.T) {
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "tok", userID.String()).Return(userGroups, nil)
		mockContainerRepo.EXPECT().GetByID(gomock.Any(), cid1).Return(containerPtrs[0], nil)
		resp, err := uc.Execute(context.Background(), GetCollectionObjectsRequest{
			CollectionID: collection.ID(), UserID: userID, UserToken: "tok", ContainerID: &cid1,
		})
		require.NoError(t, err)
		assert.Len(t, resp.Objects, 2)
		names := []string{resp.Objects[0].Object.Name().String(), resp.Objects[1].Object.Name().String()}
		assert.Contains(t, names, "Apple Juice")
		assert.Contains(t, names, "Banana Smoothie")
	})

	t.Run("property filter matches value substring", func(t *testing.T) {
		setupMocks()
		resp, err := uc.Execute(context.Background(), GetCollectionObjectsRequest{
			CollectionID: collection.ID(), UserID: userID, UserToken: "tok",
			PropertyFilters: map[string]string{"brand": "trop"},
		})
		require.NoError(t, err)
		require.Len(t, resp.Objects, 1)
		assert.Equal(t, "Apple Juice", resp.Objects[0].Object.Name().String())
	})

	t.Run("property filter returns empty when key absent", func(t *testing.T) {
		setupMocks()
		resp, err := uc.Execute(context.Background(), GetCollectionObjectsRequest{
			CollectionID: collection.ID(), UserID: userID, UserToken: "tok",
			PropertyFilters: map[string]string{"nonexistent_key": "value"},
		})
		require.NoError(t, err)
		assert.Empty(t, resp.Objects)
	})

	t.Run("combined query and tag filter", func(t *testing.T) {
		setupMocks()
		resp, err := uc.Execute(context.Background(), GetCollectionObjectsRequest{
			CollectionID: collection.ID(), UserID: userID, UserToken: "tok",
			Query: "banana", Tags: []string{"food"},
		})
		require.NoError(t, err)
		require.Len(t, resp.Objects, 1)
		assert.Equal(t, "Banana Smoothie", resp.Objects[0].Object.Name().String())
	})
}

func TestGetCollectionObjectsUseCase_AccessControl(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
	mockContainerRepo := mocks.NewMockContainerRepository(mockCtrl)
	mockAuthService := mocks.NewMockAuthService(mockCtrl)

	uc := NewGetCollectionObjectsUseCase(mockCollectionRepo, mockContainerRepo, mockAuthService)

	ownerID := entities.NewUserID()
	otherID := entities.NewUserID()
	collection := NewTestCollection(ColUserID(ownerID))

	t.Run("access denied for non-owner without group returns empty", func(t *testing.T) {
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "tok", otherID.String()).Return([]*entities.Group{}, nil)
		mockContainerRepo.EXPECT().GetByCollectionIDWithAccess(gomock.Any(), collection.ID(), otherID, gomock.Any()).Return(nil, nil)

		resp, err := uc.Execute(context.Background(), GetCollectionObjectsRequest{
			CollectionID: collection.ID(), UserID: otherID, UserToken: "tok",
		})
		require.NoError(t, err)
		assert.Empty(t, resp.Objects)
	})

	t.Run("collection not found returns empty", func(t *testing.T) {
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "tok", ownerID.String()).Return([]*entities.Group{}, nil)
		mockContainerRepo.EXPECT().GetByCollectionIDWithAccess(gomock.Any(), collection.ID(), ownerID, gomock.Any()).Return(nil, nil)

		resp, err := uc.Execute(context.Background(), GetCollectionObjectsRequest{
			CollectionID: collection.ID(), UserID: ownerID, UserToken: "tok",
		})
		require.NoError(t, err)
		assert.Empty(t, resp.Objects)
	})
}
