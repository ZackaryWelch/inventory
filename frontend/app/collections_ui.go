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
	searchField.Styler(appstyles.StyleSearchFieldWithMargin)

	// Filter chips row (Group and Container filters)
	if app.searchFilter != nil && (len(app.searchFilter.SelectedTypes) > 0 || app.searchFilter.SearchQuery != "") {
		filtersRow := core.NewFrame(content)
		filtersRow.Styler(appstyles.StyleFilterChipsRow)

		// Group filter chip
		if len(app.searchFilter.SelectedTypes) > 0 {
			for _, filterType := range app.searchFilter.SelectedTypes {
				chip := core.NewFrame(filtersRow)
				chip.Styler(appstyles.StyleFilterChip)

				chipText := core.NewText(chip).SetText(filterType)
				chipText.Styler(appstyles.StyleFilterChipText)

				closeBtn := core.NewButton(chip).SetIcon(icons.Close)
				closeBtn.Styler(appstyles.StyleFilterChipCloseButton)
			}
		}
	}

	// Sort dropdown
	sortRow := core.NewFrame(content)
	sortRow.Styler(appstyles.StyleSortRow)

	sortDropdown := core.NewButton(sortRow).SetText("Created At (Newest → Oldest)").SetIcon(icons.ArrowDropDown)
	sortDropdown.Styler(func(s *styles.Style) {
		appstyles.StyleSortDropdown(s)
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
		app.showCreateCollectionDialog()
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
			app.logger.Info("Collection card clicked!", "name", collection.Name)
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
				s.Color = colors.Uniform(appstyles.ColorBlack)
			})

			// Stats section
			statsContainer := core.NewFrame(card)
			statsContainer.Styler(appstyles.StyleStatsRow)

			// Calculate actual counts
			containerCount := len(collection.Containers)
			totalObjects := 0
			for _, container := range collection.Containers {
				totalObjects += len(container.Objects)
			}

			// Containers count
			containersStats := createFlexColumn(statsContainer, 2)
			containersStats.Styler(func(s *styles.Style) {
				s.Align.Items = styles.Center
			})
			containersCount := core.NewText(containersStats).SetText(fmt.Sprintf("%d", containerCount))
			containersCount.Styler(appstyles.StyleStatValuePrimary)
			containersLabel := core.NewText(containersStats).SetText("Containers")
			containersLabel.Styler(appstyles.StyleSmallText)

			// Objects count
			objectsStats := createFlexColumn(statsContainer, 2)
			objectsStats.Styler(func(s *styles.Style) {
				s.Align.Items = styles.Center
			})
			objectsCount := core.NewText(objectsStats).SetText(fmt.Sprintf("%d", totalObjects))
			objectsCount.Styler(appstyles.StyleStatValueAccent)
			objectsLabel := core.NewText(objectsStats).SetText("Objects")
			objectsLabel.Styler(appstyles.StyleSmallText)
		},
	})
}

// refreshCurrentCollectionView re-fetches the current collection and refreshes the detail view
func (app *App) refreshCurrentCollectionView() {
	if app.selectedCollection == nil {
		app.logger.Warn("No selected collection to refresh")
		return
	}

	// Re-fetch the collection with updated containers
	updatedCollection, err := app.collectionsClient.Get(app.currentUser.ID, app.selectedCollection.ID)
	if err != nil {
		app.logger.Error("Failed to fetch updated collection", "error", err)
		return
	}

	// Refresh the view
	app.showCollectionDetailView(*updatedCollection)
}

