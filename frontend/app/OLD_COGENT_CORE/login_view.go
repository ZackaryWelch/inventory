//go:build js && wasm

package app

import (
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/nishiki/frontend/ui/theme"
	"github.com/nishiki/frontend/ui/widgets"
)

// LoginViewState holds widget state for the login view
type LoginViewState struct {
	loginButton widget.Clickable
}

var loginState = &LoginViewState{}

// renderLoginView renders the login screen
func (ga *GioApp) renderLoginView(gtx layout.Context) layout.Dimensions {
	ga.logger.Debug("Rendering login view", "constraints", gtx.Constraints)

	// Handle login button click
	if loginState.loginButton.Clicked(gtx) {
		ga.logger.Info("Login button clicked, initiating OAuth flow")
		ga.handleLogin()
	}

	// Center the content
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		ga.logger.Debug("Inside centered layout", "constraints", gtx.Constraints)

		return layout.Flex{
			Axis:    layout.Vertical,
			Spacing: layout.SpaceEvenly,
		}.Layout(gtx,
			// Logo/Title
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				ga.logger.Debug("Rendering logo")
				return layout.Inset{
					Bottom: unit.Dp(theme.Spacing8),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					label := material.H1(ga.theme.Theme, "NISHIKI")
					label.Color = theme.ColorPrimary
					label.Alignment = text.Middle
					ga.logger.Debug("Rendering H1 label", "text", "NISHIKI", "color", theme.ColorPrimary)
					dims := label.Layout(gtx)
					ga.logger.Debug("H1 dimensions", "dims", dims)
					return dims
				})
			}),

			// Subtitle
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Bottom: unit.Dp(theme.Spacing12),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					label := material.Body1(ga.theme.Theme, "Inventory Management System")
					label.Color = theme.ColorTextSecondary
					label.Alignment = text.Middle
					return label.Layout(gtx)
				})
			}),

			// Login button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Bottom: unit.Dp(theme.Spacing6),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					// Constrain button width
					gtx.Constraints.Min.X = gtx.Dp(unit.Dp(300))
					gtx.Constraints.Max.X = gtx.Dp(unit.Dp(400))

					return widgets.LoginButton(ga.theme.Theme, &loginState.loginButton, "Sign in with Authentik")(gtx)
				})
			}),

			// Authentication info
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Caption(ga.theme.Theme, "Secure authentication powered by Authentik")
				label.Color = theme.ColorGray600
				label.Alignment = text.Middle
				return label.Layout(gtx)
			}),
		)
	})
}

// handleLogin initiates the OAuth2 login flow
func (ga *GioApp) handleLogin() {
	ga.logger.Info("Starting OAuth2 PKCE flow")

	// Initiate login (generates authorization URL and redirects)
	if err := ga.authService.InitiateLogin(); err != nil {
		ga.logger.Error("Failed to initiate login", "error", err)
		return
	}
}

// renderCallbackView renders a loading screen during OAuth callback
func (ga *GioApp) renderCallbackView(gtx layout.Context) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Vertical,
			Alignment: layout.Middle,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Bottom: unit.Dp(theme.Spacing6),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					label := material.H3(ga.theme.Theme, "Completing Sign In...")
					label.Alignment = text.Middle
					return label.Layout(gtx)
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Body1(ga.theme.Theme, "Please wait while we authenticate you with Authentik.")
				label.Color = theme.ColorTextSecondary
				label.Alignment = text.Middle
				return label.Layout(gtx)
			}),
		)
	})
}

