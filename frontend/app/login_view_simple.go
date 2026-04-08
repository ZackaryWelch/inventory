package app

import (
	"image"
	"image/color"
	"math"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"github.com/nishiki/frontend/ui/theme"
	"github.com/nishiki/frontend/ui/widgets"
)

// renderLoginViewSimple renders the login screen with a centered card layout
func (ga *GioApp) renderLoginViewSimple(gtx layout.Context) layout.Dimensions {
	ga.logger.Debug("Rendering simple login view")

	// Handle login button click
	if ga.widgetState.loginButton.Clicked(gtx) {
		ga.logger.Info("Login button clicked")
		ga.loginErrorMsg = ""
		ga.handleLogin()
	}

	// Paint background with subtle gradient effect
	paintLoginBackground(gtx)

	// Force min = max so layout.Center has the full window to position within.
	// Without this, Center constrains to child size, leaving no centering room.
	gtx.Constraints.Min = gtx.Constraints.Max

	// Center the login card vertically and horizontally
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Constrain the card width
		maxWidth := gtx.Dp(unit.Dp(420))
		if gtx.Constraints.Max.X < maxWidth {
			maxWidth = gtx.Constraints.Max.X - gtx.Dp(unit.Dp(32))
		}
		gtx.Constraints.Min.X = maxWidth
		gtx.Constraints.Max.X = maxWidth

		// Render the card with surface background
		return widgets.Card{
			BackgroundColor: theme.ColorSurface,
			CornerRadius:    unit.Dp(theme.RadiusXL),
			Inset:           layout.Inset{},
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Vertical,
				Alignment: layout.Middle,
			}.Layout(gtx,
				// Accent bar at top of card
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return paintAccentBar(gtx, maxWidth)
				}),

				// Card content with padding
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{
						Top:    unit.Dp(theme.Spacing8),
						Bottom: unit.Dp(theme.Spacing8),
						Left:   unit.Dp(theme.Spacing8),
						Right:  unit.Dp(theme.Spacing8),
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{
							Axis: layout.Vertical,
						}.Layout(gtx,
							// App name
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layout.Inset{Bottom: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									gtx.Constraints.Min.X = gtx.Constraints.Max.X
									l := material.Label(ga.theme.Theme, unit.Sp(36), "NISHIKI")
									l.Color = theme.ColorPrimary
									l.Alignment = text.Middle
									l.Font.Weight = font.Bold
									return l.Layout(gtx)
								})
							}),

							// Subtitle
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layout.Inset{Bottom: unit.Dp(theme.Spacing10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									gtx.Constraints.Min.X = gtx.Constraints.Max.X
									l := material.Body1(ga.theme.Theme, "Inventory Management System")
									l.Color = theme.ColorTextSecondary
									l.Alignment = text.Middle
									return l.Layout(gtx)
								})
							}),

							// Divider line
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layout.Inset{
									Bottom: unit.Dp(theme.Spacing10),
									Left:   unit.Dp(theme.Spacing10),
									Right:  unit.Dp(theme.Spacing10),
								}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									return paintDivider(gtx, theme.ColorBorder)
								})
							}),

							// Error message (if any)
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								if ga.loginErrorMsg == "" {
									return layout.Dimensions{}
								}
								return layout.Inset{
									Bottom: unit.Dp(theme.Spacing4),
									Left:   unit.Dp(theme.Spacing4),
									Right:  unit.Dp(theme.Spacing4),
								}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									gtx.Constraints.Min.X = gtx.Constraints.Max.X
									l := material.Body2(ga.theme.Theme, ga.loginErrorMsg)
									l.Color = theme.ColorDanger
									l.Alignment = text.Middle
									return l.Layout(gtx)
								})
							}),

							// Login button - centered
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layout.Inset{
									Bottom: unit.Dp(theme.Spacing6),
								}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										gtx.Constraints.Min.X = gtx.Dp(unit.Dp(280))
										return widgets.LoginButton(ga.theme.Theme, &ga.widgetState.loginButton, "Sign in with Authentik")(gtx)
									})
								})
							}),

							// Attribution
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min.X = gtx.Constraints.Max.X
								l := material.Label(ga.theme.Theme, unit.Sp(theme.FontSizeXS), "Secure authentication powered by Authentik")
								l.Color = color.NRGBA{
									R: theme.ColorTextSecondary.R,
									G: theme.ColorTextSecondary.G,
									B: theme.ColorTextSecondary.B,
									A: 160,
								}
								l.Alignment = text.Middle
								return l.Layout(gtx)
							}),
						)
					})
				}),
			)
		})
	})
}

