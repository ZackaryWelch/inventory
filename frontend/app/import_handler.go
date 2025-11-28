//go:build js && wasm

package app

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"syscall/js"
)

// ImportData represents parsed import data
type ImportData struct {
	Data   []map[string]interface{}
	Format string // "csv" or "json"
	Errors []string
}

// SelectImportFile opens a file picker dialog and returns the file content
func (ga *GioApp) SelectImportFile() {
	// Create file input element
	input := js.Global().Get("document").Call("createElement", "input")
	input.Set("type", "file")
	input.Set("accept", ".csv,.json")

	// Create change handler
	var changeHandler js.Func
	changeHandler = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		files := input.Get("files")
		if files.Length() == 0 {
			changeHandler.Release()
			return nil
		}

		file := files.Index(0)
		filename := file.Get("name").String()

		// Create FileReader
		reader := js.Global().Get("FileReader").New()

		// Set up load handler
		var loadHandler js.Func
		loadHandler = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			result := reader.Get("result").String()
			ga.handleImportFileContent(result, filename)
			loadHandler.Release()
			changeHandler.Release()
			return nil
		})

		// Set up error handler
		var errorHandler js.Func
		errorHandler = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			ga.logger.Error("Failed to read import file")
			errorHandler.Release()
			loadHandler.Release()
			changeHandler.Release()
			return nil
		})

		reader.Set("onload", loadHandler)
		reader.Set("onerror", errorHandler)

		// Read file as text
		reader.Call("readAsText", file)
		return nil
	})

	input.Call("addEventListener", "change", changeHandler)
	input.Call("click")
}

// handleImportFileContent processes the imported file content
func (ga *GioApp) handleImportFileContent(content string, filename string) {
	ga.logger.Info("Processing import file", "filename", filename)

	var importData *ImportData
	var err error

	// Determine file type and parse
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

	// Store import data and show preview
	ga.importData = importData
	ga.importFilename = filename
	ga.showImportPreview = true
	ga.window.Invalidate()
}

// parseCSV parses CSV content into import data
func (ga *GioApp) parseCSV(content string) (*ImportData, error) {
	reader := csv.NewReader(strings.NewReader(content))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file must have at least a header row and one data row")
	}

	// Parse header
	headers := records[0]

	// Parse rows
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
				// Try to parse as number
				if num, err := strconv.ParseFloat(value, 64); err == nil {
					row[header] = num
				} else {
					row[header] = value
				}
			}
		}

		// Validate required fields
		if _, hasName := row["name"]; !hasName {
			if _, hasTitle := row["title"]; !hasTitle {
				if _, hasItem := row["item"]; !hasItem {
					data.Errors = append(data.Errors, fmt.Sprintf("Row %d: missing required field 'name'", rowIdx+2))
					continue
				}
			}
		}

		data.Data = append(data.Data, row)
	}

	return data, nil
}

// parseJSON parses JSON content into import data
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

	// Validate rows
	for i, row := range rawData {
		if _, hasName := row["name"]; !hasName {
			if _, hasTitle := row["title"]; !hasTitle {
				if _, hasItem := row["item"]; !hasItem {
					data.Errors = append(data.Errors, fmt.Sprintf("Item %d: missing required field 'name'", i+1))
				}
			}
		}
	}

	return data, nil
}

// executeImport sends the import request to the backend
func (ga *GioApp) executeImport() {
	if ga.selectedCollection == nil || ga.importData == nil {
		ga.logger.Error("Cannot execute import: missing collection or data")
		return
	}

	ga.logger.Info("Executing import", "collection_id", ga.selectedCollection.ID, "items", len(ga.importData.Data))

	go func() {
		// Prepare import request
		req := map[string]interface{}{
			"format":             ga.importData.Format,
			"data":               ga.importData.Data,
			"distribution_mode":  "automatic",
			"target_container_id": nil,
		}

		// Call import API
		endpoint := fmt.Sprintf("/accounts/%s/collections/%s/import", ga.currentUser.ID, ga.selectedCollection.ID)
		resp, err := ga.apiClient.Post(endpoint, req)
		if err != nil {
			ga.logger.Error("Import failed", "error", err)
			return
		}

		// Parse response
		var result struct {
			Imported int      `json:"imported"`
			Failed   int      `json:"failed"`
			Total    int      `json:"total"`
			Errors   []string `json:"errors,omitempty"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			ga.logger.Error("Failed to parse import response", "error", err)
			return
		}

		ga.logger.Info("Import completed", "imported", result.Imported, "failed", result.Failed, "total", result.Total)

		// Close preview and refresh data
		ga.showImportPreview = false
		ga.importData = nil
		ga.importFilename = ""
		ga.fetchContainersAndObjects()
		ga.window.Invalidate()
	}()
}
