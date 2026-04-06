package app

import (
	"fmt"
	"sort"
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/nishiki/frontend/ui/theme"
	"github.com/nishiki/frontend/ui/widgets"
)

// Container type constants
const (
	ContainerTypeRoom      = "room"
	ContainerTypeBookshelf = "bookshelf"
	ContainerTypeShelf     = "shelf"
	ContainerTypeBinder    = "binder"
	ContainerTypeCabinet   = "cabinet"
	ContainerTypeGeneral   = "general"
)

var containerTypes = []string{
	ContainerTypeRoom,
	ContainerTypeBookshelf,
	ContainerTypeShelf,
	ContainerTypeBinder,
	ContainerTypeCabinet,
	ContainerTypeGeneral,
}

var containerTypeLabels = map[string]string{
	ContainerTypeRoom:      "Room",
	ContainerTypeBookshelf: "Bookshelf",
	ContainerTypeShelf:     "Shelf",
	ContainerTypeBinder:    "Binder",
	ContainerTypeCabinet:   "Cabinet",
	ContainerTypeGeneral:   "General",
}

// renderCollectionDetailView renders the collection detail view with containers and objects
func (ga *GioApp) renderCollectionDetailView(gtx layout.Context) layout.Dimensions {
	if ga.selectedCollection == nil {
		// Navigate back to collections if no collection is selected
		ga.currentView = ViewCollectionsGio
		return layout.Dimensions{}
	}

	// Handle back button click
	if ga.widgetState.backToCollections.Clicked(gtx) {
		ga.currentView = ViewCollectionsGio
		ga.selectedCollection = nil
		ga.containers = nil
		ga.objects = nil
		ga.activeGroupedTextFilters = nil
		ga.showContainersPanel = false
		ga.containerViewMode = ""
		return layout.Dimensions{}
	}

	// Handle create container button
	if ga.widgetState.createContainerButton.Clicked(gtx) {
		ga.logger.Info("Opening create container dialog")
		ga.showContainerDialog = true
		ga.containerDialogMode = "create"
		ga.selectedContainerID = nil
		ga.widgetState.containerNameEditor.SetText("")
		ga.widgetState.containerLocationEditor.SetText("")
		ga.selectedParentContainerID = nil
	}

	// Handle create object button
	if ga.widgetState.createObjectButton.Clicked(gtx) {
		ga.logger.Info("Opening create object dialog")
		ga.showObjectDialog = true
		ga.objectDialogMode = "create"
		ga.selectedContainerID = nil
		ga.widgetState.objectNameEditor.SetText("")
		ga.widgetState.objectDescriptionEditor.SetText("")
		ga.widgetState.objectQuantityEditor.SetText("")
		ga.widgetState.objectUnitEditor.SetText("")
	}

	// Handle import button
	if ga.widgetState.importButton.Clicked(gtx) {
		ga.logger.Info("Opening file picker for import")
		go ga.SelectImportFile()
	}

	// Handle edit schema button
	if ga.widgetState.editSchemaButton.Clicked(gtx) {
		ga.openSchemaEditor()
	}

	// Handle container panel toggle
	if ga.widgetState.toggleContainersButton.Clicked(gtx) {
		ga.showContainersPanel = !ga.showContainersPanel
		if ga.showContainersPanel && ga.containerViewMode == "" {
			ga.containerViewMode = "split"
		}
	}

	// Handle container view mode toggle
	if ga.widgetState.containerViewSplitBtn.Clicked(gtx) {
		ga.containerViewMode = "split"
	}
	if ga.widgetState.containerViewGroupBtn.Clicked(gtx) {
		ga.containerViewMode = "grouped"
	}

	// Ensure we have widget states
	ga.ensureContainerItemStates()
	ga.ensureObjectItemStates()

	// Main layout with dialogs overlay
	return layout.Stack{}.Layout(gtx,
		// Main content
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				// Header with collection name and back button
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return ga.renderCollectionDetailHeader(gtx)
				}),

				// Content area with containers and objects
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{
						Top:    unit.Dp(theme.Spacing4),
						Bottom: unit.Dp(theme.Spacing20), // Space for bottom menu
						Left:   unit.Dp(theme.Spacing4),
						Right:  unit.Dp(theme.Spacing4),
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						if ga.showContainersPanel && ga.containerViewMode == "grouped" {
							// Grouped mode: objects grouped under container headers
							return ga.renderObjectsGroupedByContainer(gtx)
						}
						if ga.showContainersPanel && ga.containerViewMode == "split" {
							// Split view: Containers on left (≤1/3), Objects on right
							return layout.Flex{
								Axis:    layout.Horizontal,
								Spacing: layout.SpaceBetween,
							}.Layout(gtx,
								layout.Flexed(0.3, func(gtx layout.Context) layout.Dimensions {
									return layout.Inset{Right: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										return ga.renderContainersColumn(gtx)
									})
								}),
								layout.Flexed(0.7, func(gtx layout.Context) layout.Dimensions {
									return layout.Inset{Left: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										return ga.renderObjectsColumn(gtx)
									})
								}),
							)
						}
						// Default: objects only (full width)
						return ga.renderObjectsColumn(gtx)
					})
				}),

				// Bottom navigation menu
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return ga.renderBottomMenu(gtx, ViewCollectionsGio)
				}),
			)
		}),

		// Dialogs overlay
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return ga.renderContainerDialog(gtx)
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return ga.renderObjectDialog(gtx)
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return ga.renderDeleteContainerDialog(gtx)
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return ga.renderDeleteObjectDialog(gtx)
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return ga.renderImportPreviewDialog(gtx)
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return ga.renderSchemaEditorDialog(gtx)
		}),
	)
}

