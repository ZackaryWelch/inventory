package app

import (
	"fmt"
	"image/color"
	"strings"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/cursors"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/text/rich"
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
	content.Styler(StyleContentColumn)

	// Action buttons row
	actionsRow := core.NewFrame(content)
	actionsRow.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(units.Dp(12))
		s.Justify.Content = styles.End
	})

	// Create group button
	createBtn := core.NewButton(actionsRow).SetText("Create Group").SetIcon(icons.Add)
	createBtn.Styler(StyleButtonPrimary)
	createBtn.Styler(StyleButtonMd)
	createBtn.OnClick(func(e events.Event) {
		app.showCreateGroupDialog()
	})

	// Join group button
	joinBtn := core.NewButton(actionsRow).SetText("Join Group").SetIcon(icons.PersonAdd)
	joinBtn.Styler(StyleButtonAccent)
	joinBtn.Styler(StyleButtonMd)
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
	card.Styler(StyleCard)
	card.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Padding.Set(units.Dp(16))
		s.Border.Style.Set(styles.BorderSolid)
		s.Border.Width.Set(units.Dp(1))
		s.Border.Color.Set(colors.Uniform(ColorGrayLight))
		s.Margin.Bottom = units.Dp(8)
	})

	// Group info section (clickable)
	infoContainer := core.NewFrame(card)
	infoContainer.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(units.Dp(8))
		s.Grow.Set(1, 0)
		s.Cursor = cursors.Pointer
	})
	infoContainer.OnClick(func(e events.Event) {
		app.showGroupDetailView(group)
	})

	// Group name and description
	nameContainer := core.NewFrame(infoContainer)
	nameContainer.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Gap.Set(units.Dp(8))
	})

	groupIcon := core.NewIcon(nameContainer).SetIcon(icons.Group)
	groupIcon.Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(ColorPrimary)
	})

	groupName := core.NewText(nameContainer).SetText(group.Name)
	groupName.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(18)
		s.Font.Weight = WeightSemiBold
		s.Color = colors.Uniform(ColorBlack)
	})

	if group.Description != "" {
		groupDesc := core.NewText(infoContainer).SetText(group.Description)
		groupDesc.Styler(func(s *styles.Style) {
			s.Font.Size = units.Dp(14)
			s.Color = colors.Uniform(ColorGrayDark)
		})
	}

	// Stats row
	statsRow := core.NewFrame(infoContainer)
	statsRow.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(units.Dp(16))
		s.Align.Items = styles.Center
	})

	membersText := core.NewText(statsRow).SetText(fmt.Sprintf("%d members", len(group.Members)))
	membersText.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(12)
		s.Color = colors.Uniform(ColorGrayDark)
	})

	// Actions menu
	actionsMenu := core.NewFrame(card)
	actionsMenu.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(units.Dp(8))
	})

	// Edit button
	editBtn := core.NewButton(actionsMenu).SetIcon(icons.Edit)
	editBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorAccent)
		s.Color = colors.Uniform(ColorBlack)
		s.Border.Radius = styles.BorderRadiusFull
		s.Padding.Set(units.Dp(8))
	})
	editBtn.OnClick(func(e events.Event) {
		app.showEditGroupDialog(group)
	})

	// Delete button
	deleteBtn := core.NewButton(actionsMenu).SetIcon(icons.Delete)
	deleteBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorDanger)
		s.Color = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusFull
		s.Padding.Set(units.Dp(8))
	})
	deleteBtn.OnClick(func(e events.Event) {
		app.showDeleteGroupDialog(group)
	})

	// Invite button
	inviteBtn := core.NewButton(actionsMenu).SetIcon(icons.PersonAdd)
	inviteBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorPrimary)
		s.Color = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusFull
		s.Padding.Set(units.Dp(8))
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
	content.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Grow.Set(1, 1)
		s.Padding.Set(units.Dp(16))
		s.Gap.Set(units.Dp(16))
	})

	// Group info card
	infoCard := core.NewFrame(content)
	infoCard.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Background = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(16))
		s.Gap.Set(units.Dp(12))
	})

	// Description
	if group.Description != "" {
		descTitle := core.NewText(infoCard).SetText("Description")
		descTitle.Styler(func(s *styles.Style) {
			s.Font.Weight = WeightSemiBold
			s.Color = colors.Uniform(ColorGrayDark)
		})
		desc := core.NewText(infoCard).SetText(group.Description)
		desc.SetTooltip("Group description")
	}

	// Members section
	membersTitle := core.NewText(content).SetText("Members")
	membersTitle.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(18)
		s.Font.Weight = WeightSemiBold
	})

	if len(group.Members) == 0 {
		emptyMembers := core.NewText(content).SetText("No members in this group yet.")
		emptyMembers.Styler(func(s *styles.Style) {
			s.Color = colors.Uniform(ColorGrayDark)
			s.Align.Self = styles.Center
		})
	} else {
		for _, member := range group.Members {
			app.createMemberCard(content, member, group)
		}
	}

	// Collections section
	collectionsTitle := core.NewText(content).SetText("Collections")
	collectionsTitle.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(18)
		s.Font.Weight = WeightSemiBold
	})

	// Filter collections for this group
	groupCollections := []Collection{}
	for _, collection := range app.collections {
		// In a real implementation, you'd filter by group membership
		groupCollections = append(groupCollections, collection)
	}

	if len(groupCollections) == 0 {
		emptyCollections := core.NewText(content).SetText("No collections in this group yet.")
		emptyCollections.Styler(func(s *styles.Style) {
			s.Color = colors.Uniform(ColorGrayDark)
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
	card.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Justify.Content = styles.SpaceBetween
		s.Background = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(units.Dp(12))
		s.Border.Style.Set(styles.BorderSolid)
		s.Border.Width.Set(units.Dp(1))
		s.Border.Color.Set(colors.Uniform(ColorGrayLight))
		s.Margin.Bottom = units.Dp(4)
	})

	// Member info
	infoContainer := core.NewFrame(card)
	infoContainer.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Gap.Set(units.Dp(12))
	})

	memberIcon := core.NewIcon(infoContainer).SetIcon(icons.Person)
	memberIcon.Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(ColorPrimary)
	})

	memberDetails := core.NewFrame(infoContainer)
	memberDetails.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(units.Dp(2))
	})

	memberName := core.NewText(memberDetails).SetText(member.Username)
	memberName.Styler(func(s *styles.Style) {
		s.Font.Weight = WeightMedium
	})

	if member.Email != "" {
		memberEmail := core.NewText(memberDetails).SetText(member.Email)
		memberEmail.Styler(func(s *styles.Style) {
			s.Font.Size = units.Dp(12)
			s.Color = colors.Uniform(ColorGrayDark)
		})
	}

	// Actions (only for group admins)
	actionsContainer := core.NewFrame(card)
	actionsContainer.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(units.Dp(8))
	})

	// Remove member button (only show if current user is admin)
	if app.currentUser != nil && app.currentUser.ID != member.ID {
		removeBtn := core.NewButton(actionsContainer).SetIcon(icons.PersonRemove)
		removeBtn.Styler(func(s *styles.Style) {
			s.Background = colors.Uniform(ColorDanger)
			s.Color = colors.Uniform(ColorWhite)
			s.Border.Radius = styles.BorderRadiusFull
			s.Padding.Set(units.Dp(6))
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
	dialog.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(24))
		s.Gap.Set(units.Dp(16))
		s.Direction = styles.Column
		s.Min.X.Set(400, units.UnitDp)
		s.Max.X.Set(500, units.UnitDp)
	})

	// Title
	title := core.NewText(dialog).SetText("Create New Group")
	title.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(20)
		s.Font.Weight = WeightSemiBold
	})

	// Form fields
	nameField := core.NewTextField(dialog)
	nameField.SetText("").SetPlaceholder("Group name")
	nameField.Styler(func(s *styles.Style) {
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(units.Dp(12))
	})

	descField := core.NewTextField(dialog)
	descField.SetText("").SetPlaceholder("Description (optional)")
	descField.Styler(func(s *styles.Style) {
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(units.Dp(12))
	})

	// Buttons
	buttonRow := core.NewFrame(dialog)
	buttonRow.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(units.Dp(12))
		s.Justify.Content = styles.End
	})

	cancelBtn := core.NewButton(buttonRow).SetText("Cancel")
	cancelBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(color.RGBA{R: 240, G: 240, B: 240, A: 255})
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(units.Dp(8), units.Dp(16))
	})
	cancelBtn.OnClick(func(e events.Event) {
		app.hideOverlay()
	})

	createBtn := core.NewButton(buttonRow).SetText("Create Group")
	createBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorPrimary)
		s.Color = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(units.Dp(8), units.Dp(16))
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
	dialog.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(24))
		s.Gap.Set(units.Dp(16))
		s.Direction = styles.Column
		s.Min.X.Set(400, units.UnitDp)
		s.Max.X.Set(500, units.UnitDp)
	})

	title := core.NewText(dialog).SetText("Edit Group")
	title.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(20)
		s.Font.Weight = WeightSemiBold
	})

	nameField := core.NewTextField(dialog)
	nameField.SetText(group.Name).SetPlaceholder("Group name")

	descField := core.NewTextField(dialog)
	descField.SetText(group.Description).SetPlaceholder("Description (optional)")

	buttonRow := core.NewFrame(dialog)
	buttonRow.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(units.Dp(12))
		s.Justify.Content = styles.End
	})

	cancelBtn := core.NewButton(buttonRow).SetText("Cancel")
	cancelBtn.OnClick(func(e events.Event) {
		app.hideOverlay()
	})

	saveBtn := core.NewButton(buttonRow).SetText("Save Changes")
	saveBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorPrimary)
		s.Color = colors.Uniform(ColorWhite)
	})
	saveBtn.OnClick(func(e events.Event) {
		app.handleEditGroup(group.ID, nameField.Text(), descField.Text())
	})

	app.showOverlay(overlay)
}

