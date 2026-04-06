package app

import (
	"fmt"
	"strconv"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/nishiki/frontend/pkg/types"
	"github.com/nishiki/frontend/ui/theme"
	"github.com/nishiki/frontend/ui/widgets"
)

// renderContainerDialog renders the create/edit container dialog
func (ga *GioApp) renderContainerDialog(gtx layout.Context) layout.Dimensions {
	if !ga.showContainerDialog {
		return layout.Dimensions{}
	}

	// Handle submit button
	if ga.widgetState.containerDialogSubmit.Clicked(gtx) {
		if ga.containerDialogMode == "create" {
			ga.handleContainerCreate()
		} else {
			ga.handleContainerUpdate()
		}
		ga.widgetState.containerDialog.Reset()
		return layout.Dimensions{}
	}

	// Handle cancel button
	if ga.widgetState.containerDialogCancel.Clicked(gtx) {
		ga.showContainerDialog = false
		ga.selectedContainer = nil
		ga.widgetState.containerDialog.Reset()
		return layout.Dimensions{}
	}

	// Determine title
	title := "Create Container"
	if ga.containerDialogMode == "edit" {
		title = "Edit Container"
	}

	// Create dialog style
	dialogStyle := widgets.DefaultDialogStyle(ga.widgetState.containerDialog, title)
	dialogStyle.Width = unit.Dp(500)

	// Render draggable dialog
	dims, dismissed := dialogStyle.Layout(gtx, ga.theme.Theme, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Name field
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ga.renderFormField(gtx, "Name *", &ga.widgetState.containerNameEditor, "Enter container name")
			}),

			// Container Type selection (only for create mode)
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if ga.containerDialogMode == "create" {
					return ga.renderContainerTypeSelector(gtx)
				}
				return layout.Dimensions{}
			}),

			// Parent container selection
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ga.renderParentContainerSelector(gtx)
			}),

			// Location field
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ga.renderFormField(gtx, "Location", &ga.widgetState.containerLocationEditor, "e.g., Living Room, Shelf 3")
			}),

			// Buttons
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:    layout.Horizontal,
					Spacing: layout.SpaceEnd,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Right: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return widgets.CancelButton(ga.theme.Theme, &ga.widgetState.containerDialogCancel, "Cancel")(gtx)
						})
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						submitText := "Create"
						if ga.containerDialogMode == "edit" {
							submitText = "Update"
						}
						return widgets.PrimaryButton(ga.theme.Theme, &ga.widgetState.containerDialogSubmit, submitText)(gtx)
					}),
				)
			}),
		)
	})

	// Handle backdrop dismissal
	if dismissed {
		ga.showContainerDialog = false
		ga.selectedContainer = nil
		ga.widgetState.containerDialog.Reset()
	}

	return dims
}

