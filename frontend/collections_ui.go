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

// Enhanced Collections View with full CRUD operations
func (app *App) showEnhancedCollectionsView() {
	app.mainContainer.DeleteChildren()
	app.currentView = "collections"

	// Header with back button
	header := app.createHeader("Collections", true)

	// Refresh collections data
	if err := app.fetchCollections(); err != nil {
		fmt.Printf("Error fetching collections: %v\n", err)
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

	// Create collection button
	createBtn := core.NewButton(actionsRow).SetText("Create Collection").SetIcon(icons.Add)
	createBtn.Style(func(s *styles.Style) {
		s.Background = ColorAccent // var(--color-accent)
		s.Color = ColorBlack
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(core.Dp(12), core.Dp(16))
		s.Gap.Set(core.Dp(8))
	})
	createBtn.OnClick(func(e events.Event) {
		app.showCreateCollectionDialog()
	})

	// Import collection button
	importBtn := core.NewButton(actionsRow).SetText("Import").SetIcon(icons.Upload)
	importBtn.Style(func(s *styles.Style) {
		s.Background = ColorPrimary
		s.Color = ColorWhite
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(core.Dp(12), core.Dp(16))
		s.Gap.Set(core.Dp(8))
	})
	importBtn.OnClick(func(e events.Event) {
		app.showImportDialog()
	})

	// Collections grid
	if len(app.collections) == 0 {
		emptyState := app.createEmptyState(content, "No collections found", "Create your first collection to start managing your inventory!", icons.FolderOpen)
		_ = emptyState
	} else {
		// Collections grid
		collectionsGrid := core.NewFrame(content)
		collectionsGrid.Style(func(s *styles.Style) {
			s.Direction = styles.Row
			s.Wrap = true
			s.Gap.Set(core.Dp(16))
		})

		for _, collection := range app.collections {
			app.createEnhancedCollectionCard(collectionsGrid, collection)
		}
	}

	_ = header
	app.mainContainer.Update()
}

// Create enhanced collection card with actions
func (app *App) createEnhancedCollectionCard(parent core.Widget, collection Collection) *core.Frame {
	card := core.NewFrame(parent)
	card.Style(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Background = ColorWhite
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(core.Dp(16))
		s.Border.Style.Set(styles.BorderSolid)
		s.Border.Width.Set(core.Dp(1))
		s.Border.Color.Set(ColorGrayLight)
		s.Min.X.Set(core.Dp(280))
		s.Max.X.Set(core.Dp(320))
		s.Gap.Set(core.Dp(12))
		s.Cursor = styles.CursorPointer
	})

	// Header with icon and actions
	cardHeader := core.NewFrame(card)
	cardHeader.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Justify.Content = styles.SpaceBetween
	})

	// Icon and title section
	titleSection := core.NewFrame(cardHeader)
	titleSection.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Gap.Set(core.Dp(8))
		s.Grow.Set(1, 0)
	})
	titleSection.OnClick(func(e events.Event) {
		app.showCollectionDetailView(collection)
	})

	// Collection type icon
	typeIcon := app.getCollectionTypeIcon(collection.ObjectType)
	collectionIcon := core.NewIcon(titleSection).SetIcon(typeIcon)
	collectionIcon.Style(func(s *styles.Style) {
		s.Color = app.getCollectionTypeColor(collection.ObjectType)
	})

	titleContainer := core.NewFrame(titleSection)
	titleContainer.Style(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(core.Dp(2))
	})

	collectionName := core.NewText(titleContainer).SetText(collection.Name)
	collectionName.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(16)
		s.Font.Weight = styles.WeightSemiBold
		s.Color = ColorBlack
	})

	objectType := core.NewText(titleContainer).SetText(strings.Title(collection.ObjectType))
	objectType.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(12)
		s.Color = ColorGrayDark
	})

	// Actions menu
	actionsMenu := core.NewFrame(cardHeader)
	actionsMenu.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(core.Dp(4))
	})

	// Edit button
	editBtn := core.NewButton(actionsMenu).SetIcon(icons.Edit)
	editBtn.Style(func(s *styles.Style) {
		s.Background = ColorAccent
		s.Color = ColorBlack
		s.Border.Radius = styles.BorderRadiusFull
		s.Padding.Set(core.Dp(6))
	})
	editBtn.OnClick(func(e events.Event) {
		app.showEditCollectionDialog(collection)
	})

	// Delete button
	deleteBtn := core.NewButton(actionsMenu).SetIcon(icons.Delete)
	deleteBtn.Style(func(s *styles.Style) {
		s.Background = ColorDanger
		s.Color = ColorWhite
		s.Border.Radius = styles.BorderRadiusFull
		s.Padding.Set(core.Dp(6))
	})
	deleteBtn.OnClick(func(e events.Event) {
		app.showDeleteCollectionDialog(collection)
	})

	// Description
	if collection.Description != "" {
		desc := core.NewText(card).SetText(collection.Description)
		desc.Style(func(s *styles.Style) {
			s.Font.Size = core.Dp(14)
			s.Color = ColorGrayDark
			s.Text.Align = styles.Start
		})
	}

	// Stats section
	statsContainer := core.NewFrame(card)
	statsContainer.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(core.Dp(16))
		s.Justify.Content = styles.SpaceBetween
		s.Background = core.RGB(248, 248, 248)
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(core.Dp(12))
	})

	// Containers count
	containersStats := core.NewFrame(statsContainer)
	containersStats.Style(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Align.Items = styles.Center
		s.Gap.Set(core.Dp(2))
	})

	containersCount := core.NewText(containersStats).SetText("0") // Would be calculated from actual data
	containersCount.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(18)
		s.Font.Weight = styles.WeightBold
		s.Color = ColorPrimary
	})

	containersLabel := core.NewText(containersStats).SetText("Containers")
	containersLabel.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(10)
		s.Color = ColorGrayDark
	})

	// Objects count
	objectsStats := core.NewFrame(statsContainer)
	objectsStats.Style(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Align.Items = styles.Center
		s.Gap.Set(core.Dp(2))
	})

	objectsCount := core.NewText(objectsStats).SetText("0") // Would be calculated from actual data
	objectsCount.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(18)
		s.Font.Weight = styles.WeightBold
		s.Color = ColorAccent
	})

	objectsLabel := core.NewText(objectsStats).SetText("Objects")
	objectsLabel.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(10)
		s.Color = ColorGrayDark
	})

	return card
}

