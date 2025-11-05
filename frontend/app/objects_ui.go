//go:build js && wasm

package app

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"time"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/cursors"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"

	"github.com/nishiki/frontend/pkg/types"
	"github.com/nishiki/frontend/ui/components"
	"github.com/nishiki/frontend/ui/layouts"
	appstyles "github.com/nishiki/frontend/ui/styles"
)

// Container Detail View with Objects Management
func (app *App) showContainerDetailView(container Container, collection Collection) {
	app.mainContainer.DeleteChildren()
	app.currentView = "container_detail"

	// Header with back button
	layouts.SimpleHeader(app.mainContainer, container.Name, true, func() {
		app.showCollectionDetailView(collection)
	})

	// Main content
	content := core.NewFrame(app.mainContainer)
	content.Styler(appstyles.StyleContentColumn)

	// Container info card
	infoCard := core.NewFrame(content)
	infoCard.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Background = colors.Uniform(appstyles.ColorWhite)
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
		s.Color = colors.Uniform(appstyles.ColorPrimary)
		s.Cursor = cursors.Pointer
	})
	collectionLink.OnClick(func(e events.Event) {
		app.showCollectionDetailView(collection)
	})

	arrow := core.NewText(breadcrumb).SetText(">")
	arrow.Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(appstyles.ColorGrayDark)
	})

	containerText := core.NewText(breadcrumb).SetText(container.Name)
	containerText.Styler(func(s *styles.Style) {
		s.Font.Weight = appstyles.WeightSemiBold
	})

	// Container location
	if container.Location != "" {
		desc := core.NewText(infoCard).SetText("Location: " + container.Location)
		desc.Styler(func(s *styles.Style) {
			s.Color = colors.Uniform(appstyles.ColorGrayDark)
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

	_ = createSearchField(searchSection, "Search objects...") // TODO: Wire up search functionality

	filterBtn := core.NewButton(searchSection).SetIcon(icons.FilterList)
	filterBtn.Styler(appstyles.StyleFilterButton)

	// Add object section
	addSection := core.NewFrame(actionsRow)
	addSection.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(units.Dp(8))
	})

	addObjectBtn := core.NewButton(addSection).SetText("Add Object").SetIcon(icons.Add)
	addObjectBtn.Styler(appstyles.StyleButtonPrimary)
	addObjectBtn.OnClick(func(e events.Event) {
		app.showCreateObjectDialog(container, collection)
	})

	bulkImportBtn := core.NewButton(addSection).SetText("Bulk Import").SetIcon(icons.Upload)
	bulkImportBtn.Styler(appstyles.StyleButtonAccent)
	bulkImportBtn.OnClick(func(e events.Event) {
		app.ShowImportDialog(container.ID, collection.ID)
	})

	// Objects section
	objectsTitle := core.NewText(content).SetText("Objects")
	objectsTitle.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(18)
		s.Font.Weight = appstyles.WeightSemiBold
	})

	// Use container's objects
	objects := container.Objects

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
			s.Background = colors.Uniform(appstyles.ColorPrimary)
			s.Color = colors.Uniform(appstyles.ColorWhite)
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

	app.mainContainer.Update()
}