// renderObjectDialog renders the create/edit object dialog
func (ga *GioApp) renderObjectDialog(gtx layout.Context) layout.Dimensions {
	if !ga.showObjectDialog {
		return layout.Dimensions{}
	}

	// Handle submit button
	if ga.widgetState.objectDialogSubmit.Clicked(gtx) {
		if ga.objectDialogMode == "create" {
			ga.handleObjectCreate()
		} else {
			ga.handleObjectUpdate()
		}
		ga.widgetState.objectDialog.Reset()
		return layout.Dimensions{}
	}

	// Handle cancel button
	if ga.widgetState.objectDialogCancel.Clicked(gtx) {
		ga.showObjectDialog = false
		ga.selectedObject = nil
		ga.widgetState.objectDialog.Reset()
		return layout.Dimensions{}
	}

	// Determine title
	title := "Create Object"
	if ga.objectDialogMode == "edit" {
		title = "Edit Object"
	}

	// Create dialog style
	dialogStyle := widgets.DefaultDialogStyle(ga.widgetState.objectDialog, title)
	dialogStyle.Width = unit.Dp(500)

	// Render draggable dialog
	dims, dismissed := dialogStyle.Layout(gtx, ga.theme.Theme, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Name field
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ga.renderFormField(gtx, "Name *", &ga.widgetState.objectNameEditor, "Enter object name")
			}),

			// Container selection
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ga.renderObjectContainerSelector(gtx)
			}),

			// Description field
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ga.renderFormField(gtx, "Description", &ga.widgetState.objectDescriptionEditor, "Optional description")
			}),

			// Quantity and Unit row
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:    layout.Horizontal,
					Spacing: layout.SpaceBetween,
				}.Layout(gtx,
					layout.Flexed(0.5, func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Right: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return ga.renderFormField(gtx, "Quantity", &ga.widgetState.objectQuantityEditor, "e.g., 1.5")
						})
					}),
					layout.Flexed(0.5, func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Left: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return ga.renderFormField(gtx, "Unit", &ga.widgetState.objectUnitEditor, "e.g., kg, lbs")
						})
					}),
				)
			}),

			// Buttons
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:    layout.Horizontal,
					Spacing: layout.SpaceEnd,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Right: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return widgets.CancelButton(ga.theme.Theme, &ga.widgetState.objectDialogCancel, "Cancel")(gtx)
						})
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						submitText := "Create"
						if ga.objectDialogMode == "edit" {
							submitText = "Update"
						}
						return widgets.PrimaryButton(ga.theme.Theme, &ga.widgetState.objectDialogSubmit, submitText)(gtx)
					}),
				)
			}),
		)
	})

	// Handle backdrop dismissal
	if dismissed {
		ga.showObjectDialog = false
		ga.selectedObject = nil
		ga.widgetState.objectDialog.Reset()
	}

	return dims
}

// renderDeleteContainerDialog renders the delete container confirmation dialog
func (ga *GioApp) renderDeleteContainerDialog(gtx layout.Context) layout.Dimensions {
	if !ga.showDeleteContainer {
		return layout.Dimensions{}
	}

	// Find container name and count children
	var containerName string
	var childCount int
	for _, container := range ga.containers {
		if container.ID == ga.deleteContainerID {
			containerName = container.Name
		}
		if container.ParentContainerID != nil && *container.ParentContainerID == ga.deleteContainerID {
			childCount++
		}
	}

	// Handle confirm button
	if ga.widgetState.containerDialogSubmit.Clicked(gtx) {
		ga.handleContainerDelete()
		ga.widgetState.deleteDialog.Reset()
		return layout.Dimensions{}
	}

	// Handle cancel button
	if ga.widgetState.containerDialogCancel.Clicked(gtx) {
		ga.showDeleteContainer = false
		ga.deleteContainerID = ""
		ga.widgetState.deleteDialog.Reset()
		return layout.Dimensions{}
	}

	// Create dialog style
	dialogStyle := widgets.DefaultDialogStyle(ga.widgetState.deleteDialog, "Delete Container")
	dialogStyle.Width = unit.Dp(500)
	dialogStyle.TitleBarColor = theme.ColorDanger

	// Render draggable dialog
	dims, dismissed := dialogStyle.Layout(gtx, ga.theme.Theme, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Message
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Bottom: unit.Dp(theme.Spacing4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					var message string
					if childCount > 0 {
						message = fmt.Sprintf("Cannot delete container \"%s\" because it has %d child container(s). Remove or reassign all child containers first.", containerName, childCount)
					} else {
						message = fmt.Sprintf("Are you sure you want to delete the container \"%s\"? All objects within it will also be deleted. This action cannot be undone.", containerName)
					}
					label := material.Body1(ga.theme.Theme, message)
					if childCount > 0 {
						label.Color = theme.ColorDanger
					}
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
							cancelLabel := "Cancel"
							if childCount > 0 {
								cancelLabel = "Close"
							}
							return widgets.CancelButton(ga.theme.Theme, &ga.widgetState.containerDialogCancel, cancelLabel)(gtx)
						})
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if childCount > 0 {
							// Don't show delete button when there are children
							return layout.Dimensions{}
						}
						return widgets.DangerButton(ga.theme.Theme, &ga.widgetState.containerDialogSubmit, "Delete")(gtx)
					}),
				)
			}),
		)
	})

	// Handle backdrop dismissal
	if dismissed {
		ga.showDeleteContainer = false
		ga.deleteContainerID = ""
		ga.widgetState.deleteDialog.Reset()
	}

	return dims
}

