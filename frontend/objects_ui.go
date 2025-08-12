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

// Container Detail View with Objects Management
func (app *App) showContainerDetailView(container Container, collection Collection) {
	app.mainContainer.DeleteChildren()
	app.currentView = "container_detail"

	// Header with back button
	header := app.createHeader(container.Name, true)

	// Main content
	content := core.NewFrame(app.mainContainer)
	content.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Grow.Set(1, 1)
		s.Padding.Set(units.Dp(16))
		s.Gap.Set(units.Dp(16))
	})

	// Container info card
	infoCard := core.NewFrame(content)
	infoCard.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Background = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(16))
		s.Gap.Set(units.Dp(12))
	})

	// Breadcrumb navigation
	breadcrumb := core.NewFrame(infoCard)
	breadcrumb.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Gap.Set(units.Dp(8))
	})

	collectionLink := core.NewText(breadcrumb).SetText(collection.Name)
	collectionLink.Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(ColorPrimary)
		s.Cursor = cursors.Pointer
	})
	collectionLink.OnClick(func(e events.Event) {
		app.showCollectionDetailView(collection)
	})

	arrow := core.NewText(breadcrumb).SetText(">")
	arrow.Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(ColorGrayDark)
	})

	containerText := core.NewText(breadcrumb).SetText(container.Name)
	containerText.Styler(func(s *styles.Style) {
		s.Font.Weight = styles.WeightSemiBold
	})

	// Container description
	if container.Description != "" {
		desc := core.NewText(infoCard).SetText(container.Description)
		desc.Styler(func(s *styles.Style) {
			s.Color = colors.Uniform(ColorGrayDark)
		})
	}

	// Action buttons row
	actionsRow := core.NewFrame(content)
	actionsRow.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(units.Dp(12))
		s.Justify.Content = styles.SpaceBetween
	})

	// Search and filter section
	searchSection := core.NewFrame(actionsRow)
	searchSection.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(units.Dp(8))
		s.Align.Items = styles.Center
	})

	searchField := core.NewTextField(searchSection)
	searchField.SetPlaceholder("Search objects...")
	searchField.Styler(func(s *styles.Style) {
		s.Min.X.Set(200, units.UnitDp)
	})

	filterBtn := core.NewButton(searchSection).SetIcon(icons.FilterList)
	filterBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(color.RGBA{R: 240, G: 240, B: 240, A: 255})
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(units.Dp(8))
	})

	// Add object section
	addSection := core.NewFrame(actionsRow)
	addSection.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(units.Dp(8))
	})

	addObjectBtn := core.NewButton(addSection).SetText("Add Object").SetIcon(icons.Add)
	addObjectBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorPrimary)
		s.Color = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(12), units.Dp(16))
		s.Gap.Set(units.Dp(8))
	})
	addObjectBtn.OnClick(func(e events.Event) {
		app.showCreateObjectDialog(container, collection)
	})

	bulkImportBtn := core.NewButton(addSection).SetText("Bulk Import").SetIcon(icons.Upload)
	bulkImportBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorAccent)
		s.Color = colors.Uniform(ColorBlack)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(12), units.Dp(16))
		s.Gap.Set(units.Dp(8))
	})

	// Objects section
	objectsTitle := core.NewText(content).SetText("Objects")
	objectsTitle.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(18)
		s.Font.Weight = styles.WeightSemiBold
	})

	// Mock objects for demonstration
	objects := []Object{
		{
			ID:          "1",
			Name:        "Organic Bananas",
			Description: "Fresh organic bananas from Ecuador",
			ContainerID: container.ID,
			Properties: map[string]interface{}{
				"expiry_date": "2024-02-15",
				"quantity":    "6 pieces",
				"brand":       "Organic Valley",
			},
			Tags: []string{"organic", "fruit", "healthy"},
		},
		{
			ID:          "2",
			Name:        "Whole Milk",
			Description: "Fresh whole milk",
			ContainerID: container.ID,
			Properties: map[string]interface{}{
				"expiry_date": "2024-02-10",
				"quantity":    "1 gallon",
				"brand":       "Local Dairy",
				"fat_content": "3.25%",
			},
			Tags: []string{"dairy", "organic"},
		},
	}

	if len(objects) == 0 {
		emptyState := app.createEmptyState(content, "No objects found", "Add objects to this container to start tracking your inventory!", icons.Inventory)
		_ = emptyState
	} else {
		// Objects view mode selector
		viewModeRow := core.NewFrame(content)
		viewModeRow.Styler(func(s *styles.Style) {
			s.Direction = styles.Row
			s.Gap.Set(units.Dp(8))
			s.Justify.Content = styles.End
		})

		gridViewBtn := core.NewButton(viewModeRow).SetIcon(icons.GridView)
		gridViewBtn.Styler(func(s *styles.Style) {
			s.Background = colors.Uniform(ColorPrimary)
			s.Color = colors.Uniform(ColorWhite)
			s.Border.Radius = styles.BorderRadiusMedium
			s.Padding.Set(units.Dp(8))
		})

		listViewBtn := core.NewButton(viewModeRow).SetIcon(icons.ViewList)
		listViewBtn.Styler(func(s *styles.Style) {
			s.Background = colors.Uniform(color.RGBA{R: 240, G: 240, B: 240, A: 255})
			s.Border.Radius = styles.BorderRadiusMedium
			s.Padding.Set(units.Dp(8))
		})

		// Objects grid
		objectsGrid := core.NewFrame(content)
		objectsGrid.Styler(func(s *styles.Style) {
			s.Direction = styles.Row
			s.Wrap = true
			s.Gap.Set(units.Dp(16))
		})

		for _, object := range objects {
			app.createObjectCard(objectsGrid, object, container, collection)
		}
	}

	_ = header
	app.mainContainer.Update()
}

