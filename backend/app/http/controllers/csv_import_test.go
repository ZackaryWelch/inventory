package controllers

import (
	"encoding/csv"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/nishiki/backend/app/http/request"
	"github.com/nishiki/backend/app/http/response"
	"github.com/nishiki/backend/domain/entities"
	"github.com/nishiki/backend/domain/usecases"
	"github.com/nishiki/backend/mocks"
)

// dataDir returns the absolute path to the project's data/ directory.
func dataDir() string {
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..", "data")
}

// parseCSVFile reads a CSV file and returns rows as []map[string]interface{},
// matching the format expected by the bulk import endpoints.
func parseCSVFile(t *testing.T, path string) []map[string]interface{} {
	t.Helper()
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	reader := csv.NewReader(f)
	reader.LazyQuotes = true
	records, err := reader.ReadAll()
	require.NoError(t, err)
	require.Greater(t, len(records), 1, "CSV must have header + at least one data row")

	headers := records[0]
	var rows []map[string]interface{}
	for _, record := range records[1:] {
		row := make(map[string]interface{}, len(headers))
		for i, header := range headers {
			if i < len(record) {
				row[header] = record[i]
			}
		}
		rows = append(rows, row)
	}
	return rows
}

func newTestCollection(userID entities.UserID, collectionID entities.CollectionID, objectType entities.ObjectType) *entities.Collection {
	collectionName, _ := entities.NewCollectionName("Test Collection")
	return entities.ReconstructCollection(
		collectionID,
		userID,
		nil,
		collectionName,
		nil,
		objectType,
		[]entities.Container{},
		[]string{},
		"",
		nil,
		time.Now(),
		time.Now(),
	)
}

