package main

import (
	"fmt"
	"image/color"
	"strings"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
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
func (app *App) initDialogState() {
	app.dialogState = &DialogState{}
}

// Enhanced Groups View with full CRUD operations
func (app *App) showEnhancedGroupsView() {
	app.mainContainer.DeleteChildren()
	app.currentView = "groups"

	// Header with back button
	header := app.createHeader("Groups", true)

	// Refresh groups data
	if err := app.fetchGroups(); err != nil {
		fmt.Printf("Error fetching groups: %v\n", err)
	}

	// Main content
	content := core.NewFrame(app.mainContainer)
	content.Style(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Grow.Set(1, 1)
		s.Padding.Set(core.Dp(16))
		s.Gap.Set(core.Dp(16))
	})

	// Action buttons row
	actionsRow := core.NewFrame(content)
	actionsRow.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(core.Dp(12))
		s.Justify.Content = styles.End
	})

	// Create group button
	createBtn := core.NewButton(actionsRow).SetText("Create Group").SetIcon(icons.Add)
	app.styleButtonPrimary(createBtn)
	createBtn.OnClick(func(e events.Event) {
		app.showCreateGroupDialog()
	})

	// Join group button
	joinBtn := core.NewButton(actionsRow).SetText("Join Group").SetIcon(icons.PersonAdd)
	app.styleButtonAccent(joinBtn)
	joinBtn.OnClick(func(e events.Event) {
		app.showJoinGroupDialog()
	})

	// Groups list
	if len(app.groups) == 0 {
		emptyState := app.createEmptyState(content, "No groups found", "Create your first group to get started!", icons.Group)
		_ = emptyState
	} else {
		for _, group := range app.groups {
			app.createEnhancedGroupCard(content, group)
		}
	}

	_ = header
	app.mainContainer.Update()
}

// Create enhanced group card with action menu
func (app *App) createEnhancedGroupCard(parent core.Widget, group Group) *core.Frame {
	card := core.NewFrame(parent)
	app.styleCard(card)
	card.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Padding.Set(core.Dp(16))
		s.Border.Style.Set(styles.BorderSolid)
		s.Border.Width.Set(core.Dp(1))
		s.Border.Color.Set(ColorGrayLight)
		s.Margin.Bottom = core.Dp(8)
	})

	// Group info section (clickable)
	infoContainer := core.NewFrame(card)
	infoContainer.Style(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(core.Dp(8))
		s.Grow.Set(1, 0)
		s.Cursor = styles.CursorPointer
	})
	infoContainer.OnClick(func(e events.Event) {
		app.showGroupDetailView(group)
	})

	// Group name and description
	nameContainer := core.NewFrame(infoContainer)
	nameContainer.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Gap.Set(core.Dp(8))
	})

	groupIcon := core.NewIcon(nameContainer).SetIcon(icons.Group)
	groupIcon.Style(func(s *styles.Style) {
		s.Color = ColorPrimary
	})

	groupName := core.NewText(nameContainer).SetText(group.Name)
	groupName.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(18)
		s.Font.Weight = styles.WeightSemiBold
		s.Color = ColorBlack
	})

	if group.Description != "" {
		groupDesc := core.NewText(infoContainer).SetText(group.Description)
		groupDesc.Style(func(s *styles.Style) {
			s.Font.Size = core.Dp(14)
			s.Color = ColorGrayDark
		})
	}

	// Stats row
	statsRow := core.NewFrame(infoContainer)
	statsRow.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(core.Dp(16))
		s.Align.Items = styles.Center
	})

	membersText := core.NewText(statsRow).SetText(fmt.Sprintf("%d members", len(group.Members)))
	membersText.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(12)
		s.Color = ColorGrayDark
	})

	// Actions menu
	actionsMenu := core.NewFrame(card)
	actionsMenu.Style(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(core.Dp(8))
	})

	// Edit button
	editBtn := core.NewButton(actionsMenu).SetIcon(icons.Edit)
	editBtn.Style(func(s *styles.Style) {
		s.Background = ColorAccent
		s.Color = ColorBlack
		s.Border.Radius = styles.BorderRadiusFull
		s.Padding.Set(core.Dp(8))
	})
	editBtn.OnClick(func(e events.Event) {
		app.showEditGroupDialog(group)
	})

	// Delete button
	deleteBtn := core.NewButton(actionsMenu).SetIcon(icons.Delete)
	deleteBtn.Style(func(s *styles.Style) {
		s.Background = ColorDanger
		s.Color = ColorWhite
		s.Border.Radius = styles.BorderRadiusFull
		s.Padding.Set(core.Dp(8))
	})
	deleteBtn.OnClick(func(e events.Event) {
		app.showDeleteGroupDialog(group)
	})

	// Invite button
	inviteBtn := core.NewButton(actionsMenu).SetIcon(icons.PersonAdd)
	inviteBtn.Style(func(s *styles.Style) {
		s.Background = ColorPrimary
		s.Color = ColorWhite
		s.Border.Radius = styles.BorderRadiusFull
		s.Padding.Set(core.Dp(8))
	})
	inviteBtn.OnClick(func(e events.Event) {
		app.showInviteToGroupDialog(group)
	})

	return card
}

