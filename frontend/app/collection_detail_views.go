package app

import (
	"fmt"
	"image"
	"image/color"
	"sort"
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/nishiki/frontend/ui/theme"
)

// ============================================================
// Shared grouped rendering
// ============================================================

// objectGroup represents a named group of object indices for grouped rendering.
type objectGroup struct {
	name    string
	indices []int
}

// renderGroupedItems renders a list of object groups with headers and separators,
// dispatching to the appropriate view type for each group's objects.
func (ga *GioApp) renderGroupedItems(gtx layout.Context, groups []objectGroup) layout.Dimensions {
	if len(groups) == 0 {
		return ga.renderEmptyObjects(gtx)
	}

	// For grid-like layouts (grid, gallery), each group is one list item that renders
	// all objects. For row-based layouts, each object is its own list item.
	useBatchLayout := ga.objectViewLayout == ObjectViewGrid || ga.objectViewLayout == ObjectViewGallery

	type listItem struct {
		isHeader    bool
		isSeparator bool
		header      string
		objIndex    int // object index (row modes) or group index (batch modes)
	}
	var items []listItem
	for gi, g := range groups {
		if gi > 0 {
			items = append(items, listItem{isSeparator: true})
		}
		items = append(items, listItem{isHeader: true, header: fmt.Sprintf("%s (%d)", g.name, len(g.indices))})
		if useBatchLayout {
			items = append(items, listItem{objIndex: gi})
		} else {
			for _, idx := range g.indices {
				items = append(items, listItem{objIndex: idx})
			}
		}
	}

	list := &ga.widgetState.objectsList
	list.Axis = layout.Vertical
	return list.Layout(gtx, len(items), func(gtx layout.Context, index int) layout.Dimensions {
		item := items[index]
		if item.isSeparator {
			return ga.renderGroupSeparator(gtx)
		}
		if item.isHeader {
			return layout.Inset{Top: unit.Dp(theme.Spacing3), Bottom: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				label := material.Body1(ga.theme.Theme, item.header)
				label.Font.Weight = font.Bold
				label.Color = theme.ColorTextSecondary
				return label.Layout(gtx)
			})
		}
		if useBatchLayout {
			g := groups[item.objIndex]
			switch ga.objectViewLayout {
			case ObjectViewGallery:
				return ga.renderGalleryGrid(gtx, g.indices)
			default: // ObjectViewGrid
				return ga.renderObjectCardGrid(gtx, g.indices)
			}
		}
		// Row-based rendering
		obj := ga.objects[item.objIndex]
		idx := item.objIndex
		switch ga.objectViewLayout {
		case ObjectViewCompact:
			return ga.renderCompactRow(gtx, obj, idx)
		case ObjectViewTable:
			columns := ga.getTableColumns()
			return ga.renderTableRow(gtx, obj, idx, columns)
		default: // ObjectViewList, ObjectViewTree, or empty
			return layout.Inset{Bottom: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return ga.renderObjectCard(gtx, obj, idx)
			})
		}
	})
}

// ============================================================
// Compact List View
// ============================================================

// renderObjectsCompact renders objects as single-line rows.
func (ga *GioApp) renderObjectsCompact(gtx layout.Context) layout.Dimensions {
	if len(ga.objects) == 0 {
		return ga.renderEmptyObjects(gtx)
	}

	filteredObjects, filteredIndices := ga.getFilteredObjects()

	list := &ga.widgetState.objectsList
	list.Axis = layout.Vertical
	return list.Layout(gtx, len(filteredObjects), func(gtx layout.Context, index int) layout.Dimensions {
		obj := filteredObjects[index]
		originalIndex := filteredIndices[index]
		return ga.renderCompactRow(gtx, obj, originalIndex)
	})
}

