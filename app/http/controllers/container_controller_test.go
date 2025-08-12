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

func TestContainerController_CreateContainer(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Mock repositories and services
	mockContainerRepo := mocks.NewMockContainerRepository(mockCtrl)
	mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
	mockAuthService := mocks.NewMockAuthService(mockCtrl)

	// Create use cases with mocked dependencies
	createContainerUC := usecases.NewCreateContainerUseCase(mockContainerRepo, mockCollectionRepo, mockAuthService)
	getAllContainersUC := usecases.NewGetAllContainersUseCase(mockContainerRepo, mockAuthService)
	getContainerByIDUC := usecases.NewGetContainerByIDUseCase(mockContainerRepo, mockCollectionRepo, mockAuthService)
	getContainersUC := usecases.NewGetContainersUseCase(mockContainerRepo, mockAuthService)

	controller := &ContainerController{
		createContainerUC:  createContainerUC,
		getAllContainersUC: getAllContainersUC,
		getContainerByIDUC: getContainerByIDUC,
		getContainersUC:    getContainersUC,
		logger:             logger,
	}

	t.Run("success - create container", func(t *testing.T) {
		testUser := randomUser()
		collectionID := entities.NewCollectionID()

		requestBody := request.CreateContainerRequest{
			CollectionID: collectionID.String(),
			Name:         "Test Container",
		}

		containerName, _ := entities.NewContainerName("Test Container")
		testContainer, _ := entities.NewContainer(entities.ContainerProps{
			CollectionID: collectionID,
			Name:         containerName,
		})

		// Mock expectations
		mockAuthService.EXPECT().ValidateToken(gomock.Any(), "test-token").Return(nil, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(nil, nil)
		mockContainerRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

		req := newTestRequest(http.MethodPost, "/containers", requestBody)
		rr, ctx := createReq(req, logger)

		ctx.Set("auth_user", testUser)
		ctx.Set("auth_token", "test-token")
		ctx.Params = []gin.Param{
			{Key: "account_id", Value: testUser.ID().String()},
			{Key: "collection_id", Value: collectionID.String()},
		}

		controller.CreateContainer(ctx)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var resp response.ContainerResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, testContainer.ID().String(), resp.ID)
		assert.Equal(t, "Test Container", resp.Name)
	})

	t.Run("error - invalid request body", func(t *testing.T) {
		testUser := randomUser()
		collectionID := entities.NewCollectionID()

		// Invalid request (empty name)
		requestBody := request.CreateContainerRequest{
			CollectionID: collectionID.String(),
			Name:         "",
		}

		req := newTestRequest(http.MethodPost, "/containers", requestBody)
		rr, ctx := createReq(req, logger)

		ctx.Set("auth_user", testUser)
		ctx.Set("auth_token", "test-token")
		ctx.Params = []gin.Param{
			{Key: "account_id", Value: testUser.ID().String()},
			{Key: "collection_id", Value: collectionID.String()},
		}

		controller.CreateContainer(ctx)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("error - invalid collection ID", func(t *testing.T) {
		testUser := randomUser()

		requestBody := request.CreateContainerRequest{
			CollectionID: "invalid-id",
			Name:         "Test Container",
		}

		req := newTestRequest(http.MethodPost, "/containers", requestBody)
		rr, ctx := createReq(req, logger)

		ctx.Set("auth_user", testUser)
		ctx.Set("auth_token", "test-token")
		ctx.Params = []gin.Param{
			{Key: "account_id", Value: testUser.ID().String()},
			{Key: "collection_id", Value: "invalid-id"},
		}

		controller.CreateContainer(ctx)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("error - access denied", func(t *testing.T) {
		testUser := randomUser()
		collectionID := entities.NewCollectionID()

		requestBody := request.CreateContainerRequest{
			CollectionID: collectionID.String(),
			Name:         "Test Container",
		}

		// Simulate access denied from auth service
		mockAuthService.EXPECT().ValidateToken(gomock.Any(), "test-token").Return(nil, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(nil, errors.New("user is not a member of the group"))

		req := newTestRequest(http.MethodPost, "/containers", requestBody)
		rr, ctx := createReq(req, logger)

		ctx.Set("auth_user", testUser)
		ctx.Set("auth_token", "test-token")
		ctx.Params = []gin.Param{
			{Key: "account_id", Value: testUser.ID().String()},
			{Key: "collection_id", Value: collectionID.String()},
		}

		controller.CreateContainer(ctx)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("error - collection not found", func(t *testing.T) {
		testUser := randomUser()
		collectionID := entities.NewCollectionID()

		requestBody := request.CreateContainerRequest{
			CollectionID: collectionID.String(),
			Name:         "Test Container",
		}

		// Simulate collection not found
		mockAuthService.EXPECT().ValidateToken(gomock.Any(), "test-token").Return(nil, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(nil, errors.New("collection not found"))

		req := newTestRequest(http.MethodPost, "/containers", requestBody)
		rr, ctx := createReq(req, logger)

		ctx.Set("auth_user", testUser)
		ctx.Set("auth_token", "test-token")
		ctx.Params = []gin.Param{
			{Key: "account_id", Value: testUser.ID().String()},
			{Key: "collection_id", Value: collectionID.String()},
		}

		controller.CreateContainer(ctx)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestContainerController_GetContainer(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Mock repositories and services
	mockContainerRepo := mocks.NewMockContainerRepository(mockCtrl)
	mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
	mockAuthService := mocks.NewMockAuthService(mockCtrl)

	// Create use case with mocked dependencies
	getContainerByIDUC := usecases.NewGetContainerByIDUseCase(mockContainerRepo, mockCollectionRepo, mockAuthService)

	controller := &ContainerController{
		getContainerByIDUC: getContainerByIDUC,
		logger:             logger,
	}

	t.Run("success - get container", func(t *testing.T) {
		testUser := randomUser()
		containerID := entities.NewContainerID()
		collectionID := entities.NewCollectionID()

		containerName, _ := entities.NewContainerName("Test Container")
		testContainer, _ := entities.NewContainer(entities.ContainerProps{
			CollectionID: collectionID,
			Name:         containerName,
		})

		// Mock expectations
		mockAuthService.EXPECT().ValidateToken(gomock.Any(), "test-token").Return(nil, nil)
		mockContainerRepo.EXPECT().GetByID(gomock.Any(), containerID).Return(testContainer, nil)

		req := newTestRequest(http.MethodGet, "/containers/"+containerID.String(), nil)
		rr, ctx := createReq(req, logger)

		ctx.Set("auth_user", testUser)
		ctx.Set("auth_token", "test-token")
		ctx.Params = []gin.Param{
			{Key: "account_id", Value: testUser.ID().String()},
			{Key: "container_id", Value: containerID.String()},
		}

		controller.GetContainer(ctx)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp response.ContainerResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, testContainer.ID().String(), resp.ID)
		assert.Equal(t, "Test Container", resp.Name)
	})

	t.Run("error - invalid container ID", func(t *testing.T) {
		testUser := randomUser()

		req := newTestRequest(http.MethodGet, "/containers/invalid-id", nil)
		rr, ctx := createReq(req, logger)

		ctx.Set("auth_user", testUser)
		ctx.Set("auth_token", "test-token")
		ctx.Params = []gin.Param{
			{Key: "account_id", Value: testUser.ID().String()},
			{Key: "container_id", Value: "invalid-id"},
		}

		controller.GetContainer(ctx)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("error - container not found", func(t *testing.T) {
		testUser := randomUser()
		containerID := entities.NewContainerID()

		// Simulate container not found
		mockAuthService.EXPECT().ValidateToken(gomock.Any(), "test-token").Return(nil, nil)
		mockContainerRepo.EXPECT().GetByID(gomock.Any(), containerID).Return(nil, errors.New("container not found"))

		req := newTestRequest(http.MethodGet, "/containers/"+containerID.String(), nil)
		rr, ctx := createReq(req, logger)

		ctx.Set("auth_user", testUser)
		ctx.Set("auth_token", "test-token")
		ctx.Params = []gin.Param{
			{Key: "account_id", Value: testUser.ID().String()},
			{Key: "container_id", Value: containerID.String()},
		}

		controller.GetContainer(ctx)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("error - access denied", func(t *testing.T) {
		testUser := randomUser()
		containerID := entities.NewContainerID()

		// Simulate access denied
		mockAuthService.EXPECT().ValidateToken(gomock.Any(), "test-token").Return(nil, errors.New("access denied: user is not a member of the container's group"))

		req := newTestRequest(http.MethodGet, "/containers/"+containerID.String(), nil)
		rr, ctx := createReq(req, logger)

		ctx.Set("auth_user", testUser)
		ctx.Set("auth_token", "test-token")
		ctx.Params = []gin.Param{
			{Key: "account_id", Value: testUser.ID().String()},
			{Key: "container_id", Value: containerID.String()},
		}

		controller.GetContainer(ctx)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})
}