// Create object card for container view
func (app *App) createObjectCard(parent core.Widget, object Object, container Container, collection Collection) *core.Frame {
	typeIcon := app.getIcon(collection.ObjectType)
	typeColor := app.getCollectionTypeColor(collection.ObjectType)

	return app.createCard(parent, CardConfig{
		Icon:        typeIcon,
		IconColor:   typeColor,
		Title:       object.Name,
		Description: object.Description,
		OnClick: func() {
			app.showObjectDetailView(object, container, collection)
		},
		Actions: []CardAction{
			{Icon: icons.Edit, Color: appstyles.ColorAccent, Tooltip: "Edit object", OnClick: func() {
				app.showEditObjectDialog(object, container, collection)
			}},
			{Icon: icons.Delete, Color: appstyles.ColorDanger, Tooltip: "Delete object", OnClick: func() {
				app.showDeleteObjectDialog(object, container, collection)
			}},
		},
		Content: func(card core.Widget) {
			// Properties section (show first 3)
			if len(object.Properties) > 0 {
				propsContainer := core.NewFrame(card)
				propsContainer.Styler(func(s *styles.Style) {
					s.Direction = styles.Column
					s.Gap.Set(units.Dp(4))
				})

				propsTitle := core.NewText(propsContainer).SetText("Properties")
				propsTitle.Styler(func(s *styles.Style) {
					s.Font.Size = units.Dp(12)
					s.Font.Weight = appstyles.WeightSemiBold
					s.Color = colors.Uniform(appstyles.ColorBlack)
				})

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
						s.Color = colors.Uniform(appstyles.ColorBlack) // Dark text for better readability
					})

					propValue := core.NewText(propRow).SetText(fmt.Sprintf("%v", value))
					propValue.Styler(func(s *styles.Style) {
						s.Font.Size = units.Dp(12)
						s.Font.Weight = appstyles.WeightMedium
						s.Color = colors.Uniform(appstyles.ColorBlack) // Dark text for better readability
					})

					count++
				}

				if len(object.Properties) > 3 {
					moreText := core.NewText(propsContainer).SetText(fmt.Sprintf("... %d more", len(object.Properties)-3))
					moreText.Styler(func(s *styles.Style) {
						s.Font.Size = units.Dp(10)
						s.Color = colors.Uniform(appstyles.ColorGrayDark)
					})
				}
			}

			// Tags section (show first 3)
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

					components.TagBadge(tagsContainer, tag)
				}

				if len(object.Tags) > 3 {
					badge := core.NewFrame(tagsContainer)
					badge.Styler(appstyles.StyleTagBadgeSecondary)
					core.NewText(badge).SetText(fmt.Sprintf("+%d", len(object.Tags)-3))
				}
			}
		},
	})
}