// renderCompactRow renders a single compact row: name (bold) | location | quantity.
func (ga *GioApp) renderCompactRow(gtx layout.Context, obj Object, index int) layout.Dimensions {
	itemState := &ga.widgetState.objectItems[index]

	if itemState.editButton.Clicked(gtx) {
		ga.openObjectEditDialog(obj)
	}

	rowHeight := gtx.Dp(unit.Dp(36))

	return layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top: unit.Dp(theme.Spacing1), Bottom: unit.Dp(theme.Spacing1),
				Left: unit.Dp(theme.Spacing2), Right: unit.Dp(theme.Spacing2),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
				}.Layout(gtx,
					// Name (bold, takes most space)
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						label := material.Body2(ga.theme.Theme, obj.Name)
						label.Font.Weight = font.Bold
						label.MaxLines = 1
						return label.Layout(gtx)
					}),

					// Location (secondary)
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						loc := ga.getObjectEffectiveLocation(obj)
						if loc == "" {
							return layout.Dimensions{}
						}
						return layout.Inset{Left: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							label := material.Body2(ga.theme.Theme, loc)
							label.Color = theme.ColorTextSecondary
							label.MaxLines = 1
							return label.Layout(gtx)
						})
					}),

					// Quantity (right-aligned)
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if obj.Quantity == nil {
							return layout.Dimensions{}
						}
						qtyText := fmt.Sprintf("%v", *obj.Quantity)
						if obj.Unit != "" {
							qtyText += " " + obj.Unit
						}
						return layout.Inset{Left: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							label := material.Body2(ga.theme.Theme, qtyText)
							label.Color = theme.ColorTextSecondary
							label.MaxLines = 1
							return label.Layout(gtx)
						})
					}),
				)
			})
		}),
		// Click target over the whole row
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			return itemState.editButton.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{Size: image.Point{X: gtx.Constraints.Max.X, Y: rowHeight}}
			})
		}),
	)
}

// ============================================================
// Table View
// ============================================================

// renderObjectsTable renders objects in a table with column headers.
func (ga *GioApp) renderObjectsTable(gtx layout.Context) layout.Dimensions {
	if len(ga.objects) == 0 {
		return ga.renderEmptyObjects(gtx)
	}

	filteredObjects, filteredIndices := ga.getFilteredObjects()
	columns := ga.getTableColumns()

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Header row
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ga.renderTableHeader(gtx, columns)
		}),
		// Separator
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ga.renderGroupSeparator(gtx)
		}),
		// Data rows
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			list := &ga.widgetState.objectsList
			list.Axis = layout.Vertical
			return list.Layout(gtx, len(filteredObjects), func(gtx layout.Context, index int) layout.Dimensions {
				obj := filteredObjects[index]
				originalIndex := filteredIndices[index]
				return ga.renderTableRow(gtx, obj, originalIndex, columns)
			})
		}),
	)
}

// tableColumn describes a column in the table view.
type tableColumn struct {
	key         string
	displayName string
	flex        float32 // relative width weight
}

// getTableColumns returns the column definitions for the table view.
func (ga *GioApp) getTableColumns() []tableColumn {
	cols := []tableColumn{
		{key: "name", displayName: "Name", flex: 2},
		{key: "location", displayName: "Location", flex: 1.5},
		{key: "quantity", displayName: "Qty", flex: 0.7},
	}
	if ga.selectedCollection != nil && ga.selectedCollection.PropertySchema != nil {
		for _, def := range ga.selectedCollection.PropertySchema.Definitions {
			name := def.DisplayName
			if name == "" {
				name = snakeToTitleCase(def.Key)
			}
			cols = append(cols, tableColumn{key: def.Key, displayName: name, flex: 1})
		}
	}
	return cols
}

// renderTableHeader renders the table column headers.
func (ga *GioApp) renderTableHeader(gtx layout.Context, columns []tableColumn) layout.Dimensions {
	return layout.Inset{
		Top: unit.Dp(theme.Spacing1), Bottom: unit.Dp(theme.Spacing1),
		Left: unit.Dp(theme.Spacing2), Right: unit.Dp(theme.Spacing2),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		children := make([]layout.FlexChild, len(columns))
		for i, col := range columns {
			col := col
			isSortField := ga.objectSortField == col.key
			label := col.displayName
			if isSortField {
				if ga.objectSortDir == "desc" {
					label += " ↓"
				} else {
					label += " ↑"
				}
			}
			children[i] = layout.Flexed(col.flex, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body2(ga.theme.Theme, label)
				lbl.Font.Weight = font.Bold
				lbl.Color = theme.ColorTextSecondary
				lbl.MaxLines = 1
				return lbl.Layout(gtx)
			})
		}
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, children...)
	})
}

