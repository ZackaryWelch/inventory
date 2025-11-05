//go:build js && wasm

package app

import (
	"fmt"
	"strings"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"

	appstyles "github.com/nishiki/frontend/ui/styles"
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

// ShowImportDialog shows a multi-step import wizard dialog
func (app *App) ShowImportDialog(containerID string, collectionID string) {
	// Set initial distribution mode based on whether a container was specified
	distributionMode := "automatic"
	if containerID != "" {
		distributionMode = "target"
	}

	state := &ImportDialogState{
		Step:              1,
		SelectedContainer: containerID,
		DistributionMode:  distributionMode,
	}

	app.logger.Info("ShowImportDialog", "containerID", containerID, "collectionID", collectionID, "distributionMode", distributionMode)

	app.showImportDialogStep(state, collectionID)
}

// showImportDialogStep displays the current step of the import wizard
func (app *App) showImportDialogStep(state *ImportDialogState, collectionID string) {
	switch state.Step {
	case 1:
		app.showImportUploadStep(state, collectionID)
	case 2:
		app.showImportPreviewStep(state, collectionID)
	case 3:
		app.showImportSettingsStep(state, collectionID)
	case 4:
		app.showImportProgressStep(state, collectionID)
	}
}

// showImportUploadStep shows step 1: file upload or text paste
func (app *App) showImportUploadStep(state *ImportDialogState, collectionID string) {
	var textField *core.TextField

	app.showDialog(DialogConfig{
		Title:   "Import Data - Step 1: Upload",
		Message: "Select a CSV or JSON file, or paste the data directly",
		ContentBuilder: func(dialog core.Widget, closeDialog func()) {
			// File upload button
			uploadBtn := core.NewButton(dialog).SetText("Select File").SetIcon(icons.UploadFile)
			uploadBtn.Styler(appstyles.StyleButtonPrimary)
			uploadBtn.OnClick(func(e events.Event) {
				// Show loading indicator
				core.MessageSnackbar(app.body, "Selecting file...")

				handler := NewImportHandler(app)
				handler.SelectFile(func(content string, filename string, err error) {
					if err != nil {
						core.ErrorSnackbar(app.body, err, "File Selection Error")
						return
					}

					// Show parsing indicator
					core.MessageSnackbar(app.body, "Parsing file...")

					// Parse the file
					data, parseErr := handler.Parse(content, filename)
					if parseErr != nil {
						core.ErrorSnackbar(app.body, parseErr, "Parse Error")
						return
					}

					state.ImportData = data
					state.Filename = filename
					state.Step = 2

					// Close current dialog and show next step
					closeDialog()
					app.showImportDialogStep(state, collectionID)
				})
			})

			// Separator
			core.NewText(dialog).SetText("OR").Styler(func(s *styles.Style) {
				s.Color = colors.Uniform(appstyles.ColorTextSecondary)
				s.Margin.Set(units.Dp(16), units.Dp(0))
				s.Text.Align = appstyles.AlignCenter
			})

			// Text area for paste
			createSectionHeader(dialog, "Paste CSV/JSON data:")

			textField = createTextField(dialog, "Paste your CSV or JSON data here...")
			// Additional styling for multi-line text area
			textField.Styler(func(s *styles.Style) {
				s.Min.Y = units.Dp(200)
			})
		},
		SubmitButtonText:  "Next",
		SubmitButtonStyle: appstyles.StyleButtonPrimary,
		OnSubmit: func() {
			if textField.Text() == "" {
				core.ErrorSnackbar(app.body, fmt.Errorf("please select a file or paste data"), "No Data")
				return
			}

			// Parse the pasted text
			handler := NewImportHandler(app)
			data, err := handler.Parse(textField.Text(), "pasted.csv")
			if err != nil {
				core.ErrorSnackbar(app.body, err, "Parse Error")
				return
			}

			state.ImportData = data
			state.Filename = "pasted data"
			state.Step = 2
			app.showImportDialogStep(state, collectionID)
		},
	})
}

// showImportPreviewStep shows step 2: preview parsed data
func (app *App) showImportPreviewStep(state *ImportDialogState, collectionID string) {
	app.showDialog(DialogConfig{
		Title:   fmt.Sprintf("Import Data - Step 2: Preview (%s)", state.Filename),
		Message: fmt.Sprintf("Found %d objects. Review the first few items:", len(state.ImportData.Objects)),
		ContentBuilder: func(dialog core.Widget, closeDialog func()) {
			// Show errors if any
			if len(state.ImportData.Errors) > 0 {
				errorFrame := core.NewFrame(dialog)
				errorFrame.Styler(appstyles.StyleErrorAlert)

				core.NewText(errorFrame).SetText(fmt.Sprintf("⚠️ %d errors found:", len(state.ImportData.Errors))).Styler(func(s *styles.Style) {
					s.Color = colors.Uniform(appstyles.ColorDanger)
					s.Font.Weight = appstyles.WeightSemiBold
				})

				// Show first few errors
				maxErrors := 5
				if len(state.ImportData.Errors) < maxErrors {
					maxErrors = len(state.ImportData.Errors)
				}
				for i := 0; i < maxErrors; i++ {
					core.NewText(errorFrame).SetText(state.ImportData.Errors[i]).Styler(func(s *styles.Style) {
						s.Color = colors.Uniform(appstyles.ColorDanger)
						s.Font.Size = units.Dp(12)
					})
				}

				if len(state.ImportData.Errors) > maxErrors {
					core.NewText(errorFrame).SetText(fmt.Sprintf("... and %d more", len(state.ImportData.Errors)-maxErrors)).Styler(func(s *styles.Style) {
						s.Color = colors.Uniform(appstyles.ColorDanger)
						s.Font.Size = units.Dp(12)
					})
				}
			}

			// Preview list (first 5 items)
			previewList := core.NewFrame(dialog)
			previewList.Styler(appstyles.StylePreviewList)

			maxPreview := 5
			if len(state.ImportData.Objects) < maxPreview {
				maxPreview = len(state.ImportData.Objects)
			}

			for i := 0; i < maxPreview; i++ {
				obj := state.ImportData.Objects[i]
				itemCard := core.NewFrame(previewList)
				itemCard.Styler(func(s *styles.Style) {
					appstyles.StyleCard(s)
					s.Padding.Set(units.Dp(12))
				})

				// Item title
				core.NewText(itemCard).SetText(obj.Name).Styler(appstyles.StylePreviewItemTitle)

				// Item details
				if obj.Description != "" {
					core.NewText(itemCard).SetText(obj.Description).Styler(appstyles.StyleSmallText)
				}

				// Tags
				if len(obj.Tags) > 0 {
					core.NewText(itemCard).SetText("Tags: " + strings.Join(obj.Tags, ", ")).Styler(appstyles.StylePreviewItemTags)
				}
			}

			if len(state.ImportData.Objects) > maxPreview {
				core.NewText(dialog).SetText(fmt.Sprintf("... and %d more items", len(state.ImportData.Objects)-maxPreview)).Styler(func(s *styles.Style) {
					s.Color = colors.Uniform(appstyles.ColorTextSecondary)
					s.Margin.Top = units.Dp(8)
				})
			}
		},
		SubmitButtonText:  "Next",
		SubmitButtonStyle: appstyles.StyleButtonPrimary,
		OnSubmit: func() {
			state.Step = 3
			app.showImportDialogStep(state, collectionID)
		},
		OnCancel: func() {
			state.Step = 1
			app.showImportDialogStep(state, collectionID)
		},
	})
}

// showImportSettingsStep shows step 3: distribution settings
func (app *App) showImportSettingsStep(state *ImportDialogState, collectionID string) {
	app.logger.Info("showImportSettingsStep", "selectedContainer", state.SelectedContainer, "distributionMode", state.DistributionMode)

	app.showDialog(DialogConfig{
		Title:   "Import Data - Step 3: Settings",
		Message: "Choose how to distribute the imported items",
		ContentBuilder: func(dialog core.Widget, closeDialog func()) {
			// Distribution mode selector
			createSectionHeader(dialog, "Distribution Mode:")

			// Radio buttons for distribution mode
			radioFrame := core.NewFrame(dialog)
			radioFrame.Styler(func(s *styles.Style) {
				s.Direction = styles.Column
				s.Gap.Set(units.Dp(8))
				s.Margin.Bottom = units.Dp(16)
			})

			// Automatic distribution
			autoBtn := core.NewButton(radioFrame).SetText("Automatic Distribution")
			autoBtn.SetIcon(icons.Stars)
			autoBtn.Styler(func(s *styles.Style) {
				if state.DistributionMode == "automatic" {
					appstyles.StyleButtonPrimary(s)
				} else {
					appstyles.StyleButtonCancel(s)
				}
			})
			autoBtn.OnClick(func(e events.Event) {
				state.DistributionMode = "automatic"
				radioFrame.Update()
			})

			core.NewText(radioFrame).SetText("Automatically distribute items across containers based on capacity and type").Styler(func(s *styles.Style) {
				s.Color = colors.Uniform(appstyles.ColorTextSecondary)
				s.Font.Size = units.Dp(12)
				s.Margin.Left = units.Dp(32)
			})

			// Target container
			if state.SelectedContainer != "" {
				targetBtn := core.NewButton(radioFrame).SetText("Import to Selected Container")
				targetBtn.SetIcon(icons.Inventory)
				targetBtn.Styler(func(s *styles.Style) {
					if state.DistributionMode == "target" {
						appstyles.StyleButtonPrimary(s)
					} else {
						appstyles.StyleButtonCancel(s)
					}
				})
				targetBtn.OnClick(func(e events.Event) {
					state.DistributionMode = "target"
					radioFrame.Update()
				})

				core.NewText(radioFrame).SetText("Import all items to the selected container").Styler(func(s *styles.Style) {
					s.Color = colors.Uniform(appstyles.ColorTextSecondary)
					s.Font.Size = units.Dp(12)
					s.Margin.Left = units.Dp(32)
				})
			}
		},
		SubmitButtonText:  "Import",
		SubmitButtonStyle: appstyles.StyleButtonPrimary,
		OnSubmit: func() {
			state.Step = 4
			// Show progress dialog first
			app.showImportProgressStep(state, collectionID)
		},
		OnCancel: func() {
			state.Step = 2
			app.showImportDialogStep(state, collectionID)
		},
	})
}

// showImportProgressStep shows step 4: import progress
func (app *App) showImportProgressStep(state *ImportDialogState, collectionID string) {
	app.showDialog(DialogConfig{
		Title:   "Import Data - Step 4: Progress",
		Message: "Importing your data...",
		ContentBuilder: func(dialog core.Widget, closeDialog func()) {
			core.NewText(dialog).SetText("Please wait while we import your data.").Styler(func(s *styles.Style) {
				s.Color = colors.Uniform(appstyles.ColorTextSecondary)
			})

			// Progress indicator (spinner)
			// TODO: Add actual progress bar when available in Cogent Core

			// Start the import asynchronously
			app.performImport(state, collectionID, closeDialog)
		},
		OnCancel: nil, // No cancel button during import
	})
}

// performImport executes the import based on the selected settings
func (app *App) performImport(state *ImportDialogState, collectionID string, closeDialog func()) {
	handler := NewImportHandler(app)

	// Run import asynchronously
	go func() {
		var err error
		if state.DistributionMode == "target" && state.SelectedContainer != "" {
			// Import to specific container
			err = handler.ImportToContainer(state.SelectedContainer, state.ImportData.Objects)
		} else {
			// Automatic distribution
			err = handler.DistributeToCollection(collectionID, state.ImportData.Objects, state.DistributionMode)
		}

		// Update UI on main thread
		app.mainContainer.AsyncLock()
		defer app.mainContainer.AsyncUnlock()

		// Close progress dialog
		closeDialog()

		if err != nil {
			app.logger.Error("Import failed", "error", err)
			core.ErrorSnackbar(app.body, err, "Import Error")
			return
		}

		app.logger.Info("Import completed successfully", "count", len(state.ImportData.Objects))
		core.MessageSnackbar(app.body, fmt.Sprintf("Successfully imported %d items", len(state.ImportData.Objects)))

		// Refresh the current view by fetching fresh data and re-showing
		if app.selectedCollection != nil {
			// Fetch fresh collection data from backend
			freshCollection, err := app.collectionsClient.Get(app.currentUser.ID, app.selectedCollection.ID)
			if err != nil {
				app.logger.Error("Failed to refresh collection after import", "error", err)
				// Fall back to showing with stale data
				app.showCollectionDetailView(*app.selectedCollection)
			} else {
				app.logger.Info("Refreshed collection data", "containers", len(freshCollection.Containers))
				app.showCollectionDetailView(*freshCollection)
			}
		}
	}()
}