// renderCollectionDetailHeader renders the header with collection info and back button
func (ga *GioApp) renderCollectionDetailHeader(gtx layout.Context) layout.Dimensions {
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
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
		}.Layout(gtx,
			// Back button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Right: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					btn := material.Button(ga.theme.Theme, &ga.widgetState.backToCollections, "← Back")
					btn.Background = theme.ColorPrimaryDark
					btn.Color = theme.ColorWhite
					btn.CornerRadius = unit.Dp(theme.RadiusDefault)
					return btn.Layout(gtx)
				})
			}),

			// Collection name and info
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := material.H6(ga.theme.Theme, ga.selectedCollection.Name)
						label.Color = theme.ColorWhite
						label.Font.Weight = font.Bold
						return label.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						info := fmt.Sprintf("%s • %s", objectTypeLabels[ga.selectedCollection.ObjectType], ga.selectedCollection.Location)
						label := material.Body2(ga.theme.Theme, info)
						label.Color = theme.ColorWhite
						return label.Layout(gtx)
					}),
				)
			}),

			// Containers page button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if ga.widgetState.containersPageButton.Clicked(gtx) {
					ga.currentView = ViewContainersGio
					ga.selectedContainer = nil
				}
				return layout.Inset{Left: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					btn := material.Button(ga.theme.Theme, &ga.widgetState.containersPageButton, "Manage")
					btn.Background = theme.ColorPrimaryDark
					btn.Color = theme.ColorWhite
					btn.CornerRadius = unit.Dp(theme.RadiusDefault)
					return btn.Layout(gtx)
				})
			}),

			// Containers toggle button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Left: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					label := "Containers"
					bg := theme.ColorPrimaryDark
					if ga.showContainersPanel {
						bg = theme.ColorAccent
					}
					btn := material.Button(ga.theme.Theme, &ga.widgetState.toggleContainersButton, label)
					btn.Background = bg
					if ga.showContainersPanel {
						btn.Color = theme.ColorBlack
					} else {
						btn.Color = theme.ColorWhite
					}
					btn.CornerRadius = unit.Dp(theme.RadiusDefault)
					return btn.Layout(gtx)
				})
			}),

			// Import button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Left: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					btn := material.Button(ga.theme.Theme, &ga.widgetState.importButton, "Import")
					btn.Background = theme.ColorAccent
					btn.Color = theme.ColorBlack
					btn.CornerRadius = unit.Dp(theme.RadiusDefault)
					return btn.Layout(gtx)
				})
			}),

			// Edit Schema button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Left: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					btn := material.Button(ga.theme.Theme, &ga.widgetState.editSchemaButton, "Schema")
					btn.Background = theme.ColorPrimaryDark
					btn.Color = theme.ColorWhite
					btn.CornerRadius = unit.Dp(theme.RadiusDefault)
					return btn.Layout(gtx)
				})
			}),
		)
	})
}

