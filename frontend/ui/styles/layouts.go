package styles

import (
	"cogentcore.org/core/colors"
	"cogentcore.org/core/cursors"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/sides"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/text/text"
)

// ====================================================================================
// LAYOUT STYLES - Matching nishiki-frontend layouts exactly
// ====================================================================================

// ====================================================================================
// Mobile Layout (from nishiki-frontend/src/components/layouts/MobileLayout.tsx)
// ====================================================================================

// MobileLayout container: className="pt-12 pb-16 min-h-screen"
// NOTE: Background color is on <body>, NOT here! Children control their own flex direction.
func StyleMobileLayoutContainer(s *styles.Style) {
	s.Padding.Set(units.Dp(Spacing12), units.Dp(0), units.Dp(Spacing16), units.Dp(0)) // pt-12 pb-16
	// min-h-screen handled by natural growth, not explicit constraint
}

// MobileLayout content: className="flex flex-col gap-2 px-4 pt-6 pb-16"
func StyleMobileLayoutContent(s *styles.Style) {
	s.Display = styles.Flex
	s.Direction = styles.Column                  // flex-col
	s.Gap.Set(units.Dp(Spacing2))                // gap-2 (8px)
	s.Padding.Set(
		units.Dp(Spacing6),  // pt-6 (24px)
		units.Dp(Spacing4),  // px-4 (16px)
		units.Dp(Spacing16), // pb-16 (64px)
		units.Dp(Spacing4),  // px-4 (16px)
	)
}

// ====================================================================================
// Main Container and Content Layouts
// ====================================================================================

// StyleMainContainer matches MobileLayout: className="pt-12 pb-16 min-h-screen"
// This is the app's main container, used for all authenticated views
// NOTE: Background color is on <body>, NOT here!
func StyleMainContainer(s *styles.Style) {
	s.Direction = styles.Column                                                        // flex-col for vertical stacking
	s.Padding.Set(units.Dp(Spacing12), units.Dp(0), units.Dp(Spacing16), units.Dp(0)) // pt-12 pb-16
	s.Grow.Set(1, 1)                                                                   // Grow to fill Body
	// Background is on Body, not here
}

// StyleContentColumn matches GroupsPage content: className="pt-6 px-4 pb-2 flex flex-col gap-2"
func StyleContentColumn(s *styles.Style) {
	s.Display = styles.Flex                                                                        // flex
	s.Direction = styles.Column                                                                    // flex-col
	s.Padding.Set(units.Dp(Spacing6), units.Dp(Spacing4), units.Dp(Spacing2), units.Dp(Spacing4)) // pt-6 px-4 pb-2
	s.Gap.Set(units.Dp(Spacing2))                                                                  // gap-2
	// No Align.Items - let children size naturally (stretch to full width by default)
}

func StyleCenteredContainer(s *styles.Style) {
	s.Display = styles.Flex           // CRITICAL: Must set Display to Flex for centering to work
	s.Direction = styles.Column
	s.Align.Items = styles.Center
	s.Justify.Content = styles.Center
	s.Grow.Set(1, 1)
	s.Gap.Set(units.Dp(Spacing8))
	s.Padding.Set(units.Dp(Spacing6))
}

// ====================================================================================
// Header Layouts (from nishiki-frontend/src/components/parts/Header.tsx)
// ====================================================================================

// Main header: className="fixed top-0 z-40 w-full h-12 bg-white flex items-center justify-center"
func StyleMainHeader(s *styles.Style) {
	// Note: fixed top-0 z-40 positioning would need positioning system
	s.Min.X.Set(100, units.UnitVw)            // w-full
	s.Min.Y.Set(48, units.UnitDp)             // h-12
	s.Background = colors.Uniform(ColorWhite) // bg-white
	s.Display = styles.Flex                   // flex
	s.Align.Items = styles.Center             // items-center
	s.Justify.Content = styles.Center         // justify-center
}

// HeaderBackButton: className="h-full aspect-square pl-4 flex items-center"
func StyleHeaderBackButton(s *styles.Style) {
	s.Min.Y.Set(100, units.UnitPh)                 // h-full (parent height)
	s.Min.X.Set(48, units.UnitDp)                  // aspect-square (48px for h-12)
	s.Padding.Left = units.Dp(Spacing4)            // pl-4
	s.Display = styles.Flex                        // flex
	s.Align.Items = styles.Center                  // items-center
}

