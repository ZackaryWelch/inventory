package app

import (
	"fmt"
	"sort"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/nishiki/frontend/ui/theme"
	"github.com/nishiki/frontend/ui/widgets"
)

// renderImportPreviewDialog renders the import preview dialog
func (ga *GioApp) renderImportPreviewDialog(gtx layout.Context) layout.Dimensions {
	if !ga.showImportPreview || ga.importData == nil {
		return layout.Dimensions{}
	}

	// Handle execute button
	if ga.widgetState.importExecuteButton.Clicked(gtx) {
		ga.logger.Info("Executing import")
		go ga.executeImport()
		return layout.Dimensions{}
	}

	// Handle cancel/close button
	if ga.widgetState.importCancelButton.Clicked(gtx) {
		ga.dismissImport()
		return layout.Dimensions{}
	}

	// Choose dialog title based on state
	dialogTitle := "Import Preview"
	if ga.importResult != nil {
		dialogTitle = "Import Results"
	}

	// Create dialog style
	dialogStyle := widgets.DefaultDialogStyle(ga.widgetState.containerDialog, dialogTitle)
	dialogStyle.Width = unit.Dp(700)

	// Render draggable dialog
	dims, dismissed := dialogStyle.Layout(gtx, ga.theme.Theme, func(gtx layout.Context) layout.Dimensions {
		// Show result view after import completes with failures
		if ga.importResult != nil {
			return ga.renderImportResultView(gtx)
		}

		// Show loading state while import is running
		if ga.importRunning {
			return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{
						Top:    unit.Dp(theme.Spacing4),
						Bottom: unit.Dp(theme.Spacing4),
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						label := material.Body1(ga.theme.Theme, "Importing...")
						label.Alignment = text.Middle
						return label.Layout(gtx)
					})
				}),
			)
		}

		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Scrollable content area
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				listStyle := material.List(ga.theme.Theme, &ga.widgetState.importDialogList)
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

						// Errors section (if any)
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if len(ga.importData.Errors) > 0 {
								return ga.renderImportErrors(gtx)
							}
							return layout.Dimensions{}
						}),

						// Preview of items (inner scroll)
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return ga.renderImportPreview(gtx)
						}),

						// Column mapping
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{
								Top:    unit.Dp(theme.Spacing3),
								Bottom: unit.Dp(theme.Spacing2),
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return ga.renderImportColumnMapping(gtx)
							})
						}),

						// Summary
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{
								Top: unit.Dp(theme.Spacing2),
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								validItems := len(ga.importData.Data)
								errorCount := len(ga.importData.Errors)
								summary := fmt.Sprintf("✓ %d items ready to import", validItems)
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

			// Buttons (pinned at bottom)
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis:    layout.Horizontal,
						Spacing: layout.SpaceEnd,
					}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Right: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return widgets.CancelButton(ga.theme.Theme, &ga.widgetState.importCancelButton, "Cancel")(gtx)
							})
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if len(ga.importData.Data) == 0 {
								label := material.Body1(ga.theme.Theme, "No valid items to import")
								label.Color = theme.ColorTextSecondary
								return label.Layout(gtx)
							}
							return widgets.PrimaryButton(ga.theme.Theme, &ga.widgetState.importExecuteButton, "Import")(gtx)
						}),
					)
				})
			}),
		)
	})

	// Handle backdrop dismissal
	if dismissed {
		ga.dismissImport()
	}

	return dims
}

