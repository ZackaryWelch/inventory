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

	"github.com/nishiki/frontend/pkg/types"
	"github.com/nishiki/frontend/ui/theme"
	"github.com/nishiki/frontend/ui/widgets"
)

// renderGroupsView renders the groups management view with CRUD operations
func (ga *GioApp) renderGroupsView(gtx layout.Context) layout.Dimensions {
	// Handle create button click
	if ga.widgetState.groupsCreateButton.Clicked(gtx) {
		ga.logger.Info("Opening create group dialog")
		ga.showGroupDialog = true
		ga.groupDialogMode = "create"
		// Clear editors
		ga.widgetState.groupNameEditor.SetText("")
		ga.widgetState.groupDescriptionEditor.SetText("")
	}

	// Handle bottom menu clicks
	if ga.widgetState.menuDashboard.Clicked(gtx) {
		ga.currentView = ViewDashboardGio
	}
	if ga.widgetState.menuCollections.Clicked(gtx) {
		ga.currentView = ViewCollectionsGio
	}
	if ga.widgetState.menuProfile.Clicked(gtx) {
		ga.currentView = ViewProfileGio
	}

	// Ensure we have group item states for all groups
	ga.ensureGroupItemStates()

	// Main layout
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		// Header
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ga.renderHeader(gtx, "Groups")
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
						return ga.renderGroupsToolbar(gtx)
					}),

					// Groups list
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return ga.renderGroupsList(gtx)
					}),
				)
			})
		}),

		// Bottom navigation menu
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ga.renderBottomMenu(gtx, "groups")
		}),
	)
}

// ensureGroupItemStates ensures we have widget states for all groups
func (ga *GioApp) ensureGroupItemStates() {
	if len(ga.widgetState.groupItems) != len(ga.groups) {
		ga.widgetState.groupItems = make([]GroupItemState, len(ga.groups))
	}
}

// renderGroupsToolbar renders the toolbar with search and create button
func (ga *GioApp) renderGroupsToolbar(gtx layout.Context) layout.Dimensions {
	return layout.Inset{Bottom: unit.Dp(theme.Spacing4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:    layout.Horizontal,
			Spacing: layout.SpaceBetween,
		}.Layout(gtx,
			// Search field
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Right: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					editor := material.Editor(ga.theme.Theme, &ga.widgetState.groupsSearchField, "Search groups...")
					editor.Color = theme.ColorTextPrimary
					return editor.Layout(gtx)
				})
			}),

			// Create button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return widgets.PrimaryButton(ga.theme.Theme, &ga.widgetState.groupsCreateButton, "+")(gtx)
			}),
		)
	})
}

// renderGroupsList renders the list of groups
func (ga *GioApp) renderGroupsList(gtx layout.Context) layout.Dimensions {
	if len(ga.groups) == 0 {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			label := material.H5(ga.theme.Theme, "No groups yet")
			label.Color = theme.ColorTextSecondary
			label.Alignment = text.Middle
			return label.Layout(gtx)
		})
	}

	// Filter groups based on search query
	searchQuery := strings.ToLower(ga.widgetState.groupsSearchField.Text())
	filteredGroups := make([]Group, 0)
	filteredIndices := make([]int, 0)

	for i, group := range ga.groups {
		if searchQuery == "" ||
			strings.Contains(strings.ToLower(group.Name), searchQuery) ||
			strings.Contains(strings.ToLower(group.Description), searchQuery) {
			filteredGroups = append(filteredGroups, group)
			filteredIndices = append(filteredIndices, i)
		}
	}

	// Render list using widget state
	list := &ga.widgetState.groupsList
	list.Axis = layout.Vertical
	return list.Layout(gtx, len(filteredGroups), func(gtx layout.Context, index int) layout.Dimensions {
		group := filteredGroups[index]
		originalIndex := filteredIndices[index]
		return layout.Inset{Bottom: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return ga.renderGroupCard(gtx, group, originalIndex)
		})
	})
}

