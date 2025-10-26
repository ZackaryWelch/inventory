package layouts

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"

	"github.com/nishiki/frontend/ui/components"
	"github.com/nishiki/frontend/ui/styles"
)

// HeaderProps defines the properties for the header component
type HeaderProps struct {
	Title      string
	ShowBack   bool
	OnBack     func()
	OnMenu     func()
	RightItems []HeaderItem
}

// HeaderItem represents an action item in the header
type HeaderItem struct {
	Icon    icons.Icon
	Text    string
	OnClick func()
}

// Header creates a header component matching nishiki-frontend Header.tsx
// Pattern: fixed top-0 z-40 w-full h-12 bg-white flex items-center justify-center
func Header(parent core.Widget, props HeaderProps) *core.Frame {
	header := core.NewFrame(parent)
	header.Styler(styles.StyleHeaderRow) // w-full h-12 bg-white flex items-center

	// Back button (if enabled) - positioned on left
	if props.ShowBack {
		backBtn := core.NewButton(header).SetIcon(icons.ArrowBack)
		backBtn.Styler(styles.StyleBackButton) // rounded-full bg-gray-light p-2
		if props.OnBack != nil {
			backBtn.OnClick(func(e events.Event) {
				props.OnBack()
			})
		}
	}

	// Title - centered in header, not in a container
	if props.Title != "" {
		title := core.NewText(header).SetText(props.Title)
		title.Styler(styles.StyleHeaderTitle) // Centered title styling
	}

	// Right side items
	if len(props.RightItems) > 0 {
		rightContainer := core.NewFrame(header)
		rightContainer.Styler(styles.StyleHeaderLeftContainer) // Reuse same style

		for _, item := range props.RightItems {
			if item.Icon != "" {
				btn := components.IconButton(rightContainer, item.Icon, func(e events.Event) {
					if item.OnClick != nil {
						item.OnClick()
					}
				})
				_ = btn
			} else if item.Text != "" {
				btn := components.Button(rightContainer, components.ButtonProps{
					Text:    item.Text,
					Variant: components.ButtonGhost,
					Size:    components.ButtonSizeMedium,
					OnClick: func(e events.Event) {
						if item.OnClick != nil {
							item.OnClick()
						}
					},
				})
				_ = btn
			}
		}
	}

	return header
}

// SimpleHeader creates a simple header with just a title and optional back button
func SimpleHeader(parent core.Widget, title string, showBack bool, onBack func()) *core.Frame {
	return Header(parent, HeaderProps{
		Title:    title,
		ShowBack: showBack,
		OnBack:   onBack,
	})
}

// HeaderWithMenu creates a header with menu button
func HeaderWithMenu(parent core.Widget, title string, onMenu func()) *core.Frame {
	return Header(parent, HeaderProps{
		Title:  title,
		OnMenu: onMenu,
		RightItems: []HeaderItem{
			{
				Icon:    icons.Menu,
				OnClick: onMenu,
			},
		},
	})
}

// ActionBar creates an action bar for buttons (matching React justify-end pattern)
func ActionBar(parent core.Widget) *core.Frame {
	bar := core.NewFrame(parent)
	bar.Styler(styles.StyleActionBar) // h-12 w-full flex items-center justify-end
	return bar
}
