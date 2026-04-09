package app

import (
	"fmt"
	"image"
	"sort"
	"strings"
	"sync"
	"time"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
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
	renderStart := time.Now()
	defer func() {
		elapsed := time.Since(renderStart)
		if elapsed > 5*time.Millisecond {
			ga.logger.Info("renderCollectionDetailView slow frame",
				"elapsed", elapsed,
				"objects", len(ga.objects),
				"containers", len(ga.containers))
		}
	}()

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
		ga.objectSortField = ""
		ga.objectSortDir = ""
		ga.objectGroupByField = ""
		ga.invalidateObjectCaches()
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
		// Clear schema property editors
		for _, ed := range ga.widgetState.objectPropertyEditors {
			ed.SetText("")
		}
		for _, b := range ga.widgetState.objectPropertyBools {
			b.Value = false
		}
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
						if ga.showContainersPanel && ga.containerViewMode == "grouped" && ga.objectGroupByField == "" {
							// Grouped mode: objects grouped under container headers (legacy, when no explicit group-by selected)
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

		// Sort/Group controls
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ga.renderSortGroupControls(gtx)
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
			if ga.objectGroupByField != "" {
				return ga.renderObjectsGroupedByField(gtx)
			}
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

// getFilteredContainers returns the cached filtered containers list, recomputing only when inputs change.
func (ga *GioApp) getFilteredContainers() ([]Container, []int) {
	searchQuery := strings.ToLower(ga.widgetState.containersSearchField.Text())
	if ga.cachedFilteredContainers != nil &&
		ga.cachedContSearchQuery == searchQuery &&
		ga.cachedContDataLen == len(ga.containers) {
		return ga.cachedFilteredContainers, ga.cachedFilteredContIndices
	}

	filtered := make([]Container, 0, len(ga.containers))
	indices := make([]int, 0, len(ga.containers))
	for i, container := range ga.containers {
		if searchQuery == "" ||
			strings.Contains(strings.ToLower(container.Name), searchQuery) ||
			strings.Contains(strings.ToLower(container.Type), searchQuery) ||
			strings.Contains(strings.ToLower(container.Location), searchQuery) {
			filtered = append(filtered, container)
			indices = append(indices, i)
		}
	}

	ga.cachedFilteredContainers = filtered
	ga.cachedFilteredContIndices = indices
	ga.cachedContSearchQuery = searchQuery
	ga.cachedContDataLen = len(ga.containers)
	return filtered, indices
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

	filteredContainers, filteredIndices := ga.getFilteredContainers()

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

// getFilteredObjects returns the cached filtered objects list, recomputing only when inputs change.
func (ga *GioApp) getFilteredObjects() ([]Object, []int) {
	searchQuery := strings.ToLower(ga.widgetState.objectsSearchField.Text())
	if ga.cachedFilteredObjects != nil &&
		ga.cachedObjSearchQuery == searchQuery &&
		ga.cachedObjDataLen == len(ga.objects) &&
		ga.cachedObjSortField == ga.objectSortField &&
		ga.cachedObjSortDir == ga.objectSortDir &&
		ga.cachedObjGroupField == ga.objectGroupByField &&
		mapsEqual(ga.cachedObjFilters, ga.activeGroupedTextFilters) {
		return ga.cachedFilteredObjects, ga.cachedFilteredObjIndices
	}

	start := time.Now()
	filtered := make([]Object, 0, len(ga.objects))
	indices := make([]int, 0, len(ga.objects))
	for i, object := range ga.objects {
		if searchQuery != "" &&
			!strings.Contains(strings.ToLower(object.Name), searchQuery) &&
			!strings.Contains(strings.ToLower(object.Description), searchQuery) {
			continue
		}
		if !ga.matchesGroupedTextFilters(object) {
			continue
		}
		filtered = append(filtered, object)
		indices = append(indices, i)
	}

	// Sort if a sort field is selected
	if ga.objectSortField != "" {
		ga.sortObjects(filtered, indices)
	}

	ga.cachedFilteredObjects = filtered
	ga.cachedFilteredObjIndices = indices
	ga.cachedObjSearchQuery = searchQuery
	ga.cachedObjDataLen = len(ga.objects)
	ga.cachedObjSortField = ga.objectSortField
	ga.cachedObjSortDir = ga.objectSortDir
	ga.cachedObjGroupField = ga.objectGroupByField
	ga.cachedObjFilters = copyStringMap(ga.activeGroupedTextFilters)
	ga.logger.Info("Recomputed filtered objects", "total", len(ga.objects), "filtered", len(filtered), "elapsed", time.Since(start))
	return filtered, indices
}

// sortObjects sorts the filtered/indices slices in place according to objectSortField/objectSortDir.
func (ga *GioApp) sortObjects(filtered []Object, indices []int) {
	desc := ga.objectSortDir == "desc"
	defMap := ga.getPropertyDefMap()

	sort.Sort(&objectSorter{
		objects:   filtered,
		indices:   indices,
		field:     ga.objectSortField,
		desc:      desc,
		defMap:    defMap,
		locFn:     ga.getObjectEffectiveLocation,
	})
}

// objectSorter implements sort.Interface for paired object + index slices.
type objectSorter struct {
	objects []Object
	indices []int
	field   string
	desc    bool
	defMap  map[string]*PropertyDefinition
	locFn   func(Object) string
}

func (s *objectSorter) Len() int { return len(s.objects) }

func (s *objectSorter) Swap(i, j int) {
	s.objects[i], s.objects[j] = s.objects[j], s.objects[i]
	s.indices[i], s.indices[j] = s.indices[j], s.indices[i]
}

func (s *objectSorter) Less(i, j int) bool {
	a := s.sortKey(s.objects[i])
	b := s.sortKey(s.objects[j])
	if s.desc {
		return strings.ToLower(a) > strings.ToLower(b)
	}
	return strings.ToLower(a) < strings.ToLower(b)
}

func (s *objectSorter) sortKey(obj Object) string {
	switch s.field {
	case "name":
		return obj.Name
	case "location":
		return s.locFn(obj)
	default:
		// Schema property key
		if tv, ok := obj.Properties[s.field]; ok {
			return RenderPropertyValueFromMap(s.field, tv, s.defMap)
		}
		return ""
	}
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

	filteredObjects, filteredIndices := ga.getFilteredObjects()

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
			// Header row (name + type label)
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
					Spacing:   layout.SpaceBetween,
				}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						label := material.H6(ga.theme.Theme, container.Name)
						label.Font.Weight = font.Bold
						return label.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Left: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							label := material.Body2(ga.theme.Theme, containerTypeLabels[container.Type])
							label.Color = theme.ColorAccent
							return label.Layout(gtx)
						})
					}),
				)
			}),

			// Location
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if container.Location != "" {
					return layout.Inset{Top: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						label := material.Body2(ga.theme.Theme, container.Location)
						label.Color = theme.ColorTextSecondary
						return label.Layout(gtx)
					})
				}
				return layout.Dimensions{}
			}),

			// Object count
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				objectCount := container.ObjectCount
				return layout.Inset{Top: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					label := material.Body2(ga.theme.Theme, fmt.Sprintf("%d objects", objectCount))
					label.Color = theme.ColorTextSecondary
					return label.Layout(gtx)
				})
			}),

			// Action buttons
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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
		// Populate schema property editors from existing object properties
		if ga.selectedCollection != nil && ga.selectedCollection.PropertySchema != nil {
			for _, def := range ga.selectedCollection.PropertySchema.Definitions {
				if def.Type == "bool" {
					b := ga.getObjectPropertyBool(def.Key)
					b.Value = false
					if tv, ok := object.Properties[def.Key]; ok {
						switch v := tv.Val.(type) {
						case bool:
							b.Value = v
						case string:
							b.Value = strings.EqualFold(v, "true") || v == "1"
						}
					}
				} else {
					ed := ga.getObjectPropertyEditor(def.Key)
					if tv, ok := object.Properties[def.Key]; ok {
						ed.SetText(fmt.Sprintf("%v", tv.Val))
					} else {
						ed.SetText("")
					}
				}
			}
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

			// Location
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				loc := ga.getObjectEffectiveLocation(object)
				if loc != "" {
					return layout.Inset{Top: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						label := material.Body2(ga.theme.Theme, loc)
						label.Color = theme.ColorPrimaryLight
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

	// Build container → objects map using cached filtered list
	_, filteredIndices := ga.getFilteredObjects()
	containerObjects := make(map[string][]int) // containerID → indices into ga.objects
	var unassigned []int

	for _, i := range filteredIndices {
		obj := ga.objects[i]
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

// getObjectEffectiveLocation returns the resolved location for an object.
// Priority: container location (if object is in a container with a location) → collection location.
func (ga *GioApp) getObjectEffectiveLocation(obj Object) string {
	if obj.ContainerID != "" {
		for _, c := range ga.containers {
			if c.ID == obj.ContainerID && c.Location != "" {
				return c.Location
			}
		}
	}
	if ga.selectedCollection != nil {
		return ga.selectedCollection.Location
	}
	return ""
}

// getObjectGroupKey returns the grouping key for an object given a group-by field.
func (ga *GioApp) getObjectGroupKey(obj Object, field string) string {
	switch field {
	case "location":
		loc := ga.getObjectEffectiveLocation(obj)
		if loc == "" {
			return "No Location"
		}
		return loc
	case "container":
		if obj.ContainerID == "" {
			return "Unassigned"
		}
		for _, c := range ga.containers {
			if c.ID == obj.ContainerID {
				return c.Name
			}
		}
		return "Unknown"
	default:
		// Schema property key
		if tv, ok := obj.Properties[field]; ok {
			defMap := ga.getPropertyDefMap()
			val := RenderPropertyValueFromMap(field, tv, defMap)
			if val != "" {
				return val
			}
		}
		return "None"
	}
}

// renderObjectsGroupedByField renders objects grouped by the selected objectGroupByField.
func (ga *GioApp) renderObjectsGroupedByField(gtx layout.Context) layout.Dimensions {
	if len(ga.objects) == 0 {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			label := material.Body1(ga.theme.Theme, "No objects yet")
			label.Color = theme.ColorTextSecondary
			label.Alignment = text.Middle
			return label.Layout(gtx)
		})
	}

	filteredObjects, filteredIndices := ga.getFilteredObjects()

	// Build groups preserving order of first occurrence
	type groupEntry struct {
		key     string
		indices []int // indices into ga.objects
	}
	var groupOrder []string
	groupMap := map[string]*groupEntry{}

	for i, obj := range filteredObjects {
		originalIdx := filteredIndices[i]
		key := ga.getObjectGroupKey(obj, ga.objectGroupByField)
		if g, ok := groupMap[key]; ok {
			g.indices = append(g.indices, originalIdx)
		} else {
			groupMap[key] = &groupEntry{key: key, indices: []int{originalIdx}}
			groupOrder = append(groupOrder, key)
		}
	}

	// Sort group keys alphabetically for stable output
	sort.Strings(groupOrder)

	// Flatten into header + object items
	type listItem struct {
		isHeader bool
		header   string
		objIndex int
	}
	var items []listItem
	for _, key := range groupOrder {
		g := groupMap[key]
		items = append(items, listItem{isHeader: true, header: fmt.Sprintf("%s (%d)", g.key, len(g.indices))})
		for _, idx := range g.indices {
			items = append(items, listItem{objIndex: idx})
		}
	}

	ga.ensureObjectItemStates()

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
}

// --- State mutation helpers (must be called inside ga.do) ---

// addContainer appends a container to the local state.
func (ga *GioApp) addContainer(c Container) {
	ga.containers = append(ga.containers, c)
	ga.invalidateObjectCaches()
}

// updateContainer replaces a container in local state by ID.
func (ga *GioApp) updateContainer(updated Container) {
	for i, c := range ga.containers {
		if c.ID == updated.ID {
			ga.containers[i] = updated
			ga.invalidateObjectCaches()
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
	ga.invalidateObjectCaches()
}

// addObject appends an object to the flat list and the parent container's embedded list.
func (ga *GioApp) addObject(obj Object) {
	ga.objects = append(ga.objects, obj)
	ga.invalidateObjectCaches()
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
	ga.invalidateObjectCaches()
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
	ga.invalidateObjectCaches()
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
		fetchStart := time.Now()
		ga.logger.Info("Fetching containers and objects", "collection_id", collectionID)

		var (
			containers []Container
			objects    []Object
			contErr    error
			objErr     error
			contTime   time.Duration
			objTime    time.Duration
		)

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			start := time.Now()
			containers, contErr = ga.containersClient.List(userID, collectionID)
			contTime = time.Since(start)
		}()
		go func() {
			defer wg.Done()
			start := time.Now()
			objects, objErr = ga.objectsClient.ListByCollection(userID, collectionID)
			objTime = time.Since(start)
		}()
		wg.Wait()

		ga.logger.Info("Fetch complete",
			"containers", len(containers), "containers_time", contTime,
			"objects", len(objects), "objects_time", objTime,
			"total_time", time.Since(fetchStart))

		if contErr != nil {
			ga.logger.Error("Failed to fetch containers", "error", contErr)
			return
		}
		if objErr != nil {
			ga.logger.Error("Failed to fetch objects", "error", objErr)
			return
		}

		ga.logger.Info("Queuing state update via ga.do()")
		doStart := time.Now()
		ga.do(func() {
			ga.logger.Info("ga.do() callback executing", "wait", time.Since(doStart))
			ga.containers = containers
			ga.objects = objects
			ga.activeGroupedTextFilters = nil
			ga.invalidateObjectCaches()
			ga.logger.Info("State updated", "objects", len(ga.objects), "containers", len(ga.containers))
		})
	}()
}

// renderObjectProperties renders the key/value properties of an object using schema-aware formatting.
func (ga *GioApp) renderObjectProperties(gtx layout.Context, props map[string]TypedValue, defs []PropertyDefinition) layout.Dimensions {
	defMap := ga.getPropertyDefMap()

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
				tv := props[k]
				displayKey := propertyDisplayNameFromMap(k, defMap)
				displayVal := RenderPropertyValueFromMap(k, tv, defMap)
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
	return snakeToTitleCase(key)
}

// propertyDisplayNameFromMap is like propertyDisplayName but uses a pre-built map for O(1) lookup.
func propertyDisplayNameFromMap(key string, defMap map[string]*PropertyDefinition) string {
	if def, ok := defMap[key]; ok && def.DisplayName != "" {
		return def.DisplayName
	}
	return snakeToTitleCase(key)
}

// snakeToTitleCase converts snake_case to Title Case.
func snakeToTitleCase(key string) string {
	parts := strings.Split(key, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}

// invalidateObjectCaches clears all caches that depend on ga.objects or ga.selectedCollection.
// Call this whenever objects are loaded, added, edited, or deleted, or when the collection changes.
func (ga *GioApp) invalidateObjectCaches() {
	ga.cachedGroupedTextValid = false
	ga.cachedGroupedTextValues = nil
	ga.cachedPropertyDefValid = false
	ga.cachedPropertyDefMap = nil
	ga.cachedFilteredObjects = nil
	ga.cachedFilteredObjIndices = nil
	ga.cachedFilteredContainers = nil
	ga.cachedFilteredContIndices = nil
}

// mapsEqual returns true if two string→string maps have identical entries.
func mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

// copyStringMap returns a shallow copy of a string→string map.
func copyStringMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	cp := make(map[string]string, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}

// getGroupedTextValues returns the cached grouped text values, recomputing only when invalidated.
func (ga *GioApp) getGroupedTextValues() map[string][]string {
	if ga.cachedGroupedTextValid {
		return ga.cachedGroupedTextValues
	}
	start := time.Now()
	ga.cachedGroupedTextValues = ga.collectGroupedTextValues()
	ga.cachedGroupedTextValid = true
	ga.logger.Info("Recomputed grouped text values", "objects", len(ga.objects), "elapsed", time.Since(start))
	return ga.cachedGroupedTextValues
}

// getPropertyDefMap returns a cached map from property key to definition.
func (ga *GioApp) getPropertyDefMap() map[string]*PropertyDefinition {
	if ga.cachedPropertyDefValid {
		return ga.cachedPropertyDefMap
	}
	ga.cachedPropertyDefMap = make(map[string]*PropertyDefinition)
	if ga.selectedCollection != nil && ga.selectedCollection.PropertySchema != nil {
		for i := range ga.selectedCollection.PropertySchema.Definitions {
			def := &ga.selectedCollection.PropertySchema.Definitions[i]
			ga.cachedPropertyDefMap[def.Key] = def
		}
	}
	ga.cachedPropertyDefValid = true
	return ga.cachedPropertyDefMap
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

// layoutFlowWrap lays out widgets in a horizontal flow that wraps to the next line.
func layoutFlowWrap(gtx layout.Context, hGap, vGap int, widgets ...layout.Widget) layout.Dimensions {
	maxWidth := gtx.Constraints.Max.X
	var x, y, rowHeight int

	type positioned struct {
		call op.CallOp
		pos  image.Point
		size image.Point
	}
	var items []positioned

	for _, w := range widgets {
		// Use min width 0 so widgets shrink-wrap their content instead of
		// expanding to fill the full available width.
		wgtx := gtx
		wgtx.Constraints.Min.X = 0
		macro := op.Record(gtx.Ops)
		dims := w(wgtx)
		call := macro.Stop()

		if x > 0 && x+dims.Size.X > maxWidth {
			y += rowHeight + vGap
			x = 0
			rowHeight = 0
		}

		items = append(items, positioned{
			call: call,
			pos:  image.Point{X: x, Y: y},
			size: dims.Size,
		})

		x += dims.Size.X + hGap
		if dims.Size.Y > rowHeight {
			rowHeight = dims.Size.Y
		}
	}

	totalHeight := y + rowHeight

	for _, item := range items {
		stack := op.Offset(item.pos).Push(gtx.Ops)
		item.call.Add(gtx.Ops)
		stack.Pop()
	}

	return layout.Dimensions{Size: image.Point{X: maxWidth, Y: totalHeight}}
}

// renderChipSelector renders a labeled field with flow-wrapping chip buttons.
// It is the chip equivalent of renderFormField.
func (ga *GioApp) renderChipSelector(gtx layout.Context, label string, chips []layout.Widget) layout.Dimensions {
	chipGap := gtx.Dp(unit.Dp(theme.Spacing1))
	return layout.Inset{Bottom: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body2(ga.theme.Theme, label)
				lbl.Color = theme.ColorTextSecondary
				return lbl.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layoutFlowWrap(gtx, chipGap, chipGap, chips...)
				})
			}),
		)
	})
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

