//go:build !js || !wasm

package app

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"cogentcore.org/core/core"
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
	// For desktop, we'll use Cogent Core's file dialog
	// This is a simplified version - you may want to enhance with proper file dialog
	core.ErrorSnackbar(h.app, fmt.Errorf("file selection not implemented for desktop"), "Not Implemented")
	callback("", "", fmt.Errorf("file selection not implemented for desktop"))
}

// ReadFile reads a file from disk (desktop only)
func (h *ImportHandler) ReadFile(filepath string) (string, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(content), nil
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

	// Get name (required)
	if idx, ok := headerMap["name"]; ok && idx < len(record) {
		obj.Name = strings.TrimSpace(record[idx])
	} else if idx, ok := headerMap["title"]; ok && idx < len(record) {
		obj.Name = strings.TrimSpace(record[idx])
	} else {
		return nil, fmt.Errorf("missing required field 'name' or 'title'")
	}

	if obj.Name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	// Get description
	if idx, ok := headerMap["description"]; ok && idx < len(record) {
		obj.Description = strings.TrimSpace(record[idx])
	}

	// Get category
	if idx, ok := headerMap["category"]; ok && idx < len(record) {
		obj.Category = strings.TrimSpace(record[idx])
	}

	// Get quantity
	if idx, ok := headerMap["quantity"]; ok && idx < len(record) {
		if qtyStr := strings.TrimSpace(record[idx]); qtyStr != "" {
			qty, err := strconv.ParseFloat(qtyStr, 64)
			if err == nil {
				obj.Quantity = qty
			}
		}
	}

	// Get unit
	if idx, ok := headerMap["unit"]; ok && idx < len(record) {
		obj.Unit = strings.TrimSpace(record[idx])
	}

	// Get tags
	if idx, ok := headerMap["tags"]; ok && idx < len(record) {
		tagsStr := strings.TrimSpace(record[idx])
		if tagsStr != "" {
			tags := strings.Split(tagsStr, ",")
			for _, tag := range tags {
				if trimmed := strings.TrimSpace(tag); trimmed != "" {
					obj.Tags = append(obj.Tags, trimmed)
				}
			}
		}
	}

	// Parse all other columns as properties
	for header, idx := range headerMap {
		if idx >= len(record) {
			continue
		}

		// Skip already handled fields
		if header == "name" || header == "title" || header == "description" ||
			header == "category" || header == "quantity" || header == "unit" || header == "tags" {
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
	// Set container ID for all objects
	for i := range objects {
		objects[i].ContainerID = containerID
	}

	// TODO: Call backend API to import objects
	// For now, this is a placeholder
	core.ErrorSnackbar(h.app, fmt.Errorf("import not yet implemented"), "Import Error")
	return fmt.Errorf("import not yet implemented")
}

// DistributeToCollection distributes objects across containers in a collection
func (h *ImportHandler) DistributeToCollection(collectionID string, objects []types.CreateObjectRequest, distributionMode string) error {
	// TODO: Implement distribution logic
	// For automatic mode, this will use the backend distribution service
	// For manual mode, user selects containers
	core.ErrorSnackbar(h.app, fmt.Errorf("distribution not yet implemented"), "Import Error")
	return fmt.Errorf("distribution not yet implemented")
}
