package controllers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	fake "github.com/brianvoe/gofakeit/v7"
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

func TestCollectionController_CreateCollection(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Mock repositories and services
	mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
	mockAuthService := mocks.NewMockAuthService(mockCtrl)

	// Create real use cases with mocked dependencies
	createCollectionUC := usecases.NewCreateCollectionUseCase(mockCollectionRepo, mockAuthService)

	controller := &CollectionController{
		createCollectionUC: createCollectionUC,
		logger:             logger,
	}

	t.Run("success - create collection", func(t *testing.T) {
		testUser := randomUser()

		requestBody := request.CreateCollectionRequest{
			UserID:     testUser.ID().String(),
			Name:       "Test Collection",
			ObjectType: "general",
		}

		// Set up mock expectations
		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), "test-token", testUser.ID().String()).
			Return([]*entities.Group{}, nil).
			Times(1)

		mockCollectionRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

		req := newTestRequest(http.MethodPost, "/collections", requestBody)
		rr, ctx := createReq(req, logger)

		ctx.Set("auth_user", testUser)
		ctx.Set("auth_token", "test-token")
		ctx.Params = []gin.Param{
			{Key: "id", Value: testUser.ID().String()},
		}

		controller.CreateCollection(ctx)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var resp response.CollectionResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, "Test Collection", resp.Name)
		assert.Equal(t, "general", resp.ObjectType)
	})

	t.Run("error - invalid request body", func(t *testing.T) {
		testUser := randomUser()

		// Invalid request (empty name)
		requestBody := request.CreateCollectionRequest{
			UserID:     testUser.ID().String(),
			Name:       "",
			ObjectType: "general",
		}

		req := newTestRequest(http.MethodPost, "/collections", requestBody)
		rr, ctx := createReq(req, logger)

		ctx.Set("auth_user", testUser)
		ctx.Set("auth_token", "test-token")
		ctx.Params = []gin.Param{
			{Key: "id", Value: testUser.ID().String()},
		}

		controller.CreateCollection(ctx)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("error - database failure", func(t *testing.T) {
		testUser := randomUser()

		requestBody := request.CreateCollectionRequest{
			UserID:     testUser.ID().String(),
			Name:       "Test Collection",
			ObjectType: "general",
		}

		// Set up mock expectations for failure
		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), "test-token", testUser.ID().String()).
			Return([]*entities.Group{}, nil).
			Times(1)

		mockCollectionRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(errors.New("database connection failed")).
			Times(1)

		req := newTestRequest(http.MethodPost, "/collections", requestBody)
		rr, ctx := createReq(req, logger)

		ctx.Set("auth_user", testUser)
		ctx.Set("auth_token", "test-token")
		ctx.Params = []gin.Param{
			{Key: "id", Value: testUser.ID().String()},
		}

		controller.CreateCollection(ctx)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("error - auth service failure", func(t *testing.T) {
		testUser := randomUser()

		requestBody := request.CreateCollectionRequest{
			UserID:     testUser.ID().String(),
			Name:       "Test Collection",
			ObjectType: "general",
		}

		// Set up mock expectations for auth failure
		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), "test-token", testUser.ID().String()).
			Return(nil, errors.New("auth service unavailable")).
			Times(1)

		req := newTestRequest(http.MethodPost, "/collections", requestBody)
		rr, ctx := createReq(req, logger)

		ctx.Set("auth_user", testUser)
		ctx.Set("auth_token", "test-token")
		ctx.Params = []gin.Param{
			{Key: "id", Value: testUser.ID().String()},
		}

		controller.CreateCollection(ctx)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestCollectionController_GetCollections(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Mock repositories and services
	mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
	mockAuthService := mocks.NewMockAuthService(mockCtrl)

	// Create real use case with mocked dependencies
	getCollectionsUC := usecases.NewGetCollectionsUseCase(mockCollectionRepo, mockAuthService)

	controller := &CollectionController{
		getCollectionsUC: getCollectionsUC,
		logger:           logger,
	}

	t.Run("success - get collections", func(t *testing.T) {
		testUser := randomUser()

		// Create test collections using ReconstructCollection
		collections := []*entities.Collection{}
		for i := 0; i < 2; i++ {
			collectionName, _ := entities.NewCollectionName(fake.Company())
			collection := entities.ReconstructCollection(
				entities.NewCollectionID(),
				testUser.ID(),
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
			collections = append(collections, collection)
		}

		// Set up mock expectations
		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), "test-token", testUser.ID().String()).
			Return([]*entities.Group{}, nil).
			Times(1)

		mockCollectionRepo.EXPECT().
			GetByUserID(gomock.Any(), testUser.ID()).
			Return(collections, nil).
			Times(1)

		req := newTestRequest(http.MethodGet, "/collections", nil)
		rr, ctx := createReq(req, logger)

		ctx.Set("auth_user", testUser)
		ctx.Set("auth_token", "test-token")
		ctx.Params = []gin.Param{
			{Key: "id", Value: testUser.ID().String()},
		}

		controller.GetCollections(ctx)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp response.CollectionListResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Len(t, resp.Collections, 2)
		assert.Equal(t, 2, resp.Total)
	})

	t.Run("error - unauthorized access", func(t *testing.T) {
		testUser := randomUser()
		differentUserID := entities.NewUserID()

		req := newTestRequest(http.MethodGet, "/collections", nil)
		rr, ctx := createReq(req, logger)

		ctx.Set("auth_user", testUser)
		ctx.Set("auth_token", "test-token")
		ctx.Params = []gin.Param{
			{Key: "id", Value: differentUserID.String()}, // Different user ID
		}

		controller.GetCollections(ctx)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("error - database failure", func(t *testing.T) {
		testUser := randomUser()

		// Set up mock expectations for failure
		mockAuthService.EXPECT().
			GetUserGroups(gomock.Any(), "test-token", testUser.ID().String()).
			Return([]*entities.Group{}, nil).
			Times(1)

		mockCollectionRepo.EXPECT().
			GetByUserID(gomock.Any(), testUser.ID()).
			Return(nil, errors.New("database connection failed")).
			Times(1)

		req := newTestRequest(http.MethodGet, "/collections", nil)
		rr, ctx := createReq(req, logger)

		ctx.Set("auth_user", testUser)
		ctx.Set("auth_token", "test-token")
		ctx.Params = []gin.Param{
			{Key: "id", Value: testUser.ID().String()},
		}

		controller.GetCollections(ctx)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestCollectionController_DeleteCollection(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Mock repositories
	mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)

	// Create real use case with mocked dependencies
	deleteCollectionUC := usecases.NewDeleteCollectionUseCase(mockCollectionRepo)

	controller := &CollectionController{
		deleteCollectionUC: deleteCollectionUC,
		logger:             logger,
	}

	t.Run("success - delete collection", func(t *testing.T) {
		testUser := randomUser()
		collectionID := entities.NewCollectionID()

		// Create test collection
		collectionName, _ := entities.NewCollectionName("Test Collection")
		testCollection := entities.ReconstructCollection(
			collectionID,
			testUser.ID(),
			nil,
			collectionName,
			nil,
			entities.ObjectTypeGeneral,
			[]entities.Container{}, // Empty containers
			[]string{},
			"",
			time.Now(),
			time.Now(),
		)

		// Set up mock expectations
		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(testCollection, nil).
			Times(1)

		mockCollectionRepo.EXPECT().
			Delete(gomock.Any(), collectionID).
			Return(nil).
			Times(1)

		req := newTestRequest(http.MethodDelete, "/collections/"+collectionID.String(), nil)
		rr, ctx := createReq(req, logger)

		ctx.Set("auth_user", testUser)
		ctx.Params = []gin.Param{
			{Key: "id", Value: testUser.ID().String()},
			{Key: "collection_id", Value: collectionID.String()},
		}

		controller.DeleteCollection(ctx)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp map[string]bool
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.True(t, resp["success"])
	})

	t.Run("error - invalid collection ID", func(t *testing.T) {
		testUser := randomUser()

		req := newTestRequest(http.MethodDelete, "/collections/invalid-id", nil)
		rr, ctx := createReq(req, logger)

		ctx.Set("auth_user", testUser)
		ctx.Params = []gin.Param{
			{Key: "id", Value: testUser.ID().String()},
			{Key: "collection_id", Value: "invalid-id"},
		}

		controller.DeleteCollection(ctx)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("error - collection not found", func(t *testing.T) {
		testUser := randomUser()
		collectionID := entities.NewCollectionID()

		// Set up mock expectations for not found
		mockCollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(nil, errors.New("collection not found")).
			Times(1)

		req := newTestRequest(http.MethodDelete, "/collections/"+collectionID.String(), nil)
		rr, ctx := createReq(req, logger)

		ctx.Set("auth_user", testUser)
		ctx.Params = []gin.Param{
			{Key: "id", Value: testUser.ID().String()},
			{Key: "collection_id", Value: collectionID.String()},
		}

		controller.DeleteCollection(ctx)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}
