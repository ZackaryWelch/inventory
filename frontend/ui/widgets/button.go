package widgets

import (
	"image/color"

	"gioui.org/layout"
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

	// Use ButtonLayoutStyle's own Background and CornerRadius so the
	// background color fills exactly the rounded rect (no color leak).
	bls := material.ButtonLayout(th, btn)
	bls.Background = b.BackgroundColor
	bls.CornerRadius = b.CornerRadius

	return bls.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			label := material.Label(th, unit.Sp(theme.FontSizeBase), b.Text)
			label.Color = b.TextColor
			return label.Layout(gtx)
		})
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
			Inset: layout.Inset{
				Top:    unit.Dp(12),
				Bottom: unit.Dp(12),
				Left:   unit.Dp(16),
				Right:  unit.Dp(16),
			},
		}
		return b.Layout(gtx, th, btn)
	}
}
