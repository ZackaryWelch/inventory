package main

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
	content.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Grow.Set(1, 1)
		s.Padding.Set(units.Dp(16))
		s.Gap.Set(units.Dp(16))
	})

	// Action buttons row
	actionsRow := core.NewFrame(content)
	actionsRow.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(units.Dp(12))
		s.Justify.Content = styles.End
	})

	// Create collection button
	createBtn := core.NewButton(actionsRow).SetText("Create Collection").SetIcon(icons.Add)
	createBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorAccent) // var(--color-accent)
		s.Color = colors.Uniform(ColorBlack)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(12), units.Dp(16))
		s.Gap.Set(units.Dp(8))
	})
	createBtn.OnClick(func(e events.Event) {
		app.showCreateCollectionDialog()
	})

	// Import collection button
	importBtn := core.NewButton(actionsRow).SetText("Import").SetIcon(icons.Upload)
	importBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorPrimary)
		s.Color = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(12), units.Dp(16))
		s.Gap.Set(units.Dp(8))
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
		collectionsGrid.Styler(func(s *styles.Style) {
			s.Direction = styles.Row
			s.Wrap = true
			s.Gap.Set(units.Dp(16))
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
	card.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Background = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(16))
		s.Border.Style.Set(styles.BorderSolid)
		s.Border.Width.Set(units.Dp(1))
		s.Border.Color.Set(colors.Uniform(ColorGrayLight))
		s.Min.X.Set(280, units.UnitDp)
		s.Max.X.Set(320, units.UnitDp)
		s.Gap.Set(units.Dp(12))
		s.Cursor = cursors.Pointer
	})

	// Header with icon and actions
	cardHeader := core.NewFrame(card)
	cardHeader.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Justify.Content = styles.SpaceBetween
	})

	// Icon and title section
	titleSection := core.NewFrame(cardHeader)
	titleSection.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Gap.Set(units.Dp(8))
		s.Grow.Set(1, 0)
	})
	titleSection.OnClick(func(e events.Event) {
		app.showCollectionDetailView(collection)
	})

	// Collection type icon
	typeIcon := app.getIcon(collection.ObjectType)
	collectionIcon := core.NewIcon(titleSection).SetIcon(typeIcon)
	collectionIcon.Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(app.getCollectionTypeColor(collection.ObjectType))
	})

	titleContainer := core.NewFrame(titleSection)
	titleContainer.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(units.Dp(2))
	})

	collectionName := core.NewText(titleContainer).SetText(collection.Name)
	collectionName.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(16)
		s.Font.Weight = styles.WeightSemiBold
		s.Color = colors.Uniform(ColorBlack)
	})

	objectType := core.NewText(titleContainer).SetText(strings.Title(collection.ObjectType))
	objectType.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(12)
		s.Color = colors.Uniform(ColorGrayDark)
	})

	// Actions menu
	actionsMenu := core.NewFrame(cardHeader)
	actionsMenu.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(units.Dp(4))
	})

	// Edit button
	editBtn := core.NewButton(actionsMenu).SetIcon(icons.Edit)
	editBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorAccent)
		s.Color = colors.Uniform(ColorBlack)
		s.Border.Radius = styles.BorderRadiusFull
		s.Padding.Set(units.Dp(6))
	})
	editBtn.OnClick(func(e events.Event) {
		app.showEditCollectionDialog(collection)
	})

	// Delete button
	deleteBtn := core.NewButton(actionsMenu).SetIcon(icons.Delete)
	deleteBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorDanger)
		s.Color = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusFull
		s.Padding.Set(units.Dp(6))
	})
	deleteBtn.OnClick(func(e events.Event) {
		app.showDeleteCollectionDialog(collection)
	})

	// Description
	if collection.Description != "" {
		desc := core.NewText(card).SetText(collection.Description)
		desc.Styler(func(s *styles.Style) {
			s.Font.Size = units.Dp(14)
			s.Color = colors.Uniform(ColorGrayDark)
			s.Text.Align = styles.Start
		})
	}

	// Stats section
	statsContainer := core.NewFrame(card)
	statsContainer.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(units.Dp(16))
		s.Justify.Content = styles.SpaceBetween
		s.Background = colors.Uniform(ColorGrayLightest)
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(units.Dp(12))
	})

	// Containers count
	containersStats := core.NewFrame(statsContainer)
	containersStats.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Align.Items = styles.Center
		s.Gap.Set(units.Dp(2))
	})

	containersCount := core.NewText(containersStats).SetText("0") // Would be calculated from actual data
	containersCount.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(18)
		s.Font.Weight = styles.WeightBold
		s.Color = colors.Uniform(ColorPrimary)
	})

	containersLabel := core.NewText(containersStats).SetText("Containers")
	containersLabel.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(10)
		s.Color = colors.Uniform(ColorGrayDark)
	})

	// Objects count
	objectsStats := core.NewFrame(statsContainer)
	objectsStats.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Align.Items = styles.Center
		s.Gap.Set(units.Dp(2))
	})

	objectsCount := core.NewText(objectsStats).SetText("0") // Would be calculated from actual data
	objectsCount.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(18)
		s.Font.Weight = styles.WeightBold
		s.Color = colors.Uniform(ColorAccent)
	})

	objectsLabel := core.NewText(objectsStats).SetText("Objects")
	objectsLabel.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(10)
		s.Color = colors.Uniform(ColorGrayDark)
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
	content.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Grow.Set(1, 1)
		s.Padding.Set(units.Dp(16))
		s.Gap.Set(units.Dp(16))
	})

	// Collection info card
	infoCard := core.NewFrame(content)
	infoCard.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Background = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(16))
		s.Gap.Set(units.Dp(12))
	})

	// Collection type
	typeRow := core.NewFrame(infoCard)
	typeRow.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Gap.Set(units.Dp(8))
	})

	typeIcon := app.getIcon(collection.ObjectType)
	icon := core.NewIcon(typeRow).SetIcon(typeIcon)
	icon.Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(app.getCollectionTypeColor(collection.ObjectType))
	})

	typeText := core.NewText(typeRow).SetText(fmt.Sprintf("Type: %s", strings.Title(collection.ObjectType)))
	typeText.Styler(func(s *styles.Style) {
		s.Font.Weight = styles.WeightMedium
	})

	// Description
	if collection.Description != "" {
		descTitle := core.NewText(infoCard).SetText("Description")
		descTitle.Styler(func(s *styles.Style) {
			s.Font.Weight = styles.WeightSemiBold
			s.Color = colors.Uniform(ColorGrayDark)
		})
		desc := core.NewText(infoCard).SetText(collection.Description)
		desc.Styler(func(s *styles.Style) {
			s.Color = colors.Uniform(ColorGrayDark)
		})
	}

	// Action buttons
	actionsRow := core.NewFrame(content)
	actionsRow.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(units.Dp(12))
		s.Justify.Content = styles.End
	})

	// Add container button
	addContainerBtn := core.NewButton(actionsRow).SetText("Add Container").SetIcon(icons.Add)
	addContainerBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorPrimary)
		s.Color = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(12), units.Dp(16))
		s.Gap.Set(units.Dp(8))
	})
	addContainerBtn.OnClick(func(e events.Event) {
		app.showCreateContainerDialog(collection)
	})

	// Import objects button
	importObjectsBtn := core.NewButton(actionsRow).SetText("Import Objects").SetIcon(icons.Upload)
	importObjectsBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorAccent)
		s.Color = colors.Uniform(ColorBlack)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(12), units.Dp(16))
		s.Gap.Set(units.Dp(8))
	})

	// Containers section
	containersTitle := core.NewText(content).SetText("Containers")
	containersTitle.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(18)
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
	card.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Justify.Content = styles.SpaceBetween
		s.Background = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(16))
		s.Border.Style.Set(styles.BorderSolid)
		s.Border.Width.Set(units.Dp(1))
		s.Border.Color.Set(colors.Uniform(ColorGrayLight))
		s.Margin.Bottom = units.Dp(8)
	})

	// Container info (clickable)
	infoContainer := core.NewFrame(card)
	infoContainer.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Gap.Set(units.Dp(12))
		s.Grow.Set(1, 0)
		s.Cursor = cursors.Pointer
	})
	infoContainer.OnClick(func(e events.Event) {
		app.showContainerDetailView(container, collection)
	})

	containerIcon := core.NewIcon(infoContainer).SetIcon(icons.FolderOpen)
	containerIcon.Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(ColorPrimary)
	})

	details := core.NewFrame(infoContainer)
	details.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(units.Dp(4))
	})

	containerName := core.NewText(details).SetText(container.Name)
	containerName.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(16)
		s.Font.Weight = styles.WeightSemiBold
	})

	if container.Description != "" {
		containerDesc := core.NewText(details).SetText(container.Description)
		containerDesc.Styler(func(s *styles.Style) {
			s.Font.Size = units.Dp(14)
			s.Color = colors.Uniform(ColorGrayDark)
		})
	}

	objectsText := core.NewText(details).SetText(fmt.Sprintf("%d objects", len(container.Objects)))
	objectsText.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(12)
		s.Color = colors.Uniform(ColorGrayDark)
	})

	// Actions
	actionsContainer := core.NewFrame(card)
	actionsContainer.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(units.Dp(8))
	})

	// Edit container button
	editBtn := core.NewButton(actionsContainer).SetIcon(icons.Edit)
	editBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorAccent)
		s.Color = colors.Uniform(ColorBlack)
		s.Border.Radius = styles.BorderRadiusFull
		s.Padding.Set(units.Dp(6))
	})
	editBtn.OnClick(func(e events.Event) {
		app.showEditContainerDialog(container, collection)
	})

	// Delete container button
	deleteBtn := core.NewButton(actionsContainer).SetIcon(icons.Delete)
	deleteBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorDanger)
		s.Color = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusFull
		s.Padding.Set(units.Dp(6))
	})
	deleteBtn.OnClick(func(e events.Event) {
		app.showDeleteContainerDialog(container, collection)
	})

	return card
}