// Object Detail View
func (app *App) showObjectDetailView(object Object, container Container, collection Collection) {
	app.mainContainer.DeleteChildren()
	app.currentView = "object_detail"

	// Header with back button
	layouts.SimpleHeader(app.mainContainer, object.Name, true, func() {
		app.showContainerDetailView(container, collection)
	})

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
		s.Background = colors.Uniform(appstyles.ColorWhite)
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
		s.Color = colors.Uniform(appstyles.ColorPrimary)
		s.Cursor = cursors.Pointer
	})
	collectionLink.OnClick(func(e events.Event) {
		app.showCollectionDetailView(collection)
	})

	arrow1 := core.NewText(breadcrumb).SetText(">")
	arrow1.Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(appstyles.ColorGrayDark)
	})

	containerLink := core.NewText(breadcrumb).SetText(container.Name)
	containerLink.Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(appstyles.ColorPrimary)
		s.Cursor = cursors.Pointer
	})
	containerLink.OnClick(func(e events.Event) {
		app.showContainerDetailView(container, collection)
	})

	arrow2 := core.NewText(breadcrumb).SetText(">")
	arrow2.Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(appstyles.ColorGrayDark)
	})

	objectText := core.NewText(breadcrumb).SetText(object.Name)
	objectText.Styler(func(s *styles.Style) {
		s.Font.Weight = appstyles.WeightSemiBold
	})

	// Object info card
	infoCard := core.NewFrame(content)
	infoCard.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Background = colors.Uniform(appstyles.ColorWhite)
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

	// Object name is already shown in the header, so only show description here
	if object.Description != "" {
		objectDesc := core.NewText(titleContainer).SetText(object.Description)
		objectDesc.Styler(func(s *styles.Style) {
			s.Color = colors.Uniform(appstyles.ColorBlack) // Better contrast than gray
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
	editBtn.Styler(appstyles.StyleButtonAccent)
	editBtn.OnClick(func(e events.Event) {
		app.showEditObjectDialog(object, container, collection)
	})

	deleteBtn := core.NewButton(actionsRow).SetText("Delete Object").SetIcon(icons.Delete)
	deleteBtn.Styler(appstyles.StyleButtonDanger)
	deleteBtn.OnClick(func(e events.Event) {
		app.showDeleteObjectDialog(object, container, collection)
	})

	// Properties section
	if len(object.Properties) > 0 {
		propsCard := core.NewFrame(content)
		propsCard.Styler(func(s *styles.Style) {
			s.Direction = styles.Column
			s.Background = colors.Uniform(appstyles.ColorWhite)
			s.Border.Radius = styles.BorderRadiusLarge
			s.Padding.Set(units.Dp(16))
			s.Gap.Set(units.Dp(12))
		})

		propsTitle := core.NewText(propsCard).SetText("Properties")
		propsTitle.Styler(func(s *styles.Style) {
			s.Font.Size = units.Dp(18)
			s.Font.Weight = appstyles.WeightSemiBold
		})

		for key, value := range object.Properties {
			propRow := core.NewFrame(propsCard)
			propRow.Styler(func(s *styles.Style) {
				s.Direction = styles.Row
				s.Justify.Content = styles.SpaceBetween
				s.Padding.Set(units.Dp(8))
				s.Background = colors.Uniform(appstyles.ColorGrayLightest)
				s.Border.Radius = styles.BorderRadiusMedium
			})

			propKey := core.NewText(propRow).SetText(strings.Title(strings.ReplaceAll(key, "_", " ")) + ":")
			propKey.Styler(func(s *styles.Style) {
				s.Font.Weight = appstyles.WeightMedium
				s.Color = colors.Uniform(appstyles.ColorBlack) // Dark text for better readability
			})

			propValue := core.NewText(propRow).SetText(fmt.Sprintf("%v", value))
			propValue.Styler(func(s *styles.Style) {
				s.Color = colors.Uniform(appstyles.ColorBlack) // Dark text for better readability
			})
		}
	}

	// Tags section
	if len(object.Tags) > 0 {
		tagsCard := core.NewFrame(content)
		tagsCard.Styler(func(s *styles.Style) {
			s.Direction = styles.Column
			s.Background = colors.Uniform(appstyles.ColorWhite)
			s.Border.Radius = styles.BorderRadiusLarge
			s.Padding.Set(units.Dp(16))
			s.Gap.Set(units.Dp(12))
		})

		tagsTitle := core.NewText(tagsCard).SetText("Tags")
		tagsTitle.Styler(func(s *styles.Style) {
			s.Font.Size = units.Dp(18)
			s.Font.Weight = appstyles.WeightSemiBold
		})

		tagsContainer := core.NewFrame(tagsCard)
		tagsContainer.Styler(func(s *styles.Style) {
			s.Direction = styles.Row
			s.Wrap = true
			s.Gap.Set(units.Dp(8))
		})

		for _, tag := range object.Tags {
			badge := components.TagBadge(tagsContainer, tag)
			// Override font size and padding for detail view
			badge.Styler(func(s *styles.Style) {
				s.Font.Size = units.Dp(14)
				s.Padding.Set(units.Dp(8), units.Dp(16))
			})
		}
	}

	app.mainContainer.Update()
}

// Object creation and editing dialogs
func (app *App) showCreateObjectDialog(container Container, collection Collection) {
	var nameField, descField, quantityField, unitField, expiresAtField, tagsField *core.TextField
	var propertyFields map[string]*core.TextField

	app.showDialog(DialogConfig{
		Title:            "Add New Object",
		SubmitButtonText: "Add Object",
		ContentBuilder: func(dialog core.Widget, closeDialog func()) {
			// Basic fields
			nameField = createTextField(dialog, "Object name")
			descField = createTextField(dialog, "Description (optional)")

			// Structured fields section
			structuredTitle := core.NewText(dialog).SetText("Quantity & Expiration")
			structuredTitle.Styler(appstyles.StyleFormLabel)

			quantityField = createTextField(dialog, "Quantity (optional)")
			unitField = createTextField(dialog, "Unit (e.g., kg, pieces, liters)")
			expiresAtField = createTextField(dialog, "Expires at (YYYY-MM-DD, optional)")

			// Properties section
			propsTitle := core.NewText(dialog).SetText("Additional Properties")
			propsTitle.Styler(appstyles.StyleFormLabel)

			// Create property fields based on object type
			propsContainer := core.NewFrame(dialog)
			propsContainer.Styler(appstyles.StylePropertiesContainer)

			propertyFields = app.createObjectTypeProperties(propsContainer, collection.ObjectType)

			// Tags section
			tagsTitle := core.NewText(dialog).SetText("Tags")
			tagsTitle.Styler(appstyles.StyleFormLabel)

			tagsField = createTextField(dialog, "Tags (comma-separated)")
		},
		OnSubmit: func() {
			app.handleCreateObject(
				nameField.Text(),
				descField.Text(),
				quantityField.Text(),
				unitField.Text(),
				expiresAtField.Text(),
				tagsField.Text(),
				propertyFields,
				container,
				collection,
			)
		},
	})
}

// Create property fields based on object type
func (app *App) createObjectTypeProperties(parent core.Widget, objectType string) map[string]*core.TextField {
	switch strings.ToLower(objectType) {
	case "food":
		return app.createFoodProperties(parent)
	case "book":
		return app.createBookProperties(parent)
	case "videogame":
		return app.createVideoGameProperties(parent)
	case "music":
		return app.createMusicProperties(parent)
	case "boardgame":
		return app.createBoardGameProperties(parent)
	default:
		return app.createGeneralProperties(parent)
	}
}

func (app *App) createFoodProperties(parent core.Widget) map[string]*core.TextField {
	fields := make(map[string]*core.TextField)

	fields["brand"] = createTextField(parent, "Brand (optional)")

	return fields
}

func (app *App) createBookProperties(parent core.Widget) map[string]*core.TextField {
	fields := make(map[string]*core.TextField)

	fields["author"] = createTextField(parent, "Author")
	fields["isbn"] = createTextField(parent, "ISBN (optional)")
	fields["genre"] = createTextField(parent, "Genre (optional)")
	fields["year"] = createTextField(parent, "Publication year (optional)")

	return fields
}

func (app *App) createVideoGameProperties(parent core.Widget) map[string]*core.TextField {
	fields := make(map[string]*core.TextField)

	fields["platform"] = createTextField(parent, "Platform")
	fields["genre"] = createTextField(parent, "Genre (optional)")
	fields["rating"] = createTextField(parent, "Rating (optional)")

	return fields
}

func (app *App) createMusicProperties(parent core.Widget) map[string]*core.TextField {
	fields := make(map[string]*core.TextField)

	fields["artist"] = createTextField(parent, "Artist")
	fields["album"] = createTextField(parent, "Album (optional)")
	fields["genre"] = createTextField(parent, "Genre (optional)")
	fields["year"] = createTextField(parent, "Release year (optional)")

	return fields
}

func (app *App) createBoardGameProperties(parent core.Widget) map[string]*core.TextField {
	fields := make(map[string]*core.TextField)

	fields["players"] = createTextField(parent, "Number of players")
	fields["age"] = createTextField(parent, "Minimum age (optional)")
	fields["duration"] = createTextField(parent, "Play duration (optional)")

	return fields
}

func (app *App) createGeneralProperties(parent core.Widget) map[string]*core.TextField {
	fields := make(map[string]*core.TextField)

	fields["property1"] = createTextField(parent, "Custom property 1")
	fields["property2"] = createTextField(parent, "Custom property 2")

	return fields
}

// Object handlers
func (app *App) handleCreateObject(
	name, description, quantityStr, unit, expiresAtStr, tagsStr string,
	propertyFields map[string]*core.TextField,
	container Container,
	collection Collection,
) {
	if strings.TrimSpace(name) == "" {
		app.logger.Error("Object name cannot be empty")
		return
	}

	// Parse quantity
	var quantity *float64
	if strings.TrimSpace(quantityStr) != "" {
		if q, err := strconv.ParseFloat(quantityStr, 64); err == nil {
			quantity = &q
		} else {
			app.logger.Warn("Invalid quantity format", "error", err)
		}
	}

	// Parse expires at
	var expiresAt *time.Time
	if strings.TrimSpace(expiresAtStr) != "" {
		if t, err := time.Parse("2006-01-02", expiresAtStr); err == nil {
			expiresAt = &t
		} else {
			app.logger.Warn("Invalid expires_at format", "error", err)
		}
	}

	// Parse tags
	var tags []string
	if strings.TrimSpace(tagsStr) != "" {
		for _, tag := range strings.Split(tagsStr, ",") {
			tags = append(tags, strings.TrimSpace(tag))
		}
	}

	// Build properties map from property fields
	properties := make(map[string]interface{})
	for key, field := range propertyFields {
		if field != nil && strings.TrimSpace(field.Text()) != "" {
			properties[key] = field.Text()
		}
	}

	// Create request
	req := types.CreateObjectRequest{
		ContainerID: container.ID,
		Name:        name,
		Description: description,
		ObjectType:  collection.ObjectType,
		Quantity:    quantity,
		Unit:        unit,
		Properties:  properties,
		Tags:        tags,
		ExpiresAt:   expiresAt,
	}

	app.logger.Info("Creating object", "name", name, "container_id", container.ID)

	// Make API call
	object, err := app.objectsClient.Create(app.currentUser.ID, req)
	if err != nil {
		app.logger.Error("Failed to create object", "error", err)
		return
	}

	app.logger.Info("Object created successfully", "object_id", object.ID)

	// Refresh the container view
	app.showContainerDetailView(container, collection)
}

func (app *App) showEditObjectDialog(object Object, container Container, collection Collection) {
	var nameField, descField, quantityField, unitField, expiresAtField, tagsField *core.TextField
	var propertyFields map[string]*core.TextField

	app.showDialog(DialogConfig{
		Title:             "Edit Object",
		SubmitButtonText:  "Save Changes",
		SubmitButtonStyle: appstyles.StyleButtonPrimary,
		ContentBuilder: func(dialog core.Widget, closeDialog func()) {
			// Basic fields
			nameField = createTextField(dialog, "Object name")
			nameField.SetText(object.Name)

			descField = createTextField(dialog, "Description (optional)")
			descField.SetText(object.Description)

			// Structured fields section
			structuredTitle := core.NewText(dialog).SetText("Quantity & Expiration")
			structuredTitle.Styler(appstyles.StyleFormLabel)

			quantityField = createTextField(dialog, "Quantity (optional)")
			if object.Quantity != nil {
				quantityField.SetText(fmt.Sprintf("%v", *object.Quantity))
			}

			unitField = createTextField(dialog, "Unit (e.g., kg, pieces, liters)")
			unitField.SetText(object.Unit)

			expiresAtField = createTextField(dialog, "Expires at (YYYY-MM-DD, optional)")
			if object.ExpiresAt != nil {
				expiresAtField.SetText(object.ExpiresAt.Format("2006-01-02"))
			}

			// Properties section
			propsTitle := core.NewText(dialog).SetText("Additional Properties")
			propsTitle.Styler(appstyles.StyleFormLabel)

			// Create property fields based on object type
			propsContainer := core.NewFrame(dialog)
			propsContainer.Styler(appstyles.StylePropertiesContainer)

			propertyFields = app.createObjectTypeProperties(propsContainer, object.ObjectType)

			// Pre-fill property fields from existing object
			for key, field := range propertyFields {
				if val, exists := object.Properties[key]; exists {
					if strVal, ok := val.(string); ok {
						field.SetText(strVal)
					}
				}
			}

			// Tags section
			tagsTitle := core.NewText(dialog).SetText("Tags")
			tagsTitle.Styler(appstyles.StyleFormLabel)

			tagsField = createTextField(dialog, "Tags (comma-separated)")
			if len(object.Tags) > 0 {
				tagsField.SetText(strings.Join(object.Tags, ", "))
			}
		},
		OnSubmit: func() {
			app.handleEditObject(
				object.ID,
				nameField.Text(),
				descField.Text(),
				quantityField.Text(),
				unitField.Text(),
				expiresAtField.Text(),
				tagsField.Text(),
				propertyFields,
				container,
				collection,
			)
		},
	})
}

func (app *App) handleEditObject(
	objectID, name, description, quantityStr, unit, expiresAtStr, tagsStr string,
	propertyFields map[string]*core.TextField,
	container Container,
	collection Collection,
) {
	// Parse quantity
	var quantity *float64
	if strings.TrimSpace(quantityStr) != "" {
		if q, err := strconv.ParseFloat(quantityStr, 64); err == nil {
			quantity = &q
		}
	}

	// Parse expires at
	var expiresAt *time.Time
	if strings.TrimSpace(expiresAtStr) != "" {
		if t, err := time.Parse("2006-01-02", expiresAtStr); err == nil {
			expiresAt = &t
		}
	}

	// Parse tags
	var tags []string
	if strings.TrimSpace(tagsStr) != "" {
		for _, tag := range strings.Split(tagsStr, ",") {
			tags = append(tags, strings.TrimSpace(tag))
		}
	}

	// Build properties map
	properties := make(map[string]interface{})
	for key, field := range propertyFields {
		if field != nil && strings.TrimSpace(field.Text()) != "" {
			properties[key] = field.Text()
		}
	}

	// Create update request
	namePtr := &name
	descPtr := &description
	unitPtr := &unit

	req := types.UpdateObjectRequest{
		Name:        namePtr,
		Description: descPtr,
		Quantity:    quantity,
		Unit:        unitPtr,
		Properties:  properties,
		Tags:        tags,
		ExpiresAt:   expiresAt,
	}

	app.logger.Info("Updating object", "object_id", objectID)

	// Make API call
	updatedObject, err := app.objectsClient.Update(app.currentUser.ID, objectID, req)
	if err != nil {
		app.logger.Error("Failed to update object", "error", err)
		return
	}

	app.logger.Info("Object updated successfully", "object_id", updatedObject.ID)

	// Refresh the container view
	app.showContainerDetailView(container, collection)
}

func (app *App) showDeleteObjectDialog(object Object, container Container, collection Collection) {
	app.showDialog(DialogConfig{
		Title:             "Delete Object",
		Message:           fmt.Sprintf("Are you sure you want to delete '%s'? This action cannot be undone.", object.Name),
		SubmitButtonText:  "Delete",
		SubmitButtonStyle: appstyles.StyleButtonDanger,
		OnSubmit: func() {
			app.handleDeleteObject(object.ID, container, collection)
		},
	})
}

func (app *App) handleDeleteObject(objectID string, container Container, collection Collection) {
	app.logger.Info("Deleting object", "object_id", objectID, "container_id", container.ID)

	// Make API call with container ID
	err := app.objectsClient.Delete(app.currentUser.ID, objectID, container.ID)
	if err != nil {
		app.logger.Error("Failed to delete object", "error", err)
		core.ErrorSnackbar(app.body, err, "Failed to Delete Object")
		return
	}

	app.logger.Info("Object deleted successfully", "object_id", objectID)
	core.MessageSnackbar(app.body, "Object deleted successfully")

	// Remove the object from the local container's Objects array
	updatedObjects := make([]Object, 0, len(container.Objects)-1)
	for _, obj := range container.Objects {
		if obj.ID != objectID {
			updatedObjects = append(updatedObjects, obj)
		}
	}

	// Update the container with the filtered objects
	container.Objects = updatedObjects

	app.logger.Info("Updated local container state", "objects_count", len(container.Objects))

	// Re-render the container view with updated local data
	app.showContainerDetailView(container, collection)
}
