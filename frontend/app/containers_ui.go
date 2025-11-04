//go:build js && wasm

package app

import (
	"fmt"
	"image/color"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/cursors"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/sides"
	"cogentcore.org/core/styles/units"

	"github.com/nishiki/frontend/pkg/types"
	appstyles "github.com/nishiki/frontend/ui/styles"
)

// ContainerNode represents a node in the container hierarchy tree
type ContainerNode struct {
	Container *types.Container
	Children  []*ContainerNode
	Parent    *ContainerNode
	Expanded  bool
}

// BuildContainerHierarchy builds a hierarchical tree from a flat list of containers
func (app *App) BuildContainerHierarchy(containers []types.Container) []*ContainerNode {
	// Create lookup map
	nodeMap := make(map[string]*ContainerNode)
	roots := make([]*ContainerNode, 0)

	// First pass: create all nodes
	for i := range containers {
		node := &ContainerNode{
			Container: &containers[i],
			Children:  make([]*ContainerNode, 0),
			Expanded:  false,
		}
		nodeMap[containers[i].ID] = node
	}

	// Second pass: build hierarchy
	for _, node := range nodeMap {
		if node.Container.ParentContainerID == nil {
			// Root container
			roots = append(roots, node)
		} else {
			// Has parent - add to parent's children
			if parent, exists := nodeMap[*node.Container.ParentContainerID]; exists {
				node.Parent = parent
				parent.Children = append(parent.Children, node)
			} else {
				// Parent not found, treat as root
				roots = append(roots, node)
			}
		}
	}

	return roots
}

// GetContainerTypeIcon returns the appropriate icon for a container type
func GetContainerTypeIcon(containerType string) icons.Icon {
	switch containerType {
	case string(types.ContainerTypeRoom):
		return icons.Home
	case string(types.ContainerTypeBookshelf):
		return icons.MenuBook
	case string(types.ContainerTypeShelf):
		return icons.Bookmarks
	case string(types.ContainerTypeBinder):
		return icons.Folder
	case string(types.ContainerTypeCabinet):
		return icons.Inventory
	default:
		return icons.Folder
	}
}

// RenderContainerTree renders a hierarchical tree of containers
func (app *App) RenderContainerTree(parent core.Widget, nodes []*ContainerNode, level int) {
	for _, node := range nodes {
		app.RenderContainerTreeNode(parent, node, level)
	}
}