// Collection Detail View
func (app *App) showCollectionDetailView(collection Collection) {
	app.mainContainer.DeleteChildren()
	app.currentView = "collection_detail"

	// Header with back button
	header := app.createHeader(collection.Name, true)

	// Main content
	content := core.NewFrame(app.mainContainer)
	content.Style(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Grow.Set(1, 1)
		s.Padding.Set(core.Dp(16))
		s.Gap.Set(core.Dp(16))
	})

	// Collection info card
	infoCard := core.NewFrame(content)
	infoCard.Style(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Background = ColorWhite
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(core.Dp(16))
		s.Gap.Set(core.Dp(12))
	})

	// Collection type
	typeRow := core.NewFrame(infoCard)
	typeRow.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Gap.Set(core.Dp(8))
	})

	typeIcon := app.getCollectionTypeIcon(collection.ObjectType)
	icon := core.NewIcon(typeRow).SetIcon(typeIcon)
	icon.Style(func(s *styles.Style) {
		s.Color = app.getCollectionTypeColor(collection.ObjectType)
	})

	typeText := core.NewText(typeRow).SetText(fmt.Sprintf("Type: %s", strings.Title(collection.ObjectType)))
	typeText.Style(func(s *styles.Style) {
		s.Font.Weight = styles.WeightMedium
	})

	// Description
	if collection.Description != "" {
		descTitle := core.NewText(infoCard).SetText("Description")
		descTitle.Style(func(s *styles.Style) {
			s.Font.Weight = styles.WeightSemiBold
			s.Color = ColorGrayDark
		})
		desc := core.NewText(infoCard).SetText(collection.Description)
	}

	// Action buttons
	actionsRow := core.NewFrame(content)
	actionsRow.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(core.Dp(12))
		s.Justify.Content = styles.End
	})

	// Add container button
	addContainerBtn := core.NewButton(actionsRow).SetText("Add Container").SetIcon(icons.Add)
	addContainerBtn.Style(func(s *styles.Style) {
		s.Background = ColorPrimary
		s.Color = ColorWhite
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(core.Dp(12), core.Dp(16))
		s.Gap.Set(core.Dp(8))
	})
	addContainerBtn.OnClick(func(e events.Event) {
		app.showCreateContainerDialog(collection)
	})

	// Import objects button
	importObjectsBtn := core.NewButton(actionsRow).SetText("Import Objects").SetIcon(icons.Upload)
	importObjectsBtn.Style(func(s *styles.Style) {
		s.Background = ColorAccent
		s.Color = ColorBlack
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(core.Dp(12), core.Dp(16))
		s.Gap.Set(core.Dp(8))
	})

	// Containers section
	containersTitle := core.NewText(content).SetText("Containers")
	containersTitle.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(18)
		s.Font.Weight = styles.WeightSemiBold
	})

	// Mock containers for now
	containers := []Container{
		{ID: "1", Name: "Kitchen Pantry", Description: "Main pantry storage", CollectionID: collection.ID},
		{ID: "2", Name: "Refrigerator", Description: "Cold storage", CollectionID: collection.ID},
	}

	if len(containers) == 0 {
		emptyContainers := app.createEmptyState(content, "No containers found", "Add containers to organize your objects!", icons.FolderOpen)
		_ = emptyContainers
	} else {
		for _, container := range containers {
			app.createContainerCard(content, container, collection)
		}
	}

	_ = header
	app.mainContainer.Update()
}

