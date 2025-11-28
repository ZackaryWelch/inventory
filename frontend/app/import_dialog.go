//go:build js && wasm

package app

import (
	"fmt"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
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

	// Handle cancel button
	if ga.widgetState.importCancelButton.Clicked(gtx) {
		ga.showImportPreview = false
		ga.importData = nil
		ga.importFilename = ""
		return layout.Dimensions{}
	}

	// Create dialog style
	dialogStyle := widgets.DefaultDialogStyle(ga.widgetState.containerDialog, "Import Preview")
	dialogStyle.Width = unit.Dp(700)

	// Render draggable dialog
	dims, dismissed := dialogStyle.Layout(gtx, ga.theme.Theme, func(gtx layout.Context) layout.Dimensions {
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

			// Preview of first few items
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ga.renderImportPreview(gtx)
			}),

			// Summary
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Top:    unit.Dp(theme.Spacing3),
					Bottom: unit.Dp(theme.Spacing3),
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

			// Buttons
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
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
							// Disable import button if no valid data
							label := material.Body1(ga.theme.Theme, "No valid items to import")
							label.Color = theme.ColorTextSecondary
							return label.Layout(gtx)
						}
						return widgets.PrimaryButton(ga.theme.Theme, &ga.widgetState.importExecuteButton, "Import")(gtx)
					}),
				)
			}),
		)
	})

	// Handle backdrop dismissal
	if dismissed {
		ga.showImportPreview = false
		ga.importData = nil
		ga.importFilename = ""
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
					maxErrors := 5
					if len(ga.importData.Errors) < maxErrors {
						maxErrors = len(ga.importData.Errors)
					}

					return layout.Flex{Axis: layout.Vertical}.Layout(gtx, func() []layout.FlexChild {
						children := make([]layout.FlexChild, maxErrors)
						for i := 0; i < maxErrors; i++ {
							errMsg := ga.importData.Errors[i]
							children[i] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								label := material.Caption(ga.theme.Theme, "• "+errMsg)
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
						label := material.Caption(ga.theme.Theme, fmt.Sprintf("...and %d more errors", remaining))
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

// renderImportPreview renders a preview of the first few items
func (ga *GioApp) renderImportPreview(gtx layout.Context) layout.Dimensions {
	if len(ga.importData.Data) == 0 {
		return layout.Dimensions{}
	}

	return layout.Inset{Bottom: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Body2(ga.theme.Theme, "Preview (first 5 items):")
				label.Font.Weight = font.Bold
				return label.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				// Show first 5 items
				maxItems := 5
				if len(ga.importData.Data) < maxItems {
					maxItems = len(ga.importData.Data)
				}

				return layout.Flex{Axis: layout.Vertical}.Layout(gtx, func() []layout.FlexChild {
					children := make([]layout.FlexChild, maxItems)
					for i := 0; i < maxItems; i++ {
						item := ga.importData.Data[i]
						children[i] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return ga.renderPreviewItem(gtx, item, i+1)
						})
					}
					return children
				}()...)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if len(ga.importData.Data) > 5 {
					remaining := len(ga.importData.Data) - 5
					label := material.Caption(ga.theme.Theme, fmt.Sprintf("...and %d more items", remaining))
					label.Color = theme.ColorTextSecondary
					label.Font.Style = font.Italic
					return label.Layout(gtx)
				}
				return layout.Dimensions{}
			}),
		)
	})
}

// renderPreviewItem renders a single preview item
func (ga *GioApp) renderPreviewItem(gtx layout.Context, item map[string]interface{}, index int) layout.Dimensions {
	return layout.Inset{
		Top:    unit.Dp(theme.Spacing1),
		Bottom: unit.Dp(theme.Spacing1),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		card := widgets.Card{
			BackgroundColor: theme.ColorGrayLightest,
			CornerRadius:    unit.Dp(theme.RadiusDefault),
			Inset: layout.Inset{
				Top:    unit.Dp(theme.Spacing2),
				Bottom: unit.Dp(theme.Spacing2),
				Left:   unit.Dp(theme.Spacing2),
				Right:  unit.Dp(theme.Spacing2),
			},
		}

		return card.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			// Get name from item (try multiple field names)
			name := ""
			if n, ok := item["name"].(string); ok {
				name = n
			} else if n, ok := item["title"].(string); ok {
				name = n
			} else if n, ok := item["item"].(string); ok {
				name = n
			}

			// Get description if available
			desc := ""
			if d, ok := item["description"].(string); ok {
				desc = d
			}

			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					text := fmt.Sprintf("%d. %s", index, name)
					label := material.Body2(ga.theme.Theme, text)
					label.Font.Weight = font.Bold
					return label.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if desc != "" {
						label := material.Caption(ga.theme.Theme, desc)
						label.Color = theme.ColorTextSecondary
						label.MaxLines = 1
						return label.Layout(gtx)
					}
					return layout.Dimensions{}
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					// Show field count
					fieldCount := len(item)
					label := material.Caption(ga.theme.Theme, fmt.Sprintf("%d fields", fieldCount))
					label.Color = theme.ColorTextSecondary
					label.Alignment = text.End
					return label.Layout(gtx)
				}),
			)
		})
	})
}
