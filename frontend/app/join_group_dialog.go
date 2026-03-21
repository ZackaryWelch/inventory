package app

import (
	"strings"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"github.com/nishiki/frontend/ui/theme"
	"github.com/nishiki/frontend/ui/widgets"
)

// openJoinGroupDialog opens the join-group-by-invite-hash dialog.
func (ga *GioApp) openJoinGroupDialog() {
	ga.widgetState.joinHashEditor.SetText("")
	ga.widgetState.joinGroupDialog.Reset()
	ga.showJoinGroupDialog = true
}

// renderJoinGroupDialog renders the join group dialog overlay.
func (ga *GioApp) renderJoinGroupDialog(gtx layout.Context) layout.Dimensions {
	if !ga.showJoinGroupDialog {
		return layout.Dimensions{}
	}

	if ga.widgetState.joinGroupClose.Clicked(gtx) {
		ga.showJoinGroupDialog = false
		ga.widgetState.joinGroupDialog.Reset()
		return layout.Dimensions{}
	}

	if ga.widgetState.joinGroupButton.Clicked(gtx) {
		ga.handleJoinGroup()
	}

	dialogStyle := widgets.DefaultDialogStyle(ga.widgetState.joinGroupDialog, "Join Group")
	dialogStyle.Width = unit.Dp(400)

	dims, dismissed := dialogStyle.Layout(gtx, ga.theme.Theme, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Bottom: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					label := material.Body2(ga.theme.Theme, "Invite Hash")
					label.Color = theme.ColorTextSecondary
					return label.Layout(gtx)
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Bottom: unit.Dp(theme.Spacing4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return ga.renderFormField(gtx, "", &ga.widgetState.joinHashEditor, "Enter invite hash")
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceEnd}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Right: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return widgets.CancelButton(ga.theme.Theme, &ga.widgetState.joinGroupClose, "Cancel")(gtx)
						})
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return widgets.PrimaryButton(ga.theme.Theme, &ga.widgetState.joinGroupButton, "Join")(gtx)
					}),
				)
			}),
		)
	})

	if dismissed {
		ga.showJoinGroupDialog = false
		ga.widgetState.joinGroupDialog.Reset()
	}

	return dims
}

// handleJoinGroup joins a group using the entered invite hash.
func (ga *GioApp) handleJoinGroup() {
	hash := strings.TrimSpace(ga.widgetState.joinHashEditor.Text())
	if hash == "" {
		return
	}

	ga.widgetState.joinHashEditor.SetText("")
	ga.showJoinGroupDialog = false
	ga.widgetState.joinGroupDialog.Reset()

	go func() {
		group, err := ga.groupsClient.JoinByHash(hash)
		if err != nil {
			ga.logger.Error("Failed to join group", "error", err)
			return
		}
		ga.logger.Info("Joined group", "group_id", group.ID, "group_name", group.Name)
		ga.fetchGroups()
		ga.window.Invalidate()
	}()
}