// renderImportErrors renders the errors section
func (ga *GioApp) renderImportErrors(gtx layout.Context) layout.Dimensions {
	return layout.Inset{Bottom: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		card := widgets.Card{
			BackgroundColor: theme.ColorDanger,
			CornerRadius:    unit.Dp(theme.RadiusDefault),
			Inset: layout.Inset{
				Top:    unit.Dp(theme.Spacing2),
				Bottom: unit.Dp(theme.Spacing2),
				Left:   unit.Dp(theme.Spacing3),
				Right:  unit.Dp(theme.Spacing3),
			},
		}

		return card.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := material.Body2(ga.theme.Theme, fmt.Sprintf("⚠️  %d Errors", len(ga.importData.Errors)))
					label.Font.Weight = font.Bold
					label.Color = theme.ColorWhite
					return label.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					// Show first 5 errors
					maxErrors := min(len(ga.importData.Errors), 5)

					return layout.Flex{Axis: layout.Vertical}.Layout(gtx, func() []layout.FlexChild {
						children := make([]layout.FlexChild, maxErrors)
						for i := range maxErrors {
							errMsg := ga.importData.Errors[i]
							children[i] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								label := material.Body2(ga.theme.Theme, "• "+errMsg)
								label.Color = theme.ColorWhite
								return label.Layout(gtx)
							})
						}
						return children
					}()...)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if len(ga.importData.Errors) > 5 {
						remaining := len(ga.importData.Errors) - 5
						label := material.Body2(ga.theme.Theme, fmt.Sprintf("...and %d more errors", remaining))
						label.Color = theme.ColorWhite
						label.Font.Style = font.Italic
						return label.Layout(gtx)
					}
					return layout.Dimensions{}
				}),
			)
		})
	})
}

// renderImportResultView renders the post-import result with error details.
func (ga *GioApp) renderImportResultView(gtx layout.Context) layout.Dimensions {
	r := ga.importResult
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Result summary
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				summary := fmt.Sprintf("%d of %d items imported successfully", r.Imported, r.Total)
				label := material.Body1(ga.theme.Theme, summary)
				label.Font.Weight = font.Bold
				return label.Layout(gtx)
			})
		}),

		// Errors section
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if len(ga.importData.Errors) > 0 {
				return ga.renderImportErrors(gtx)
			}
			return layout.Dimensions{}
		}),

		// Close button
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return widgets.PrimaryButton(ga.theme.Theme, &ga.widgetState.importCancelButton, "Close")(gtx)
			})
		}),
	)
}

// renderImportPreview renders a scrollable preview of items
func (ga *GioApp) renderImportPreview(gtx layout.Context) layout.Dimensions {
	if len(ga.importData.Data) == 0 {
		return layout.Dimensions{}
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.Body2(ga.theme.Theme, fmt.Sprintf("Preview (%d items):", len(ga.importData.Data)))
			label.Font.Weight = font.Bold
			return label.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			// Cap the preview list height
			maxHeight := gtx.Dp(unit.Dp(300))
			if gtx.Constraints.Max.Y > maxHeight {
				gtx.Constraints.Max.Y = maxHeight
			}

			listStyle := material.List(ga.theme.Theme, &ga.widgetState.importPreviewList)
			return listStyle.Layout(gtx, len(ga.importData.Data), func(gtx layout.Context, i int) layout.Dimensions {
				return ga.renderPreviewItem(gtx, ga.importData.Data[i], i+1)
			})
		}),
	)
}

// renderPreviewItem renders a single preview item
func (ga *GioApp) renderPreviewItem(gtx layout.Context, item map[string]any, index int) layout.Dimensions {
	return layout.Inset{
		Top:    unit.Dp(theme.Spacing1),
		Bottom: unit.Dp(theme.Spacing1),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		card := widgets.Card{
			BackgroundColor: theme.ColorSurfaceAlt,
			CornerRadius:    unit.Dp(theme.RadiusDefault),
			Inset: layout.Inset{
				Top:    unit.Dp(theme.Spacing2),
				Bottom: unit.Dp(theme.Spacing2),
				Left:   unit.Dp(theme.Spacing2),
				Right:  unit.Dp(theme.Spacing2),
			},
		}

		return card.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			// Get name from item using case-insensitive key lookup
			name := findStringField(item, "name", "title", "item")

			// Get description if available
			desc := findStringField(item, "description")

			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					text := fmt.Sprintf("%d. %s", index, name)
					label := material.Body2(ga.theme.Theme, text)
					label.Font.Weight = font.Bold
					return label.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if desc != "" {
						label := material.Body2(ga.theme.Theme, desc)
						label.Color = theme.ColorTextSecondary
						label.MaxLines = 1
						return label.Layout(gtx)
					}
					return layout.Dimensions{}
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					// Show field count
					fieldCount := len(item)
					label := material.Body2(ga.theme.Theme, fmt.Sprintf("%d fields", fieldCount))
					label.Color = theme.ColorTextSecondary
					label.Alignment = text.End
					return label.Layout(gtx)
				}),
			)
		})
	})
}

