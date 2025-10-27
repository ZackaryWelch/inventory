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

// showDialog creates and displays a generic dialog using Cogent Core's built-in dialog system
func (app *App) showDialog(config DialogConfig) {
	// Use Cogent Core's built-in dialog system which handles overlay properly
	d := core.NewBody().SetTitle(config.Title)

	// Optional message/description
	if config.Message != "" {
		core.NewText(d).SetText(config.Message).Styler(func(s *styles.Style) {
			s.Color = colors.Uniform(appstyles.ColorTextSecondary)
			s.Margin.Bottom = units.Dp(16)
		})
	}

	// Content (delegate to caller)
	if config.ContentBuilder != nil {
		config.ContentBuilder(d)
	}

	// Add dialog buttons
	d.AddBottomBar(func(bar *core.Frame) {
		// Cancel button (always shown)
		cancelBtn := core.NewButton(bar).SetText("Cancel")
		cancelBtn.Styler(appstyles.StyleButtonCancel)
		cancelBtn.OnClick(func(e events.Event) {
			if config.OnCancel != nil {
				config.OnCancel()
			}
			d.Close()
		})

		// Submit button (only show if OnSubmit is provided)
		if config.OnSubmit != nil {
			submitText := config.SubmitButtonText
			if submitText == "" {
				submitText = "Submit"
			}
			submitBtn := core.NewButton(bar).SetText(submitText)
			if config.SubmitButtonStyle != nil {
				submitBtn.Styler(config.SubmitButtonStyle)
			} else {
				submitBtn.Styler(appstyles.StyleButtonPrimary)
			}
			submitBtn.OnClick(func(e events.Event) {
				config.OnSubmit()
				d.Close()
			})
		}
	})

	// Run the dialog
	d.RunDialog(app.body)
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
	// Set a name to avoid "Expected source id" errors
	field.SetName(placeholder)
	field.Styler(func(s *styles.Style) {
		appstyles.StyleInputRounded(s)
		// Ensure input respects parent constraints
		s.Max.X.Set(100, units.UnitPw) // Don't exceed parent width
	})
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

