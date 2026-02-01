package widgets

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"

	"github.com/nishiki/frontend/ui/theme"
)

// Card renders a card container with rounded corners and background
type Card struct {
	BackgroundColor color.NRGBA
	CornerRadius    unit.Dp
	Inset           layout.Inset
}

// Layout renders the card with the provided content
func (c Card) Layout(gtx layout.Context, w layout.Widget) layout.Dimensions {
	// Default values if not specified
	if c.BackgroundColor == (color.NRGBA{}) {
		c.BackgroundColor = theme.ColorWhite
	}
	if c.CornerRadius == 0 {
		c.CornerRadius = unit.Dp(theme.RadiusDefault)
	}
	if c.Inset == (layout.Inset{}) {
		c.Inset = layout.UniformInset(unit.Dp(theme.Spacing4))
	}

	// Render content first to determine size
	macro := op.Record(gtx.Ops)
	dims := c.Inset.Layout(gtx, w)
	call := macro.Stop()

	// Draw background with the content's dimensions
	rr := gtx.Dp(c.CornerRadius)
	defer clip.UniformRRect(image.Rectangle{Max: dims.Size}, rr).Push(gtx.Ops).Pop()
	paint.ColorOp{Color: c.BackgroundColor}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)

	// Draw the content
	call.Add(gtx.Ops)

	return dims
}

// DefaultCard creates a card with default styling
func DefaultCard() Card {
	return Card{
		BackgroundColor: theme.ColorWhite,
		CornerRadius:    unit.Dp(theme.RadiusDefault),
		Inset:           layout.UniformInset(unit.Dp(theme.Spacing4)),
	}
}

// SurfaceCard creates a card with surface background (lighter)
func SurfaceCard() Card {
	return Card{
		BackgroundColor: theme.ColorGrayLightest,
		CornerRadius:    unit.Dp(theme.RadiusDefault),
		Inset:           layout.UniformInset(unit.Dp(theme.Spacing4)),
	}
}