// renderContainersColumn renders the containers list column
func (ga *GioApp) renderContainersColumn(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Header
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
					Spacing:   layout.SpaceBetween,
				}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						label := material.H6(ga.theme.Theme, "Containers")
						label.Font.Weight = font.Bold
						return label.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return widgets.PrimaryButton(ga.theme.Theme, &ga.widgetState.createContainerButton, "+")(gtx)
					}),
				)
			})
		}),

		// Search field
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				editor := material.Editor(ga.theme.Theme, &ga.widgetState.containersSearchField, "Search containers...")
				editor.Color = theme.ColorTextPrimary
				return editor.Layout(gtx)
			})
		}),

		// Containers list
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return ga.renderContainersList(gtx)
		}),
	)
}

// renderContainerViewModeToggle renders split/grouped toggle chips when the container panel is visible.
func (ga *GioApp) renderContainerViewModeToggle(gtx layout.Context) layout.Dimensions {
	if !ga.showContainersPanel {
		return layout.Dimensions{}
	}
	return layout.Inset{Bottom: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Right: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return ga.renderFilterChip(gtx, &ga.widgetState.containerViewSplitBtn, "Split", ga.containerViewMode == "split")
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ga.renderFilterChip(gtx, &ga.widgetState.containerViewGroupBtn, "Grouped", ga.containerViewMode == "grouped")
			}),
		)
	})
}

// renderObjectsColumn renders the objects list column
func (ga *GioApp) renderObjectsColumn(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Header
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
					Spacing:   layout.SpaceBetween,
				}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						label := material.H6(ga.theme.Theme, "Objects")
						label.Font.Weight = font.Bold
						return label.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return widgets.PrimaryButton(ga.theme.Theme, &ga.widgetState.createObjectButton, "+")(gtx)
					}),
				)
			})
		}),

		// Container view mode toggle
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ga.renderContainerViewModeToggle(gtx)
		}),

		// Grouped-text filter chips
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ga.renderGroupedTextFilters(gtx)
		}),

		// Search field
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				editor := material.Editor(ga.theme.Theme, &ga.widgetState.objectsSearchField, "Search objects...")
				editor.Color = theme.ColorTextPrimary
				return editor.Layout(gtx)
			})
		}),

		// Objects list
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return ga.renderObjectsList(gtx)
		}),
	)
}

// ensureContainerItemStates ensures we have widget states for all containers
func (ga *GioApp) ensureContainerItemStates() {
	if len(ga.widgetState.containerItems) != len(ga.containers) {
		ga.widgetState.containerItems = make([]ContainerItemState, len(ga.containers))
	}
}

// ensureObjectItemStates ensures we have widget states for all objects
func (ga *GioApp) ensureObjectItemStates() {
	if len(ga.widgetState.objectItems) != len(ga.objects) {
		ga.widgetState.objectItems = make([]ObjectItemState, len(ga.objects))
	}
}

// containersWithObjects returns the set of container IDs that have at least one object.
func (ga *GioApp) containersWithObjects() map[string]bool {
	m := make(map[string]bool)
	for _, obj := range ga.objects {
		if obj.ContainerID != "" {
			m[obj.ContainerID] = true
		}
	}
	return m
}

// renderContainersList renders the list of containers
func (ga *GioApp) renderContainersList(gtx layout.Context) layout.Dimensions {
	if len(ga.containers) == 0 {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			label := material.Body1(ga.theme.Theme, "No containers yet")
			label.Color = theme.ColorTextSecondary
			label.Alignment = text.Middle
			return label.Layout(gtx)
		})
	}

	// Filter containers based on search query and whether they have objects
	searchQuery := strings.ToLower(ga.widgetState.containersSearchField.Text())
	hasObjects := ga.containersWithObjects()
	filteredContainers := make([]Container, 0)
	filteredIndices := make([]int, 0)

	for i, container := range ga.containers {
		// In collection detail split view, only show containers with objects
		if !hasObjects[container.ID] {
			continue
		}
		if searchQuery == "" ||
			strings.Contains(strings.ToLower(container.Name), searchQuery) ||
			strings.Contains(strings.ToLower(container.Type), searchQuery) ||
			strings.Contains(strings.ToLower(container.Location), searchQuery) {
			filteredContainers = append(filteredContainers, container)
			filteredIndices = append(filteredIndices, i)
		}
	}

	// Render list
	list := &ga.widgetState.containersList
	list.Axis = layout.Vertical
	return list.Layout(gtx, len(filteredContainers), func(gtx layout.Context, index int) layout.Dimensions {
		container := filteredContainers[index]
		originalIndex := filteredIndices[index]
		return layout.Inset{Bottom: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return ga.renderContainerCard(gtx, container, originalIndex)
		})
	})
}