// renderDeleteObjectDialog renders the delete object confirmation dialog
func (ga *GioApp) renderDeleteObjectDialog(gtx layout.Context) layout.Dimensions {
	if !ga.showDeleteObject {
		return layout.Dimensions{}
	}

	// Find object name for confirmation
	var objectName string
	for _, object := range ga.objects {
		if object.ID == ga.deleteObjectID {
			objectName = object.Name
			break
		}
	}

	// Handle confirm button
	if ga.widgetState.objectDialogSubmit.Clicked(gtx) {
		ga.handleObjectDelete()
		ga.widgetState.deleteDialog.Reset()
		return layout.Dimensions{}
	}

	// Handle cancel button
	if ga.widgetState.objectDialogCancel.Clicked(gtx) {
		ga.showDeleteObject = false
		ga.deleteObjectID = ""
		ga.widgetState.deleteDialog.Reset()
		return layout.Dimensions{}
	}

	// Create dialog style
	dialogStyle := widgets.DefaultDialogStyle(ga.widgetState.deleteDialog, "Delete Object")
	dialogStyle.Width = unit.Dp(500)
	dialogStyle.TitleBarColor = theme.ColorDanger

	// Render draggable dialog
	dims, dismissed := dialogStyle.Layout(gtx, ga.theme.Theme, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Message
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Bottom: unit.Dp(theme.Spacing4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					message := fmt.Sprintf("Are you sure you want to delete the object \"%s\"? This action cannot be undone.", objectName)
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
							return widgets.CancelButton(ga.theme.Theme, &ga.widgetState.objectDialogCancel, "Cancel")(gtx)
						})
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return widgets.DangerButton(ga.theme.Theme, &ga.widgetState.objectDialogSubmit, "Delete")(gtx)
					}),
				)
			}),
		)
	})

	// Handle backdrop dismissal
	if dismissed {
		ga.showDeleteObject = false
		ga.deleteObjectID = ""
		ga.widgetState.deleteDialog.Reset()
	}

	return dims
}

// renderContainerTypeSelector renders container type selection buttons
func (ga *GioApp) renderContainerTypeSelector(gtx layout.Context) layout.Dimensions {
	return layout.Inset{Bottom: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Body2(ga.theme.Theme, "Type *")
				label.Color = theme.ColorTextSecondary
				return label.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					// Create buttons for each type (2 rows of 3)
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{
								Axis:    layout.Horizontal,
								Spacing: layout.SpaceEvenly,
							}.Layout(gtx,
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									return ga.renderContainerTypeButton(gtx, ContainerTypeRoom)
								}),
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									return ga.renderContainerTypeButton(gtx, ContainerTypeBookshelf)
								}),
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									return ga.renderContainerTypeButton(gtx, ContainerTypeShelf)
								}),
							)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Top: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return layout.Flex{
									Axis:    layout.Horizontal,
									Spacing: layout.SpaceEvenly,
								}.Layout(gtx,
									layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
										return ga.renderContainerTypeButton(gtx, ContainerTypeBinder)
									}),
									layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
										return ga.renderContainerTypeButton(gtx, ContainerTypeCabinet)
									}),
									layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
										return ga.renderContainerTypeButton(gtx, ContainerTypeGeneral)
									}),
								)
							})
						}),
					)
				})
			}),
		)
	})
}

// renderContainerTypeButton renders a container type selection button
func (ga *GioApp) renderContainerTypeButton(gtx layout.Context, containerType string) layout.Dimensions {
	// Get or create button state
	if ga.widgetState.containerTypeButtons[containerType] == nil {
		ga.widgetState.containerTypeButtons[containerType] = &widget.Clickable{}
	}
	btn := ga.widgetState.containerTypeButtons[containerType]

	// Handle click - store in a temporary state variable
	if btn.Clicked(gtx) {
		ga.selectedContainerType = containerType
	}

	// Render button with appropriate style
	isSelected := ga.selectedContainerType == containerType
	return layout.Inset{Right: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if isSelected {
			return widgets.AccentButton(ga.theme.Theme, btn, containerTypeLabels[containerType])(gtx)
		}
		return widgets.CancelButton(ga.theme.Theme, btn, containerTypeLabels[containerType])(gtx)
	})
}

