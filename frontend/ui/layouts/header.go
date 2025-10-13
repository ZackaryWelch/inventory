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
	header.Styler(styles.StyleHeaderRow) // w-full h-12 bg-white flex items-center justify-between

	// Left side container
	leftContainer := core.NewFrame(header)
	leftContainer.Styler(styles.StyleHeaderLeftContainer) // flex items-center gap-3

	// Back button (if enabled)
	if props.ShowBack {
		backBtn := core.NewButton(leftContainer).SetIcon(icons.ArrowBack)
		backBtn.Styler(styles.StyleBackButton) // rounded-full bg-gray-light p-2
		if props.OnBack != nil {
			backBtn.OnClick(func(e events.Event) {
				props.OnBack()
			})
		}
	}

	// Title
	if props.Title != "" {
		title := core.NewText(leftContainer).SetText(props.Title)
		title.Styler(styles.StyleSectionTitle) // text-xl font-semibold
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
