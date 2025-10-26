//go:build js && wasm

package app

import (
	"image/color"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/cursors"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"

	"github.com/nishiki/frontend/ui/layouts"
	appstyles "github.com/nishiki/frontend/ui/styles"
)

// Search and Filter functionality

// SearchFilter holds search and filter state
type SearchFilter struct {
	SearchQuery   string
	SelectedTags  []string
	SelectedTypes []string
	SortBy        string
	SortDirection string
	DateRange     DateRange
}

type DateRange struct {
	From string
	To   string
}

// Add search filter to App struct

// Global search view
func (app *App) showGlobalSearchView() {
	app.mainContainer.DeleteChildren()
	app.currentView = "search"

	// Header with back button
	layouts.SimpleHeader(app.mainContainer, "Search", true, func() {
		app.showDashboardView()
	})

	// Main content
	content := core.NewFrame(app.mainContainer)
	content.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Grow.Set(1, 1)
		s.Padding.Set(units.Dp(16))
		s.Gap.Set(units.Dp(16))
	})

	// Search bar
	searchContainer := core.NewFrame(content)
	searchContainer.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(units.Dp(12))
		s.Align.Items = styles.Center
		s.Background = colors.Uniform(appstyles.ColorWhite)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(16))
	})

	searchIcon := core.NewIcon(searchContainer).SetIcon(icons.Search)
	searchIcon.Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(appstyles.ColorGrayDark)
	})

	searchField := core.NewTextField(searchContainer)
	searchField.SetPlaceholder("Search across all collections, containers, and objects...")
	searchField.Styler(func(s *styles.Style) {
		appstyles.StyleInputRounded(s) // Apply proper input styling with white background
		s.Grow.Set(1, 1)
		// Keep border from StyleInputRounded, don't override to BorderNone
	})

	// Search button
	searchBtn := core.NewButton(searchContainer).SetText("Search").SetIcon(icons.Search)
	searchBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(appstyles.ColorPrimary)
		s.Color = colors.Uniform(appstyles.ColorWhite)
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(units.Dp(8), units.Dp(16))
		s.Gap.Set(units.Dp(4))
	})

	// Filters section
	filtersContainer := core.NewFrame(content)
	filtersContainer.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Background = colors.Uniform(appstyles.ColorWhite)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(16))
		s.Gap.Set(units.Dp(16))
	})

	filtersTitle := core.NewText(filtersContainer).SetText("Filters")
	filtersTitle.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(18)
		s.Font.Weight = appstyles.WeightSemiBold
	})

	// Filter row
	filterRow := core.NewFrame(filtersContainer)
	filterRow.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Wrap = true
		s.Gap.Set(units.Dp(12))
	})

	// Object type filter
	app.createTypeFilter(filterRow)

	// Sort options
	app.createSortOptions(filterRow)

	// Date range filter
	app.createDateRangeFilter(filterRow)

	// Active filters display
	if app.hasActiveFilters() {
		app.createActiveFiltersDisplay(content)
	}

	// Search results
	resultsContainer := core.NewFrame(content)
	resultsContainer.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(units.Dp(16))
	})

	// Mock search results
	app.createSearchResults(resultsContainer)

	app.mainContainer.Update()
}

// Create type filter dropdown
func (app *App) createTypeFilter(parent core.Widget) {
	typeFilterContainer := core.NewFrame(parent)
	typeFilterContainer.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(units.Dp(4))
	})

	typeLabel := core.NewText(typeFilterContainer).SetText("Object Type")
	typeLabel.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(12)
		s.Font.Weight = appstyles.WeightSemiBold
		s.Color = colors.Uniform(appstyles.ColorBlack) // Form labels should be black for visibility
	})

	typeDropdown := core.NewButton(typeFilterContainer).SetText("All Types").SetIcon(icons.ArrowDropDown)
	typeDropdown.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(appstyles.ColorGrayLightest)
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(units.Dp(8), units.Dp(12))
		s.Gap.Set(units.Dp(4))
		s.Min.X.Set(120, units.UnitDp)
	})
}