// Group Detail View
func (app *App) showGroupDetailView(group Group) {
	app.mainContainer.DeleteChildren()
	app.currentView = "group_detail"

	// Header with back button
	header := app.createHeader(group.Name, true)

	// Main content
	content := core.NewFrame(app.mainContainer)
	content.Style(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Grow.Set(1, 1)
		s.Padding.Set(core.Dp(16))
		s.Gap.Set(core.Dp(16))
	})

	// Group info card
	infoCard := core.NewFrame(content)
	infoCard.Style(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Background = ColorWhite
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(core.Dp(16))
		s.Gap.Set(core.Dp(12))
	})

	// Description
	if group.Description != "" {
		descTitle := core.NewText(infoCard).SetText("Description")
		descTitle.Style(func(s *styles.Style) {
			s.Font.Weight = styles.WeightSemiBold
			s.Color = ColorGrayDark
		})
		desc := core.NewText(infoCard).SetText(group.Description)
	}

	// Members section
	membersTitle := core.NewText(content).SetText("Members")
	membersTitle.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(18)
		s.Font.Weight = styles.WeightSemiBold
	})

	if len(group.Members) == 0 {
		emptyMembers := core.NewText(content).SetText("No members in this group yet.")
		emptyMembers.Style(func(s *styles.Style) {
			s.Color = ColorGrayDark
			s.Align.Self = styles.Center
		})
	} else {
		for _, member := range group.Members {
			app.createMemberCard(content, member, group)
		}
	}

	// Collections section
	collectionsTitle := core.NewText(content).SetText("Collections")
	collectionsTitle.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(18)
		s.Font.Weight = styles.WeightSemiBold
	})

	// Filter collections for this group
	groupCollections := []Collection{}
	for _, collection := range app.collections {
		// In a real implementation, you'd filter by group membership
		groupCollections = append(groupCollections, collection)
	}

	if len(groupCollections) == 0 {
		emptyCollections := core.NewText(content).SetText("No collections in this group yet.")
		emptyCollections.Style(func(s *styles.Style) {
			s.Color = ColorGrayDark
			s.Align.Self = styles.Center
		})
	} else {
		for _, collection := range groupCollections {
			app.createCollectionCard(content, collection)
		}
	}

	_ = header
	app.mainContainer.Update()
}