// Helper functions for collection types
func (app *App) getIcon(objectType string) icons.Icon {
	switch strings.ToLower(objectType) {
	case "food":
		return icons.Dining
	case "book":
		return icons.Book
	case "videogame":
		return icons.VideogameAsset
	case "music":
		return icons.MusicNote
	case "boardgame":
		return icons.Extension
	default:
		return icons.Folder
	}
}

func (app *App) getCollectionTypeColor(objectType string) color.RGBA {
	switch strings.ToLower(objectType) {
	case "food":
		return color.RGBA{R: 76, G: 175, B: 80, A: 255} // Green
	case "book":
		return color.RGBA{R: 63, G: 81, B: 181, A: 255} // Indigo
	case "videogame":
		return color.RGBA{R: 156, G: 39, B: 176, A: 255} // Purple
	case "music":
		return color.RGBA{R: 255, G: 152, B: 0, A: 255} // Orange
	case "boardgame":
		return color.RGBA{R: 233, G: 30, B: 99, A: 255} // Pink
	default:
		return ColorPrimary // Primary teal
	}
}

// Collection dialogs
func (app *App) showCreateCollectionDialog() {
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

	title := core.NewText(dialog).SetText("Create New Collection")
	title.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(20)
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
	typeLabel.Styler(func(s *styles.Style) {
		s.Font.Weight = styles.WeightSemiBold
	})

	// Type selection buttons
	typeContainer := core.NewFrame(dialog)
	typeContainer.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Wrap = true
		s.Gap.Set(units.Dp(8))
	})

	objectTypes := []string{"food", "book", "videogame", "music", "boardgame", "general"}
	var selectedType string

	for _, objType := range objectTypes {
		typeBtn := core.NewButton(typeContainer).SetText(strings.Title(objType))
		typeBtn.Styler(func(s *styles.Style) {
			if selectedType == objType {
				s.Background = colors.Uniform(ColorPrimary)
				s.Color = colors.Uniform(ColorWhite)
			} else {
				s.Background = colors.Uniform(color.RGBA{R: 240, G: 240, B: 240, A: 255})
				s.Color = colors.Uniform(ColorBlack)
			}
			s.Border.Radius = styles.BorderRadiusMedium
			s.Padding.Set(units.Dp(8), units.Dp(12))
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
	buttonRow.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(units.Dp(12))
		s.Justify.Content = styles.End
	})

	cancelBtn := core.NewButton(buttonRow).SetText("Cancel")
	cancelBtn.OnClick(func(e events.Event) {
		app.hideOverlay()
	})

	createBtn := core.NewButton(buttonRow).SetText("Create Collection")
	createBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorAccent)
		s.Color = colors.Uniform(ColorBlack)
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
