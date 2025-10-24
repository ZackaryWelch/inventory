//go:build js && wasm

package app

import (
	"fmt"
	"image/color"
	"strings"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/text/rich"

	"github.com/nishiki/frontend/ui/components"
	"github.com/nishiki/frontend/ui/layouts"
	appstyles "github.com/nishiki/frontend/ui/styles"
)

// Container represents a container within a collection
type Container struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	CollectionID string    `json:"collection_id"`
	Objects      []Object  `json:"objects,omitempty"`
}

// Object represents an object within a container
type Object struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	ContainerID string                 `json:"container_id"`
	Properties  map[string]interface{} `json:"properties"`
	Tags        []string               `json:"tags"`
}

// Dialog states for UI management
type DialogState struct {
	createGroupOpen       bool
	editGroupOpen         bool
	deleteGroupOpen       bool
	createCollectionOpen  bool
	editCollectionOpen    bool
	deleteCollectionOpen  bool
	createContainerOpen   bool
	editContainerOpen     bool
	createObjectOpen      bool
	editObjectOpen        bool
	selectedGroup         *Group
	selectedCollection    *Collection
	selectedContainer     *Container
	selectedObject        *Object
}

// Add dialog state to App

// Enhanced Groups View with full CRUD operations
func (app *App) showEnhancedGroupsView() {
	app.mainContainer.DeleteChildren()
	app.currentView = "groups"

	// Header with back button
	layouts.SimpleHeader(app.mainContainer, "Groups", true, func() {
		app.showDashboardView()
	})

	// Refresh groups data
	if err := app.fetchGroups(); err != nil {
		fmt.Printf("Error fetching groups: %v\n", err)
	}

	// Main content
	content := layouts.ContentColumn(app.mainContainer)

	// Action buttons row
	actionsRow := core.NewFrame(content)
	actionsRow.Styler(appstyles.StyleActionsSplit)

	// Create group button using component
	components.Button(actionsRow, components.ButtonProps{
		Text:    "Create Group",
		Icon:    icons.Add,
		Variant: components.ButtonPrimary,
		Size:    components.ButtonSizeMedium,
		OnClick: func(e events.Event) {
			app.showCreateGroupDialog()
		},
	})

	// Join group button using component
	components.Button(actionsRow, components.ButtonProps{
		Text:    "Join Group",
		Icon:    icons.PersonAdd,
		Variant: components.ButtonAccent,
		Size:    components.ButtonSizeMedium,
		OnClick: func(e events.Event) {
			app.showJoinGroupDialog()
		},
	})

	// Groups list
	if len(app.groups) == 0 {
		components.EmptyState(content, "No groups found. Create your first group to get started!")
	} else {
		for _, group := range app.groups {
			app.createEnhancedGroupCard(content, group)
		}
	}

	app.mainContainer.Update()
}

// Create enhanced group card with action menu
func (app *App) createEnhancedGroupCard(parent core.Widget, group Group) *core.Frame {
	return app.createCard(parent, CardConfig{
		Icon:        icons.Group,
		IconColor:   appstyles.ColorPrimary,
		Title:       group.Name,
		Description: group.Description,
		Stats: []CardStat{
			{Label: "members", Value: fmt.Sprintf("%d", len(group.Members))},
		},
		OnClick: func() {
			app.showGroupDetailView(group)
		},
		Actions: []CardAction{
			{Icon: icons.Edit, Color: appstyles.ColorAccent, Tooltip: "Edit group", OnClick: func() {
				app.showEditGroupDialog(group)
			}},
			{Icon: icons.Delete, Color: appstyles.ColorDanger, Tooltip: "Delete group", OnClick: func() {
				app.showDeleteGroupDialog(group)
			}},
			{Icon: icons.PersonAdd, Color: appstyles.ColorPrimary, Tooltip: "Invite to group", OnClick: func() {
				app.showInviteToGroupDialog(group)
			}},
		},
	})
}

