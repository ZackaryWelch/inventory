package app

import (
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/nishiki/frontend/ui/theme"
	"github.com/nishiki/frontend/ui/widgets"
)

// openMembersDialog opens the member management dialog for a group.
func (ga *GioApp) openMembersDialog(group *Group) {
	ga.groupMembersOf = group
	ga.groupMembers = nil
	ga.knownUsers = nil
	ga.widgetState.memberItems = nil
	ga.widgetState.memberUserIDEditor.SetText("")
	ga.widgetState.memberSearchEditor.SetText("")
	ga.widgetState.membersDialog.Reset()
	ga.showMembersDialog = true

	go func() {
		members, err := ga.groupsClient.GetMembers(group.ID)
		if err != nil {
			ga.logger.Error("Failed to fetch group members", "error", err)
			return
		}
		ga.groupMembers = members
		ga.widgetState.memberItems = make([]MemberItemState, len(members))

		// Load known users from all other groups
		ga.loadKnownUsers(group.ID, members)
		ga.window.Invalidate()
	}()
}

// loadKnownUsers fetches all users from all groups, excluding current members,
// and stores them in ga.knownUsers for the user picker.
func (ga *GioApp) loadKnownUsers(excludeGroupID string, currentMembers []User) {
	memberIDs := make(map[string]bool, len(currentMembers))
	for _, m := range currentMembers {
		memberIDs[m.ID] = true
	}

	seen := make(map[string]bool)
	var known []User

	for _, g := range ga.groups {
		if g.ID == excludeGroupID {
			continue
		}
		users, err := ga.groupsClient.GetMembers(g.ID)
		if err != nil {
			ga.logger.Warn("Failed to fetch members for group", "group_id", g.ID, "error", err)
			continue
		}
		for _, u := range users {
			if !seen[u.ID] && !memberIDs[u.ID] {
				seen[u.ID] = true
				known = append(known, u)
			}
		}
	}

	ga.knownUsers = known
	ga.window.Invalidate()
}

// renderMembersDialog renders the group member management dialog overlay.
func (ga *GioApp) renderMembersDialog(gtx layout.Context) layout.Dimensions {
	if !ga.showMembersDialog {
		return layout.Dimensions{}
	}

	// Handle close
	if ga.widgetState.membersDialogClose.Clicked(gtx) {
		ga.showMembersDialog = false
		ga.groupMembersOf = nil
		ga.groupMembers = nil
		ga.knownUsers = nil
		ga.widgetState.membersDialog.Reset()
		return layout.Dimensions{}
	}

	// Handle add member via explicit user ID field
	if ga.widgetState.membersAddButton.Clicked(gtx) {
		ga.handleAddMember()
	}

	// Collect known-user click before layout
	var addKnownUserID string
	searchText := strings.ToLower(ga.widgetState.memberSearchEditor.Text())
	var filtered []User
	for _, u := range ga.knownUsers {
		if searchText == "" ||
			strings.Contains(strings.ToLower(u.Name), searchText) ||
			strings.Contains(strings.ToLower(u.Email), searchText) {
			filtered = append(filtered, u)
		}
	}
	for _, u := range filtered {
		btn := ga.knownUserClickable(u.ID)
		if btn.Clicked(gtx) {
			addKnownUserID = u.ID
		}
	}

	title := "Members"
	if ga.groupMembersOf != nil {
		title = "Members — " + ga.groupMembersOf.Name
	}

	var removeMemberUserID string

	dialogStyle := widgets.DefaultDialogStyle(ga.widgetState.membersDialog, title)
	dialogStyle.Width = unit.Dp(500)

	dims, dismissed := dialogStyle.Layout(gtx, ga.theme.Theme, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Current members list
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if len(ga.groupMembers) == 0 {
					return layout.Inset{Bottom: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						label := material.Body2(ga.theme.Theme, "No members yet.")
						label.Color = theme.ColorTextSecondary
						return label.Layout(gtx)
					})
				}
				maxH := gtx.Dp(unit.Dp(220))
				if gtx.Constraints.Max.Y > maxH {
					gtx.Constraints.Max.Y = maxH
				}
				list := &ga.widgetState.membersList
				list.Axis = layout.Vertical
				return list.Layout(gtx, len(ga.groupMembers), func(gtx layout.Context, i int) layout.Dimensions {
					if i < len(ga.widgetState.memberItems) {
						if ga.widgetState.memberItems[i].removeButton.Clicked(gtx) {
							removeMemberUserID = ga.groupMembers[i].ID
						}
					}
					return layout.Inset{Bottom: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return ga.renderMemberRow(gtx, ga.groupMembers[i], i)
					})
				})
			}),

			// Add member section header
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: unit.Dp(theme.Spacing3), Bottom: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					label := material.Body2(ga.theme.Theme, "Add Member")
					label.Color = theme.ColorTextSecondary
					return label.Layout(gtx)
				})
			}),

			// User picker: search field + scrollable list of known users
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Bottom: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Bottom: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return ga.renderFormField(gtx, "Search users", &ga.widgetState.memberSearchEditor, "Name or email…")
							})
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if len(ga.knownUsers) == 0 && searchText == "" {
								label := material.Body2(ga.theme.Theme, "No other users found in your groups.")
								label.Color = theme.ColorTextSecondary
								return label.Layout(gtx)
							}
							maxH := gtx.Dp(unit.Dp(160))
							if gtx.Constraints.Max.Y > maxH {
								gtx.Constraints.Max.Y = maxH
							}
							list := &ga.widgetState.knownUsersList
							list.Axis = layout.Vertical
							return list.Layout(gtx, len(filtered), func(gtx layout.Context, i int) layout.Dimensions {
								u := filtered[i]
								return layout.Inset{Bottom: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									return ga.renderKnownUserRow(gtx, u)
								})
							})
						}),
					)
				})
			}),

			// Manual user ID fallback
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Bottom: unit.Dp(theme.Spacing4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Right: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return ga.renderFormField(gtx, "User ID", &ga.widgetState.memberUserIDEditor, "Paste user ID directly")
							})
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return widgets.PrimaryButton(ga.theme.Theme, &ga.widgetState.membersAddButton, "Add")(gtx)
						}),
					)
				})
			}),

			// Close button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceEnd}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return widgets.CancelButton(ga.theme.Theme, &ga.widgetState.membersDialogClose, "Close")(gtx)
					}),
				)
			}),
		)
	})

	if removeMemberUserID != "" {
		ga.handleRemoveMember(removeMemberUserID)
	}
	if addKnownUserID != "" {
		ga.handleAddMemberByID(addKnownUserID)
	}

	if dismissed {
		ga.showMembersDialog = false
		ga.groupMembersOf = nil
		ga.groupMembers = nil
		ga.knownUsers = nil
		ga.widgetState.membersDialog.Reset()
	}

	return dims
}

