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
	"cogentcore.org/core/styles/sides"
	"cogentcore.org/core/styles/units"

	"github.com/nishiki/frontend/pkg/types"
	"github.com/nishiki/frontend/ui/components"
	"github.com/nishiki/frontend/ui/layouts"
	appstyles "github.com/nishiki/frontend/ui/styles"
)

// Enhanced Collections View with full CRUD operations
func (app *App) showEnhancedCollectionsView() {
	app.mainContainer.DeleteChildren()
	app.currentView = "collections"

	// Refresh collections data
	if err := app.fetchCollections(); err != nil {
		fmt.Printf("Error fetching collections: %v\n", err)
	}

	// Page title - using helper function
	layouts.PageTitle(app.mainContainer, "Foods")

	// Main content - using existing layout function
	content := layouts.ContentColumn(app.mainContainer)

	// Search bar matching React design
	searchField := core.NewTextField(content)
	searchField.SetPlaceholder("Search Foods...")
	searchField.Styler(func(s *styles.Style) {
		appstyles.StyleInputRounded(s) // Apply rounded input style
		s.Margin.Bottom = units.Dp(appstyles.Spacing4)
	})

	// Filter chips row (Group and Container filters)
	if app.searchFilter != nil && (len(app.searchFilter.SelectedTypes) > 0 || app.searchFilter.SearchQuery != "") {
		filtersRow := core.NewFrame(content)
		filtersRow.Styler(func(s *styles.Style) {
			s.Direction = styles.Row
			s.Gap.Set(units.Dp(appstyles.Spacing2))
			s.Wrap = true
			s.Margin.Bottom = units.Dp(appstyles.Spacing4)
		})

		// Group filter chip
		if len(app.searchFilter.SelectedTypes) > 0 {
			for _, filterType := range app.searchFilter.SelectedTypes {
				chip := core.NewFrame(filtersRow)
				chip.Styler(func(s *styles.Style) {
					s.Direction = styles.Row
					s.Align.Items = styles.Center
					s.Gap.Set(units.Dp(4))
					s.Background = colors.Uniform(appstyles.ColorPrimaryLightest)
					s.Border.Radius = sides.NewValues(units.Dp(appstyles.RadiusFull))
					s.Padding.Set(units.Dp(6), units.Dp(12))
				})

				chipText := core.NewText(chip).SetText(filterType)
				chipText.Styler(func(s *styles.Style) {
					s.Font.Size = units.Dp(appstyles.FontSizeSM)
					s.Color = colors.Uniform(appstyles.ColorPrimary)
				})

				closeBtn := core.NewButton(chip).SetIcon(icons.Close)
				closeBtn.Styler(func(s *styles.Style) {
					s.Background = nil
					s.Color = colors.Uniform(appstyles.ColorPrimary)
					s.Padding.Set(units.Dp(2))
					s.Font.Size = units.Dp(12)
				})
			}
		}
	}

	// Sort dropdown
	sortRow := core.NewFrame(content)
	sortRow.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Justify.Content = styles.End
		s.Margin.Bottom = units.Dp(appstyles.Spacing4)
	})

	sortDropdown := core.NewButton(sortRow).SetText("Created At (Newest → Oldest)").SetIcon(icons.ArrowDropDown)
	sortDropdown.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(appstyles.ColorGrayLightest)
		s.Border.Radius = sides.NewValues(units.Dp(appstyles.RadiusMD))
		s.Padding.Set(units.Dp(8), units.Dp(12))
		s.Gap.Set(units.Dp(4))
		s.Font.Size = units.Dp(appstyles.FontSizeSM)
	})

	// Collections grid
	if len(app.collections) == 0 {
		components.EmptyState(content, "No collections found. Create your first collection to start managing your inventory!")
	} else {
		collectionsGrid := core.NewFrame(content)
		collectionsGrid.Styler(appstyles.StyleCollectionsGrid)

		for _, collection := range app.collections {
			app.createEnhancedCollectionCard(collectionsGrid, collection)
		}
	}

	// Fixed FAB at bottom-right (React pattern: fixed bottom-[5.5rem] right-4)
	// Bottom nav is at bottom-0, so FAB is ~88px above it
	fab := core.NewButton(app.mainContainer).SetIcon(icons.Add)
	fab.Styler(func(s *styles.Style) {
		s.Min.X.Set(56, units.UnitDp)                        // w-14 (56px)
		s.Min.Y.Set(56, units.UnitDp)                        // aspect-square
		s.Background = colors.Uniform(appstyles.ColorAccent) // bg-accent (yellow)
		s.Color = colors.Uniform(appstyles.ColorBlack)       // Black icon
		s.Border.Radius = sides.NewValues(units.Dp(9999))    // rounded-full
		// TODO: Need to position fixed at bottom-right
		// For now it will appear in flow, but ideally: bottom-[5.5rem] right-4
	})
	fab.OnClick(func(e events.Event) {
		app.showCreateCollectionDialog() // Open dialog using Cogent Core's built-in system
	})

	// Bottom navigation bar
	app.updateBottomMenu("collections")

	app.body.Update()
}