// sortGroupField describes a field available for sorting/grouping.
type sortGroupField struct {
	key         string // internal key: "name", "location", or a property key
	displayName string // human label
}

// getSortableFields returns the list of fields available for sorting.
// Always includes Name and Location, plus each schema property.
func (ga *GioApp) getSortableFields() []sortGroupField {
	fields := []sortGroupField{
		{key: "name", displayName: "Name"},
		{key: "location", displayName: "Location"},
	}
	if ga.selectedCollection != nil && ga.selectedCollection.PropertySchema != nil {
		for _, def := range ga.selectedCollection.PropertySchema.Definitions {
			name := def.DisplayName
			if name == "" {
				name = snakeToTitleCase(def.Key)
			}
			fields = append(fields, sortGroupField{key: def.Key, displayName: name})
		}
	}
	return fields
}

// getGroupableFields returns the list of fields available for grouping.
// Includes Location, Container, and each schema property.
func (ga *GioApp) getGroupableFields() []sortGroupField {
	fields := []sortGroupField{
		{key: "location", displayName: "Location"},
		{key: "container", displayName: "Container"},
	}
	if ga.selectedCollection != nil && ga.selectedCollection.PropertySchema != nil {
		for _, def := range ga.selectedCollection.PropertySchema.Definitions {
			name := def.DisplayName
			if name == "" {
				name = snakeToTitleCase(def.Key)
			}
			fields = append(fields, sortGroupField{key: def.Key, displayName: name})
		}
	}
	return fields
}