// knownUserClickable returns (creating if needed) the clickable for a known user row.
func (ga *GioApp) knownUserClickable(userID string) *widget.Clickable {
	if btn, ok := ga.widgetState.knownUserClickables[userID]; ok {
		return btn
	}
	btn := new(widget.Clickable)
	ga.widgetState.knownUserClickables[userID] = btn
	return btn
}

// renderMemberRow renders a single current-member row card.
func (ga *GioApp) renderMemberRow(gtx layout.Context, member User, index int) layout.Dimensions {
	card := widgets.DefaultCard()
	return card.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := material.Body1(ga.theme.Theme, member.Name)
						label.Font.Weight = font.Bold
						return label.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := material.Body2(ga.theme.Theme, member.Email)
						label.Color = theme.ColorTextSecondary
						return label.Layout(gtx)
					}),
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if index < len(ga.widgetState.memberItems) {
					return widgets.DangerButton(ga.theme.Theme, &ga.widgetState.memberItems[index].removeButton, "Remove")(gtx)
				}
				return layout.Dimensions{}
			}),
		)
	})
}

// renderKnownUserRow renders a known user row that can be clicked to add.
func (ga *GioApp) renderKnownUserRow(gtx layout.Context, u User) layout.Dimensions {
	btn := ga.knownUserClickable(u.ID)
	card := widgets.DefaultCard()
	return card.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := material.Body2(ga.theme.Theme, u.Name)
						label.Font.Weight = font.Bold
						return label.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := material.Body2(ga.theme.Theme, u.Email)
						label.Color = theme.ColorTextSecondary
						return label.Layout(gtx)
					}),
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return widgets.AccentButton(ga.theme.Theme, btn, "Add")(gtx)
			}),
		)
	})
}

// handleAddMember adds a member using the manual user ID field.
func (ga *GioApp) handleAddMember() {
	userID := strings.TrimSpace(ga.widgetState.memberUserIDEditor.Text())
	if userID == "" {
		return
	}
	ga.widgetState.memberUserIDEditor.SetText("")
	ga.handleAddMemberByID(userID)
}

// handleAddMemberByID adds a member to the current group by user ID.
func (ga *GioApp) handleAddMemberByID(userID string) {
	if ga.groupMembersOf == nil {
		return
	}
	groupID := ga.groupMembersOf.ID

	go func() {
		if err := ga.groupsClient.AddMember(groupID, userID); err != nil {
			ga.logger.Error("Failed to add member", "error", err)
			return
		}
		ga.logger.Info("Member added", "group_id", groupID, "user_id", userID)
		ga.refreshGroupMembers(groupID)
	}()
}

// handleRemoveMember removes a member from the current group.
func (ga *GioApp) handleRemoveMember(userID string) {
	if ga.groupMembersOf == nil {
		return
	}
	groupID := ga.groupMembersOf.ID

	go func() {
		if err := ga.groupsClient.RemoveMember(groupID, userID); err != nil {
			ga.logger.Error("Failed to remove member", "error", err)
			return
		}
		ga.logger.Info("Member removed", "group_id", groupID, "user_id", userID)
		ga.refreshGroupMembers(groupID)
	}()
}

// refreshGroupMembers reloads the members list and known users for the current group dialog.
func (ga *GioApp) refreshGroupMembers(groupID string) {
	members, err := ga.groupsClient.GetMembers(groupID)
	if err != nil {
		ga.logger.Error("Failed to refresh members", "error", err)
		return
	}
	ga.groupMembers = members
	ga.widgetState.memberItems = make([]MemberItemState, len(members))
	if ga.groupMembersOf != nil {
		ga.loadKnownUsers(groupID, members)
	}
	ga.window.Invalidate()
}