// HeaderMenuCircleButton: className="h-full px-4 flex items-center"
func StyleHeaderMenuButton(s *styles.Style) {
	s.Min.Y.Set(100, units.UnitPh)                       // h-full (parent height)
	s.Padding.Set(units.Dp(0), units.Dp(Spacing4))       // px-4
	s.Display = styles.Flex                              // flex
	s.Align.Items = styles.Center                        // items-center
}

func StyleHeaderRow(s *styles.Style) {
	s.Direction = styles.Row
	s.Align.Items = styles.Center
	s.Justify.Content = styles.Center // Center the title
	// No background - header should be transparent to match React design
	s.Min.Y.Set(48, units.UnitDp)                        // h-12 (48px)
	// Let it size naturally to parent width
	s.Padding.Set(units.Dp(0), units.Dp(Spacing4))       // px-4 for proper edge spacing
}

func StyleHeaderTitle(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeXL)          // text-xl (matching React)
	s.Font.Weight = WeightSemiBold
	s.Color = colors.Uniform(ColorBlack)
	s.Text.WhiteSpace = text.WrapNever          // Never wrap header titles
	s.Text.Align = AlignCenter                   // Center text
	s.Grow.Set(1, 0)                             // Grow to fill space horizontally
}

func StyleHeaderLeftContainer(s *styles.Style) {
	s.Direction = styles.Row
	s.Align.Items = styles.Center
	s.Gap.Set(units.Dp(Spacing3))
}

// Navigation header with actions: className="h-16 px-4 border-b border-gray-light flex items-center justify-between"
func StyleNavHeader(s *styles.Style) {
	s.Min.Y.Set(64, units.UnitDp)                      // h-16
	s.Padding.Set(units.Dp(0), units.Dp(Spacing4))     // px-4
	s.Border.Style.Bottom = styles.BorderSolid         // border-b
	s.Border.Width.Bottom = units.Dp(1)                // border-b
	s.Border.Color.Set(colors.Uniform(ColorGrayLight)) // border-gray-light
	s.Display = styles.Flex                            // flex
	s.Align.Items = styles.Center                      // items-center
	s.Justify.Content = styles.SpaceBetween            // justify-between
}

// StyleBottomMenu matches React BottomMenu.tsx:
// className="fixed bottom-0 left-0 z-40 w-full h-16 bg-white border-t border-gray-light"
func StyleBottomMenu(s *styles.Style) {
	s.Min.X.Set(100, units.UnitPw)                     // w-full
	s.Min.Y.Set(64, units.UnitDp)                      // h-16 (64px)
	s.Background = colors.Uniform(ColorWhite)          // bg-white
	s.Border.Style.Top = styles.BorderSolid            // border-t
	s.Border.Width.Top = units.Dp(1)                   // border-t
	s.Border.Color.Set(colors.Uniform(ColorGrayLight)) // border-gray-light
	s.Display = styles.Flex                            // flex
	s.Direction = styles.Row                           // flex-row
	s.Justify.Content = styles.SpaceAround             // justify-around (for items)
	s.Align.Items = styles.Center                      // items-center
}

// ====================================================================================
// Action Bars and Button Rows
// ====================================================================================

func StyleActionsRow(s *styles.Style) {
	s.Direction = styles.Row
	s.Gap.Set(units.Dp(Spacing3))
	s.Align.Items = styles.Center
	s.Margin.Bottom = units.Dp(Spacing4)
}

// ActionBar - matches React's h-12 w-full flex items-center justify-end pattern
func StyleActionBar(s *styles.Style) {
	s.Display = styles.Flex
	s.Min.Y.Set(48, units.UnitDp)                      // h-12 (48px)
	s.Min.X.Set(100, units.UnitPw)                     // w-full (parent width)
	s.Align.Items = styles.Center                      // items-center
	s.Justify.Content = styles.End                     // justify-end (matching React GroupsPage)
}

func StyleActionsSplit(s *styles.Style) {
	s.Direction = styles.Row
	s.Gap.Set(units.Dp(Spacing3))
	s.Justify.Content = styles.End  // Align buttons to the right (matching React justify-end)
	s.Min.Y.Set(48, units.UnitDp)   // h-12 container height
	s.Min.X.Set(100, units.UnitPw)  // w-full (parent width) - CRITICAL: container must span full width for justify-end to work
}

// ====================================================================================
// Grid and Container Layouts
// ====================================================================================