// Create sort options
func (app *App) createSortOptions(parent core.Widget) {
	sortContainer := core.NewFrame(parent)
	sortContainer.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(units.Dp(4))
	})

	sortLabel := core.NewText(sortContainer).SetText("Sort By")
	sortLabel.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(12)
		s.Font.Weight = appstyles.WeightSemiBold
		s.Color = colors.Uniform(appstyles.ColorBlack) // Form labels should be black for visibility
	})

	sortRow := core.NewFrame(sortContainer)
	sortRow.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(units.Dp(4))
	})

	sortDropdown := core.NewButton(sortRow).SetText("Name").SetIcon(icons.ArrowDropDown)
	sortDropdown.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(appstyles.ColorGrayLightest)
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(units.Dp(8), units.Dp(12))
		s.Gap.Set(units.Dp(4))
	})

	sortDirectionBtn := core.NewButton(sortRow).SetIcon(icons.ArrowUpward)
	sortDirectionBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(appstyles.ColorGrayLightest)
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(units.Dp(8))
	})
}

// Create date range filter
func (app *App) createDateRangeFilter(parent core.Widget) {
	dateContainer := core.NewFrame(parent)
	dateContainer.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(units.Dp(4))
	})

	dateLabel := core.NewText(dateContainer).SetText("Date Range")
	dateLabel.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(12)
		s.Font.Weight = appstyles.WeightSemiBold
		s.Color = colors.Uniform(appstyles.ColorBlack) // Form labels should be black for visibility
	})

	dateRow := core.NewFrame(dateContainer)
	dateRow.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(units.Dp(4))
		s.Align.Items = styles.Center
	})

	fromField := core.NewTextField(dateRow)
	fromField.SetPlaceholder("From")
	fromField.Styler(func(s *styles.Style) {
		s.Min.X.Set(80, units.UnitDp)
	})

	toText := core.NewText(dateRow).SetText("to")
	toText.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(12)
		s.Color = colors.Uniform(appstyles.ColorGrayDark)
	})

	toField := core.NewTextField(dateRow)
	toField.SetPlaceholder("To")
	toField.Styler(func(s *styles.Style) {
		s.Min.X.Set(80, units.UnitDp)
	})
}

// Check if there are active filters
func (app *App) hasActiveFilters() bool {
	if app.searchFilter == nil {
		return false
	}
	return app.searchFilter.SearchQuery != "" ||
		len(app.searchFilter.SelectedTags) > 0 ||
		len(app.searchFilter.SelectedTypes) > 0 ||
		app.searchFilter.DateRange.From != "" ||
		app.searchFilter.DateRange.To != ""
}

// Create active filters display
func (app *App) createActiveFiltersDisplay(parent core.Widget) {
	activeFiltersContainer := core.NewFrame(parent)
	activeFiltersContainer.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Background = colors.Uniform(appstyles.ColorWhite)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(16))
		s.Gap.Set(units.Dp(12))
	})

	filtersTitle := core.NewText(activeFiltersContainer).SetText("Active Filters")
	filtersTitle.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(14)
		s.Font.Weight = appstyles.WeightSemiBold
	})

	filtersRow := core.NewFrame(activeFiltersContainer)
	filtersRow.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Wrap = true
		s.Gap.Set(units.Dp(8))
	})

	// Mock active filters
	filters := []string{"Food Items", "Expires Soon", "Tagged: Organic"}
	for _, filter := range filters {
		filterBadge := core.NewFrame(filtersRow)
		filterBadge.Styler(func(s *styles.Style) {
			s.Direction = styles.Row
			s.Align.Items = styles.Center
			s.Gap.Set(units.Dp(4))
			s.Background = colors.Uniform(appstyles.ColorPrimaryLightest)
			s.Border.Radius = styles.BorderRadiusFull
			s.Padding.Set(units.Dp(6), units.Dp(12))
		})

		filterText := core.NewText(filterBadge).SetText(filter)
		filterText.Styler(func(s *styles.Style) {
			s.Font.Size = units.Dp(12)
			s.Color = colors.Uniform(appstyles.ColorPrimary)
		})

		removeBtn := core.NewButton(filterBadge).SetIcon(icons.Close)
		removeBtn.Styler(func(s *styles.Style) {
			s.Background = colors.Uniform(appstyles.ColorPrimary)
			s.Color = colors.Uniform(appstyles.ColorWhite)
			s.Border.Radius = styles.BorderRadiusFull
			s.Padding.Set(units.Dp(2))
			s.Font.Size = units.Dp(10)
		})
	}

	// Clear all filters button
	clearAllBtn := core.NewButton(activeFiltersContainer).SetText("Clear All Filters")
	clearAllBtn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(color.RGBA{R: 240, G: 240, B: 240, A: 255})
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(units.Dp(8), units.Dp(12))
		s.Align.Self = styles.Start
	})
}

