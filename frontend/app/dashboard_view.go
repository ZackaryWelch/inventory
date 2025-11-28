//go:build js && wasm

package app

import (
	"fmt"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"github.com/nishiki/frontend/ui/theme"
	"github.com/nishiki/frontend/ui/widgets"
)

// renderDashboardView renders the dashboard with stats and navigation
func (ga *GioApp) renderDashboardView(gtx layout.Context) layout.Dimensions {
	// Handle button clicks
	if ga.widgetState.groupsButton.Clicked(gtx) {
		ga.logger.Info("Navigating to groups view")
		ga.currentView = ViewGroupsGio
	}
	if ga.widgetState.collectionsButton.Clicked(gtx) {
		ga.logger.Info("Navigating to collections view")
		ga.currentView = ViewCollectionsGio
	}
	if ga.widgetState.profileButton.Clicked(gtx) {
		ga.logger.Info("Navigating to profile view")
		ga.currentView = ViewProfileGio
	}
	if ga.widgetState.searchButton.Clicked(gtx) {
		ga.logger.Info("Navigating to search view")
		ga.currentView = ViewSearchGio
	}

	// Bottom menu navigation
	if ga.widgetState.menuGroups.Clicked(gtx) {
		ga.currentView = ViewGroupsGio
	}
	if ga.widgetState.menuCollections.Clicked(gtx) {
		ga.currentView = ViewCollectionsGio
	}
	if ga.widgetState.menuProfile.Clicked(gtx) {
		ga.currentView = ViewProfileGio
	}

	// Main layout
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		// Header
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ga.renderHeader(gtx, "Dashboard")
		}),

		// Content area with scrolling
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
					// Navigation buttons
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{
							Bottom: unit.Dp(theme.Spacing6),
						}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return ga.renderNavigationButtons(gtx)
						})
					}),

					// Stats section
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return ga.renderStats(gtx)
					}),
				)
			})
		}),

		// Bottom navigation menu
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ga.renderBottomMenu(gtx, "dashboard")
		}),
	)
}

// renderHeader renders a page header with title
func (ga *GioApp) renderHeader(gtx layout.Context, title string) layout.Dimensions {
	return layout.Inset{
		Top:    unit.Dp(theme.Spacing4),
		Bottom: unit.Dp(theme.Spacing4),
		Left:   unit.Dp(theme.Spacing4),
		Right:  unit.Dp(theme.Spacing4),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
			Spacing:   layout.SpaceBetween,
		}.Layout(gtx,
			// Title
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				label := material.H5(ga.theme.Theme, title)
				label.Font.Weight = font.Bold
				return label.Layout(gtx)
			}),

			// Username (if available)
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if ga.currentUser != nil {
					label := material.Body2(ga.theme.Theme, ga.currentUser.Name)
					label.Color = theme.ColorTextSecondary
					return label.Layout(gtx)
				}
				return layout.Dimensions{}
			}),
		)
	})
}

// renderNavigationButtons renders the main navigation buttons
func (ga *GioApp) renderNavigationButtons(gtx layout.Context) layout.Dimensions {
	return layout.Flex{
		Axis:    layout.Vertical,
		Spacing: layout.SpaceEvenly,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return widgets.PrimaryButton(ga.theme.Theme, &ga.widgetState.groupsButton, "Groups")(gtx)
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return widgets.PrimaryButton(ga.theme.Theme, &ga.widgetState.collectionsButton, "Collections")(gtx)
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return widgets.PrimaryButton(ga.theme.Theme, &ga.widgetState.profileButton, "Profile")(gtx)
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return widgets.PrimaryButton(ga.theme.Theme, &ga.widgetState.searchButton, "Search")(gtx)
		}),
	)
}

// renderStats renders the statistics cards
func (ga *GioApp) renderStats(gtx layout.Context) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		// Stats title
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Bottom: unit.Dp(theme.Spacing4),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				label := material.H6(ga.theme.Theme, "Quick Stats")
				return label.Layout(gtx)
			})
		}),

		// Stats grid
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:    layout.Horizontal,
				Spacing: layout.SpaceEvenly,
			}.Layout(gtx,
				// Groups stat
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Right: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return ga.renderStatCard(gtx, fmt.Sprintf("%d", len(ga.groups)), "Groups", theme.ColorPrimary)
					})
				}),

				// Collections stat
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Left: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return ga.renderStatCard(gtx, fmt.Sprintf("%d", len(ga.collections)), "Collections", theme.ColorAccent)
					})
				}),
			)
		}),
	)
}

// renderStatCard renders a single stat card
func (ga *GioApp) renderStatCard(gtx layout.Context, value, label string, bgColor color.NRGBA) layout.Dimensions {
	card := widgets.DefaultCard()
	card.BackgroundColor = bgColor

	return card.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Vertical,
			Alignment: layout.Middle,
		}.Layout(gtx,
			// Value
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				valueLabel := material.H4(ga.theme.Theme, value)
				valueLabel.Color = theme.ColorWhite
				valueLabel.Alignment = text.Middle
				return valueLabel.Layout(gtx)
			}),

			// Label
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				labelWidget := material.Body2(ga.theme.Theme, label)
				labelWidget.Color = theme.ColorWhite
				labelWidget.Alignment = text.Middle
				return labelWidget.Layout(gtx)
			}),
		)
	})
}

// renderBottomMenu renders the bottom navigation menu
func (ga *GioApp) renderBottomMenu(gtx layout.Context, activeView string) layout.Dimensions {
	return layout.Inset{
		Top:    unit.Dp(theme.Spacing2),
		Bottom: unit.Dp(theme.Spacing2),
		Left:   unit.Dp(theme.Spacing2),
		Right:  unit.Dp(theme.Spacing2),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Background
		card := widgets.Card{
			BackgroundColor: theme.ColorWhite,
			CornerRadius:    unit.Dp(theme.RadiusDefault),
			Inset:           layout.UniformInset(unit.Dp(theme.Spacing2)),
		}

		return card.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:    layout.Horizontal,
				Spacing: layout.SpaceEvenly,
			}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					variant := widgets.PrimaryButton
					if activeView == "dashboard" {
						variant = widgets.AccentButton
					}
					return variant(ga.theme.Theme, &ga.widgetState.menuDashboard, "Home")(gtx)
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					variant := widgets.PrimaryButton
					if activeView == "groups" {
						variant = widgets.AccentButton
					}
					return variant(ga.theme.Theme, &ga.widgetState.menuGroups, "Groups")(gtx)
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					variant := widgets.PrimaryButton
					if activeView == "collections" {
						variant = widgets.AccentButton
					}
					return variant(ga.theme.Theme, &ga.widgetState.menuCollections, "Collections")(gtx)
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					variant := widgets.PrimaryButton
					if activeView == "profile" {
						variant = widgets.AccentButton
					}
					return variant(ga.theme.Theme, &ga.widgetState.menuProfile, "Profile")(gtx)
				}),
			)
		})
	})
}
