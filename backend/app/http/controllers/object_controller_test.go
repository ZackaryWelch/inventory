package controllers

import (
	"encoding/json/v2"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/nishiki/backend/app/http/request"
	"github.com/nishiki/backend/app/http/response"
	"github.com/nishiki/backend/domain/entities"
)

func TestObjectController_CreateObject(t *testing.T) {
	t.Parallel()

	c, m := newTestContainer(t)
	controller := NewObjectController(c, c.GetLogger())

	t.Run("success - create object", func(t *testing.T) {
		testUser := randomUser()
		collectionID := entities.NewCollectionID()
		containerID := entities.NewContainerID()

		requestBody := request.CreateObjectRequest{
			ContainerID: containerID.String(),
			Name:        "Test Object",
			ObjectType:  "general",
			Properties:  map[string]any{"description": "Test description"},
			Tags:        []string{"test", "example"},
		}

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
			nil,
			time.Now(),
			time.Now(),
		)

		m.ContainerRepo.EXPECT().GetByID(gomock.Any(), containerID).Return(testContainer, nil)
		m.AuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", testUser.ID().String()).Return([]*entities.Group{}, nil)
		m.CollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(testCollection, nil)
		m.ContainerRepo.EXPECT().AddObject(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		req := newTestRequest(http.MethodPost, "/accounts/"+testUser.ID().String()+"/objects", requestBody)
		req.SetPathValue("id", testUser.ID().String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.CreateObject(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var resp response.ObjectResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.NotEmpty(t, resp.ID)
		assert.Equal(t, "Test Object", resp.Name)
	})

	t.Run("error - invalid request body", func(t *testing.T) {
		testUser := randomUser()

		requestBody := map[string]any{
			"name": "",
		}

		req := newTestRequest(http.MethodPost, "/accounts/"+testUser.ID().String()+"/objects", requestBody)
		req.SetPathValue("id", testUser.ID().String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.CreateObject(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("error - access denied", func(t *testing.T) {
		testUser := randomUser()
		collectionID := entities.NewCollectionID()
		containerID := entities.NewContainerID()

		requestBody := request.CreateObjectRequest{
			ContainerID: containerID.String(),
			Name:        "Test Object",
			ObjectType:  "general",
		}

		// Container found, but collection owned by different user
		containerName, _ := entities.NewContainerName("Test Container")
		testContainer, _ := entities.NewContainer(entities.ContainerProps{
			CollectionID: collectionID,
			Name:         containerName,
		})

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
			nil,
			time.Now(),
			time.Now(),
		)

		m.ContainerRepo.EXPECT().GetByID(gomock.Any(), containerID).Return(testContainer, nil)
		m.AuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", testUser.ID().String()).Return([]*entities.Group{}, nil)
		m.CollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(testCollection, nil)

		req := newTestRequest(http.MethodPost, "/accounts/"+testUser.ID().String()+"/objects", requestBody)
		req.SetPathValue("id", testUser.ID().String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.CreateObject(rr, req)

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

		m.ContainerRepo.EXPECT().GetByID(gomock.Any(), containerID).Return(nil, errors.New("container not found"))

		req := newTestRequest(http.MethodPost, "/accounts/"+testUser.ID().String()+"/objects", requestBody)
		req.SetPathValue("id", testUser.ID().String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.CreateObject(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestObjectController_DeleteObject(t *testing.T) {
	t.Parallel()

	c, m := newTestContainer(t)
	controller := NewObjectController(c, c.GetLogger())

	t.Run("success - delete object", func(t *testing.T) {
		testUser := randomUser()
		objectID := entities.NewObjectID()
		containerID := entities.NewContainerID()
		collectionID := entities.NewCollectionID()

		// Create an object with the specific ID so RemoveObject succeeds
		objectName, _ := entities.NewObjectName("Test Object")
		objectDesc := entities.NewObjectDescription("")
		testObject := entities.ReconstructObject(objectID, objectName, objectDesc, entities.ObjectTypeGeneral, "", nil, "", nil, nil, "", nil, time.Now(), time.Now())

		// Create a container that already holds the object
		containerName, _ := entities.NewContainerName("Test Container")
		testContainer := entities.ReconstructContainer(
			entities.NewContainerID(),
			collectionID,
			containerName,
			entities.ContainerTypeGeneral,
			nil, nil, nil,
			[]entities.Object{*testObject},
			"", nil, nil, nil, nil,
			time.Now(), time.Now(),
		)

		// Collection owned by testUser
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
			nil,
			time.Now(),
			time.Now(),
		)

		m.ContainerRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(testContainer, nil)
		m.AuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", testUser.ID().String()).Return([]*entities.Group{}, nil)
		m.CollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(testCollection, nil)
		m.ContainerRepo.EXPECT().RemoveObject(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		req := newTestRequest(http.MethodDelete, "/accounts/"+testUser.ID().String()+"/objects/"+objectID.String()+"?container_id="+containerID.String(), nil)
		req.SetPathValue("id", testUser.ID().String())
		req.SetPathValue("object_id", objectID.String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.DeleteObject(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp response.DeleteObjectResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
	})

	t.Run("error - invalid object ID", func(t *testing.T) {
		testUser := randomUser()
		containerID := entities.NewContainerID()

		req := newTestRequest(http.MethodDelete, "/accounts/"+testUser.ID().String()+"/objects/invalid-id?container_id="+containerID.String(), nil)
		req.SetPathValue("id", testUser.ID().String())
		req.SetPathValue("object_id", "invalid-id")
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.DeleteObject(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("error - object not found", func(t *testing.T) {
		testUser := randomUser()
		objectID := entities.NewObjectID()
		containerID := entities.NewContainerID()

		// Container not found causes early return
		m.ContainerRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(nil, errors.New("container not found"))

		req := newTestRequest(http.MethodDelete, "/accounts/"+testUser.ID().String()+"/objects/"+objectID.String()+"?container_id="+containerID.String(), nil)
		req.SetPathValue("id", testUser.ID().String())
		req.SetPathValue("object_id", objectID.String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.DeleteObject(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestObjectController_BulkImport(t *testing.T) {
	t.Parallel()

	c, m := newTestContainer(t)
	controller := NewObjectController(c, c.GetLogger())

	t.Run("success - bulk import objects", func(t *testing.T) {
		testUser := randomUser()
		collectionID := entities.NewCollectionID()
		containerID := entities.NewContainerID()

		requestBody := request.BulkImportRequest{
			ContainerID: containerID.String(),
			Format:      "json",
			ObjectType:  "general",
			Data: []map[string]any{
				{"name": "Object 1", "description": "First object"},
				{"name": "Object 2", "description": "Second object"},
			},
			DefaultTags: []string{"imported"},
		}

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
			nil,
			time.Now(),
			time.Now(),
		)

		m.ContainerRepo.EXPECT().GetByID(gomock.Any(), containerID).Return(testContainer, nil)
		m.AuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", testUser.ID().String()).Return([]*entities.Group{}, nil)
		m.CollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(testCollection, nil)
		m.ContainerRepo.EXPECT().Update(gomock.Any(), testContainer).Return(nil).Times(1)

		req := newTestRequest(http.MethodPost, "/accounts/"+testUser.ID().String()+"/import", requestBody)
		req.SetPathValue("id", testUser.ID().String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.BulkImport(rr, req)

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

		requestBody := request.BulkImportRequest{
			ContainerID: containerID.String(),
			Format:      "json",
			ObjectType:  "general",
			Data:        []map[string]any{},
		}

		req := newTestRequest(http.MethodPost, "/accounts/"+testUser.ID().String()+"/import", requestBody)
		req.SetPathValue("id", testUser.ID().String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.BulkImport(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
