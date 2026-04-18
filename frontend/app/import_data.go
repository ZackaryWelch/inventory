package app

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/nishiki/frontend/pkg/types"
)

// ImportData represents parsed import data.
type ImportData struct {
	Data   []map[string]any
	Format string // "csv" or "json"
	Errors []string
}

// importResult holds the outcome of a completed import for display.
type importResult struct {
	Imported          int
	Failed            int
	Total             int
	ContainersCreated int
}

func (ga *GioApp) dismissImport() {
	ga.showImportPreview = false
	ga.importData = nil
	ga.importFilename = ""
	ga.importRunning = false
	ga.importResult = nil
	ga.importOmittedColumns = nil
	ga.schemaEditorForImport = false
	ga.importSchemaReturnTo = ""
}

// handleImportFileContent processes file content and stores it as import data.
func (ga *GioApp) handleImportFileContent(content string, filename string) {
	ga.logger.Info("Processing import file", "filename", filename)

	var importData *ImportData
	var err error

	if strings.HasSuffix(strings.ToLower(filename), ".csv") {
		importData, err = ga.parseCSV(content)
	} else if strings.HasSuffix(strings.ToLower(filename), ".json") {
		importData, err = ga.parseJSON(content)
	} else {
		ga.logger.Error("Unsupported file format")
		return
	}

	if err != nil {
		ga.logger.Error("Failed to parse import file", "error", err)
		return
	}

	ga.importData = importData
	ga.importFilename = filename
	ga.importOmittedColumns = make(map[string]bool)

	// Initialize column mapping with auto-detected values
	ga.importNameColumn = detectNameColumn(importData.Data)
	if loc := detectLocationColumn(importData.Data); loc != "" {
		ga.importLocationColumn = &loc
	} else {
		ga.importLocationColumn = nil
	}

	if ga.importCreateMode {
		// Import & Create Collection mode
		ga.importCreateMode = false
		ga.showImportCreateDialog = true
		ga.importCreateError = ""
		ga.selectedObjectType = ObjectTypeGeneral
		ga.selectedGroupID = nil
		ga.widgetState.importCreateNameEditor.SetText(filenameToCollectionName(filename))
		ga.widgetState.importCreateLocationEditor.SetText("")
		ga.widgetState.importCreateInferSchemaCheck.Value = true

		// Auto-detect container column
		if col := detectContainerColumn(importData.Data); col != "" {
			ga.importContainerCol = &col
		} else {
			ga.importContainerCol = nil
		}
	} else {
		// Regular import into existing collection
		ga.showImportPreview = true
		ga.widgetState.importInferSchemaCheck.Value = true
	}

	ga.window.Invalidate()
}

// parseCSV parses CSV content into import data.
func (ga *GioApp) parseCSV(content string) (*ImportData, error) {
	reader := csv.NewReader(strings.NewReader(content))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, errors.New("CSV file must have at least a header row and one data row")
	}

	headers := records[0]
	data := &ImportData{
		Data:   make([]map[string]any, 0),
		Format: "csv",
		Errors: make([]string, 0),
	}

	for rowIdx, record := range records[1:] {
		if len(record) != len(headers) {
			data.Errors = append(data.Errors, fmt.Sprintf("Row %d: column count mismatch", rowIdx+2))
			continue
		}

		row := make(map[string]any)
		for i, header := range headers {
			header = strings.TrimSpace(header)
			value := strings.TrimSpace(record[i])
			if value != "" {
				if num, err := strconv.ParseFloat(value, 64); err == nil {
					row[header] = num
				} else {
					row[header] = value
				}
			}
		}

		data.Data = append(data.Data, row)
	}

	return data, nil
}

// parseJSON parses JSON content into import data.
func (ga *GioApp) parseJSON(content string) (*ImportData, error) {
	var rawData []map[string]any
	if err := json.Unmarshal([]byte(content), &rawData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	data := &ImportData{
		Data:   rawData,
		Format: "json",
		Errors: make([]string, 0),
	}

	return data, nil
}

// filterOmittedColumns returns a deep copy of data with omitted columns removed.
// If no columns are omitted, the original slice is returned unchanged.
func filterOmittedColumns(data []map[string]any, omitted map[string]bool) []map[string]any {
	anyOmitted := false
	for _, v := range omitted {
		if v {
			anyOmitted = true
			break
		}
	}
	if !anyOmitted {
		return data
	}
	out := make([]map[string]any, len(data))
	for i, row := range data {
		filtered := make(map[string]any, len(row))
		for k, v := range row {
			if omitted[k] {
				continue
			}
			filtered[k] = v
		}
		out[i] = filtered
	}
	return out
}

// nonOmittedColumns returns the columns sorted, filtered by the omit set.
func nonOmittedColumns(data []map[string]any, omitted map[string]bool) []string {
	all := importColumns(data)
	if len(omitted) == 0 {
		return all
	}
	out := make([]string, 0, len(all))
	for _, c := range all {
		if !omitted[c] {
			out = append(out, c)
		}
	}
	return out
}

// findStringField returns the first string value found for any of the given
// field names, matched case-insensitively against the map keys.
func findStringField(m map[string]any, fields ...string) string {
	for k, v := range m {
		kLower := strings.ToLower(k)
		for _, f := range fields {
			if kLower == f {
				if s, ok := v.(string); ok {
					return s
				}
				return fmt.Sprintf("%v", v)
			}
		}
	}
	return ""
}

// detectColumnByName returns the actual key from the first data row that
// matches one of the given names case-insensitively, checked in order.
func detectColumnByName(data []map[string]any, names ...string) string {
	if len(data) == 0 {
		return ""
	}
	for _, target := range names {
		for key := range data[0] {
			if strings.EqualFold(key, target) {
				return key
			}
		}
	}
	return ""
}

// detectLocationColumn returns the location column name if found in the data headers.
func detectLocationColumn(data []map[string]any) string {
	return detectColumnByName(data, "location")
}

// detectNameColumn returns the name column if "name", "title", or "item" is present.
func detectNameColumn(data []map[string]any) string {
	return detectColumnByName(data, "name", "title", "item")
}

// detectContainerColumn returns the container column if found in the data headers.
func detectContainerColumn(data []map[string]any) string {
	return detectColumnByName(data, "container", "shelf", "room", "box", "bin")
}

// filenameToCollectionName converts a filename to a human-readable collection name.
// e.g., "my_board_games.csv" → "My Board Games"
func filenameToCollectionName(filename string) string {
	// Strip extension
	name := filename
	if idx := strings.LastIndex(name, "."); idx > 0 {
		name = name[:idx]
	}
	// Replace separators with spaces
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "-", " ")
	// Title case
	words := strings.Fields(name)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		}
	}
	return strings.Join(words, " ")
}

