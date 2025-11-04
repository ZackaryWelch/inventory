//go:build js && wasm

package app

import (
	"fmt"
	"log/slog"
	"strings"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"

	"github.com/nishiki/frontend/pkg/types"
	"github.com/nishiki/frontend/ui/components"
	"github.com/nishiki/frontend/ui/layouts"
	appstyles "github.com/nishiki/frontend/ui/styles"
)

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

	// Refresh groups data
	if err := app.fetchGroups(); err != nil {
		fmt.Printf("Error fetching groups: %v\n", err)
	}

	// Page title - using helper function
	layouts.PageTitle(app.mainContainer, "Groups")

	// Main content - using existing layout function
	content := layouts.ContentColumn(app.mainContainer)

	// Action button row - React pattern: h-12 w-full flex items-center justify-end
	// This goes BEFORE the group cards list
	actionRow := core.NewFrame(content)
	actionRow.Styler(appstyles.StyleActionRowRight)

	// Create Group button (ghost icon button like React)
	createGroupBtn := core.NewButton(actionRow).SetIcon(icons.Add)
	createGroupBtn.Styler(appstyles.StyleCreateButton)
	createGroupBtn.OnClick(func(e events.Event) {
		app.showCreateGroupDialog() // Open dialog using Cogent Core's built-in system
	})

	// Groups list
	if len(app.groups) == 0 {
		components.EmptyState(content, "No groups found. Create your first group to get started!")
	} else {
		for _, group := range app.groups {
			app.createEnhancedGroupCard(content, group)
		}
	}

	// Bottom navigation bar - FIXED at bottom (React pattern)
	app.updateBottomMenu("groups")

	app.body.Update()
}

// Create enhanced group card - EXACTLY matches React GroupCard.tsx structure
func (app *App) createEnhancedGroupCard(parent core.Widget, group Group) *core.Frame {
	// Container for group name + card (React structure has name ABOVE card)
	container := core.NewFrame(parent)
	container.Styler(appstyles.StyleGroupCardContainer)

	// Group name ABOVE the card (React: text-lg font-semibold mb-2)
	groupName := core.NewText(container).SetText(group.Name)
	groupName.Styler(appstyles.StyleGroupName)

	// Card - single horizontal row (React: className="flex justify-between items-center p-4")
	card := components.Card(container, components.CardProps{})
	card.Styler(appstyles.StyleGroupCard)

	// LEFT SECTION: Icon + Container Count (React: flex gap-2 items-center)
	leftSection := core.NewFrame(card)
	leftSection.Styler(appstyles.StyleGroupCardLeftSection)
	leftSection.OnClick(func(e events.Event) {
		app.showGroupDetailView(group)
	})

	// Icon (cheese emoji in React - using folder icon as placeholder)
	icon := core.NewIcon(leftSection).SetIcon(icons.Folder)
	icon.Styler(appstyles.StyleGroupIconAccent)

	// Container count
	containerCount := core.NewText(leftSection).SetText("0")
	containerCount.Styler(appstyles.StyleContainerCount)

	// RIGHT SECTION: User Avatars + User Count + Menu (React: flex gap-2 items-center)
	rightSection := core.NewFrame(card)
	rightSection.Styler(appstyles.StyleGroupCardRightSection)

	// User avatars (show up to 3)
	avatarsContainer := core.NewFrame(rightSection)
	avatarsContainer.Styler(appstyles.StyleAvatarsContainer)

	// User count (will be updated when members are fetched)
	userCount := core.NewText(rightSection).SetText("×...")
	userCount.Styler(appstyles.StyleUserCount)

	// Fetch members asynchronously
	go func() {
		members, err := app.groupsClient.GetMembers(group.ID)
		if err != nil {
			slog.Error("Failed to fetch group members", "error", err, "group_id", group.ID)
			return
		}

		// Update UI on main thread
		app.mainContainer.AsyncLock()
		defer app.mainContainer.AsyncUnlock()

		// Clear placeholder and add member avatars (max 3)
		avatarsContainer.DeleteChildren()
		maxAvatars := 3
		for i, member := range members {
			if i >= maxAvatars {
				break
			}
			avatar := core.NewIcon(avatarsContainer).SetIcon(icons.Person)
			avatar.Tooltip = member.Name
			avatar.Styler(appstyles.StyleMemberAvatarSmall)
		}

		// Update user count
		userCount.SetText(fmt.Sprintf("×%d", len(members)))
		avatarsContainer.Update()
	}()

	// Three-dot menu button
	menuButton := core.NewButton(rightSection).SetIcon(icons.MoreVert)
	menuButton.Styler(appstyles.StyleGroupMenuButton)
	menuButton.OnClick(func(e events.Event) {
		app.showEditGroupDialog(group)
	})

	return container
}

