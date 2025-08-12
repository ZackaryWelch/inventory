package main

import (
	"image/color"
	"strings"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
)

// Search and Filter functionality

// SearchFilter holds search and filter state
type SearchFilter struct {
	SearchQuery    string
	SelectedTags   []string
	SelectedTypes  []string
	SortBy         string
	SortDirection  string
	DateRange      DateRange
}

type DateRange struct {
	From string
	To   string
}

// Add search filter to App struct
func (app *App) initSearchFilter() {
	app.searchFilter = &SearchFilter{
		SortBy:        "name",
		SortDirection: "asc",
	}
}

// Global search view
func (app *App) showGlobalSearchView() {
	app.mainContainer.DeleteChildren()
	app.currentView = "search"

	// Header with back button
	header := app.createHeader("Search", true)

	// Main content
	content := core.NewFrame(app.mainContainer)
	content.Style(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Grow.Set(1, 1)
		s.Padding.Set(core.Dp(16))
		s.Gap.Set(core.Dp(16))
	})

	// Search bar
	searchContainer := core.NewFrame(content)
	searchContainer.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(core.Dp(12))
		s.Align.Items = styles.Center
		s.Background = ColorWhite
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(core.Dp(16))
	})

	searchIcon := core.NewIcon(searchContainer).SetIcon(icons.Search)
	searchIcon.Style(func(s *styles.Style) {
		s.Color = ColorGrayDark
	})

	searchField := core.NewTextField(searchContainer)
	searchField.SetPlaceholder("Search across all collections, containers, and objects...")
	searchField.Style(func(s *styles.Style) {
		s.Grow.Set(1, 1)
		s.Border.Style.Set(styles.BorderNone)
	})

	// Search button
	searchBtn := core.NewButton(searchContainer).SetText("Search").SetIcon(icons.Search)
	searchBtn.Style(func(s *styles.Style) {
		s.Background = ColorPrimary
		s.Color = ColorWhite
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(core.Dp(8), core.Dp(16))
		s.Gap.Set(core.Dp(4))
	})

	// Filters section
	filtersContainer := core.NewFrame(content)
	filtersContainer.Style(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Background = ColorWhite
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(core.Dp(16))
		s.Gap.Set(core.Dp(16))
	})

	filtersTitle := core.NewText(filtersContainer).SetText("Filters")
	filtersTitle.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(18)
		s.Font.Weight = styles.WeightSemiBold
	})

	// Filter row
	filterRow := core.NewFrame(filtersContainer)
	filterRow.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Wrap = true
		s.Gap.Set(core.Dp(12))
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
	resultsContainer.Style(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(core.Dp(16))
	})

	// Mock search results
	app.createSearchResults(resultsContainer)

	_ = header
	app.mainContainer.Update()
}

// Create type filter dropdown
func (app *App) createTypeFilter(parent core.Widget) {
	typeFilterContainer := core.NewFrame(parent)
	typeFilterContainer.Style(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(core.Dp(4))
	})

	typeLabel := core.NewText(typeFilterContainer).SetText("Object Type")
	typeLabel.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(12)
		s.Font.Weight = styles.WeightSemiBold
		s.Color = ColorGrayDark
	})

	typeDropdown := core.NewButton(typeFilterContainer).SetText("All Types").SetIcon(icons.ArrowDropDown)
	typeDropdown.Style(func(s *styles.Style) {
		s.Background = ColorGrayLightest
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(core.Dp(8), core.Dp(12))
		s.Gap.Set(core.Dp(4))
		s.Min.X.Set(core.Dp(120))
	})
}

// Create sort options
func (app *App) createSortOptions(parent core.Widget) {
	sortContainer := core.NewFrame(parent)
	sortContainer.Style(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(core.Dp(4))
	})

	sortLabel := core.NewText(sortContainer).SetText("Sort By")
	sortLabel.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(12)
		s.Font.Weight = styles.WeightSemiBold
		s.Color = ColorGrayDark
	})

	sortRow := core.NewFrame(sortContainer)
	sortRow.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(core.Dp(4))
	})

	sortDropdown := core.NewButton(sortRow).SetText("Name").SetIcon(icons.ArrowDropDown)
	sortDropdown.Style(func(s *styles.Style) {
		s.Background = ColorGrayLightest
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(core.Dp(8), core.Dp(12))
		s.Gap.Set(core.Dp(4))
	})

	sortDirectionBtn := core.NewButton(sortRow).SetIcon(icons.ArrowUpward)
	sortDirectionBtn.Style(func(s *styles.Style) {
		s.Background = ColorGrayLightest
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(core.Dp(8))
	})
}