// RenderContainerTreeNode renders a single node in the container tree
func (app *App) RenderContainerTreeNode(parent core.Widget, node *ContainerNode, level int) {
	// Container row
	row := core.NewFrame(parent)
	row.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Gap.Set(units.Dp(8))
		s.Padding.Set(units.Dp(8), units.Dp(12))
		s.Padding.Left = units.Dp(12 + float32(level*24)) // Indent based on level
		s.Background = colors.Uniform(appstyles.ColorWhite)
		s.Border.Radius = sides.NewValues(units.Dp(appstyles.RadiusDefault))
		s.Margin.Bottom = units.Dp(4)
		s.Cursor = cursors.Pointer
	})

	// Expand/collapse button (only if has children)
	if len(node.Children) > 0 {
		expandBtn := core.NewButton(row)
		if node.Expanded {
			expandBtn.SetIcon(icons.ExpandMore)
		} else {
			expandBtn.SetIcon(icons.ChevronRight)
		}
		expandBtn.Styler(func(s *styles.Style) {
			s.Background = nil
			s.Padding.Set(units.Dp(4))
			s.Min.Set(units.Dp(24))
		})
		expandBtn.OnClick(func(e events.Event) {
			node.Expanded = !node.Expanded
			if app.selectedCollection != nil {
				app.showCollectionDetailView(*app.selectedCollection)
			}
			e.SetHandled()
		})
	} else {
		// Spacer for alignment
		spacer := core.NewFrame(row)
		spacer.Styler(func(s *styles.Style) {
			s.Min.X = units.Dp(24)
		})
	}

	// Container type icon
	iconBtn := core.NewButton(row).SetIcon(GetContainerTypeIcon(node.Container.Type))
	iconBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(appstyles.ColorPrimaryLightest)
		s.Color = colors.Uniform(appstyles.ColorPrimary)
		s.Padding.Set(units.Dp(8))
		s.Min.Set(units.Dp(36))
		s.Border.Radius = sides.NewValues(units.Dp(appstyles.RadiusDefault))
	})

	// Container info
	infoFrame := core.NewFrame(row)
	infoFrame.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Grow.Set(1, 0)
	})

	// Container name
	nameText := core.NewText(infoFrame).SetText(node.Container.Name)
	nameText.Styler(func(s *styles.Style) {
		s.Font.Weight = appstyles.WeightSemiBold
		s.Font.Size = units.Dp(14)
	})

	// Container details (type + object count)
	detailsText := core.NewText(infoFrame).SetText(fmt.Sprintf("%s • %d objects", node.Container.Type, len(node.Container.Objects)))
	detailsText.Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(appstyles.ColorBlack)
		s.Font.Size = units.Dp(12)
	})

	// Capacity indicator
	if node.Container.Capacity != nil && node.Container.UsedCapacity != nil {
		app.RenderCapacityIndicator(row, node.Container)
	}

	// View Objects button
	viewObjectsBtn := core.NewButton(row).SetIcon(icons.Visibility)
	viewObjectsBtn.Styler(func(s *styles.Style) {
		s.Background = nil
		s.Padding.Set(units.Dp(4))
		s.Color = colors.Uniform(appstyles.ColorPrimary) // Teal color for visibility
	})
	viewObjectsBtn.SetTooltip("View objects in this container")
	viewObjectsBtn.OnClick(func(e events.Event) {
		app.showContainerDetailView(*node.Container, *app.selectedCollection)
		e.SetHandled()
	})

	// Actions button
	actionsBtn := core.NewButton(row).SetIcon(icons.MoreVert)
	actionsBtn.Styler(func(s *styles.Style) {
		s.Background = nil
		s.Padding.Set(units.Dp(4))
	})
	actionsBtn.SetTooltip("Container actions")
	actionsBtn.OnClick(func(e events.Event) {
		app.showContainerActions(node.Container)
		e.SetHandled()
	})

	// Click handler for row
	row.OnClick(func(e events.Event) {
		app.showContainerDetail(node.Container)
	})

	// Render children if expanded
	if node.Expanded && len(node.Children) > 0 {
		app.RenderContainerTree(parent, node.Children, level+1)
	}
}

// RenderCapacityIndicator renders a visual capacity indicator
func (app *App) RenderCapacityIndicator(parent core.Widget, container *types.Container) {
	if container.Capacity == nil || container.UsedCapacity == nil {
		return
	}

	capacityFrame := core.NewFrame(parent)
	capacityFrame.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Align.Items = styles.End
		s.Min.X = units.Dp(100)
	})

	// Utilization percentage
	utilization := 0.0
	if container.CapacityUtilization != nil {
		utilization = *container.CapacityUtilization
	}

	// Color based on utilization
	var barColor, textColor color.RGBA
	if utilization > 100 {
		barColor = appstyles.ColorDanger
		textColor = appstyles.ColorDanger
	} else if utilization >= 80 {
		barColor = appstyles.ColorAccent
		textColor = appstyles.ColorAccent
	} else {
		barColor = appstyles.ColorPrimary
		textColor = appstyles.ColorTextSecondary
	}

	// Percentage text
	percentText := core.NewText(capacityFrame).SetText(fmt.Sprintf("%.0f%%", utilization))
	percentText.Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(textColor)
		s.Font.Size = units.Dp(11)
		s.Font.Weight = appstyles.WeightSemiBold
	})

	// Progress bar background
	progressBg := core.NewFrame(capacityFrame)
	progressBg.Styler(func(s *styles.Style) {
		s.Min.X = units.Dp(100)
		s.Min.Y = units.Dp(6)
		s.Background = colors.Uniform(appstyles.ColorGrayLightest)
		s.Border.Radius = sides.NewValues(units.Dp(3))
	})

	// Progress bar fill
	progressFill := core.NewFrame(progressBg)
	progressFill.Styler(func(s *styles.Style) {
		// Cap at 100% width
		fillPercent := utilization
		if fillPercent > 100 {
			fillPercent = 100
		}
		// Use Grow to fill proportionally
		s.Grow.Set(float32(fillPercent), 0)
		s.Min.Y = units.Dp(6)
		s.Background = colors.Uniform(barColor)
		s.Border.Radius = sides.NewValues(units.Dp(3))
	})
}

