package layouts

import (
	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	corestyles "cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"

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
	menu.Styler(styles.StyleBottomMenu) // Fixed bottom menu styling

	// Create menu items directly in menu (no inner container needed for flex layout)
	for i, item := range props.Items {
		isActive := i == props.ActiveIndex
		menuItem := createBottomMenuItem(menu, item, isActive)
		_ = menuItem
	}

	return menu
}

// createBottomMenuItem creates a single bottom menu item
func createBottomMenuItem(parent core.Widget, item BottomMenuItem, isActive bool) *core.Button {
	// Use Button instead of Frame for proper click handling
	btn := core.NewButton(parent)
	btn.SetType(core.ButtonAction) // Action button type

	btn.Styler(func(s *corestyles.Style) {
		// Apply StyleBottomMenuItem styling
		s.Direction = corestyles.Column        // Stack icon above text
		s.Align.Items = corestyles.Center      // Center horizontally
		s.Justify.Content = corestyles.Center  // Center vertically
		s.Background = nil                     // Transparent background
		s.Padding.Set(units.Dp(styles.Spacing2))
		s.Gap.Set(units.Dp(2))                 // Small gap between icon and text
		s.Text.WhiteSpace = corestyles.WhiteSpaceNowrap // Prevent text wrapping

		// Text sizing
		s.Font.Size.Set(12, units.UnitDp) // text-xs for label

		// Color based on active state
		if isActive {
			s.Color = colors.Uniform(styles.ColorPrimary)
		} else {
			s.Color = colors.Uniform(styles.ColorGrayDark)
		}
	})

	// Icon - size={6} in React = 24px
	if item.Icon != "" {
		btn.SetIcon(item.Icon)
	}

	// Label - text-xs in React = 12px
	if item.Label != "" {
		btn.SetText(item.Label)
	}

	// Click handler
	if item.OnClick != nil {
		btn.OnClick(func(e events.Event) {
			item.OnClick()
		})
	}

	return btn
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