// renderObjectsList renders the list of objects
func (ga *GioApp) renderObjectsList(gtx layout.Context) layout.Dimensions {
	if len(ga.objects) == 0 {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			label := material.Body1(ga.theme.Theme, "No objects yet")
			label.Color = theme.ColorTextSecondary
			label.Alignment = text.Middle
			return label.Layout(gtx)
		})
	}

	// Filter objects based on search query
	searchQuery := strings.ToLower(ga.widgetState.objectsSearchField.Text())
	filteredObjects := make([]Object, 0)
	filteredIndices := make([]int, 0)

	for i, object := range ga.objects {
		if searchQuery != "" &&
			!strings.Contains(strings.ToLower(object.Name), searchQuery) &&
			!strings.Contains(strings.ToLower(object.Description), searchQuery) {
			continue
		}
		if !ga.matchesGroupedTextFilters(object) {
			continue
		}
		filteredObjects = append(filteredObjects, object)
		filteredIndices = append(filteredIndices, i)
	}

	// Render list
	list := &ga.widgetState.objectsList
	list.Axis = layout.Vertical
	return list.Layout(gtx, len(filteredObjects), func(gtx layout.Context, index int) layout.Dimensions {
		object := filteredObjects[index]
		originalIndex := filteredIndices[index]
		return layout.Inset{Bottom: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return ga.renderObjectCard(gtx, object, originalIndex)
		})
	})
}

// renderContainerCard renders a single container card
func (ga *GioApp) renderContainerCard(gtx layout.Context, container Container, index int) layout.Dimensions {
	itemState := &ga.widgetState.containerItems[index]

	// Handle edit button click
	if itemState.editButton.Clicked(gtx) {
		ga.logger.Info("Opening edit container dialog", "container_id", container.ID)
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

	// Handle delete button click
	if itemState.deleteButton.Clicked(gtx) {
		ga.logger.Info("Opening delete confirmation", "container_id", container.ID)
		ga.showDeleteContainer = true
		ga.deleteContainerID = container.ID
	}

	card := widgets.DefaultCard()
	return card.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Container name and type
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
					Spacing:   layout.SpaceBetween,
				}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						label := material.Body1(ga.theme.Theme, container.Name)
						label.Font.Weight = font.Bold
						return label.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						badge := widgets.Card{
							BackgroundColor: theme.ColorAccent,
							CornerRadius:    unit.Dp(theme.RadiusFull),
							Inset: layout.Inset{
								Top:    unit.Dp(theme.Spacing1),
								Bottom: unit.Dp(theme.Spacing1),
								Left:   unit.Dp(theme.Spacing2),
								Right:  unit.Dp(theme.Spacing2),
							},
						}
						return badge.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							label := material.Body2(ga.theme.Theme, containerTypeLabels[container.Type])
							label.Color = theme.ColorBlack
							return label.Layout(gtx)
						})
					}),
				)
			}),

			// Location
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if container.Location != "" {
					return layout.Inset{Top: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						label := material.Body2(ga.theme.Theme, container.Location)
						label.Color = theme.ColorTextSecondary
						return label.Layout(gtx)
					})
				}
				return layout.Dimensions{}
			}),

			// Object count
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				objectCount := len(container.Objects)
				return layout.Inset{Top: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					label := material.Body2(ga.theme.Theme, fmt.Sprintf("%d objects", objectCount))
					label.Color = theme.ColorTextSecondary
					return label.Layout(gtx)
				})
			}),

			// Action buttons
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis:    layout.Horizontal,
						Spacing: layout.SpaceStart,
					}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Right: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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
}