// Group Detail View
func (app *App) showGroupDetailView(group Group) {
	app.mainContainer.DeleteChildren()
	app.currentView = "group_detail"

	// Header with back button
	layouts.SimpleHeader(app.mainContainer, group.Name, true, func() {
		app.showEnhancedGroupsView()
	})

	// Main content
	content := layouts.ContentColumn(app.mainContainer)

	// Group info card using component
	infoCard := components.Card(content, components.CardProps{})

	// Description
	if group.Description != "" {
		components.CardTitle(infoCard, "Description")
		components.CardDescription(infoCard, group.Description)
	}

	// Members section
	membersTitle := core.NewText(content).SetText("Members")
	membersTitle.Styler(appstyles.StyleH2)

	if len(group.Members) == 0 {
		components.EmptyState(content, "No members in this group yet.")
	} else {
		for _, member := range group.Members {
			app.createMemberCard(content, member, group)
		}
	}

	// Collections section
	collectionsTitle := core.NewText(content).SetText("Collections")
	collectionsTitle.Styler(appstyles.StyleH2)

	// Filter collections for this group
	groupCollections := []Collection{}
	for _, collection := range app.collections {
		// In a real implementation, you'd filter by group membership
		groupCollections = append(groupCollections, collection)
	}

	if len(groupCollections) == 0 {
		emptyCollections := core.NewText(content).SetText("No collections in this group yet.")
		emptyCollections.Styler(func(s *styles.Style) {
			s.Color = colors.Uniform(appstyles.ColorGrayDark)
			s.Align.Self = styles.Center
		})
	} else {
		for _, collection := range groupCollections {
			app.createCollectionCard(content, collection)
		}
	}

	app.mainContainer.Update()
}

// Member card for group detail view
func (app *App) createMemberCard(parent core.Widget, member User, group Group) *core.Frame {
	// Build actions - only show remove button if not the current user
	var actions []CardAction
	if app.currentUser != nil && app.currentUser.ID != member.ID {
		actions = []CardAction{
			{Icon: icons.PersonRemove, Color: appstyles.ColorDanger, Tooltip: "Remove member", OnClick: func() {
				app.showRemoveMemberDialog(member, group)
			}},
		}
	}

	return app.createCard(parent, CardConfig{
		Icon:        icons.Person,
		IconColor:   appstyles.ColorPrimary,
		Title:       member.Username,
		Description: member.Email,
		Actions:     actions,
	})
}

// Dialog functions
func (app *App) showCreateGroupDialog() {
	var nameField, descField *core.TextField

	app.showDialog(DialogConfig{
		Title:            "Create New Group",
		SubmitButtonText: "Create Group",
		SubmitButtonStyle: appstyles.StyleButtonPrimary,
		ContentBuilder: func(dialog core.Widget) {
			nameField = createTextField(dialog, "Group name")
			descField = createTextField(dialog, "Description (optional)")
		},
		OnSubmit: func() {
			app.handleCreateGroup(nameField.Text(), descField.Text())
		},
	})
}

func (app *App) showEditGroupDialog(group Group) {
	var nameField, descField *core.TextField

	app.showDialog(DialogConfig{
		Title:            "Edit Group",
		SubmitButtonText: "Save Changes",
		SubmitButtonStyle: appstyles.StyleButtonPrimary,
		ContentBuilder: func(dialog core.Widget) {
			nameField = createTextField(dialog, "Group name")
			nameField.SetText(group.Name)
			descField = createTextField(dialog, "Description (optional)")
			descField.SetText(group.Description)
		},
		OnSubmit: func() {
			app.handleEditGroup(group.ID, nameField.Text(), descField.Text())
		},
	})
}

func (app *App) showDeleteGroupDialog(group Group) {
	app.showDialog(DialogConfig{
		Title:            "Delete Group",
		Message:          fmt.Sprintf("Are you sure you want to delete \"%s\"? This action cannot be undone.", group.Name),
		SubmitButtonText: "Delete",
		SubmitButtonStyle: appstyles.StyleButtonDanger,
		OnSubmit: func() {
			app.handleDeleteGroup(group.ID)
		},
	})
}

// Helper functions for overlay management
func (app *App) createOverlay() *core.Frame {
	overlay := core.NewFrame(app.mainContainer)
	overlay.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(appstyles.ColorOverlay) // Semi-transparent black
		s.Display = styles.Flex
		s.Align.Items = styles.Center
		s.Justify.Content = styles.Center
		// s.ZIndex = 1000 // ZIndex removed in v0.3.12
	})
	return overlay
}

func (app *App) showOverlay(overlay *core.Frame) {
	// Store reference for later hiding
	app.currentOverlay = overlay
	app.mainContainer.Update()
}

