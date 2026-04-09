package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/nishiki/backend/domain/entities"
	"github.com/nishiki/backend/mocks"
)

// buildMultiContainerCollection creates a collection with two containers, each holding
// distinct objects.  Returns the collection plus the IDs of the two containers.
func buildMultiContainerCollection(
	userID entities.UserID,
	objs1, objs2 []entities.Object,
) (*entities.Collection, entities.ContainerID, entities.ContainerID) {
	collectionID := entities.NewCollectionID()
	collectionName, _ := entities.NewCollectionName("Test Collection")

	cid1 := entities.NewContainerID()
	name1, _ := entities.NewContainerName("Container A")
	c1 := entities.ReconstructContainer(
		cid1, collectionID, name1, entities.ContainerTypeGeneral,
		nil, nil, nil, objs1, "", nil, nil, nil, nil,
		time.Now(), time.Now(),
	)

	cid2 := entities.NewContainerID()
	name2, _ := entities.NewContainerName("Container B")
	c2 := entities.ReconstructContainer(
		cid2, collectionID, name2, entities.ContainerTypeGeneral,
		nil, nil, nil, objs2, "", nil, nil, nil, nil,
		time.Now(), time.Now(),
	)

	col := entities.ReconstructCollection(
		collectionID, userID, nil, collectionName, nil,
		entities.ObjectTypeGeneral,
		[]entities.Container{*c1, *c2},
		[]string{}, "", nil,
		time.Now(), time.Now(),
	)
	return col, cid1, cid2
}