// Create object card for container view
func (app *App) createObjectCard(parent core.Widget, object Object, container Container, collection Collection) *core.Frame {
	card := core.NewFrame(parent)
	card.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Background = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(16))
		s.Border.Style.Set(styles.BorderSolid)
		s.Border.Width.Set(units.Dp(1))
		s.Border.Color.Set(colors.Uniform(ColorGrayLight))
		s.Min.X.Set(250, units.UnitDp)
		s.Max.X.Set(300, units.UnitDp)
		s.Gap.Set(units.Dp(12))
		s.Cursor = cursors.Pointer
	})

	// Header with name and actions
	cardHeader := core.NewFrame(card)
	cardHeader.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Justify.Content = styles.SpaceBetween
	})

	// Object name and type icon
	nameSection := core.NewFrame(cardHeader)
	nameSection.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Gap.Set(units.Dp(8))
		s.Grow.Set(1, 0)
	})
	nameSection.OnClick(func(e events.Event) {
		app.showObjectDetailView(object, container, collection)
	})

	objectIcon := core.NewIcon(nameSection).SetIcon(app.getIcon(collection.ObjectType))
	objectIcon.Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(app.getCollectionTypeColor(collection.ObjectType))
	})

	objectName := core.NewText(nameSection).SetText(object.Name)
	objectName.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(16)
		s.Font.Weight = styles.WeightSemiBold
		s.Color = colors.Uniform(ColorBlack)
	})

	// Actions menu
	actionsMenu := core.NewFrame(cardHeader)
	actionsMenu.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(units.Dp(4))
	})

	editBtn := core.NewButton(actionsMenu).SetIcon(icons.Edit)
	editBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorAccent)
		s.Color = colors.Uniform(ColorBlack)
		s.Border.Radius = styles.BorderRadiusFull
		s.Padding.Set(units.Dp(6))
	})
	editBtn.OnClick(func(e events.Event) {
		app.showEditObjectDialog(object, container, collection)
	})

	deleteBtn := core.NewButton(actionsMenu).SetIcon(icons.Delete)
	deleteBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorDanger)
		s.Color = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusFull
		s.Padding.Set(units.Dp(6))
	})
	deleteBtn.OnClick(func(e events.Event) {
		app.showDeleteObjectDialog(object, container, collection)
	})

	// Description
	if object.Description != "" {
		desc := core.NewText(card).SetText(object.Description)
		desc.Styler(func(s *styles.Style) {
			s.Font.Size = units.Dp(14)
			s.Color = colors.Uniform(ColorGrayDark)
		})
	}

	// Properties
	if len(object.Properties) > 0 {
		propsContainer := core.NewFrame(card)
		propsContainer.Styler(func(s *styles.Style) {
			s.Direction = styles.Column
			s.Gap.Set(units.Dp(4))
		})

		propsTitle := core.NewText(propsContainer).SetText("Properties")
		propsTitle.Styler(func(s *styles.Style) {
			s.Font.Size = units.Dp(12)
			s.Font.Weight = styles.WeightSemiBold
			s.Color = colors.Uniform(ColorGrayDark)
		})

		// Show first few properties
		count := 0
		for key, value := range object.Properties {
			if count >= 3 {
				break
			}

			propRow := core.NewFrame(propsContainer)
			propRow.Styler(func(s *styles.Style) {
				s.Direction = styles.Row
				s.Justify.Content = styles.SpaceBetween
			})

			propKey := core.NewText(propRow).SetText(strings.Title(strings.ReplaceAll(key, "_", " ")) + ":")
			propKey.Styler(func(s *styles.Style) {
				s.Font.Size = units.Dp(12)
				s.Color = colors.Uniform(ColorGrayDark)
			})

			propValue := core.NewText(propRow).SetText(fmt.Sprintf("%v", value))
			propValue.Styler(func(s *styles.Style) {
				s.Font.Size = units.Dp(12)
				s.Font.Weight = styles.WeightMedium
			})

			count++
		}

		if len(object.Properties) > 3 {
			moreText := core.NewText(propsContainer).SetText(fmt.Sprintf("... %d more", len(object.Properties)-3))
			moreText.Styler(func(s *styles.Style) {
				s.Font.Size = units.Dp(10)
				s.Color = colors.Uniform(ColorGrayDark)
			})
		}
	}

	// Tags
	if len(object.Tags) > 0 {
		tagsContainer := core.NewFrame(card)
		tagsContainer.Styler(func(s *styles.Style) {
			s.Direction = styles.Row
			s.Wrap = true
			s.Gap.Set(units.Dp(4))
		})

		for i, tag := range object.Tags {
			if i >= 3 {
				break
			}

			tagBadge := core.NewText(tagsContainer).SetText(tag)
			tagBadge.Styler(func(s *styles.Style) {
				s.Font.Size = units.Dp(10)
				s.Background = colors.Uniform(ColorToBeFixed(230, 247, 245))
				s.Color = colors.Uniform(ColorPrimary)
				s.Padding.Set(units.Dp(4), units.Dp(8))
				s.Border.Radius = styles.BorderRadiusFull
			})
		}

		if len(object.Tags) > 3 {
			moreTags := core.NewText(tagsContainer).SetText(fmt.Sprintf("+%d", len(object.Tags)-3))
			moreTags.Styler(func(s *styles.Style) {
				s.Font.Size = units.Dp(10)
				s.Background = colors.Uniform(color.RGBA{R: 240, G: 240, B: 240, A: 255})
				s.Color = colors.Uniform(ColorGrayDark)
				s.Padding.Set(units.Dp(4), units.Dp(8))
				s.Border.Radius = styles.BorderRadiusFull
			})
		}
	}

	return card
}