// createMemberCard creates a card for displaying a group member
func (app *App) createMemberCard(parent core.Widget, member User) *core.Frame {
	card := components.Card(parent, components.CardProps{})
	card.Styler(appstyles.StyleMemberCard)

	// Avatar
	avatar := core.NewIcon(card).SetIcon(icons.Person)
	avatar.Styler(appstyles.StyleMemberAvatarLarge)

	// Info section
	infoSection := core.NewFrame(card)
	infoSection.Styler(appstyles.StyleMemberInfo)

	// Name
	nameText := core.NewText(infoSection).SetText(member.Name)
	nameText.Styler(appstyles.StyleMemberName)

	// Email
	emailText := core.NewText(infoSection).SetText(member.Email)
	emailText.Styler(appstyles.StyleMemberEmail)

	return card
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

	// Container for members (will be populated asynchronously)
	membersContainer := core.NewFrame(content)
	membersContainer.Styler(appstyles.StyleMembersContainer)

	// Show loading state initially
	components.EmptyState(membersContainer, "Loading members...")

	// Fetch members asynchronously
	go func() {
		members, err := app.groupsClient.GetMembers(group.ID)
		if err != nil {
			slog.Error("Failed to fetch group members", "error", err, "group_id", group.ID)
			app.mainContainer.AsyncLock()
			membersContainer.DeleteChildren()
			components.EmptyState(membersContainer, "Failed to load members")
			membersContainer.Update()
			app.mainContainer.AsyncUnlock()
			return
		}

		// Update UI on main thread
		app.mainContainer.AsyncLock()
		defer app.mainContainer.AsyncUnlock()

		membersContainer.DeleteChildren()

		if len(members) == 0 {
			components.EmptyState(membersContainer, "No members in this group yet.")
		} else {
			for _, member := range members {
				app.createMemberCard(membersContainer, member)
			}
		}

		membersContainer.Update()
	}()

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
		emptyCollections.Styler(appstyles.StyleEmptyText)
	} else {
		for _, collection := range groupCollections {
			app.createCollectionCard(content, collection)
		}
	}

	app.mainContainer.Update()
}

// Member card for group detail view
// Removed showGroupActionsMenu - button now opens dialog directly

// Dialog functions
func (app *App) showCreateGroupDialog() {
	var nameField, descField *core.TextField

	app.showDialog(DialogConfig{
		Title:            "Create New Group",
		SubmitButtonText: "Create Group",
		SubmitButtonStyle: appstyles.StyleButtonPrimary,
		ContentBuilder: func(dialog core.Widget, closeDialog func()) {
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
		ContentBuilder: func(dialog core.Widget, closeDialog func()) {
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

// NOTE: Overlay management removed - using Cogent Core's built-in dialog system
// See ui_helpers.go showDialog() which uses d.RunDialog(app.body)

// Empty state component
func (app *App) createEmptyState(parent core.Widget, title, message string, icon icons.Icon) *core.Frame {
	emptyState := core.NewFrame(parent)
	emptyState.Styler(appstyles.StyleEmptyStateContainer)

	emptyIcon := core.NewIcon(emptyState).SetIcon(icon)
	emptyIcon.Styler(appstyles.StyleEmptyStateIcon)

	emptyTitle := core.NewText(emptyState).SetText(title)
	emptyTitle.Styler(appstyles.StyleEmptyStateTitle)

	emptyMessage := core.NewText(emptyState).SetText(message)
	emptyMessage.Styler(appstyles.StyleEmptyStateMessage)

	return emptyState
}

// API handlers (these would make actual HTTP requests)
func (app *App) handleCreateGroup(name, description string) {
	if strings.TrimSpace(name) == "" {
		app.logger.Error("Group name cannot be empty")
		return
	}

	// Create request using types
	req := types.CreateGroupRequest{
		Name:        name,
		Description: description,
	}

	// Make API call to create group using client
	app.logger.Info("Creating group", "name", name)
	group, err := app.groupsClient.Create(req)
	if err != nil {
		app.logger.Error("Failed to create group", "error", err)
		return
	}

	app.logger.Info("Group created successfully", "group_id", group.ID)

	// Dialog closes automatically
	app.fetchGroups() // Refresh the list
	app.showEnhancedGroupsView() // Refresh the view
}

func (app *App) handleEditGroup(groupID, name, description string) {
	if strings.TrimSpace(name) == "" {
		app.logger.Error("Group name cannot be empty")
		return
	}

	// Create request using types
	req := types.UpdateGroupRequest{
		Name:        name,
		Description: description,
	}

	// Make API call to update group using client
	app.logger.Info("Updating group", "group_id", groupID, "name", name)
	group, err := app.groupsClient.Update(groupID, req)
	if err != nil {
		app.logger.Error("Failed to update group", "error", err)
		return
	}

	app.logger.Info("Group updated successfully", "group_id", group.ID)

	// Dialog closes automatically
	app.fetchGroups()
	app.showEnhancedGroupsView()
}

func (app *App) handleDeleteGroup(groupID string) {
	// Make API call to delete group using client
	app.logger.Info("Deleting group", "group_id", groupID)
	err := app.groupsClient.Delete(groupID)
	if err != nil {
		app.logger.Error("Failed to delete group", "error", err)
		return
	}

	app.logger.Info("Group deleted successfully")

	// Dialog closes automatically
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
		ContentBuilder: func(dialog core.Widget, closeDialog func()) {
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
		ContentBuilder: func(dialog core.Widget, closeDialog func()) {
			inviteCode := core.NewText(dialog).SetText("ABC123XYZ") // Would be generated
			inviteCode.Styler(appstyles.StyleInviteCode)
		},
		OnSubmit: nil, // No submit button, just a close button
	})
}

func (app *App) handleJoinGroup(inviteCode string) {
	fmt.Printf("Joining group with code: %s\n", inviteCode)
	// Dialog closes automatically
}