// Member card for group detail view
func (app *App) createMemberCard(parent core.Widget, member User, group Group) *core.Frame {
	card := core.NewFrame(parent)
	card.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Justify.Content = styles.SpaceBetween
		s.Background = ColorWhite
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(core.Dp(12))
		s.Border.Style.Set(styles.BorderSolid)
		s.Border.Width.Set(core.Dp(1))
		s.Border.Color.Set(ColorGrayLight)
		s.Margin.Bottom = core.Dp(4)
	})

	// Member info
	infoContainer := core.NewFrame(card)
	infoContainer.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Gap.Set(core.Dp(12))
	})

	memberIcon := core.NewIcon(infoContainer).SetIcon(icons.Person)
	memberIcon.Style(func(s *styles.Style) {
		s.Color = ColorPrimary
	})

	memberDetails := core.NewFrame(infoContainer)
	memberDetails.Style(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(core.Dp(2))
	})

	memberName := core.NewText(memberDetails).SetText(member.Username)
	memberName.Style(func(s *styles.Style) {
		s.Font.Weight = styles.WeightMedium
	})

	if member.Email != "" {
		memberEmail := core.NewText(memberDetails).SetText(member.Email)
		memberEmail.Style(func(s *styles.Style) {
			s.Font.Size = core.Dp(12)
			s.Color = ColorGrayDark
		})
	}

	// Actions (only for group admins)
	actionsContainer := core.NewFrame(card)
	actionsContainer.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(core.Dp(8))
	})

	// Remove member button (only show if current user is admin)
	if app.currentUser != nil && app.currentUser.ID != member.ID {
		removeBtn := core.NewButton(actionsContainer).SetIcon(icons.PersonRemove)
		removeBtn.Style(func(s *styles.Style) {
			s.Background = ColorDanger
			s.Color = ColorWhite
			s.Border.Radius = styles.BorderRadiusFull
			s.Padding.Set(core.Dp(6))
		})
		removeBtn.OnClick(func(e events.Event) {
			app.showRemoveMemberDialog(member, group)
		})
	}

	return card
}

// Dialog functions
func (app *App) showCreateGroupDialog() {
	// Create overlay
	overlay := app.createOverlay()

	// Dialog container
	dialog := core.NewFrame(overlay)
	dialog.Style(func(s *styles.Style) {
		s.Background = ColorWhite
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(core.Dp(24))
		s.Gap.Set(core.Dp(16))
		s.Direction = styles.Column
		s.Min.X.Set(core.Dp(400))
		s.Max.X.Set(core.Dp(500))
	})

	// Title
	title := core.NewText(dialog).SetText("Create New Group")
	title.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(20)
		s.Font.Weight = styles.WeightSemiBold
	})

	// Form fields
	nameField := core.NewTextField(dialog)
	nameField.SetText("").SetPlaceholder("Group name")
	nameField.Style(func(s *styles.Style) {
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(core.Dp(12))
	})

	descField := core.NewTextField(dialog)
	descField.SetText("").SetPlaceholder("Description (optional)")
	descField.Style(func(s *styles.Style) {
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(core.Dp(12))
	})

	// Buttons
	buttonRow := core.NewFrame(dialog)
	buttonRow.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(core.Dp(12))
		s.Justify.Content = styles.End
	})

	cancelBtn := core.NewButton(buttonRow).SetText("Cancel")
	cancelBtn.Style(func(s *styles.Style) {
		s.Background = color.RGBA{R: 240, G: 240, B: 240, A: 255}
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(core.Dp(8), core.Dp(16))
	})
	cancelBtn.OnClick(func(e events.Event) {
		app.hideOverlay()
	})

	createBtn := core.NewButton(buttonRow).SetText("Create Group")
	createBtn.Style(func(s *styles.Style) {
		s.Background = ColorPrimary
		s.Color = ColorWhite
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(core.Dp(8), core.Dp(16))
	})
	createBtn.OnClick(func(e events.Event) {
		app.handleCreateGroup(nameField.Text(), descField.Text())
	})

	app.showOverlay(overlay)
}

