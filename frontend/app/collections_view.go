package app

import (
	"fmt"
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/nishiki/frontend/pkg/types"
	"github.com/nishiki/frontend/ui/theme"
	"github.com/nishiki/frontend/ui/widgets"
)

// Object type constants
const (
	ObjectTypeFood      = "food"
	ObjectTypeBook      = "book"
	ObjectTypeVideoGame = "videogame"
	ObjectTypeMusic     = "music"
	ObjectTypeBoardGame = "boardgame"
	ObjectTypeGeneral   = "general"
)

var objectTypes = []string{
	ObjectTypeFood,
	ObjectTypeBook,
	ObjectTypeVideoGame,
	ObjectTypeMusic,
	ObjectTypeBoardGame,
	ObjectTypeGeneral,
}

var objectTypeLabels = map[string]string{
	ObjectTypeFood:      "Food",
	ObjectTypeBook:      "Books",
	ObjectTypeVideoGame: "Video Games",
	ObjectTypeMusic:     "Music",
	ObjectTypeBoardGame: "Board Games",
	ObjectTypeGeneral:   "General",
}

// renderCollectionsView renders the collections view with CRUD operations
func (ga *GioApp) renderCollectionsView(gtx layout.Context) layout.Dimensions {
	// Handle create button click
	if ga.widgetState.collectionsCreateButton.Clicked(gtx) {
		ga.logger.Info("Opening create collection dialog")
		ga.showCollectionDialog = true
		ga.collectionDialogMode = "create"
		ga.selectedObjectType = ObjectTypeGeneral
		ga.selectedGroupID = nil
		// Clear editors
		ga.widgetState.collectionNameEditor.SetText("")
		ga.widgetState.collectionLocationEditor.SetText("")
		ga.widgetState.collectionTagsEditor.SetText("")
	}

	// Ensure we have collection item states
	ga.ensureCollectionItemStates()

	// Main layout
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		// Header
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ga.renderHeader(gtx, "Collections")
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
					// Toolbar (search + create button)
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return ga.renderCollectionsToolbar(gtx)
					}),

					// Collections list
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return ga.renderCollectionsList(gtx)
					}),
				)
			})
		}),

		// Bottom navigation menu
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ga.renderBottomMenu(gtx, ViewCollectionsGio)
		}),
	)
}

// ensureCollectionItemStates ensures we have widget states for all collections
func (ga *GioApp) ensureCollectionItemStates() {
	if len(ga.widgetState.collectionItems) != len(ga.collections) {
		ga.widgetState.collectionItems = make([]CollectionItemState, len(ga.collections))
	}
}

// renderCollectionsToolbar renders the toolbar with search and create button
func (ga *GioApp) renderCollectionsToolbar(gtx layout.Context) layout.Dimensions {
	return layout.Inset{Bottom: unit.Dp(theme.Spacing4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:    layout.Horizontal,
			Spacing: layout.SpaceBetween,
		}.Layout(gtx,
			// Search field
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Right: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					editor := material.Editor(ga.theme.Theme, &ga.widgetState.collectionsSearchField, "Search collections...")
					editor.Color = theme.ColorTextPrimary
					return editor.Layout(gtx)
				})
			}),

			// Create button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return widgets.PrimaryButton(ga.theme.Theme, &ga.widgetState.collectionsCreateButton, "+")(gtx)
			}),
		)
	})
}

// renderCollectionsList renders the list of collections
func (ga *GioApp) renderCollectionsList(gtx layout.Context) layout.Dimensions {
	if len(ga.collections) == 0 {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			label := material.H5(ga.theme.Theme, "No collections yet")
			label.Color = theme.ColorTextSecondary
			label.Alignment = text.Middle
			return label.Layout(gtx)
		})
	}

	// Filter collections based on search query
	searchQuery := strings.ToLower(ga.widgetState.collectionsSearchField.Text())
	filteredCollections := make([]Collection, 0)
	filteredIndices := make([]int, 0)

	for i, collection := range ga.collections {
		if searchQuery == "" ||
			strings.Contains(strings.ToLower(collection.Name), searchQuery) ||
			strings.Contains(strings.ToLower(collection.Location), searchQuery) ||
			strings.Contains(strings.ToLower(collection.ObjectType), searchQuery) {
			filteredCollections = append(filteredCollections, collection)
			filteredIndices = append(filteredIndices, i)
		}
	}

	// Render list using widget state
	list := &ga.widgetState.collectionsList
	list.Axis = layout.Vertical
	return list.Layout(gtx, len(filteredCollections), func(gtx layout.Context, index int) layout.Dimensions {
		collection := filteredCollections[index]
		originalIndex := filteredIndices[index]
		return layout.Inset{Bottom: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return ga.renderCollectionCard(gtx, collection, originalIndex)
		})
	})
}