// paintLoginBackground paints the login page background with a subtle gradient
func paintLoginBackground(gtx layout.Context) {
	size := gtx.Constraints.Max
	bg := theme.ColorBackground

	// Paint base background
	defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
	paint.ColorOp{Color: bg}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)

	// Paint a subtle radial-ish glow in the center using concentric rectangles
	// This creates a vignette effect that draws the eye to the center
	centerX := size.X / 2
	centerY := size.Y / 2
	steps := 5
	for i := steps; i >= 0; i-- {
		frac := float64(i) / float64(steps)
		// Shrink the rectangle as we go inward
		halfW := int(float64(size.X) * (0.3 + 0.7*frac) / 2)
		halfH := int(float64(size.Y) * (0.3 + 0.7*frac) / 2)
		rect := image.Rect(centerX-halfW, centerY-halfH, centerX+halfW, centerY+halfH)

		// Alpha decreases toward the edges (outermost is most transparent)
		alpha := uint8(math.Max(0, math.Min(255, 12*(float64(steps)-float64(i)))))

		// Tint with primary color
		glowColor := color.NRGBA{
			R: theme.ColorPrimary.R / 4,
			G: theme.ColorPrimary.G / 4,
			B: theme.ColorPrimary.B / 4,
			A: alpha,
		}

		paintRect(gtx, rect.Min, rect.Max, glowColor)
	}
}

// paintAccentBar paints a horizontal gradient accent bar at the top of the card
func paintAccentBar(gtx layout.Context, width int) layout.Dimensions {
	barHeight := gtx.Dp(unit.Dp(4))
	rr := gtx.Dp(unit.Dp(theme.RadiusXL))

	// Use a rounded rect that matches the card's top corners
	size := image.Point{X: width, Y: barHeight + rr}
	rect := image.Rect(0, 0, size.X, size.Y)

	// Clip to just the visible bar area (top portion)
	barRect := image.Rect(0, 0, width, barHeight)

	// Draw with rounded top corners by using a taller rounded rect clipped to the bar height
	func() {
		defer clip.UniformRRect(rect, rr).Push(gtx.Ops).Pop()
		// Clip further to just the visible bar height
		defer clip.Rect(barRect).Push(gtx.Ops).Pop()

		// Paint gradient from primary to accent using horizontal steps
		steps := 20
		for i := 0; i < steps; i++ {
			frac := float64(i) / float64(steps)
			x0 := int(float64(width) * frac)
			x1 := int(float64(width) * (frac + 1.0/float64(steps)))

			col := lerpColor(theme.ColorPrimary, theme.ColorAccent, float32(frac))
			paintRect(gtx, image.Pt(x0, 0), image.Pt(x1, barHeight), col)
		}
	}()

	// Advance by bar height only
	d := layout.Dimensions{Size: image.Point{X: width, Y: barHeight}}
	op.Offset(image.Point{}).Add(gtx.Ops)
	return d
}

// paintDivider paints a thin horizontal line
func paintDivider(gtx layout.Context, col color.NRGBA) layout.Dimensions {
	height := gtx.Dp(unit.Dp(1))
	width := gtx.Constraints.Max.X

	size := image.Point{X: width, Y: height}
	paintRect(gtx, image.Pt(0, 0), size, col)

	return layout.Dimensions{Size: size}
}

// paintRect paints a colored rectangle
func paintRect(gtx layout.Context, minSize, maxSize image.Point, col color.NRGBA) {
	rect := clip.Rect{Max: maxSize}
	if minSize.X > 0 && minSize.Y > 0 {
		rect.Min = minSize
	}
	defer rect.Push(gtx.Ops).Pop()
	paint.ColorOp{Color: col}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
}

// lerpColor linearly interpolates between two colors
func lerpColor(a, b color.NRGBA, t float32) color.NRGBA {
	return color.NRGBA{
		R: uint8(float32(a.R) + t*(float32(b.R)-float32(a.R))),
		G: uint8(float32(a.G) + t*(float32(b.G)-float32(a.G))),
		B: uint8(float32(a.B) + t*(float32(b.B)-float32(a.B))),
		A: uint8(float32(a.A) + t*(float32(b.A)-float32(a.A))),
	}
}
