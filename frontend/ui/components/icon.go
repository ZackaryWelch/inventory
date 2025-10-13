package components

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/icons"

	"github.com/nishiki/frontend/ui/styles"
)

// IconCircleVariant defines the visual style variant of an icon circle
type IconCircleVariant int

const (
	IconCircleAccent IconCircleVariant = iota
	IconCircleFood
	IconCircleCategory
	IconCircleCategoryLarge
)

// IconCircleProps defines the properties for an icon circle component
type IconCircleProps struct {
	Icon    icons.Icon
	Emoji   string
	Variant IconCircleVariant
}

// IconCircle creates an icon circle component (bg with rounded border)
func IconCircle(parent core.Widget, props IconCircleProps) *core.Frame {
	circle := core.NewFrame(parent)

	// Apply variant styles
	switch props.Variant {
	case IconCircleAccent:
		circle.Styler(styles.StyleIconCircleAccent)
	case IconCircleFood:
		circle.Styler(styles.StyleFoodEmojiCircle)
	case IconCircleCategory:
		circle.Styler(styles.StyleCategoryIcon)
	case IconCircleCategoryLarge:
		circle.Styler(styles.StyleCategoryIconLarge)
	default:
		circle.Styler(styles.StyleIconCircleAccent)
	}

	// Add icon or emoji
	if props.Icon != "" {
		icon := core.NewIcon(circle).SetIcon(props.Icon)
		_ = icon
	} else if props.Emoji != "" {
		emoji := core.NewText(circle).SetText(props.Emoji)
		_ = emoji
	}

	return circle
}

// AccentIconCircle creates an accent icon circle (convenience function)
func AccentIconCircle(parent core.Widget, icon icons.Icon) *core.Frame {
	return IconCircle(parent, IconCircleProps{
		Icon:    icon,
		Variant: IconCircleAccent,
	})
}

// FoodEmojiCircle creates a food emoji circle (convenience function)
func FoodEmojiCircle(parent core.Widget, emoji string) *core.Frame {
	return IconCircle(parent, IconCircleProps{
		Emoji:   emoji,
		Variant: IconCircleFood,
	})
}

// CategoryIconCircle creates a category icon circle (convenience function)
func CategoryIconCircle(parent core.Widget, emoji string) *core.Frame {
	return IconCircle(parent, IconCircleProps{
		Emoji:   emoji,
		Variant: IconCircleCategory,
	})
}

// EmptyState creates an empty state message component
func EmptyState(parent core.Widget, message string) *core.Frame {
	container := core.NewFrame(parent)
	container.Styler(styles.StyleEmptyState)

	text := core.NewText(container).SetText(message)
	_ = text

	return container
}

// LoadingSpinner creates a loading spinner component
func LoadingSpinner(parent core.Widget) *core.Frame {
	spinner := core.NewFrame(parent)
	spinner.Styler(styles.StyleLoadingSpinner)
	return spinner
}

// LoadingSkeleton creates a loading skeleton placeholder
func LoadingSkeleton(parent core.Widget) *core.Frame {
	skeleton := core.NewFrame(parent)
	skeleton.Styler(styles.StyleLoadingSkeleton)
	return skeleton
}

// Divider creates a horizontal divider line
func Divider(parent core.Widget) *core.Frame {
	divider := core.NewFrame(parent)
	// Add divider styling if needed
	return divider
}