func (app *App) showEditGroupDialog(group Group) {
	// Similar to create dialog but with pre-filled values
	overlay := app.createOverlay()

	dialog := core.NewFrame(overlay)
	dialog.Style(func(s *styles.Style) {
		s.Background = ColorWhite
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(core.Dp(24))
		s.Gap.Set(core.Dp(16))
		s.Direction = styles.Column
		s.Min.X.Set(core.Dp(400))
		s.Max.X.Set(core.Dp(500))
	})

	title := core.NewText(dialog).SetText("Edit Group")
	title.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(20)
		s.Font.Weight = styles.WeightSemiBold
	})

	nameField := core.NewTextField(dialog)
	nameField.SetText(group.Name).SetPlaceholder("Group name")

	descField := core.NewTextField(dialog)
	descField.SetText(group.Description).SetPlaceholder("Description (optional)")

	buttonRow := core.NewFrame(dialog)
	buttonRow.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(core.Dp(12))
		s.Justify.Content = styles.End
	})

	cancelBtn := core.NewButton(buttonRow).SetText("Cancel")
	cancelBtn.OnClick(func(e events.Event) {
		app.hideOverlay()
	})

	saveBtn := core.NewButton(buttonRow).SetText("Save Changes")
	saveBtn.Style(func(s *styles.Style) {
		s.Background = ColorPrimary
		s.Color = ColorWhite
	})
	saveBtn.OnClick(func(e events.Event) {
		app.handleEditGroup(group.ID, nameField.Text(), descField.Text())
	})

	app.showOverlay(overlay)
}

func (app *App) showDeleteGroupDialog(group Group) {
	overlay := app.createOverlay()

	dialog := core.NewFrame(overlay)
	dialog.Style(func(s *styles.Style) {
		s.Background = ColorWhite
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(core.Dp(24))
		s.Gap.Set(core.Dp(16))
		s.Direction = styles.Column
		s.Min.X.Set(core.Dp(400))
	})

	title := core.NewText(dialog).SetText("Delete Group")
	title.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(20)
		s.Font.Weight = styles.WeightSemiBold
		s.Color = ColorDanger
	})

	message := core.NewText(dialog).SetText(fmt.Sprintf("Are you sure you want to delete \"%s\"? This action cannot be undone.", group.Name))
	message.Style(func(s *styles.Style) {
		s.Color = ColorGrayDark
	})

	buttonRow := core.NewFrame(dialog)
	buttonRow.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(core.Dp(12))
		s.Justify.Content = styles.End
	})

	cancelBtn := core.NewButton(buttonRow).SetText("Cancel")
	cancelBtn.OnClick(func(e events.Event) {
		app.hideOverlay()
	})

	deleteBtn := core.NewButton(buttonRow).SetText("Delete")
	deleteBtn.Style(func(s *styles.Style) {
		s.Background = ColorDanger
		s.Color = ColorWhite
	})
	deleteBtn.OnClick(func(e events.Event) {
		app.handleDeleteGroup(group.ID)
	})

	app.showOverlay(overlay)
}

// Helper functions for overlay management
func (app *App) createOverlay() *core.Frame {
	overlay := core.NewFrame(app.mainContainer)
	overlay.Style(func(s *styles.Style) {
		s.Position = styles.PositionAbsolute
		s.Top = core.Dp(0)
		s.Left = core.Dp(0)
		s.Right = core.Dp(0)
		s.Bottom = core.Dp(0)
		s.Background = ColorOverlay // Semi-transparent black
		s.Display = styles.Flex
		s.Align.Items = styles.Center
		s.Justify.Content = styles.Center
		s.Z = 1000 // High z-index
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
	emptyState.Style(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Align.Items = styles.Center
		s.Justify.Content = styles.Center
		s.Gap.Set(core.Dp(16))
		s.Padding.Set(core.Dp(32))
		s.Margin.Top = core.Dp(32)
	})

	emptyIcon := core.NewIcon(emptyState).SetIcon(icon)
	emptyIcon.Style(func(s *styles.Style) {
		s.Color = ColorGray
		s.Font.Size = core.Dp(48)
	})

	emptyTitle := core.NewText(emptyState).SetText(title)
	emptyTitle.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(18)
		s.Font.Weight = styles.WeightSemiBold
		s.Color = ColorGrayDark
	})

	emptyMessage := core.NewText(emptyState).SetText(message)
	emptyMessage.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(14)
		s.Color = ColorGrayDark
		s.Text.Align = styles.Center
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
	overlay := app.createOverlay()

	dialog := core.NewFrame(overlay)
	dialog.Style(func(s *styles.Style) {
		s.Background = ColorWhite
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(core.Dp(24))
		s.Gap.Set(core.Dp(16))
		s.Direction = styles.Column
		s.Min.X.Set(core.Dp(400))
	})

	title := core.NewText(dialog).SetText("Join Group")
	title.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(20)
		s.Font.Weight = styles.WeightSemiBold
	})

	inviteField := core.NewTextField(dialog)
	inviteField.SetText("").SetPlaceholder("Invitation code")

	buttonRow := core.NewFrame(dialog)
	buttonRow.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(core.Dp(12))
		s.Justify.Content = styles.End
	})

	cancelBtn := core.NewButton(buttonRow).SetText("Cancel")
	cancelBtn.OnClick(func(e events.Event) {
		app.hideOverlay()
	})

	joinBtn := core.NewButton(buttonRow).SetText("Join Group")
	joinBtn.Style(func(s *styles.Style) {
		s.Background = ColorPrimary
		s.Color = ColorWhite
	})
	joinBtn.OnClick(func(e events.Event) {
		app.handleJoinGroup(inviteField.Text())
	})

	app.showOverlay(overlay)
}