// renderTableRow renders a single table data row.
func (ga *GioApp) renderTableRow(gtx layout.Context, obj Object, index int, columns []tableColumn) layout.Dimensions {
	itemState := &ga.widgetState.objectItems[index]

	if itemState.editButton.Clicked(gtx) {
		ga.openObjectEditDialog(obj)
	}

	defMap := ga.getPropertyDefMap()
	rowHeight := gtx.Dp(unit.Dp(32))

	return layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top: unit.Dp(2), Bottom: unit.Dp(2),
				Left: unit.Dp(theme.Spacing2), Right: unit.Dp(theme.Spacing2),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				children := make([]layout.FlexChild, len(columns))
				for i, col := range columns {
					col := col
					cellText := ga.tableCellValue(obj, col.key, defMap)
					children[i] = layout.Flexed(col.flex, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Body2(ga.theme.Theme, cellText)
						lbl.MaxLines = 1
						if col.key == "name" {
							lbl.Font.Weight = font.Bold
						} else {
							lbl.Color = theme.ColorTextSecondary
						}
						return lbl.Layout(gtx)
					})
				}
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, children...)
			})
		}),
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			return itemState.editButton.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{Size: image.Point{X: gtx.Constraints.Max.X, Y: rowHeight}}
			})
		}),
	)
}

// tableCellValue returns the display string for a table cell.
func (ga *GioApp) tableCellValue(obj Object, key string, defMap map[string]*PropertyDefinition) string {
	switch key {
	case "name":
		return obj.Name
	case "location":
		return ga.getObjectEffectiveLocation(obj)
	case "quantity":
		if obj.Quantity == nil {
			return ""
		}
		q := fmt.Sprintf("%v", *obj.Quantity)
		if obj.Unit != "" {
			q += " " + obj.Unit
		}
		return q
	default:
		if tv, ok := obj.Properties[key]; ok {
			return RenderPropertyValueFromMap(key, tv, defMap)
		}
		return ""
	}
}

// ============================================================
// Gallery View
// ============================================================

// galleryColumns returns the number of columns for the gallery grid.
func galleryColumns(gtx layout.Context) int {
	widthDp := float32(gtx.Constraints.Max.X) / gtx.Metric.PxPerDp
	const minColWidth = 180
	cols := int(widthDp / minColWidth)
	if cols < 2 {
		cols = 2
	}
	return cols
}

// renderObjectsGallery renders objects as a tight thumbnail grid.
func (ga *GioApp) renderObjectsGallery(gtx layout.Context) layout.Dimensions {
	if len(ga.objects) == 0 {
		return ga.renderEmptyObjects(gtx)
	}

	_, filteredIndices := ga.getFilteredObjects()
	return ga.renderGalleryGrid(gtx, filteredIndices)
}

// renderGalleryGrid renders a gallery grid for a set of object indices.
func (ga *GioApp) renderGalleryGrid(gtx layout.Context, indices []int) layout.Dimensions {
	cols := galleryColumns(gtx)
	gap := gtx.Dp(unit.Dp(theme.Spacing1))
	thumbSize := (gtx.Constraints.Max.X - gap*(cols-1)) / cols

	// Chunk into rows
	type row struct{ indices []int }
	var rows []row
	for i := 0; i < len(indices); i += cols {
		end := i + cols
		if end > len(indices) {
			end = len(indices)
		}
		rows = append(rows, row{indices: indices[i:end]})
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			var totalHeight int
			for _, r := range rows {
				cellH := thumbSize + gtx.Dp(unit.Dp(24)) // thumbnail + name overlay space
				for ci, objIdx := range r.indices {
					xOffset := ci * (thumbSize + gap)
					stack := op.Offset(image.Point{X: xOffset, Y: totalHeight}).Push(gtx.Ops)
					ga.renderObjectThumbnail(gtx, ga.objects[objIdx], objIdx, thumbSize)
					stack.Pop()
				}
				totalHeight += cellH + gap
			}
			return layout.Dimensions{Size: image.Point{X: gtx.Constraints.Max.X, Y: totalHeight}}
		}),
	)
}

