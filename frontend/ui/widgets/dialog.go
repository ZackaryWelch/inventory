package widgets

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/nishiki/frontend/ui/theme"
)

// Dialog represents a draggable modal dialog
type Dialog struct {
	// Position of the dialog (top-left corner)
	Position f32.Point
	// Whether the dialog has been positioned (false = will be centered)
	positioned bool
	// Drag gesture for moving the dialog
	drag gesture.Drag
	// Click for backdrop dismissal
	backdropClick widget.Clickable
	// Offset during drag
	dragOffset f32.Point
}

// DialogStyle configures the appearance of a dialog
type DialogStyle struct {
	Dialog          *Dialog
	Title           string
	Width           unit.Dp
	BackgroundColor color.NRGBA
	TitleBarColor   color.NRGBA
	CornerRadius    unit.Dp
	CloseOnBackdrop bool
}

// NewDialog creates a new dialog instance
func NewDialog() *Dialog {
	return &Dialog{
		positioned: false,
	}
}

// DefaultDialogStyle creates a dialog style with default values
func DefaultDialogStyle(dialog *Dialog, title string) DialogStyle {
	return DialogStyle{
		Dialog:          dialog,
		Title:           title,
		Width:           unit.Dp(500),
		BackgroundColor: theme.ColorWhite,
		TitleBarColor:   theme.ColorPrimary,
		CornerRadius:    unit.Dp(theme.RadiusLG),
		CloseOnBackdrop: true,
	}
}

// Layout renders the dialog with the given content
func (ds DialogStyle) Layout(gtx layout.Context, th *material.Theme, content layout.Widget) (layout.Dimensions, bool) {
	dismissed := false

	// Center dialog on first render
	if !ds.Dialog.positioned {
		// Will be positioned after we know the dialog size
		ds.Dialog.positioned = true
	}

	// Handle backdrop click for dismissal
	if ds.CloseOnBackdrop && ds.Dialog.backdropClick.Clicked(gtx) {
		dismissed = true
	}

	// Handle drag events for moving the dialog
	for {
		ev, ok := ds.Dialog.drag.Update(gtx.Metric, gtx.Source, gesture.Axis(gesture.Both))
		if !ok {
			break
		}

		if ev.Kind == pointer.Press {
			ds.Dialog.dragOffset = ds.Dialog.Position.Sub(ev.Position)
		} else if ev.Kind == pointer.Drag {
			ds.Dialog.Position = ev.Position.Add(ds.Dialog.dragOffset)
		}
	}

	return layout.Stack{}.Layout(gtx,
		// Semi-transparent backdrop
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			// Draw backdrop
			defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
			paint.ColorOp{Color: theme.ColorOverlay}.Add(gtx.Ops)
			paint.PaintOp{}.Add(gtx.Ops)

			// Handle backdrop clicks
			if ds.CloseOnBackdrop {
				ds.Dialog.backdropClick.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Dimensions{Size: gtx.Constraints.Max}
				})
			}

			return layout.Dimensions{Size: gtx.Constraints.Max}
		}),

		// Dialog content
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			// First render: calculate size and center
			if ds.Dialog.Position.X == 0 && ds.Dialog.Position.Y == 0 {
				// Render to measure size
				macro := op.Record(gtx.Ops)
				dims := ds.layoutDialog(gtx, th, content)
				call := macro.Stop()

				// Center the dialog
				centerX := float32(gtx.Constraints.Max.X-dims.Size.X) / 2
				centerY := float32(gtx.Constraints.Max.Y-dims.Size.Y) / 2
				ds.Dialog.Position = f32.Point{X: centerX, Y: centerY}

				// Clamp to screen bounds
				ds.Dialog.Position = ds.clampPosition(gtx, dims.Size)

				// Draw at centered position
				offset := op.Offset(ds.Dialog.Position.Round()).Push(gtx.Ops)
				call.Add(gtx.Ops)
				offset.Pop()

				return dims
			}

			// Subsequent renders: use stored position
			// Clamp position to keep dialog on screen
			macro := op.Record(gtx.Ops)
			dims := ds.layoutDialog(gtx, th, content)
			call := macro.Stop()

			ds.Dialog.Position = ds.clampPosition(gtx, dims.Size)

			offset := op.Offset(ds.Dialog.Position.Round()).Push(gtx.Ops)
			call.Add(gtx.Ops)
			offset.Pop()

			return dims
		}),
	), dismissed
}

// layoutDialog renders the dialog structure (title bar + content)
func (ds DialogStyle) layoutDialog(gtx layout.Context, th *material.Theme, content layout.Widget) layout.Dimensions {
	// Constrain dialog width
	gtx.Constraints.Max.X = gtx.Dp(ds.Width)
	gtx.Constraints.Min.X = gtx.Dp(ds.Width)

	// Record dialog content for background
	macro := op.Record(gtx.Ops)
	dims := layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Title bar (draggable)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ds.layoutTitleBar(gtx, th)
		}),
		// Content area
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top:    unit.Dp(theme.Spacing4),
				Bottom: unit.Dp(theme.Spacing4),
				Left:   unit.Dp(theme.Spacing4),
				Right:  unit.Dp(theme.Spacing4),
			}.Layout(gtx, content)
		}),
	)
	call := macro.Stop()

	// Draw rounded background
	rr := gtx.Dp(ds.CornerRadius)
	defer clip.UniformRRect(image.Rectangle{Max: dims.Size}, rr).Push(gtx.Ops).Pop()
	paint.ColorOp{Color: ds.BackgroundColor}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)

	// Draw content
	call.Add(gtx.Ops)

	return dims
}

// layoutTitleBar renders the draggable title bar
func (ds DialogStyle) layoutTitleBar(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Inset{
		Top:    unit.Dp(theme.Spacing3),
		Bottom: unit.Dp(theme.Spacing3),
		Left:   unit.Dp(theme.Spacing4),
		Right:  unit.Dp(theme.Spacing4),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Title bar background with rounded top corners
		macro := op.Record(gtx.Ops)
		titleLabel := material.H6(th, ds.Title)
		titleLabel.Color = theme.ColorWhite
		dims := titleLabel.Layout(gtx)
		call := macro.Stop()

		// Draw title bar background
		rr := gtx.Dp(ds.CornerRadius)
		// Only round top corners
		rect := image.Rectangle{Max: image.Point{X: dims.Size.X, Y: dims.Size.Y}}
		defer clip.UniformRRect(rect, rr).Push(gtx.Ops).Pop()
		paint.ColorOp{Color: ds.TitleBarColor}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)

		// Draw title text
		call.Add(gtx.Ops)

		// Add pointer input area for dragging the dialog
		defer clip.Rect{Max: dims.Size}.Push(gtx.Ops).Pop()
		ds.Dialog.drag.Add(gtx.Ops)

		return dims
	})
}

// clampPosition ensures the dialog stays within screen bounds
func (ds DialogStyle) clampPosition(gtx layout.Context, size image.Point) f32.Point {
	pos := ds.Dialog.Position

	// Clamp X
	if pos.X < 0 {
		pos.X = 0
	}
	maxX := float32(gtx.Constraints.Max.X - size.X)
	if pos.X > maxX {
		pos.X = maxX
	}

	// Clamp Y
	if pos.Y < 0 {
		pos.Y = 0
	}
	maxY := float32(gtx.Constraints.Max.Y - size.Y)
	if pos.Y > maxY {
		pos.Y = maxY
	}

	return pos
}

// Reset resets the dialog position (will re-center on next layout)
func (d *Dialog) Reset() {
	d.Position = f32.Point{}
	d.positioned = false
}