// Object Detail View
func (app *App) showObjectDetailView(object Object, container Container, collection Collection) {
	app.mainContainer.DeleteChildren()
	app.currentView = "object_detail"

	// Header with back button
	header := app.createHeader(object.Name, true)

	// Main content
	content := core.NewFrame(app.mainContainer)
	content.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Grow.Set(1, 1)
		s.Padding.Set(units.Dp(16))
		s.Gap.Set(units.Dp(16))
	})

	// Breadcrumb navigation
	breadcrumbCard := core.NewFrame(content)
	breadcrumbCard.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(16))
	})

	breadcrumb := core.NewFrame(breadcrumbCard)
	breadcrumb.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Gap.Set(units.Dp(8))
	})

	collectionLink := core.NewText(breadcrumb).SetText(collection.Name)
	collectionLink.Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(ColorPrimary)
		s.Cursor = cursors.Pointer
	})
	collectionLink.OnClick(func(e events.Event) {
		app.showCollectionDetailView(collection)
	})

	arrow1 := core.NewText(breadcrumb).SetText(">")
	arrow1.Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(ColorGrayDark)
	})

	containerLink := core.NewText(breadcrumb).SetText(container.Name)
	containerLink.Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(ColorPrimary)
		s.Cursor = cursors.Pointer
	})
	containerLink.OnClick(func(e events.Event) {
		app.showContainerDetailView(container, collection)
	})

	arrow2 := core.NewText(breadcrumb).SetText(">")
	arrow2.Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(ColorGrayDark)
	})

	objectText := core.NewText(breadcrumb).SetText(object.Name)
	objectText.Styler(func(s *styles.Style) {
		s.Font.Weight = styles.WeightSemiBold
	})

	// Object info card
	infoCard := core.NewFrame(content)
	infoCard.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Background = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(16))
		s.Gap.Set(units.Dp(12))
	})

	// Object header with icon
	objectHeader := core.NewFrame(infoCard)
	objectHeader.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Gap.Set(units.Dp(12))
	})

	objectIcon := core.NewIcon(objectHeader).SetIcon(app.getIcon(collection.ObjectType))
	objectIcon.Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(app.getCollectionTypeColor(collection.ObjectType))
		s.Font.Size = units.Dp(24)
	})

	titleContainer := core.NewFrame(objectHeader)
	titleContainer.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(units.Dp(2))
	})

	objectName := core.NewText(titleContainer).SetText(object.Name)
	objectName.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(20)
		s.Font.Weight = styles.WeightBold
	})

	if object.Description != "" {
		objectDesc := core.NewText(titleContainer).SetText(object.Description)
		objectDesc.Styler(func(s *styles.Style) {
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

	editBtn := core.NewButton(actionsRow).SetText("Edit Object").SetIcon(icons.Edit)
	editBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorAccent)
		s.Color = colors.Uniform(ColorBlack)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(12), units.Dp(16))
		s.Gap.Set(units.Dp(8))
	})

	deleteBtn := core.NewButton(actionsRow).SetText("Delete Object").SetIcon(icons.Delete)
	deleteBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorDanger)
		s.Color = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(12), units.Dp(16))
		s.Gap.Set(units.Dp(8))
	})

	// Properties section
	if len(object.Properties) > 0 {
		propsCard := core.NewFrame(content)
		propsCard.Styler(func(s *styles.Style) {
			s.Direction = styles.Column
			s.Background = colors.Uniform(ColorWhite)
			s.Border.Radius = styles.BorderRadiusLarge
			s.Padding.Set(units.Dp(16))
			s.Gap.Set(units.Dp(12))
		})

		propsTitle := core.NewText(propsCard).SetText("Properties")
		propsTitle.Styler(func(s *styles.Style) {
			s.Font.Size = units.Dp(18)
			s.Font.Weight = styles.WeightSemiBold
		})

		for key, value := range object.Properties {
			propRow := core.NewFrame(propsCard)
			propRow.Styler(func(s *styles.Style) {
				s.Direction = styles.Row
				s.Justify.Content = styles.SpaceBetween
				s.Padding.Set(units.Dp(8))
				s.Background = colors.Uniform(ColorToBeFixed(248, 248, 248))
				s.Border.Radius = styles.BorderRadiusMedium
			})

			propKey := core.NewText(propRow).SetText(strings.Title(strings.ReplaceAll(key, "_", " ")) + ":")
			propKey.Styler(func(s *styles.Style) {
				s.Font.Weight = styles.WeightMedium
			})

			propValue := core.NewText(propRow).SetText(fmt.Sprintf("%v", value))
			propValue.Styler(func(s *styles.Style) {
				s.Color = colors.Uniform(ColorGrayDark)
			})
		}
	}

	// Tags section
	if len(object.Tags) > 0 {
		tagsCard := core.NewFrame(content)
		tagsCard.Styler(func(s *styles.Style) {
			s.Direction = styles.Column
			s.Background = colors.Uniform(ColorWhite)
			s.Border.Radius = styles.BorderRadiusLarge
			s.Padding.Set(units.Dp(16))
			s.Gap.Set(units.Dp(12))
		})

		tagsTitle := core.NewText(tagsCard).SetText("Tags")
		tagsTitle.Styler(func(s *styles.Style) {
			s.Font.Size = units.Dp(18)
			s.Font.Weight = styles.WeightSemiBold
		})

		tagsContainer := core.NewFrame(tagsCard)
		tagsContainer.Styler(func(s *styles.Style) {
			s.Direction = styles.Row
			s.Wrap = true
			s.Gap.Set(units.Dp(8))
		})

		for _, tag := range object.Tags {
			tagBadge := core.NewText(tagsContainer).SetText(tag)
			tagBadge.Styler(func(s *styles.Style) {
				s.Font.Size = units.Dp(14)
				s.Background = colors.Uniform(ColorToBeFixed(230, 247, 245))
				s.Color = colors.Uniform(ColorPrimary)
				s.Padding.Set(units.Dp(8), units.Dp(16))
				s.Border.Radius = styles.BorderRadiusFull
			})
		}
	}

	_ = header
	app.mainContainer.Update()
}