// renderCollectionCard renders a single collection card
func (ga *GioApp) renderCollectionCard(gtx layout.Context, collection Collection, index int) layout.Dimensions {
	itemState := &ga.widgetState.collectionItems[index]

	// Handle view button click
	if itemState.viewButton.Clicked(gtx) {
		ga.logger.Info("Viewing collection details", "collection_id", collection.ID)
		ga.selectedCollection = &collection
		ga.currentView = ViewCollectionDetailGio
		// Fetch containers and objects for this collection
		ga.fetchContainersAndObjects()
	}

	// Handle edit button click
	if itemState.editButton.Clicked(gtx) {
		ga.logger.Info("Opening edit collection dialog", "collection_id", collection.ID)
		ga.selectedCollection = &collection
		ga.showCollectionDialog = true
		ga.collectionDialogMode = "edit"
		ga.selectedObjectType = collection.ObjectType
		ga.selectedGroupID = collection.GroupID
		ga.widgetState.collectionNameEditor.SetText(collection.Name)
		ga.widgetState.collectionLocationEditor.SetText(collection.Location)
		ga.widgetState.collectionTagsEditor.SetText(strings.Join(collection.Tags, ", "))
	}

	// Handle delete button click
	if itemState.deleteButton.Clicked(gtx) {
		ga.logger.Info("Opening delete confirmation", "collection_id", collection.ID)
		ga.showDeleteCollection = true
		ga.deleteCollectionID = collection.ID
	}

	card := widgets.DefaultCard()
	return card.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			// Header row (name + type badge)
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
					Spacing:   layout.SpaceBetween,
				}.Layout(gtx,
					// Collection name
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						label := material.H6(ga.theme.Theme, collection.Name)
						label.Font.Weight = font.Bold
						return label.Layout(gtx)
					}),

					// Type label
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Left: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							label := material.Body2(ga.theme.Theme, objectTypeLabels[collection.ObjectType])
							label.Color = theme.ColorAccent
							return label.Layout(gtx)
						})
					}),
				)
			}),

			// Location
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if collection.Location != "" {
					return layout.Inset{Top: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						label := material.Body2(ga.theme.Theme, collection.Location)
						label.Color = theme.ColorTextSecondary
						return label.Layout(gtx)
					})
				}
				return layout.Dimensions{}
			}),

			// Tags
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if len(collection.Tags) > 0 {
					return layout.Inset{Top: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						tagsText := strings.Join(collection.Tags, ", ")
						label := material.Body2(ga.theme.Theme, "Tags: "+tagsText)
						label.Color = theme.ColorTextSecondary
						return label.Layout(gtx)
					})
				}
				return layout.Dimensions{}
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
								return widgets.PrimaryButton(ga.theme.Theme, &itemState.viewButton, "View")(gtx)
							})
						}),
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