// Create date range filter
func (app *App) createDateRangeFilter(parent core.Widget) {
	dateContainer := core.NewFrame(parent)
	dateContainer.Style(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(core.Dp(4))
	})

	dateLabel := core.NewText(dateContainer).SetText("Date Range")
	dateLabel.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(12)
		s.Font.Weight = styles.WeightSemiBold
		s.Color = ColorGrayDark
	})

	dateRow := core.NewFrame(dateContainer)
	dateRow.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(core.Dp(4))
		s.Align.Items = styles.Center
	})

	fromField := core.NewTextField(dateRow)
	fromField.SetPlaceholder("From")
	fromField.Style(func(s *styles.Style) {
		s.Min.X.Set(core.Dp(80))
	})

	toText := core.NewText(dateRow).SetText("to")
	toText.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(12)
		s.Color = ColorGrayDark
	})

	toField := core.NewTextField(dateRow)
	toField.SetPlaceholder("To")
	toField.Style(func(s *styles.Style) {
		s.Min.X.Set(core.Dp(80))
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
	activeFiltersContainer.Style(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Background = ColorWhite
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(core.Dp(16))
		s.Gap.Set(core.Dp(12))
	})

	filtersTitle := core.NewText(activeFiltersContainer).SetText("Active Filters")
	filtersTitle.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(14)
		s.Font.Weight = styles.WeightSemiBold
	})

	filtersRow := core.NewFrame(activeFiltersContainer)
	filtersRow.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Wrap = true
		s.Gap.Set(core.Dp(8))
	})

	// Mock active filters
	filters := []string{"Food Items", "Expires Soon", "Tagged: Organic"}
	for _, filter := range filters {
		filterBadge := core.NewFrame(filtersRow)
		filterBadge.Style(func(s *styles.Style) {
			s.Direction = styles.Row
			s.Align.Items = styles.Center
			s.Gap.Set(core.Dp(4))
			s.Background = ColorPrimaryLightest
			s.Border.Radius = styles.BorderRadiusFull
			s.Padding.Set(core.Dp(6), core.Dp(12))
		})

		filterText := core.NewText(filterBadge).SetText(filter)
		filterText.Style(func(s *styles.Style) {
			s.Font.Size = core.Dp(12)
			s.Color = ColorPrimary
		})

		removeBtn := core.NewButton(filterBadge).SetIcon(icons.Close)
		removeBtn.Style(func(s *styles.Style) {
			s.Background = ColorPrimary
			s.Color = ColorWhite
			s.Border.Radius = styles.BorderRadiusFull
			s.Padding.Set(core.Dp(2))
			s.Font.Size = core.Dp(10)
		})
	}

	// Clear all filters button
	clearAllBtn := core.NewButton(activeFiltersContainer).SetText("Clear All Filters")
	clearAllBtn.Style(func(s *styles.Style) {
		s.Background = color.RGBA{R: 240, G: 240, B: 240, A: 255}
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(core.Dp(8), core.Dp(12))
		s.Align.Self = styles.Start
	})
}

// Create search results display
func (app *App) createSearchResults(parent core.Widget) {
	resultsTitle := core.NewText(parent).SetText("Search Results")
	resultsTitle.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(18)
		s.Font.Weight = styles.WeightSemiBold
	})

	// Results summary
	summaryText := core.NewText(parent).SetText("Found 15 results across 3 collections")
	summaryText.Style(func(s *styles.Style) {
		s.Color = ColorGrayDark
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
	sectionContainer.Style(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Background = ColorWhite
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(core.Dp(16))
		s.Gap.Set(core.Dp(12))
	})

	sectionTitleText := core.NewText(sectionContainer).SetText(sectionTitle)
	sectionTitleText.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(16)
		s.Font.Weight = styles.WeightSemiBold
	})

	for _, result := range results {
		app.createSearchResultCard(sectionContainer, result)
	}
}

