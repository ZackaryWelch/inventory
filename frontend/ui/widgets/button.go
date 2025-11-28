package widgets

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/nishiki/frontend/ui/theme"
)

// Button renders a custom styled button
type Button struct {
	Text            string
	BackgroundColor color.NRGBA
	TextColor       color.NRGBA
	CornerRadius    unit.Dp
	Inset           layout.Inset
}

// Layout renders the button
func (b Button) Layout(gtx layout.Context, th *material.Theme, btn *widget.Clickable) layout.Dimensions {
	// Default inset if not specified
	inset := b.Inset
	if inset == (layout.Inset{}) {
		inset = layout.UniformInset(unit.Dp(theme.Spacing4))
	}

	// Create button with material design ripple effect
	return material.ButtonLayout(th, btn).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Measure and render the content (text with padding)
		macro := op.Record(gtx.Ops)
		dims := inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			label := material.Label(th, unit.Sp(theme.FontSizeBase), b.Text)
			label.Color = b.TextColor
			return label.Layout(gtx)
		})
		call := macro.Stop()

		// Draw rounded rectangle background with the measured size
		rr := gtx.Dp(b.CornerRadius)
		bounds := image.Rectangle{Max: dims.Size}
		defer clip.UniformRRect(bounds, rr).Push(gtx.Ops).Pop()
		paint.ColorOp{Color: b.BackgroundColor}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)

		// Render the recorded content on top
		call.Add(gtx.Ops)

		return dims
	})
}

// PrimaryButton creates a primary styled button
func PrimaryButton(th *material.Theme, btn *widget.Clickable, text string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		b := Button{
			Text:            text,
			BackgroundColor: theme.ColorPrimary,
			TextColor:       theme.ColorWhite,
			CornerRadius:    unit.Dp(theme.RadiusDefault),
			Inset:           layout.UniformInset(unit.Dp(theme.Spacing4)),
		}
		return b.Layout(gtx, th, btn)
	}
}

// DangerButton creates a danger styled button
func DangerButton(th *material.Theme, btn *widget.Clickable, text string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		b := Button{
			Text:            text,
			BackgroundColor: theme.ColorDanger,
			TextColor:       theme.ColorWhite,
			CornerRadius:    unit.Dp(theme.RadiusDefault),
			Inset:           layout.UniformInset(unit.Dp(theme.Spacing4)),
		}
		return b.Layout(gtx, th, btn)
	}
}

// AccentButton creates an accent styled button
func AccentButton(th *material.Theme, btn *widget.Clickable, text string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		b := Button{
			Text:            text,
			BackgroundColor: theme.ColorAccent,
			TextColor:       theme.ColorBlack,
			CornerRadius:    unit.Dp(theme.RadiusDefault),
			Inset:           layout.UniformInset(unit.Dp(theme.Spacing4)),
		}
		return b.Layout(gtx, th, btn)
	}
}

// CancelButton creates a cancel styled button
func CancelButton(th *material.Theme, btn *widget.Clickable, text string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		b := Button{
			Text:            text,
			BackgroundColor: theme.ColorGrayLightest,
			TextColor:       theme.ColorBlack,
			CornerRadius:    unit.Dp(theme.RadiusDefault),
			Inset:           layout.UniformInset(unit.Dp(theme.Spacing4)),
		}
		return b.Layout(gtx, th, btn)
	}
}

// LoginButton creates a login styled button (Authentik blue)
func LoginButton(th *material.Theme, btn *widget.Clickable, text string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		b := Button{
			Text:            text,
			BackgroundColor: theme.ColorBlue600,
			TextColor:       theme.ColorWhite,
			CornerRadius:    unit.Dp(theme.RadiusMD),
			Inset:           layout.Inset{
				Top:    unit.Dp(12),
				Bottom: unit.Dp(12),
				Left:   unit.Dp(16),
				Right:  unit.Dp(16),
			},
		}
		return b.Layout(gtx, th, btn)
	}
}