func (app *App) showDeleteGroupDialog(group Group) {
	overlay := app.createOverlay()

	dialog := core.NewFrame(overlay)
	dialog.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(24))
		s.Gap.Set(units.Dp(16))
		s.Direction = styles.Column
		s.Min.X.Set(400, units.UnitDp)
	})

	title := core.NewText(dialog).SetText("Delete Group")
	title.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(20)
		s.Font.Weight = WeightSemiBold
		s.Color = colors.Uniform(ColorDanger)
	})

	message := core.NewText(dialog).SetText(fmt.Sprintf("Are you sure you want to delete \"%s\"? This action cannot be undone.", group.Name))
	message.Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(ColorGrayDark)
	})

	buttonRow := core.NewFrame(dialog)
	buttonRow.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(units.Dp(12))
		s.Justify.Content = styles.End
	})

	cancelBtn := core.NewButton(buttonRow).SetText("Cancel")
	cancelBtn.OnClick(func(e events.Event) {
		app.hideOverlay()
	})

	deleteBtn := core.NewButton(buttonRow).SetText("Delete")
	deleteBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorDanger)
		s.Color = colors.Uniform(ColorWhite)
	})
	deleteBtn.OnClick(func(e events.Event) {
		app.handleDeleteGroup(group.ID)
	})

	app.showOverlay(overlay)
}