// renderGroupCard renders a single group card
func (ga *GioApp) renderGroupCard(gtx layout.Context, group Group, index int) layout.Dimensions {
	itemState := &ga.widgetState.groupItems[index]

	// Handle edit button click
	if itemState.editButton.Clicked(gtx) {
		ga.logger.Info("Opening edit group dialog", "group_id", group.ID, "group_name", group.Name)
		ga.selectedGroup = &group
		ga.showGroupDialog = true
		ga.groupDialogMode = "edit"
		ga.widgetState.groupNameEditor.SetText(group.Name)
		ga.widgetState.groupDescriptionEditor.SetText(group.Description)
	}

	// Handle delete button click
	if itemState.deleteButton.Clicked(gtx) {
		ga.logger.Info("Opening delete confirmation", "group_id", group.ID, "group_name", group.Name)
		ga.showDeleteConfirm = true
		ga.deleteGroupID = group.ID
	}

	card := widgets.DefaultCard()
	return card.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			// Header row (name + buttons)
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
					Spacing:   layout.SpaceBetween,
				}.Layout(gtx,
					// Group name
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						label := material.H6(ga.theme.Theme, group.Name)
						label.Font.Weight = font.Bold
						return label.Layout(gtx)
					}),

					// Action buttons
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
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
					}),
				)
			}),

			// Description
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if group.Description != "" {
					return layout.Inset{Top: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						label := material.Body2(ga.theme.Theme, group.Description)
						label.Color = theme.ColorTextSecondary
						return label.Layout(gtx)
					})
				}
				return layout.Dimensions{}
			}),
		)
	})
}

// handleGroupCreate handles creating a new group
func (ga *GioApp) handleGroupCreate() {
	name := ga.widgetState.groupNameEditor.Text()
	description := ga.widgetState.groupDescriptionEditor.Text()

	if name == "" {
		ga.logger.Warn("Group name is required")
		return
	}

	ga.logger.Info("Creating group", "name", name)

	go func() {
		req := types.CreateGroupRequest{
			Name:        name,
			Description: description,
		}

		group, err := ga.groupsClient.Create(req)
		if err != nil {
			ga.logger.Error("Failed to create group", "error", err)
			ga.ops <- Operation{Type: "group_create_failed", Err: err}
			return
		}

		ga.logger.Info("Group created successfully", "group_id", group.ID)
		ga.ops <- Operation{Type: "group_created", Data: group}
		// Refresh groups list
		ga.fetchGroups()
		ga.window.Invalidate()
	}()

	// Close dialog
	ga.showGroupDialog = false
	ga.window.Invalidate()
}

// handleGroupUpdate handles updating an existing group
func (ga *GioApp) handleGroupUpdate() {
	if ga.selectedGroup == nil {
		ga.logger.Error("No group selected for update")
		return
	}

	name := ga.widgetState.groupNameEditor.Text()
	description := ga.widgetState.groupDescriptionEditor.Text()

	if name == "" {
		ga.logger.Warn("Group name is required")
		return
	}

	ga.logger.Info("Updating group", "group_id", ga.selectedGroup.ID, "name", name)

	groupID := ga.selectedGroup.ID

	go func() {
		req := types.UpdateGroupRequest{
			Name:        name,
			Description: description,
		}

		_, err := ga.groupsClient.Update(groupID, req)
		if err != nil {
			ga.logger.Error("Failed to update group", "error", err)
			ga.ops <- Operation{Type: "group_update_failed", Err: err}
			return
		}

		ga.logger.Info("Group updated successfully", "group_id", groupID)
		ga.ops <- Operation{Type: "group_updated"}
		// Refresh groups list
		ga.fetchGroups()
		ga.window.Invalidate()
	}()

	// Close dialog
	ga.showGroupDialog = false
	ga.selectedGroup = nil
	ga.window.Invalidate()
}

// handleGroupDelete handles deleting a group
func (ga *GioApp) handleGroupDelete() {
	if ga.deleteGroupID == "" {
		ga.logger.Error("No group ID for deletion")
		return
	}

	ga.logger.Info("Deleting group", "group_id", ga.deleteGroupID)

	groupID := ga.deleteGroupID

	go func() {
		err := ga.groupsClient.Delete(groupID)
		if err != nil {
			ga.logger.Error("Failed to delete group", "error", err)
			ga.ops <- Operation{Type: "group_delete_failed", Err: err}
			return
		}

		ga.logger.Info("Group deleted successfully", "group_id", groupID)
		ga.ops <- Operation{Type: "group_deleted"}
		// Refresh groups list
		ga.fetchGroups()
		ga.window.Invalidate()
	}()

	// Close dialog
	ga.showDeleteConfirm = false
	ga.deleteGroupID = ""
	ga.window.Invalidate()
}