// renderParentContainerSelector renders parent container selection buttons
func (ga *GioApp) renderParentContainerSelector(gtx layout.Context) layout.Dimensions {
	if len(ga.containers) == 0 {
		return layout.Dimensions{}
	}

	return layout.Inset{Bottom: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Body2(ga.theme.Theme, "Parent Container")
				label.Color = theme.ColorTextSecondary
				return label.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					// "(none)" chip to make this a root container
					noneBtn := ga.getParentContainerButton("")
					if noneBtn.Clicked(gtx) {
						ga.selectedParentContainerID = nil
					}
					children := []layout.FlexChild{
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Right: unit.Dp(theme.Spacing1), Bottom: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return ga.renderFilterChip(gtx, noneBtn, "(none)", ga.selectedParentContainerID == nil)
							})
						}),
					}
					// Exclude the container being edited from parent options
					editingID := ""
					if ga.selectedContainer != nil {
						editingID = ga.selectedContainer.ID
					}
					for _, c := range ga.containers {
						c := c
						if c.ID == editingID {
							continue
						}
						btn := ga.getParentContainerButton(c.ID)
						if btn.Clicked(gtx) {
							cid := c.ID
							ga.selectedParentContainerID = &cid
						}
						active := ga.selectedParentContainerID != nil && *ga.selectedParentContainerID == c.ID
						children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Right: unit.Dp(theme.Spacing1), Bottom: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return ga.renderFilterChip(gtx, btn, c.Name, active)
							})
						}))
					}
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, children...)
				})
			}),
		)
	})
}

func (ga *GioApp) getParentContainerButton(containerID string) *widget.Clickable {
	if ga.widgetState.parentContainerButtons == nil {
		ga.widgetState.parentContainerButtons = make(map[string]*widget.Clickable)
	}
	if btn, ok := ga.widgetState.parentContainerButtons[containerID]; ok {
		return btn
	}
	btn := new(widget.Clickable)
	ga.widgetState.parentContainerButtons[containerID] = btn
	return btn
}

// renderObjectContainerSelector renders container selection for objects (optional)
func (ga *GioApp) renderObjectContainerSelector(gtx layout.Context) layout.Dimensions {
	if len(ga.containers) == 0 {
		return layout.Dimensions{}
	}

	return layout.Inset{Bottom: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Body2(ga.theme.Theme, "Container")
				label.Color = theme.ColorTextSecondary
				return label.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					// "(none)" chip
					noneBtn := ga.getObjectContainerButton("")
					if noneBtn.Clicked(gtx) {
						ga.selectedContainerID = nil
					}
					children := []layout.FlexChild{
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Right: unit.Dp(theme.Spacing1), Bottom: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return ga.renderFilterChip(gtx, noneBtn, "(none)", ga.selectedContainerID == nil)
							})
						}),
					}
					for _, c := range ga.containers {
						c := c
						btn := ga.getObjectContainerButton(c.ID)
						if btn.Clicked(gtx) {
							cid := c.ID
							ga.selectedContainerID = &cid
						}
						active := ga.selectedContainerID != nil && *ga.selectedContainerID == c.ID
						children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Right: unit.Dp(theme.Spacing1), Bottom: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return ga.renderFilterChip(gtx, btn, c.Name, active)
							})
						}))
					}
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, children...)
				})
			}),
		)
	})
}

func (ga *GioApp) getObjectContainerButton(containerID string) *widget.Clickable {
	if ga.widgetState.objectContainerButtons == nil {
		ga.widgetState.objectContainerButtons = make(map[string]*widget.Clickable)
	}
	if btn, ok := ga.widgetState.objectContainerButtons[containerID]; ok {
		return btn
	}
	btn := new(widget.Clickable)
	ga.widgetState.objectContainerButtons[containerID] = btn
	return btn
}

