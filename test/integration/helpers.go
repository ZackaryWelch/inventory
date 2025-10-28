package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/mock/gomock"

	"github.com/nishiki/backend-go/app/config"
	appContainer "github.com/nishiki/backend-go/app/container"
	"github.com/nishiki/backend-go/app/http/routes"
	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/services"
	"github.com/nishiki/backend-go/external/adapters"
	"github.com/nishiki/backend-go/external/repositories"
	"github.com/nishiki/backend-go/mocks"
)

// TestServer holds the test server and dependencies
type TestServer struct {
	Server         *httptest.Server
	Router         *gin.Engine
	Container      *appContainer.Container
	MongoDB        *adapters.MongoDatabase
	MongoContainer testcontainers.Container
	MockAuth       *mocks.MockAuthService
	Logger         *slog.Logger
}

// SetupTestServer creates a test server with MongoDB testcontainer
func SetupTestServer(t *testing.T) *TestServer {
	ctx := context.Background()

	// Start MongoDB testcontainer
	mongoContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "mongo:7.0",
			ExposedPorts: []string{"27017/tcp"},
			WaitingFor:   wait.ForLog("Waiting for connections"),
		},
		Started: true,
	})
	require.NoError(t, err)

	// Get MongoDB connection details
	mongoHost, err := mongoContainer.Host(ctx)
	require.NoError(t, err)

	mongoPort, err := mongoContainer.MappedPort(ctx, "27017")
	require.NoError(t, err)

	mongoURI := fmt.Sprintf("mongodb://%s:%s", mongoHost, mongoPort.Port())

	// Setup MongoDB connection
	dbConfig := config.DatabaseConfig{
		URI:      mongoURI,
		Database: "nishiki_test",
		Timeout:  10,
	}

	mongoDB := adapters.NewMongoDatabase(dbConfig)
	err = mongoDB.Connect(ctx)
	require.NoError(t, err)

	// Setup logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Setup repositories
	collectionRepo := repositories.NewMongoCollectionRepository(mongoDB)
	containerRepo := repositories.NewMongoContainerRepository(mongoDB)
	categoryRepo := repositories.NewMongoCategoryRepository(mongoDB)

	// Create a mock controller for auth service (will be set per test)
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)
	mockAuth := mocks.NewMockAuthService(mockCtrl)

	// Setup container - manually set public fields for testing
	container := &appContainer.Container{
		CollectionRepo: collectionRepo,
		ContainerRepo:  containerRepo,
		CategoryRepo:   categoryRepo,
		AuthService:    mockAuth,
	}

	// Set logger for testing (routes.Setup needs this)
	container.SetLogger(logger)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	routes.Setup(router, container)

	// Create test server
	server := httptest.NewServer(router)

	return &TestServer{
		Server:         server,
		Router:         router,
		Container:      container,
		MongoDB:        mongoDB,
		MongoContainer: mongoContainer,
		MockAuth:       mockAuth,
		Logger:         logger,
	}
}

// Teardown cleans up test resources
func (ts *TestServer) Teardown(t *testing.T) {
	ctx := context.Background()

	// Close server
	if ts.Server != nil {
		ts.Server.Close()
	}

	// Disconnect from MongoDB
	if ts.MongoDB != nil {
		err := ts.MongoDB.Disconnect(ctx)
		require.NoError(t, err)
	}

	// Stop MongoDB container
	if ts.MongoContainer != nil {
		err := ts.MongoContainer.Terminate(ctx)
		require.NoError(t, err)
	}
}

// CleanDatabase removes all data from test database
func (ts *TestServer) CleanDatabase(t *testing.T) {
	ctx := context.Background()
	db := ts.MongoDB.Database()

	collections, err := db.ListCollectionNames(ctx, map[string]interface{}{})
	require.NoError(t, err)

	for _, collection := range collections {
		err := db.Collection(collection).Drop(ctx)
		require.NoError(t, err)
	}
}