// Create search results display
func (app *App) createSearchResults(parent core.Widget) {
	resultsTitle := core.NewText(parent).SetText("Search Results")
	resultsTitle.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(18)
		s.Font.Weight = appstyles.WeightSemiBold
	})

	// Results summary
	summaryText := core.NewText(parent).SetText("Found 15 results across 3 collections")
	summaryText.Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(appstyles.ColorGrayDark)
	})

	// Group results by type
	app.createSearchResultsSection(parent, "Collections", app.getMockCollectionResults())
	app.createSearchResultsSection(parent, "Containers", app.getMockContainerResults())
	app.createSearchResultsSection(parent, "Objects", app.getMockObjectResults())
}

// Create a section of search results
func (app *App) createSearchResultsSection(parent core.Widget, sectionTitle string, results []SearchResult) {
	if len(results) == 0 {
		return
	}

	sectionContainer := core.NewFrame(parent)
	sectionContainer.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Background = colors.Uniform(appstyles.ColorWhite)
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(units.Dp(16))
		s.Gap.Set(units.Dp(12))
	})

	sectionTitleText := core.NewText(sectionContainer).SetText(sectionTitle)
	sectionTitleText.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(16)
		s.Font.Weight = appstyles.WeightSemiBold
	})

	for _, result := range results {
		app.createSearchResultCard(sectionContainer, result)
	}
}

// Search result card (horizontal layout variant)
func (app *App) createSearchResultCard(parent core.Widget, result SearchResult) *core.Frame {
	card := core.NewFrame(parent)
	card.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Gap.Set(units.Dp(12))
		s.Background = colors.Uniform(appstyles.ColorGrayLightest)
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(units.Dp(12))
		s.Cursor = cursors.Pointer
	})

	// Icon
	resultIcon := core.NewIcon(card).SetIcon(result.Icon)
	resultIcon.Styler(func(s *styles.Style) {
		s.Color = colors.Uniform(result.Color)
	})

	// Content
	contentContainer := createFlexColumn(card, 4)
	contentContainer.Styler(func(s *styles.Style) {
		s.Grow.Set(1, 0)
	})

	titleText := core.NewText(contentContainer).SetText(result.Title)
	titleText.Styler(func(s *styles.Style) {
		s.Font.Weight = appstyles.WeightSemiBold
	})

	if result.Description != "" {
		descText := core.NewText(contentContainer).SetText(result.Description)
		descText.Styler(func(s *styles.Style) {
			s.Font.Size = units.Dp(12)
			s.Color = colors.Uniform(appstyles.ColorGrayDark)
		})
	}

	pathText := core.NewText(contentContainer).SetText(result.Path)
	pathText.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(10)
		s.Color = colors.Uniform(appstyles.ColorPrimary)
	})

	// Action button
	viewBtn := core.NewButton(card).SetText("View").SetIcon(icons.ArrowForward)
	viewBtn.Styler(appstyles.StyleButtonPrimary)
	viewBtn.Styler(appstyles.StyleButtonSm)

	return card
}

// Search result data structure
type SearchResult struct {
	Title       string
	Description string
	Path        string
	Type        string
	Icon        icons.Icon
	Color       color.RGBA
}

// Mock search results
func (app *App) getMockCollectionResults() []SearchResult {
	return []SearchResult{
		{
			Title:       "Kitchen Pantry",
			Description: "Main food storage collection",
			Path:        "Groups > Home > Kitchen Pantry",
			Type:        "collection",
			Icon:        icons.Dining,
			Color:       color.RGBA{R: 76, G: 175, B: 80, A: 255},
		},
	}
}

func (app *App) getMockContainerResults() []SearchResult {
	return []SearchResult{
		{
			Title:       "Refrigerator",
			Description: "Cold storage container",
			Path:        "Groups > Home > Kitchen Pantry > Refrigerator",
			Type:        "container",
			Icon:        icons.FolderOpen,
			Color:       appstyles.ColorPrimary,
		},
	}
}

func (app *App) getMockObjectResults() []SearchResult {
	return []SearchResult{
		{
			Title:       "Organic Bananas",
			Description: "Fresh organic bananas from Ecuador",
			Path:        "Groups > Home > Kitchen Pantry > Refrigerator > Organic Bananas",
			Type:        "object",
			Icon:        icons.Dining,
			Color:       appstyles.ColorAccent,
		},
		{
			Title:       "Whole Milk",
			Description: "Fresh whole milk, expires 2024-02-10",
			Path:        "Groups > Home > Kitchen Pantry > Refrigerator > Whole Milk",
			Type:        "object",
			Icon:        icons.Dining,
			Color:       appstyles.ColorAccent,
		},
	}
}

// Advanced filter dialog
