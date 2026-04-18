package app

import (
	"encoding/json"
	"fmt"
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/nishiki/frontend/pkg/types"
	"github.com/nishiki/frontend/ui/theme"
	"github.com/nishiki/frontend/ui/widgets"
)

// dismissImportCreate clears all import-create dialog state.
func (ga *GioApp) dismissImportCreate() {
	ga.showImportCreateDialog = false
	ga.importData = nil
	ga.importFilename = ""
	ga.importCreateRunning = false
	ga.importCreateError = ""
	ga.importContainerCol = nil
	ga.importOmittedColumns = nil
	ga.schemaEditorForImport = false
	ga.importSchemaReturnTo = ""
}

func (ga *GioApp) getImportCreateNameColButton(col string) *widget.Clickable {
	if btn, ok := ga.widgetState.importCreateNameColButtons[col]; ok {
		return btn
	}
	btn := new(widget.Clickable)
	ga.widgetState.importCreateNameColButtons[col] = btn
	return btn
}

func (ga *GioApp) getImportCreateContainerColButton(col string) *widget.Clickable {
	if btn, ok := ga.widgetState.importCreateContainerColButtons[col]; ok {
		return btn
	}
	btn := new(widget.Clickable)
	ga.widgetState.importCreateContainerColButtons[col] = btn
	return btn
}

// renderImportCreateDialog renders the import & create collection dialog.
func (ga *GioApp) renderImportCreateDialog(gtx layout.Context) layout.Dimensions {
	if !ga.showImportCreateDialog || ga.importData == nil {
		return layout.Dimensions{}
	}

	// Handle execute button: "Create & Import" when inferring, "Next" when
	// the user will set the schema manually in the following step.
	if ga.widgetState.importCreateExecuteButton.Clicked(gtx) {
		name := strings.TrimSpace(ga.widgetState.importCreateNameEditor.Text())
		if name != "" && !ga.importCreateRunning {
			if ga.widgetState.importCreateInferSchemaCheck.Value {
				go ga.executeImportCreate()
			} else {
				ga.openSchemaEditorForImport("create")
			}
		}
		return layout.Dimensions{}
	}

	// Handle cancel button
	if ga.widgetState.importCreateCancelButton.Clicked(gtx) {
		ga.dismissImportCreate()
		return layout.Dimensions{}
	}

	dialogStyle := widgets.DefaultDialogStyle(ga.widgetState.importCreateDialog, "Import & Create Collection")
	dialogStyle.Width = unit.Dp(700)

	dims, dismissed := dialogStyle.Layout(gtx, ga.theme.Theme, func(gtx layout.Context) layout.Dimensions {
		// Loading state
		if ga.importCreateRunning {
			return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{
						Top:    unit.Dp(theme.Spacing4),
						Bottom: unit.Dp(theme.Spacing4),
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						label := material.Body1(ga.theme.Theme, "Creating collection and importing...")
						label.Alignment = text.Middle
						return label.Layout(gtx)
					})
				}),
			)
		}

		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Scrollable content
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				listStyle := material.List(ga.theme.Theme, &ga.widgetState.importCreateDialogList)
				return listStyle.Layout(gtx, 1, func(gtx layout.Context, _ int) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						// File info
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Bottom: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										label := material.Body1(ga.theme.Theme, "File: "+ga.importFilename)
										label.Font.Weight = font.Bold
										return label.Layout(gtx)
									}),
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										info := fmt.Sprintf("Format: %s | Items: %d", ga.importData.Format, len(ga.importData.Data))
										label := material.Body2(ga.theme.Theme, info)
										label.Color = theme.ColorTextSecondary
										return label.Layout(gtx)
									}),
								)
							})
						}),

						// Error display
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if ga.importCreateError != "" {
								return layout.Inset{Bottom: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									card := widgets.Card{
										BackgroundColor: theme.ColorDanger,
										CornerRadius:    unit.Dp(theme.RadiusDefault),
										Inset: layout.Inset{
											Top: unit.Dp(theme.Spacing2), Bottom: unit.Dp(theme.Spacing2),
											Left: unit.Dp(theme.Spacing3), Right: unit.Dp(theme.Spacing3),
										},
									}
									return card.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										label := material.Body2(ga.theme.Theme, ga.importCreateError)
										label.Color = theme.ColorWhite
										return label.Layout(gtx)
									})
								})
							}
							return layout.Dimensions{}
						}),

						// Parse errors
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if len(ga.importData.Errors) > 0 {
								return ga.renderImportErrors(gtx)
							}
							return layout.Dimensions{}
						}),

						// Collection name
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return ga.renderFormField(gtx, "Collection Name *", &ga.widgetState.importCreateNameEditor, "Enter collection name")
						}),

						// Object type
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return ga.renderObjectTypeSelector(gtx)
						}),

						// Group selector
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return ga.renderGroupSelector(gtx)
						}),

						// Column mapping
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{
								Top:    unit.Dp(theme.Spacing3),
								Bottom: unit.Dp(theme.Spacing2),
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return ga.renderImportCreateColumnMapping(gtx)
							})
						}),

						// Preview
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return ga.renderImportPreview(gtx, &ga.widgetState.importCreatePreviewList)
						}),

						// Summary
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Top: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								validItems := len(ga.importData.Data)
								errorCount := len(ga.importData.Errors)
								summary := fmt.Sprintf("%d items ready to import", validItems)
								if errorCount > 0 {
									summary += fmt.Sprintf(" (%d errors)", errorCount)
								}
								label := material.Body1(ga.theme.Theme, summary)
								label.Font.Weight = font.Bold
								return label.Layout(gtx)
							})
						}),
					)
				})
			}),

			// Buttons
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis:    layout.Horizontal,
						Spacing: layout.SpaceEnd,
					}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Right: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return widgets.CancelButton(ga.theme.Theme, &ga.widgetState.importCreateCancelButton, "Cancel")(gtx)
							})
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							name := strings.TrimSpace(ga.widgetState.importCreateNameEditor.Text())
							if len(ga.importData.Data) == 0 || name == "" {
								label := material.Body1(ga.theme.Theme, "Name required")
								label.Color = theme.ColorTextSecondary
								return label.Layout(gtx)
							}
							buttonText := "Create & Import"
							if !ga.widgetState.importCreateInferSchemaCheck.Value {
								buttonText = "Next"
							}
							return widgets.PrimaryButton(ga.theme.Theme, &ga.widgetState.importCreateExecuteButton, buttonText)(gtx)
						}),
					)
				})
			}),
		)
	})

	if dismissed {
		ga.dismissImportCreate()
	}

	return dims
}