// newSimpleObject builds a minimal Object with the given name and optional tags/properties.
func newSimpleObject(name string, tags []string, props map[string]entities.TypedValue) entities.Object {
	objName, _ := entities.NewObjectName(name)
	obj := entities.ReconstructObject(
		entities.NewObjectID(), objName, entities.NewObjectDescription(""),
		entities.ObjectTypeGeneral, "", nil, "", props, tags, nil,
		time.Now(), time.Now(),
	)
	return *obj
}

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

	obj1 := newSimpleObject("Apple Juice", []string{"food", "beverage"}, map[string]entities.TypedValue{"brand": {Val: "Tropicana"}, "for_sale": {Val: "true"}})
	obj2 := newSimpleObject("Banana Smoothie", []string{"food"}, map[string]entities.TypedValue{"brand": {Val: "Dole"}})
	obj3 := newSimpleObject("Code Book", []string{"book"}, map[string]entities.TypedValue{"author": {Val: "Clean Coder"}})

	collection, cid1, _ := buildMultiContainerCollection(userID, []entities.Object{obj1, obj2}, []entities.Object{obj3})

	// Build container pointers for mock returns
	containerSlice := collection.Containers()
	containerPtrs := make([]*entities.Container, len(containerSlice))
	for i := range containerSlice {
		containerPtrs[i] = &containerSlice[i]
	}

	// Full setup: auth + single aggregation query
	setupMocks := func() {
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "tok", userID.String()).Return(userGroups, nil)
		mockContainerRepo.EXPECT().GetByCollectionIDWithAccess(gomock.Any(), collection.ID(), userID, gomock.Any()).Return(containerPtrs, nil)
	}

	t.Run("no filters returns all objects", func(t *testing.T) {
		setupMocks()
		resp, err := uc.Execute(context.Background(), GetCollectionObjectsRequest{
			CollectionID: collection.ID(),
			UserID:       userID,
			UserToken:    "tok",
		})
		require.NoError(t, err)
		assert.Len(t, resp.Objects, 3)
	})

	t.Run("query filter matches name substring case-insensitively", func(t *testing.T) {
		setupMocks()
		resp, err := uc.Execute(context.Background(), GetCollectionObjectsRequest{
			CollectionID: collection.ID(),
			UserID:       userID,
			UserToken:    "tok",
			Query:        "juice",
		})
		require.NoError(t, err)
		require.Len(t, resp.Objects, 1)
		assert.Equal(t, "Apple Juice", resp.Objects[0].Object.Name().String())
	})

	t.Run("query filter returns empty when no match", func(t *testing.T) {
		setupMocks()
		resp, err := uc.Execute(context.Background(), GetCollectionObjectsRequest{
			CollectionID: collection.ID(),
			UserID:       userID,
			UserToken:    "tok",
			Query:        "zzznomatch",
		})
		require.NoError(t, err)
		assert.Empty(t, resp.Objects)
	})

	t.Run("tag filter matches single tag", func(t *testing.T) {
		setupMocks()
		resp, err := uc.Execute(context.Background(), GetCollectionObjectsRequest{
			CollectionID: collection.ID(),
			UserID:       userID,
			UserToken:    "tok",
			Tags:         []string{"beverage"},
		})
		require.NoError(t, err)
		require.Len(t, resp.Objects, 1)
		assert.Equal(t, "Apple Juice", resp.Objects[0].Object.Name().String())
	})

	t.Run("tag filter requires all tags", func(t *testing.T) {
		setupMocks()
		resp, err := uc.Execute(context.Background(), GetCollectionObjectsRequest{
			CollectionID: collection.ID(),
			UserID:       userID,
			UserToken:    "tok",
			Tags:         []string{"food", "beverage"},
		})
		require.NoError(t, err)
		require.Len(t, resp.Objects, 1)
		assert.Equal(t, "Apple Juice", resp.Objects[0].Object.Name().String())
	})

	t.Run("container_id filter restricts to that container", func(t *testing.T) {
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "tok", userID.String()).Return(userGroups, nil)
		mockContainerRepo.EXPECT().GetByID(gomock.Any(), cid1).Return(containerPtrs[0], nil)
		resp, err := uc.Execute(context.Background(), GetCollectionObjectsRequest{
			CollectionID: collection.ID(),
			UserID:       userID,
			UserToken:    "tok",
			ContainerID:  &cid1,
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
			CollectionID:    collection.ID(),
			UserID:          userID,
			UserToken:       "tok",
			PropertyFilters: map[string]string{"brand": "trop"},
		})
		require.NoError(t, err)
		require.Len(t, resp.Objects, 1)
		assert.Equal(t, "Apple Juice", resp.Objects[0].Object.Name().String())
	})

	t.Run("property filter returns empty when key absent", func(t *testing.T) {
		setupMocks()
		resp, err := uc.Execute(context.Background(), GetCollectionObjectsRequest{
			CollectionID:    collection.ID(),
			UserID:          userID,
			UserToken:       "tok",
			PropertyFilters: map[string]string{"nonexistent_key": "value"},
		})
		require.NoError(t, err)
		assert.Empty(t, resp.Objects)
	})

	t.Run("combined query and tag filter", func(t *testing.T) {
		setupMocks()
		resp, err := uc.Execute(context.Background(), GetCollectionObjectsRequest{
			CollectionID: collection.ID(),
			UserID:       userID,
			UserToken:    "tok",
			Query:        "banana",
			Tags:         []string{"food"},
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

	collection, _, _ := buildMultiContainerCollection(ownerID, nil, nil)

	t.Run("access denied for non-owner without group returns empty", func(t *testing.T) {
		// With $lookup access check, unauthorized users get empty results (no containers match)
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "tok", otherID.String()).Return([]*entities.Group{}, nil)
		mockContainerRepo.EXPECT().GetByCollectionIDWithAccess(gomock.Any(), collection.ID(), otherID, gomock.Any()).Return(nil, nil)

		resp, err := uc.Execute(context.Background(), GetCollectionObjectsRequest{
			CollectionID: collection.ID(),
			UserID:       otherID,
			UserToken:    "tok",
		})
		require.NoError(t, err)
		assert.Empty(t, resp.Objects)
	})

	t.Run("collection not found returns empty", func(t *testing.T) {
		// With $lookup, non-existent collection also returns empty (no containers match)
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "tok", ownerID.String()).Return([]*entities.Group{}, nil)
		mockContainerRepo.EXPECT().GetByCollectionIDWithAccess(gomock.Any(), collection.ID(), ownerID, gomock.Any()).Return(nil, nil)

		resp, err := uc.Execute(context.Background(), GetCollectionObjectsRequest{
			CollectionID: collection.ID(),
			UserID:       ownerID,
			UserToken:    "tok",
		})
		require.NoError(t, err)
		assert.Empty(t, resp.Objects)
	})
}
