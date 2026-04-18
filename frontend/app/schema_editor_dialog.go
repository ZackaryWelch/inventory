package app

import (
	"image"
	"strings"
	"unicode"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/nishiki/frontend/pkg/types"
	"github.com/nishiki/frontend/ui/theme"
	"github.com/nishiki/frontend/ui/widgets"
)

var propertyTypes = []string{
	"text",
	"currency",
	"date",
	"bool",
	"url",
	"numeric",
	"grouped_text",
}

var propertyTypeLabels = map[string]string{
	"text":         "Text",
	"currency":     "Currency",
	"date":         "Date",
	"bool":         "Boolean",
	"url":          "URL",
	"numeric":      "Numeric",
	"grouped_text": "Grouped",
}

// openSchemaEditor initializes schema editor state and opens the dialog.
func (ga *GioApp) openSchemaEditor() {
	var rows []SchemaRowState
	if ga.selectedCollection != nil && ga.selectedCollection.PropertySchema != nil {
		for _, def := range ga.selectedCollection.PropertySchema.Definitions {
			row := SchemaRowState{
				selectedType: def.Type,
				typeButtons:  make(map[string]*widget.Clickable),
			}
			// Use DisplayName if set, otherwise fall back to Key
			name := def.DisplayName
			if name == "" {
				name = def.Key
			}
			row.nameEditor.SetText(name)
			row.requiredCheck.Value = def.Required
			rows = append(rows, row)
		}
	}
	ga.widgetState.schemaRows = rows
	ga.widgetState.schemaDialog.Reset()
	ga.showSchemaDialog = true
}

// openSchemaEditorForImport opens the schema editor as a handoff step in the
// import flow, pre-populating rows from the non-omitted import columns.
// returnTo is "preview" or "create", selecting which import dialog to restore
// if the user cancels the schema editor.
func (ga *GioApp) openSchemaEditorForImport(returnTo string) {
	if ga.importData == nil {
		return
	}
	cols := nonOmittedColumns(ga.importData.Data, ga.importOmittedColumns)
	rows := make([]SchemaRowState, 0, len(cols))
	for _, col := range cols {
		row := SchemaRowState{
			selectedType: "text",
			typeButtons:  make(map[string]*widget.Clickable),
		}
		row.nameEditor.SetText(col)
		rows = append(rows, row)
	}
	ga.widgetState.schemaRows = rows
	ga.widgetState.schemaDialog.Reset()
	ga.schemaEditorForImport = true
	ga.importSchemaReturnTo = returnTo
	ga.showSchemaDialog = true
	// Hide originating dialogs so schema editor is the only modal.
	ga.showImportPreview = false
	ga.showImportCreateDialog = false
}

// buildSchemaRequestFromRows converts current schema editor rows into a schema
// request. Rows with empty names are dropped.
func (ga *GioApp) buildSchemaRequestFromRows() *types.PropertySchemaRequest {
	defs := make([]types.PropertyDefinitionRequest, 0, len(ga.widgetState.schemaRows))
	for _, row := range ga.widgetState.schemaRows {
		name := strings.TrimSpace(row.nameEditor.Text())
		if name == "" {
			continue
		}
		propType := row.selectedType
		if propType == "" {
			propType = "text"
		}
		defs = append(defs, types.PropertyDefinitionRequest{
			Key:         toSnakeCase(name),
			DisplayName: name,
			Type:        propType,
			Required:    row.requiredCheck.Value,
		})
	}
	return &types.PropertySchemaRequest{Definitions: defs}
}

// closeSchemaEditorAndRestoreImport hides the schema editor and re-opens the
// originating import dialog (preview or create) so the user can continue.
func (ga *GioApp) closeSchemaEditorAndRestoreImport() {
	returnTo := ga.importSchemaReturnTo
	ga.showSchemaDialog = false
	ga.widgetState.schemaRows = nil
	ga.widgetState.schemaDialog.Reset()
	ga.schemaEditorForImport = false
	ga.importSchemaReturnTo = ""
	switch returnTo {
	case "preview":
		ga.showImportPreview = true
	case "create":
		ga.showImportCreateDialog = true
	}
}

