package components

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"

	"github.com/nishiki/frontend/ui/styles"
)

// ButtonVariant defines the visual style variant of a button
type ButtonVariant int

const (
	ButtonPrimary ButtonVariant = iota
	ButtonDanger
	ButtonCancel
	ButtonAccent
	ButtonGhost
)

// ButtonSize defines the size of a button
type ButtonSize int

const (
	ButtonSizeSmall ButtonSize = iota
	ButtonSizeMedium
	ButtonSizeLarge
	ButtonSizeIcon
)

// ButtonProps defines the properties for a button component
type ButtonProps struct {
	Text     string
	Icon     icons.Icon
	Variant  ButtonVariant
	Size     ButtonSize
	OnClick  func(e events.Event)
	Disabled bool
}

// Button creates a button component matching nishiki-frontend Button.tsx
func Button(parent core.Widget, props ButtonProps) *core.Button {
	btn := core.NewButton(parent)

	// Set text if provided
	if props.Text != "" {
		btn.SetText(props.Text)
	}

	// Set icon if provided
	if props.Icon != "" {
		btn.SetIcon(props.Icon)
	}

	// Apply variant styles
	switch props.Variant {
	case ButtonPrimary:
		btn.Styler(styles.StyleButtonPrimary)
	case ButtonDanger:
		btn.Styler(styles.StyleButtonDanger)
	case ButtonCancel:
		btn.Styler(styles.StyleButtonCancel)
	case ButtonAccent:
		btn.Styler(styles.StyleButtonAccent)
	case ButtonGhost:
		btn.Styler(styles.StyleButtonGhost)
	default:
		btn.Styler(styles.StyleButtonPrimary)
	}

	// Apply size styles
	switch props.Size {
	case ButtonSizeSmall:
		btn.Styler(styles.StyleButtonSm)
	case ButtonSizeMedium:
		btn.Styler(styles.StyleButtonMd)
	case ButtonSizeLarge:
		btn.Styler(styles.StyleButtonLg)
	case ButtonSizeIcon:
		btn.Styler(styles.StyleButtonIcon)
	}

	// Set disabled state
	if props.Disabled {
		btn.SetEnabled(false)
	}

	// Set click handler
	if props.OnClick != nil {
		btn.OnClick(props.OnClick)
	}

	return btn
}

// PrimaryButton creates a primary button (convenience function)
func PrimaryButton(parent core.Widget, text string, onClick func(e events.Event)) *core.Button {
	return Button(parent, ButtonProps{
		Text:    text,
		Variant: ButtonPrimary,
		Size:    ButtonSizeMedium,
		OnClick: onClick,
	})
}

// DangerButton creates a danger button (convenience function)
func DangerButton(parent core.Widget, text string, onClick func(e events.Event)) *core.Button {
	return Button(parent, ButtonProps{
		Text:    text,
		Variant: ButtonDanger,
		Size:    ButtonSizeMedium,
		OnClick: onClick,
	})
}

// CancelButton creates a cancel button (convenience function)
func CancelButton(parent core.Widget, text string, onClick func(e events.Event)) *core.Button {
	return Button(parent, ButtonProps{
		Text:    text,
		Variant: ButtonCancel,
		Size:    ButtonSizeMedium,
		OnClick: onClick,
	})
}

// IconButton creates an icon-only button (convenience function)
func IconButton(parent core.Widget, icon icons.Icon, onClick func(e events.Event)) *core.Button {
	return Button(parent, ButtonProps{
		Icon:    icon,
		Variant: ButtonGhost,
		Size:    ButtonSizeIcon,
		OnClick: onClick,
	})
}