// renderCollectionDialog renders the create/edit collection dialog
func (ga *GioApp) renderCollectionDialog(gtx layout.Context) layout.Dimensions {
	if !ga.showCollectionDialog {
		return layout.Dimensions{}
	}

	// Handle submit button
	if ga.widgetState.collectionDialogSubmit.Clicked(gtx) {
		if ga.collectionDialogMode == "create" {
			ga.handleCollectionCreate()
		} else {
			ga.handleCollectionUpdate()
		}
		ga.widgetState.collectionDialog.Reset()
		return layout.Dimensions{}
	}

	// Handle cancel button
	if ga.widgetState.collectionDialogCancel.Clicked(gtx) {
		ga.showCollectionDialog = false
		ga.selectedCollection = nil
		ga.widgetState.collectionDialog.Reset()
		return layout.Dimensions{}
	}

	// Determine title
	title := "Create Collection"
	if ga.collectionDialogMode == "edit" {
		title = "Edit Collection"
	}

	// Create dialog style
	dialogStyle := widgets.DefaultDialogStyle(ga.widgetState.collectionDialog, title)
	dialogStyle.Width = unit.Dp(600)

	// Render draggable dialog
	dims, dismissed := dialogStyle.Layout(gtx, ga.theme.Theme, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			// Name field
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ga.renderFormField(gtx, "Name *", &ga.widgetState.collectionNameEditor, "Enter collection name")
			}),

			// Object Type selection (only for create mode)
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if ga.collectionDialogMode == "create" {
					return ga.renderObjectTypeSelector(gtx)
				}
				return layout.Dimensions{}
			}),

			// Group selection
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ga.renderGroupSelector(gtx)
			}),

			// Location field
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ga.renderFormField(gtx, "Location", &ga.widgetState.collectionLocationEditor, "e.g., Kitchen, Living Room")
			}),

			// Tags field
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ga.renderFormField(gtx, "Tags", &ga.widgetState.collectionTagsEditor, "Comma-separated tags")
			}),

			// Buttons
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:    layout.Horizontal,
					Spacing: layout.SpaceEnd,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Right: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return widgets.CancelButton(ga.theme.Theme, &ga.widgetState.collectionDialogCancel, "Cancel")(gtx)
						})
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						submitText := "Create"
						if ga.collectionDialogMode == "edit" {
							submitText = "Update"
						}
						return widgets.PrimaryButton(ga.theme.Theme, &ga.widgetState.collectionDialogSubmit, submitText)(gtx)
					}),
				)
			}),
		)
	})

	// Handle backdrop dismissal
	if dismissed {
		ga.showCollectionDialog = false
		ga.selectedCollection = nil
		ga.widgetState.collectionDialog.Reset()
	}

	return dims
}

// renderFormField renders a labeled form field
func (ga *GioApp) renderFormField(gtx layout.Context, label string, editor *widget.Editor, hint string) layout.Dimensions {
	return layout.Inset{Bottom: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				labelWidget := material.Body2(ga.theme.Theme, label)
				labelWidget.Color = theme.ColorTextSecondary
				return labelWidget.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				editorWidget := material.Editor(ga.theme.Theme, editor, hint)
				return editorWidget.Layout(gtx)
			}),
		)
	})
}

// renderObjectTypeSelector renders object type selection chips.
func (ga *GioApp) renderObjectTypeSelector(gtx layout.Context) layout.Dimensions {
	chips := make([]layout.Widget, len(objectTypes))
	for i, ot := range objectTypes {
		ot := ot
		if ga.widgetState.collectionTypeButtons[ot] == nil {
			ga.widgetState.collectionTypeButtons[ot] = &widget.Clickable{}
		}
		btn := ga.widgetState.collectionTypeButtons[ot]
		if btn.Clicked(gtx) {
			ga.selectedObjectType = ot
		}
		active := ga.selectedObjectType == ot
		chips[i] = func(gtx layout.Context) layout.Dimensions {
			return ga.renderFilterChip(gtx, btn, objectTypeLabels[ot], active)
		}
	}
	return ga.renderChipSelector(gtx, "Object Type *", chips)
}

// renderGroupSelector renders group selection chips.
func (ga *GioApp) renderGroupSelector(gtx layout.Context) layout.Dimensions {
	noneBtn := ga.widgetState.collectionGroupButtons["none"]
	if noneBtn == nil {
		noneBtn = &widget.Clickable{}
		ga.widgetState.collectionGroupButtons["none"] = noneBtn
	}
	if noneBtn.Clicked(gtx) {
		ga.selectedGroupID = nil
	}
	chips := []layout.Widget{
		func(gtx layout.Context) layout.Dimensions {
			return ga.renderFilterChip(gtx, noneBtn, "None", ga.selectedGroupID == nil)
		},
	}
	for _, group := range ga.groups {
		grp := group
		btn := ga.widgetState.collectionGroupButtons[grp.ID]
		if btn == nil {
			btn = &widget.Clickable{}
			ga.widgetState.collectionGroupButtons[grp.ID] = btn
		}
		if btn.Clicked(gtx) {
			gid := grp.ID
			ga.selectedGroupID = &gid
		}
		active := ga.selectedGroupID != nil && *ga.selectedGroupID == grp.ID
		chips = append(chips, func(gtx layout.Context) layout.Dimensions {
			return ga.renderFilterChip(gtx, btn, grp.Name, active)
		})
	}
	return ga.renderChipSelector(gtx, "Group (Optional)", chips)
}