func (app *App) hideOverlay() {
	if app.currentOverlay != nil {
		app.currentOverlay.Delete()
		app.currentOverlay = nil
		app.mainContainer.Update()
	}
}

// Empty state component
func (app *App) createEmptyState(parent core.Widget, title, message string, icon icons.Icon) *core.Frame {
	emptyState := core.NewFrame(parent)
	emptyState.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Align.Items = styles.Center
		s.Justify.Content = styles.Center
		s.Gap.Set(units.Dp(16))
		s.Padding.Set(units.Dp(32))
		s.Margin.Top = units.Dp(32)
	})

	emptyIcon := core.NewIcon(emptyState).SetIcon(icon)
	emptyIcon.Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(appstyles.ColorGray)
		s.Font.Size = units.Dp(48)
	})

	emptyTitle := core.NewText(emptyState).SetText(title)
	emptyTitle.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(18)
		s.Font.Weight = appstyles.WeightSemiBold
		s.Color = colors.Uniform(appstyles.ColorGrayDark)
	})

	emptyMessage := core.NewText(emptyState).SetText(message)
	emptyMessage.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(14)
		s.Color = colors.Uniform(appstyles.ColorGrayDark)
		s.Text.Align = appstyles.AlignCenter
	})

	return emptyState
}

// API handlers (these would make actual HTTP requests)
func (app *App) handleCreateGroup(name, description string) {
	if strings.TrimSpace(name) == "" {
		// Show error message
		return
	}

	// Here you would make the API call to create the group
	// For now, we'll simulate success
	fmt.Printf("Creating group: %s - %s\n", name, description)
	
	app.hideOverlay()
	app.fetchGroups() // Refresh the list
	app.showEnhancedGroupsView() // Refresh the view
}

func (app *App) handleEditGroup(groupID, name, description string) {
	fmt.Printf("Editing group %s: %s - %s\n", groupID, name, description)
	
	app.hideOverlay()
	app.fetchGroups()
	app.showEnhancedGroupsView()
}

func (app *App) handleDeleteGroup(groupID string) {
	fmt.Printf("Deleting group: %s\n", groupID)
	
	app.hideOverlay()
	app.fetchGroups()
	app.showEnhancedGroupsView()
}

// Additional dialog functions
func (app *App) showJoinGroupDialog() {
	var inviteField *core.TextField

	app.showDialog(DialogConfig{
		Title:            "Join Group",
		SubmitButtonText: "Join Group",
		SubmitButtonStyle: appstyles.StyleButtonPrimary,
		ContentBuilder: func(dialog core.Widget) {
			inviteField = createTextField(dialog, "Invitation code")
		},
		OnSubmit: func() {
			app.handleJoinGroup(inviteField.Text())
		},
	})
}

func (app *App) showInviteToGroupDialog(group Group) {
	app.showDialog(DialogConfig{
		Title:   "Invite to Group",
		Message: "Share this invitation code:",
		ContentBuilder: func(dialog core.Widget) {
			inviteCode := core.NewText(dialog).SetText("ABC123XYZ") // Would be generated
			inviteCode.Styler(func(s *styles.Style) {
				s.Font.Family = rich.Monospace
				s.Background = colors.Uniform(color.RGBA{R: 240, G: 240, B: 240, A: 255})
				s.Padding.Set(units.Dp(8))
				s.Border.Radius = styles.BorderRadiusMedium
			})
		},
		OnSubmit: nil, // No submit button, just a close button
	})
}

func (app *App) showRemoveMemberDialog(member User, group Group) {
	app.showDialog(DialogConfig{
		Title:            "Remove Member",
		Message:          fmt.Sprintf("Remove \"%s\" from \"%s\"?", member.Username, group.Name),
		SubmitButtonText: "Remove",
		SubmitButtonStyle: appstyles.StyleButtonDanger,
		OnSubmit: func() {
			app.handleRemoveMember(member.ID, group.ID)
		},
	})
}

func (app *App) handleJoinGroup(inviteCode string) {
	fmt.Printf("Joining group with code: %s\n", inviteCode)
	app.hideOverlay()
}

func (app *App) handleRemoveMember(userID, groupID string) {
	fmt.Printf("Removing user %s from group %s\n", userID, groupID)
	app.hideOverlay()
}