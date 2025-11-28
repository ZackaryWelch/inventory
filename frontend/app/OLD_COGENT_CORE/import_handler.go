//go:build js && wasm

package app

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"syscall/js"

	"github.com/nishiki/frontend/pkg/types"
)

// ImportHandler handles file uploads and data parsing for imports
type ImportHandler struct {
	app *App
}

// NewImportHandler creates a new import handler
func NewImportHandler(app *App) *ImportHandler {
	return &ImportHandler{app: app}
}

// ImportData represents parsed import data
type ImportData struct {
	Objects []types.CreateObjectRequest
	Format  string // "csv" or "json"
	Errors  []string
}

// SelectFile opens a file picker dialog and returns the file content
func (h *ImportHandler) SelectFile(callback func(content string, filename string, err error)) {
	// Create file input element
	input := js.Global().Get("document").Call("createElement", "input")
	input.Set("type", "file")
	input.Set("accept", ".csv,.json")

	// Create change handler - will be released after file is processed
	var changeHandler js.Func
	changeHandler = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		files := input.Get("files")
		if files.Length() == 0 {
			callback("", "", fmt.Errorf("no file selected"))
			changeHandler.Release()
			return nil
		}

		file := files.Index(0)
		filename := file.Get("name").String()

		// Create FileReader
		reader := js.Global().Get("FileReader").New()

		// Set up load handler - release after callback executes
		var loadHandler js.Func
		loadHandler = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			result := reader.Get("result").String()
			callback(result, filename, nil)
			// Release handlers after successful load
			loadHandler.Release()
			changeHandler.Release()
			return nil
		})

		// Set up error handler - release after callback executes
		var errorHandler js.Func
		errorHandler = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			callback("", filename, fmt.Errorf("failed to read file"))
			// Release handlers after error
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

// ParseCSV parses CSV content into import data
func (h *ImportHandler) ParseCSV(content string) (*ImportData, error) {
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
	headerMap := make(map[string]int)
	for i, header := range headers {
		headerMap[strings.ToLower(strings.TrimSpace(header))] = i
	}

	// Parse rows
	data := &ImportData{
		Objects: make([]types.CreateObjectRequest, 0),
		Format:  "csv",
		Errors:  make([]string, 0),
	}

	for rowIdx, record := range records[1:] {
		obj, err := h.parseCSVRow(record, headerMap)
		if err != nil {
			data.Errors = append(data.Errors, fmt.Sprintf("Row %d: %v", rowIdx+2, err))
			continue
		}
		data.Objects = append(data.Objects, *obj)
	}

	return data, nil
}

// parseCSVRow parses a single CSV row into a CreateObjectRequest
func (h *ImportHandler) parseCSVRow(record []string, headerMap map[string]int) (*types.CreateObjectRequest, error) {
	obj := &types.CreateObjectRequest{
		Properties: make(map[string]interface{}),
		Tags:       make([]string, 0),
	}

	// Get name (required) - try multiple field names
	nameFound := false
	nameFields := []string{"name", "title", "item"}
	for _, field := range nameFields {
		if idx, ok := headerMap[field]; ok && idx < len(record) {
			if name := strings.TrimSpace(record[idx]); name != "" {
				obj.Name = name
				nameFound = true
				break
			}
		}
	}

	if !nameFound {
		return nil, fmt.Errorf("missing required field 'name', 'title', or 'item'")
	}

	// Get description - try multiple field names
	descFields := []string{"description", "notes", "summary"}
	for _, field := range descFields {
		if idx, ok := headerMap[field]; ok && idx < len(record) {
			if desc := strings.TrimSpace(record[idx]); desc != "" {
				obj.Description = desc
				break
			}
		}
	}

	// Get quantity - try multiple field names
	qtyFields := []string{"quantity", "copies", "amount"}
	for _, field := range qtyFields {
		if idx, ok := headerMap[field]; ok && idx < len(record) {
			if qtyStr := strings.TrimSpace(record[idx]); qtyStr != "" {
				qty, err := strconv.ParseFloat(qtyStr, 64)
				if err == nil {
					obj.Quantity = &qty
					break
				}
			}
		}
	}

	// Get unit
	if idx, ok := headerMap["unit"]; ok && idx < len(record) {
		obj.Unit = strings.TrimSpace(record[idx])
	}

	// Get tags - try multiple field names and formats
	tagsFields := []string{"tags", "tag", "categories"}
	for _, field := range tagsFields {
		if idx, ok := headerMap[field]; ok && idx < len(record) {
			tagsStr := strings.TrimSpace(record[idx])
			if tagsStr != "" {
				// Split by comma, semicolon, or pipe
				var tags []string
				if strings.Contains(tagsStr, ",") {
					tags = strings.Split(tagsStr, ",")
				} else if strings.Contains(tagsStr, ";") {
					tags = strings.Split(tagsStr, ";")
				} else if strings.Contains(tagsStr, "|") {
					tags = strings.Split(tagsStr, "|")
				} else {
					tags = []string{tagsStr}
				}

				for _, tag := range tags {
					if trimmed := strings.TrimSpace(tag); trimmed != "" {
						obj.Tags = append(obj.Tags, trimmed)
					}
				}
				break
			}
		}
	}

	// Define fields to skip when adding to properties
	skipFields := map[string]bool{
		"name":        true,
		"title":       true,
		"item":        true,
		"description": true,
		"notes":       true,
		"summary":     true,
		"quantity":    true,
		"copies":      true,
		"amount":      true,
		"unit":        true,
		"tags":        true,
		"tag":         true,
		"categories":  true,
		"category":    true,
		"id":          true, // Skip ID fields from source data
		"_id":         true,
	}

	// Parse all other columns as properties
	for header, idx := range headerMap {
		if idx >= len(record) {
			continue
		}

		// Skip already handled fields
		if skipFields[header] {
			continue
		}

		value := strings.TrimSpace(record[idx])
		if value != "" {
			obj.Properties[header] = value
		}
	}

	return obj, nil
}

