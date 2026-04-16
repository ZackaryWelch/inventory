package usecases

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/nishiki/backend/domain/entities"
	"github.com/nishiki/backend/mocks"
)

func parseCSV(t *testing.T, data []byte) [][]string {
	t.Helper()
	r := csv.NewReader(bytes.NewReader(data))
	rows, err := r.ReadAll()
	require.NoError(t, err)
	return rows
}

// collectionWithObjects creates a collection owned by userID with one container holding objects.
func collectionWithObjects(userID entities.UserID, groupID *entities.GroupID, schema *entities.PropertySchema, objects []entities.Object) *entities.Collection {
	collectionID := entities.NewCollectionID()
	ctr := NewTestContainer(CtrCollectionID(collectionID), CtrObjects(objects...))
	return NewTestCollection(ColID(collectionID), ColUserID(userID), ColGroupID(groupID), ColSchema(schema), ColContainers(*ctr))
}

func TestExportCollectionUseCase_Execute(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
	mockAuthService := mocks.NewMockAuthService(mockCtrl)

	useCase := NewExportCollectionUseCase(mockCollectionRepo, mockAuthService)

	t.Run("success - owner, no schema, properties sorted alphabetically", func(t *testing.T) {
		userID := entities.NewUserID()

		qty := 3.0
		exp := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
		obj := NewTestObject(
			ObjName("Widget A"), ObjDesc("A widget"),
			ObjQuantity(qty), ObjUnit("pcs"), ObjExpiresAt(exp),
			ObjProps(Props("sku", "SKU-001", "color", "red")),
			ObjTags("sale", "new"),
		)

		collection := collectionWithObjects(userID, nil, nil, []entities.Object{*obj})

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "tok", userID.String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collection.ID()).Return(collection, nil)

		resp, err := useCase.Execute(context.Background(), ExportCollectionRequest{
			CollectionID: collection.ID(), UserID: userID, UserToken: "tok",
		})

		require.NoError(t, err)
		rows := parseCSV(t, resp.CSV)
		require.Len(t, rows, 2)
		assert.Equal(t, []string{"Name", "Description", "Quantity", "Unit", "Tags", "Expires At", "color", "sku"}, rows[0])
		assert.Equal(t, "Widget A", rows[1][0])
		assert.Equal(t, "A widget", rows[1][1])
		assert.Equal(t, "3", rows[1][2])
		assert.Equal(t, "pcs", rows[1][3])
		assert.Equal(t, "sale|new", rows[1][4])
		assert.Equal(t, exp.Format(time.RFC3339), rows[1][5])
		assert.Equal(t, "red", rows[1][6])
		assert.Equal(t, "SKU-001", rows[1][7])
	})

	t.Run("success - schema drives property column order and display names", func(t *testing.T) {
		userID := entities.NewUserID()

		schema := &entities.PropertySchema{
			Definitions: []entities.PropertyDefinition{
				{Key: "sku", DisplayName: "SKU", Type: entities.PropertyTypeText},
				{Key: "price", DisplayName: "Price (USD)", Type: entities.PropertyTypeCurrency},
			},
		}

		obj := NewTestObject(ObjName("Gadget"), ObjProps(Props("sku", "G-42", "price", entities.TypedValue{Val: 9.99})))
		collection := collectionWithObjects(userID, nil, schema, []entities.Object{*obj})

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "tok", userID.String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collection.ID()).Return(collection, nil)

		resp, err := useCase.Execute(context.Background(), ExportCollectionRequest{
			CollectionID: collection.ID(), UserID: userID, UserToken: "tok",
		})

		require.NoError(t, err)
		rows := parseCSV(t, resp.CSV)
		require.Len(t, rows, 2)
		assert.Equal(t, []string{"Name", "Description", "Quantity", "Unit", "Tags", "Expires At", "SKU", "Price (USD)"}, rows[0])
		assert.Equal(t, "Gadget", rows[1][0])
		assert.Equal(t, "G-42", rows[1][6])
		assert.Equal(t, "9.99", rows[1][7])
	})

	t.Run("success - schema overrides display name for fixed field", func(t *testing.T) {
		userID := entities.NewUserID()

		schema := &entities.PropertySchema{
			Definitions: []entities.PropertyDefinition{
				{Key: "name", DisplayName: "Title", Type: entities.PropertyTypeText},
			},
		}

		obj := NewTestObject(ObjName("Go Programming"))
		collection := collectionWithObjects(userID, nil, schema, []entities.Object{*obj})

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "tok", userID.String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collection.ID()).Return(collection, nil)

		resp, err := useCase.Execute(context.Background(), ExportCollectionRequest{
			CollectionID: collection.ID(), UserID: userID, UserToken: "tok",
		})

		require.NoError(t, err)
		rows := parseCSV(t, resp.CSV)
		require.Len(t, rows, 2)
		assert.Equal(t, "Title", rows[0][0])
		assert.Equal(t, "Go Programming", rows[1][0])
	})

	t.Run("success - group member access", func(t *testing.T) {
		memberID := entities.NewUserID()
		ownerID := entities.NewUserID()
		groupID := entities.NewGroupID()

		group := NewTestGroup(GrpID(groupID), GrpName("Team"))
		collection := collectionWithObjects(ownerID, &groupID, nil, []entities.Object{})

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "tok", memberID.String()).Return([]*entities.Group{group}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collection.ID()).Return(collection, nil)

		resp, err := useCase.Execute(context.Background(), ExportCollectionRequest{
			CollectionID: collection.ID(), UserID: memberID, UserToken: "tok",
		})

		require.NoError(t, err)
		rows := parseCSV(t, resp.CSV)
		require.Len(t, rows, 1)
		assert.Equal(t, "Name", rows[0][0])
	})

	t.Run("success - empty quantity and expires_at omitted", func(t *testing.T) {
		userID := entities.NewUserID()

		obj := NewTestObject(ObjName("Simple Item"))
		collection := collectionWithObjects(userID, nil, nil, []entities.Object{*obj})

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "tok", userID.String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collection.ID()).Return(collection, nil)

		resp, err := useCase.Execute(context.Background(), ExportCollectionRequest{
			CollectionID: collection.ID(), UserID: userID, UserToken: "tok",
		})

		require.NoError(t, err)
		rows := parseCSV(t, resp.CSV)
		require.Len(t, rows, 2)
		assert.Empty(t, rows[1][2]) // quantity
		assert.Empty(t, rows[1][5]) // expires_at
	})

	t.Run("error - auth service failure", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "tok", userID.String()).Return(nil, errors.New("auth unavailable"))

		resp, err := useCase.Execute(context.Background(), ExportCollectionRequest{
			CollectionID: collectionID, UserID: userID, UserToken: "tok",
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to get user groups")
	})

	t.Run("error - collection not found", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "tok", userID.String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(nil, errors.New("not found"))

		resp, err := useCase.Execute(context.Background(), ExportCollectionRequest{
			CollectionID: collectionID, UserID: userID, UserToken: "tok",
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "collection not found")
	})

	t.Run("error - access denied", func(t *testing.T) {
		userID := entities.NewUserID()
		ownerID := entities.NewUserID()

		collection := collectionWithObjects(ownerID, nil, nil, []entities.Object{})

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "tok", userID.String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collection.ID()).Return(collection, nil)

		resp, err := useCase.Execute(context.Background(), ExportCollectionRequest{
			CollectionID: collection.ID(), UserID: userID, UserToken: "tok",
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "access denied")
	})
}
