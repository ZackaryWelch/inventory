package components

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"

	"github.com/nishiki/frontend/ui/styles"
)

// InputVariant defines the visual style variant of an input
type InputVariant int

const (
	InputBase InputVariant = iota
	InputRounded
	InputSearch
)

// InputProps defines the properties for an input component
type InputProps struct {
	Placeholder string
	Value       string
	Variant     InputVariant
	OnChange    func(value string)
	Disabled    bool
}

// Input creates an input component matching nishiki-frontend Input.tsx
func Input(parent core.Widget, props InputProps) *core.TextField {
	input := core.NewTextField(parent)

	// Set placeholder
	if props.Placeholder != "" {
		input.SetPlaceholder(props.Placeholder)
	}

	// Set initial value
	if props.Value != "" {
		input.SetText(props.Value)
	}

	// Apply variant styles
	switch props.Variant {
	case InputBase:
		input.Styler(styles.StyleInputBase)
	case InputRounded:
		input.Styler(styles.StyleInputRounded)
	case InputSearch:
		input.Styler(styles.StyleSearchInputWithIcon)
	default:
		input.Styler(styles.StyleInputBase)
	}

	// Set disabled state
	if props.Disabled {
		input.SetEnabled(false)
	}

	// Set change handler
	if props.OnChange != nil {
		input.OnChange(func(e events.Event) {
			props.OnChange(input.Text())
		})
	}

	return input
}

// TextInput creates a basic text input (convenience function)
func TextInput(parent core.Widget, placeholder string, onChange func(value string)) *core.TextField {
	return Input(parent, InputProps{
		Placeholder: placeholder,
		Variant:     InputRounded,
		OnChange:    onChange,
	})
}

// SearchInput creates a search input with icon (convenience function)
func SearchInput(parent core.Widget, placeholder string, onChange func(value string)) *core.TextField {
	return Input(parent, InputProps{
		Placeholder: placeholder,
		Variant:     InputSearch,
		OnChange:    onChange,
	})
}

// Label creates a form label
func Label(parent core.Widget, text string) *core.Text {
	label := core.NewText(parent).SetText(text)
	label.Styler(styles.StyleUserFieldLabel)
	return label
}

// FormField creates a complete form field with label and input
func FormField(parent core.Widget, labelText, placeholder string, onChange func(value string)) (*core.Frame, *core.TextField) {
	container := core.NewFrame(parent)
	container.Styler(styles.StyleFormContainer)

	label := Label(container, labelText)
	input := TextInput(container, placeholder, onChange)

	_ = label
	return container, input
}