// MakeRequest makes an HTTP request to the test server
func (ts *TestServer) MakeRequest(method, path string, body interface{}, headers map[string]string) (*http.Response, []byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, ts.Server.URL+path, reqBody)
	if err != nil {
		return nil, nil, err
	}

	// Set default headers
	req.Header.Set("Content-Type", "application/json")

	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	return resp, responseBody, nil
}

// CreateTestUser creates a test user entity
func CreateTestUser() *entities.User {
	username, _ := entities.NewUsername("testuser")
	email, _ := entities.NewEmailAddress("test@example.com")
	return entities.ReconstructUser(
		entities.NewUserID(),
		username,
		email,
		"test-authentik-id",
		time.Now(),
		time.Now(),
	)
}

// CreateTestAuthClaims creates test auth claims
func CreateTestAuthClaims(userID string) *services.AuthClaims {
	return &services.AuthClaims{
		Subject:  userID,
		Email:    "test@example.com",
		Username: "testuser",
		Groups:   []string{},
		Name:     "Test User",
	}
}

// SetupDefaultAuthMocks sets up the default authentication mock expectations
func (ts *TestServer) SetupDefaultAuthMocks(testUser *entities.User, testClaims *services.AuthClaims) {
	ts.MockAuth.EXPECT().
		ValidateToken(gomock.Any(), gomock.Any()).
		Return(testClaims, nil).
		AnyTimes()
	ts.MockAuth.EXPECT().
		GetUserFromClaims(gomock.Any(), gomock.Any()).
		Return(testUser, nil).
		AnyTimes()
	ts.MockAuth.EXPECT().
		GetUserGroups(gomock.Any(), gomock.Any(), testUser.ID().String()).
		Return([]*entities.Group{}, nil).
		AnyTimes()
}

// CreateTestCollection creates and saves a test collection
func (ts *TestServer) CreateTestCollection(t *testing.T, userID entities.UserID, objectType entities.ObjectType) *entities.Collection {
	ctx := context.Background()

	collectionName, err := entities.NewCollectionName("Test Collection")
	require.NoError(t, err)

	collection, err := entities.NewCollection(entities.CollectionProps{
		UserID:     userID,
		GroupID:    nil,
		Name:       collectionName,
		ObjectType: objectType,
		Tags:       []string{"test"},
		Location:   "Test Location",
	})
	require.NoError(t, err)

	err = ts.Container.CollectionRepo.Create(ctx, collection)
	require.NoError(t, err)

	// Verify collection was saved by retrieving it
	retrieved, err := ts.Container.CollectionRepo.GetByID(ctx, collection.ID())
	if err != nil {
		t.Fatalf("Failed to retrieve newly created collection: %v", err)
	}
	if retrieved == nil {
		t.Fatal("Retrieved collection is nil")
	}

	return collection
}

// CreateTestContainer creates and saves a test container
func (ts *TestServer) CreateTestContainer(t *testing.T, collectionID entities.CollectionID) *entities.Container {
	ctx := context.Background()

	containerName, err := entities.NewContainerName("Test Container")
	require.NoError(t, err)

	container, err := entities.NewContainer(entities.ContainerProps{
		CollectionID: collectionID,
		Name:         containerName,
	})
	require.NoError(t, err)

	// Save container to container repository
	err = ts.Container.ContainerRepo.Create(ctx, container)
	require.NoError(t, err)

	// Get collection and add container
	collection, err := ts.Container.CollectionRepo.GetByID(ctx, collectionID)
	require.NoError(t, err)

	err = collection.AddContainer(*container)
	require.NoError(t, err)

	err = ts.Container.CollectionRepo.Update(ctx, collection)
	require.NoError(t, err)

	return container
}

// CreateTestObject creates a test object entity (not saved)
func CreateTestObject(name string) (*entities.Object, error) {
	objectName, err := entities.NewObjectName(name)
	if err != nil {
		return nil, err
	}

	return entities.NewObject(entities.ObjectProps{
		Name:       objectName,
		ObjectType: entities.ObjectTypeGeneral,
		Properties: map[string]interface{}{"description": "Test object"},
		Tags:       []string{"test"},
	})
}
