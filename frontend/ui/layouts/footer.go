package layouts

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/colors"

	appstyles "github.com/nishiki/frontend/ui/styles"
)

// BottomNavProps defines the properties for the bottom navigation
type BottomNavProps struct {
	CurrentView string
	OnGroups    func()
	OnFoods     func()
	OnProfile   func()
}

// BottomNav creates a bottom navigation bar matching React's mobile nav
// Pattern: fixed bottom-0 w-full h-16 bg-white flex justify-around items-center
func BottomNav(parent core.Widget, props BottomNavProps) *core.Frame {
	nav := core.NewFrame(parent)
	nav.Styler(func(s *styles.Style) {
		s.Min.X.Set(100, units.UnitEw)              // w-full
		s.Min.Y.Set(64, units.UnitDp)               // h-16 (64px)
		s.Background = colors.Uniform(appstyles.ColorWhite)
		s.Direction = styles.Row
		s.Justify.Content = styles.SpaceAround      // justify-around
		s.Align.Items = styles.Center
		s.Border.Width.Top = units.Dp(1)
		s.Border.Color.Top = colors.Uniform(appstyles.ColorGray)
	})

	// Groups button
	createNavButton(nav, "Groups", icons.Group, props.CurrentView == "groups", props.OnGroups)

	// Foods button (using FolderOpen icon for collections/foods)
	createNavButton(nav, "Foods", icons.FolderOpen, props.CurrentView == "collections", props.OnFoods)

	// Profile button
	createNavButton(nav, "Profile", icons.Person, props.CurrentView == "profile", props.OnProfile)

	return nav
}

// createNavButton creates a bottom nav button with icon and label
func createNavButton(parent core.Widget, label string, icon icons.Icon, isActive bool, onClick func()) {
	button := core.NewButton(parent)
	button.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Align.Items = styles.Center
		s.Gap.Set(units.Dp(4))
		s.Background = nil // Transparent background
		s.Padding.Set(units.Dp(8))

		// Active state: teal color
		if isActive {
			s.Color = colors.Uniform(appstyles.ColorPrimary)
		} else {
			s.Color = colors.Uniform(appstyles.ColorGrayDark)
		}
	})

	// Icon
	buttonIcon := core.NewIcon(button).SetIcon(icon)
	buttonIcon.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(24)
	})

	// Label
	buttonText := core.NewText(button).SetText(label)
	buttonText.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(12)
	})

	if onClick != nil {
		button.OnClick(func(e events.Event) {
			onClick()
		})
	}
}