// ParseJSON parses JSON content into import data
func (h *ImportHandler) ParseJSON(content string) (*ImportData, error) {
	var objects []types.CreateObjectRequest
	if err := json.Unmarshal([]byte(content), &objects); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if len(objects) == 0 {
		return nil, fmt.Errorf("JSON file must contain at least one object")
	}

	data := &ImportData{
		Objects: objects,
		Format:  "json",
		Errors:  make([]string, 0),
	}

	// Validate objects
	for i, obj := range objects {
		if obj.Name == "" {
			data.Errors = append(data.Errors, fmt.Sprintf("Object %d: name is required", i+1))
		}
	}

	return data, nil
}

// Parse automatically detects format and parses content
func (h *ImportHandler) Parse(content string, filename string) (*ImportData, error) {
	// Detect format from filename
	if strings.HasSuffix(strings.ToLower(filename), ".json") {
		return h.ParseJSON(content)
	} else if strings.HasSuffix(strings.ToLower(filename), ".csv") {
		return h.ParseCSV(content)
	}

	// Try to auto-detect
	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, "{") {
		return h.ParseJSON(content)
	}

	// Default to CSV
	return h.ParseCSV(content)
}

// ImportToContainer imports objects to a specific container
func (h *ImportHandler) ImportToContainer(containerID string, objects []types.CreateObjectRequest) error {
	h.app.logger.Info("ImportToContainer called",
		"containerID", containerID,
		"collectionID", h.app.selectedCollection.ID,
		"objectCount", len(objects))

	// Convert CreateObjectRequest to map[string]interface{} for backend
	data := make([]map[string]interface{}, len(objects))
	for i, obj := range objects {
		data[i] = map[string]interface{}{
			"name":        obj.Name,
			"description": obj.Description,
			"quantity":    obj.Quantity,
			"unit":        obj.Unit,
			"tags":        obj.Tags,
		}
		// Add properties
		for key, value := range obj.Properties {
			data[i][key] = value
		}
	}

	h.app.logger.Info("Sample object data", "first", data[0])

	// Call backend bulk import API using backend request type
	req := types.BulkImportCollectionRequest{
		CollectionID:      h.app.selectedCollection.ID,
		TargetContainerID: &containerID,
		DistributionMode:  "target",
		Format:            "json",
		Data:              data,
	}

	h.app.logger.Info("Sending import request",
		"url", fmt.Sprintf("/accounts/%s/collections/%s/import", h.app.currentUser.ID, h.app.selectedCollection.ID),
		"targetContainerID", containerID,
		"distributionMode", "target")

	err := h.app.collectionsClient.ImportObjects(h.app.currentUser.ID, h.app.selectedCollection.ID, req)
	if err != nil {
		h.app.logger.Error("Import API call failed", "error", err)
		return fmt.Errorf("failed to import objects: %w", err)
	}

	h.app.logger.Info("Import API call succeeded")
	return nil
}

// DistributeToCollection distributes objects across containers in a collection
func (h *ImportHandler) DistributeToCollection(collectionID string, objects []types.CreateObjectRequest, distributionMode string) error {
	h.app.logger.Info("DistributeToCollection called",
		"collectionID", collectionID,
		"distributionMode", distributionMode,
		"objectCount", len(objects))

	// Convert CreateObjectRequest to map[string]interface{} for backend
	data := make([]map[string]interface{}, len(objects))
	for i, obj := range objects {
		data[i] = map[string]interface{}{
			"name":        obj.Name,
			"description": obj.Description,
			"quantity":    obj.Quantity,
			"unit":        obj.Unit,
			"tags":        obj.Tags,
		}
		// Add properties
		for key, value := range obj.Properties {
			data[i][key] = value
		}
	}

	h.app.logger.Info("Sample object data", "first", data[0])

	// Call backend bulk import API with automatic distribution using backend request type
	req := types.BulkImportCollectionRequest{
		CollectionID:     collectionID,
		DistributionMode: distributionMode,
		Format:           "json",
		Data:             data,
	}

	h.app.logger.Info("Sending import request",
		"url", fmt.Sprintf("/accounts/%s/collections/%s/import", h.app.currentUser.ID, collectionID),
		"distributionMode", distributionMode)

	err := h.app.collectionsClient.ImportObjects(h.app.currentUser.ID, collectionID, req)
	if err != nil {
		h.app.logger.Error("Import API call failed", "error", err)
		return fmt.Errorf("failed to distribute objects: %w", err)
	}

	h.app.logger.Info("Import API call succeeded")
	return nil
}
