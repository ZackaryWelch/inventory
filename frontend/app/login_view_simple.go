//go:build js && wasm

package app

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"github.com/nishiki/frontend/ui/theme"
	"github.com/nishiki/frontend/ui/widgets"
)

// renderLoginViewSimple renders a simplified login screen for debugging
func (ga *GioApp) renderLoginViewSimple(gtx layout.Context) layout.Dimensions {
	ga.logger.Debug("Rendering simple login view")

	// Handle login button click
	if ga.widgetState.loginButton.Clicked(gtx) {
		ga.logger.Info("Login button clicked")
		ga.handleLogin()
	}

	// Paint a colored background to ensure canvas is working
	paintRect(gtx, gtx.Constraints.Max, theme.ColorGrayLightest)

	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Vertical,
			Alignment: layout.Middle,
			Spacing:   layout.SpaceEvenly,
		}.Layout(gtx,
			// Title with explicit styling
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Top:    unit.Dp(theme.Spacing8),
					Bottom: unit.Dp(theme.Spacing8),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					// Create label with explicit font and color
					l := material.H1(ga.theme.Theme, "NISHIKI")
					l.Color = theme.ColorPrimary
					l.Alignment = text.Middle
					l.Font.Weight = font.Bold

					ga.logger.Debug("Rendering title", "text", "NISHIKI", "color", l.Color)
					dims := l.Layout(gtx)
					ga.logger.Debug("Title dimensions", "size", dims.Size)
					return dims
				})
			}),

			// Subtitle
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Bottom: unit.Dp(theme.Spacing4),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					l := material.Body1(ga.theme.Theme, "Inventory Management System")
					l.Color = theme.ColorTextSecondary
					l.Alignment = text.Middle
					return l.Layout(gtx)
				})
			}),

			// Login button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Top:    unit.Dp(theme.Spacing4),
					Bottom: unit.Dp(theme.Spacing4),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					// Constrain width
					gtx.Constraints.Min.X = gtx.Dp(unit.Dp(300))
					gtx.Constraints.Max.X = gtx.Dp(unit.Dp(400))

					return widgets.LoginButton(ga.theme.Theme, &ga.widgetState.loginButton, "Sign in with Authentik")(gtx)
				})
			}),

			// Caption
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				l := material.Caption(ga.theme.Theme, "Secure authentication powered by Authentik")
				l.Color = theme.ColorGray600
				l.Alignment = text.Middle
				return l.Layout(gtx)
			}),
		)
	})
}

// paintRect paints a colored rectangle
func paintRect(gtx layout.Context, size image.Point, col color.NRGBA) {
	defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
	paint.ColorOp{Color: col}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
}
