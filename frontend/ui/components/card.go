package components

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"

	"github.com/nishiki/frontend/ui/styles"
)

// CardProps defines the properties for a card component
type CardProps struct {
	OnClick     func(e events.Event)
	StyleFunc   func(s *any) // Allow custom styling
	FlexBetween bool         // Use flex justify-between layout
}

// Card creates a card component matching nishiki-frontend Card.tsx
func Card(parent core.Widget, props CardProps) *core.Frame {
	card := core.NewFrame(parent)

	// Apply base card styling or flex-between variant
	if props.FlexBetween {
		card.Styler(styles.StyleCardFlexBetween)
	} else {
		card.Styler(styles.StyleCard)
	}

	// Make card clickable if onClick is provided
	if props.OnClick != nil {
		card.OnClick(props.OnClick)
	}

	return card
}

// CardHeader creates a card header section
func CardHeader(parent core.Widget) *core.Frame {
	header := core.NewFrame(parent)
	header.Styler(styles.StyleCollectionCardHeader)
	return header
}

// CardContent creates a card content section
func CardContent(parent core.Widget) *core.Frame {
	content := core.NewFrame(parent)
	content.Styler(styles.StyleCardContentColumn)
	return content
}

// CardContentGrow creates a card content section with grow layout
func CardContentGrow(parent core.Widget) *core.Frame {
	content := core.NewFrame(parent)
	content.Styler(styles.StyleCardContentGrow)
	return content
}

// CardTitle creates a card title text element
func CardTitle(parent core.Widget, text string) *core.Text {
	title := core.NewText(parent).SetText(text)
	title.Styler(styles.StyleCardTitle)
	return title
}

// CardDescription creates a card description text element
func CardDescription(parent core.Widget, text string) *core.Text {
	desc := core.NewText(parent).SetText(text)
	desc.Styler(styles.StyleDescriptionText)
	return desc
}

// CardInfo creates a card info container (flex column with gap)
func CardInfo(parent core.Widget) *core.Frame {
	info := core.NewFrame(parent)
	info.Styler(styles.StyleCardInfo)
	return info
}