// Container card for collection detail view
func (app *App) createContainerCard(parent core.Widget, container Container, collection Collection) *core.Frame {
	card := core.NewFrame(parent)
	card.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Justify.Content = styles.SpaceBetween
		s.Background = ColorWhite
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(core.Dp(16))
		s.Border.Style.Set(styles.BorderSolid)
		s.Border.Width.Set(core.Dp(1))
		s.Border.Color.Set(ColorGrayLight)
		s.Margin.Bottom = core.Dp(8)
	})

	// Container info (clickable)
	infoContainer := core.NewFrame(card)
	infoContainer.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Gap.Set(core.Dp(12))
		s.Grow.Set(1, 0)
		s.Cursor = styles.CursorPointer
	})
	infoContainer.OnClick(func(e events.Event) {
		app.showContainerDetailView(container, collection)
	})

	containerIcon := core.NewIcon(infoContainer).SetIcon(icons.FolderOpen)
	containerIcon.Style(func(s *styles.Style) {
		s.Color = ColorPrimary
	})

	details := core.NewFrame(infoContainer)
	details.Style(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(core.Dp(4))
	})

	containerName := core.NewText(details).SetText(container.Name)
	containerName.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(16)
		s.Font.Weight = styles.WeightSemiBold
	})

	if container.Description != "" {
		containerDesc := core.NewText(details).SetText(container.Description)
		containerDesc.Style(func(s *styles.Style) {
			s.Font.Size = core.Dp(14)
			s.Color = ColorGrayDark
		})
	}

	objectsText := core.NewText(details).SetText(fmt.Sprintf("%d objects", len(container.Objects)))
	objectsText.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(12)
		s.Color = ColorGrayDark
	})

	// Actions
	actionsContainer := core.NewFrame(card)
	actionsContainer.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(core.Dp(8))
	})

	// Edit container button
	editBtn := core.NewButton(actionsContainer).SetIcon(icons.Edit)
	editBtn.Style(func(s *styles.Style) {
		s.Background = ColorAccent
		s.Color = ColorBlack
		s.Border.Radius = styles.BorderRadiusFull
		s.Padding.Set(core.Dp(6))
	})
	editBtn.OnClick(func(e events.Event) {
		app.showEditContainerDialog(container, collection)
	})

	// Delete container button
	deleteBtn := core.NewButton(actionsContainer).SetIcon(icons.Delete)
	deleteBtn.Style(func(s *styles.Style) {
		s.Background = ColorDanger
		s.Color = ColorWhite
		s.Border.Radius = styles.BorderRadiusFull
		s.Padding.Set(core.Dp(6))
	})
	deleteBtn.OnClick(func(e events.Event) {
		app.showDeleteContainerDialog(container, collection)
	})

	return card
}