// renderObjectThumbnail renders a single gallery thumbnail cell.
func (ga *GioApp) renderObjectThumbnail(gtx layout.Context, obj Object, index int, size int) layout.Dimensions {
	itemState := &ga.widgetState.objectItems[index]

	if itemState.editButton.Clicked(gtx) {
		ga.openObjectEditDialog(obj)
	}

	cellH := size + gtx.Dp(unit.Dp(24))

	// Background
	bgRect := image.Rectangle{Max: image.Point{X: size, Y: cellH}}
	defer clip.RRect{Rect: bgRect, SE: gtx.Dp(unit.Dp(theme.RadiusDefault)), SW: gtx.Dp(unit.Dp(theme.RadiusDefault)), NW: gtx.Dp(unit.Dp(theme.RadiusDefault)), NE: gtx.Dp(unit.Dp(theme.RadiusDefault))}.Push(gtx.Ops).Pop()
	paint.ColorOp{Color: theme.ColorSurfaceAlt}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)

	// Image or placeholder
	img := ga.getImage(obj.ImageURL)
	if img != nil {
		imgOp := paint.NewImageOp(img)
		wImg := widget.Image{Src: imgOp, Fit: widget.Contain, Position: layout.Center}
		cgtx := gtx
		cgtx.Constraints = layout.Exact(image.Point{X: size, Y: size})
		wImg.Layout(cgtx)
	} else {
		// Placeholder: first letter on colored background
		ga.renderThumbnailPlaceholder(gtx, obj.Name, size)
	}

	// Name overlay at bottom
	nameOffset := op.Offset(image.Point{X: 0, Y: size}).Push(gtx.Ops)
	nameGtx := gtx
	nameGtx.Constraints = layout.Exact(image.Point{X: size, Y: gtx.Dp(unit.Dp(24))})
	layout.Center.Layout(nameGtx, func(gtx layout.Context) layout.Dimensions {
		lbl := material.Caption(ga.theme.Theme, obj.Name)
		lbl.MaxLines = 1
		lbl.Alignment = text.Middle
		return lbl.Layout(gtx)
	})
	nameOffset.Pop()

	// Clickable overlay
	itemState.editButton.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Point{X: size, Y: cellH}}
	})

	return layout.Dimensions{Size: image.Point{X: size, Y: cellH}}
}

// renderThumbnailPlaceholder renders a colored rectangle with the first letter of the name.
func (ga *GioApp) renderThumbnailPlaceholder(gtx layout.Context, name string, size int) {
	// Pick a color based on name hash
	letter := "?"
	if len(name) > 0 {
		letter = strings.ToUpper(name[:1])
	}
	var hash uint32
	for _, c := range name {
		hash = hash*31 + uint32(c)
	}
	colors := []color.NRGBA{
		theme.ColorPrimary, theme.ColorAccent, theme.ColorPrimaryDark, theme.ColorPrimaryLight,
	}
	bg := colors[hash%uint32(len(colors))]
	bg.A = 80

	rect := image.Rectangle{Max: image.Point{X: size, Y: size}}
	defer clip.Rect(rect).Push(gtx.Ops).Pop()
	paint.ColorOp{Color: bg}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)

	// Center the letter
	cgtx := gtx
	cgtx.Constraints = layout.Exact(image.Point{X: size, Y: size})
	layout.Center.Layout(cgtx, func(gtx layout.Context) layout.Dimensions {
		lbl := material.H4(ga.theme.Theme, letter)
		lbl.Color = theme.ColorTextSecondary
		lbl.Alignment = text.Middle
		return lbl.Layout(gtx)
	})
}

// ============================================================
// Tree View
// ============================================================