// Collection Detail View
func (app *App) showCollectionDetailView(collection Collection) {
	app.logger.Info("showCollectionDetailView called", "collection_id", collection.ID, "collection_name", collection.Name)

	app.selectedCollection = &collection
	app.mainContainer.DeleteChildren()
	app.currentView = "collection_detail"

	// Fetch full collection details with containers
	fullCollection, err := app.collectionsClient.Get(app.currentUser.ID, collection.ID)
	if err != nil {
		app.logger.Error("Failed to fetch collection details", "error", err)
		// Fall back to using the collection we have
		fullCollection = &collection
	} else {
		// Update selected collection with full details
		app.selectedCollection = fullCollection
		collection = *fullCollection
		app.logger.Info("Fetched collection details",
			"collection_id", collection.ID,
			"containers_count", len(collection.Containers))
	}

	// Header with back button
	layouts.SimpleHeader(app.mainContainer, collection.Name, true, func() {
		app.showEnhancedCollectionsView()
	})

	// Main content
	content := core.NewFrame(app.mainContainer)
	content.Styler(appstyles.StyleCollectionDetailContent)

	// Collection info card
	infoCard := core.NewFrame(content)
	infoCard.Styler(func(s *styles.Style) {
		appstyles.StyleCollectionInfoCard(s)
		s.Background = colors.Uniform(appstyles.ColorWhite)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(16))
	})

	// Collection type
	typeRow := core.NewFrame(infoCard)
	typeRow.Styler(appstyles.StyleTypeRow)

	typeIcon := app.getIcon(collection.ObjectType)
	icon := core.NewIcon(typeRow).SetIcon(typeIcon)
	icon.Styler(func(s *styles.Style) {
		appstyles.StyleTypeIcon(s)
		s.Color = colors.Uniform(app.getCollectionTypeColor(collection.ObjectType))
	})

	typeText := core.NewText(typeRow).SetText(fmt.Sprintf("Type: %s", strings.Title(collection.ObjectType)))
	typeText.Styler(appstyles.StyleTypeText)

	// Location
	if collection.Location != "" {
		locTitle := core.NewText(infoCard).SetText("Location")
		locTitle.Styler(appstyles.StyleLocationTitle)
		loc := core.NewText(infoCard).SetText(collection.Location)
		loc.Styler(appstyles.StyleLocationText)
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
	importObjectsBtn.OnClick(func(e events.Event) {
		app.ShowImportDialog("", collection.ID) // No specific container, will distribute across collection
	})

	// Containers section header
	containersHeaderRow := core.NewFrame(content)
	containersHeaderRow.Styler(appstyles.StyleContainersHeaderRow)

	containersTitle := core.NewText(containersHeaderRow).SetText("Containers")
	containersTitle.Styler(appstyles.StyleContainersTitle)

	// View toggle (tree vs list)
	viewToggle := core.NewButton(containersHeaderRow).SetIcon(icons.List)
	viewToggle.SetTooltip("Hierarchy View")
	viewToggle.Styler(func(s *styles.Style) {
		appstyles.StyleViewToggle(s)
		s.Color = colors.Uniform(appstyles.ColorPrimary)
		s.Padding.Set(units.Dp(8))
	})

	// Render containers
	if len(collection.Containers) == 0 {
		emptyContainers := app.createEmptyState(content, "No containers found", "Add containers to organize your objects!", icons.FolderOpen)
		_ = emptyContainers
	} else {
		// Build and render container hierarchy
		hierarchy := app.BuildContainerHierarchy(collection.Containers)

		containersFrame := core.NewFrame(content)
		containersFrame.Styler(func(s *styles.Style) {
			appstyles.StyleContainersFrame(s)
			s.Gap.Set(units.Dp(4))
		})

		app.RenderContainerTree(containersFrame, hierarchy, 0)
	}

	app.mainContainer.Update()
	app.body.Update()
}

