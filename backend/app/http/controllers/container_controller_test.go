package controllers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

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

	mockContainerRepo := mocks.NewMockContainerRepository(mockCtrl)
	mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
	mockAuthService := mocks.NewMockAuthService(mockCtrl)

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

		collectionName, _ := entities.NewCollectionName("Test Collection")
		testCollection := entities.ReconstructCollection(
			collectionID,
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

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", testUser.ID().String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(testCollection, nil)
		mockContainerRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
		mockCollectionRepo.EXPECT().Update(gomock.Any(), testCollection).Return(nil)

		req := newTestRequest(http.MethodPost, "/containers", requestBody)
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.CreateContainer(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var resp response.ContainerResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.NotEmpty(t, resp.ID)
		assert.Equal(t, "Test Container", resp.Name)
	})

	t.Run("error - invalid request body", func(t *testing.T) {
		testUser := randomUser()
		collectionID := entities.NewCollectionID()

		requestBody := request.CreateContainerRequest{
			CollectionID: collectionID.String(),
			Name:         "",
		}

		req := newTestRequest(http.MethodPost, "/containers", requestBody)
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.CreateContainer(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("error - invalid collection ID", func(t *testing.T) {
		testUser := randomUser()

		requestBody := request.CreateContainerRequest{
			CollectionID: "invalid-id",
			Name:         "Test Container",
		}

		req := newTestRequest(http.MethodPost, "/containers", requestBody)
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.CreateContainer(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("error - access denied", func(t *testing.T) {
		testUser := randomUser()
		collectionID := entities.NewCollectionID()

		requestBody := request.CreateContainerRequest{
			CollectionID: collectionID.String(),
			Name:         "Test Container",
		}

		// Collection owned by a different user — use case returns access denied
		otherUser := randomUser()
		collectionName, _ := entities.NewCollectionName("Test Collection")
		testCollection := entities.ReconstructCollection(
			collectionID,
			otherUser.ID(),
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

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", testUser.ID().String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(testCollection, nil)

		req := newTestRequest(http.MethodPost, "/containers", requestBody)
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.CreateContainer(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("error - collection not found", func(t *testing.T) {
		testUser := randomUser()
		collectionID := entities.NewCollectionID()

		requestBody := request.CreateContainerRequest{
			CollectionID: collectionID.String(),
			Name:         "Test Container",
		}

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", testUser.ID().String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(nil, errors.New("collection not found"))

		req := newTestRequest(http.MethodPost, "/containers", requestBody)
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.CreateContainer(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestContainerController_GetContainer(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	mockContainerRepo := mocks.NewMockContainerRepository(mockCtrl)
	mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
	mockAuthService := mocks.NewMockAuthService(mockCtrl)

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

		collectionName, _ := entities.NewCollectionName("Test Collection")
		testCollection := entities.ReconstructCollection(
			collectionID,
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

		mockContainerRepo.EXPECT().GetByID(gomock.Any(), containerID).Return(testContainer, nil)
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", testUser.ID().String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(testCollection, nil)

		req := newTestRequest(http.MethodGet, "/containers/"+containerID.String(), nil)
		req.SetPathValue("container_id", containerID.String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.GetContainer(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp response.ContainerResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, testContainer.ID().String(), resp.ID)
		assert.Equal(t, "Test Container", resp.Name)
	})

	t.Run("error - invalid container ID", func(t *testing.T) {
		testUser := randomUser()

		req := newTestRequest(http.MethodGet, "/containers/invalid-id", nil)
		req.SetPathValue("container_id", "invalid-id")
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.GetContainer(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("error - container not found", func(t *testing.T) {
		testUser := randomUser()
		containerID := entities.NewContainerID()

		mockContainerRepo.EXPECT().GetByID(gomock.Any(), containerID).Return(nil, errors.New("container not found"))

		req := newTestRequest(http.MethodGet, "/containers/"+containerID.String(), nil)
		req.SetPathValue("container_id", containerID.String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.GetContainer(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("error - access denied", func(t *testing.T) {
		testUser := randomUser()
		containerID := entities.NewContainerID()
		collectionID := entities.NewCollectionID()

		containerName, _ := entities.NewContainerName("Test Container")
		testContainer, _ := entities.NewContainer(entities.ContainerProps{
			CollectionID: collectionID,
			Name:         containerName,
		})

		// Collection owned by a different user — use case returns access denied
		otherUser := randomUser()
		collectionName, _ := entities.NewCollectionName("Test Collection")
		testCollection := entities.ReconstructCollection(
			collectionID,
			otherUser.ID(),
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

		mockContainerRepo.EXPECT().GetByID(gomock.Any(), containerID).Return(testContainer, nil)
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", testUser.ID().String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(testCollection, nil)

		req := newTestRequest(http.MethodGet, "/containers/"+containerID.String(), nil)
		req.SetPathValue("container_id", containerID.String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.GetContainer(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})
}