// Object creation and editing dialogs
func (app *App) showCreateObjectDialog(container Container, collection Collection) {
	overlay := app.createOverlay()

	dialog := core.NewFrame(overlay)
	dialog.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(24))
		s.Gap.Set(units.Dp(16))
		s.Direction = styles.Column
		s.Min.X.Set(500, units.UnitDp)
		s.Max.X.Set(600, units.UnitDp)
		s.Max.Y.Set(500, units.UnitDp)
	})

	title := core.NewText(dialog).SetText("Add New Object")
	title.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(20)
		s.Font.Weight = styles.WeightSemiBold
	})

	// Basic fields
	nameField := core.NewTextField(dialog)
	nameField.SetText("").SetPlaceholder("Object name")

	descField := core.NewTextField(dialog)
	descField.SetText("").SetPlaceholder("Description (optional)")

	// Properties section
	propsTitle := core.NewText(dialog).SetText("Properties")
	propsTitle.Styler(func(s *styles.Style) {
		s.Font.Weight = styles.WeightSemiBold
	})

	// Create property fields based on object type
	propsContainer := core.NewFrame(dialog)
	propsContainer.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(units.Dp(8))
		s.Background = colors.Uniform(ColorToBeFixed(248, 248, 248))
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(units.Dp(12))
	})

	app.createObjectTypeProperties(propsContainer, collection.ObjectType)

	// Tags section
	tagsTitle := core.NewText(dialog).SetText("Tags")
	tagsTitle.Styler(func(s *styles.Style) {
		s.Font.Weight = styles.WeightSemiBold
	})

	tagsField := core.NewTextField(dialog)
	tagsField.SetText("").SetPlaceholder("Tags (comma-separated)")

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

	addBtn := core.NewButton(buttonRow).SetText("Add Object")
	addBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorPrimary)
		s.Color = colors.Uniform(ColorWhite)
	})
	addBtn.OnClick(func(e events.Event) {
		app.handleCreateObject(nameField.Text(), descField.Text(), tagsField.Text(), container, collection)
	})

	app.showOverlay(overlay)
}