// renderObjectCard renders a single object card
func (ga *GioApp) renderObjectCard(gtx layout.Context, object Object, index int) layout.Dimensions {
	itemState := &ga.widgetState.objectItems[index]

	// Handle edit button click
	if itemState.editButton.Clicked(gtx) {
		ga.logger.Info("Opening edit object dialog", "object_id", object.ID)
		ga.selectedObject = &object
		ga.showObjectDialog = true
		ga.objectDialogMode = "edit"
		ga.widgetState.objectNameEditor.SetText(object.Name)
		ga.widgetState.objectDescriptionEditor.SetText(object.Description)
		if object.Quantity != nil {
			ga.widgetState.objectQuantityEditor.SetText(fmt.Sprintf("%v", *object.Quantity))
		} else {
			ga.widgetState.objectQuantityEditor.SetText("")
		}
		ga.widgetState.objectUnitEditor.SetText(object.Unit)
		if object.ContainerID != "" {
			cid := object.ContainerID
			ga.selectedContainerID = &cid
		} else {
			ga.selectedContainerID = nil
		}
	}

	// Handle delete button click
	if itemState.deleteButton.Clicked(gtx) {
		ga.logger.Info("Opening delete confirmation", "object_id", object.ID)
		ga.showDeleteObject = true
		ga.deleteObjectID = object.ID
	}

	card := widgets.DefaultCard()
	return card.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Object name
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Body1(ga.theme.Theme, object.Name)
				label.Font.Weight = font.Bold
				return label.Layout(gtx)
			}),

			// Description
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if object.Description != "" {
					return layout.Inset{Top: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						label := material.Body2(ga.theme.Theme, object.Description)
						label.Color = theme.ColorTextSecondary
						return label.Layout(gtx)
					})
				}
				return layout.Dimensions{}
			}),

			// Quantity
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if object.Quantity != nil && object.Unit != "" {
					return layout.Inset{Top: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						label := material.Body2(ga.theme.Theme, fmt.Sprintf("Qty: %v %s", *object.Quantity, object.Unit))
						label.Color = theme.ColorTextSecondary
						return label.Layout(gtx)
					})
				}
				return layout.Dimensions{}
			}),

			// Tags
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if len(object.Tags) > 0 {
					return layout.Inset{Top: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						tagsText := strings.Join(object.Tags, ", ")
						label := material.Body2(ga.theme.Theme, "Tags: "+tagsText)
						label.Color = theme.ColorTextSecondary
						return label.Layout(gtx)
					})
				}
				return layout.Dimensions{}
			}),

			// Properties
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if len(object.Properties) == 0 {
					return layout.Dimensions{}
				}
				var defs []PropertyDefinition
				if ga.selectedCollection != nil && ga.selectedCollection.PropertySchema != nil {
					defs = ga.selectedCollection.PropertySchema.Definitions
				}
				return layout.Inset{Top: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return ga.renderObjectProperties(gtx, object.Properties, defs)
				})
			}),

			// Action buttons
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis:    layout.Horizontal,
						Spacing: layout.SpaceStart,
					}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Right: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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
}