// Container card for collection detail view
func (app *App) createContainerCard(parent core.Widget, container types.Container, collection Collection) *core.Frame {
	return app.createCard(parent, CardConfig{
		Icon:        icons.FolderOpen,
		IconColor:   appstyles.ColorPrimary,
		Title:       container.Name,
		Description: container.Location,
		Stats: []CardStat{
			{Label: "objects", Value: fmt.Sprintf("%d", len(container.Objects))},
		},
		OnClick: func() {
			app.showContainerDetail(&container)
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
		ContentBuilder: func(dialog core.Widget, closeDialog func()) {
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
			typeLabel.Styler(appstyles.StyleFormLabel)

			typeContainer := createFlexRow(dialog, 8, styles.Start)
			typeContainer.Styler(appstyles.StyleTypeButtonContainer)

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

func (app *App) handleCreateContainer(collectionID, name string) {
	app.handleCreateContainerWithDetails(collectionID, name, "", "general", nil, nil)
}

func (app *App) handleCreateContainerWithDetails(collectionID, name, location, containerType string, parentID, groupID *string) {
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
		CollectionID:      collectionID,
		Name:              name,
		Type:              containerType,
		Location:          location,
		ParentContainerID: parentID,
		GroupID:           groupID,
	}

	// Make API call to create container using client
	app.logger.Info("Creating container", "collection_id", collectionID, "name", name, "type", containerType)
	container, err := app.containersClient.Create(app.currentUser.ID, collectionID, req)
	if err != nil {
		app.logger.Error("Failed to create container", "error", err)
		return
	}

	app.logger.Info("Container created successfully", "container_id", container.ID)

	// Refresh the collection detail view to show the new container
	app.refreshCurrentCollectionView()
}

func (app *App) handleEditContainer(collectionID, containerID, name string) {
	app.handleEditContainerWithDetails(collectionID, containerID, name, "", "general", nil)
}

func (app *App) handleEditContainerWithDetails(collectionID, containerID, name, location, containerType string, groupID *string) {
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
		Name:     name,
		Type:     containerType,
		Location: location,
		GroupID:  groupID,
	}

	// Make API call to update container using client
	app.logger.Info("Updating container", "collection_id", collectionID, "container_id", containerID, "name", name)
	container, err := app.containersClient.Update(app.currentUser.ID, collectionID, containerID, req)
	if err != nil {
		app.logger.Error("Failed to update container", "error", err)
		return
	}

	app.logger.Info("Container updated successfully", "container_id", container.ID)

	// Refresh the collection detail view to show the updated container
	app.refreshCurrentCollectionView()

	// Re-open the container actions dialog with the updated container
	// Find the updated container in the refreshed collection
	if app.selectedCollection != nil {
		for i := range app.selectedCollection.Containers {
			if app.selectedCollection.Containers[i].ID == containerID {
				app.showContainerActions(&app.selectedCollection.Containers[i])
				break
			}
		}
	}
}

func (app *App) showEditCollectionDialog(collection Collection) {
	var nameField, descField *core.TextField

	app.showDialog(DialogConfig{
		Title:             "Edit Collection",
		SubmitButtonText:  "Save Changes",
		SubmitButtonStyle: appstyles.StyleButtonPrimary,
		ContentBuilder: func(dialog core.Widget, closeDialog func()) {
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
		ContentBuilder: func(dialog core.Widget, closeDialog func()) {
			fileField = createTextField(dialog, "File path or data")
		},
		OnSubmit: func() {
			app.handleImport(fileField.Text())
		},
	})
}

func (app *App) showCreateContainerDialog(collection Collection) {
	app.showCreateContainerDialogWithParent(collection, nil)
}

func (app *App) showCreateContainerDialogWithParent(collection Collection, parentContainer *types.Container) {
	var nameField, locationField *core.TextField
	var selectedGroupID *string
	var selectedType string = "general"

	containerTypes := []string{"room", "bookshelf", "shelf", "binder", "cabinet", "general"}

	app.showDialog(DialogConfig{
		Title:             "Create Container",
		SubmitButtonText:  "Create",
		SubmitButtonStyle: appstyles.StyleButtonPrimary,
		ContentBuilder: func(dialog core.Widget, closeDialog func()) {
			nameField = createTextField(dialog, "Container name")
			locationField = createTextField(dialog, "Location (optional)")

			// Container type selection
			typeLabel := core.NewText(dialog).SetText("Container Type")
			typeLabel.Styler(appstyles.StyleFormLabel)

			typeContainer := createFlexRow(dialog, 8, styles.Start)
			typeContainer.Styler(appstyles.StyleTypeButtonContainer)

			// Store button references for dynamic styling updates
			typeButtons := make([]*core.Button, 0, len(containerTypes))

			for _, containerType := range containerTypes {
				typeBtn := core.NewButton(typeContainer).SetText(strings.Title(containerType))
				typeButtons = append(typeButtons, typeBtn)

				// Apply initial styling
				if selectedType == containerType {
					typeBtn.Styler(appstyles.StyleObjectTypeButtonSelected)
				} else {
					typeBtn.Styler(appstyles.StyleObjectTypeButtonUnselected)
				}

				capturedType := containerType
				typeBtn.OnClick(func(e events.Event) {
					selectedType = capturedType

					// Update all button styles to reflect new selection
					for i, btnRef := range typeButtons {
						if containerTypes[i] == selectedType {
							btnRef.Styler(appstyles.StyleObjectTypeButtonSelected)
						} else {
							btnRef.Styler(appstyles.StyleObjectTypeButtonUnselected)
						}
					}

					// Force re-render of container
					typeContainer.Update()
				})
			}

			// Group selection dropdown
			if len(app.groups) > 0 {
				groupLabel := core.NewText(dialog).SetText("Assign to Group (optional)")
				groupLabel.Styler(appstyles.StyleGroupLabelWithMargin)

				groupDropdown := core.NewButton(dialog).SetText("Select Group").SetIcon(icons.ArrowDropDown)
				groupDropdown.Styler(appstyles.StyleGroupDropdownButtonGrow)
				groupDropdown.OnClick(func(e events.Event) {
					// Create menu with group options
					m := core.NewMenu(func(m *core.Scene) {
						// Add "No Group" option
						core.NewButton(m).SetText("No Group").OnClick(func(e events.Event) {
							selectedGroupID = nil
							groupDropdown.SetText("No Group")
							// Apply unselected styling
							groupDropdown.Styler(appstyles.StyleGroupDropdownButtonGrow)
							groupDropdown.Update()
						})

						// Add each group
						for _, group := range app.groups {
							groupCopy := group // Capture for closure
							core.NewButton(m).SetText(group.Name).OnClick(func(e events.Event) {
								selectedGroupID = &groupCopy.ID
								groupDropdown.SetText(groupCopy.Name)
								// Apply selected styling
								groupDropdown.Styler(appstyles.StyleGroupDropdownButtonSelectedGrow)
								groupDropdown.Update()
							})
						}
					}, groupDropdown, groupDropdown.ContextMenuPos(e))
					m.Run()
				})
			}

			// Show parent container info if creating sub-container
			if parentContainer != nil {
				parentInfo := core.NewText(dialog).SetText(fmt.Sprintf("Parent: %s", parentContainer.Name))
				parentInfo.Styler(func(s *styles.Style) {
					appstyles.StyleParentInfo(s)
					s.Margin.Top = units.Dp(12)
				})
			}
		},
		OnSubmit: func() {
			var parentID *string
			if parentContainer != nil {
				parentID = &parentContainer.ID
			}
			app.handleCreateContainerWithDetails(collection.ID, nameField.Text(), locationField.Text(), selectedType, parentID, selectedGroupID)
		},
	})
}

func (app *App) showEditContainerDialog(container types.Container, collection Collection) {
	var nameField, locationField *core.TextField
	selectedGroupID := container.GroupID
	selectedType := container.Type
	if selectedType == "" {
		selectedType = "general"
	}

	containerTypes := []string{"room", "bookshelf", "shelf", "binder", "cabinet", "general"}

	app.showDialog(DialogConfig{
		Title:             "Edit Container",
		SubmitButtonText:  "Save Changes",
		SubmitButtonStyle: appstyles.StyleButtonPrimary,
		ContentBuilder: func(dialog core.Widget, closeDialog func()) {
			nameField = createTextField(dialog, "Container name")
			nameField.SetText(container.Name)

			locationField = createTextField(dialog, "Location (optional)")
			locationField.SetText(container.Location)

			// Container type selection
			typeLabel := core.NewText(dialog).SetText("Container Type")
			typeLabel.Styler(appstyles.StyleFormLabel)

			typeContainer := createFlexRow(dialog, 8, styles.Start)
			typeContainer.Styler(appstyles.StyleTypeButtonContainer)

			// Store button references for dynamic styling updates
			typeButtons := make([]*core.Button, 0, len(containerTypes))

			for _, containerType := range containerTypes {
				typeBtn := core.NewButton(typeContainer).SetText(strings.Title(containerType))
				typeButtons = append(typeButtons, typeBtn)

				// Apply initial styling
				if selectedType == containerType {
					typeBtn.Styler(appstyles.StyleObjectTypeButtonSelected)
				} else {
					typeBtn.Styler(appstyles.StyleObjectTypeButtonUnselected)
				}

				capturedType := containerType
				typeBtn.OnClick(func(e events.Event) {
					selectedType = capturedType

					// Update all button styles to reflect new selection
					for i, btnRef := range typeButtons {
						if containerTypes[i] == selectedType {
							btnRef.Styler(appstyles.StyleObjectTypeButtonSelected)
						} else {
							btnRef.Styler(appstyles.StyleObjectTypeButtonUnselected)
						}
					}

					// Force re-render of container
					typeContainer.Update()
				})
			}

			// Group selection dropdown
			if len(app.groups) > 0 {
				groupLabel := core.NewText(dialog).SetText("Assign to Group (optional)")
				groupLabel.Styler(appstyles.StyleFormLabel)

				initialGroupText := "Select Group"
				hasInitialGroup := false
				if selectedGroupID != nil {
					for _, g := range app.groups {
						if g.ID == *selectedGroupID {
							initialGroupText = g.Name
							hasInitialGroup = true
							break
						}
					}
				}

				groupDropdown := core.NewButton(dialog).SetText(initialGroupText).SetIcon(icons.ArrowDropDown)
				// Apply initial styling based on whether a group is already selected
				if hasInitialGroup {
					groupDropdown.Styler(appstyles.StyleGroupDropdownButtonSelectedGrow)
				} else {
					groupDropdown.Styler(appstyles.StyleGroupDropdownButtonGrow)
				}

				groupDropdown.OnClick(func(e events.Event) {
					// Create menu with group options
					m := core.NewMenu(func(m *core.Scene) {
						// Add "No Group" option
						core.NewButton(m).SetText("No Group").OnClick(func(e events.Event) {
							selectedGroupID = nil
							groupDropdown.SetText("No Group")
							// Apply unselected styling
							groupDropdown.Styler(appstyles.StyleGroupDropdownButtonGrow)
							groupDropdown.Update()
						})

						// Add each group
						for _, group := range app.groups {
							groupCopy := group // Capture for closure
							core.NewButton(m).SetText(group.Name).OnClick(func(e events.Event) {
								selectedGroupID = &groupCopy.ID
								groupDropdown.SetText(groupCopy.Name)
								// Apply selected styling
								groupDropdown.Styler(appstyles.StyleGroupDropdownButtonSelectedGrow)
								groupDropdown.Update()
							})
						}
					}, groupDropdown, groupDropdown.ContextMenuPos(e))
					m.Run()
				})
			}
		},
		OnSubmit: func() {
			app.handleEditContainerWithDetails(collection.ID, container.ID, nameField.Text(), locationField.Text(), selectedType, selectedGroupID)
		},
	})
}

func (app *App) showDeleteContainerDialog(container types.Container, collection Collection) {
	app.showDialog(DialogConfig{
		Title:             "Delete Container",
		Message:           fmt.Sprintf("Are you sure you want to delete \"%s\"? This will also delete all objects within it. This action cannot be undone.", container.Name),
		SubmitButtonText:  "Delete",
		SubmitButtonStyle: appstyles.StyleButtonDanger,
		OnSubmit: func() {
			app.handleDeleteContainer(container, collection)
		},
	})
}

func (app *App) handleDeleteContainer(container types.Container, collection Collection) {
	app.logger.Info("Deleting container", "container_id", container.ID, "collection_id", collection.ID)

	// Make API call
	err := app.containersClient.Delete(app.currentUser.ID, collection.ID, container.ID)
	if err != nil {
		app.logger.Error("Failed to delete container", "error", err)
		core.ErrorSnackbar(app.body, err, "Failed to Delete Container")
		return
	}

	app.logger.Info("Container deleted successfully", "container_id", container.ID)
	core.MessageSnackbar(app.body, "Container deleted successfully")

	// Remove the container from the local collection's Containers array
	updatedContainers := make([]types.Container, 0, len(collection.Containers)-1)
	for _, c := range collection.Containers {
		if c.ID != container.ID {
			updatedContainers = append(updatedContainers, c)
		}
	}

	// Update the collection with the filtered containers
	collection.Containers = updatedContainers

	// Update selected collection if it matches
	if app.selectedCollection != nil && app.selectedCollection.ID == collection.ID {
		app.selectedCollection.Containers = updatedContainers
	}

	app.logger.Info("Updated local collection state", "containers_count", len(collection.Containers))

	// Re-render the collection view with updated local data
	app.showCollectionDetailView(collection)
}