func (app *App) showInviteToGroupDialog(group Group) {
	overlay := app.createOverlay()

	dialog := core.NewFrame(overlay)
	dialog.Style(func(s *styles.Style) {
		s.Background = ColorWhite
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(core.Dp(24))
		s.Gap.Set(core.Dp(16))
		s.Direction = styles.Column
		s.Min.X.Set(core.Dp(400))
	})

	title := core.NewText(dialog).SetText("Invite to Group")
	title.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(20)
		s.Font.Weight = styles.WeightSemiBold
	})

	message := core.NewText(dialog).SetText("Share this invitation code:")
	inviteCode := core.NewText(dialog).SetText("ABC123XYZ") // Would be generated
	inviteCode.Style(func(s *styles.Style) {
		s.Font.Family = "monospace"
		s.Background = color.RGBA{R: 240, G: 240, B: 240, A: 255}
		s.Padding.Set(core.Dp(8))
		s.Border.Radius = styles.BorderRadiusMedium
	})

	buttonRow := core.NewFrame(dialog)
	buttonRow.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(core.Dp(12))
		s.Justify.Content = styles.End
	})

	closeBtn := core.NewButton(buttonRow).SetText("Close")
	closeBtn.OnClick(func(e events.Event) {
		app.hideOverlay()
	})

	app.showOverlay(overlay)
}

func (app *App) showRemoveMemberDialog(member User, group Group) {
	overlay := app.createOverlay()

	dialog := core.NewFrame(overlay)
	dialog.Style(func(s *styles.Style) {
		s.Background = ColorWhite
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(core.Dp(24))
		s.Gap.Set(core.Dp(16))
		s.Direction = styles.Column
		s.Min.X.Set(core.Dp(400))
	})

	title := core.NewText(dialog).SetText("Remove Member")
	title.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(20)
		s.Font.Weight = styles.WeightSemiBold
		s.Color = ColorDanger
	})

	message := core.NewText(dialog).SetText(fmt.Sprintf("Remove \"%s\" from \"%s\"?", member.Username, group.Name))

	buttonRow := core.NewFrame(dialog)
	buttonRow.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(core.Dp(12))
		s.Justify.Content = styles.End
	})

	cancelBtn := core.NewButton(buttonRow).SetText("Cancel")
	cancelBtn.OnClick(func(e events.Event) {
		app.hideOverlay()
	})

	removeBtn := core.NewButton(buttonRow).SetText("Remove")
	removeBtn.Style(func(s *styles.Style) {
		s.Background = ColorDanger
		s.Color = ColorWhite
	})
	removeBtn.OnClick(func(e events.Event) {
		app.handleRemoveMember(member.ID, group.ID)
	})

	app.showOverlay(overlay)
}

func (app *App) handleJoinGroup(inviteCode string) {
	fmt.Printf("Joining group with code: %s\n", inviteCode)
	app.hideOverlay()
}

func (app *App) handleRemoveMember(userID, groupID string) {
	fmt.Printf("Removing user %s from group %s\n", userID, groupID)
	app.hideOverlay()
}