package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/nishiki/backend-go/app/http/httputil"
	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/mocks"
)

func TestUserController_GetUser(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	t.Cleanup(mockCtrl.Finish)

	mockAuthService := mocks.NewMockAuthService(mockCtrl)

	controller := &UserController{
		authService: mockAuthService,
		logger:      logger,
	}

	t.Run("success - user found", func(t *testing.T) {
		t.Parallel()

		// Create test user
		username, _ := entities.NewUsername("testuser")
		email, _ := entities.NewEmailAddress("test@example.com")
		testUser := entities.ReconstructUser(
			entities.NewUserID(),
			username,
			email,
			"test-authentik-id",
			time.Now(),
			time.Now(),
		)

		// Create request with path value
		req := newTestRequest(http.MethodGet, "/users/"+testUser.ID().String(), nil)
		req.SetPathValue("id", testUser.ID().String())

		// Set auth context
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()

		// Setup mock expectations
		mockAuthService.EXPECT().
			GetUserByID(gomock.Any(), "test-token", testUser.ID().String()).
			Return(testUser, nil)

		// Call controller method
		controller.GetUser(rr, req)

		// Assert response
		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Equal(t, testUser.ID().String(), response["id"])
		assert.Equal(t, testUser.Username().String(), response["name"])
		assert.Equal(t, testUser.EmailAddress().String(), response["email"])
	})

	t.Run("error - user not found", func(t *testing.T) {
		t.Parallel()

		userID := entities.NewUserID()

		// Create test user for auth context
		username, _ := entities.NewUsername("authuser")
		email, _ := entities.NewEmailAddress("auth@example.com")
		authUser := entities.ReconstructUser(
			entities.NewUserID(),
			username,
			email,
			"auth-authentik-id",
			time.Now(),
			time.Now(),
		)

		// Create request
		req := newTestRequest(http.MethodGet, "/users/"+userID.String(), nil)
		req.SetPathValue("id", userID.String())
		req = setAuthContext(req, authUser, "test-token")

		rr := httptest.NewRecorder()

		// Setup mock expectations
		mockAuthService.EXPECT().
			GetUserByID(gomock.Any(), "test-token", userID.String()).
			Return(nil, errors.New("user not found"))

		// Call controller method
		controller.GetUser(rr, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, rr.Code)

		var response map[string]interface{}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Equal(t, "user not found", response["error"])
	})

	t.Run("error - invalid user ID", func(t *testing.T) {
		t.Parallel()

		// Create test user for auth context
		username, _ := entities.NewUsername("authuser")
		email, _ := entities.NewEmailAddress("auth@example.com")
		authUser := entities.ReconstructUser(
			entities.NewUserID(),
			username,
			email,
			"auth-authentik-id",
			time.Now(),
			time.Now(),
		)

		// Create request with invalid UUID
		req := newTestRequest(http.MethodGet, "/users/invalid-uuid", nil)
		req.SetPathValue("id", "invalid-uuid")
		req = setAuthContext(req, authUser, "test-token")

		rr := httptest.NewRecorder()

		// Call controller method (no mock expectation needed as validation should fail first)
		controller.GetUser(rr, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response map[string]interface{}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Contains(t, response["error"], "invalid")
	})

	t.Run("error - auth service error", func(t *testing.T) {
		t.Parallel()

		// Create test user
		username, _ := entities.NewUsername("testuser")
		email, _ := entities.NewEmailAddress("test@example.com")
		testUser := entities.ReconstructUser(
			entities.NewUserID(),
			username,
			email,
			"test-authentik-id",
			time.Now(),
			time.Now(),
		)

		// Create request
		req := newTestRequest(http.MethodGet, "/users/"+testUser.ID().String(), nil)
		req.SetPathValue("id", testUser.ID().String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()

		// Setup mock expectations
		mockAuthService.EXPECT().
			GetUserByID(gomock.Any(), "test-token", testUser.ID().String()).
			Return(nil, errors.New("internal server error"))

		// Call controller method
		controller.GetUser(rr, req)

		// Assert response
		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		var response map[string]interface{}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Equal(t, "failed to get user", response["error"])
	})
}

func newTestRequest(method, path string, requestBody interface{}) *http.Request {
	var body *bytes.Buffer
	if requestBody != nil {
		reqBodyBytes, err := json.Marshal(requestBody)
		if err != nil {
			panic(err.Error())
		}
		body = bytes.NewBuffer(reqBodyBytes)
	} else {
		body = bytes.NewBuffer([]byte{})
	}

	req, err := http.NewRequestWithContext(context.Background(), method, path, body)
	if err != nil {
		panic(err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}

// setAuthContext sets authentication context values on the request
func setAuthContext(r *http.Request, user *entities.User, token string) *http.Request {
	r = httputil.SetContextValue(r, httputil.AuthUserKey, user)
	r = httputil.SetContextValue(r, httputil.AuthTokenKey, token)
	return r
}

func randomUser() *entities.User {
	username, _ := entities.NewUsername(fake.Username())
	email, _ := entities.NewEmailAddress(fake.Email())
	return entities.ReconstructUser(
		entities.NewUserID(),
		username,
		email,
		fake.UUID(),
		time.Now(),
		time.Now(),
	)
}