// showContainerDetail shows detailed view of a container
func (app *App) showContainerDetail(container *types.Container) {
	app.showDialog(DialogConfig{
		Title: container.Name,
		ContentBuilder: func(dialog core.Widget, closeDialog func()) {
			// Container type
			typeRow := core.NewFrame(dialog)
			typeRow.Styler(func(s *styles.Style) {
				s.Direction = styles.Row
				s.Align.Items = styles.Center
				s.Gap.Set(units.Dp(8))
				s.Margin.Bottom = units.Dp(12)
			})

			typeIcon := core.NewButton(typeRow).SetIcon(GetContainerTypeIcon(container.Type))
			typeIcon.Styler(func(s *styles.Style) {
				s.Background = colors.Uniform(appstyles.ColorPrimaryLightest)
				s.Color = colors.Uniform(appstyles.ColorPrimary)
				s.Padding.Set(units.Dp(8))
			})

			core.NewText(typeRow).SetText(container.Type).Styler(func(s *styles.Style) {
				s.Font.Weight = appstyles.WeightSemiBold
			})

			// Stats
			statsFrame := core.NewFrame(dialog)
			statsFrame.Styler(func(s *styles.Style) {
				s.Direction = styles.Row
				s.Gap.Set(units.Dp(16))
				s.Margin.Bottom = units.Dp(16)
			})

			// Object count
			app.renderStat(statsFrame, "Objects", fmt.Sprintf("%d", len(container.Objects)))

			// Capacity info
			if container.Capacity != nil {
				app.renderStat(statsFrame, "Capacity", fmt.Sprintf("%.1f / %.1f", *container.UsedCapacity, *container.Capacity))
			}

			// Dimensions
			if container.Width != nil && container.Depth != nil {
				app.renderStat(statsFrame, "Dimensions", fmt.Sprintf("%.1f\" x %.1f\"", *container.Width, *container.Depth))
			}

			// Location
			if container.Location != "" {
				core.NewText(dialog).SetText(fmt.Sprintf("Location: %s", container.Location)).Styler(func(s *styles.Style) {
					s.Color = colors.Uniform(appstyles.ColorTextSecondary)
					s.Margin.Bottom = units.Dp(12)
				})
			}

			// Full capacity indicator
			if container.Capacity != nil {
				app.RenderFullCapacityBar(dialog, container)
			}

			// Actions
			actionsRow := core.NewFrame(dialog)
			actionsRow.Styler(func(s *styles.Style) {
				s.Direction = styles.Row
				s.Gap.Set(units.Dp(8))
				s.Margin.Top = units.Dp(16)
			})

			importBtn := core.NewButton(actionsRow).SetText("Import to Container").SetIcon(icons.Upload)
			importBtn.Styler(appstyles.StyleButtonPrimary)
			importBtn.OnClick(func(e events.Event) {
				app.ShowImportDialog(container.ID, container.CollectionID)
			})

			viewObjectsBtn := core.NewButton(actionsRow).SetText("View Objects").SetIcon(icons.Visibility)
			viewObjectsBtn.Styler(appstyles.StyleButtonCancel)
			viewObjectsBtn.OnClick(func(e events.Event) {
				// Navigate to container detail view showing all objects
				app.showContainerDetailView(*container, *app.selectedCollection)
			})
		},
	})
}

// renderStat renders a stat item
func (app *App) renderStat(parent core.Widget, label, value string) {
	statFrame := core.NewFrame(parent)
	statFrame.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Background = colors.Uniform(appstyles.ColorGrayLightest)
		s.Padding.Set(units.Dp(8), units.Dp(12))
		s.Border.Radius = sides.NewValues(units.Dp(appstyles.RadiusDefault))
	})

	core.NewText(statFrame).SetText(label).Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(appstyles.ColorTextSecondary)
		s.Font.Size = units.Dp(11)
	})

	core.NewText(statFrame).SetText(value).Styler(func(s *styles.Style) {
		s.Font.Weight = appstyles.WeightSemiBold
		s.Font.Size = units.Dp(16)
	})
}

