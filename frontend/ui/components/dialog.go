package components

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"

	"github.com/nishiki/frontend/ui/styles"
)

// DialogProps defines the properties for a dialog component
type DialogProps struct {
	Title      string
	OnClose    func()
	OnConfirm  func()
	OnCancel   func()
	ShowCancel bool
	ConfirmText string
	CancelText  string
}

// Dialog creates a dialog/modal component matching nishiki-frontend Dialog pattern
func Dialog(parent core.Widget, props DialogProps) *core.Frame {
	// Overlay background
	overlay := core.NewFrame(parent)
	overlay.Styler(styles.StyleOverlayBackground)

	// Dialog container
	dialog := core.NewFrame(overlay)
	dialog.Styler(styles.StyleDialogContainer)

	// Dialog title
	if props.Title != "" {
		title := core.NewText(dialog).SetText(props.Title)
		title.Styler(styles.StyleDialogTitle)
	}

	// Dialog body (returned for adding content)
	body := core.NewFrame(dialog)
	body.Styler(styles.StyleDrawerBody)

	// Dialog buttons
	buttonsRow := core.NewFrame(dialog)
	buttonsRow.Styler(styles.StyleDialogButtonRow)

	// Cancel button (if enabled)
	if props.ShowCancel {
		cancelText := props.CancelText
		if cancelText == "" {
			cancelText = "Cancel"
		}

		cancelBtn := CancelButton(buttonsRow, cancelText, func(e events.Event) {
			if props.OnCancel != nil {
				props.OnCancel()
			}
			if props.OnClose != nil {
				props.OnClose()
			}
		})
		_ = cancelBtn
	}

	// Confirm button
	confirmText := props.ConfirmText
	if confirmText == "" {
		confirmText = "Confirm"
	}

	confirmBtn := PrimaryButton(buttonsRow, confirmText, func(e events.Event) {
		if props.OnConfirm != nil {
			props.OnConfirm()
		}
		if props.OnClose != nil {
			props.OnClose()
		}
	})
	_ = confirmBtn

	return body // Return body so caller can add content
}

// ConfirmDialog creates a confirmation dialog (convenience function)
func ConfirmDialog(parent core.Widget, title, message string, onConfirm, onCancel func()) *core.Frame {
	overlay := core.NewFrame(parent)
	overlay.Styler(styles.StyleOverlayBackground)

	dialog := core.NewFrame(overlay)
	dialog.Styler(styles.StyleDialogContainer)

	// Title
	titleText := core.NewText(dialog).SetText(title)
	titleText.Styler(styles.StyleDialogTitle)

	// Message
	messageText := core.NewText(dialog).SetText(message)
	messageText.Styler(styles.StyleDescriptionText)

	// Buttons
	buttonsRow := core.NewFrame(dialog)
	buttonsRow.Styler(styles.StyleDialogButtonRow)

	cancelBtn := CancelButton(buttonsRow, "Cancel", func(e events.Event) {
		if onCancel != nil {
			onCancel()
		}
		overlay.Delete()
	})

	confirmBtn := PrimaryButton(buttonsRow, "Confirm", func(e events.Event) {
		if onConfirm != nil {
			onConfirm()
		}
		overlay.Delete()
	})

	_ = cancelBtn
	_ = confirmBtn

	return dialog
}

// DeleteConfirmDialog creates a delete confirmation dialog (convenience function)
func DeleteConfirmDialog(parent core.Widget, itemName string, onConfirm, onCancel func()) *core.Frame {
	overlay := core.NewFrame(parent)
	overlay.Styler(styles.StyleOverlayBackground)

	dialog := core.NewFrame(overlay)
	dialog.Styler(styles.StyleDialogContainer)

	// Title
	titleText := core.NewText(dialog).SetText("Confirm Delete")
	titleText.Styler(styles.StyleDialogTitle)

	// Message
	message := "Are you sure you want to delete \"" + itemName + "\"? This action cannot be undone."
	messageText := core.NewText(dialog).SetText(message)
	messageText.Styler(styles.StyleDescriptionText)

	// Buttons
	buttonsRow := core.NewFrame(dialog)
	buttonsRow.Styler(styles.StyleDialogButtonRow)

	cancelBtn := CancelButton(buttonsRow, "Cancel", func(e events.Event) {
		if onCancel != nil {
			onCancel()
		}
		overlay.Delete()
	})

	deleteBtn := DangerButton(buttonsRow, "Delete", func(e events.Event) {
		if onConfirm != nil {
			onConfirm()
		}
		overlay.Delete()
	})

	_ = cancelBtn
	_ = deleteBtn

	return dialog
}

// FormDialog creates a form dialog with title and form fields
func FormDialog(parent core.Widget, title string, onSubmit, onCancel func()) (*core.Frame, *core.Frame) {
	overlay := core.NewFrame(parent)
	overlay.Styler(styles.StyleOverlayBackground)

	dialog := core.NewFrame(overlay)
	dialog.Styler(styles.StyleDialogContainer)

	// Title
	titleText := core.NewText(dialog).SetText(title)
	titleText.Styler(styles.StyleDialogTitle)

	// Form body (returned for adding fields)
	formBody := core.NewFrame(dialog)
	formBody.Styler(styles.StyleDrawerBody)

	// Buttons
	buttonsRow := core.NewFrame(dialog)
	buttonsRow.Styler(styles.StyleDialogButtonRow)

	cancelBtn := CancelButton(buttonsRow, "Cancel", func(e events.Event) {
		if onCancel != nil {
			onCancel()
		}
		overlay.Delete()
	})

	submitBtn := PrimaryButton(buttonsRow, "Submit", func(e events.Event) {
		if onSubmit != nil {
			onSubmit()
		}
		overlay.Delete()
	})

	_ = cancelBtn
	_ = submitBtn

	return overlay, formBody // Return both overlay (for closing) and body (for adding fields)
}
