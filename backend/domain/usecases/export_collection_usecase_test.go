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

// parseCSV is a test helper that reads all rows from CSV bytes.
func parseCSV(t *testing.T, data []byte) [][]string {
	t.Helper()
	r := csv.NewReader(bytes.NewReader(data))
	rows, err := r.ReadAll()
	require.NoError(t, err)
	return rows
}

// buildCollection creates a collection with one container holding the given objects.
func buildCollection(userID entities.UserID, groupID *entities.GroupID, schema *entities.PropertySchema, objects []entities.Object) *entities.Collection {
	collectionID := entities.NewCollectionID()
	collectionName, _ := entities.NewCollectionName("Test Collection")
	containerName, _ := entities.NewContainerName("Test Container")
	container := entities.ReconstructContainer(
		entities.NewContainerID(),
		collectionID,
		containerName,
		entities.ContainerTypeGeneral,
		nil, nil, nil,
		objects,
		"",
		nil, nil, nil, nil,
		time.Now(), time.Now(),
	)
	return entities.ReconstructCollection(
		collectionID,
		userID,
		groupID,
		collectionName,
		nil,
		entities.ObjectTypeGeneral,
		[]entities.Container{*container},
		[]string{},
		"",
		schema,
		time.Now(), time.Now(),
	)
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
		objName, _ := entities.NewObjectName("Widget A")
		obj := entities.ReconstructObject(
			entities.NewObjectID(),
			objName,
			entities.NewObjectDescription("A widget"),
			entities.ObjectTypeGeneral,
			&qty,
			"pcs",
			map[string]interface{}{"sku": "SKU-001", "color": "red"},
			[]string{"sale", "new"},
			&exp,
			time.Now(), time.Now(),
		)

		collection := buildCollection(userID, nil, nil, []entities.Object{*obj})

		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), "tok", userID.String()).
			Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collection.ID()).
			Return(collection, nil)

		resp, err := useCase.Execute(context.Background(), ExportCollectionRequest{
			CollectionID: collection.ID(),
			UserID:       userID,
			UserToken:    "tok",
		})

		require.NoError(t, err)
		rows := parseCSV(t, resp.CSV)
		require.Len(t, rows, 2)

		// Fixed headers auto-derived, then property keys alpha-sorted
		assert.Equal(t, []string{"Name", "Description", "Quantity", "Unit", "Tags", "Expires At", "color", "sku"}, rows[0])

		// Data row
		assert.Equal(t, "Widget A", rows[1][0])
		assert.Equal(t, "A widget", rows[1][1])
		assert.Equal(t, "3", rows[1][2])
		assert.Equal(t, "pcs", rows[1][3])
		assert.Equal(t, "sale|new", rows[1][4])
		assert.Equal(t, exp.Format(time.RFC3339), rows[1][5])
		assert.Equal(t, "red", rows[1][6])     // color
		assert.Equal(t, "SKU-001", rows[1][7]) // sku
	})

	t.Run("success - schema drives property column order and display names", func(t *testing.T) {
		userID := entities.NewUserID()

		schema := &entities.PropertySchema{
			Definitions: []entities.PropertyDefinition{
				{Key: "sku", DisplayName: "SKU", Type: entities.PropertyTypeText},
				{Key: "price", DisplayName: "Price (USD)", Type: entities.PropertyTypeCurrency},
			},
		}

		objName, _ := entities.NewObjectName("Gadget")
		obj := entities.ReconstructObject(
			entities.NewObjectID(),
			objName,
			entities.NewObjectDescription(""),
			entities.ObjectTypeGeneral,
			nil, "",
			map[string]interface{}{"sku": "G-42", "price": 9.99},
			[]string{},
			nil,
			time.Now(), time.Now(),
		)

		collection := buildCollection(userID, nil, schema, []entities.Object{*obj})

		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), "tok", userID.String()).
			Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collection.ID()).
			Return(collection, nil)

		resp, err := useCase.Execute(context.Background(), ExportCollectionRequest{
			CollectionID: collection.ID(),
			UserID:       userID,
			UserToken:    "tok",
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

		objName, _ := entities.NewObjectName("Go Programming")
		obj := entities.ReconstructObject(
			entities.NewObjectID(),
			objName,
			entities.NewObjectDescription(""),
			entities.ObjectTypeGeneral,
			nil, "", map[string]interface{}{}, []string{}, nil,
			time.Now(), time.Now(),
		)

		collection := buildCollection(userID, nil, schema, []entities.Object{*obj})

		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), "tok", userID.String()).
			Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collection.ID()).
			Return(collection, nil)

		resp, err := useCase.Execute(context.Background(), ExportCollectionRequest{
			CollectionID: collection.ID(),
			UserID:       userID,
			UserToken:    "tok",
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

		groupName, _ := entities.NewGroupName("Team")
		group := entities.ReconstructGroup(groupID, groupName, entities.NewGroupDescription(""), time.Now(), time.Now())

		collection := buildCollection(ownerID, &groupID, nil, []entities.Object{})

		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), "tok", memberID.String()).
			Return([]*entities.Group{group}, nil)
		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collection.ID()).
			Return(collection, nil)

		resp, err := useCase.Execute(context.Background(), ExportCollectionRequest{
			CollectionID: collection.ID(),
			UserID:       memberID,
			UserToken:    "tok",
		})

		require.NoError(t, err)
		rows := parseCSV(t, resp.CSV)
		// Header only, no objects
		require.Len(t, rows, 1)
		assert.Equal(t, "Name", rows[0][0])
	})

	t.Run("success - empty quantity and expires_at omitted", func(t *testing.T) {
		userID := entities.NewUserID()

		objName, _ := entities.NewObjectName("Simple Item")
		obj := entities.ReconstructObject(
			entities.NewObjectID(),
			objName,
			entities.NewObjectDescription(""),
			entities.ObjectTypeGeneral,
			nil, "", map[string]interface{}{}, []string{}, nil,
			time.Now(), time.Now(),
		)

		collection := buildCollection(userID, nil, nil, []entities.Object{*obj})

		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), "tok", userID.String()).
			Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collection.ID()).
			Return(collection, nil)

		resp, err := useCase.Execute(context.Background(), ExportCollectionRequest{
			CollectionID: collection.ID(),
			UserID:       userID,
			UserToken:    "tok",
		})

		require.NoError(t, err)
		rows := parseCSV(t, resp.CSV)
		require.Len(t, rows, 2)
		assert.Equal(t, "", rows[1][2]) // quantity
		assert.Equal(t, "", rows[1][5]) // expires_at
	})

	t.Run("error - auth service failure", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), "tok", userID.String()).
			Return(nil, errors.New("auth unavailable"))

		resp, err := useCase.Execute(context.Background(), ExportCollectionRequest{
			CollectionID: collectionID,
			UserID:       userID,
			UserToken:    "tok",
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to get user groups")
	})

	t.Run("error - collection not found", func(t *testing.T) {
		userID := entities.NewUserID()
		collectionID := entities.NewCollectionID()

		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), "tok", userID.String()).
			Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(nil, errors.New("not found"))

		resp, err := useCase.Execute(context.Background(), ExportCollectionRequest{
			CollectionID: collectionID,
			UserID:       userID,
			UserToken:    "tok",
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "collection not found")
	})

	t.Run("error - access denied", func(t *testing.T) {
		userID := entities.NewUserID()
		ownerID := entities.NewUserID()

		collection := buildCollection(ownerID, nil, nil, []entities.Object{})

		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), "tok", userID.String()).
			Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collection.ID()).
			Return(collection, nil)

		resp, err := useCase.Execute(context.Background(), ExportCollectionRequest{
			CollectionID: collection.ID(),
			UserID:       userID,
			UserToken:    "tok",
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "access denied")
	})
}