// renderObjectsGroupedByContainer renders all objects grouped under container section headers.
func (ga *GioApp) renderObjectsGroupedByContainer(gtx layout.Context) layout.Dimensions {
	ga.ensureObjectItemStates()

	// Build container → objects map
	containerObjects := make(map[string][]int) // containerID → indices into ga.objects
	var unassigned []int
	searchQuery := strings.ToLower(ga.widgetState.objectsSearchField.Text())

	for i, obj := range ga.objects {
		if searchQuery != "" &&
			!strings.Contains(strings.ToLower(obj.Name), searchQuery) &&
			!strings.Contains(strings.ToLower(obj.Description), searchQuery) {
			continue
		}
		if !ga.matchesGroupedTextFilters(obj) {
			continue
		}
		if obj.ContainerID == "" {
			unassigned = append(unassigned, i)
		} else {
			containerObjects[obj.ContainerID] = append(containerObjects[obj.ContainerID], i)
		}
	}

	// Build ordered container list (only those with matching objects)
	type containerGroup struct {
		name    string
		indices []int
	}
	var groups []containerGroup
	for _, c := range ga.containers {
		if indices, ok := containerObjects[c.ID]; ok && len(indices) > 0 {
			groups = append(groups, containerGroup{name: c.Name, indices: indices})
		}
	}
	if len(unassigned) > 0 {
		groups = append(groups, containerGroup{name: "Unassigned", indices: unassigned})
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Header row
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						label := material.H6(ga.theme.Theme, "Objects")
						label.Font.Weight = font.Bold
						return label.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return widgets.PrimaryButton(ga.theme.Theme, &ga.widgetState.createObjectButton, "+")(gtx)
					}),
				)
			})
		}),

		// View mode toggle
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ga.renderContainerViewModeToggle(gtx)
		}),

		// Grouped-text filters
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ga.renderGroupedTextFilters(gtx)
		}),

		// Search field
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				editor := material.Editor(ga.theme.Theme, &ga.widgetState.objectsSearchField, "Search objects...")
				editor.Color = theme.ColorTextPrimary
				return editor.Layout(gtx)
			})
		}),

		// Grouped list
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			if len(groups) == 0 {
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					label := material.Body1(ga.theme.Theme, "No objects yet")
					label.Color = theme.ColorTextSecondary
					label.Alignment = text.Middle
					return label.Layout(gtx)
				})
			}

			// Flatten groups into a list of items (headers + objects)
			type listItem struct {
				isHeader bool
				header   string
				objIndex int
			}
			var items []listItem
			for _, g := range groups {
				items = append(items, listItem{isHeader: true, header: fmt.Sprintf("%s (%d)", g.name, len(g.indices))})
				for _, idx := range g.indices {
					items = append(items, listItem{objIndex: idx})
				}
			}

			list := &ga.widgetState.objectsList
			list.Axis = layout.Vertical
			return list.Layout(gtx, len(items), func(gtx layout.Context, index int) layout.Dimensions {
				item := items[index]
				if item.isHeader {
					return layout.Inset{Top: unit.Dp(theme.Spacing3), Bottom: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						label := material.Body1(ga.theme.Theme, item.header)
						label.Font.Weight = font.Bold
						label.Color = theme.ColorTextSecondary
						return label.Layout(gtx)
					})
				}
				obj := ga.objects[item.objIndex]
				return layout.Inset{Bottom: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return ga.renderObjectCard(gtx, obj, item.objIndex)
				})
			})
		}),
	)
}

// --- State mutation helpers (must be called inside ga.do) ---

// addContainer appends a container to the local state.
func (ga *GioApp) addContainer(c Container) {
	ga.containers = append(ga.containers, c)
}

// updateContainer replaces a container in local state by ID.
func (ga *GioApp) updateContainer(updated Container) {
	for i, c := range ga.containers {
		if c.ID == updated.ID {
			ga.containers[i] = updated
			return
		}
	}
}

// removeContainer removes a container and its objects from local state.
func (ga *GioApp) removeContainer(containerID string) {
	for i, c := range ga.containers {
		if c.ID == containerID {
			ga.containers = append(ga.containers[:i], ga.containers[i+1:]...)
			break
		}
	}
	filtered := ga.objects[:0]
	for _, obj := range ga.objects {
		if obj.ContainerID != containerID {
			filtered = append(filtered, obj)
		}
	}
	ga.objects = filtered
}

// addObject appends an object to the flat list and the parent container's embedded list.
func (ga *GioApp) addObject(obj Object) {
	ga.objects = append(ga.objects, obj)
	if obj.ContainerID != "" {
		for i, c := range ga.containers {
			if c.ID == obj.ContainerID {
				ga.containers[i].Objects = append(ga.containers[i].Objects, obj)
				return
			}
		}
	}
}

// updateObject replaces an object in the flat list and keeps container embedded lists in sync.
func (ga *GioApp) updateObject(updated Object, oldContainerID string) {
	for i, obj := range ga.objects {
		if obj.ID == updated.ID {
			ga.objects[i] = updated
			break
		}
	}
	if oldContainerID != updated.ContainerID {
		ga.removeObjectFromContainer(updated.ID, oldContainerID)
		if updated.ContainerID != "" {
			for i, c := range ga.containers {
				if c.ID == updated.ContainerID {
					ga.containers[i].Objects = append(ga.containers[i].Objects, updated)
					break
				}
			}
		}
	} else if updated.ContainerID != "" {
		for i, c := range ga.containers {
			if c.ID == updated.ContainerID {
				for j, obj := range c.Objects {
					if obj.ID == updated.ID {
						ga.containers[i].Objects[j] = updated
						break
					}
				}
				break
			}
		}
	}
}