// renderImportCreateColumnMapping renders column mapping for the import-create dialog.
func (ga *GioApp) renderImportCreateColumnMapping(gtx layout.Context) layout.Dimensions {
	cols := importColumns(ga.importData.Data)
	if len(cols) == 0 {
		return layout.Dimensions{}
	}

	availableCols := nonOmittedColumns(ga.importData.Data, ga.importOmittedColumns)

	// Build name column chips
	autoBtn := ga.getImportCreateNameColButton("")
	if autoBtn.Clicked(gtx) {
		ga.importNameColumn = ""
	}
	nameChips := []layout.Widget{
		func(gtx layout.Context) layout.Dimensions {
			return ga.renderFilterChip(gtx, autoBtn, "(auto)", ga.importNameColumn == "")
		},
	}
	for _, col := range availableCols {
		btn := ga.getImportCreateNameColButton(col)
		if btn.Clicked(gtx) {
			ga.importNameColumn = col
		}
		active := ga.importNameColumn == col
		nameChips = append(nameChips, func(gtx layout.Context) layout.Dimensions {
			return ga.renderFilterChip(gtx, btn, col, active)
		})
	}

	// Build container column chips
	noneContainerBtn := ga.getImportCreateContainerColButton("")
	if noneContainerBtn.Clicked(gtx) {
		ga.importContainerCol = nil
	}
	containerChips := []layout.Widget{
		func(gtx layout.Context) layout.Dimensions {
			return ga.renderFilterChip(gtx, noneContainerBtn, "(none)", ga.importContainerCol == nil)
		},
	}
	for _, col := range availableCols {
		btn := ga.getImportCreateContainerColButton(col)
		if btn.Clicked(gtx) {
			c := col
			ga.importContainerCol = &c
		}
		active := ga.importContainerCol != nil && *ga.importContainerCol == col
		containerChips = append(containerChips, func(gtx layout.Context) layout.Dimensions {
			return ga.renderFilterChip(gtx, btn, col, active)
		})
	}

	// Omit column chips (multi-select toggles; show every column).
	omitChips := make([]layout.Widget, 0, len(cols))
	for _, col := range cols {
		btn := ga.getImportOmitColButton(col)
		if btn.Clicked(gtx) {
			ga.toggleOmittedColumn(col)
		}
		active := ga.importOmittedColumns[col]
		omitChips = append(omitChips, func(gtx layout.Context) layout.Dimensions {
			return ga.renderFilterChip(gtx, btn, col, active)
		})
	}

	children := []layout.FlexChild{
		// Section header
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.Body2(ga.theme.Theme, "Column Mapping")
			label.Font.Weight = font.Bold
			return label.Layout(gtx)
		}),
		// Name column
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return ga.renderChipSelector(gtx, "Name column:", nameChips)
			})
		}),
		// Container column
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ga.renderChipSelector(gtx, "Container column:", containerChips)
		}),
		// Omit columns
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ga.renderChipSelector(gtx, "Omit columns:", omitChips)
		}),
	}

	// Location field (shown when no container column selected)
	if ga.importContainerCol == nil {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ga.renderFormField(gtx, "Location", &ga.widgetState.importCreateLocationEditor, "e.g., Kitchen, Living Room")
		}))
	}

	// Infer schema checkbox
	children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return material.CheckBox(ga.theme.Theme, &ga.widgetState.importCreateInferSchemaCheck, "Infer property types from data").Layout(gtx)
	}))

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