// renderSchemaEditorDialog renders the schema editor dialog overlay.
func (ga *GioApp) renderSchemaEditorDialog(gtx layout.Context) layout.Dimensions {
	if !ga.showSchemaDialog {
		return layout.Dimensions{}
	}

	// Handle submit button
	if ga.widgetState.schemaDialogSubmit.Clicked(gtx) {
		if ga.schemaEditorForImport {
			ga.handleImportSchemaContinue()
		} else {
			ga.handleSchemaUpdate()
		}
		ga.widgetState.schemaDialog.Reset()
		return layout.Dimensions{}
	}

	// Handle cancel button
	if ga.widgetState.schemaDialogCancel.Clicked(gtx) {
		if ga.schemaEditorForImport {
			ga.closeSchemaEditorAndRestoreImport()
		} else {
			ga.showSchemaDialog = false
			ga.widgetState.schemaRows = nil
			ga.widgetState.schemaDialog.Reset()
		}
		return layout.Dimensions{}
	}

	// Handle add row button (before layout so slice is stable during rendering)
	if ga.widgetState.schemaAddRowButton.Clicked(gtx) {
		ga.widgetState.schemaRows = append(ga.widgetState.schemaRows, SchemaRowState{
			selectedType: "text",
			typeButtons:  make(map[string]*widget.Clickable),
		})
	}

	// Collect row deletions during list rendering
	var deleteIndices []int

	title := "Edit Schema"
	if ga.schemaEditorForImport {
		title = "Define Import Schema"
	}
	dialogStyle := widgets.DefaultDialogStyle(ga.widgetState.schemaDialog, title)
	dialogStyle.Width = unit.Dp(700)

	dims, dismissed := dialogStyle.Layout(gtx, ga.theme.Theme, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Schema rows list (scrollable, fills available dialog space)
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				list := &ga.widgetState.schemaList
				list.Axis = layout.Vertical
				return list.Layout(gtx, len(ga.widgetState.schemaRows), func(gtx layout.Context, i int) layout.Dimensions {
					row := &ga.widgetState.schemaRows[i]
					if row.deleteButton.Clicked(gtx) {
						deleteIndices = append(deleteIndices, i)
					}
					return layout.Inset{Bottom: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return ga.renderSchemaRow(gtx, i)
					})
				})
			}),

			// Empty state hint
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if len(ga.widgetState.schemaRows) > 0 {
					return layout.Dimensions{}
				}
				return layout.Inset{Top: unit.Dp(theme.Spacing4), Bottom: unit.Dp(theme.Spacing4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					label := material.Body2(ga.theme.Theme, "No properties defined. Add one below.")
					label.Color = theme.ColorTextSecondary
					return label.Layout(gtx)
				})
			}),

			// Add Property button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: unit.Dp(theme.Spacing2), Bottom: unit.Dp(theme.Spacing4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return widgets.AccentButton(ga.theme.Theme, &ga.widgetState.schemaAddRowButton, "+ Add Property")(gtx)
				})
			}),

			// Action buttons
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:    layout.Horizontal,
					Spacing: layout.SpaceEnd,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Right: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return widgets.CancelButton(ga.theme.Theme, &ga.widgetState.schemaDialogCancel, "Cancel")(gtx)
						})
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						submitText := "Save Schema"
						if ga.schemaEditorForImport {
							submitText = "Import"
						}
						return widgets.PrimaryButton(ga.theme.Theme, &ga.widgetState.schemaDialogSubmit, submitText)(gtx)
					}),
				)
			}),
		)
	})

	// Process row deletions after layout (reverse order to preserve indices)
	for i := len(deleteIndices) - 1; i >= 0; i-- {
		idx := deleteIndices[i]
		rows := ga.widgetState.schemaRows
		ga.widgetState.schemaRows = append(rows[:idx], rows[idx+1:]...)
	}
	if len(deleteIndices) > 0 {
		ga.window.Invalidate()
	}

	if dismissed {
		if ga.schemaEditorForImport {
			ga.closeSchemaEditorAndRestoreImport()
		} else {
			ga.showSchemaDialog = false
			ga.widgetState.schemaRows = nil
			ga.widgetState.schemaDialog.Reset()
		}
	}

	return dims
}

// renderSchemaRow renders a single property definition row card.
func (ga *GioApp) renderSchemaRow(gtx layout.Context, i int) layout.Dimensions {
	row := &ga.widgetState.schemaRows[i]

	card := widgets.DefaultCard()
	return card.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Line 1: Name | Required | Delete
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Bottom: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Right: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return ga.renderFormField(gtx, "Name", &row.nameEditor, "e.g. Brand Name")
							})
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Right: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return material.CheckBox(ga.theme.Theme, &row.requiredCheck, "Required").Layout(gtx)
							})
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return widgets.DangerButton(ga.theme.Theme, &row.deleteButton, "X")(gtx)
						}),
					)
				})
			}),

			// Line 2: Type selector chips
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				chipGap := gtx.Dp(unit.Dp(theme.Spacing1))
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Start}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Right: unit.Dp(theme.Spacing2), Top: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							label := material.Body2(ga.theme.Theme, "Type:")
							label.Color = theme.ColorTextSecondary
							return label.Layout(gtx)
						})
					}),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return ga.renderPropertyTypeChips(gtx, row, chipGap)
					}),
				)
			}),
		)
	})
}