// executeImport sends the import request to the backend.
func (ga *GioApp) executeImport() {
	if ga.selectedCollection == nil || ga.importData == nil {
		ga.logger.Error("Cannot execute import: missing collection or data")
		return
	}

	ga.logger.Info("Executing import", "collection_id", ga.selectedCollection.ID, "items", len(ga.importData.Data))
	ga.importRunning = true

	go func() {
		locationCol := ga.importLocationColumn
		nameCol := ga.importNameColumn
		inferSchema := ga.widgetState.importInferSchemaCheck.Value
		filteredData := filterOmittedColumns(ga.importData.Data, ga.importOmittedColumns)
		// schemaChanged covers both inferred and user-supplied schemas; either
		// one means the in-memory collection is now stale and must be refetched.
		schemaChanged := inferSchema

		// A user-defined schema overrides inference: apply it to the collection
		// first, then run the import without inferring.
		if ga.pendingImportSchema != nil {
			inferSchema = false
			schemaChanged = true
			if err := ga.collectionsClient.UpdateSchema(ga.currentUser.ID, ga.selectedCollection.ID, types.UpdatePropertySchemaRequest{
				PropertySchema: *ga.pendingImportSchema,
			}); err != nil {
				ga.do(func() {
					ga.importRunning = false
					if ga.importData != nil {
						ga.importData.Errors = []string{fmt.Sprintf("Failed to save schema: %v", err)}
					}
				})
				return
			}
			ga.pendingImportSchema = nil
		}

		distMode := "automatic"
		if locationCol != nil {
			distMode = "location"
		}

		req := map[string]any{
			"format":            ga.importData.Format,
			"data":              filteredData,
			"distribution_mode": distMode,
			"infer_schema":      inferSchema,
		}
		if locationCol != nil {
			req["location_column"] = *locationCol
		}
		if nameCol != "" {
			req["name_column"] = nameCol
		}

		endpoint := fmt.Sprintf("/accounts/%s/collections/%s/import", ga.currentUser.ID, ga.selectedCollection.ID)
		resp, err := ga.apiClient.Post(endpoint, req)
		if err != nil {
			ga.importRunning = false
			ga.logger.Error("Import failed", "error", err)
			return
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			var errResp struct {
				Error string `json:"error"`
			}
			_ = json.NewDecoder(resp.Body).Decode(&errResp)
			resp.Body.Close()
			ga.importRunning = false
			errMsg := errResp.Error
			if errMsg == "" {
				errMsg = fmt.Sprintf("server error (status %d)", resp.StatusCode)
			}
			ga.logger.Error("Import failed", "error", errMsg)
			ga.importData.Errors = []string{errMsg}
			ga.window.Invalidate()
			return
		}

		var result struct {
			Imported          int      `json:"imported"`
			Failed            int      `json:"failed"`
			Total             int      `json:"total"`
			ContainersCreated int      `json:"containers_created"`
			Errors            []string `json:"errors,omitempty"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			ga.importRunning = false
			ga.logger.Error("Failed to parse import response", "error", err)
			return
		}

		ga.logger.Info("Import completed",
			"imported", result.Imported,
			"failed", result.Failed,
			"total", result.Total,
			"containers_created", result.ContainersCreated)

		ga.importRunning = false
		if result.Failed > 0 {
			for _, errMsg := range result.Errors {
				ga.logger.Warn("Import item failed", "error", errMsg)
			}
			// Keep dialog open and show errors so the user can see what failed
			ga.importData.Data = nil // Clear preview data
			ga.importData.Errors = result.Errors
			ga.importResult = &importResult{
				Imported:          result.Imported,
				Failed:            result.Failed,
				Total:             result.Total,
				ContainersCreated: result.ContainersCreated,
			}
		} else {
			ga.dismissImport()
		}

		// Refetch collection to pick up inferred or user-defined schema
		if schemaChanged {
			userID := ga.currentUser.ID
			collectionID := ga.selectedCollection.ID
			updated, err := ga.collectionsClient.Get(userID, collectionID)
			if err != nil {
				ga.logger.Warn("Failed to refetch collection after import", "error", err)
			} else {
				ga.do(func() {
					ga.selectedCollection = updated
					for i, c := range ga.collections {
						if c.ID == updated.ID {
							ga.collections[i] = *updated
							break
						}
					}
					// Reset sort/group since schema may have changed
					ga.objectSortSpecs = nil
					ga.objectGroupByField = ""
					ga.invalidateObjectCaches()
				})
			}
		}

		ga.fetchContainersAndObjects()
	}()
}