// Create enhanced collection card with actions
func (app *App) createEnhancedCollectionCard(parent core.Widget, collection Collection) *core.Frame {
	typeIcon := app.getIcon(collection.ObjectType)
	typeColor := app.getCollectionTypeColor(collection.ObjectType)

	return app.createCard(parent, CardConfig{
		Icon:        typeIcon,
		IconColor:   typeColor,
		Title:       collection.Name,
		Description: collection.Location,
		OnClick: func() {
			app.showCollectionDetailView(collection)
		},
		Actions: []CardAction{
			{Icon: icons.Edit, Color: appstyles.ColorAccent, Tooltip: "Edit collection", OnClick: func() {
				app.showEditCollectionDialog(collection)
			}},
			{Icon: icons.Delete, Color: appstyles.ColorDanger, Tooltip: "Delete collection", OnClick: func() {
				app.showDeleteCollectionDialog(collection)
			}},
		},
		Content: func(card core.Widget) {
			// Object type badge
			typeBadge := core.NewText(card).SetText(strings.Title(collection.ObjectType))
			typeBadge.Styler(func(s *styles.Style) {
				s.Font.Size = units.Dp(12)
				s.Color = colors.Uniform(appstyles.ColorGrayDark)
			})

			// Stats section
			statsContainer := core.NewFrame(card)
			statsContainer.Styler(appstyles.StyleStatsRow)

			// Containers count
			containersStats := createFlexColumn(statsContainer, 2)
			containersStats.Styler(func(s *styles.Style) {
				s.Align.Items = styles.Center
			})
			containersCount := core.NewText(containersStats).SetText("0")
			containersCount.Styler(appstyles.StyleStatValuePrimary)
			containersLabel := core.NewText(containersStats).SetText("Containers")
			containersLabel.Styler(appstyles.StyleSmallText)

			// Objects count
			objectsStats := createFlexColumn(statsContainer, 2)
			objectsStats.Styler(func(s *styles.Style) {
				s.Align.Items = styles.Center
			})
			objectsCount := core.NewText(objectsStats).SetText("0")
			objectsCount.Styler(appstyles.StyleStatValueAccent)
			objectsLabel := core.NewText(objectsStats).SetText("Objects")
			objectsLabel.Styler(appstyles.StyleSmallText)
		},
	})
}