// renderObjectsTree renders objects in an expandable container hierarchy.
func (ga *GioApp) renderObjectsTree(gtx layout.Context) layout.Dimensions {
	if len(ga.objects) == 0 {
		return ga.renderEmptyObjects(gtx)
	}

	_, filteredIndices := ga.getFilteredObjects()

	// Build container tree and object mapping
	type treeItem struct {
		isContainer bool
		isSeparator bool
		containerID string
		name        string
		depth       int
		objIndex    int // for object items
		childCount  int // for container items
	}

	// Map objects to containers
	containerObjs := make(map[string][]int)
	var unassigned []int
	for _, idx := range filteredIndices {
		obj := ga.objects[idx]
		if obj.ContainerID == "" {
			unassigned = append(unassigned, idx)
		} else {
			containerObjs[obj.ContainerID] = append(containerObjs[obj.ContainerID], idx)
		}
	}

	// Build parent→children map for containers
	childContainers := make(map[string][]Container) // parentID → children
	rootContainers := make([]Container, 0)
	for _, c := range ga.containers {
		if c.ParentContainerID != nil && *c.ParentContainerID != "" {
			childContainers[*c.ParentContainerID] = append(childContainers[*c.ParentContainerID], c)
		} else {
			rootContainers = append(rootContainers, c)
		}
	}

	// Flatten tree using DFS
	var items []treeItem
	var walkContainer func(c Container, depth int)
	walkContainer = func(c Container, depth int) {
		objCount := len(containerObjs[c.ID])
		items = append(items, treeItem{
			isContainer: true,
			containerID: c.ID,
			name:        c.Name,
			depth:       depth,
			childCount:  objCount,
		})

		if ga.treeExpandedNodes[c.ID] {
			// Child containers
			for _, child := range childContainers[c.ID] {
				walkContainer(child, depth+1)
			}
			// Objects in this container
			for _, idx := range containerObjs[c.ID] {
				items = append(items, treeItem{
					objIndex: idx,
					depth:    depth + 1,
					name:     ga.objects[idx].Name,
				})
			}
		}
	}

	for _, c := range rootContainers {
		walkContainer(c, 0)
	}

	// Unassigned objects
	if len(unassigned) > 0 {
		items = append(items, treeItem{isSeparator: true})
		items = append(items, treeItem{
			isContainer: true,
			containerID: "__unassigned__",
			name:        "Unassigned",
			depth:       0,
			childCount:  len(unassigned),
		})
		if ga.treeExpandedNodes["__unassigned__"] {
			for _, idx := range unassigned {
				items = append(items, treeItem{
					objIndex: idx,
					depth:    1,
					name:     ga.objects[idx].Name,
				})
			}
		}
	}

	list := &ga.widgetState.objectsList
	list.Axis = layout.Vertical
	return list.Layout(gtx, len(items), func(gtx layout.Context, index int) layout.Dimensions {
		item := items[index]
		if item.isSeparator {
			return ga.renderGroupSeparator(gtx)
		}
		return ga.renderTreeNode(gtx, item.isContainer, item.containerID, item.name, item.depth, item.objIndex, item.childCount)
	})
}

// getTreeNodeClickable returns or creates a clickable for a tree node.
func (ga *GioApp) getTreeNodeClickable(key string) *widget.Clickable {
	if btn, ok := ga.treeNodeClickables[key]; ok {
		return btn
	}
	btn := new(widget.Clickable)
	ga.treeNodeClickables[key] = btn
	return btn
}

// renderTreeNode renders a single tree node (container header or object leaf).
func (ga *GioApp) renderTreeNode(gtx layout.Context, isContainer bool, containerID, name string, depth, objIndex, childCount int) layout.Dimensions {
	indent := unit.Dp(float32(depth) * 20)

	if isContainer {
		btn := ga.getTreeNodeClickable("tree_" + containerID)
		if btn.Clicked(gtx) {
			ga.treeExpandedNodes[containerID] = !ga.treeExpandedNodes[containerID]
		}
		expanded := ga.treeExpandedNodes[containerID]
		arrow := "▶"
		if expanded {
			arrow = "▼"
		}
		label := fmt.Sprintf("%s %s (%d)", arrow, name, childCount)

		return layout.Inset{Left: indent, Top: unit.Dp(theme.Spacing1), Bottom: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body1(ga.theme.Theme, label)
				lbl.Font.Weight = font.Bold
				lbl.Color = theme.ColorTextPrimary
				return lbl.Layout(gtx)
			})
		})
	}

	// Object leaf
	itemState := &ga.widgetState.objectItems[objIndex]
	if itemState.editButton.Clicked(gtx) {
		ga.openObjectEditDialog(ga.objects[objIndex])
	}

	return layout.Inset{Left: indent, Top: unit.Dp(1), Bottom: unit.Dp(1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return itemState.editButton.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body2(ga.theme.Theme, name)
			lbl.MaxLines = 1
			return lbl.Layout(gtx)
		})
	})
}

// ============================================================
// Stats / Summary View
// ============================================================

// renderCollectionStats renders a stats dashboard above the object list.
func (ga *GioApp) renderCollectionStats(gtx layout.Context) layout.Dimensions {
	if len(ga.objects) == 0 {
		return layout.Dimensions{}
	}

	return layout.Inset{Bottom: unit.Dp(theme.Spacing3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Total count
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Bottom: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body1(ga.theme.Theme, fmt.Sprintf("Total Objects: %d", len(ga.objects)))
					lbl.Font.Weight = font.Bold
					return lbl.Layout(gtx)
				})
			}),

			// Objects per container
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ga.renderContainerDistribution(gtx)
			}),

			// Tag cloud
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ga.renderTagCloud(gtx)
			}),

			// Property distributions
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ga.renderPropertyDistributions(gtx)
			}),

			// Separator
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ga.renderGroupSeparator(gtx)
			}),
		)
	})
}