// executeImportCreate creates a new collection and imports data into it.
func (ga *GioApp) executeImportCreate() {
	if ga.importData == nil || ga.currentUser == nil {
		ga.logger.Error("Cannot execute import-create: missing data or user")
		return
	}

	name := strings.TrimSpace(ga.widgetState.importCreateNameEditor.Text())
	if name == "" {
		return
	}

	ga.importCreateRunning = true
	ga.importCreateError = ""
	ga.window.Invalidate()

	objectType := ga.selectedObjectType
	groupID := ga.selectedGroupID
	location := strings.TrimSpace(ga.widgetState.importCreateLocationEditor.Text())
	containerCol := ga.importContainerCol
	nameCol := ga.importNameColumn
	inferSchema := ga.widgetState.importCreateInferSchemaCheck.Value
	userID := ga.currentUser.ID

	// A user-defined schema supplied via the schema editor overrides inference.
	userSchema := ga.pendingImportSchema
	if userSchema != nil {
		inferSchema = false
	}
	ga.pendingImportSchema = nil

	// Step 1: Create collection
	createReq := types.CreateCollectionRequest{
		Name:           name,
		ObjectType:     objectType,
		GroupID:        groupID,
		PropertySchema: userSchema,
	}
	// Only set location when no container column (container col creates containers instead)
	if containerCol == nil {
		createReq.Location = location
	}

	collection, err := ga.collectionsClient.Create(userID, createReq)
	if err != nil {
		ga.do(func() {
			ga.importCreateRunning = false
			ga.importCreateError = fmt.Sprintf("Failed to create collection: %v", err)
		})
		return
	}

	ga.logger.Info("Collection created for import", "collection_id", collection.ID, "name", name)

	// Step 2: Import data
	distMode := "automatic"
	if containerCol != nil {
		distMode = "location"
	}

	filteredData := filterOmittedColumns(ga.importData.Data, ga.importOmittedColumns)
	importReq := map[string]any{
		"format":            ga.importData.Format,
		"data":              filteredData,
		"distribution_mode": distMode,
		"infer_schema":      inferSchema,
	}
	if containerCol != nil {
		importReq["location_column"] = *containerCol
	}
	if nameCol != "" {
		importReq["name_column"] = nameCol
	}

	endpoint := fmt.Sprintf("/accounts/%s/collections/%s/import", userID, collection.ID)
	resp, err := ga.apiClient.Post(endpoint, importReq)
	if err != nil {
		ga.do(func() {
			ga.importCreateRunning = false
			ga.importCreateError = fmt.Sprintf("Collection created but import failed: %v", err)
			// Add collection to list so user can navigate to it
			ga.collections = append(ga.collections, *collection)
		})
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		resp.Body.Close()
		errMsg := errResp.Error
		if errMsg == "" {
			errMsg = fmt.Sprintf("server error (status %d)", resp.StatusCode)
		}
		ga.do(func() {
			ga.importCreateRunning = false
			ga.importCreateError = "Collection created but import failed: " + errMsg
			ga.collections = append(ga.collections, *collection)
		})
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
		ga.do(func() {
			ga.importCreateRunning = false
			ga.importCreateError = fmt.Sprintf("Failed to parse import response: %v", err)
			ga.collections = append(ga.collections, *collection)
		})
		return
	}

	ga.logger.Info("Import-create completed",
		"imported", result.Imported,
		"failed", result.Failed,
		"total", result.Total,
		"containers_created", result.ContainersCreated)

	if result.Failed > 0 {
		ga.do(func() {
			ga.importCreateRunning = false
			var errSummary string
			if len(result.Errors) > 0 {
				maxShow := min(len(result.Errors), 5)
				errSummary = strings.Join(result.Errors[:maxShow], "; ")
				if len(result.Errors) > 5 {
					errSummary += fmt.Sprintf(" ...and %d more", len(result.Errors)-5)
				}
			}
			ga.importCreateError = fmt.Sprintf("Imported %d of %d items (%d failed). %s",
				result.Imported, result.Total, result.Failed, errSummary)
			ga.collections = append(ga.collections, *collection)
		})
		return
	}

	// Success — navigate to the new collection
	collectionID := collection.ID
	ga.do(func() {
		ga.collections = append(ga.collections, *collection)
		ga.selectedCollection = collection
		ga.currentView = ViewCollectionDetailGio
		ga.dismissImportCreate()
	})

	// Refetch collection to pick up the schema (inferred or user-defined).
	if inferSchema || userSchema != nil {
		updated, err := ga.collectionsClient.Get(userID, collectionID)
		if err == nil {
			ga.do(func() {
				ga.selectedCollection = updated
				for i, c := range ga.collections {
					if c.ID == updated.ID {
						ga.collections[i] = *updated
						break
					}
				}
				ga.objectSortSpecs = nil
				ga.objectGroupByField = ""
				ga.invalidateObjectCaches()
			})
		}
	}

	ga.fetchContainersAndObjects()
}