// handleCollectionCreate handles creating a new collection
func (ga *GioApp) handleCollectionCreate() {
	name := ga.widgetState.collectionNameEditor.Text()
	location := ga.widgetState.collectionLocationEditor.Text()
	tagsText := ga.widgetState.collectionTagsEditor.Text()

	if name == "" {
		ga.logger.Warn("Collection name is required")
		return
	}

	// Parse tags
	var tags []string
	if tagsText != "" {
		for _, tag := range strings.Split(tagsText, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	ga.logger.Info("Creating collection", "name", name, "type", ga.selectedObjectType)

	go func() {
		req := types.CreateCollectionRequest{
			Name:       name,
			ObjectType: ga.selectedObjectType,
			GroupID:    ga.selectedGroupID,
			Location:   location,
			Tags:       tags,
		}

		collection, err := ga.collectionsClient.Create(ga.currentUser.ID, req)
		if err != nil {
			ga.logger.Error("Failed to create collection", "error", err)
			return
		}

		ga.logger.Info("Collection created successfully", "collection_id", collection.ID)
		ga.do(func() {
			ga.collections = append(ga.collections, *collection)
		})
	}()

	// Close dialog
	ga.showCollectionDialog = false
	ga.window.Invalidate()
}

// handleCollectionUpdate handles updating an existing collection
func (ga *GioApp) handleCollectionUpdate() {
	if ga.selectedCollection == nil {
		ga.logger.Error("No collection selected for update")
		return
	}

	name := ga.widgetState.collectionNameEditor.Text()
	location := ga.widgetState.collectionLocationEditor.Text()
	tagsText := ga.widgetState.collectionTagsEditor.Text()

	if name == "" {
		ga.logger.Warn("Collection name is required")
		return
	}

	// Parse tags
	var tags []string
	if tagsText != "" {
		for _, tag := range strings.Split(tagsText, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	ga.logger.Info("Updating collection", "collection_id", ga.selectedCollection.ID, "name", name)

	collectionID := ga.selectedCollection.ID

	go func() {
		req := types.UpdateCollectionRequest{
			Name:     name,
			Location: location,
			Tags:     tags,
		}

		updated, err := ga.collectionsClient.Update(ga.currentUser.ID, collectionID, req)
		if err != nil {
			ga.logger.Error("Failed to update collection", "error", err)
			return
		}

		ga.logger.Info("Collection updated successfully", "collection_id", collectionID)
		ga.do(func() {
			for i, c := range ga.collections {
				if c.ID == updated.ID {
					ga.collections[i] = *updated
					break
				}
			}
		})
	}()

	// Close dialog
	ga.showCollectionDialog = false
	ga.selectedCollection = nil
	ga.window.Invalidate()
}

// renderDeleteCollectionDialog renders the delete confirmation dialog
func (ga *GioApp) renderDeleteCollectionDialog(gtx layout.Context) layout.Dimensions {
	if !ga.showDeleteCollection {
		return layout.Dimensions{}
	}

	// Find collection name for confirmation
	var collectionName string
	for _, collection := range ga.collections {
		if collection.ID == ga.deleteCollectionID {
			collectionName = collection.Name
			break
		}
	}

	// Handle confirm button
	if ga.widgetState.collectionDialogSubmit.Clicked(gtx) {
		ga.handleCollectionDelete()
		ga.widgetState.deleteDialog.Reset()
		return layout.Dimensions{}
	}

	// Handle cancel button
	if ga.widgetState.collectionDialogCancel.Clicked(gtx) {
		ga.showDeleteCollection = false
		ga.deleteCollectionID = ""
		ga.widgetState.deleteDialog.Reset()
		return layout.Dimensions{}
	}

	// Create dialog style
	dialogStyle := widgets.DefaultDialogStyle(ga.widgetState.deleteDialog, "Delete Collection")
	dialogStyle.Width = unit.Dp(500)
	dialogStyle.TitleBarColor = theme.ColorDanger

	// Render draggable dialog
	dims, dismissed := dialogStyle.Layout(gtx, ga.theme.Theme, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			// Message
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Bottom: unit.Dp(theme.Spacing4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					message := fmt.Sprintf("Are you sure you want to delete the collection \"%s\"? This will also delete all containers and objects within it. This action cannot be undone.", collectionName)
					label := material.Body1(ga.theme.Theme, message)
					return label.Layout(gtx)
				})
			}),

			// Buttons
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:    layout.Horizontal,
					Spacing: layout.SpaceEnd,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Right: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return widgets.CancelButton(ga.theme.Theme, &ga.widgetState.collectionDialogCancel, "Cancel")(gtx)
						})
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return widgets.DangerButton(ga.theme.Theme, &ga.widgetState.collectionDialogSubmit, "Delete")(gtx)
					}),
				)
			}),
		)
	})

	// Handle backdrop dismissal
	if dismissed {
		ga.showDeleteCollection = false
		ga.deleteCollectionID = ""
		ga.widgetState.deleteDialog.Reset()
	}

	return dims
}