// TestBulkImportToCollection_ElectronicSuppliesCSV tests importing the Electronic_Supplies.csv
// file using location-based distribution. The CSV has a "Location" column with values like
// OD1, OD4, OD5, Desk, Kitchen, Car, KD3 — each should become a container.
func TestBulkImportToCollection_ElectronicSuppliesCSV(t *testing.T) {
	t.Parallel()

	csvPath := filepath.Join(dataDir(), "Electronic_Supplies.csv")
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		t.Skip("test data file not found: ", csvPath)
	}

	data := parseCSVFile(t, csvPath)
	require.NotEmpty(t, data)

	// Filter out rows with empty Name (the CSV has a blank row)
	var filtered []map[string]interface{}
	for _, row := range data {
		if name, _ := row["Name"].(string); name != "" {
			filtered = append(filtered, row)
		}
	}
	data = filtered

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	mockContainerRepo := mocks.NewMockContainerRepository(mockCtrl)
	mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
	mockAuthService := mocks.NewMockAuthService(mockCtrl)

	bulkImportCollectionUC := usecases.NewBulkImportCollectionUseCase(
		mockCollectionRepo, mockContainerRepo, mockAuthService, nil, nil,
	)
	controller := &ObjectController{
		bulkImportCollectionUC: bulkImportCollectionUC,
		logger:                 logger,
	}

	t.Run("location distribution", func(t *testing.T) {
		testUser := randomUser()
		collectionID := entities.NewCollectionID()
		testCollection := newTestCollection(testUser.ID(), collectionID, entities.ObjectTypeGeneral)

		// Mock auth and collection lookup
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", testUser.ID().String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(testCollection, nil)

		// Location distribution fetches existing containers (none initially)
		mockContainerRepo.EXPECT().GetByCollectionID(gomock.Any(), collectionID).Return([]*entities.Container{}, nil)

		// Expect container creation for each unique location + Default container
		// Locations in CSV: OD1, OD4, OD5, Desk, Kitchen, Car, KD3
		mockContainerRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		// Collection updated after containers created
		mockCollectionRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		// Dirty containers saved after objects added
		mockContainerRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		requestBody := request.BulkImportCollectionRequest{
			Format:           "csv",
			DistributionMode: "location",
			Data:             data,
			LocationColumn:   "Location",
			NameColumn:       "Name",
		}

		req := newTestRequest(http.MethodPost,
			"/accounts/"+testUser.ID().String()+"/collections/"+collectionID.String()+"/import",
			requestBody,
		)
		req.SetPathValue("id", testUser.ID().String())
		req.SetPathValue("collection_id", collectionID.String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.BulkImportToCollection(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp response.BulkImportResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.Equal(t, len(data), resp.Total, "total should match input rows")
		assert.Equal(t, len(data), resp.Imported, "all valid rows should import")
		assert.Equal(t, 0, resp.Failed, "no rows should fail")
		assert.Empty(t, resp.Errors)
	})

	t.Run("default distribution", func(t *testing.T) {
		testUser := randomUser()
		collectionID := entities.NewCollectionID()
		testCollection := newTestCollection(testUser.ID(), collectionID, entities.ObjectTypeGeneral)

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", testUser.ID().String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(testCollection, nil)

		// Default distribution creates a "Default Container" since collection has no containers
		mockCollectionRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		mockContainerRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		requestBody := request.BulkImportCollectionRequest{
			Format:     "csv",
			Data:       data,
			NameColumn: "Name",
		}

		req := newTestRequest(http.MethodPost,
			"/accounts/"+testUser.ID().String()+"/collections/"+collectionID.String()+"/import",
			requestBody,
		)
		req.SetPathValue("id", testUser.ID().String())
		req.SetPathValue("collection_id", collectionID.String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.BulkImportToCollection(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp response.BulkImportResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.Equal(t, len(data), resp.Total)
		assert.Equal(t, len(data), resp.Imported)
		assert.Equal(t, 0, resp.Failed)
	})

	t.Run("with schema inference", func(t *testing.T) {
		testUser := randomUser()
		collectionID := entities.NewCollectionID()
		testCollection := newTestCollection(testUser.ID(), collectionID, entities.ObjectTypeGeneral)

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", testUser.ID().String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(testCollection, nil)

		// Schema inference triggers collection update for schema save, then again for container add
		mockCollectionRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		mockContainerRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		requestBody := request.BulkImportCollectionRequest{
			Format:      "csv",
			Data:        data,
			NameColumn:  "Name",
			InferSchema: true,
		}

		req := newTestRequest(http.MethodPost,
			"/accounts/"+testUser.ID().String()+"/collections/"+collectionID.String()+"/import",
			requestBody,
		)
		req.SetPathValue("id", testUser.ID().String())
		req.SetPathValue("collection_id", collectionID.String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.BulkImportToCollection(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp response.BulkImportResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.Equal(t, len(data), resp.Total)
		assert.Equal(t, len(data), resp.Imported)
		assert.Equal(t, 0, resp.Failed)
	})
}

// TestBulkImportToCollection_LibibBooksCSV tests importing a Libib library export
// (books). The CSV uses "title" for the name column, which exercises auto-detection.
func TestBulkImportToCollection_LibibBooksCSV(t *testing.T) {
	t.Parallel()

	csvPath := filepath.Join(dataDir(), "libib", "library_20250705_145919.csv")
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		t.Skip("test data file not found: ", csvPath)
	}

	data := parseCSVFile(t, csvPath)
	require.NotEmpty(t, data)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	mockContainerRepo := mocks.NewMockContainerRepository(mockCtrl)
	mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
	mockAuthService := mocks.NewMockAuthService(mockCtrl)

	bulkImportCollectionUC := usecases.NewBulkImportCollectionUseCase(
		mockCollectionRepo, mockContainerRepo, mockAuthService, nil, nil,
	)
	controller := &ObjectController{
		bulkImportCollectionUC: bulkImportCollectionUC,
		logger:                 logger,
	}

	t.Run("default distribution with title auto-detect", func(t *testing.T) {
		testUser := randomUser()
		collectionID := entities.NewCollectionID()
		testCollection := newTestCollection(testUser.ID(), collectionID, entities.ObjectTypeBook)

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", testUser.ID().String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(testCollection, nil)

		// Default distribution path
		mockCollectionRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		mockContainerRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		requestBody := request.BulkImportCollectionRequest{
			Format:      "csv",
			Data:        data,
			DefaultTags: []string{"libib-import"},
		}

		req := newTestRequest(http.MethodPost,
			"/accounts/"+testUser.ID().String()+"/collections/"+collectionID.String()+"/import",
			requestBody,
		)
		req.SetPathValue("id", testUser.ID().String())
		req.SetPathValue("collection_id", collectionID.String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.BulkImportToCollection(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp response.BulkImportResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.Equal(t, len(data), resp.Total)
		assert.Equal(t, len(data), resp.Imported, "all books should import successfully")
		assert.Equal(t, 0, resp.Failed)
	})

	t.Run("with schema inference", func(t *testing.T) {
		testUser := randomUser()
		collectionID := entities.NewCollectionID()
		testCollection := newTestCollection(testUser.ID(), collectionID, entities.ObjectTypeBook)

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", testUser.ID().String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(testCollection, nil)
		mockCollectionRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		mockContainerRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		requestBody := request.BulkImportCollectionRequest{
			Format:      "csv",
			Data:        data,
			InferSchema: true,
		}

		req := newTestRequest(http.MethodPost,
			"/accounts/"+testUser.ID().String()+"/collections/"+collectionID.String()+"/import",
			requestBody,
		)
		req.SetPathValue("id", testUser.ID().String())
		req.SetPathValue("collection_id", collectionID.String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.BulkImportToCollection(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp response.BulkImportResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.Equal(t, len(data), resp.Total)
		assert.Equal(t, len(data), resp.Imported)
		assert.Equal(t, 0, resp.Failed)
	})
}

// TestBulkImportToCollection_LibibVideoGamesCSV tests importing a Libib videogame library export.
func TestBulkImportToCollection_LibibVideoGamesCSV(t *testing.T) {
	t.Parallel()

	csvPath := filepath.Join(dataDir(), "libib", "library_20250705_145938.csv")
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		t.Skip("test data file not found: ", csvPath)
	}

	data := parseCSVFile(t, csvPath)
	require.NotEmpty(t, data)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	mockContainerRepo := mocks.NewMockContainerRepository(mockCtrl)
	mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
	mockAuthService := mocks.NewMockAuthService(mockCtrl)

	bulkImportCollectionUC := usecases.NewBulkImportCollectionUseCase(
		mockCollectionRepo, mockContainerRepo, mockAuthService, nil, nil,
	)
	controller := &ObjectController{
		bulkImportCollectionUC: bulkImportCollectionUC,
		logger:                 logger,
	}

	t.Run("default distribution", func(t *testing.T) {
		testUser := randomUser()
		collectionID := entities.NewCollectionID()
		testCollection := newTestCollection(testUser.ID(), collectionID, entities.ObjectTypeVideoGame)

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", testUser.ID().String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(testCollection, nil)
		mockCollectionRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		mockContainerRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		requestBody := request.BulkImportCollectionRequest{
			Format: "csv",
			Data:   data,
		}

		req := newTestRequest(http.MethodPost,
			"/accounts/"+testUser.ID().String()+"/collections/"+collectionID.String()+"/import",
			requestBody,
		)
		req.SetPathValue("id", testUser.ID().String())
		req.SetPathValue("collection_id", collectionID.String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.BulkImportToCollection(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp response.BulkImportResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.Equal(t, len(data), resp.Total)
		assert.Equal(t, len(data), resp.Imported)
		assert.Equal(t, 0, resp.Failed)
	})
}

// TestBulkImportToCollection_LibibMusicCSV tests importing a Libib music library export.
func TestBulkImportToCollection_LibibMusicCSV(t *testing.T) {
	t.Parallel()

	csvPath := filepath.Join(dataDir(), "libib", "library_20250705_145950.csv")
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		t.Skip("test data file not found: ", csvPath)
	}

	data := parseCSVFile(t, csvPath)
	require.NotEmpty(t, data)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	mockContainerRepo := mocks.NewMockContainerRepository(mockCtrl)
	mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
	mockAuthService := mocks.NewMockAuthService(mockCtrl)

	bulkImportCollectionUC := usecases.NewBulkImportCollectionUseCase(
		mockCollectionRepo, mockContainerRepo, mockAuthService, nil, nil,
	)
	controller := &ObjectController{
		bulkImportCollectionUC: bulkImportCollectionUC,
		logger:                 logger,
	}

	t.Run("default distribution", func(t *testing.T) {
		testUser := randomUser()
		collectionID := entities.NewCollectionID()
		testCollection := newTestCollection(testUser.ID(), collectionID, entities.ObjectTypeMusic)

		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", testUser.ID().String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(testCollection, nil)
		mockCollectionRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		mockContainerRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		requestBody := request.BulkImportCollectionRequest{
			Format: "csv",
			Data:   data,
		}

		req := newTestRequest(http.MethodPost,
			"/accounts/"+testUser.ID().String()+"/collections/"+collectionID.String()+"/import",
			requestBody,
		)
		req.SetPathValue("id", testUser.ID().String())
		req.SetPathValue("collection_id", collectionID.String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.BulkImportToCollection(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp response.BulkImportResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.Equal(t, len(data), resp.Total)
		assert.Equal(t, len(data), resp.Imported)
		assert.Equal(t, 0, resp.Failed)
	})
}

// TestBulkImport_ElectronicSuppliesCSV tests the simpler BulkImport endpoint
// (import to a specific container) using the Electronic_Supplies.csv data.
func TestBulkImport_ElectronicSuppliesCSV(t *testing.T) {
	t.Parallel()

	csvPath := filepath.Join(dataDir(), "Electronic_Supplies.csv")
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		t.Skip("test data file not found: ", csvPath)
	}

	data := parseCSVFile(t, csvPath)
	require.NotEmpty(t, data)

	// Filter out blank-name rows
	var filtered []map[string]interface{}
	for _, row := range data {
		if name, _ := row["Name"].(string); name != "" {
			filtered = append(filtered, row)
		}
	}
	data = filtered

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	mockContainerRepo := mocks.NewMockContainerRepository(mockCtrl)
	mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
	mockAuthService := mocks.NewMockAuthService(mockCtrl)

	bulkImportUC := usecases.NewBulkImportObjectsUseCase(mockContainerRepo, mockCollectionRepo, mockAuthService, nil)
	controller := &ObjectController{
		bulkImportUC: bulkImportUC,
		logger:       logger,
	}

	t.Run("import all to single container", func(t *testing.T) {
		testUser := randomUser()
		collectionID := entities.NewCollectionID()
		containerID := entities.NewContainerID()

		// Rename "Name" key to "name" so the BulkImport controller can find it
		normalizedData := make([]map[string]interface{}, len(data))
		for i, row := range data {
			normalized := make(map[string]interface{}, len(row))
			for k, v := range row {
				if k == "Name" {
					normalized["name"] = v
				} else {
					normalized[k] = v
				}
			}
			normalizedData[i] = normalized
		}

		containerName, _ := entities.NewContainerName("Electronics Box")
		testContainer, _ := entities.NewContainer(entities.ContainerProps{
			CollectionID: collectionID,
			Name:         containerName,
		})

		testCollection := newTestCollection(testUser.ID(), collectionID, entities.ObjectTypeGeneral)

		mockContainerRepo.EXPECT().GetByID(gomock.Any(), containerID).Return(testContainer, nil)
		mockAuthService.EXPECT().GetUserGroups(gomock.Any(), "test-token", testUser.ID().String()).Return([]*entities.Group{}, nil)
		mockCollectionRepo.EXPECT().GetByID(gomock.Any(), collectionID).Return(testCollection, nil)
		mockContainerRepo.EXPECT().Update(gomock.Any(), testContainer).Return(nil)

		requestBody := request.BulkImportRequest{
			ContainerID: containerID.String(),
			Format:      "csv",
			ObjectType:  "general",
			Data:        normalizedData,
			DefaultTags: []string{"electronics"},
		}

		req := newTestRequest(http.MethodPost,
			"/accounts/"+testUser.ID().String()+"/import",
			requestBody,
		)
		req.SetPathValue("id", testUser.ID().String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.BulkImport(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp response.BulkImportResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.Equal(t, len(normalizedData), resp.Total)
		assert.Equal(t, len(normalizedData), resp.Imported)
		assert.Equal(t, 0, resp.Failed)
	})
}

// TestBulkImportToCollection_ValidationErrors tests that the controller returns
// proper error responses for invalid requests.
func TestBulkImportToCollection_ValidationErrors(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	mockContainerRepo := mocks.NewMockContainerRepository(mockCtrl)
	mockCollectionRepo := mocks.NewMockCollectionRepository(mockCtrl)
	mockAuthService := mocks.NewMockAuthService(mockCtrl)

	bulkImportCollectionUC := usecases.NewBulkImportCollectionUseCase(
		mockCollectionRepo, mockContainerRepo, mockAuthService, nil, nil,
	)
	controller := &ObjectController{
		bulkImportCollectionUC: bulkImportCollectionUC,
		logger:                 logger,
	}

	t.Run("empty data", func(t *testing.T) {
		testUser := randomUser()
		collectionID := entities.NewCollectionID()

		requestBody := request.BulkImportCollectionRequest{
			Format: "csv",
			Data:   []map[string]interface{}{},
		}

		req := newTestRequest(http.MethodPost,
			"/accounts/"+testUser.ID().String()+"/collections/"+collectionID.String()+"/import",
			requestBody,
		)
		req.SetPathValue("id", testUser.ID().String())
		req.SetPathValue("collection_id", collectionID.String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.BulkImportToCollection(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid format", func(t *testing.T) {
		testUser := randomUser()
		collectionID := entities.NewCollectionID()

		requestBody := request.BulkImportCollectionRequest{
			Format: "xml",
			Data:   []map[string]interface{}{{"name": "test"}},
		}

		req := newTestRequest(http.MethodPost,
			"/accounts/"+testUser.ID().String()+"/collections/"+collectionID.String()+"/import",
			requestBody,
		)
		req.SetPathValue("id", testUser.ID().String())
		req.SetPathValue("collection_id", collectionID.String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.BulkImportToCollection(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid distribution mode", func(t *testing.T) {
		testUser := randomUser()
		collectionID := entities.NewCollectionID()

		requestBody := request.BulkImportCollectionRequest{
			Format:           "csv",
			DistributionMode: "invalid",
			Data:             []map[string]interface{}{{"name": "test"}},
		}

		req := newTestRequest(http.MethodPost,
			"/accounts/"+testUser.ID().String()+"/collections/"+collectionID.String()+"/import",
			requestBody,
		)
		req.SetPathValue("id", testUser.ID().String())
		req.SetPathValue("collection_id", collectionID.String())
		req = setAuthContext(req, testUser, "test-token")

		rr := httptest.NewRecorder()
		controller.BulkImportToCollection(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