// removeObject removes an object from the flat list and its container's embedded list.
func (ga *GioApp) removeObject(objectID, containerID string) {
	for i, obj := range ga.objects {
		if obj.ID == objectID {
			ga.objects = append(ga.objects[:i], ga.objects[i+1:]...)
			break
		}
	}
	ga.removeObjectFromContainer(objectID, containerID)
}

// removeObjectFromContainer removes an object from a container's embedded Objects slice.
func (ga *GioApp) removeObjectFromContainer(objectID, containerID string) {
	if containerID == "" {
		return
	}
	for i, c := range ga.containers {
		if c.ID == containerID {
			for j, obj := range c.Objects {
				if obj.ID == objectID {
					ga.containers[i].Objects = append(c.Objects[:j], c.Objects[j+1:]...)
					return
				}
			}
			return
		}
	}
}

// fetchContainersAndObjects launches a goroutine that fetches containers and objects
// for the current collection, then updates state via ga.do(). Safe to call from anywhere.
func (ga *GioApp) fetchContainersAndObjects() {
	if ga.selectedCollection == nil || ga.currentUser == nil {
		return
	}
	collectionID := ga.selectedCollection.ID
	userID := ga.currentUser.ID

	go func() {
		ga.logger.Info("Fetching containers and objects", "collection_id", collectionID)

		containers, err := ga.containersClient.List(userID, collectionID)
		if err != nil {
			ga.logger.Error("Failed to fetch containers", "error", err)
			return
		}

		objects, err := ga.objectsClient.ListByCollection(userID, collectionID)
		if err != nil {
			ga.logger.Error("Failed to fetch objects", "error", err)
			return
		}

		ga.do(func() {
			ga.containers = containers
			ga.objects = objects
			ga.activeGroupedTextFilters = nil
		})
	}()
}

// renderObjectProperties renders the key/value properties of an object using schema-aware formatting.
func (ga *GioApp) renderObjectProperties(gtx layout.Context, props map[string]interface{}, defs []PropertyDefinition) layout.Dimensions {
	// Build ordered keys: schema-defined first (in order), then remaining alpha-sorted
	var keys []string
	seen := map[string]bool{}
	for _, def := range defs {
		if _, ok := props[def.Key]; ok {
			keys = append(keys, def.Key)
			seen[def.Key] = true
		}
	}
	var unseen []string
	for k := range props {
		if !seen[k] {
			unseen = append(unseen, k)
		}
	}
	sort.Strings(unseen)
	keys = append(keys, unseen...)

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		func() []layout.FlexChild {
			children := make([]layout.FlexChild, 0, len(keys))
			for _, k := range keys {
				k := k
				v := props[k]
				displayKey := propertyDisplayName(k, defs)
				displayVal := RenderPropertyValue(k, v, defs)
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := material.Body2(ga.theme.Theme, displayKey+": "+displayVal)
					label.Color = theme.ColorTextSecondary
					return label.Layout(gtx)
				}))
			}
			return children
		}()...,
	)
}

