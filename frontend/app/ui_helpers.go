//go:build js && wasm

package app

import (
	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/cursors"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
	"image/color"

	appstyles "github.com/nishiki/frontend/ui/styles"
)

// DialogConfig defines configuration for a generic dialog
type DialogConfig struct {
	Title            string
	Message          string // Optional message/description text
	MinWidth         int
	MaxWidth         int
	OnCancel         func()
	OnSubmit         func()
	SubmitButtonText string
	SubmitButtonStyle func(*styles.Style)
	ContentBuilder   func(dialog core.Widget) // Callback to build dialog content
}

// showDialog creates and displays a generic dialog with consistent styling
func (app *App) showDialog(config DialogConfig) {
	// CRITICAL FIX: Close any existing overlay before creating a new one
	// This prevents multiple dialogs from stacking on top of each other
	if app.currentOverlay != nil {
		app.hideOverlay()
	}

	overlay := app.createOverlay()

	// Add click handler to overlay background to close dialog when clicking outside
	overlay.OnClick(func(e events.Event) {
		app.hideOverlay()
	})

	dialog := core.NewFrame(overlay)
	dialog.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(appstyles.ColorWhite)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(24))
		s.Gap.Set(units.Dp(16))
		s.Direction = styles.Column
		if config.MinWidth > 0 {
			s.Min.X.Set(float32(config.MinWidth), units.UnitDp)
		} else {
			s.Min.X.Set(400, units.UnitDp)
		}
		if config.MaxWidth > 0 {
			s.Max.X.Set(float32(config.MaxWidth), units.UnitDp)
		} else {
			s.Max.X.Set(500, units.UnitDp)
		}
	})

	// Prevent clicks on dialog itself from closing the overlay
	dialog.OnClick(func(e events.Event) {
		e.SetHandled() // Stop event from propagating to overlay
	})

	// Title
	title := core.NewText(dialog).SetText(config.Title)
	title.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(20)
		s.Font.Weight = appstyles.WeightSemiBold
		s.Color = colors.Uniform(appstyles.ColorBlack) // Ensure title is visible
	})

	// Optional message/description
	if config.Message != "" {
		message := core.NewText(dialog).SetText(config.Message)
		message.Styler(func(s *styles.Style) {
			s.Color = colors.Uniform(appstyles.ColorTextSecondary)
		})
	}

	// Content (delegate to caller)
	if config.ContentBuilder != nil {
		config.ContentBuilder(dialog)
	}

	// Button row
	buttonRow := core.NewFrame(dialog)
	buttonRow.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(units.Dp(12))
		s.Justify.Content = styles.End
	})

	// Cancel button
	cancelBtn := core.NewButton(buttonRow).SetText("Cancel")
	cancelBtn.Styler(appstyles.StyleButtonCancel)
	cancelBtn.OnClick(func(e events.Event) {
		if config.OnCancel != nil {
			config.OnCancel()
		}
		app.hideOverlay()
	})

	// Submit button (only show if OnSubmit is provided)
	if config.OnSubmit != nil {
		submitText := config.SubmitButtonText
		if submitText == "" {
			submitText = "Submit"
		}
		submitBtn := core.NewButton(buttonRow).SetText(submitText)
		if config.SubmitButtonStyle != nil {
			submitBtn.Styler(config.SubmitButtonStyle)
		} else {
			submitBtn.Styler(appstyles.StyleButtonPrimary)
		}
		submitBtn.OnClick(func(e events.Event) {
			config.OnSubmit()
		})
	}

	app.showOverlay(overlay)
}

// CardConfig defines configuration for a generic card
type CardConfig struct {
	Icon        icons.Icon
	IconColor   color.RGBA
	Title       string
	Description string
	Stats       []CardStat // Optional stats display
	OnClick     func()
	Actions     []CardAction
	Content     func(card core.Widget) // Optional custom content
}

// CardStat represents a stat to display in a card
type CardStat struct {
	Label string
	Value string
	Icon  icons.Icon
}

// CardAction represents an action button in a card
type CardAction struct {
	Icon    icons.Icon
	Color   color.RGBA
	Tooltip string
	OnClick func()
}

