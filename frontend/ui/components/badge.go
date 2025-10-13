package components

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/icons"

	"github.com/nishiki/frontend/ui/styles"
)

// BadgeVariant defines the visual style variant of a badge
type BadgeVariant int

const (
	BadgeLight BadgeVariant = iota
	BadgeLightest
	BadgeOutline
	BadgeTag
	BadgeTagSecondary
)

// BadgeProps defines the properties for a badge component
type BadgeProps struct {
	Text    string
	Icon    icons.Icon
	Variant BadgeVariant
}

// Badge creates a badge component matching nishiki-frontend Badge.tsx
func Badge(parent core.Widget, props BadgeProps) *core.Frame {
	badge := core.NewFrame(parent)

	// Apply variant styles
	switch props.Variant {
	case BadgeLight:
		badge.Styler(styles.StyleBadgeLight)
	case BadgeLightest:
		badge.Styler(styles.StyleBadgeLightest)
	case BadgeOutline:
		badge.Styler(styles.StyleBadgeOutline)
	case BadgeTag:
		badge.Styler(styles.StyleTagBadge)
	case BadgeTagSecondary:
		badge.Styler(styles.StyleTagBadgeSecondary)
	default:
		badge.Styler(styles.StyleBadgeLight)
	}

	// Add icon if provided
	if props.Icon != "" {
		icon := core.NewIcon(badge).SetIcon(props.Icon)
		_ = icon
	}

	// Add text
	if props.Text != "" {
		text := core.NewText(badge).SetText(props.Text)
		_ = text
	}

	return badge
}

// LightBadge creates a light badge (convenience function)
func LightBadge(parent core.Widget, text string) *core.Frame {
	return Badge(parent, BadgeProps{
		Text:    text,
		Variant: BadgeLight,
	})
}

// OutlineBadge creates an outline badge (convenience function)
func OutlineBadge(parent core.Widget, text string) *core.Frame {
	return Badge(parent, BadgeProps{
		Text:    text,
		Variant: BadgeOutline,
	})
}

// TagBadge creates a tag badge (convenience function)
func TagBadge(parent core.Widget, text string) *core.Frame {
	return Badge(parent, BadgeProps{
		Text:    text,
		Variant: BadgeTag,
	})
}

// CategoryBadge creates a category badge with icon and text
func CategoryBadge(parent core.Widget, emoji, text string) *core.Frame {
	badge := core.NewFrame(parent)
	badge.Styler(styles.StyleBadgeLight)

	// Emoji icon circle
	iconCircle := core.NewFrame(badge)
	iconCircle.Styler(styles.StyleCategoryIcon)

	emojiText := core.NewText(iconCircle).SetText(emoji)
	_ = emojiText

	// Category name
	nameText := core.NewText(badge).SetText(text)
	_ = nameText

	return badge
}