func StyleGridContainer(s *styles.Style) {
	s.Display = styles.Grid
	s.Gap.Set(units.Dp(Spacing4))
	s.Wrap = true
}

// Grid patterns: className="grid grid-cols-2 gap-6"
func StyleGrid2Cols(s *styles.Style) {
	s.Display = styles.Grid                // grid
	s.Gap.Set(units.Dp(Spacing6))          // gap-6
	// Note: grid-cols-2 would need CSS Grid implementation in Cogent Core
}

// Navigation button grids: className="grid grid-cols-2 gap-4 w-full"
func StyleNavGrid(s *styles.Style) {
	s.Display = styles.Grid                // grid
	s.Gap.Set(units.Dp(Spacing4))          // gap-4
	s.Min.X.Set(100, units.UnitPw)         // w-full (parent width)
	// Note: grid-cols-2 would need CSS Grid implementation
}

func StyleObjectsGrid(s *styles.Style) {
	s.Direction = styles.Row
	s.Wrap = true
	s.Gap.Set(units.Dp(Spacing4))
}

func StyleCollectionsGrid(s *styles.Style) {
	s.Direction = styles.Row
	s.Wrap = true
	s.Gap.Set(units.Dp(Spacing4))
}

func StyleNavContainer(s *styles.Style) {
	s.Direction = styles.Column  // Stack buttons vertically on mobile
	s.Gap.Set(units.Dp(Spacing3))
	// Items in column layout naturally take full width of parent
}

// ====================================================================================
// Stats and Info Containers
// ====================================================================================

func StyleStatsContainer(s *styles.Style) {
	s.Direction = styles.Column
	s.Gap.Set(units.Dp(Spacing3))
	s.Background = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(RadiusDefault)) // rounded
	s.Padding.Set(units.Dp(Spacing4))
}

func StyleStatsGrid(s *styles.Style) {
	s.Direction = styles.Row
	s.Gap.Set(units.Dp(Spacing4))
	s.Wrap = true
}

func StyleStatsRow(s *styles.Style) {
	s.Direction = styles.Row
	s.Gap.Set(units.Dp(Spacing4))
	s.Justify.Content = styles.SpaceBetween
	s.Background = colors.Uniform(ColorGrayLightest)
	s.Border.Radius = sides.NewValues(units.Dp(RadiusLG)) // BorderRadiusMedium
	s.Padding.Set(units.Dp(Spacing3))
}

func StyleStatColumn(s *styles.Style) {
	s.Direction = styles.Column
	s.Align.Items = styles.Center
	s.Gap.Set(units.Dp(Spacing0_5))
}

// ====================================================================================
// Search and Filter Containers
// ====================================================================================

func StyleSearchContainer(s *styles.Style) {
	s.Direction = styles.Row
	s.Gap.Set(units.Dp(Spacing3))
	s.Align.Items = styles.Center
	s.Background = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(Radius2XL)) // BorderRadiusLarge
	s.Padding.Set(units.Dp(Spacing4))
}

func StyleSearchField(s *styles.Style) {
	s.Grow.Set(1, 1)
	s.Border.Style.Set(styles.BorderNone)
}

// Search bar with margin: className="mb-2"
func StyleSearchBar(s *styles.Style) {
	s.Margin.Bottom = units.Dp(Spacing2) // mb-2
}

func StyleSearchSection(s *styles.Style) {
	s.Direction = styles.Row
	s.Gap.Set(units.Dp(Spacing2))
	s.Align.Items = styles.Center
}

func StyleSearchFieldContainer(s *styles.Style) {
	s.Min.X.Set(200, units.UnitDp)
}

func StyleFilterContainer(s *styles.Style) {
	s.Direction = styles.Column
	s.Background = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(Radius2XL)) // BorderRadiusLarge
	s.Padding.Set(units.Dp(Spacing4))
	s.Gap.Set(units.Dp(Spacing4))
}

func StyleFilterRow(s *styles.Style) {
	s.Direction = styles.Row
	s.Wrap = true
	s.Gap.Set(units.Dp(Spacing3))
}

func StyleActiveFiltersContainer(s *styles.Style) {
	s.Direction = styles.Column
	s.Background = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(Radius2XL)) // BorderRadiusLarge
	s.Padding.Set(units.Dp(Spacing4))
	s.Gap.Set(units.Dp(Spacing3))
}

func StyleActiveFiltersRow(s *styles.Style) {
	s.Direction = styles.Row
	s.Wrap = true
	s.Gap.Set(units.Dp(Spacing2))
}