// Helper functions for overlay management
func (app *App) createOverlay() *core.Frame {
	overlay := core.NewFrame(app.mainContainer)
	overlay.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorOverlay) // Semi-transparent black
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
		s.Color = colors.Uniform(ColorGray)
		s.Font.Size = units.Dp(48)
	})

	emptyTitle := core.NewText(emptyState).SetText(title)
	emptyTitle.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(18)
		s.Font.Weight = WeightSemiBold
		s.Color = colors.Uniform(ColorGrayDark)
	})

	emptyMessage := core.NewText(emptyState).SetText(message)
	emptyMessage.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(14)
		s.Color = colors.Uniform(ColorGrayDark)
		s.Text.Align = AlignCenter
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
	dialog.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(24))
		s.Gap.Set(units.Dp(16))
		s.Direction = styles.Column
		s.Min.X.Set(400, units.UnitDp)
	})

	title := core.NewText(dialog).SetText("Join Group")
	title.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(20)
		s.Font.Weight = WeightSemiBold
	})

	inviteField := core.NewTextField(dialog)
	inviteField.SetText("").SetPlaceholder("Invitation code")

	buttonRow := core.NewFrame(dialog)
	buttonRow.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(units.Dp(12))
		s.Justify.Content = styles.End
	})

	cancelBtn := core.NewButton(buttonRow).SetText("Cancel")
	cancelBtn.OnClick(func(e events.Event) {
		app.hideOverlay()
	})

	joinBtn := core.NewButton(buttonRow).SetText("Join Group")
	joinBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorPrimary)
		s.Color = colors.Uniform(ColorWhite)
	})
	joinBtn.OnClick(func(e events.Event) {
		app.handleJoinGroup(inviteField.Text())
	})

	app.showOverlay(overlay)
}

func (app *App) showInviteToGroupDialog(group Group) {
	overlay := app.createOverlay()

	dialog := core.NewFrame(overlay)
	dialog.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(24))
		s.Gap.Set(units.Dp(16))
		s.Direction = styles.Column
		s.Min.X.Set(400, units.UnitDp)
	})

	title := core.NewText(dialog).SetText("Invite to Group")
	title.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(20)
		s.Font.Weight = WeightSemiBold
	})

	message := core.NewText(dialog).SetText("Share this invitation code:")
	message.SetTooltip("Share this code with others to invite them")
	inviteCode := core.NewText(dialog).SetText("ABC123XYZ") // Would be generated
	inviteCode.Styler(func(s *styles.Style) {
		s.Font.Family = rich.Monospace
		s.Background = colors.Uniform(color.RGBA{R: 240, G: 240, B: 240, A: 255})
		s.Padding.Set(units.Dp(8))
		s.Border.Radius = styles.BorderRadiusMedium
	})

	buttonRow := core.NewFrame(dialog)
	buttonRow.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(units.Dp(12))
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
	dialog.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(24))
		s.Gap.Set(units.Dp(16))
		s.Direction = styles.Column
		s.Min.X.Set(400, units.UnitDp)
	})

	title := core.NewText(dialog).SetText("Remove Member")
	title.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(20)
		s.Font.Weight = WeightSemiBold
		s.Color = colors.Uniform(ColorDanger)
	})

	message := core.NewText(dialog).SetText(fmt.Sprintf("Remove \"%s\" from \"%s\"?", member.Username, group.Name))
	message.SetTooltip("Confirm member removal")

	buttonRow := core.NewFrame(dialog)
	buttonRow.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(units.Dp(12))
		s.Justify.Content = styles.End
	})

	cancelBtn := core.NewButton(buttonRow).SetText("Cancel")
	cancelBtn.OnClick(func(e events.Event) {
		app.hideOverlay()
	})

	removeBtn := core.NewButton(buttonRow).SetText("Remove")
	removeBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorDanger)
		s.Color = colors.Uniform(ColorWhite)
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