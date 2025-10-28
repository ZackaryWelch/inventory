package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nishiki/backend-go/app/http/response"
	"github.com/nishiki/backend-go/domain/entities"
)

func TestObjectAPI_CreateObject(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	testUser := CreateTestUser()
	testClaims := CreateTestAuthClaims(testUser.ID().String())
	ts.SetupDefaultAuthMocks(testUser, testClaims)

	t.Run("success - create object", func(t *testing.T) {
		ts.CleanDatabase(t)

		// Create test collection and container
		collection := ts.CreateTestCollection(t, testUser.ID(), entities.ObjectTypeGeneral)
		container := ts.CreateTestContainer(t, collection.ID())

		reqBody := map[string]interface{}{
			"container_id": container.ID().String(),
			"name":         "Test Object",
			"object_type":  "general",
			"properties": map[string]interface{}{
				"description": "A test object",
				"quantity":    5,
			},
			"tags": []string{"test", "sample"},
		}

		headers := map[string]string{
			"Authorization": "Bearer test-token",
		}

		resp, body, err := ts.MakeRequest(
			http.MethodPost,
			"/accounts/"+testUser.ID().String()+"/objects",
			reqBody,
			headers,
		)
		require.NoError(t, err)

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var objectResp response.ObjectResponse
		err = json.Unmarshal(body, &objectResp)
		require.NoError(t, err)

		assert.Equal(t, "Test Object", objectResp.Name)
		assert.Equal(t, "general", objectResp.ObjectType)
		assert.NotEmpty(t, objectResp.ID)
	})

	t.Run("error - invalid request body", func(t *testing.T) {
		ts.CleanDatabase(t)

		collection := ts.CreateTestCollection(t, testUser.ID(), entities.ObjectTypeGeneral)
		container := ts.CreateTestContainer(t, collection.ID())

		reqBody := map[string]interface{}{
			"container_id": container.ID().String(),
			"name":         "", // Empty name is invalid
		}

		headers := map[string]string{
			"Authorization": "Bearer test-token",
		}

		resp, _, err := ts.MakeRequest(
			http.MethodPost,
			"/accounts/"+testUser.ID().String()+"/objects",
			reqBody,
			headers,
		)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("error - container not found", func(t *testing.T) {
		ts.CleanDatabase(t)

		fakeContainerID := entities.NewContainerID()

		reqBody := map[string]interface{}{
			"container_id": fakeContainerID.String(),
			"name":         "Test Object",
			"object_type":  "general",
		}

		headers := map[string]string{
			"Authorization": "Bearer test-token",
		}

		resp, _, err := ts.MakeRequest(
			http.MethodPost,
			"/accounts/"+testUser.ID().String()+"/objects",
			reqBody,
			headers,
		)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestObjectAPI_GetCollectionObjects(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	testUser := CreateTestUser()
	testClaims := CreateTestAuthClaims(testUser.ID().String())
	ts.SetupDefaultAuthMocks(testUser, testClaims)

	t.Run("success - get collection objects", func(t *testing.T) {
		ts.CleanDatabase(t)

		// Create test collection and container with objects
		collection := ts.CreateTestCollection(t, testUser.ID(), entities.ObjectTypeGeneral)
		container := ts.CreateTestContainer(t, collection.ID())

		// Add objects to container
		ctx := context.Background()
		object1, _ := CreateTestObject("Object 1")
		object2, _ := CreateTestObject("Object 2")

		err := container.AddObject(*object1)
		require.NoError(t, err)
		err = container.AddObject(*object2)
		require.NoError(t, err)

		// Update container in database
		err = ts.Container.ContainerRepo.Update(ctx, container)
		require.NoError(t, err)

		headers := map[string]string{
			"Authorization": "Bearer test-token",
		}

		resp, body, err := ts.MakeRequest(
			http.MethodGet,
			"/accounts/"+testUser.ID().String()+"/collections/"+collection.ID().String()+"/objects",
			nil,
			headers,
		)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var listResp response.ObjectListResponse
		err = json.Unmarshal(body, &listResp)
		require.NoError(t, err)

		assert.Equal(t, 2, listResp.Total)
		assert.Len(t, listResp.Objects, 2)
	})

	t.Run("success - get empty objects list", func(t *testing.T) {
		ts.CleanDatabase(t)

		collection := ts.CreateTestCollection(t, testUser.ID(), entities.ObjectTypeGeneral)

		headers := map[string]string{
			"Authorization": "Bearer test-token",
		}

		resp, body, err := ts.MakeRequest(
			http.MethodGet,
			"/accounts/"+testUser.ID().String()+"/collections/"+collection.ID().String()+"/objects",
			nil,
			headers,
		)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var listResp response.ObjectListResponse
		err = json.Unmarshal(body, &listResp)
		require.NoError(t, err)

		assert.Equal(t, 0, listResp.Total)
		assert.Len(t, listResp.Objects, 0)
	})
}

func TestObjectAPI_UpdateObject(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	testUser := CreateTestUser()
	testClaims := CreateTestAuthClaims(testUser.ID().String())
	ts.SetupDefaultAuthMocks(testUser, testClaims)

	t.Run("success - update object", func(t *testing.T) {
		ts.CleanDatabase(t)

		// Create test collection and container with object
		collection := ts.CreateTestCollection(t, testUser.ID(), entities.ObjectTypeGeneral)
		container := ts.CreateTestContainer(t, collection.ID())

		ctx := context.Background()
		testObject, _ := CreateTestObject("Original Name")
		err := container.AddObject(*testObject)
		require.NoError(t, err)
		err = ts.Container.ContainerRepo.Update(ctx, container)
		require.NoError(t, err)

		reqBody := map[string]interface{}{
			"name": "Updated Name",
			"properties": map[string]interface{}{
				"updated": true,
			},
			"tags": []string{"updated"},
		}

		headers := map[string]string{
			"Authorization": "Bearer test-token",
		}

		resp, body, err := ts.MakeRequest(
			http.MethodPut,
			"/accounts/"+testUser.ID().String()+"/objects/"+testObject.ID().String()+"?container_id="+container.ID().String(),
			reqBody,
			headers,
		)
		require.NoError(t, err)

		if resp.StatusCode != http.StatusOK {
			t.Logf("Response body: %s", string(body))
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var objectResp response.ObjectResponse
		err = json.Unmarshal(body, &objectResp)
		require.NoError(t, err)

		assert.Equal(t, "Updated Name", objectResp.Name)
	})

	t.Run("error - object not found", func(t *testing.T) {
		ts.CleanDatabase(t)

		collection := ts.CreateTestCollection(t, testUser.ID(), entities.ObjectTypeGeneral)
		container := ts.CreateTestContainer(t, collection.ID())

		fakeObjectID := entities.NewObjectID()

		reqBody := map[string]interface{}{
			"name": "Updated Name",
		}

		headers := map[string]string{
			"Authorization": "Bearer test-token",
		}

		resp, _, err := ts.MakeRequest(
			http.MethodPut,
			"/accounts/"+testUser.ID().String()+"/objects/"+fakeObjectID.String()+"?container_id="+container.ID().String(),
			reqBody,
			headers,
		)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestObjectAPI_DeleteObject(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	testUser := CreateTestUser()
	testClaims := CreateTestAuthClaims(testUser.ID().String())
	ts.SetupDefaultAuthMocks(testUser, testClaims)

	t.Run("success - delete object", func(t *testing.T) {
		ts.CleanDatabase(t)

		// Create test collection and container with object
		collection := ts.CreateTestCollection(t, testUser.ID(), entities.ObjectTypeGeneral)
		container := ts.CreateTestContainer(t, collection.ID())

		ctx := context.Background()
		testObject, _ := CreateTestObject("Test Object")
		err := container.AddObject(*testObject)
		require.NoError(t, err)
		err = ts.Container.ContainerRepo.Update(ctx, container)
		require.NoError(t, err)

		headers := map[string]string{
			"Authorization": "Bearer test-token",
		}

		resp, body, err := ts.MakeRequest(
			http.MethodDelete,
			"/accounts/"+testUser.ID().String()+"/objects/"+testObject.ID().String()+"?container_id="+container.ID().String(),
			nil,
			headers,
		)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var deleteResp response.DeleteObjectResponse
		err = json.Unmarshal(body, &deleteResp)
		require.NoError(t, err)

		assert.True(t, deleteResp.Success)

		// Verify object was deleted
		updatedContainer, err := ts.Container.ContainerRepo.GetByID(ctx, container.ID())
		require.NoError(t, err)
		assert.Len(t, updatedContainer.Objects(), 0)
	})

	t.Run("error - object not found", func(t *testing.T) {
		ts.CleanDatabase(t)

		collection := ts.CreateTestCollection(t, testUser.ID(), entities.ObjectTypeGeneral)
		container := ts.CreateTestContainer(t, collection.ID())

		fakeObjectID := entities.NewObjectID()

		headers := map[string]string{
			"Authorization": "Bearer test-token",
		}

		resp, _, err := ts.MakeRequest(
			http.MethodDelete,
			"/accounts/"+testUser.ID().String()+"/objects/"+fakeObjectID.String()+"?container_id="+container.ID().String(),
			nil,
			headers,
		)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestObjectAPI_BulkImport(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	testUser := CreateTestUser()
	testClaims := CreateTestAuthClaims(testUser.ID().String())
	ts.SetupDefaultAuthMocks(testUser, testClaims)

	t.Run("success - bulk import objects", func(t *testing.T) {
		ts.CleanDatabase(t)

		collection := ts.CreateTestCollection(t, testUser.ID(), entities.ObjectTypeGeneral)
		container := ts.CreateTestContainer(t, collection.ID())

		reqBody := map[string]interface{}{
			"format":      "json",
			"object_type": "general",
			"data": []map[string]interface{}{
				{
					"name":        "Object 1",
					"description": "First object",
				},
				{
					"name":        "Object 2",
					"description": "Second object",
				},
				{
					"name":        "Object 3",
					"description": "Third object",
				},
			},
			"default_tags": []string{"imported"},
		}

		headers := map[string]string{
			"Authorization": "Bearer test-token",
		}

		resp, body, err := ts.MakeRequest(
			http.MethodPost,
			"/accounts/"+testUser.ID().String()+"/import?container_id="+container.ID().String(),
			reqBody,
			headers,
		)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var importResp response.BulkImportResponse
		err = json.Unmarshal(body, &importResp)
		require.NoError(t, err)

		assert.Equal(t, 3, importResp.Total)
		assert.Equal(t, 3, importResp.Imported)
		assert.Equal(t, 0, importResp.Failed)
	})

	t.Run("error - empty data", func(t *testing.T) {
		ts.CleanDatabase(t)

		collection := ts.CreateTestCollection(t, testUser.ID(), entities.ObjectTypeGeneral)
		container := ts.CreateTestContainer(t, collection.ID())

		reqBody := map[string]interface{}{
			"format":      "json",
			"object_type": "general",
			"data":        []map[string]interface{}{}, // Empty data
		}

		headers := map[string]string{
			"Authorization": "Bearer test-token",
		}

		resp, _, err := ts.MakeRequest(
			http.MethodPost,
			"/accounts/"+testUser.ID().String()+"/import?container_id="+container.ID().String(),
			reqBody,
			headers,
		)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
