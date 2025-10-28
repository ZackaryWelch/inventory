package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nishiki/backend-go/app/http/response"
	"github.com/nishiki/backend-go/domain/entities"
)

func TestCollectionAPI_CreateCollection(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	// Create test user
	testUser := CreateTestUser()
	testClaims := CreateTestAuthClaims(testUser.ID().String())
	ts.SetupDefaultAuthMocks(testUser, testClaims)

	t.Run("success - create collection", func(t *testing.T) {
		ts.CleanDatabase(t)

		reqBody := map[string]interface{}{
			"user_id":     testUser.ID().String(),
			"name":        "My Food Collection",
			"object_type": "food",
			"tags":        []string{"pantry"},
			"location":    "Kitchen",
		}

		headers := map[string]string{
			"Authorization": "Bearer test-token",
		}

		resp, body, err := ts.MakeRequest(http.MethodPost, "/accounts/"+testUser.ID().String()+"/collections", reqBody, headers)
		require.NoError(t, err)

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var collectionResp response.CollectionResponse
		err = json.Unmarshal(body, &collectionResp)
		require.NoError(t, err)

		assert.Equal(t, "My Food Collection", collectionResp.Name)
		assert.Equal(t, "food", collectionResp.ObjectType)
		assert.NotEmpty(t, collectionResp.ID)
	})

	t.Run("error - invalid request body", func(t *testing.T) {
		ts.CleanDatabase(t)

		reqBody := map[string]interface{}{
			"user_id": testUser.ID().String(),
			"name":    "", // Empty name is invalid
		}

		headers := map[string]string{
			"Authorization": "Bearer test-token",
		}

		resp, _, err := ts.MakeRequest(http.MethodPost, "/accounts/"+testUser.ID().String()+"/collections", reqBody, headers)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("error - unauthorized (no token)", func(t *testing.T) {
		ts.CleanDatabase(t)

		reqBody := map[string]interface{}{
			"user_id":     testUser.ID().String(),
			"name":        "My Collection",
			"object_type": "general",
		}

		resp, _, err := ts.MakeRequest(http.MethodPost, "/accounts/"+testUser.ID().String()+"/collections", reqBody, nil)
		require.NoError(t, err)

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestCollectionAPI_GetCollections(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	testUser := CreateTestUser()
	testClaims := CreateTestAuthClaims(testUser.ID().String())
	ts.SetupDefaultAuthMocks(testUser, testClaims)

	t.Run("success - get user collections", func(t *testing.T) {
		ts.CleanDatabase(t)

		// Create test collections
		collection1 := ts.CreateTestCollection(t, testUser.ID(), entities.ObjectTypeFood)
		collection2 := ts.CreateTestCollection(t, testUser.ID(), entities.ObjectTypeBook)

		headers := map[string]string{
			"Authorization": "Bearer test-token",
		}

		resp, body, err := ts.MakeRequest(http.MethodGet, "/accounts/"+testUser.ID().String()+"/collections", nil, headers)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var listResp response.CollectionListResponse
		err = json.Unmarshal(body, &listResp)
		require.NoError(t, err)

		assert.Equal(t, 2, listResp.Total)
		assert.Len(t, listResp.Collections, 2)

		// Verify collection IDs are present
		ids := []string{listResp.Collections[0].ID, listResp.Collections[1].ID}
		assert.Contains(t, ids, collection1.ID().String())
		assert.Contains(t, ids, collection2.ID().String())
	})

	t.Run("success - get empty collection list", func(t *testing.T) {
		ts.CleanDatabase(t)

		headers := map[string]string{
			"Authorization": "Bearer test-token",
		}

		resp, body, err := ts.MakeRequest(http.MethodGet, "/accounts/"+testUser.ID().String()+"/collections", nil, headers)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var listResp response.CollectionListResponse
		err = json.Unmarshal(body, &listResp)
		require.NoError(t, err)

		assert.Equal(t, 0, listResp.Total)
		assert.Len(t, listResp.Collections, 0)
	})
}

func TestCollectionAPI_GetCollection(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	testUser := CreateTestUser()
	testClaims := CreateTestAuthClaims(testUser.ID().String())
	ts.SetupDefaultAuthMocks(testUser, testClaims)

	t.Run("success - get collection by ID", func(t *testing.T) {
		ts.CleanDatabase(t)

		// Create test collection
		collection := ts.CreateTestCollection(t, testUser.ID(), entities.ObjectTypeFood)

		path := "/accounts/" + testUser.ID().String() + "/collections/" + collection.ID().String()
		t.Logf("Making GET request to: %s", path)
		t.Logf("Collection ID: %s", collection.ID().String())

		headers := map[string]string{
			"Authorization": "Bearer test-token",
		}

		resp, body, err := ts.MakeRequest(
			http.MethodGet,
			path,
			nil,
			headers,
		)
		require.NoError(t, err)

		if resp.StatusCode != http.StatusOK {
			t.Logf("Response body: %s", string(body))
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var collectionResp response.CollectionResponse
		err = json.Unmarshal(body, &collectionResp)
		require.NoError(t, err)

		assert.Equal(t, collection.ID().String(), collectionResp.ID)
		assert.Equal(t, "Test Collection", collectionResp.Name)
	})

	t.Run("error - collection not found", func(t *testing.T) {
		ts.CleanDatabase(t)

		fakeCollectionID := entities.NewCollectionID()

		headers := map[string]string{
			"Authorization": "Bearer test-token",
		}

		resp, _, err := ts.MakeRequest(
			http.MethodGet,
			"/accounts/"+testUser.ID().String()+"/collections/"+fakeCollectionID.String(),
			nil,
			headers,
		)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestCollectionAPI_UpdateCollection(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	testUser := CreateTestUser()
	testClaims := CreateTestAuthClaims(testUser.ID().String())
	ts.SetupDefaultAuthMocks(testUser, testClaims)

	t.Run("success - update collection", func(t *testing.T) {
		ts.CleanDatabase(t)

		// Create test collection
		collection := ts.CreateTestCollection(t, testUser.ID(), entities.ObjectTypeFood)

		reqBody := map[string]interface{}{
			"name":     "Updated Collection Name",
			"tags":     []string{"updated", "new-tag"},
			"location": "New Location",
		}

		headers := map[string]string{
			"Authorization": "Bearer test-token",
		}

		resp, body, err := ts.MakeRequest(
			http.MethodPut,
			"/accounts/"+testUser.ID().String()+"/collections/"+collection.ID().String(),
			reqBody,
			headers,
		)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var collectionResp response.CollectionResponse
		err = json.Unmarshal(body, &collectionResp)
		require.NoError(t, err)

		assert.Equal(t, "Updated Collection Name", collectionResp.Name)
		assert.Equal(t, "New Location", collectionResp.Location)
	})

	t.Run("error - invalid name", func(t *testing.T) {
		ts.CleanDatabase(t)

		collection := ts.CreateTestCollection(t, testUser.ID(), entities.ObjectTypeFood)

		reqBody := map[string]interface{}{
			"name": "", // Empty name is invalid
		}

		headers := map[string]string{
			"Authorization": "Bearer test-token",
		}

		resp, _, err := ts.MakeRequest(
			http.MethodPut,
			"/accounts/"+testUser.ID().String()+"/collections/"+collection.ID().String(),
			reqBody,
			headers,
		)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestCollectionAPI_DeleteCollection(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	testUser := CreateTestUser()
	testClaims := CreateTestAuthClaims(testUser.ID().String())
	ts.SetupDefaultAuthMocks(testUser, testClaims)

	t.Run("success - delete empty collection", func(t *testing.T) {
		ts.CleanDatabase(t)

		// Create test collection
		collection := ts.CreateTestCollection(t, testUser.ID(), entities.ObjectTypeFood)

		headers := map[string]string{
			"Authorization": "Bearer test-token",
		}

		resp, body, err := ts.MakeRequest(
			http.MethodDelete,
			"/accounts/"+testUser.ID().String()+"/collections/"+collection.ID().String(),
			nil,
			headers,
		)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var deleteResp map[string]bool
		err = json.Unmarshal(body, &deleteResp)
		require.NoError(t, err)

		assert.True(t, deleteResp["success"])
	})

	t.Run("error - collection not found", func(t *testing.T) {
		ts.CleanDatabase(t)

		fakeCollectionID := entities.NewCollectionID()

		headers := map[string]string{
			"Authorization": "Bearer test-token",
		}

		resp, _, err := ts.MakeRequest(
			http.MethodDelete,
			"/accounts/"+testUser.ID().String()+"/collections/"+fakeCollectionID.String(),
			nil,
			headers,
		)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