// Collection Detail View
func (app *App) showCollectionDetailView(collection Collection) {
	app.mainContainer.DeleteChildren()
	app.currentView = "collection_detail"

	// Header with back button
	layouts.SimpleHeader(app.mainContainer, collection.Name, true, func() {
		app.showEnhancedCollectionsView()
	})

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
		s.Background = colors.Uniform(appstyles.ColorWhite)
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
		s.Font.Weight = appstyles.WeightMedium
	})

	// Location
	if collection.Location != "" {
		locTitle := core.NewText(infoCard).SetText("Location")
		locTitle.Styler(func(s *styles.Style) {
			s.Font.Weight = appstyles.WeightSemiBold
			s.Color = colors.Uniform(appstyles.ColorGrayDark)
		})
		loc := core.NewText(infoCard).SetText(collection.Location)
		loc.Styler(func(s *styles.Style) {
			s.Color = colors.Uniform(appstyles.ColorGrayDark)
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
	addContainerBtn.Styler(appstyles.StyleButtonPrimary)
	addContainerBtn.OnClick(func(e events.Event) {
		app.showCreateContainerDialog(collection)
	})

	// Import objects button
	importObjectsBtn := core.NewButton(actionsRow).SetText("Import Objects").SetIcon(icons.Upload)
	importObjectsBtn.Styler(appstyles.StyleButtonAccent)

	// Containers section
	containersTitle := core.NewText(content).SetText("Containers")
	containersTitle.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(18)
		s.Font.Weight = appstyles.WeightSemiBold
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

	app.mainContainer.Update()
}

// Container card for collection detail view
func (app *App) createContainerCard(parent core.Widget, container Container, collection Collection) *core.Frame {
	return app.createCard(parent, CardConfig{
		Icon:        icons.FolderOpen,
		IconColor:   appstyles.ColorPrimary,
		Title:       container.Name,
		Description: container.Description,
		Stats: []CardStat{
			{Label: "objects", Value: fmt.Sprintf("%d", len(container.Objects))},
		},
		OnClick: func() {
			app.showContainerDetailView(container, collection)
		},
		Actions: []CardAction{
			{Icon: icons.Edit, Color: appstyles.ColorAccent, Tooltip: "Edit container", OnClick: func() {
				app.showEditContainerDialog(container, collection)
			}},
			{Icon: icons.Delete, Color: appstyles.ColorDanger, Tooltip: "Delete container", OnClick: func() {
				app.showDeleteContainerDialog(container, collection)
			}},
		},
	})
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
		return appstyles.ColorPrimary // Primary teal
	}
}

// Removed showCollectionActionsMenu - FAB opens Create Collection directly

// Collection dialogs
func (app *App) showCreateCollectionDialog() {
	app.showCreateCollectionDialogWithTypeAndValues("", "", "")
}

func (app *App) showCreateCollectionDialogWithType(selectedType string) {
	app.showCreateCollectionDialogWithTypeAndValues(selectedType, "", "")
}

func (app *App) showCreateCollectionDialogWithTypeAndValues(selectedType, currentName, currentDesc string) {
	var nameField, descField *core.TextField
	objectTypes := []string{"food", "book", "videogame", "music", "boardgame", "general"}

	app.showDialog(DialogConfig{
		Title:             "Create New Collection",
		SubmitButtonText:  "Create Collection",
		SubmitButtonStyle: appstyles.StyleButtonAccent,
		ContentBuilder: func(dialog core.Widget) {
			nameField = createTextField(dialog, "Collection name")
			if currentName != "" {
				nameField.SetText(currentName) // Restore previous value
			}

			descField = createTextField(dialog, "Location (optional)")
			if currentDesc != "" {
				descField.SetText(currentDesc) // Restore previous value
			}

			// Object type selection
			typeLabel := core.NewText(dialog).SetText("Object Type")
			typeLabel.Styler(func(s *styles.Style) {
				s.Font.Weight = appstyles.WeightSemiBold
				s.Color = colors.Uniform(appstyles.ColorBlack) // Ensure label is visible
			})

			typeContainer := createFlexRow(dialog, 8, styles.Start)
			typeContainer.Styler(func(s *styles.Style) {
				s.Wrap = true
				s.Max.X.Set(100, units.UnitPw) // Constrain to parent width for wrapping
			})

			for _, objType := range objectTypes {
				typeBtn := core.NewButton(typeContainer).SetText(strings.Title(objType))
				if selectedType == objType {
					typeBtn.Styler(appstyles.StyleObjectTypeButtonSelected)
				} else {
					typeBtn.Styler(appstyles.StyleObjectTypeButtonUnselected)
				}

				capturedType := objType
				typeBtn.OnClick(func(e events.Event) {
					// Update button styling without recreating dialog
					// The selectedType will be captured when form is submitted
					selectedType = capturedType
					// TODO: Update button styles dynamically to show selection
					// For now, selection state is only stored, visual feedback on submit
				})
			}
		},
		OnSubmit: func() {
			typeToUse := selectedType
			if typeToUse == "" {
				typeToUse = "general"
			}
			app.handleCreateCollection(nameField.Text(), descField.Text(), typeToUse)
		},
	})
}

// Collection API handlers
func (app *App) handleCreateCollection(name, location, objectType string) {
	if strings.TrimSpace(name) == "" {
		app.logger.Error("Collection name cannot be empty")
		return
	}

	if app.currentUser == nil {
		app.logger.Error("No current user for collection creation")
		return
	}

	// Create request using types
	req := types.CreateCollectionRequest{
		UserID:     app.currentUser.ID,
		Name:       name,
		ObjectType: objectType,
		Location:   location,
	}

	// Make API call to create collection using client
	app.logger.Info("Creating collection", "name", name, "type", objectType)
	collection, err := app.collectionsClient.Create(app.currentUser.ID, req)
	if err != nil {
		app.logger.Error("Failed to create collection", "error", err)
		return
	}

	app.logger.Info("Collection created successfully", "collection_id", collection.ID)

	app.fetchCollections()
	app.showEnhancedCollectionsView()
}

func (app *App) handleEditCollection(collectionID, name, location string) {
	if strings.TrimSpace(name) == "" {
		app.logger.Error("Collection name cannot be empty")
		return
	}

	if app.currentUser == nil {
		app.logger.Error("No current user for collection update")
		return
	}

	// Create request using types
	req := types.UpdateCollectionRequest{
		Name:     name,
		Location: location,
	}

	// Make API call to update collection using client
	app.logger.Info("Updating collection", "collection_id", collectionID, "name", name)
	collection, err := app.collectionsClient.Update(app.currentUser.ID, collectionID, req)
	if err != nil {
		app.logger.Error("Failed to update collection", "error", err)
		return
	}

	app.logger.Info("Collection updated successfully", "collection_id", collection.ID)

	app.fetchCollections()
	app.showEnhancedCollectionsView()
}

func (app *App) handleDeleteCollection(collectionID string) {
	if app.currentUser == nil {
		app.logger.Error("No current user for collection deletion")
		return
	}

	// Make API call to delete collection using client
	app.logger.Info("Deleting collection", "collection_id", collectionID)
	err := app.collectionsClient.Delete(app.currentUser.ID, collectionID)
	if err != nil {
		app.logger.Error("Failed to delete collection", "error", err)
		return
	}

	app.logger.Info("Collection deleted successfully")

	app.fetchCollections()
	app.showEnhancedCollectionsView()
}

func (app *App) handleImport(fileData string) {
	fmt.Printf("Importing data: %s\n", fileData)
	app.fetchCollections()
	app.showEnhancedCollectionsView()
}

func (app *App) handleCreateContainer(collectionID, name, description string) {
	if strings.TrimSpace(name) == "" {
		app.logger.Error("Container name cannot be empty")
		return
	}

	if app.currentUser == nil {
		app.logger.Error("No current user for container creation")
		return
	}

	// Create request using types
	req := types.CreateContainerRequest{
		Name:        name,
		Description: description,
	}

	// Make API call to create container using client
	app.logger.Info("Creating container", "collection_id", collectionID, "name", name)
	container, err := app.containersClient.Create(app.currentUser.ID, collectionID, req)
	if err != nil {
		app.logger.Error("Failed to create container", "error", err)
		return
	}

	app.logger.Info("Container created successfully", "container_id", container.ID)

	// Refresh the collection view
	app.fetchCollections()
}

func (app *App) handleEditContainer(collectionID, containerID, name, description string) {
	if strings.TrimSpace(name) == "" {
		app.logger.Error("Container name cannot be empty")
		return
	}

	if app.currentUser == nil {
		app.logger.Error("No current user for container update")
		return
	}

	// Create request using types
	req := types.UpdateContainerRequest{
		Name:        name,
		Description: description,
	}

	// Make API call to update container using client
	app.logger.Info("Updating container", "collection_id", collectionID, "container_id", containerID, "name", name)
	container, err := app.containersClient.Update(app.currentUser.ID, collectionID, containerID, req)
	if err != nil {
		app.logger.Error("Failed to update container", "error", err)
		return
	}

	app.logger.Info("Container updated successfully", "container_id", container.ID)

	// Refresh the collection view
	app.fetchCollections()
}

func (app *App) handleDeleteContainer(collectionID, containerID string) {
	if app.currentUser == nil {
		app.logger.Error("No current user for container deletion")
		return
	}

	// Make API call to delete container using client
	app.logger.Info("Deleting container", "collection_id", collectionID, "container_id", containerID)
	err := app.containersClient.Delete(app.currentUser.ID, collectionID, containerID)
	if err != nil {
		app.logger.Error("Failed to delete container", "error", err)
		return
	}

	app.logger.Info("Container deleted successfully")

	// Refresh the collection view
	app.fetchCollections()
}

func (app *App) showEditCollectionDialog(collection Collection) {
	var nameField, descField *core.TextField

	app.showDialog(DialogConfig{
		Title:             "Edit Collection",
		SubmitButtonText:  "Save Changes",
		SubmitButtonStyle: appstyles.StyleButtonPrimary,
		ContentBuilder: func(dialog core.Widget) {
			nameField = createTextField(dialog, "Collection name")
			nameField.SetText(collection.Name)
			descField = createTextField(dialog, "Location (optional)")
			descField.SetText(collection.Location)
		},
		OnSubmit: func() {
			app.handleEditCollection(collection.ID, nameField.Text(), descField.Text())
		},
	})
}

func (app *App) showDeleteCollectionDialog(collection Collection) {
	app.showDialog(DialogConfig{
		Title:             "Delete Collection",
		Message:           fmt.Sprintf("Are you sure you want to delete \"%s\"? This will also delete all containers and objects within it. This action cannot be undone.", collection.Name),
		SubmitButtonText:  "Delete",
		SubmitButtonStyle: appstyles.StyleButtonDanger,
		OnSubmit: func() {
			app.handleDeleteCollection(collection.ID)
		},
	})
}

func (app *App) showImportDialog() {
	var fileField *core.TextField

	app.showDialog(DialogConfig{
		Title:             "Import Data",
		Message:           "Upload a JSON or CSV file to import objects into a new collection",
		SubmitButtonText:  "Import",
		SubmitButtonStyle: appstyles.StyleButtonPrimary,
		ContentBuilder: func(dialog core.Widget) {
			fileField = createTextField(dialog, "File path or data")
		},
		OnSubmit: func() {
			app.handleImport(fileField.Text())
		},
	})
}

func (app *App) showCreateContainerDialog(collection Collection) {
	var nameField, descField *core.TextField

	app.showDialog(DialogConfig{
		Title:             "Create Container",
		SubmitButtonText:  "Create",
		SubmitButtonStyle: appstyles.StyleButtonPrimary,
		ContentBuilder: func(dialog core.Widget) {
			nameField = createTextField(dialog, "Container name")
			descField = createTextField(dialog, "Description (optional)")
		},
		OnSubmit: func() {
			app.handleCreateContainer(collection.ID, nameField.Text(), descField.Text())
		},
	})
}

func (app *App) showEditContainerDialog(container Container, collection Collection) {
	var nameField, descField *core.TextField

	app.showDialog(DialogConfig{
		Title:             "Edit Container",
		SubmitButtonText:  "Save Changes",
		SubmitButtonStyle: appstyles.StyleButtonPrimary,
		ContentBuilder: func(dialog core.Widget) {
			nameField = createTextField(dialog, "Container name")
			nameField.SetText(container.Name)
			descField = createTextField(dialog, "Description (optional)")
			descField.SetText(container.Description)
		},
		OnSubmit: func() {
			app.handleEditContainer(collection.ID, container.ID, nameField.Text(), descField.Text())
		},
	})
}

func (app *App) showDeleteContainerDialog(container Container, collection Collection) {
	app.showDialog(DialogConfig{
		Title:             "Delete Container",
		Message:           fmt.Sprintf("Are you sure you want to delete \"%s\"? This will also delete all objects within it. This action cannot be undone.", container.Name),
		SubmitButtonText:  "Delete",
		SubmitButtonStyle: appstyles.StyleButtonDanger,
		OnSubmit: func() {
			app.handleDeleteContainer(collection.ID, container.ID)
		},
	})
}
