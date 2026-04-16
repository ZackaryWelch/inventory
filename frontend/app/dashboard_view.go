package app

import (
	"fmt"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
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
			return ga.renderBottomMenu(gtx, ViewDashboardGio)
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
						return ga.renderStatCard(gtx, strconv.Itoa(len(ga.groups)), "Groups", theme.ColorPrimary, theme.ColorWhite)
					})
				}),

				// Collections stat
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Left: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return ga.renderStatCard(gtx, strconv.Itoa(len(ga.collections)), "Collections", theme.ColorAccent, theme.ColorBlack)
					})
				}),
			)
		}),
	)
}

// renderStatCard renders a single stat card
func (ga *GioApp) renderStatCard(gtx layout.Context, value, label string, bgColor color.NRGBA, textColor color.NRGBA) layout.Dimensions {
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
				valueLabel.Color = textColor
				valueLabel.Alignment = text.Middle
				return valueLabel.Layout(gtx)
			}),

			// Label
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				labelWidget := material.Body2(ga.theme.Theme, label)
				labelWidget.Color = textColor
				labelWidget.Alignment = text.Middle
				return labelWidget.Layout(gtx)
			}),
		)
	})
}

// renderBottomMenu renders the bottom navigation menu and handles its click events.
// activeView identifies the current view so its tab is highlighted.
func (ga *GioApp) renderBottomMenu(gtx layout.Context, activeView ViewID) layout.Dimensions {
	// Menu items: clickable, label, target view
	type menuItem struct {
		btn    *widget.Clickable
		label  string
		target ViewID
	}
	items := []menuItem{
		{&ga.widgetState.menuDashboard, "Home", ViewDashboardGio},
		{&ga.widgetState.menuGroups, "Groups", ViewGroupsGio},
		{&ga.widgetState.menuCollections, "Collections", ViewCollectionsGio},
		{&ga.widgetState.menuProfile, "Profile", ViewProfileGio},
	}

	// Handle clicks — navigate to target view if not already active
	for _, item := range items {
		if item.btn.Clicked(gtx) && item.target != activeView {
			ga.currentView = item.target
		}
	}

	return layout.Inset{
		Top:    unit.Dp(theme.Spacing2),
		Bottom: unit.Dp(theme.Spacing2),
		Left:   unit.Dp(theme.Spacing2),
		Right:  unit.Dp(theme.Spacing2),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		card := widgets.Card{
			BackgroundColor: theme.ColorSurface,
			CornerRadius:    unit.Dp(theme.RadiusDefault),
			Inset:           layout.UniformInset(unit.Dp(theme.Spacing2)),
		}

		return card.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			children := make([]layout.FlexChild, len(items))
			for i, item := range items {
				children[i] = layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					variant := widgets.PrimaryButton
					if item.target == activeView {
						variant = widgets.AccentButton
					}
					return variant(ga.theme.Theme, item.btn, item.label)(gtx)
				})
			}
			return layout.Flex{
				Axis:    layout.Horizontal,
				Spacing: layout.SpaceEvenly,
			}.Layout(gtx, children...)
		})
	})
}
