//go:build js && wasm

package app

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"github.com/nishiki/frontend/ui/theme"
	"github.com/nishiki/frontend/ui/widgets"
)

// renderProfileView renders the user profile view
func (ga *GioApp) renderProfileView(gtx layout.Context) layout.Dimensions {
	// Handle logout button click
	if ga.widgetState.logoutButton.Clicked(gtx) {
		ga.logger.Info("User logged out")
		ga.handleLogout()
	}

	// Handle bottom menu clicks
	if ga.widgetState.menuDashboard.Clicked(gtx) {
		ga.currentView = ViewDashboardGio
	}
	if ga.widgetState.menuGroups.Clicked(gtx) {
		ga.currentView = ViewGroupsGio
	}
	if ga.widgetState.menuCollections.Clicked(gtx) {
		ga.currentView = ViewCollectionsGio
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		// Header
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ga.renderHeader(gtx, "Profile")
		}),

		// Content
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top:    unit.Dp(theme.Spacing4),
				Bottom: unit.Dp(theme.Spacing20), // Space for bottom menu
				Left:   unit.Dp(theme.Spacing4),
				Right:  unit.Dp(theme.Spacing4),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis: layout.Vertical,
				}.Layout(gtx,
					// User info card
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if ga.currentUser != nil {
							return layout.Inset{
								Bottom: unit.Dp(theme.Spacing4),
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return ga.renderUserInfoCard(gtx)
							})
						}
						return layout.Dimensions{}
					}),

					// Logout button
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return widgets.DangerButton(ga.theme.Theme, &ga.widgetState.logoutButton, "Sign Out")(gtx)
					}),
				)
			})
		}),

		// Bottom menu
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ga.renderBottomMenu(gtx, "profile")
		}),
	)
}

// renderUserInfoCard renders the user information card
func (ga *GioApp) renderUserInfoCard(gtx layout.Context) layout.Dimensions {
	card := widgets.DefaultCard()

	return card.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			// Name
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Bottom: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis: layout.Vertical,
					}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							label := material.Body2(ga.theme.Theme, "Name:")
							label.Color = theme.ColorTextSecondary
							return label.Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							label := material.Body1(ga.theme.Theme, ga.currentUser.Name)
							return label.Layout(gtx)
						}),
					)
				})
			}),

			// Email
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis: layout.Vertical,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := material.Body2(ga.theme.Theme, "Email:")
						label.Color = theme.ColorTextSecondary
						return label.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := material.Body1(ga.theme.Theme, ga.currentUser.Email)
						return label.Layout(gtx)
					}),
				)
			}),
		)
	})
}

// handleLogout logs out the current user
func (ga *GioApp) handleLogout() {
	// Clear token from localStorage
	ga.authService.ClearToken()

	// Reset app state
	ga.currentUser = nil
	ga.groups = nil
	ga.collections = nil
	ga.isSignedIn = false

	// Navigate to login view
	ga.currentView = ViewLoginGio
	ga.window.Invalidate()
}
