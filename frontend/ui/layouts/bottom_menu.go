package layouts

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	corestyles "cogentcore.org/core/styles"

	"github.com/nishiki/frontend/ui/styles"
)

// BottomMenuProps defines the properties for the bottom menu
type BottomMenuProps struct {
	Items       []BottomMenuItem
	ActiveIndex int
}

// BottomMenuItem represents a navigation item in the bottom menu
type BottomMenuItem struct {
	Icon    icons.Icon
	Label   string
	OnClick func()
}

// BottomMenu creates a bottom navigation menu matching nishiki-frontend BottomMenu.tsx
// Pattern: fixed bottom-0 z-40 w-full bg-white border-t border-gray-light
func BottomMenu(parent core.Widget, props BottomMenuProps) *core.Frame {
	menu := core.NewFrame(parent)
	menu.Styler(styles.StyleNavHeader) // Using nav header style for bottom menu

	// Menu items container
	itemsContainer := core.NewFrame(menu)
	itemsContainer.Styler(styles.StyleNavContainer) // flex gap-3 mx-auto max-w-lg

	// Create menu items
	for i, item := range props.Items {
		isActive := i == props.ActiveIndex
		menuItem := createBottomMenuItem(itemsContainer, item, isActive)
		_ = menuItem
	}

	return menu
}

// createBottomMenuItem creates a single bottom menu item
func createBottomMenuItem(parent core.Widget, item BottomMenuItem, isActive bool) *core.Frame {
	container := core.NewFrame(parent)
	container.Styler(styles.StyleNavButton) // Base button style

	// Icon
	if item.Icon != "" {
		icon := core.NewIcon(container).SetIcon(item.Icon)
		if isActive {
			icon.Styler(styles.StyleIconPrimary) // Active state uses primary color
		} else {
			icon.Styler(styles.StyleIconGray) // Inactive state uses gray
		}
	}

	// Label
	if item.Label != "" {
		label := core.NewText(container).SetText(item.Label)
		if isActive {
			label.Styler(func(s *corestyles.Style) {
				// Active text styling (use primary color for active state)
				// For now, just apply default styling - can enhance later
			})
		} else {
			label.Styler(styles.StyleSmallText) // Inactive text
		}
		_ = label
	}

	// Click handler
	if item.OnClick != nil {
		container.OnClick(func(e events.Event) {
			item.OnClick()
		})
	}

	return container
}

// CreateDefaultBottomMenu creates a bottom menu with default navigation items
func CreateDefaultBottomMenu(parent core.Widget, activeView string, onNavigate func(view string)) *core.Frame {
	items := []BottomMenuItem{
		{
			Icon:  icons.Home,
			Label: "Home",
			OnClick: func() {
				onNavigate("dashboard")
			},
		},
		{
			Icon:  icons.Group,
			Label: "Groups",
			OnClick: func() {
				onNavigate("groups")
			},
		},
		{
			Icon:  icons.FolderOpen,
			Label: "Collections",
			OnClick: func() {
				onNavigate("collections")
			},
		},
		{
			Icon:  icons.Search,
			Label: "Search",
			OnClick: func() {
				onNavigate("search")
			},
		},
		{
			Icon:  icons.Person,
			Label: "Profile",
			OnClick: func() {
				onNavigate("profile")
			},
		},
	}

	// Determine active index based on current view
	activeIndex := 0
	switch activeView {
	case "dashboard":
		activeIndex = 0
	case "groups":
		activeIndex = 1
	case "collections":
		activeIndex = 2
	case "search":
		activeIndex = 3
	case "profile":
		activeIndex = 4
	}

	return BottomMenu(parent, BottomMenuProps{
		Items:       items,
		ActiveIndex: activeIndex,
	})
}
