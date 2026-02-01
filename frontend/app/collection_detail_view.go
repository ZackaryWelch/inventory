//go:build js && wasm

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
	}

	// Handle create container button
	if ga.widgetState.createContainerButton.Clicked(gtx) {
		ga.logger.Info("Opening create container dialog")
		ga.showContainerDialog = true
		ga.containerDialogMode = "create"
		ga.selectedContainerID = nil
		ga.widgetState.containerNameEditor.SetText("")
		ga.widgetState.containerLocationEditor.SetText("")
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
				// Split view: Containers on left, Objects on right
				return layout.Flex{
					Axis:    layout.Horizontal,
					Spacing: layout.SpaceBetween,
				}.Layout(gtx,
					// Containers column
					layout.Flexed(0.4, func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Right: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return ga.renderContainersColumn(gtx)
						})
					}),

					// Objects column
					layout.Flexed(0.6, func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Left: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return ga.renderObjectsColumn(gtx)
						})
					}),
				)
			})
		}),

				// Bottom navigation menu
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return ga.renderBottomMenu(gtx, "collections")
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

			// Import button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Left: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					btn := material.Button(ga.theme.Theme, &ga.widgetState.importButton, "📥 Import")
					btn.Background = theme.ColorAccent
					btn.Color = theme.ColorBlack
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

	// Filter containers based on search query
	searchQuery := strings.ToLower(ga.widgetState.containersSearchField.Text())
	filteredContainers := make([]Container, 0)
	filteredIndices := make([]int, 0)

	for i, container := range ga.containers {
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
		if searchQuery == "" ||
			strings.Contains(strings.ToLower(object.Name), searchQuery) ||
			strings.Contains(strings.ToLower(object.Description), searchQuery) {
			filteredObjects = append(filteredObjects, object)
			filteredIndices = append(filteredIndices, i)
		}
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
							label := material.Caption(ga.theme.Theme, containerTypeLabels[container.Type])
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
						label := material.Caption(ga.theme.Theme, "📍 "+container.Location)
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
					label := material.Caption(ga.theme.Theme, fmt.Sprintf("📦 %d objects", objectCount))
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
						label := material.Caption(ga.theme.Theme, object.Description)
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
						label := material.Caption(ga.theme.Theme, fmt.Sprintf("Qty: %v %s", *object.Quantity, object.Unit))
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
						label := material.Caption(ga.theme.Theme, "🏷️  "+tagsText)
						label.Color = theme.ColorTextSecondary
						return label.Layout(gtx)
					})
				}
				return layout.Dimensions{}
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

// fetchContainersAndObjects fetches containers and objects for the current collection
func (ga *GioApp) fetchContainersAndObjects() {
	if ga.selectedCollection == nil {
		return
	}

	ga.logger.Info("Fetching containers and objects", "collection_id", ga.selectedCollection.ID)

	// Fetch containers
	containers, err := ga.containersClient.List(ga.currentUser.ID, ga.selectedCollection.ID)
	if err != nil {
		ga.logger.Error("Failed to fetch containers", "error", err)
		return
	}

	// Fetch objects
	objects, err := ga.objectsClient.ListByCollection(ga.currentUser.ID, ga.selectedCollection.ID)
	if err != nil {
		ga.logger.Error("Failed to fetch objects", "error", err)
		return
	}

	ga.containers = containers
	ga.objects = objects
	ga.window.Invalidate()
}
