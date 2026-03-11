//go:build !js || !wasm

package app

import (
	"os"
	"path/filepath"
)

// SelectImportFile on desktop is a no-op; the import dialog provides a file path
// input field and calls SelectImportFileByPath directly.
func (ga *GioApp) SelectImportFile() {}

// SelectImportFileByPath reads a file from disk and processes it as import data.
func (ga *GioApp) SelectImportFileByPath(filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		ga.logger.Error("Failed to read import file", "path", filePath, "error", err)
		return
	}
	ga.handleImportFileContent(string(content), filepath.Base(filePath))
}
