package controllers

import (
	"encoding/json/v2"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/nishiki/backend/app/http/request"
	"github.com/nishiki/backend/app/http/response"
	"github.com/nishiki/backend/domain/entities"
)

func TestCollectionController_CreateCollection(t *testing.T) {
	t.Parallel()

	c, m := newTestContainer(t)
	controller := NewCollectionController(c, c.GetLogger())

	t.Run("success - create collection", func(t *testing.T) {
		testUser := randomUser()

		requestBody := request.CreateCollectionRequest{
			Name:       "Test Collection",
			ObjectType: "general",
		}

		// No GroupID -> GetUserGroups is NOT called; only Create is called
		m.CollectionRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

		req := newTestRequest(http.MethodPost, "/accounts/"+testUser.ID().String()+"/collections", requestBody)
		req.SetPathValue("id", testUser.ID().String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.CreateCollection(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var resp response.CollectionResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, "Test Collection", resp.Name)
		assert.Equal(t, "general", resp.ObjectType)
	})

	t.Run("error - invalid request body", func(t *testing.T) {
		testUser := randomUser()

		requestBody := request.CreateCollectionRequest{
			Name:       "",
			ObjectType: "general",
		}

		req := newTestRequest(http.MethodPost, "/accounts/"+testUser.ID().String()+"/collections", requestBody)
		req.SetPathValue("id", testUser.ID().String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.CreateCollection(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("error - database failure", func(t *testing.T) {
		testUser := randomUser()

		requestBody := request.CreateCollectionRequest{
			Name:       "Test Collection",
			ObjectType: "general",
		}

		// No GroupID -> GetUserGroups NOT called; Create fails
		m.CollectionRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(errors.New("database connection failed")).
			Times(1)

		req := newTestRequest(http.MethodPost, "/accounts/"+testUser.ID().String()+"/collections", requestBody)
		req.SetPathValue("id", testUser.ID().String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.CreateCollection(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("error - auth service failure", func(t *testing.T) {
		testUser := randomUser()
		groupIDStr := "group-123" // non-empty so GetGroupID returns non-nil -> GetUserGroups is called

		// Provide a GroupID so GetUserGroups IS called
		requestBody := request.CreateCollectionRequest{
			Name:       "Test Collection",
			ObjectType: "general",
			GroupID:    &groupIDStr,
		}

		m.AuthService.EXPECT().
			GetUserGroups(gomock.Any(), "test-token", testUser.ID().String()).
			Return(nil, errors.New("auth service unavailable")).
			Times(1)

		req := newTestRequest(http.MethodPost, "/accounts/"+testUser.ID().String()+"/collections", requestBody)
		req.SetPathValue("id", testUser.ID().String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.CreateCollection(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestCollectionController_GetCollections(t *testing.T) {
	t.Parallel()

	c, m := newTestContainer(t)
	controller := NewCollectionController(c, c.GetLogger())

	t.Run("success - get collections", func(t *testing.T) {
		testUser := randomUser()

		collections := []*entities.Collection{}
		for range 2 {
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
				nil,
				time.Now(),
				time.Now(),
			)
			collections = append(collections, collection)
		}

		// GetCollectionsUseCase calls GetByUserIDSummary (no GetUserGroups in simple path)
		m.CollectionRepo.EXPECT().
			GetByUserIDSummary(gomock.Any(), testUser.ID()).
			Return(collections, nil).
			Times(1)

		req := newTestRequest(http.MethodGet, "/accounts/"+testUser.ID().String()+"/collections", nil)
		req.SetPathValue("id", testUser.ID().String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.GetCollections(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp []response.CollectionResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Len(t, resp, 2)
	})

	t.Run("error - unauthorized access", func(t *testing.T) {
		testUser := randomUser()
		differentUserID := entities.NewUserID()

		req := newTestRequest(http.MethodGet, "/accounts/"+differentUserID.String()+"/collections", nil)
		req.SetPathValue("id", differentUserID.String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.GetCollections(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("error - database failure", func(t *testing.T) {
		testUser := randomUser()

		// GetCollectionsUseCase calls GetByUserIDSummary which returns error
		m.CollectionRepo.EXPECT().
			GetByUserIDSummary(gomock.Any(), testUser.ID()).
			Return(nil, errors.New("database connection failed")).
			Times(1)

		req := newTestRequest(http.MethodGet, "/accounts/"+testUser.ID().String()+"/collections", nil)
		req.SetPathValue("id", testUser.ID().String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.GetCollections(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestCollectionController_DeleteCollection(t *testing.T) {
	t.Parallel()

	c, m := newTestContainer(t)
	controller := NewCollectionController(c, c.GetLogger())

	t.Run("success - delete collection", func(t *testing.T) {
		testUser := randomUser()
		collectionID := entities.NewCollectionID()

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

		m.CollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(testCollection, nil).
			Times(1)

		m.CollectionRepo.EXPECT().
			Delete(gomock.Any(), collectionID).
			Return(nil).
			Times(1)

		req := newTestRequest(http.MethodDelete, "/accounts/"+testUser.ID().String()+"/collections/"+collectionID.String(), nil)
		req.SetPathValue("id", testUser.ID().String())
		req.SetPathValue("collection_id", collectionID.String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.DeleteCollection(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp map[string]bool
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.True(t, resp["success"])
	})

	t.Run("error - invalid collection ID", func(t *testing.T) {
		testUser := randomUser()

		req := newTestRequest(http.MethodDelete, "/accounts/"+testUser.ID().String()+"/collections/invalid-id", nil)
		req.SetPathValue("id", testUser.ID().String())
		req.SetPathValue("collection_id", "invalid-id")
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.DeleteCollection(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("error - collection not found", func(t *testing.T) {
		testUser := randomUser()
		collectionID := entities.NewCollectionID()

		m.CollectionRepo.EXPECT().
			GetByID(gomock.Any(), collectionID).
			Return(nil, errors.New("collection not found")).
			Times(1)

		req := newTestRequest(http.MethodDelete, "/accounts/"+testUser.ID().String()+"/collections/"+collectionID.String(), nil)
		req.SetPathValue("id", testUser.ID().String())
		req.SetPathValue("collection_id", collectionID.String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.DeleteCollection(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}