// Create property fields based on object type
func (app *App) createObjectTypeProperties(parent core.Widget, objectType string) {
	switch strings.ToLower(objectType) {
	case "food":
		app.createFoodProperties(parent)
	case "book":
		app.createBookProperties(parent)
	case "videogame":
		app.createVideoGameProperties(parent)
	case "music":
		app.createMusicProperties(parent)
	case "boardgame":
		app.createBoardGameProperties(parent)
	default:
		app.createGeneralProperties(parent)
	}
}

func (app *App) createFoodProperties(parent core.Widget) {
	expiryField := core.NewTextField(parent)
	expiryField.SetPlaceholder("Expiry date (YYYY-MM-DD)")

	quantityField := core.NewTextField(parent)
	quantityField.SetPlaceholder("Quantity")

	brandField := core.NewTextField(parent)
	brandField.SetPlaceholder("Brand (optional)")
}

func (app *App) createBookProperties(parent core.Widget) {
	authorField := core.NewTextField(parent)
	authorField.SetPlaceholder("Author")

	isbnField := core.NewTextField(parent)
	isbnField.SetPlaceholder("ISBN (optional)")

	genreField := core.NewTextField(parent)
	genreField.SetPlaceholder("Genre (optional)")

	yearField := core.NewTextField(parent)
	yearField.SetPlaceholder("Publication year (optional)")
}