// RenderFullCapacityBar renders a detailed capacity bar with labels
func (app *App) RenderFullCapacityBar(parent core.Widget, container *types.Container) {
	capacityFrame := core.NewFrame(parent)
	capacityFrame.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(units.Dp(8))
		s.Margin.Top = units.Dp(12)
	})

	// Header
	headerRow := core.NewFrame(capacityFrame)
	headerRow.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Justify.Content = styles.SpaceBetween
	})

	core.NewText(headerRow).SetText("Capacity Utilization").Styler(func(s *styles.Style) {
		s.Font.Weight = appstyles.WeightSemiBold
		s.Font.Size = units.Dp(14)
	})

	utilization := 0.0
	if container.CapacityUtilization != nil {
		utilization = *container.CapacityUtilization
	}

	statusText := "Good"
	statusColor := appstyles.ColorPrimary
	if utilization > 100 {
		statusText = "Over Capacity!"
		statusColor = appstyles.ColorDanger
	} else if utilization >= 80 {
		statusText = "Near Capacity"
		statusColor = appstyles.ColorAccent
	}

	core.NewText(headerRow).SetText(statusText).Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(statusColor)
		s.Font.Weight = appstyles.WeightSemiBold
		s.Font.Size = units.Dp(12)
	})

	// Progress bar
	progressBg := core.NewFrame(capacityFrame)
	progressBg.Styler(func(s *styles.Style) {
		s.Grow.Set(1, 0) // Fill available width
		s.Min.Y = units.Dp(24)
		s.Background = colors.Uniform(appstyles.ColorGrayLightest)
		s.Border.Radius = sides.NewValues(units.Dp(appstyles.RadiusDefault))
	})

	progressFill := core.NewFrame(progressBg)
	progressFill.Styler(func(s *styles.Style) {
		fillPercent := utilization
		if fillPercent > 100 {
			fillPercent = 100
		}
		s.Grow.Set(float32(fillPercent), 0)
		s.Min.Y = units.Dp(24)
		s.Background = colors.Uniform(statusColor)
		s.Border.Radius = sides.NewValues(units.Dp(appstyles.RadiusDefault))
	})

	// Percentage text inside bar
	percentText := core.NewText(progressFill).SetText(fmt.Sprintf("%.1f%%", utilization))
	percentText.Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(appstyles.ColorWhite)
		s.Font.Weight = appstyles.WeightBold
		s.Font.Size = units.Dp(12)
		s.Align.Self = styles.Center
		s.Padding.Left = units.Dp(8)
	})

	// Details
	detailsRow := core.NewFrame(capacityFrame)
	detailsRow.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Justify.Content = styles.SpaceBetween
	})

	core.NewText(detailsRow).SetText(fmt.Sprintf("Used: %.1f units", *container.UsedCapacity)).Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(appstyles.ColorTextSecondary)
		s.Font.Size = units.Dp(12)
	})

	core.NewText(detailsRow).SetText(fmt.Sprintf("Total: %.1f units", *container.Capacity)).Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(appstyles.ColorTextSecondary)
		s.Font.Size = units.Dp(12)
	})
}