// propertyDisplayName returns the display name for a property key.
// Uses the schema definition's DisplayName when available, otherwise converts snake_case to Title Case.
func propertyDisplayName(key string, defs []PropertyDefinition) string {
	for _, def := range defs {
		if def.Key == key && def.DisplayName != "" {
			return def.DisplayName
		}
	}
	// snake_case → Title Case
	parts := strings.Split(key, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}

// collectGroupedTextValues returns unique values per grouped_text property key.
// Returns nil if the collection has no grouped_text properties.
func (ga *GioApp) collectGroupedTextValues() map[string][]string {
	if ga.selectedCollection == nil || ga.selectedCollection.PropertySchema == nil {
		return nil
	}
	result := map[string][]string{}
	seen := map[string]map[string]bool{}
	for _, def := range ga.selectedCollection.PropertySchema.Definitions {
		if def.Type == "grouped_text" {
			result[def.Key] = nil
			seen[def.Key] = map[string]bool{}
		}
	}
	if len(result) == 0 {
		return nil
	}
	for _, obj := range ga.objects {
		for k := range result {
			if v, ok := obj.Properties[k]; ok {
				s := fmt.Sprintf("%v", v)
				if s != "" && !seen[k][s] {
					seen[k][s] = true
					result[k] = append(result[k], s)
				}
			}
		}
	}
	for k := range result {
		sort.Strings(result[k])
	}
	return result
}

// matchesGroupedTextFilters returns true if the object passes all active grouped-text filters.
func (ga *GioApp) matchesGroupedTextFilters(obj Object) bool {
	for propKey, selectedVal := range ga.activeGroupedTextFilters {
		if selectedVal == "" {
			continue
		}
		v, ok := obj.Properties[propKey]
		if !ok {
			return false
		}
		if fmt.Sprintf("%v", v) != selectedVal {
			return false
		}
	}
	return true
}

// getGroupedTextChipButton lazily creates a widget.Clickable for the given chip key.
func (ga *GioApp) getGroupedTextChipButton(key string) *widget.Clickable {
	if btn, ok := ga.widgetState.groupedTextFilterButtons[key]; ok {
		return btn
	}
	btn := new(widget.Clickable)
	ga.widgetState.groupedTextFilterButtons[key] = btn
	return btn
}

// renderFilterChip renders a single filter chip button, highlighted when active.
func (ga *GioApp) renderFilterChip(gtx layout.Context, btn *widget.Clickable, label string, active bool) layout.Dimensions {
	bg := theme.ColorSurfaceAlt
	textColor := theme.ColorTextPrimary
	if active {
		bg = theme.ColorPrimary
		textColor = theme.ColorWhite
	}
	b := widgets.Button{
		Text:            label,
		BackgroundColor: bg,
		TextColor:       textColor,
		CornerRadius:    unit.Dp(theme.RadiusFull),
		Inset: layout.Inset{
			Top:    unit.Dp(theme.Spacing1),
			Bottom: unit.Dp(theme.Spacing1),
			Left:   unit.Dp(theme.Spacing2),
			Right:  unit.Dp(theme.Spacing2),
		},
	}
	return b.Layout(gtx, ga.theme.Theme, btn)
}

// renderGroupedTextFilters renders a filter bar of chips for each grouped_text property.
func (ga *GioApp) renderGroupedTextFilters(gtx layout.Context) layout.Dimensions {
	values := ga.collectGroupedTextValues()
	if len(values) == 0 {
		return layout.Dimensions{}
	}
	if ga.activeGroupedTextFilters == nil {
		ga.activeGroupedTextFilters = map[string]string{}
	}

	// Stable order: sort property keys
	var keys []string
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var defs []PropertyDefinition
	if ga.selectedCollection.PropertySchema != nil {
		defs = ga.selectedCollection.PropertySchema.Definitions
	}

	rows := make([]layout.FlexChild, 0, len(keys))
	for _, propKey := range keys {
		propKey := propKey
		vals := values[propKey]
		if len(vals) == 0 {
			continue
		}
		displayName := propertyDisplayName(propKey, defs)
		activeVal := ga.activeGroupedTextFilters[propKey]

		rows = append(rows, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				// Label
				var children []layout.FlexChild
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Right: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						label := material.Body2(ga.theme.Theme, displayName+":")
						label.Color = theme.ColorTextSecondary
						return label.Layout(gtx)
					})
				}))

				// "All" chip
				allKey := propKey + "||"
				allBtn := ga.getGroupedTextChipButton(allKey)
				if allBtn.Clicked(gtx) {
					ga.activeGroupedTextFilters[propKey] = ""
				}
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Right: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return ga.renderFilterChip(gtx, allBtn, "All", activeVal == "")
					})
				}))

				// Value chips
				for _, val := range vals {
					val := val
					chipKey := propKey + "||" + val
					btn := ga.getGroupedTextChipButton(chipKey)
					if btn.Clicked(gtx) {
						ga.activeGroupedTextFilters[propKey] = val
					}
					isActive := activeVal == val
					children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Right: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return ga.renderFilterChip(gtx, btn, val, isActive)
						})
					}))
				}

				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, children...)
			})
		}))
	}

	return layout.Inset{Bottom: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx, rows...)
	})
}