// handleContainerCreate handles creating a new container
func (ga *GioApp) handleContainerCreate() {
	if ga.selectedCollection == nil {
		ga.logger.Error("No collection selected for container creation")
		return
	}

	name := ga.widgetState.containerNameEditor.Text()
	location := ga.widgetState.containerLocationEditor.Text()

	if name == "" {
		ga.logger.Warn("Container name is required")
		return
	}

	containerType := ga.selectedContainerType
	if containerType == "" {
		containerType = ContainerTypeGeneral
	}

	ga.logger.Info("Creating container", "name", name, "type", containerType)

	collectionID := ga.selectedCollection.ID
	userID := ga.currentUser.ID
	parentContainerID := ga.selectedParentContainerID

	go func() {
		req := types.CreateContainerRequest{
			CollectionID:      collectionID,
			Name:              name,
			Type:              containerType,
			Location:          location,
			ParentContainerID: parentContainerID,
		}

		container, err := ga.containersClient.Create(userID, collectionID, req)
		if err != nil {
			ga.logger.Error("Failed to create container", "error", err)
			return
		}

		ga.logger.Info("Container created successfully", "container_id", container.ID)
		ga.do(func() { ga.addContainer(*container) })
	}()

	// Close dialog
	ga.showContainerDialog = false
	ga.selectedContainerType = ""
	ga.selectedParentContainerID = nil
}

// handleContainerUpdate handles updating an existing container
func (ga *GioApp) handleContainerUpdate() {
	if ga.selectedContainer == nil {
		ga.logger.Error("No container selected for update")
		return
	}

	name := ga.widgetState.containerNameEditor.Text()
	location := ga.widgetState.containerLocationEditor.Text()

	if name == "" {
		ga.logger.Warn("Container name is required")
		return
	}

	ga.logger.Info("Updating container", "container_id", ga.selectedContainer.ID, "name", name)

	containerID := ga.selectedContainer.ID
	collectionID := ga.selectedCollection.ID
	userID := ga.currentUser.ID

	// Capture parent ID decision before launching goroutine
	var parentID *string
	if ga.selectedParentContainerID != nil {
		parentID = ga.selectedParentContainerID
	} else if ga.selectedContainer != nil && ga.selectedContainer.ParentContainerID != nil {
		// User explicitly selected "(none)" to remove parent
		empty := ""
		parentID = &empty
	}

	go func() {
		req := types.UpdateContainerRequest{
			Name:              name,
			Location:          location,
			ParentContainerID: parentID,
		}

		updated, err := ga.containersClient.Update(userID, collectionID, containerID, req)
		if err != nil {
			ga.logger.Error("Failed to update container", "error", err)
			return
		}

		ga.logger.Info("Container updated successfully", "container_id", containerID)
		ga.do(func() { ga.updateContainer(*updated) })
	}()

	// Close dialog
	ga.showContainerDialog = false
	ga.selectedContainer = nil
	ga.selectedParentContainerID = nil
}

// handleContainerDelete handles deleting a container
func (ga *GioApp) handleContainerDelete() {
	if ga.deleteContainerID == "" {
		ga.logger.Error("No container ID for deletion")
		return
	}

	ga.logger.Info("Deleting container", "container_id", ga.deleteContainerID)

	containerID := ga.deleteContainerID
	collectionID := ga.selectedCollection.ID
	userID := ga.currentUser.ID

	go func() {
		err := ga.containersClient.Delete(userID, collectionID, containerID)
		if err != nil {
			ga.logger.Error("Failed to delete container", "error", err)
			return
		}

		ga.logger.Info("Container deleted successfully", "container_id", containerID)
		ga.do(func() { ga.removeContainer(containerID) })
	}()

	// Close dialog
	ga.showDeleteContainer = false
	ga.deleteContainerID = ""
}