// renderSortGroupControls renders sort and group-by chip selectors.
func (ga *GioApp) renderSortGroupControls(gtx layout.Context) layout.Dimensions {
	return layout.Inset{Bottom: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Sort row
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ga.renderSortRow(gtx)
			}),
			// Group row
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ga.renderGroupRow(gtx)
			}),
		)
	})
}

// renderSortRow renders the sort-by chip row.
func (ga *GioApp) renderSortRow(gtx layout.Context) layout.Dimensions {
	fields := ga.getSortableFields()

	// Process clicks
	for _, f := range fields {
		chipKey := "sort||" + f.key
		btn := ga.getGroupedTextChipButton(chipKey)
		if btn.Clicked(gtx) {
			if ga.objectSortField == f.key {
				// Toggle direction on re-click
				if ga.objectSortDir == "desc" {
					// Third click: clear sort
					ga.objectSortField = ""
					ga.objectSortDir = ""
				} else {
					ga.objectSortDir = "desc"
				}
			} else {
				ga.objectSortField = f.key
				ga.objectSortDir = "asc"
			}
			ga.invalidateFilteredObjects()
		}
	}

	chipGap := gtx.Dp(unit.Dp(theme.Spacing1))

	return layout.Inset{Bottom: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Start}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Right: unit.Dp(theme.Spacing2), Top: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					label := material.Body2(ga.theme.Theme, "Sort:")
					label.Color = theme.ColorTextSecondary
					return label.Layout(gtx)
				})
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				var chips []layout.Widget
				for _, f := range fields {
					f := f
					chipKey := "sort||" + f.key
					btn := ga.getGroupedTextChipButton(chipKey)
					isActive := ga.objectSortField == f.key
					label := f.displayName
					if isActive && ga.objectSortDir == "desc" {
						label += " ↓"
					} else if isActive {
						label += " ↑"
					}
					chips = append(chips, func(gtx layout.Context) layout.Dimensions {
						return ga.renderFilterChip(gtx, btn, label, isActive)
					})
				}
				return layoutFlowWrap(gtx, chipGap, chipGap, chips...)
			}),
		)
	})
}