// ====================================================================================
// Form and Input Layouts
// ====================================================================================

// Form patterns: className="flex flex-col gap-4" (common drawer body)
func StyleFormContainer(s *styles.Style) {
	s.Display = styles.Flex                         // flex
	s.Direction = styles.Column                     // flex-col
	s.Gap.Set(units.Dp(Spacing4))                   // gap-4
}

// className="flex flex-wrap gap-1.5 whitespace-nowrap"
func StyleFlexWrap(s *styles.Style) {
	s.Display = styles.Flex                         // flex
	s.Wrap = true                                   // flex-wrap
	s.Gap.Set(units.Dp(Spacing1_5))                 // gap-1.5
	// Note: whitespace-nowrap would need text wrapping control
}

func StylePropertiesContainer(s *styles.Style) {
	s.Direction = styles.Column
	s.Gap.Set(units.Dp(Spacing2))                       // gap-2 (8px)
	s.Background = colors.Uniform(ColorGrayLightest)
	s.Border.Radius = styles.BorderRadiusMedium
	s.Padding.Set(units.Dp(Spacing3))                   // p-3 (12px)
}

func StylePropertyRow(s *styles.Style) {
	s.Direction = styles.Row
	s.Justify.Content = styles.SpaceBetween
}

func StylePropertyRowHighlight(s *styles.Style) {
	s.Direction = styles.Row
	s.Justify.Content = styles.SpaceBetween
	s.Padding.Set(units.Dp(Spacing2))
	s.Background = colors.Uniform(ColorGrayLightest)
	s.Border.Radius = sides.NewValues(units.Dp(RadiusLG)) // BorderRadiusMedium
}

func StylePropertiesFormContainer(s *styles.Style) {
	s.Direction = styles.Column
	s.Gap.Set(units.Dp(Spacing2))
	s.Background = colors.Uniform(ColorGrayLightest)
	s.Border.Radius = sides.NewValues(units.Dp(RadiusLG)) // BorderRadiusMedium
	s.Padding.Set(units.Dp(Spacing3))
}

// ====================================================================================
// Card and Detail View Layouts
// ====================================================================================

func StyleCollectionCardHeader(s *styles.Style) {
	s.Direction = styles.Row
	s.Align.Items = styles.Center
	s.Justify.Content = styles.SpaceBetween
}

func StyleCollectionTitleSection(s *styles.Style) {
	s.Direction = styles.Row
	s.Align.Items = styles.Center
	s.Gap.Set(units.Dp(Spacing2))
	s.Grow.Set(1, 0)
}

func StyleCollectionTitleContainer(s *styles.Style) {
	s.Direction = styles.Column
	s.Gap.Set(units.Dp(Spacing0_5))
}

func StyleObjectCardHeader(s *styles.Style) {
	s.Direction = styles.Row
	s.Align.Items = styles.Center
	s.Justify.Content = styles.SpaceBetween
}

func StyleObjectCardNameSection(s *styles.Style) {
	s.Direction = styles.Row
	s.Align.Items = styles.Center
	s.Gap.Set(units.Dp(Spacing2))
	s.Grow.Set(1, 0)
}

func StyleContainerInfoSection(s *styles.Style) {
	s.Direction = styles.Row
	s.Align.Items = styles.Center
	s.Gap.Set(units.Dp(Spacing3))
	s.Grow.Set(1, 0)
	s.Cursor = cursors.Pointer
}

func StyleContainerDetails(s *styles.Style) {
	s.Direction = styles.Column
	s.Gap.Set(units.Dp(Spacing1))
}

func StyleObjectDetailHeader(s *styles.Style) {
	s.Direction = styles.Row
	s.Align.Items = styles.Center
	s.Gap.Set(units.Dp(Spacing3))
}

func StyleObjectDetailTitleContainer(s *styles.Style) {
	s.Direction = styles.Column
	s.Gap.Set(units.Dp(Spacing0_5))
}

// ====================================================================================
// Breadcrumb Navigation
// ====================================================================================

func StyleBreadcrumbContainer(s *styles.Style) {
	s.Background = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(Radius2XL)) // BorderRadiusLarge
	s.Padding.Set(units.Dp(Spacing4))
}

func StyleBreadcrumbRow(s *styles.Style) {
	s.Direction = styles.Row
	s.Align.Items = styles.Center
	s.Gap.Set(units.Dp(Spacing2))
}