// handleObjectCreate handles creating a new object
func (ga *GioApp) handleObjectCreate() {
	if ga.selectedCollection == nil {
		ga.logger.Error("No collection selected for object creation")
		return
	}

	name := ga.widgetState.objectNameEditor.Text()
	description := ga.widgetState.objectDescriptionEditor.Text()
	quantityText := ga.widgetState.objectQuantityEditor.Text()
	unit := ga.widgetState.objectUnitEditor.Text()

	if name == "" {
		ga.logger.Warn("Object name is required")
		return
	}

	// Parse quantity if provided
	var quantity *float64
	if quantityText != "" {
		if val, err := strconv.ParseFloat(quantityText, 64); err == nil {
			quantity = &val
		}
	}

	ga.logger.Info("Creating object", "name", name)

	userID := ga.currentUser.ID
	objectType := ga.selectedCollection.ObjectType
	collectionID := ga.selectedCollection.ID
	selectedContainerID := ga.selectedContainerID

	go func() {
		req := types.CreateObjectRequest{
			Name:        name,
			Description: description,
			ObjectType:  objectType,
			Quantity:    quantity,
			Unit:        unit,
			Properties:  make(map[string]interface{}),
			Tags:        []string{},
		}

		// Add container ID if selected
		if selectedContainerID != nil {
			req.ContainerID = *selectedContainerID
		}

		object, err := ga.objectsClient.Create(userID, req, collectionID)
		if err != nil {
			ga.logger.Error("Failed to create object", "error", err)
			return
		}

		ga.logger.Info("Object created successfully", "object_id", object.ID)
		ga.do(func() { ga.addObject(*object) })
	}()

	// Close dialog
	ga.showObjectDialog = false
	ga.selectedContainerID = nil
}

// handleObjectUpdate handles updating an existing object
func (ga *GioApp) handleObjectUpdate() {
	if ga.selectedObject == nil {
		ga.logger.Error("No object selected for update")
		return
	}

	name := ga.widgetState.objectNameEditor.Text()
	description := ga.widgetState.objectDescriptionEditor.Text()
	quantityText := ga.widgetState.objectQuantityEditor.Text()
	unit := ga.widgetState.objectUnitEditor.Text()

	if name == "" {
		ga.logger.Warn("Object name is required")
		return
	}

	// Parse quantity if provided
	var quantity *float64
	if quantityText != "" {
		if val, err := strconv.ParseFloat(quantityText, 64); err == nil {
			quantity = &val
		}
	}

	ga.logger.Info("Updating object", "object_id", ga.selectedObject.ID, "name", name)

	objectID := ga.selectedObject.ID
	userID := ga.currentUser.ID

	containerID := ""
	if ga.selectedContainerID != nil {
		containerID = *ga.selectedContainerID
	} else if ga.selectedObject != nil {
		containerID = ga.selectedObject.ContainerID
	}

	// Preserve existing properties and tags
	properties := ga.selectedObject.Properties
	tags := ga.selectedObject.Tags

	oldContainerID := ""
	if ga.selectedObject != nil {
		oldContainerID = ga.selectedObject.ContainerID
	}

	go func() {
		req := types.UpdateObjectRequest{
			ContainerID: containerID,
			Name:        &name,
			Description: &description,
			Quantity:    quantity,
			Unit:        &unit,
			Properties:  properties,
			Tags:        tags,
		}

		updated, err := ga.objectsClient.Update(userID, objectID, req)
		if err != nil {
			ga.logger.Error("Failed to update object", "error", err)
			return
		}

		ga.logger.Info("Object updated successfully", "object_id", objectID)
		ga.do(func() { ga.updateObject(*updated, oldContainerID) })
	}()

	// Close dialog
	ga.showObjectDialog = false
	ga.selectedObject = nil
}

// handleObjectDelete handles deleting an object
func (ga *GioApp) handleObjectDelete() {
	if ga.deleteObjectID == "" {
		ga.logger.Error("No object ID for deletion")
		return
	}

	// Find the object to get its container ID
	var containerID string
	for _, obj := range ga.objects {
		if obj.ID == ga.deleteObjectID {
			containerID = obj.ContainerID
			break
		}
	}

	ga.logger.Info("Deleting object", "object_id", ga.deleteObjectID)

	objectID := ga.deleteObjectID
	userID := ga.currentUser.ID

	go func() {
		err := ga.objectsClient.Delete(userID, objectID, containerID)
		if err != nil {
			ga.logger.Error("Failed to delete object", "error", err)
			return
		}

		ga.logger.Info("Object deleted successfully", "object_id", objectID)
		ga.do(func() { ga.removeObject(objectID, containerID) })
	}()

	// Close dialog
	ga.showDeleteObject = false
	ga.deleteObjectID = ""
}