// renderGroupRow renders the group-by chip row.
func (ga *GioApp) renderGroupRow(gtx layout.Context) layout.Dimensions {
	fields := ga.getGroupableFields()

	// Process clicks
	noneKey := "group||"
	noneBtn := ga.getGroupedTextChipButton(noneKey)
	if noneBtn.Clicked(gtx) {
		ga.objectGroupByField = ""
		ga.invalidateFilteredObjects()
	}
	for _, f := range fields {
		chipKey := "group||" + f.key
		btn := ga.getGroupedTextChipButton(chipKey)
		if btn.Clicked(gtx) {
			if ga.objectGroupByField == f.key {
				ga.objectGroupByField = ""
			} else {
				ga.objectGroupByField = f.key
			}
			ga.invalidateFilteredObjects()
		}
	}

	chipGap := gtx.Dp(unit.Dp(theme.Spacing1))

	return layout.Inset{Bottom: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Start}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Right: unit.Dp(theme.Spacing2), Top: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					label := material.Body2(ga.theme.Theme, "Group:")
					label.Color = theme.ColorTextSecondary
					return label.Layout(gtx)
				})
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				var chips []layout.Widget

				// "None" chip
				chips = append(chips, func(gtx layout.Context) layout.Dimensions {
					return ga.renderFilterChip(gtx, noneBtn, "None", ga.objectGroupByField == "")
				})

				for _, f := range fields {
					f := f
					chipKey := "group||" + f.key
					btn := ga.getGroupedTextChipButton(chipKey)
					isActive := ga.objectGroupByField == f.key
					chips = append(chips, func(gtx layout.Context) layout.Dimensions {
						return ga.renderFilterChip(gtx, btn, f.displayName, isActive)
					})
				}
				return layoutFlowWrap(gtx, chipGap, chipGap, chips...)
			}),
		)
	})
}