// handleCollectionDelete handles deleting a collection
func (ga *GioApp) handleCollectionDelete() {
	if ga.deleteCollectionID == "" {
		ga.logger.Error("No collection ID for deletion")
		return
	}

	ga.logger.Info("Deleting collection", "collection_id", ga.deleteCollectionID)

	collectionID := ga.deleteCollectionID

	go func() {
		err := ga.collectionsClient.Delete(ga.currentUser.ID, collectionID, true)
		if err != nil {
			ga.logger.Error("Failed to delete collection", "error", err)
			ga.do(func() {
				ga.showDeleteCollectionError = true
				ga.deleteCollectionErrorMsg = err.Error()
			})
			return
		}

		ga.logger.Info("Collection deleted successfully", "collection_id", collectionID)
		ga.do(func() {
			for i, c := range ga.collections {
				if c.ID == collectionID {
					ga.collections = append(ga.collections[:i], ga.collections[i+1:]...)
					break
				}
			}
		})
	}()

	// Close dialog
	ga.showDeleteCollection = false
	ga.deleteCollectionID = ""
	ga.window.Invalidate()
}

// renderDeleteCollectionErrorDialog renders an error dialog when collection deletion fails
func (ga *GioApp) renderDeleteCollectionErrorDialog(gtx layout.Context) layout.Dimensions {
	if !ga.showDeleteCollectionError {
		return layout.Dimensions{}
	}

	// Handle dismiss button
	if ga.widgetState.collectionErrorDialogDismiss.Clicked(gtx) {
		ga.showDeleteCollectionError = false
		ga.deleteCollectionErrorMsg = ""
		ga.widgetState.collectionErrorDialog.Reset()
		return layout.Dimensions{}
	}

	dialogStyle := widgets.DefaultDialogStyle(ga.widgetState.collectionErrorDialog, "Delete Failed")
	dialogStyle.Width = unit.Dp(500)
	dialogStyle.TitleBarColor = theme.ColorDanger

	dims, dismissed := dialogStyle.Layout(gtx, ga.theme.Theme, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Bottom: unit.Dp(theme.Spacing4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					label := material.Body1(ga.theme.Theme, ga.deleteCollectionErrorMsg)
					return label.Layout(gtx)
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:    layout.Horizontal,
					Spacing: layout.SpaceEnd,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return widgets.PrimaryButton(ga.theme.Theme, &ga.widgetState.collectionErrorDialogDismiss, "OK")(gtx)
					}),
				)
			}),
		)
	})

	if dismissed {
		ga.showDeleteCollectionError = false
		ga.deleteCollectionErrorMsg = ""
		ga.widgetState.collectionErrorDialog.Reset()
	}

	return dims
}