// Search result card
func (app *App) createSearchResultCard(parent core.Widget, result SearchResult) *core.Frame {
	card := core.NewFrame(parent)
	card.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Align.Items = styles.Center
		s.Gap.Set(core.Dp(12))
		s.Background = ColorGrayLightest
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(core.Dp(12))
		s.Cursor = styles.CursorPointer
	})

	// Icon
	resultIcon := core.NewIcon(card).SetIcon(result.Icon)
	resultIcon.Style(func(s *styles.Style) {
		s.Color = result.Color
	})

	// Content
	contentContainer := core.NewFrame(card)
	contentContainer.Style(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(core.Dp(4))
		s.Grow.Set(1, 0)
	})

	titleText := core.NewText(contentContainer).SetText(result.Title)
	titleText.Style(func(s *styles.Style) {
		s.Font.Weight = styles.WeightSemiBold
	})

	if result.Description != "" {
		descText := core.NewText(contentContainer).SetText(result.Description)
		descText.Style(func(s *styles.Style) {
			s.Font.Size = core.Dp(12)
			s.Color = ColorGrayDark
		})
	}

	pathText := core.NewText(contentContainer).SetText(result.Path)
	pathText.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(10)
		s.Color = ColorPrimary
	})

	// Action button
	viewBtn := core.NewButton(card).SetText("View").SetIcon(icons.ArrowForward)
	viewBtn.Style(func(s *styles.Style) {
		s.Background = ColorPrimary
		s.Color = ColorWhite
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(core.Dp(6), core.Dp(12))
		s.Gap.Set(core.Dp(4))
	})

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
			Icon:        icons.Restaurant,
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
			Color:       ColorPrimary,
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
			Icon:        icons.Restaurant,
			Color:       ColorAccent,
		},
		{
			Title:       "Whole Milk",
			Description: "Fresh whole milk, expires 2024-02-10",
			Path:        "Groups > Home > Kitchen Pantry > Refrigerator > Whole Milk",
			Type:        "object",
			Icon:        icons.Restaurant,
			Color:       ColorAccent,
		},
	}
}

// Advanced filter dialog
func (app *App) showAdvancedFilterDialog() {
	overlay := app.createOverlay()

	dialog := core.NewFrame(overlay)
	dialog.Style(func(s *styles.Style) {
		s.Background = ColorWhite
		s.Border.Radius = styles.BorderRadiusLarge
		s.Padding.Set(core.Dp(24))
		s.Gap.Set(core.Dp(16))
		s.Direction = styles.Column
		s.Min.X.Set(core.Dp(500))
		s.Max.X.Set(core.Dp(600))
		s.Max.Y.Set(core.Dp(500))
	})

	title := core.NewText(dialog).SetText("Advanced Filters")
	title.Style(func(s *styles.Style) {
		s.Font.Size = core.Dp(20)
		s.Font.Weight = styles.WeightSemiBold
	})

	// Tags filter
	tagsSection := core.NewFrame(dialog)
	tagsSection.Style(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(core.Dp(8))
	})

	tagsTitle := core.NewText(tagsSection).SetText("Filter by Tags")
	tagsTitle.Style(func(s *styles.Style) {
		s.Font.Weight = styles.WeightSemiBold
	})

	// Mock available tags
	tagsGrid := core.NewFrame(tagsSection)
	tagsGrid.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Wrap = true
		s.Gap.Set(core.Dp(8))
	})

	availableTags := []string{"organic", "dairy", "fruit", "vegetable", "snack", "beverage"}
	for _, tag := range availableTags {
		tagBtn := core.NewButton(tagsGrid).SetText(tag)
		tagBtn.Style(func(s *styles.Style) {
			s.Background = color.RGBA{R: 240, G: 240, B: 240, A: 255}
			s.Border.Radius = styles.BorderRadiusFull
			s.Padding.Set(core.Dp(6), core.Dp(12))
		})
	}

	// Property filters
	propsSection := core.NewFrame(dialog)
	propsSection.Style(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Gap.Set(core.Dp(8))
	})

	propsTitle := core.NewText(propsSection).SetText("Property Filters")
	propsTitle.Style(func(s *styles.Style) {
		s.Font.Weight = styles.WeightSemiBold
	})

	// Expiry date filter for food items
	expiryContainer := core.NewFrame(propsSection)
	expiryContainer.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(core.Dp(8))
		s.Align.Items = styles.Center
	})

	expiryLabel := core.NewText(expiryContainer).SetText("Expires within:")
	expiryField := core.NewTextField(expiryContainer)
	expiryField.SetPlaceholder("7")
	expiryField.Style(func(s *styles.Style) {
		s.Min.X.Set(core.Dp(60))
	})
	dayLabel := core.NewText(expiryContainer).SetText("days")

	// Buttons
	buttonRow := core.NewFrame(dialog)
	buttonRow.Style(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Gap.Set(core.Dp(12))
		s.Justify.Content = styles.End
	})

	clearBtn := core.NewButton(buttonRow).SetText("Clear All")
	clearBtn.Style(func(s *styles.Style) {
		s.Background = color.RGBA{R: 240, G: 240, B: 240, A: 255}
	})

	cancelBtn := core.NewButton(buttonRow).SetText("Cancel")
	cancelBtn.OnClick(func(e events.Event) {
		app.hideOverlay()
	})

	applyBtn := core.NewButton(buttonRow).SetText("Apply Filters")
	applyBtn.Style(func(s *styles.Style) {
		s.Background = ColorPrimary
		s.Color = ColorWhite
	})
	applyBtn.OnClick(func(e events.Event) {
		app.hideOverlay()
		// Apply filters and refresh search results
	})

	app.showOverlay(overlay)
}