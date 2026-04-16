package app

import (
	"fmt"
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"github.com/nishiki/frontend/ui/theme"
	"github.com/nishiki/frontend/ui/widgets"
)

// renderContainersPageView renders the dedicated containers management page.
// Shows a container list and, when a container is selected, its objects.
func (ga *GioApp) renderContainersPageView(gtx layout.Context) layout.Dimensions {
	if ga.selectedCollection == nil {
		ga.currentView = ViewCollectionDetailGio
		return layout.Dimensions{}
	}

	// Handle back button
	if ga.widgetState.containersBackButton.Clicked(gtx) {
		ga.currentView = ViewCollectionDetailGio
		ga.selectedContainer = nil
		return layout.Dimensions{}
	}

	// Handle create container button
	if ga.widgetState.createContainerButton.Clicked(gtx) {
		ga.showContainerDialog = true
		ga.containerDialogMode = "create"
		ga.selectedContainerID = nil
		ga.widgetState.containerNameEditor.SetText("")
		ga.widgetState.containerLocationEditor.SetText("")
		ga.selectedParentContainerID = nil
	}

	ga.ensureContainerItemStates()
	ga.ensureObjectItemStates()

	// Handle container selection clicks
	for i, c := range ga.containers {
		if i < len(ga.widgetState.containerItems) && ga.widgetState.containerItems[i].clickable.Clicked(gtx) {
			container := c
			ga.selectedContainer = &container
		}
	}

	return layout.Stack{}.Layout(gtx,
		// Main content
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				// Header
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return ga.renderContainersPageHeader(gtx)
				}),

				// Content: container list + detail
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{
						Top:    unit.Dp(theme.Spacing4),
						Bottom: unit.Dp(theme.Spacing20),
						Left:   unit.Dp(theme.Spacing4),
						Right:  unit.Dp(theme.Spacing4),
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceBetween}.Layout(gtx,
							// Container list (left)
							layout.Flexed(0.35, func(gtx layout.Context) layout.Dimensions {
								return layout.Inset{Right: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									return ga.renderContainersPageList(gtx)
								})
							}),
							// Container detail / objects (right)
							layout.Flexed(0.65, func(gtx layout.Context) layout.Dimensions {
								return layout.Inset{Left: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									return ga.renderContainerDetailPanel(gtx)
								})
							}),
						)
					})
				}),

				// Bottom menu
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return ga.renderBottomMenu(gtx, ViewCollectionsGio)
				}),
			)
		}),

		// Dialogs
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return ga.renderContainerDialog(gtx)
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return ga.renderDeleteContainerDialog(gtx)
		}),
	)
}

// renderContainersPageHeader renders the header for the containers page.
func (ga *GioApp) renderContainersPageHeader(gtx layout.Context) layout.Dimensions {
	return widgets.Card{
		BackgroundColor: theme.ColorPrimary,
		CornerRadius:    0,
		Inset: layout.Inset{
			Top:    unit.Dp(theme.Spacing4),
			Bottom: unit.Dp(theme.Spacing4),
			Left:   unit.Dp(theme.Spacing4),
			Right:  unit.Dp(theme.Spacing4),
		},
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
			// Back button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Right: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					btn := material.Button(ga.theme.Theme, &ga.widgetState.containersBackButton, "< Back")
					btn.Background = theme.ColorPrimaryDark
					btn.Color = theme.ColorWhite
					btn.CornerRadius = unit.Dp(theme.RadiusDefault)
					return btn.Layout(gtx)
				})
			}),
			// Title
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := material.H6(ga.theme.Theme, "Containers")
						label.Color = theme.ColorWhite
						label.Font.Weight = font.Bold
						return label.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := material.Body2(ga.theme.Theme, ga.selectedCollection.Name)
						label.Color = theme.ColorWhite
						return label.Layout(gtx)
					}),
				)
			}),
			// Create button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return widgets.PrimaryButton(ga.theme.Theme, &ga.widgetState.createContainerButton, "+ Container")(gtx)
			}),
		)
	})
}

// renderContainersPageList renders all containers (not filtered by objects) for the containers page.
func (ga *GioApp) renderContainersPageList(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Search
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				editor := material.Editor(ga.theme.Theme, &ga.widgetState.containersSearchField, "Search containers...")
				editor.Color = theme.ColorTextPrimary
				return editor.Layout(gtx)
			})
		}),

		// List
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			if len(ga.containers) == 0 {
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					label := material.Body1(ga.theme.Theme, "No containers yet")
					label.Color = theme.ColorTextSecondary
					label.Alignment = text.Middle
					return label.Layout(gtx)
				})
			}

			searchQuery := strings.ToLower(ga.widgetState.containersSearchField.Text())
			var filtered []int
			for i, c := range ga.containers {
				if searchQuery != "" &&
					!strings.Contains(strings.ToLower(c.Name), searchQuery) &&
					!strings.Contains(strings.ToLower(c.Type), searchQuery) &&
					!strings.Contains(strings.ToLower(c.Location), searchQuery) {
					continue
				}
				filtered = append(filtered, i)
			}

			list := &ga.widgetState.containersList
			list.Axis = layout.Vertical
			return list.Layout(gtx, len(filtered), func(gtx layout.Context, index int) layout.Dimensions {
				origIdx := filtered[index]
				container := ga.containers[origIdx]
				isSelected := ga.selectedContainer != nil && ga.selectedContainer.ID == container.ID

				return layout.Inset{Bottom: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return ga.renderContainersPageCard(gtx, container, origIdx, isSelected)
				})
			})
		}),
	)
}

