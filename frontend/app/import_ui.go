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
	"cogentcore.org/core/styles/sides"
	"cogentcore.org/core/styles/units"

	appstyles "github.com/nishiki/frontend/ui/styles"
)

// ImportDialogState tracks the state of the import wizard
type ImportDialogState struct {
	Step            int // 1=upload, 2=preview, 3=settings, 4=progress
	ImportData      *ImportData
	Filename        string
	SelectedContainer string
	DistributionMode string // "automatic" or "manual" or "target"
	ImportErrors    []string
}

// ShowImportDialog shows a multi-step import wizard dialog
func (app *App) ShowImportDialog(containerID string, collectionID string) {
	state := &ImportDialogState{
		Step:              1,
		SelectedContainer: containerID,
		DistributionMode:  "automatic",
	}

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
		ContentBuilder: func(dialog core.Widget) {
			// File upload button
			uploadBtn := core.NewButton(dialog).SetText("Select File").SetIcon(icons.UploadFile)
			uploadBtn.Styler(appstyles.StyleButtonPrimary)
			uploadBtn.OnClick(func(e events.Event) {
				handler := NewImportHandler(app)
				handler.SelectFile(func(content string, filename string, err error) {
					if err != nil {
						core.ErrorSnackbar(app.body, err, "File Selection Error")
						return
					}

					// Parse the file
					data, parseErr := handler.Parse(content, filename)
					if parseErr != nil {
						core.ErrorSnackbar(app.body, parseErr, "Parse Error")
						return
					}

					state.ImportData = data
					state.Filename = filename
					state.Step = 2
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
			core.NewText(dialog).SetText("Paste CSV/JSON data:").Styler(func(s *styles.Style) {
				s.Font.Weight = appstyles.WeightSemiBold
				s.Margin.Bottom = units.Dp(8)
			})

			textField = core.NewTextField(dialog)
			textField.SetPlaceholder("Paste your CSV or JSON data here...")
			textField.Styler(func(s *styles.Style) {
				appstyles.StyleInputRounded(s)
				s.Min.Y = units.Dp(200)
			})
		},
		SubmitButtonText: "Next",
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
		ContentBuilder: func(dialog core.Widget) {
			// Show errors if any
			if len(state.ImportData.Errors) > 0 {
				errorFrame := core.NewFrame(dialog)
				errorFrame.Styler(func(s *styles.Style) {
					s.Background = colors.Uniform(appstyles.ColorDanger)
					s.Padding.Set(units.Dp(12))
					s.Border.Radius = sides.NewValues(units.Dp(appstyles.RadiusDefault))
					s.Margin.Bottom = units.Dp(16)
				})

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
			previewList.Styler(func(s *styles.Style) {
				s.Direction = styles.Column
				s.Gap.Set(units.Dp(8))
				s.Max.Y = units.Dp(300)
				s.Overflow.Y = styles.OverflowAuto
			})

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
				core.NewText(itemCard).SetText(obj.Name).Styler(func(s *styles.Style) {
					s.Font.Weight = appstyles.WeightSemiBold
					s.Font.Size = units.Dp(14)
				})

				// Item details
				if obj.Description != "" {
					core.NewText(itemCard).SetText(obj.Description).Styler(func(s *styles.Style) {
						s.Color = colors.Uniform(appstyles.ColorTextSecondary)
						s.Font.Size = units.Dp(12)
					})
				}

				// Tags
				if len(obj.Tags) > 0 {
					core.NewText(itemCard).SetText("Tags: " + strings.Join(obj.Tags, ", ")).Styler(func(s *styles.Style) {
						s.Color = colors.Uniform(appstyles.ColorPrimary)
						s.Font.Size = units.Dp(11)
					})
				}
			}

			if len(state.ImportData.Objects) > maxPreview {
				core.NewText(dialog).SetText(fmt.Sprintf("... and %d more items", len(state.ImportData.Objects)-maxPreview)).Styler(func(s *styles.Style) {
					s.Color = colors.Uniform(appstyles.ColorTextSecondary)
					s.Margin.Top = units.Dp(8)
				})
			}
		},
		SubmitButtonText: "Next",
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
	app.showDialog(DialogConfig{
		Title:   "Import Data - Step 3: Settings",
		Message: "Choose how to distribute the imported items",
		ContentBuilder: func(dialog core.Widget) {
			// Distribution mode selector
			core.NewText(dialog).SetText("Distribution Mode:").Styler(func(s *styles.Style) {
				s.Font.Weight = appstyles.WeightSemiBold
				s.Margin.Bottom = units.Dp(8)
			})

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
				})

				core.NewText(radioFrame).SetText("Import all items to the selected container").Styler(func(s *styles.Style) {
					s.Color = colors.Uniform(appstyles.ColorTextSecondary)
					s.Font.Size = units.Dp(12)
					s.Margin.Left = units.Dp(32)
				})
			}
		},
		SubmitButtonText: "Import",
		SubmitButtonStyle: appstyles.StyleButtonPrimary,
		OnSubmit: func() {
			state.Step = 4
			app.performImport(state, collectionID)
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
		ContentBuilder: func(dialog core.Widget) {
			core.NewText(dialog).SetText("Please wait while we import your data.").Styler(func(s *styles.Style) {
				s.Color = colors.Uniform(appstyles.ColorTextSecondary)
			})

			// Progress indicator (spinner)
			// TODO: Add actual progress bar when available in Cogent Core
		},
		OnCancel: nil, // No cancel button during import
	})
}

// performImport executes the import based on the selected settings
func (app *App) performImport(state *ImportDialogState, collectionID string) {
	handler := NewImportHandler(app)

	var err error
	if state.DistributionMode == "target" && state.SelectedContainer != "" {
		// Import to specific container
		err = handler.ImportToContainer(state.SelectedContainer, state.ImportData.Objects)
	} else {
		// Automatic distribution
		err = handler.DistributeToCollection(collectionID, state.ImportData.Objects, state.DistributionMode)
	}

	if err != nil {
		core.ErrorSnackbar(app.body, err, "Import Error")
		return
	}

	core.MessageSnackbar(app.body, fmt.Sprintf("Successfully imported %d items", len(state.ImportData.Objects)))

	// Refresh the current view by re-showing the collection detail
	if app.selectedCollection != nil {
		app.showCollectionDetailView(*app.selectedCollection)
	}
}
