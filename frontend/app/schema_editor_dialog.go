package app

import (
	"strings"

	"gioui.org/layout"
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
	"currency":     "Cur$",
	"date":         "Date",
	"bool":         "Bool",
	"url":          "URL",
	"numeric":      "Num",
	"grouped_text": "Grpd",
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
			row.keyEditor.SetText(def.Key)
			row.displayNameEditor.SetText(def.DisplayName)
			row.requiredCheck.Value = def.Required
			rows = append(rows, row)
		}
	}
	ga.widgetState.schemaRows = rows
	ga.widgetState.schemaDialog.Reset()
	ga.showSchemaDialog = true
}

// renderSchemaEditorDialog renders the schema editor dialog overlay.
func (ga *GioApp) renderSchemaEditorDialog(gtx layout.Context) layout.Dimensions {
	if !ga.showSchemaDialog {
		return layout.Dimensions{}
	}

	// Handle submit button
	if ga.widgetState.schemaDialogSubmit.Clicked(gtx) {
		ga.handleSchemaUpdate()
		ga.widgetState.schemaDialog.Reset()
		return layout.Dimensions{}
	}

	// Handle cancel button
	if ga.widgetState.schemaDialogCancel.Clicked(gtx) {
		ga.showSchemaDialog = false
		ga.widgetState.schemaRows = nil
		ga.widgetState.schemaDialog.Reset()
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

	dialogStyle := widgets.DefaultDialogStyle(ga.widgetState.schemaDialog, "Edit Schema")
	dialogStyle.Width = unit.Dp(700)

	dims, dismissed := dialogStyle.Layout(gtx, ga.theme.Theme, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Schema rows list (scrollable, capped height)
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				maxH := gtx.Dp(unit.Dp(420))
				if gtx.Constraints.Max.Y > maxH {
					gtx.Constraints.Max.Y = maxH
				}
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
						return widgets.PrimaryButton(ga.theme.Theme, &ga.widgetState.schemaDialogSubmit, "Save Schema")(gtx)
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
		ga.showSchemaDialog = false
		ga.widgetState.schemaRows = nil
		ga.widgetState.schemaDialog.Reset()
	}

	return dims
}

// renderSchemaRow renders a single property definition row card.
func (ga *GioApp) renderSchemaRow(gtx layout.Context, i int) layout.Dimensions {
	row := &ga.widgetState.schemaRows[i]

	card := widgets.DefaultCard()
	return card.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Line 1: Key | Display Name | Required | Delete
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Bottom: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(0.3, func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Right: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return ga.renderFormField(gtx, "Key", &row.keyEditor, "e.g. brand")
							})
						}),
						layout.Flexed(0.4, func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Right: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return ga.renderFormField(gtx, "Display Name", &row.displayNameEditor, "e.g. Brand Name")
							})
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Right: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return material.CheckBox(ga.theme.Theme, &row.requiredCheck, "Required").Layout(gtx)
							})
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return widgets.DangerButton(ga.theme.Theme, &row.deleteButton, "✕")(gtx)
						}),
					)
				})
			}),

			// Line 2: Type selector chips
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				typeChildren := []layout.FlexChild{
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := material.Caption(ga.theme.Theme, "Type:")
						label.Color = theme.ColorTextSecondary
						return layout.Inset{Right: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return label.Layout(gtx)
						})
					}),
				}
				typeChildren = append(typeChildren, ga.renderPropertyTypeButtons(gtx, row)...)
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, typeChildren...)
			}),
		)
	})
}

// renderPropertyTypeButtons returns flex children for each property type chip.
func (ga *GioApp) renderPropertyTypeButtons(gtx layout.Context, row *SchemaRowState) []layout.FlexChild {
	if row.typeButtons == nil {
		row.typeButtons = make(map[string]*widget.Clickable)
	}
	children := make([]layout.FlexChild, len(propertyTypes))
	for j, pt := range propertyTypes {
		pt := pt
		if row.typeButtons[pt] == nil {
			row.typeButtons[pt] = &widget.Clickable{}
		}
		btn := row.typeButtons[pt]
		// Check click before layout (Gio immediate-mode pattern)
		if btn.Clicked(gtx) {
			row.selectedType = pt
		}
		isSelected := row.selectedType == pt
		label := propertyTypeLabels[pt]
		children[j] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Right: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				if isSelected {
					return widgets.AccentButton(ga.theme.Theme, btn, label)(gtx)
				}
				return widgets.CancelButton(ga.theme.Theme, btn, label)(gtx)
			})
		})
	}
	return children
}

// handleSchemaUpdate builds and saves the schema to the backend.
func (ga *GioApp) handleSchemaUpdate() {
	if ga.selectedCollection == nil {
		ga.logger.Error("No collection selected for schema update")
		return
	}

	defs := make([]types.PropertyDefinitionRequest, 0, len(ga.widgetState.schemaRows))
	for _, row := range ga.widgetState.schemaRows {
		key := strings.TrimSpace(row.keyEditor.Text())
		if key == "" {
			continue
		}
		propType := row.selectedType
		if propType == "" {
			propType = "text"
		}
		defs = append(defs, types.PropertyDefinitionRequest{
			Key:         key,
			DisplayName: strings.TrimSpace(row.displayNameEditor.Text()),
			Type:        propType,
			Required:    row.requiredCheck.Value,
		})
	}

	collectionID := ga.selectedCollection.ID
	accountID := ga.currentUser.ID

	go func() {
		err := ga.collectionsClient.UpdateSchema(accountID, collectionID, types.UpdatePropertySchemaRequest{
			PropertySchema: types.PropertySchemaRequest{
				Definitions: defs,
			},
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
		ga.selectedCollection = updated
		ga.window.Invalidate()
	}()

	ga.showSchemaDialog = false
	ga.widgetState.schemaRows = nil
	ga.window.Invalidate()
}