// renderContainerDistribution renders a bar chart of objects per container.
func (ga *GioApp) renderContainerDistribution(gtx layout.Context) layout.Dimensions {
	if len(ga.containers) == 0 {
		return layout.Dimensions{}
	}

	// Count objects per container
	type bar struct {
		name  string
		count int
	}
	containerCounts := make(map[string]int)
	unassigned := 0
	for _, obj := range ga.objects {
		if obj.ContainerID == "" {
			unassigned++
		} else {
			containerCounts[obj.ContainerID]++
		}
	}

	var bars []bar
	maxCount := 0
	for _, c := range ga.containers {
		cnt := containerCounts[c.ID]
		if cnt > 0 {
			bars = append(bars, bar{name: c.Name, count: cnt})
			if cnt > maxCount {
				maxCount = cnt
			}
		}
	}
	if unassigned > 0 {
		bars = append(bars, bar{name: "Unassigned", count: unassigned})
		if unassigned > maxCount {
			maxCount = unassigned
		}
	}

	if len(bars) == 0 {
		return layout.Dimensions{}
	}

	return layout.Inset{Bottom: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body2(ga.theme.Theme, "Objects per Container:")
				lbl.Font.Weight = font.Bold
				lbl.Color = theme.ColorTextSecondary
				return lbl.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				children := make([]layout.FlexChild, len(bars))
				maxWidth := gtx.Constraints.Max.X
				for i, b := range bars {
					b := b
					children[i] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Top: unit.Dp(2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
								// Label
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									lbl := material.Caption(ga.theme.Theme, fmt.Sprintf("%s (%d)", b.name, b.count))
									lbl.Color = theme.ColorTextSecondary
									cgtx := gtx
									cgtx.Constraints.Min.X = gtx.Dp(unit.Dp(120))
									cgtx.Constraints.Max.X = gtx.Dp(unit.Dp(120))
									return lbl.Layout(cgtx)
								}),
								// Bar
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									barMaxW := maxWidth - gtx.Dp(unit.Dp(130))
									barW := barMaxW * b.count / maxCount
									if barW < 4 {
										barW = 4
									}
									barH := gtx.Dp(unit.Dp(12))
									sz := image.Point{X: barW, Y: barH}
									defer clip.RRect{Rect: image.Rectangle{Max: sz}, SE: 3, SW: 3, NW: 3, NE: 3}.Push(gtx.Ops).Pop()
									paint.ColorOp{Color: theme.ColorPrimary}.Add(gtx.Ops)
									paint.PaintOp{}.Add(gtx.Ops)
									return layout.Dimensions{Size: sz}
								}),
							)
						})
					})
				}
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
			}),
		)
	})
}

// renderTagCloud renders frequently used tags.
func (ga *GioApp) renderTagCloud(gtx layout.Context) layout.Dimensions {
	// Collect tag frequencies
	tagFreq := make(map[string]int)
	for _, obj := range ga.objects {
		for _, tag := range obj.Tags {
			tagFreq[tag]++
		}
	}
	if len(tagFreq) == 0 {
		return layout.Dimensions{}
	}

	// Sort by frequency descending
	type tagEntry struct {
		tag   string
		count int
	}
	var tags []tagEntry
	for t, c := range tagFreq {
		tags = append(tags, tagEntry{tag: t, count: c})
	}
	sort.Slice(tags, func(i, j int) bool { return tags[i].count > tags[j].count })

	// Limit to top 20
	if len(tags) > 20 {
		tags = tags[:20]
	}

	chipGap := gtx.Dp(unit.Dp(theme.Spacing1))

	return layout.Inset{Bottom: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body2(ga.theme.Theme, "Tags:")
				lbl.Font.Weight = font.Bold
				lbl.Color = theme.ColorTextSecondary
				return lbl.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				var widgets []layout.Widget
				for _, t := range tags {
					t := t
					widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Caption(ga.theme.Theme, fmt.Sprintf("%s (%d)", t.tag, t.count))
						lbl.Color = theme.ColorTextPrimary
						// Wrap in a small pill
						return layout.Inset{
							Top: unit.Dp(2), Bottom: unit.Dp(2),
							Left: unit.Dp(theme.Spacing1), Right: unit.Dp(theme.Spacing1),
						}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return lbl.Layout(gtx)
						})
					})
				}
				return layoutFlowWrap(gtx, chipGap, chipGap, widgets...)
			}),
		)
	})
}