// renderContainersPageCard renders a compact container card for the containers page list.
func (ga *GioApp) renderContainersPageCard(gtx layout.Context, container Container, index int, selected bool) layout.Dimensions {
	itemState := &ga.widgetState.containerItems[index]

	// Handle click to select
	if itemState.clickable.Clicked(gtx) {
		c := container
		ga.selectedContainer = &c
	}

	// Handle edit/delete
	if itemState.editButton.Clicked(gtx) {
		ga.selectedContainer = &container
		ga.showContainerDialog = true
		ga.containerDialogMode = "edit"
		ga.widgetState.containerNameEditor.SetText(container.Name)
		ga.widgetState.containerLocationEditor.SetText(container.Location)
		if container.ParentContainerID != nil {
			ga.selectedParentContainerID = container.ParentContainerID
		} else {
			ga.selectedParentContainerID = nil
		}
	}
	if itemState.deleteButton.Clicked(gtx) {
		ga.showDeleteContainer = true
		ga.deleteContainerID = container.ID
	}

	bg := theme.ColorSurface
	if selected {
		bg = theme.ColorSurfaceAlt
	}

	card := widgets.Card{
		BackgroundColor: bg,
		CornerRadius:    unit.Dp(theme.RadiusDefault),
		Inset: layout.Inset{
			Top:    unit.Dp(theme.Spacing2),
			Bottom: unit.Dp(theme.Spacing2),
			Left:   unit.Dp(theme.Spacing3),
			Right:  unit.Dp(theme.Spacing3),
		},
	}
	return material.Clickable(gtx, &itemState.clickable, func(gtx layout.Context) layout.Dimensions {
		return card.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				// Name + type badge
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							label := material.Body1(ga.theme.Theme, container.Name)
							label.Font.Weight = font.Bold
							return label.Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							label := material.Body2(ga.theme.Theme, containerTypeLabels[container.Type])
							label.Color = theme.ColorTextSecondary
							return label.Layout(gtx)
						}),
					)
				}),
				// Object count + actions
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Top: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								label := material.Body2(ga.theme.Theme, fmt.Sprintf("%d objects", container.ObjectCount))
								label.Color = theme.ColorTextSecondary
								return label.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layout.Inset{Right: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									return widgets.AccentButton(ga.theme.Theme, &itemState.editButton, "Edit")(gtx)
								})
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return widgets.DangerButton(ga.theme.Theme, &itemState.deleteButton, "Delete")(gtx)
							}),
						)
					})
				}),
			)
		})
	})
}

// renderContainerDetailPanel renders the selected container's objects.
func (ga *GioApp) renderContainerDetailPanel(gtx layout.Context) layout.Dimensions {
	if ga.selectedContainer == nil {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			label := material.Body1(ga.theme.Theme, "Select a container to view its objects")
			label.Color = theme.ColorTextSecondary
			label.Alignment = text.Middle
			return label.Layout(gtx)
		})
	}

	// Find objects for the selected container
	var containerObjects []Object
	var objectIndices []int
	for i, obj := range ga.objects {
		if obj.ContainerID == ga.selectedContainer.ID {
			containerObjects = append(containerObjects, obj)
			objectIndices = append(objectIndices, i)
		}
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Container detail header
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := material.H6(ga.theme.Theme, ga.selectedContainer.Name)
						label.Font.Weight = font.Bold
						return label.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						info := containerTypeLabels[ga.selectedContainer.Type]
						if ga.selectedContainer.Location != "" {
							info += " - " + ga.selectedContainer.Location
						}
						label := material.Body2(ga.theme.Theme, info)
						label.Color = theme.ColorTextSecondary
						return label.Layout(gtx)
					}),
				)
			})
		}),

		// Objects list
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			if len(containerObjects) == 0 {
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					label := material.Body1(ga.theme.Theme, "No objects in this container")
					label.Color = theme.ColorTextSecondary
					label.Alignment = text.Middle
					return label.Layout(gtx)
				})
			}

			list := &ga.widgetState.containerDetailList
			list.Axis = layout.Vertical
			return list.Layout(gtx, len(containerObjects), func(gtx layout.Context, index int) layout.Dimensions {
				obj := containerObjects[index]
				origIdx := objectIndices[index]
				return layout.Inset{Bottom: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return ga.renderObjectCard(gtx, obj, origIdx)
				})
			})
		}),
	)
}