// importColumns returns sorted column names from the first data row.
func importColumns(data []map[string]any) []string {
	if len(data) == 0 {
		return nil
	}
	cols := make([]string, 0, len(data[0]))
	for k := range data[0] {
		cols = append(cols, k)
	}
	sort.Strings(cols)
	return cols
}

func (ga *GioApp) getImportNameColButton(col string) *widget.Clickable {
	if btn, ok := ga.widgetState.importNameColumnButtons[col]; ok {
		return btn
	}
	btn := new(widget.Clickable)
	ga.widgetState.importNameColumnButtons[col] = btn
	return btn
}

func (ga *GioApp) getImportLocationColButton(col string) *widget.Clickable {
	if btn, ok := ga.widgetState.importLocationColumnButtons[col]; ok {
		return btn
	}
	btn := new(widget.Clickable)
	ga.widgetState.importLocationColumnButtons[col] = btn
	return btn
}

// renderImportColumnMapping renders the column mapping section of the import dialog.
func (ga *GioApp) renderImportColumnMapping(gtx layout.Context) layout.Dimensions {
	cols := importColumns(ga.importData.Data)
	if len(cols) == 0 {
		return layout.Dimensions{}
	}

	// Build name column chips
	autoBtn := ga.getImportNameColButton("")
	if autoBtn.Clicked(gtx) {
		ga.importNameColumn = ""
	}
	nameChips := []layout.Widget{
		func(gtx layout.Context) layout.Dimensions {
			return ga.renderFilterChip(gtx, autoBtn, "(auto)", ga.importNameColumn == "")
		},
	}
	for _, col := range cols {
		btn := ga.getImportNameColButton(col)
		if btn.Clicked(gtx) {
			ga.importNameColumn = col
		}
		active := ga.importNameColumn == col
		nameChips = append(nameChips, func(gtx layout.Context) layout.Dimensions {
			return ga.renderFilterChip(gtx, btn, col, active)
		})
	}

	// Build location column chips
	noneBtn := ga.getImportLocationColButton("")
	if noneBtn.Clicked(gtx) {
		ga.importLocationColumn = nil
	}
	locationChips := []layout.Widget{
		func(gtx layout.Context) layout.Dimensions {
			return ga.renderFilterChip(gtx, noneBtn, "(none)", ga.importLocationColumn == nil)
		},
	}
	for _, col := range cols {
		btn := ga.getImportLocationColButton(col)
		if btn.Clicked(gtx) {
			c := col
			ga.importLocationColumn = &c
		}
		active := ga.importLocationColumn != nil && *ga.importLocationColumn == col
		locationChips = append(locationChips, func(gtx layout.Context) layout.Dimensions {
			return ga.renderFilterChip(gtx, btn, col, active)
		})
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Section header
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.Body2(ga.theme.Theme, "Column Mapping")
			label.Font.Weight = font.Bold
			return label.Layout(gtx)
		}),

		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return ga.renderChipSelector(gtx, "Name column:", nameChips)
			})
		}),

		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ga.renderChipSelector(gtx, "Location column:", locationChips)
		}),

		// Infer schema checkbox
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.CheckBox(ga.theme.Theme, &ga.widgetState.importInferSchemaCheck, "Infer property types from data").Layout(gtx)
		}),
	)
}