// showContainerActions shows a comprehensive menu of actions for a container
func (app *App) showContainerActions(container *types.Container) {
	if app.selectedCollection == nil {
		app.logger.Error("No selected collection for container actions")
		return
	}

	app.showDialog(DialogConfig{
		Title: fmt.Sprintf("Actions: %s", container.Name),
		ContentBuilder: func(dialog core.Widget, closeDialog func()) {
			// Container info summary
			infoCard := core.NewFrame(dialog)
			infoCard.Styler(func(s *styles.Style) {
				s.Direction = styles.Column
				s.Background = colors.Uniform(appstyles.ColorGrayLightest)
				s.Padding.Set(units.Dp(12))
				s.Border.Radius = sides.NewValues(units.Dp(appstyles.RadiusDefault))
				s.Gap.Set(units.Dp(4))
				s.Margin.Bottom = units.Dp(16)
			})

			typeRow := core.NewFrame(infoCard)
			typeRow.Styler(func(s *styles.Style) {
				s.Direction = styles.Row
				s.Align.Items = styles.Center
				s.Gap.Set(units.Dp(8))
			})

			core.NewIcon(typeRow).SetIcon(GetContainerTypeIcon(container.Type)).Styler(func(s *styles.Style) {
				s.Color = colors.Uniform(appstyles.ColorPrimary)
			})

			core.NewText(typeRow).SetText(fmt.Sprintf("%s • %d objects", container.Type, len(container.Objects))).Styler(func(s *styles.Style) {
				s.Color = colors.Uniform(appstyles.ColorTextSecondary)
				s.Font.Size = units.Dp(12)
			})

			// Action buttons
			actionsFrame := core.NewFrame(dialog)
			actionsFrame.Styler(func(s *styles.Style) {
				s.Direction = styles.Column
				s.Gap.Set(units.Dp(8))
			})

			// Add Sub-Container action
			addSubContainerBtn := core.NewButton(actionsFrame).SetText("Add Sub-Container").SetIcon(icons.CreateNewFolder)
			addSubContainerBtn.Styler(appstyles.StyleButtonPrimary)
			addSubContainerBtn.OnClick(func(e events.Event) {
				app.showCreateContainerDialogWithParent(*app.selectedCollection, container)
			})

			// Add Objects action
			addObjectsBtn := core.NewButton(actionsFrame).SetText("Add Objects").SetIcon(icons.Add)
			addObjectsBtn.Styler(appstyles.StyleButtonAccent)
			addObjectsBtn.OnClick(func(e events.Event) {
				// Close parent dialog before opening object creation dialog
				closeDialog()
				// Open create object dialog for this container
				app.showCreateObjectDialog(*container, *app.selectedCollection)
			})

			// Import to Container action
			importBtn := core.NewButton(actionsFrame).SetText("Import to Container").SetIcon(icons.Upload)
			importBtn.Styler(appstyles.StyleButtonPrimary)
			importBtn.OnClick(func(e events.Event) {
				app.ShowImportDialog(container.ID, container.CollectionID)
			})

			// Edit Container action
			editBtn := core.NewButton(actionsFrame).SetText("Edit Container").SetIcon(icons.Edit)
			editBtn.Styler(func(s *styles.Style) {
				s.Background = colors.Uniform(appstyles.ColorGrayLightest)
				s.Color = colors.Uniform(appstyles.ColorBlack)
				s.Padding.Set(units.Dp(12), units.Dp(16))
				s.Border.Radius = sides.NewValues(units.Dp(appstyles.RadiusDefault))
				s.Gap.Set(units.Dp(8))
			})
			editBtn.OnClick(func(e events.Event) {
				app.showEditContainerDialog(*container, *app.selectedCollection)
			})

			// View Details action
			viewDetailsBtn := core.NewButton(actionsFrame).SetText("View Details").SetIcon(icons.Visibility)
			viewDetailsBtn.Styler(func(s *styles.Style) {
				s.Background = colors.Uniform(appstyles.ColorGrayLightest)
				s.Color = colors.Uniform(appstyles.ColorBlack)
				s.Padding.Set(units.Dp(12), units.Dp(16))
				s.Border.Radius = sides.NewValues(units.Dp(appstyles.RadiusDefault))
				s.Gap.Set(units.Dp(8))
			})
			viewDetailsBtn.OnClick(func(e events.Event) {
				app.showContainerDetail(container)
			})

			// Delete Container action
			deleteBtn := core.NewButton(actionsFrame).SetText("Delete Container").SetIcon(icons.Delete)
			deleteBtn.Styler(appstyles.StyleButtonDanger)
			deleteBtn.OnClick(func(e events.Event) {
				app.showDeleteContainerDialog(*container, *app.selectedCollection)
			})
		},
	})
}

// showObjectsViewWithContainer navigates to objects view with a specific container pre-selected
func (app *App) showObjectsViewWithContainer(container *types.Container) {
	app.logger.Info("Navigating to objects view", "container_id", container.ID, "container_name", container.Name)

	// TODO: Implement navigation to objects view
	// For now, show a placeholder message
	app.SafeShowSnackbar(fmt.Sprintf("Navigate to objects for container: %s", container.Name))

	// In the future, this should:
	// 1. Set the selected container in app state
	// 2. Navigate to objects view
	// 3. Pre-filter objects by this container
	// app.selectedContainer = container
	// app.showObjectsView()
}
