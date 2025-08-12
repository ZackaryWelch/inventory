package controllers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/nishiki/backend-go/app/http/request"
	"github.com/nishiki/backend-go/app/http/response"
	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/usecases"
	"github.com/nishiki/backend-go/mocks"
)

func TestObjectController_CreateObject(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Mock repositories and services
	mockContainerRepo := mocks.NewMockContainerRepository(mockCtrl)
	mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
	mockAuthService := mocks.NewMockAuthService(mockCtrl)

	// Create use cases with mocked dependencies
	createObjectUC := usecases.NewCreateObjectUseCase(mockContainerRepo, mockCollectionRepo, mockAuthService)
	updateObjectUC := usecases.NewUpdateObjectUseCase(mockContainerRepo, mockCollectionRepo, mockAuthService)
	deleteObjectUC := usecases.NewDeleteObjectUseCase(mockContainerRepo, mockCollectionRepo, mockAuthService)
	getCollectionObjectsUC := usecases.NewGetCollectionObjectsUseCase(mockCollectionRepo, mockContainerRepo, mockAuthService)
	bulkImportUC := usecases.NewBulkImportObjectsUseCase(mockContainerRepo, mockCollectionRepo, mockAuthService)
	bulkImportCollectionUC := usecases.NewBulkImportCollectionUseCase(mockCollectionRepo, mockContainerRepo, mockAuthService)

	controller := &ObjectController{
		createObjectUC:         createObjectUC,
		updateObjectUC:         updateObjectUC,
		deleteObjectUC:         deleteObjectUC,
		getCollectionObjectsUC: getCollectionObjectsUC,
		bulkImportUC:           bulkImportUC,
		bulkImportCollectionUC: bulkImportCollectionUC,
		logger:                 logger,
	}

	t.Run("success - create object", func(t *testing.T) {
		// Create test data
		testUser := randomUser()
		containerID := entities.NewContainerID()

		requestBody := request.CreateObjectRequest{
			ContainerID: containerID.String(),
			Name:        "Test Object",
			ObjectType:  "general",
			Properties:  map[string]interface{}{"description": "Test description"},
			Tags:        []string{"test", "example"},
		}

		objectName, _ := entities.NewObjectName("Test Object")
		testObject, _ := entities.NewObject(entities.ObjectProps{
			Name:       objectName,
			ObjectType: entities.ObjectTypeGeneral,
			Properties: requestBody.Properties,
			Tags:       requestBody.Tags,
		})

		// Create test container for mocking
		containerName, _ := entities.NewContainerName("Test Container")
		testContainer, _ := entities.NewContainer(entities.ContainerProps{
			CollectionID: entities.NewCollectionID(),
			Name:         containerName,
		})

		// Mock expectations
		mockAuthService.EXPECT().ValidateToken(gomock.Any(), "test-token").Return(nil, nil)
		mockContainerRepo.EXPECT().GetByID(gomock.Any(), containerID).Return(testContainer, nil)
		mockContainerRepo.EXPECT().Update(gomock.Any(), testContainer).Return(nil)

		// Create request
		req := newTestRequest(http.MethodPost, "/objects", requestBody)
		rr, ctx := createReq(req, logger)

		// Set authenticated user and token
		ctx.Set("auth_user", testUser)
		ctx.Set("auth_token", "test-token")
		ctx.Params = []gin.Param{
			{Key: "account_id", Value: testUser.ID().String()},
			{Key: "container_id", Value: containerID.String()},
		}

		// Call controller
		controller.CreateObject(ctx)

		// Assert response
		assert.Equal(t, http.StatusCreated, rr.Code)

		var resp response.ObjectResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, testObject.ID().String(), resp.ID)
		assert.Equal(t, "Test Object", resp.Name)
	})

	t.Run("error - invalid request body", func(t *testing.T) {
		testUser := randomUser()
		containerID := entities.NewContainerID()

		// Invalid request body (missing required fields)
		requestBody := map[string]interface{}{
			"name": "", // Empty name
		}

		req := newTestRequest(http.MethodPost, "/objects", requestBody)
		rr, ctx := createReq(req, logger)

		ctx.Set("auth_user", testUser)
		ctx.Set("auth_token", "test-token")
		ctx.Params = []gin.Param{
			{Key: "account_id", Value: testUser.ID().String()},
			{Key: "container_id", Value: containerID.String()},
		}

		controller.CreateObject(ctx)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("error - access denied", func(t *testing.T) {
		testUser := randomUser()
		containerID := entities.NewContainerID()

		requestBody := request.CreateObjectRequest{
			ContainerID: containerID.String(),
			Name:        "Test Object",
			ObjectType:  "general",
		}

		// Simulate access denied
		mockAuthService.EXPECT().ValidateToken(gomock.Any(), "test-token").Return(nil, nil)
		mockContainerRepo.EXPECT().GetByID(gomock.Any(), containerID).Return(nil, errors.New("access denied"))

		req := newTestRequest(http.MethodPost, "/objects", requestBody)
		rr, ctx := createReq(req, logger)

		ctx.Set("auth_user", testUser)
		ctx.Set("auth_token", "test-token")
		ctx.Params = []gin.Param{
			{Key: "account_id", Value: testUser.ID().String()},
			{Key: "container_id", Value: containerID.String()},
		}

		controller.CreateObject(ctx)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("error - container not found", func(t *testing.T) {
		testUser := randomUser()
		containerID := entities.NewContainerID()

		requestBody := request.CreateObjectRequest{
			ContainerID: containerID.String(),
			Name:        "Test Object",
			ObjectType:  "general",
		}

		// Simulate container not found
		mockAuthService.EXPECT().ValidateToken(gomock.Any(), "test-token").Return(nil, nil)
		mockContainerRepo.EXPECT().GetByID(gomock.Any(), containerID).Return(nil, errors.New("container not found"))

		req := newTestRequest(http.MethodPost, "/objects", requestBody)
		rr, ctx := createReq(req, logger)

		ctx.Set("auth_user", testUser)
		ctx.Set("auth_token", "test-token")
		ctx.Params = []gin.Param{
			{Key: "account_id", Value: testUser.ID().String()},
			{Key: "container_id", Value: containerID.String()},
		}

		controller.CreateObject(ctx)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestObjectController_DeleteObject(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Mock repositories and services
	mockContainerRepo := mocks.NewMockContainerRepository(mockCtrl)
	mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
	mockAuthService := mocks.NewMockAuthService(mockCtrl)

	// Create use case with mocked dependencies
	deleteObjectUC := usecases.NewDeleteObjectUseCase(mockContainerRepo, mockCollectionRepo, mockAuthService)

	controller := &ObjectController{
		deleteObjectUC: deleteObjectUC,
		logger:         logger,
	}

	t.Run("success - delete object", func(t *testing.T) {
		testUser := randomUser()
		objectID := entities.NewObjectID()

		// Create test container for mocking
		containerName, _ := entities.NewContainerName("Test Container")
		testContainer, _ := entities.NewContainer(entities.ContainerProps{
			CollectionID: entities.NewCollectionID(),
			Name:         containerName,
		})

		// Mock expectations
		mockAuthService.EXPECT().ValidateToken(gomock.Any(), "test-token").Return(nil, nil)
		mockContainerRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(testContainer, nil)
		mockContainerRepo.EXPECT().Update(gomock.Any(), testContainer).Return(nil)

		req := newTestRequest(http.MethodDelete, "/objects/"+objectID.String(), nil)
		rr, ctx := createReq(req, logger)

		ctx.Set("auth_user", testUser)
		ctx.Set("auth_token", "test-token")
		ctx.Params = []gin.Param{
			{Key: "account_id", Value: testUser.ID().String()},
			{Key: "object_id", Value: objectID.String()},
		}

		controller.DeleteObject(ctx)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp response.DeleteObjectResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
	})

	t.Run("error - invalid object ID", func(t *testing.T) {
		testUser := randomUser()

		req := newTestRequest(http.MethodDelete, "/objects/invalid-id", nil)
		rr, ctx := createReq(req, logger)

		ctx.Set("auth_user", testUser)
		ctx.Set("auth_token", "test-token")
		ctx.Params = []gin.Param{
			{Key: "account_id", Value: testUser.ID().String()},
			{Key: "object_id", Value: "invalid-id"},
		}

		controller.DeleteObject(ctx)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("error - object not found", func(t *testing.T) {
		testUser := randomUser()
		objectID := entities.NewObjectID()

		// Simulate object not found
		mockAuthService.EXPECT().ValidateToken(gomock.Any(), "test-token").Return(nil, nil)
		mockContainerRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(nil, errors.New("object not found"))

		req := newTestRequest(http.MethodDelete, "/objects/"+objectID.String(), nil)
		rr, ctx := createReq(req, logger)

		ctx.Set("auth_user", testUser)
		ctx.Set("auth_token", "test-token")
		ctx.Params = []gin.Param{
			{Key: "account_id", Value: testUser.ID().String()},
			{Key: "object_id", Value: objectID.String()},
		}

		controller.DeleteObject(ctx)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestObjectController_BulkImport(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Mock repositories and services
	mockContainerRepo := mocks.NewMockContainerRepository(mockCtrl)
	mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
	mockAuthService := mocks.NewMockAuthService(mockCtrl)

	// Create use case with mocked dependencies
	bulkImportUC := usecases.NewBulkImportObjectsUseCase(mockContainerRepo, mockCollectionRepo, mockAuthService)

	controller := &ObjectController{
		bulkImportUC: bulkImportUC,
		logger:       logger,
	}

	t.Run("success - bulk import objects", func(t *testing.T) {
		testUser := randomUser()
		containerID := entities.NewContainerID()

		requestBody := request.BulkImportRequest{
			Format:     "json",
			ObjectType: "general",
			Data: []map[string]interface{}{
				{"name": "Object 1", "description": "First object"},
				{"name": "Object 2", "description": "Second object"},
			},
			DefaultTags: []string{"imported"},
		}

		// Create test container for mocking
		containerName, _ := entities.NewContainerName("Test Container")
		testContainer, _ := entities.NewContainer(entities.ContainerProps{
			CollectionID: entities.NewCollectionID(),
			Name:         containerName,
		})

		// Mock expectations
		mockAuthService.EXPECT().ValidateToken(gomock.Any(), "test-token").Return(nil, nil)
		mockContainerRepo.EXPECT().GetByID(gomock.Any(), containerID).Return(testContainer, nil)
		mockContainerRepo.EXPECT().Update(gomock.Any(), testContainer).Return(nil).Times(2)

		req := newTestRequest(http.MethodPost, "/objects/bulk-import", requestBody)
		rr, ctx := createReq(req, logger)

		ctx.Set("auth_user", testUser)
		ctx.Set("auth_token", "test-token")
		ctx.Params = []gin.Param{
			{Key: "account_id", Value: testUser.ID().String()},
			{Key: "container_id", Value: containerID.String()},
		}

		controller.BulkImport(ctx)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp response.BulkImportResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, 2, resp.Imported)
		assert.Equal(t, 0, resp.Failed)
		assert.Equal(t, 2, resp.Total)
	})

	t.Run("error - validation failed", func(t *testing.T) {
		testUser := randomUser()
		containerID := entities.NewContainerID()

		// Invalid request (empty data)
		requestBody := request.BulkImportRequest{
			Format:     "json",
			ObjectType: "general",
			Data:       []map[string]interface{}{}, // Empty data
		}

		req := newTestRequest(http.MethodPost, "/objects/bulk-import", requestBody)
		rr, ctx := createReq(req, logger)

		ctx.Set("auth_user", testUser)
		ctx.Set("auth_token", "test-token")
		ctx.Params = []gin.Param{
			{Key: "account_id", Value: testUser.ID().String()},
			{Key: "container_id", Value: containerID.String()},
		}

		controller.BulkImport(ctx)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