// renderPropertyDistributions renders top values for grouped_text properties.
func (ga *GioApp) renderPropertyDistributions(gtx layout.Context) layout.Dimensions {
	if ga.selectedCollection == nil || ga.selectedCollection.PropertySchema == nil {
		return layout.Dimensions{}
	}

	// Find grouped_text properties
	var propDefs []PropertyDefinition
	for _, def := range ga.selectedCollection.PropertySchema.Definitions {
		if def.Type == "grouped_text" {
			propDefs = append(propDefs, def)
		}
	}
	if len(propDefs) == 0 {
		return layout.Dimensions{}
	}

	children := make([]layout.FlexChild, len(propDefs))
	for i, def := range propDefs {
		def := def
		// Count values
		valFreq := make(map[string]int)
		for _, obj := range ga.objects {
			if tv, ok := obj.Properties[def.Key]; ok {
				v := fmt.Sprintf("%v", tv.Val)
				if v != "" {
					valFreq[v]++
				}
			}
		}
		if len(valFreq) == 0 {
			children[i] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{}
			})
			continue
		}

		type valEntry struct {
			val   string
			count int
		}
		var vals []valEntry
		for v, c := range valFreq {
			vals = append(vals, valEntry{val: v, count: c})
		}
		sort.Slice(vals, func(a, b int) bool { return vals[a].count > vals[b].count })
		if len(vals) > 5 {
			vals = vals[:5]
		}

		displayName := def.DisplayName
		if displayName == "" {
			displayName = snakeToTitleCase(def.Key)
		}

		children[i] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: unit.Dp(theme.Spacing1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				parts := make([]string, len(vals))
				for j, v := range vals {
					parts[j] = fmt.Sprintf("%s (%d)", v.val, v.count)
				}
				lbl := material.Caption(ga.theme.Theme, displayName+": "+strings.Join(parts, ", "))
				lbl.Color = theme.ColorTextSecondary
				return lbl.Layout(gtx)
			})
		})
	}

	return layout.Inset{Bottom: unit.Dp(theme.Spacing2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
	})
}

// ============================================================
// Shared helpers
// ============================================================

// renderEmptyObjects renders the "No objects yet" placeholder.
func (ga *GioApp) renderEmptyObjects(gtx layout.Context) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		label := material.Body1(ga.theme.Theme, "No objects yet")
		label.Color = theme.ColorTextSecondary
		label.Alignment = text.Middle
		return label.Layout(gtx)
	})
}

// openObjectEditDialog opens the object edit dialog populated with the given object's data.
func (ga *GioApp) openObjectEditDialog(obj Object) {
	ga.selectedObject = &obj
	ga.showObjectDialog = true
	ga.objectDialogMode = "edit"
	ga.widgetState.objectNameEditor.SetText(obj.Name)
	ga.widgetState.objectDescriptionEditor.SetText(obj.Description)
	if obj.Quantity != nil {
		ga.widgetState.objectQuantityEditor.SetText(fmt.Sprintf("%v", *obj.Quantity))
	} else {
		ga.widgetState.objectQuantityEditor.SetText("")
	}
	ga.widgetState.objectUnitEditor.SetText(obj.Unit)
	if obj.ContainerID != "" {
		cid := obj.ContainerID
		ga.selectedContainerID = &cid
	} else {
		ga.selectedContainerID = nil
	}
	// Populate schema property editors
	if ga.selectedCollection != nil && ga.selectedCollection.PropertySchema != nil {
		for _, def := range ga.selectedCollection.PropertySchema.Definitions {
			if def.Type == "bool" {
				b := ga.getObjectPropertyBool(def.Key)
				b.Value = false
				if tv, ok := obj.Properties[def.Key]; ok {
					switch v := tv.Val.(type) {
					case bool:
						b.Value = v
					case string:
						b.Value = strings.EqualFold(v, "true") || v == "1"
					}
				}
			} else {
				ed := ga.getObjectPropertyEditor(def.Key)
				if tv, ok := obj.Properties[def.Key]; ok {
					ed.SetText(fmt.Sprintf("%v", tv.Val))
				} else {
					ed.SetText("")
				}
			}
		}
	}
}