// renderPropertyTypeChips renders property type chips in a flow-wrap layout.
func (ga *GioApp) renderPropertyTypeChips(gtx layout.Context, row *SchemaRowState, gap int) layout.Dimensions {
	if row.typeButtons == nil {
		row.typeButtons = make(map[string]*widget.Clickable)
	}

	// Process clicks before layout
	for _, pt := range propertyTypes {
		if row.typeButtons[pt] == nil {
			row.typeButtons[pt] = &widget.Clickable{}
		}
		if row.typeButtons[pt].Clicked(gtx) {
			row.selectedType = pt
		}
	}

	maxWidth := gtx.Constraints.Max.X
	var x, y, rowHeight int

	type positioned struct {
		call op.CallOp
		pos  image.Point
	}
	items := make([]positioned, 0, len(propertyTypes))

	chipInset := layout.Inset{
		Top:    unit.Dp(theme.Spacing1),
		Bottom: unit.Dp(theme.Spacing1),
		Left:   unit.Dp(theme.Spacing2),
		Right:  unit.Dp(theme.Spacing2),
	}

	for _, pt := range propertyTypes {
		btn := row.typeButtons[pt]
		isSelected := row.selectedType == pt
		label := propertyTypeLabels[pt]

		// Shrink-wrap: don't let buttons expand to fill width
		cgtx := gtx
		cgtx.Constraints.Min.X = 0

		macro := op.Record(gtx.Ops)
		var b widgets.Button
		if isSelected {
			b = widgets.Button{
				Text:            label,
				BackgroundColor: theme.ColorAccent,
				TextColor:       theme.ColorBlack,
				CornerRadius:    unit.Dp(theme.RadiusFull),
				Inset:           chipInset,
			}
		} else {
			b = widgets.Button{
				Text:            label,
				BackgroundColor: theme.ColorSurfaceAlt,
				TextColor:       theme.ColorTextPrimary,
				CornerRadius:    unit.Dp(theme.RadiusFull),
				Inset:           chipInset,
			}
		}
		dims := b.Layout(cgtx, ga.theme.Theme, btn)
		call := macro.Stop()

		if x > 0 && x+dims.Size.X > maxWidth {
			y += rowHeight + gap
			x = 0
			rowHeight = 0
		}

		items = append(items, positioned{
			call: call,
			pos:  image.Point{X: x, Y: y},
		})

		x += dims.Size.X + gap
		if dims.Size.Y > rowHeight {
			rowHeight = dims.Size.Y
		}
	}

	totalHeight := y + rowHeight

	for _, item := range items {
		stack := op.Offset(item.pos).Push(gtx.Ops)
		item.call.Add(gtx.Ops)
		stack.Pop()
	}

	return layout.Dimensions{Size: image.Point{X: maxWidth, Y: totalHeight}}
}

// handleSchemaUpdate builds and saves the schema to the backend.
func (ga *GioApp) handleSchemaUpdate() {
	if ga.selectedCollection == nil {
		ga.logger.Error("No collection selected for schema update")
		return
	}

	schemaReq := ga.buildSchemaRequestFromRows()

	collectionID := ga.selectedCollection.ID
	accountID := ga.currentUser.ID

	go func() {
		err := ga.collectionsClient.UpdateSchema(accountID, collectionID, types.UpdatePropertySchemaRequest{
			PropertySchema: *schemaReq,
		})
		if err != nil {
			ga.logger.Error("Failed to update schema", "error", err)
			return
		}
		ga.logger.Info("Schema updated successfully", "collection_id", collectionID)

		// Refresh the collection to pick up the new schema
		updated, err := ga.collectionsClient.Get(accountID, collectionID)
		if err != nil {
			ga.logger.Error("Failed to refresh collection after schema update", "error", err)
			return
		}
		ga.do(func() {
			ga.selectedCollection = updated
			for i, c := range ga.collections {
				if c.ID == updated.ID {
					ga.collections[i] = *updated
					break
				}
			}
			// The schema drives object rendering (types, sort keys, grouping);
			// drop the sort/group selections and render caches so the view
			// rebuilds from the new schema.
			ga.objectSortSpecs = nil
			ga.objectGroupByField = ""
			ga.invalidateObjectCaches()
		})
		// Refetch objects so any re-coercion the backend applied is reflected.
		ga.fetchContainersAndObjects()
	}()

	ga.showSchemaDialog = false
	ga.widgetState.schemaRows = nil
	ga.window.Invalidate()
}

// handleImportSchemaContinue is invoked when the user clicks the primary button
// in the schema editor dialog while it was opened from an import flow. It stores
// the user-defined schema as a pending import schema and resumes the import.
func (ga *GioApp) handleImportSchemaContinue() {
	schemaReq := ga.buildSchemaRequestFromRows()
	ga.pendingImportSchema = schemaReq

	returnTo := ga.importSchemaReturnTo
	ga.schemaEditorForImport = false
	ga.importSchemaReturnTo = ""
	ga.showSchemaDialog = false
	ga.widgetState.schemaRows = nil

	switch returnTo {
	case "preview":
		// Run the import against the existing collection using the user schema.
		go ga.executeImport()
	case "create":
		// Resume the import-create flow with the user schema.
		go ga.executeImportCreate()
	}
	ga.window.Invalidate()
}

// toSnakeCase converts a display name to a snake_case key.
// "Brand Name" -> "brand_name", "ISBN" -> "isbn", "brand" -> "brand"
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 && result.Len() > 0 {
				prev := rune(s[i-1])
				if unicode.IsLower(prev) || unicode.IsDigit(prev) {
					result.WriteRune('_')
				} else if unicode.IsUpper(prev) && i+1 < len(s) && unicode.IsLower(rune(s[i+1])) {
					result.WriteRune('_')
				}
			}
			result.WriteRune(unicode.ToLower(r))
		} else if r == ' ' || r == '-' {
			result.WriteRune('_')
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