// ====================================================================================
// Login and Loading Layouts
// ====================================================================================

// className="flex items-center justify-center h-screen"
// This is applied to mainContainer for the login page
// Using Column direction so Justify.Content centers vertically (main axis)
func StyleLoginContainer(s *styles.Style) {
	s.Display = styles.Flex            // flex
	s.Direction = styles.Column        // Column: main axis = vertical
	s.Justify.Content = styles.Center  // justify-center (vertical centering on main axis)
	s.Align.Items = styles.Center      // items-center (horizontal centering on cross axis)
	s.Align.Content = styles.Center    // Also center the content collection
	s.Grow.Set(1, 1)                   // Grow to fill parent container (Body)
}

// className="transform flex flex-col items-center justify-center"
func StyleLoginContent(s *styles.Style) {
	s.Display = styles.Flex           // flex
	s.Direction = styles.Column       // flex-col
	s.Align.Items = styles.Center     // items-center (center children horizontally)
	s.Justify.Content = styles.Center // justify-center (center children vertically)
	s.Grow.Set(1, 0)                  // Grow to fill parent width (but not force height)
}

// Loading Patterns: className="min-h-screen flex items-center justify-center"
func StyleLoadingScreen(s *styles.Style) {
	s.Min.Y.Set(100, units.UnitVh)    // min-h-screen
	s.Display = styles.Flex           // flex
	s.Align.Items = styles.Center     // items-center
	s.Justify.Content = styles.Center // justify-center
}

// ====================================================================================
// Misc Layout Patterns
// ====================================================================================

func StyleDevSection(s *styles.Style) {
	s.Direction = styles.Column
	s.Background = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(RadiusDefault)) // rounded
	s.Padding.Set(units.Dp(Spacing4))
	s.Gap.Set(units.Dp(Spacing3))
	s.Margin.Top = units.Dp(Spacing4)
}

func StyleViewModeRow(s *styles.Style) {
	s.Direction = styles.Row
	s.Gap.Set(units.Dp(Spacing2))
	s.Justify.Content = styles.End
}

func StyleAddSection(s *styles.Style) {
	s.Direction = styles.Row
	s.Gap.Set(units.Dp(Spacing2))
}

func StyleTagsSection(s *styles.Style) {
	s.Direction = styles.Column
	s.Gap.Set(units.Dp(Spacing2))
}

func StyleTagsGrid(s *styles.Style) {
	s.Direction = styles.Row
	s.Wrap = true
	s.Gap.Set(units.Dp(Spacing2))
}

func StyleObjectTypeContainer(s *styles.Style) {
	s.Direction = styles.Row
	s.Wrap = true
	s.Gap.Set(units.Dp(Spacing2))
}

func StyleExpiryContainer(s *styles.Style) {
	s.Direction = styles.Row
	s.Gap.Set(units.Dp(Spacing2))
	s.Align.Items = styles.Center
}

func StyleQuantityContainer(s *styles.Style) {
	s.Display = styles.Flex       // flex
	s.Align.Items = styles.Center // items-center
	s.Gap.Set(units.Dp(Spacing1)) // gap-1
}

// Empty state patterns: className="text-center p-8 text-gray-dark"
func StyleEmptyState(s *styles.Style) {
	s.Text.Align = text.Center              // text-center
	s.Padding.Set(units.Dp(Spacing8))       // p-8
	s.Color = colors.Uniform(ColorGrayDark) // text-gray-dark
	s.Max.X.Set(400, units.UnitDp)          // Max width to prevent awkward text wrapping
}

// ====================================================================================
// Constants for Spacing Values that Need Specific Values
// ====================================================================================

const (
	Spacing0_5  = 2  // 0.125rem - gap-0.5 / spacing-0.5
	Spacing1_5  = 6  // 0.375rem - gap-1.5 / spacing-1.5
	Spacing2_5  = 10 // 0.625rem - gap-2.5 / spacing-2.5
	Spacing3_5  = 14 // 0.875rem - gap-3.5 / spacing-3.5
	Spacing7    = 28 // 1.75rem - gap-7 / spacing-7
	Spacing9    = 36 // 2.25rem - gap-9 / spacing-9
	Spacing11   = 44 // 2.75rem - gap-11 / spacing-11
	Spacing14   = 56 // 3.5rem - gap-14 / spacing-14
	Spacing32   = 128 // 8rem - gap-32 / spacing-32
)
