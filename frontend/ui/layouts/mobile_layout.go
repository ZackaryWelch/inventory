package layouts

import (
	"cogentcore.org/core/core"

	"github.com/nishiki/frontend/ui/styles"
)

// MobileLayoutProps defines the properties for the mobile layout
type MobileLayoutProps struct {
	ShowHeader     bool
	ShowBottomMenu bool
}

// MobileLayout creates the main mobile layout matching nishiki-frontend MobileLayout.tsx
// Pattern: flex min-h-screen flex-col bg-gray-lightest
func MobileLayout(parent core.Widget, props MobileLayoutProps) (*core.Frame, *core.Frame) {
	// Main container
	container := core.NewFrame(parent)
	container.Styler(styles.StyleMobileLayoutContainer) // flex min-h-screen flex-col bg-gray-lightest

	// Content area (flex flex-col gap-2 px-4 pt-6 pb-16)
	content := core.NewFrame(container)
	content.Styler(styles.StyleMobileLayoutContent) // flex flex-col gap-2 px-4 pt-6 pb-16

	// Return container and content so caller can add header, content, and bottom menu
	return container, content
}

// CenteredLayout creates a centered layout for login/loading screens
func CenteredLayout(parent core.Widget) *core.Frame {
	container := core.NewFrame(parent)
	container.Styler(styles.StyleCenteredContainer)
	return container
}

// ContentColumn creates a main content column layout
func ContentColumn(parent core.Widget) *core.Frame {
	content := core.NewFrame(parent)
	content.Styler(styles.StyleContentColumn)
	return content
}

// PageTitle creates a centered page title (React MobileLayout pattern)
func PageTitle(parent core.Widget, title string) *core.Text {
	pageTitle := core.NewText(parent).SetText(title)
	pageTitle.Styler(styles.StylePageTitle)
	return pageTitle
}

// MainContainer creates the main app container
func MainContainer(parent core.Widget) *core.Frame {
	container := core.NewFrame(parent)
	container.Styler(styles.StyleMainContainer)
	return container
}