// Helper functions for collection types
func (app *App) getCollectionTypeIcon(objectType string) icons.Icon {
	switch strings.ToLower(objectType) {
	case "food":
		return icons.Restaurant
	case "book":
		return icons.Book
	case "videogame":
		return icons.Games
	case "music":
		return icons.MusicNote
	case "boardgame":
		return icons.Casino
	default:
		return icons.Inventory
	}
}

func (app *App) getCollectionTypeColor(objectType string) core.RGBA {
	switch strings.ToLower(objectType) {
	case "food":
		return core.RGB(76, 175, 80)   // Green
	case "book":
		return core.RGB(63, 81, 181)   // Indigo
	case "videogame":
		return core.RGB(156, 39, 176)  // Purple
	case "music":
		return core.RGB(255, 152, 0)   // Orange
	case "boardgame":
		return core.RGB(233, 30, 99)   // Pink
	default:
		return ColorPrimary // Primary teal
	}
}

// Collection dialogs
func (app *App) showCreateCollectionDialog() {
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

	title := core.NewText(dialog).SetText("Create New Collection")
	title.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(20)
		s.Font.Weight = styles.WeightSemiBold
	})

	// Collection name
	nameField := core.NewTextField(dialog)
	nameField.SetText("").SetPlaceholder("Collection name")

	// Collection description
	descField := core.NewTextField(dialog)
	descField.SetText("").SetPlaceholder("Description (optional)")

	// Object type selection
	typeLabel := core.NewText(dialog).SetText("Object Type")
	typeLabel.Style(func(s *styles.Style) {
		s.Font.Weight = styles.WeightSemiBold
	})

	// Type selection buttons
	typeContainer := core.NewFrame(dialog)
	typeContainer.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Wrap = true
		s.Gap.Set(core.Dp(8))
	})

	objectTypes := []string{"food", "book", "videogame", "music", "boardgame", "general"}
	var selectedType string

	for _, objType := range objectTypes {
		typeBtn := core.NewButton(typeContainer).SetText(strings.Title(objType))
		typeBtn.Style(func(s *styles.Style) {
			if selectedType == objType {
				s.Background = ColorPrimary
				s.Color = ColorWhite
			} else {
				s.Background = core.RGB(240, 240, 240)
				s.Color = ColorBlack
			}
			s.Border.Radius = styles.BorderRadiusMedium
			s.Padding.Set(core.Dp(8), core.Dp(12))
		})
		
		// Capture the type value for the closure
		capturedType := objType
		typeBtn.OnClick(func(e events.Event) {
			selectedType = capturedType
			// Refresh the dialog to show selected state
			app.hideOverlay()
			app.showCreateCollectionDialog()
		})
	}

	// Buttons
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

	createBtn := core.NewButton(buttonRow).SetText("Create Collection")
	createBtn.Style(func(s *styles.Style) {
		s.Background = ColorAccent
		s.Color = ColorBlack
	})
	createBtn.OnClick(func(e events.Event) {
		if selectedType == "" {
			selectedType = "general"
		}
		app.handleCreateCollection(nameField.Text(), descField.Text(), selectedType)
	})

	app.showOverlay(overlay)
}

// Collection API handlers
func (app *App) handleCreateCollection(name, description, objectType string) {
	if strings.TrimSpace(name) == "" {
		return
	}

	fmt.Printf("Creating collection: %s - %s (type: %s)\n", name, description, objectType)
	
	app.hideOverlay()
	app.fetchCollections()
	app.showEnhancedCollectionsView()
}

func (app *App) showEditCollectionDialog(collection Collection) {
	// Similar implementation to create dialog but with pre-filled values
	fmt.Printf("Edit collection dialog for: %s\n", collection.Name)
}

func (app *App) showDeleteCollectionDialog(collection Collection) {
	// Similar implementation to delete group dialog
	fmt.Printf("Delete collection dialog for: %s\n", collection.Name)
}

func (app *App) showImportDialog() {
	fmt.Printf("Import dialog opened\n")
}

func (app *App) showCreateContainerDialog(collection Collection) {
	fmt.Printf("Create container dialog for collection: %s\n", collection.Name)
}


func (app *App) showEditContainerDialog(container Container, collection Collection) {
	fmt.Printf("Edit container dialog: %s\n", container.Name)
}

func (app *App) showDeleteContainerDialog(container Container, collection Collection) {
	fmt.Printf("Delete container dialog: %s\n", container.Name)
}