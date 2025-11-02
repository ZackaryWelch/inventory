//go:build !js || !wasm

package app

import (
	"fmt"

	"cogentcore.org/core/core"
)

// ImportDialogState tracks the state of the import wizard
type ImportDialogState struct {
	Step              int // 1=upload, 2=preview, 3=settings, 4=progress
	ImportData        *ImportData
	Filename          string
	SelectedContainer string
	DistributionMode  string // "automatic" or "manual" or "target"
	ImportErrors      []string
}

// ShowImportDialog shows a simplified import dialog for desktop
func (app *App) ShowImportDialog(containerID string, collectionID string) {
	core.ErrorSnackbar(app, fmt.Errorf("import not yet implemented for desktop"), "Not Implemented")
}

// performImport executes the import based on the selected settings
func (app *App) performImport(state *ImportDialogState, collectionID string) {
	core.ErrorSnackbar(app, fmt.Errorf("import not yet implemented for desktop"), "Not Implemented")
}