// invalidateFilteredObjects clears only the filtered objects cache (not the grouped text / property def caches).
func (ga *GioApp) invalidateFilteredObjects() {
	ga.cachedFilteredObjects = nil
	ga.cachedFilteredObjIndices = nil
}

// renderGroupedTextFilters renders a filter bar of chips for each grouped_text property.
func (ga *GioApp) renderGroupedTextFilters(gtx layout.Context) layout.Dimensions {
	values := ga.getGroupedTextValues()
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
				// Process all clicks before layout
				allKey := propKey + "||"
				allBtn := ga.getGroupedTextChipButton(allKey)
				if allBtn.Clicked(gtx) {
					ga.activeGroupedTextFilters[propKey] = ""
				}
				for _, val := range vals {
					chipKey := propKey + "||" + val
					btn := ga.getGroupedTextChipButton(chipKey)
					if btn.Clicked(gtx) {
						ga.activeGroupedTextFilters[propKey] = val
					}
				}

				chipGap := gtx.Dp(unit.Dp(theme.Spacing1))

				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Start}.Layout(gtx,
					// Label
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Right: unit.Dp(theme.Spacing2), Top: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							label := material.Body2(ga.theme.Theme, displayName+":")
							label.Color = theme.ColorTextSecondary
							return label.Layout(gtx)
						})
					}),
					// Flow-wrapping chips
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						var chips []layout.Widget

						// "All" chip
						chips = append(chips, func(gtx layout.Context) layout.Dimensions {
							return ga.renderFilterChip(gtx, allBtn, "All", activeVal == "")
						})

						// Value chips
						for _, val := range vals {
							val := val
							chipKey := propKey + "||" + val
							btn := ga.getGroupedTextChipButton(chipKey)
							isActive := activeVal == val
							chips = append(chips, func(gtx layout.Context) layout.Dimensions {
								return ga.renderFilterChip(gtx, btn, val, isActive)
							})
						}

						return layoutFlowWrap(gtx, chipGap, chipGap, chips...)
					}),
				)
			})
		}))
	}

	return layout.Inset{Bottom: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx, rows...)
	})
}
