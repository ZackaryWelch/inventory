//go:build js && wasm

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
	"cogentcore.org/core/styles/sides"
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
	actionRow.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Justify.Content = styles.End // justify-end
		s.Align.Items = styles.Center   // items-center
		s.Min.Y.Set(48, units.UnitDp)   // h-12
		s.Min.X.Set(100, units.UnitPw)  // w-full (parent width)
		s.Margin.Bottom = units.Dp(appstyles.Spacing2)
	})

	// Create Group button (ghost icon button like React)
	createGroupBtn := core.NewButton(actionRow).SetIcon(icons.Add)
	createGroupBtn.Styler(func(s *styles.Style) {
		s.Background = nil // ghost variant
		s.Color = colors.Uniform(appstyles.ColorGrayDark)
		s.Padding.Set(units.Dp(8))
		s.Min.X.Set(48, units.UnitDp)
		s.Min.Y.Set(48, units.UnitDp)
	})
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
	container.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(units.Dp(8))
		s.Margin.Bottom = units.Dp(appstyles.Spacing2)
	})

	// Group name ABOVE the card (React: text-lg font-semibold mb-2)
	groupName := core.NewText(container).SetText(group.Name)
	groupName.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(appstyles.FontSizeLG)
		s.Font.Weight = appstyles.WeightSemiBold
		s.Color = colors.Uniform(appstyles.ColorBlack)
	})

	// Card - single horizontal row (React: className="flex justify-between items-center p-4")
	card := components.Card(container, components.CardProps{})
	card.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Justify.Content = styles.SpaceBetween
		s.Align.Items = styles.Center
		s.Padding.Set(units.Dp(appstyles.Spacing4))
		s.Gap.Set(units.Dp(appstyles.Spacing4))
	})

	// LEFT SECTION: Icon + Container Count (React: flex gap-2 items-center)
	leftSection := core.NewFrame(card)
	leftSection.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Gap.Set(units.Dp(8))
		s.Cursor = cursors.Pointer
	})
	leftSection.OnClick(func(e events.Event) {
		app.showGroupDetailView(group)
	})

	// Icon (cheese emoji in React - using folder icon as placeholder)
	icon := core.NewIcon(leftSection).SetIcon(icons.Folder)
	icon.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(24)
		s.Color = colors.Uniform(appstyles.ColorAccent) // Yellow like cheese
	})

	// Container count
	containerCount := core.NewText(leftSection).SetText("0")
	containerCount.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(appstyles.FontSizeBase)
		s.Color = colors.Uniform(appstyles.ColorBlack)
	})

	// RIGHT SECTION: User Avatars + User Count + Menu (React: flex gap-2 items-center)
	rightSection := core.NewFrame(card)
	rightSection.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Gap.Set(units.Dp(8))
	})

	// User avatars (3 gray circles in React)
	avatarsContainer := core.NewFrame(rightSection)
	avatarsContainer.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(units.Dp(4))
	})

	// Show up to 3 avatar circles
	numAvatars := len(group.Members)
	if numAvatars > 3 {
		numAvatars = 3
	}
	for i := 0; i < numAvatars; i++ {
		avatar := core.NewIcon(avatarsContainer).SetIcon(icons.Person)
		avatar.Styler(func(s *styles.Style) {
			s.Font.Size = units.Dp(20)
			s.Color = colors.Uniform(appstyles.ColorGray)
			s.Background = colors.Uniform(appstyles.ColorGrayLight)
			s.Border.Radius = sides.NewValues(units.Dp(9999)) // Circular
			s.Padding.Set(units.Dp(4))
		})
	}

	// User count (×N)
	userCount := core.NewText(rightSection).SetText(fmt.Sprintf("×%d", len(group.Members)))
	userCount.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(appstyles.FontSizeBase)
		s.Color = colors.Uniform(appstyles.ColorGrayDark)
	})

	// Three-dot menu button
	menuButton := core.NewButton(rightSection).SetIcon(icons.MoreVert)
	menuButton.Styler(func(s *styles.Style) {
		s.Background = nil
		s.Border.Width.Set(units.Dp(1))
		s.Border.Color.Set(colors.Uniform(appstyles.ColorGray))
		s.Border.Radius = sides.NewValues(units.Dp(appstyles.RadiusMD))
		s.Padding.Set(units.Dp(8))
		s.Color = colors.Uniform(appstyles.ColorGray)
	})
	menuButton.OnClick(func(e events.Event) {
		app.showEditGroupDialog(group)
	})

	return container
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
		Title:       member.Name,
		Description: member.Email,
		Actions:     actions,
	})
}

// Removed showGroupActionsMenu - button now opens dialog directly

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

// NOTE: Overlay management removed - using Cogent Core's built-in dialog system
// See ui_helpers.go showDialog() which uses d.RunDialog(app.body)

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
	
	// Dialog closes automatically
	app.fetchGroups() // Refresh the list
	app.showEnhancedGroupsView() // Refresh the view
}

func (app *App) handleEditGroup(groupID, name, description string) {
	fmt.Printf("Editing group %s: %s - %s\n", groupID, name, description)
	
	// Dialog closes automatically
	app.fetchGroups()
	app.showEnhancedGroupsView()
}

func (app *App) handleDeleteGroup(groupID string) {
	fmt.Printf("Deleting group: %s\n", groupID)
	
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
		Message:          fmt.Sprintf("Remove \"%s\" from \"%s\"?", member.Name, group.Name),
		SubmitButtonText: "Remove",
		SubmitButtonStyle: appstyles.StyleButtonDanger,
		OnSubmit: func() {
			app.handleRemoveMember(member.ID, group.ID)
		},
	})
}

func (app *App) handleJoinGroup(inviteCode string) {
	fmt.Printf("Joining group with code: %s\n", inviteCode)
	// Dialog closes automatically
}

func (app *App) handleRemoveMember(userID, groupID string) {
	fmt.Printf("Removing user %s from group %s\n", userID, groupID)
	// Dialog closes automatically
}