// renderGroupDialog renders the create/edit group dialog
func (ga *GioApp) renderGroupDialog(gtx layout.Context) layout.Dimensions {
	if !ga.showGroupDialog {
		return layout.Dimensions{}
	}

	// Handle submit button
	if ga.widgetState.groupDialogSubmit.Clicked(gtx) {
		if ga.groupDialogMode == "create" {
			ga.handleGroupCreate()
		} else {
			ga.handleGroupUpdate()
		}
		return layout.Dimensions{}
	}

	// Handle cancel button
	if ga.widgetState.groupDialogCancel.Clicked(gtx) {
		ga.showGroupDialog = false
		ga.selectedGroup = nil
		return layout.Dimensions{}
	}

	// Render modal overlay
	return ga.renderModal(gtx, func(gtx layout.Context) layout.Dimensions {
		card := widgets.Card{
			BackgroundColor: theme.ColorWhite,
			CornerRadius:    unit.Dp(theme.RadiusLG),
			Inset:           layout.UniformInset(unit.Dp(theme.Spacing6)),
		}

		return card.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				// Title
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					title := "Create Group"
					if ga.groupDialogMode == "edit" {
						title = "Edit Group"
					}
					return layout.Inset{Bottom: unit.Dp(theme.Spacing4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						label := material.H6(ga.theme.Theme, title)
						label.Font.Weight = font.Bold
						return label.Layout(gtx)
					})
				}),

				// Name field
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								label := material.Body2(ga.theme.Theme, "Name *")
								label.Color = theme.ColorTextSecondary
								return label.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								editor := material.Editor(ga.theme.Theme, &ga.widgetState.groupNameEditor, "Enter group name")
								return editor.Layout(gtx)
							}),
						)
					})
				}),

				// Description field
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(theme.Spacing4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								label := material.Body2(ga.theme.Theme, "Description")
								label.Color = theme.ColorTextSecondary
								return label.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								editor := material.Editor(ga.theme.Theme, &ga.widgetState.groupDescriptionEditor, "Enter description (optional)")
								return editor.Layout(gtx)
							}),
						)
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
								return widgets.CancelButton(ga.theme.Theme, &ga.widgetState.groupDialogCancel, "Cancel")(gtx)
							})
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							submitText := "Create"
							if ga.groupDialogMode == "edit" {
								submitText = "Update"
							}
							return widgets.PrimaryButton(ga.theme.Theme, &ga.widgetState.groupDialogSubmit, submitText)(gtx)
						}),
					)
				}),
			)
		})
	})
}

// renderDeleteConfirmDialog renders the delete confirmation dialog
func (ga *GioApp) renderDeleteConfirmDialog(gtx layout.Context) layout.Dimensions {
	if !ga.showDeleteConfirm {
		return layout.Dimensions{}
	}

	// Find group name for confirmation
	var groupName string
	for _, group := range ga.groups {
		if group.ID == ga.deleteGroupID {
			groupName = group.Name
			break
		}
	}

	// Handle confirm button
	if ga.widgetState.groupDialogSubmit.Clicked(gtx) {
		ga.handleGroupDelete()
		return layout.Dimensions{}
	}

	// Handle cancel button
	if ga.widgetState.groupDialogCancel.Clicked(gtx) {
		ga.showDeleteConfirm = false
		ga.deleteGroupID = ""
		return layout.Dimensions{}
	}

	// Render modal overlay
	return ga.renderModal(gtx, func(gtx layout.Context) layout.Dimensions {
		card := widgets.Card{
			BackgroundColor: theme.ColorWhite,
			CornerRadius:    unit.Dp(theme.RadiusLG),
			Inset:           layout.UniformInset(unit.Dp(theme.Spacing6)),
		}

		return card.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				// Title
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(theme.Spacing4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						label := material.H6(ga.theme.Theme, "Delete Group")
						label.Font.Weight = font.Bold
						return label.Layout(gtx)
					})
				}),

				// Message
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(theme.Spacing4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						message := fmt.Sprintf("Are you sure you want to delete the group \"%s\"? This action cannot be undone.", groupName)
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
								return widgets.CancelButton(ga.theme.Theme, &ga.widgetState.groupDialogCancel, "Cancel")(gtx)
							})
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return widgets.DangerButton(ga.theme.Theme, &ga.widgetState.groupDialogSubmit, "Delete")(gtx)
						}),
					)
				}),
			)
		})
	})
}

// renderModal renders a modal overlay with content
func (ga *GioApp) renderModal(gtx layout.Context, content layout.Widget) layout.Dimensions {
	return layout.Stack{}.Layout(gtx,
		// Semi-transparent overlay
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			return widgets.Card{
				BackgroundColor: theme.ColorOverlay,
				CornerRadius:    0,
				Inset:           layout.Inset{},
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{Size: gtx.Constraints.Max}
			})
		}),
		// Content centered
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				// Constrain width for dialogs
				gtx.Constraints.Max.X = gtx.Dp(unit.Dp(400))
				return content(gtx)
			})
		}),
	)
}