// createCard creates a generic card with consistent styling
func (app *App) createCard(parent core.Widget, config CardConfig) *core.Frame {
	card := core.NewFrame(parent)
	card.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Background = colors.Uniform(appstyles.ColorWhite)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(16))
		s.Border.Style.Set(styles.BorderSolid)
		s.Border.Width.Set(units.Dp(1))
		s.Border.Color.Set(colors.Uniform(appstyles.ColorGrayLight))
		s.Gap.Set(units.Dp(12))
		if config.OnClick != nil {
			s.Cursor = cursors.Pointer
		}
	})

	if config.OnClick != nil {
		card.OnClick(func(e events.Event) {
			config.OnClick()
		})
	}

	// Header row (icon + title + actions)
	cardHeader := core.NewFrame(card)
	cardHeader.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Justify.Content = styles.SpaceBetween
	})

	// Left side (icon + title)
	leftSide := core.NewFrame(cardHeader)
	leftSide.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Gap.Set(units.Dp(12))
		s.Grow.Set(1, 0)
	})

	// Icon
	if config.Icon != "" {
		iconWidget := core.NewIcon(leftSide).SetIcon(config.Icon)
		iconWidget.Styler(func(s *styles.Style) {
			s.Font.Size = units.Dp(24)
			s.Color = colors.Uniform(config.IconColor)
		})
	}

	// Title
	titleText := core.NewText(leftSide).SetText(config.Title)
	titleText.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(18)
		s.Font.Weight = appstyles.WeightSemiBold
		s.Color = colors.Uniform(appstyles.ColorBlack) // Ensure title is visible
	})

	// Actions
	if len(config.Actions) > 0 {
		actionsRow := core.NewFrame(cardHeader)
		actionsRow.Styler(func(s *styles.Style) {
			s.Direction = styles.Row
			s.Gap.Set(units.Dp(8))
		})

		for _, action := range config.Actions {
			actionBtn := core.NewButton(actionsRow).SetIcon(action.Icon)
			actionBtn.SetTooltip(action.Tooltip)
			actionBtn.Styler(func(s *styles.Style) {
				s.Background = colors.Uniform(action.Color)
				s.Color = colors.Uniform(appstyles.ColorWhite)
				s.Border.Radius = styles.BorderRadiusLarge
				s.Padding.Set(units.Dp(8))
			})
			actionBtn.OnClick(func(e events.Event) {
				action.OnClick()
			})
		}
	}

	// Description
	if config.Description != "" {
		desc := core.NewText(card).SetText(config.Description)
		desc.Styler(func(s *styles.Style) {
			s.Font.Size = units.Dp(14)
			s.Color = colors.Uniform(appstyles.ColorTextSecondary)
		})
	}

	// Stats
	if len(config.Stats) > 0 {
		statsRow := core.NewFrame(card)
		statsRow.Styler(func(s *styles.Style) {
			s.Direction = styles.Row
			s.Gap.Set(units.Dp(16))
		})

		for _, stat := range config.Stats {
			statItem := core.NewFrame(statsRow)
			statItem.Styler(func(s *styles.Style) {
				s.Direction = styles.Row
				s.Align.Items = styles.Center
				s.Gap.Set(units.Dp(6))
			})

			if stat.Icon != "" {
				statIcon := core.NewIcon(statItem).SetIcon(stat.Icon)
				statIcon.Styler(func(s *styles.Style) {
					s.Font.Size = units.Dp(16)
					s.Color = colors.Uniform(appstyles.ColorGray) // Use medium grey for icons
				})
			}

			statText := core.NewText(statItem).SetText(stat.Label + ": " + stat.Value)
			statText.Styler(func(s *styles.Style) {
				s.Font.Size = units.Dp(14)
				s.Color = colors.Uniform(appstyles.ColorBlack) // Use black for better visibility
			})
		}
	}

	// Custom content
	if config.Content != nil {
		config.Content(card)
	}

	return card
}

// createTextField creates a text field with consistent styling
func createTextField(parent core.Widget, placeholder string) *core.TextField {
	field := core.NewTextField(parent)
	field.SetText("").SetPlaceholder(placeholder)
	field.Styler(appstyles.StyleInputRounded) // Apply proper input styling
	return field
}

// createFlexRow creates a flex row container with consistent styling
func createFlexRow(parent core.Widget, gap int, justify styles.Aligns) *core.Frame {
	row := core.NewFrame(parent)
	row.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(units.Dp(float32(gap)))
		s.Justify.Content = justify
	})
	return row
}

// createFlexColumn creates a flex column container with consistent styling
func createFlexColumn(parent core.Widget, gap int) *core.Frame {
	col := core.NewFrame(parent)
	col.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(units.Dp(float32(gap)))
	})
	return col
}

// BreadcrumbItem represents a single breadcrumb link
type BreadcrumbItem struct {
	Label   string
	OnClick func()
}

