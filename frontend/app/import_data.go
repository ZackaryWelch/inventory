package app

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ImportData represents parsed import data.
type ImportData struct {
	Data   []map[string]interface{}
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
	ga.showImportPreview = true

	// Initialize column mapping with auto-detected values
	ga.importNameColumn = detectNameColumn(importData.Data)
	if loc := detectLocationColumn(importData.Data); loc != "" {
		ga.importLocationColumn = &loc
	} else {
		ga.importLocationColumn = nil
	}
	ga.widgetState.importInferSchemaCheck.Value = true

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
		return nil, fmt.Errorf("CSV file must have at least a header row and one data row")
	}

	headers := records[0]
	data := &ImportData{
		Data:   make([]map[string]interface{}, 0),
		Format: "csv",
		Errors: make([]string, 0),
	}

	for rowIdx, record := range records[1:] {
		if len(record) != len(headers) {
			data.Errors = append(data.Errors, fmt.Sprintf("Row %d: column count mismatch", rowIdx+2))
			continue
		}

		row := make(map[string]interface{})
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
	var rawData []map[string]interface{}
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

// findStringField returns the first string value found for any of the given
// field names, matched case-insensitively against the map keys.
func findStringField(m map[string]interface{}, fields ...string) string {
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
func detectColumnByName(data []map[string]interface{}, names ...string) string {
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
func detectLocationColumn(data []map[string]interface{}) string {
	return detectColumnByName(data, "location")
}

// detectNameColumn returns the name column if "name", "title", or "item" is present.
func detectNameColumn(data []map[string]interface{}) string {
	return detectColumnByName(data, "name", "title", "item")
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

		distMode := "automatic"
		if locationCol != nil {
			distMode = "location"
		}

		req := map[string]interface{}{
			"format":            ga.importData.Format,
			"data":              ga.importData.Data,
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
		ga.fetchContainersAndObjects()
	}()
}