func (app *App) createVideoGameProperties(parent core.Widget) {
	platformField := core.NewTextField(parent)
	platformField.SetPlaceholder("Platform")

	genreField := core.NewTextField(parent)
	genreField.SetPlaceholder("Genre (optional)")

	ratingField := core.NewTextField(parent)
	ratingField.SetPlaceholder("Rating (optional)")
}

func (app *App) createMusicProperties(parent core.Widget) {
	artistField := core.NewTextField(parent)
	artistField.SetPlaceholder("Artist")

	albumField := core.NewTextField(parent)
	albumField.SetPlaceholder("Album (optional)")

	genreField := core.NewTextField(parent)
	genreField.SetPlaceholder("Genre (optional)")

	yearField := core.NewTextField(parent)
	yearField.SetPlaceholder("Release year (optional)")
}

func (app *App) createBoardGameProperties(parent core.Widget) {
	playersField := core.NewTextField(parent)
	playersField.SetPlaceholder("Number of players")

	ageField := core.NewTextField(parent)
	ageField.SetPlaceholder("Minimum age (optional)")

	durationField := core.NewTextField(parent)
	durationField.SetPlaceholder("Play duration (optional)")
}

func (app *App) createGeneralProperties(parent core.Widget) {
	prop1Field := core.NewTextField(parent)
	prop1Field.SetPlaceholder("Custom property 1")

	prop2Field := core.NewTextField(parent)
	prop2Field.SetPlaceholder("Custom property 2")
}

// Object handlers
func (app *App) handleCreateObject(name, description, tags string, container Container, collection Collection) {
	if strings.TrimSpace(name) == "" {
		return
	}

	fmt.Printf("Creating object: %s in container %s\n", name, container.Name)

	app.hideOverlay()
	// Refresh the container view
	app.showContainerDetailView(container, collection)
}

func (app *App) showEditObjectDialog(object Object, container Container, collection Collection) {
	fmt.Printf("Edit object dialog for: %s\n", object.Name)
	// Implementation similar to create dialog but with pre-filled values
}

func (app *App) showDeleteObjectDialog(object Object, container Container, collection Collection) {
	fmt.Printf("Delete object dialog for: %s\n", object.Name)
	// Implementation similar to other delete dialogs
}
