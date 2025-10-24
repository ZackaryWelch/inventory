package styles

import (
	"image/color"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/cursors"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/sides"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/text/text"
)

// ====================================================================================
// UTILITY STYLES - Text, Icons, and Other Utilities
// ====================================================================================

// ====================================================================================
// Typography Hierarchy (matching nishiki-frontend Typography components)
// ====================================================================================

// H1: className="text-2xl"
func StyleH1(s *styles.Style) {
	s.Font.Size = units.Dp(FontSize2XL) // text-2xl
}

// H2: className="text-xl"
func StyleH2(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeXL) // text-xl
}

// H3: className="text-lg"
func StyleH3(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeLG) // text-lg
}

func StyleAppTitle(s *styles.Style) {
	s.Font.Size = units.Dp(FontSize2XL)         // text-2xl
	s.Font.Weight = WeightBold
	s.Color = colors.Uniform(ColorPrimary)
}

func StyleSubtitle(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeBase)        // text-base
	s.Color = colors.Uniform(ColorGrayDark)
}

func StyleSectionTitle(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeXL)          // text-xl (matching H2 component)
	s.Font.Weight = WeightSemiBold
	s.Color = colors.Uniform(ColorBlack)
}

func StyleCardTitle(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeLG)          // text-lg (matching group card titles)
	s.Font.Weight = WeightSemiBold
	s.Margin.Bottom = units.Dp(Spacing0_5)      // Simulate React's "leading-6" line height spacing
}

func StyleSmallText(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeXS)          // text-xs
	s.Color = colors.Uniform(ColorGrayDark)
}

// Description text styling (matching frontend text-sm text-gray-dark)
func StyleDescriptionText(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeSM)          // text-sm
	s.Color = colors.Uniform(ColorGrayDark)
}

// ====================================================================================
// Text Patterns from Frontend
// ====================================================================================

// className="text-xs text-gray-dark flex items-center gap-1 my-1.5"
func StyleTextXsGrayWithIcon(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeXS)          // text-xs
	s.Color = colors.Uniform(ColorGrayDark)     // text-gray-dark
	s.Display = styles.Flex                     // flex
	s.Align.Items = styles.Center               // items-center
	s.Gap.Set(units.Dp(Spacing1))               // gap-1
	s.Margin.Set(units.Dp(Spacing1_5), units.Dp(0)) // my-1.5
}

// className="text-sm flex items-center gap-1"
func StyleTextSmWithIcon(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeSM)          // text-sm
	s.Display = styles.Flex                     // flex
	s.Align.Items = styles.Center               // items-center
	s.Gap.Set(units.Dp(Spacing1))               // gap-1
}

func StyleStatsTitle(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeLG)          // text-lg
	s.Font.Weight = WeightSemiBold
}

func StyleDevTitle(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeBase)        // text-base
	s.Font.Weight = WeightSemiBold
	s.Color = colors.Uniform(ColorGrayDark)
}

func StyleUserFieldLabel(s *styles.Style) {
	s.Font.Weight = WeightSemiBold
	s.Color = colors.Uniform(ColorGrayDark)
}

func StyleEmptyText(s *styles.Style) {
	s.Color = colors.Uniform(ColorGrayDark)
	s.Align.Self = styles.Center
	s.Margin.Top = units.Dp(Spacing8)
}

// Stat text styles
func StyleStatValue(s *styles.Style) {
	s.Font.Size = units.Dp(FontSize2XL)         // text-2xl
	s.Font.Weight = WeightBold
	s.Color = colors.Uniform(ColorWhite)
}

func StyleStatLabel(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeSM)          // text-sm
	s.Color = colors.Uniform(ColorWhite)
}

func StyleStatValuePrimary(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeLG)          // text-lg
	s.Font.Weight = WeightBold
	s.Color = colors.Uniform(ColorPrimary)
}

func StyleStatValueAccent(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeLG)          // text-lg
	s.Font.Weight = WeightBold
	s.Color = colors.Uniform(ColorAccent)
}

// ====================================================================================
// Object and Card Text Styles
// ====================================================================================

func StyleObjectCardName(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeBase)        // text-base
	s.Font.Weight = WeightSemiBold
	s.Color = colors.Uniform(ColorBlack)
}

func StyleObjectCardDescription(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeSM)          // text-sm
	s.Color = colors.Uniform(ColorGrayDark)
}

func StyleCollectionName(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeBase)        // text-base
	s.Font.Weight = WeightSemiBold
	s.Color = colors.Uniform(ColorBlack)
}

func StyleCollectionType(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeXS)          // text-xs
	s.Color = colors.Uniform(ColorGrayDark)
}

func StyleContainerName(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeBase)        // text-base
	s.Font.Weight = WeightSemiBold
}

func StyleContainerDescription(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeSM)          // text-sm
	s.Color = colors.Uniform(ColorGrayDark)
}

func StyleContainerCount(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeXS)          // text-xs
	s.Color = colors.Uniform(ColorGrayDark)
}

// ====================================================================================
// Property Display Text Styles
// ====================================================================================

func StylePropertiesTitle(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeXS)          // text-xs
	s.Font.Weight = WeightSemiBold
	s.Color = colors.Uniform(ColorGrayDark)
}

func StylePropertyKey(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeXS)          // text-xs
	s.Color = colors.Uniform(ColorGrayDark)
}

func StylePropertyValue(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeXS)          // text-xs
	s.Font.Weight = WeightMedium
}

// ====================================================================================
// Filter and Search Text Styles
// ====================================================================================

func StyleFilterLabel(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeXS)          // text-xs
	s.Font.Weight = WeightSemiBold
	s.Color = colors.Uniform(ColorGrayDark)
}

func StyleFilterTitle(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeLG)          // text-lg
	s.Font.Weight = WeightSemiBold
}

func StyleFilterSubtitle(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeSM)          // text-sm
	s.Font.Weight = WeightSemiBold
}

func StyleSectionSubtitle(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeBase)        // text-base
	s.Font.Weight = WeightSemiBold
}

func StyleObjectTitle(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeXL)          // text-xl
	s.Font.Weight = WeightBold
}

func StyleMoreText(s *styles.Style) {
	s.Font.Size = units.Dp(FontSize2XS)         // text-xs
	s.Color = colors.Uniform(ColorGrayDark)
}

func StyleActiveFilterText(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeXS)          // text-xs
	s.Color = colors.Uniform(ColorPrimary)
}

func StyleSearchResultTitle(s *styles.Style) {
	s.Font.Weight = WeightSemiBold
}

func StyleSearchResultDescription(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeXS)          // text-xs
	s.Color = colors.Uniform(ColorGrayDark)
}

func StyleSearchResultPath(s *styles.Style) {
	s.Font.Size = units.Dp(FontSize2XS)         // text-xs
	s.Color = colors.Uniform(ColorPrimary)
}

// ====================================================================================
// Food Card Text Styles
// ====================================================================================

// FoodCard container info: className="text-xs text-gray-dark flex items-center gap-1 my-1.5"
func StyleFoodCardContainerInfo(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeXS)              // text-xs
	s.Color = colors.Uniform(ColorGrayDark)         // text-gray-dark
	s.Display = styles.Flex                         // flex
	s.Align.Items = styles.Center                   // items-center
	s.Gap.Set(units.Dp(Spacing1))                   // gap-1
	s.Margin.Set(units.Dp(Spacing1_5), units.Dp(0)) // my-1.5
}

// FoodCard quantity info: className="text-sm flex items-center gap-1"
func StyleFoodCardQuantityInfo(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeSM)              // text-sm
	s.Display = styles.Flex                         // flex
	s.Align.Items = styles.Center                   // items-center
	s.Gap.Set(units.Dp(Spacing1))                   // gap-1
}

// FoodCard time display: className="ml-auto"
func StyleFoodCardTime(s *styles.Style) {
	// Note: ml-auto equivalent - push to right via parent layout
	s.Margin.Left = units.Dp(-1) // ml-auto
}

func StyleTimeDisplay(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeSM)          // text-sm (matching quantity)
	s.Color = colors.Uniform(ColorGrayDark)     // subtle color for dates
}

// ====================================================================================
// Text Alignment Utilities
// ====================================================================================

func StyleTextLeft(s *styles.Style) {
	s.Text.Align = AlignStart // text-left
}

func StyleTextRight(s *styles.Style) {
	s.Text.Align = AlignEnd // text-right
}

func StyleTextCenter(s *styles.Style) {
	s.Text.Align = text.Center // text-center
}

func StyleTextAutoRight(s *styles.Style) {
	s.Margin.Left = units.Dp(-1) // ml-auto equivalent (grow to push right)
}

// ====================================================================================
// Breadcrumb Text Styles
// ====================================================================================

func StyleBreadcrumbLink(s *styles.Style) {
	s.Color = colors.Uniform(ColorPrimary)
	s.Cursor = cursors.Pointer
}

func StyleBreadcrumbArrow(s *styles.Style) {
	s.Color = colors.Uniform(ColorGrayDark)
}

func StyleBreadcrumbCurrent(s *styles.Style) {
	s.Font.Weight = WeightSemiBold
}

// ====================================================================================
// Icon System (from nishiki-frontend Icon.tsx)
// ====================================================================================

// Complete icon size variants matching frontend exactly
func StyleIconSize2(s *styles.Style) {
	s.Min.X.Set(8, units.UnitDp) // w-2 h-2
	s.Min.Y.Set(8, units.UnitDp)
}

func StyleIconSize2_5(s *styles.Style) {
	s.Min.X.Set(10, units.UnitDp) // w-2.5 h-2.5
	s.Min.Y.Set(10, units.UnitDp)
}

func StyleIconSize3(s *styles.Style) {
	s.Min.X.Set(12, units.UnitDp) // w-3 h-3
	s.Min.Y.Set(12, units.UnitDp)
}

func StyleIconSize3_5(s *styles.Style) {
	s.Min.X.Set(14, units.UnitDp) // w-3.5 h-3.5
	s.Min.Y.Set(14, units.UnitDp)
}

func StyleIconSize4(s *styles.Style) {
	s.Min.X.Set(16, units.UnitDp) // w-4 h-4
	s.Min.Y.Set(16, units.UnitDp)
}

func StyleIconSize4_5(s *styles.Style) {
	s.Min.X.Set(18, units.UnitDp) // w-4.5 h-4.5 (custom size)
	s.Min.Y.Set(18, units.UnitDp)
}

func StyleIconSize5(s *styles.Style) {
	s.Min.X.Set(20, units.UnitDp) // w-5 h-5
	s.Min.Y.Set(20, units.UnitDp)
}

func StyleIconSize6(s *styles.Style) {
	s.Min.X.Set(24, units.UnitDp) // w-6 h-6
	s.Min.Y.Set(24, units.UnitDp)
}

func StyleIconSize7(s *styles.Style) {
	s.Min.X.Set(28, units.UnitDp) // w-7 h-7
	s.Min.Y.Set(28, units.UnitDp)
}

func StyleIconSize8(s *styles.Style) {
	s.Min.X.Set(32, units.UnitDp) // w-8 h-8
	s.Min.Y.Set(32, units.UnitDp)
}

func StyleIconSize9(s *styles.Style) {
	s.Min.X.Set(36, units.UnitDp) // w-9 h-9
	s.Min.Y.Set(36, units.UnitDp)
}

func StyleIconSize10(s *styles.Style) {
	s.Min.X.Set(40, units.UnitDp) // w-10 h-10
	s.Min.Y.Set(40, units.UnitDp)
}

func StyleIconSize11(s *styles.Style) {
	s.Min.X.Set(44, units.UnitDp) // w-11 h-11
	s.Min.Y.Set(44, units.UnitDp)
}

func StyleIconSize12(s *styles.Style) {
	s.Min.X.Set(48, units.UnitDp) // w-12 h-12
	s.Min.Y.Set(48, units.UnitDp)
}

func StyleIconSize14(s *styles.Style) {
	s.Min.X.Set(56, units.UnitDp) // w-14 h-14
	s.Min.Y.Set(56, units.UnitDp)
}

func StyleIconSize16(s *styles.Style) {
	s.Min.X.Set(64, units.UnitDp) // w-16 h-16
	s.Min.Y.Set(64, units.UnitDp)
}

// Icon color variants (from Icon.tsx)
func StyleIconWhite(s *styles.Style) {
	s.Color = colors.Uniform(ColorWhite) // text-white
}

func StyleIconBlack(s *styles.Style) {
	s.Color = colors.Uniform(ColorBlack) // text-black
}

func StyleIconPrimary(s *styles.Style) {
	s.Color = colors.Uniform(ColorPrimary) // text-primary
}

func StyleIconDanger(s *styles.Style) {
	s.Color = colors.Uniform(ColorDanger) // text-danger
}

func StyleIconGray(s *styles.Style) {
	s.Color = colors.Uniform(ColorGray) // text-gray
}

func StyleIconGrayDark(s *styles.Style) {
	s.Color = colors.Uniform(ColorGrayDark) // text-gray-dark
}

func StyleObjectDetailIcon(s *styles.Style) {
	s.Font.Size = units.Dp(FontSize2XL) // text-2xl for larger icons
}

// ====================================================================================
// Action Button Styles
// ====================================================================================

func StyleActionButtonEdit(s *styles.Style) {
	s.Background = colors.Uniform(ColorAccent)
	s.Color = colors.Uniform(ColorBlack)
	s.Border.Radius = sides.NewValues(units.Dp(RadiusFull))
	s.Padding.Set(units.Dp(Spacing1_5))
}

func StyleActionButtonDelete(s *styles.Style) {
	s.Background = colors.Uniform(ColorDanger)
	s.Color = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(RadiusFull))
	s.Padding.Set(units.Dp(Spacing1_5))
}

func StyleActionButtonInvite(s *styles.Style) {
	s.Background = colors.Uniform(ColorPrimary)
	s.Color = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(RadiusFull))
	s.Padding.Set(units.Dp(Spacing2))
}

func StyleActionButtonRemove(s *styles.Style) {
	s.Background = colors.Uniform(ColorDanger)
	s.Color = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(RadiusFull))
	s.Padding.Set(units.Dp(Spacing1_5))
}

func StyleActionButtonLarge(s *styles.Style) {
	s.Background = colors.Uniform(ColorPrimary)
	s.Color = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(Radius2XL)) // BorderRadiusLarge
	s.Padding.Set(units.Dp(Spacing3), units.Dp(Spacing4))
	s.Gap.Set(units.Dp(Spacing2))
}

func StyleActionButtonLargeAccent(s *styles.Style) {
	s.Background = colors.Uniform(ColorAccent)
	s.Color = colors.Uniform(ColorBlack)
	s.Border.Radius = sides.NewValues(units.Dp(Radius2XL)) // BorderRadiusLarge
	s.Padding.Set(units.Dp(Spacing3), units.Dp(Spacing4))
	s.Gap.Set(units.Dp(Spacing2))
}

func StyleActionButtonLargeDanger(s *styles.Style) {
	s.Background = colors.Uniform(ColorDanger)
	s.Color = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(Radius2XL)) // BorderRadiusLarge
	s.Padding.Set(units.Dp(Spacing3), units.Dp(Spacing4))
	s.Gap.Set(units.Dp(Spacing2))
}

func StyleSearchResultAction(s *styles.Style) {
	s.Background = colors.Uniform(ColorPrimary)
	s.Color = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(RadiusLG)) // BorderRadiusMedium
	s.Padding.Set(units.Dp(Spacing1_5), units.Dp(Spacing3))
	s.Gap.Set(units.Dp(Spacing1))
}

func StyleContainerActionsMenu(s *styles.Style) {
	s.Direction = styles.Row
	s.Gap.Set(units.Dp(Spacing2))
}

func StyleObjectCardActionsMenu(s *styles.Style) {
	s.Direction = styles.Row
	s.Gap.Set(units.Dp(Spacing1))
}

// ====================================================================================
// Specific Button Styles
// ====================================================================================

func StyleNavButton(s *styles.Style) {
	s.Background = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(RadiusDefault)) // rounded
	s.Padding.Set(units.Dp(Spacing4))
	s.Gap.Set(units.Dp(Spacing2))
	s.Min.X.Set(120, units.UnitDp)
	s.Border.Style.Set(styles.BorderSolid)
	s.Border.Width.Set(units.Dp(1))
	s.Border.Color.Set(colors.Uniform(ColorGrayLight))
}

func StyleUserButton(s *styles.Style) {
	s.Background = colors.Uniform(color.RGBA{R: 240, G: 240, B: 240, A: 255})
	s.Border.Radius = sides.NewValues(units.Dp(RadiusDefault)) // rounded
	s.Padding.Set(units.Dp(Spacing2), units.Dp(Spacing4))
}

func StyleBackButton(s *styles.Style) {
	s.Background = colors.Uniform(color.RGBA{R: 240, G: 240, B: 240, A: 255})
	s.Border.Radius = sides.NewValues(units.Dp(RadiusFull)) // rounded-full
	s.Padding.Set(units.Dp(Spacing2))
}

func StyleLogoutButton(s *styles.Style) {
	s.Align.Self = styles.Start
	s.Margin.Top = units.Dp(Spacing4)
}

func StyleCreateButton(s *styles.Style) {
	s.Align.Self = styles.End
}

func StyleViewButton(s *styles.Style) {
	s.Background = colors.Uniform(ColorPrimary)
	s.Color = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(RadiusDefault)) // rounded
	s.Padding.Set(units.Dp(Spacing2), units.Dp(Spacing3))
	s.Gap.Set(units.Dp(Spacing1))
}

func StyleViewButtonAccent(s *styles.Style) {
	s.Background = colors.Uniform(ColorAccent)
	s.Color = colors.Uniform(ColorBlack)
	s.Border.Radius = sides.NewValues(units.Dp(RadiusDefault)) // rounded
	s.Padding.Set(units.Dp(Spacing2), units.Dp(Spacing3))
	s.Gap.Set(units.Dp(Spacing1))
}

func StyleClearCacheButton(s *styles.Style) {
	s.Background = colors.Uniform(ColorAccent)
	s.Color = colors.Uniform(ColorBlack)
	s.Border.Radius = sides.NewValues(units.Dp(RadiusDefault)) // rounded
	s.Padding.Set(units.Dp(Spacing2), units.Dp(Spacing3))
	s.Gap.Set(units.Dp(Spacing1))
	s.Align.Self = styles.Start
}

func StyleViewModeButtonActive(s *styles.Style) {
	s.Background = colors.Uniform(ColorPrimary)
	s.Color = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(RadiusLG)) // BorderRadiusMedium
	s.Padding.Set(units.Dp(Spacing2))
}

func StyleViewModeButtonInactive(s *styles.Style) {
	s.Background = colors.Uniform(color.RGBA{R: 240, G: 240, B: 240, A: 255})
	s.Border.Radius = sides.NewValues(units.Dp(RadiusLG)) // BorderRadiusMedium
	s.Padding.Set(units.Dp(Spacing2))
}

func StyleObjectTypeButtonSelected(s *styles.Style) {
	s.Background = colors.Uniform(ColorPrimary)
	s.Color = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(RadiusLG)) // BorderRadiusMedium
	s.Padding.Set(units.Dp(Spacing2), units.Dp(Spacing3))
}

func StyleObjectTypeButtonUnselected(s *styles.Style) {
	s.Background = colors.Uniform(color.RGBA{R: 240, G: 240, B: 240, A: 255})
	s.Color = colors.Uniform(ColorBlack)
	s.Border.Radius = sides.NewValues(units.Dp(RadiusLG)) // BorderRadiusMedium
	s.Padding.Set(units.Dp(Spacing2), units.Dp(Spacing3))
}

func StyleTagButton(s *styles.Style) {
	s.Background = colors.Uniform(color.RGBA{R: 240, G: 240, B: 240, A: 255})
	s.Border.Radius = sides.NewValues(units.Dp(RadiusFull))
	s.Padding.Set(units.Dp(Spacing1_5), units.Dp(Spacing3))
}

func StyleActiveFilterBadge(s *styles.Style) {
	s.Direction = styles.Row
	s.Align.Items = styles.Center
	s.Gap.Set(units.Dp(Spacing1))
	s.Background = colors.Uniform(ColorPrimaryLightest)
	s.Border.Radius = sides.NewValues(units.Dp(RadiusFull))
	s.Padding.Set(units.Dp(Spacing1_5), units.Dp(Spacing3))
}

func StyleActiveFilterRemove(s *styles.Style) {
	s.Background = colors.Uniform(ColorPrimary)
	s.Color = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(RadiusFull))
	s.Padding.Set(units.Dp(Spacing0_5))
	s.Font.Size = units.Dp(FontSize2XS) // text-xs
}

func StyleClearFiltersButton(s *styles.Style) {
	s.Background = colors.Uniform(color.RGBA{R: 240, G: 240, B: 240, A: 255})
	s.Border.Radius = sides.NewValues(units.Dp(RadiusLG)) // BorderRadiusMedium
	s.Padding.Set(units.Dp(Spacing2), units.Dp(Spacing3))
	s.Align.Self = styles.Start
}

// ====================================================================================
// Loading and Spinner Styles
// ====================================================================================

// className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"
func StyleLoadingSpinner(s *styles.Style) {
	s.Border.Radius = sides.NewValues(units.Dp(RadiusFull))     // rounded-full
	s.Min.Y.Set(48, units.UnitDp)                               // h-12
	s.Min.X.Set(48, units.UnitDp)                               // w-12
	s.Border.Style.Set(styles.BorderSolid)                      // border
	s.Border.Width.Bottom = units.Dp(2)                         // border-b-2
	s.Border.Color.Set(colors.Uniform(ColorPrimary))            // border-primary (using primary instead of blue-600)
	s.Margin.Set(units.Dp(0), units.Dp(-1), units.Dp(Spacing4), units.Dp(-1)) // mx-auto mb-4
}

// LoadingSkeleton - matches React's loading card skeletons
func StyleLoadingSkeleton(s *styles.Style) {
	s.Background = colors.Uniform(color.RGBA{R: 229, G: 231, B: 235, A: 255}) // bg-gray-200
	s.Min.Y.Set(80, units.UnitDp)                                             // h-20 (matching React skeleton height)
	s.Border.Radius = sides.NewValues(units.Dp(RadiusLG))                     // rounded-lg
	s.Margin.Bottom = units.Dp(Spacing2)                                      // mb-2 (matching card spacing)
}

// Logo patterns: className="w-32 h-26 mb-20"
func StyleLoginLogo(s *styles.Style) {
	s.Min.X.Set(128, units.UnitDp)      // w-32
	s.Min.Y.Set(104, units.UnitDp)      // h-26 (104px)
	s.Margin.Bottom = units.Dp(Spacing20) // mb-20
}

// ====================================================================================
// Margin and Spacing Utilities
// ====================================================================================

func StyleMarginLeftAuto(s *styles.Style) {
	s.Margin.Left = units.Dp(-1) // ml-auto equivalent
}

func StyleMarginRightAuto(s *styles.Style) {
	s.Margin.Right = units.Dp(-1) // mr-auto equivalent
}

func StyleMarginAuto(s *styles.Style) {
	s.Margin.Left = units.Dp(-1)  // mx-auto equivalent
	s.Margin.Right = units.Dp(-1)
}

// ====================================================================================
// Grow and Flex Utilities
// ====================================================================================

func StyleGrow(s *styles.Style) {
	s.Grow.Set(1, 0) // grow (flex-grow: 1)
}

func StyleGrowFull(s *styles.Style) {
	s.Grow.Set(1, 1) // grow with shrink
}

// ====================================================================================
// Aspect Ratio Utilities
// ====================================================================================

func StyleAspectSquare(s *styles.Style) {
	// For square aspect ratio, set min width equal to min height
	// This assumes height is already set on the element
	s.Min.X.Set(100, units.UnitEh) // aspect-square equivalent
}

// ====================================================================================
// Filter Badge Styles
// ====================================================================================

// Badge with custom padding: className="pl-1 pr-0 gap-0"
func StyleFilterBadge(s *styles.Style) {
	StyleBadgeBase(s)
	s.Padding.Left = units.Dp(Spacing1)  // pl-1
	s.Padding.Right = units.Dp(0)        // pr-0
	s.Gap.Set(units.Dp(0))               // gap-0
}

// FilterBadge icon circle: className="bg-white w-4 h-4 rounded-full p-[3.5px] mr-1 flex items-center justify-center"
func StyleFilterBadgeIconCircle(s *styles.Style) {
	s.Background = colors.Uniform(ColorWhite)         // bg-white
	s.Min.X.Set(16, units.UnitDp)                     // w-4
	s.Min.Y.Set(16, units.UnitDp)                     // h-4
	s.Border.Radius = sides.NewValues(units.Dp(RadiusFull)) // rounded-full
	s.Padding.Set(units.Dp(3.5))                      // p-[3.5px]
	s.Margin.Right = units.Dp(Spacing1)               // mr-1
	s.Display = styles.Flex                           // flex
	s.Align.Items = styles.Center                     // items-center
	s.Justify.Content = styles.Center                 // justify-center
}

// FilterBadge emoji circle: className="bg-white w-4 h-4 rounded-full p-[3px] mr-1 flex items-center justify-center text-2xs select-none"
func StyleFilterBadgeEmojiCircle(s *styles.Style) {
	s.Background = colors.Uniform(ColorWhite)         // bg-white
	s.Min.X.Set(16, units.UnitDp)                     // w-4
	s.Min.Y.Set(16, units.UnitDp)                     // h-4
	s.Border.Radius = sides.NewValues(units.Dp(RadiusFull)) // rounded-full
	s.Padding.Set(units.Dp(3))                        // p-[3px]
	s.Margin.Right = units.Dp(Spacing1)               // mr-1
	s.Display = styles.Flex                           // flex
	s.Align.Items = styles.Center                     // items-center
	s.Justify.Content = styles.Center                 // justify-center
	s.Font.Size = units.Dp(FontSize2XS)               // text-2xs (10px)
	// Note: select-none would need text selection control
}

// FilterBadge close button: className="h-full w-6 flex items-center relative"
func StyleFilterBadgeCloseButton(s *styles.Style) {
	s.Min.Y.Set(100, units.UnitEh) // h-full
	s.Min.X.Set(24, units.UnitDp)  // w-6
	s.Display = styles.Flex        // flex
	s.Align.Items = styles.Center  // items-center
	// Note: relative positioning would need positioning system
}

// ====================================================================================
// Food Card Styles (Specific Patterns)
// ====================================================================================

// Complete FoodCard pattern: Card className="mb-2 w-full flex"
func StyleFoodCardContainer(s *styles.Style) {
	StyleCard(s)                       // Apply base card styles
	s.Margin.Bottom = units.Dp(Spacing2)  // mb-2
	s.Min.X.Set(100, units.UnitEw)     // w-full
	s.Display = styles.Flex            // flex
}

// FoodCard button: className="flex grow gap-4 items-center text-left pl-4 py-2"
func StyleFoodCardButton(s *styles.Style) {
	s.Display = styles.Flex                                                        // flex
	s.Grow.Set(1, 0)                                                               // grow
	s.Gap.Set(units.Dp(Spacing4))                                                  // gap-4
	s.Align.Items = styles.Center                                                  // items-center
	s.Text.Align = AlignStart                                                      // text-left
	s.Padding.Set(units.Dp(Spacing2), units.Dp(0), units.Dp(Spacing2), units.Dp(Spacing4)) // pl-4 py-2
	s.Background = colors.Uniform(color.RGBA{R: 0, G: 0, B: 0, A: 0})              // transparent button
	s.Border.Style.Set(styles.BorderNone)                                          // no border
	s.Cursor = cursors.Pointer
}

// FoodCard emoji figure: className="bg-white w-10 h-10 rounded-full flex items-center justify-center border border-primary select-none text-2xl"
func StyleFoodCardEmojiCircle(s *styles.Style) {
	s.Background = colors.Uniform(ColorWhite)               // bg-white
	s.Min.X.Set(40, units.UnitDp)                           // w-10
	s.Min.Y.Set(40, units.UnitDp)                           // h-10
	s.Border.Radius = sides.NewValues(units.Dp(RadiusFull)) // rounded-full
	s.Display = styles.Flex                                 // flex
	s.Align.Items = styles.Center                           // items-center
	s.Justify.Content = styles.Center                       // justify-center
	s.Border.Style.Set(styles.BorderSolid)                  // border
	s.Border.Width.Set(units.Dp(1))                         // border
	s.Border.Color.Set(colors.Uniform(ColorPrimary))        // border-primary
	s.Font.Size = units.Dp(FontSize2XL)                     // text-2xl
	// Note: select-none would need text selection control
}

// FoodCard content grow area: className="grow"
func StyleFoodCardContent(s *styles.Style) {
	s.Grow.Set(1, 0) // grow
}

func StyleSearchResultContent(s *styles.Style) {
	s.Direction = styles.Column
	s.Gap.Set(units.Dp(Spacing1))
	s.Grow.Set(1, 0)
}

func StyleExpiryField(s *styles.Style) {
	s.Min.X.Set(60, units.UnitDp)
}

// Filter dot pattern
// className="absolute -top-[3px] -right-[5px] w-2 h-2 rounded-full bg-danger"
func StyleFilterDot(s *styles.Style) {
	s.Min.X.Set(8, units.UnitDp)                      // w-2
	s.Min.Y.Set(8, units.UnitDp)                      // h-2
	s.Border.Radius = sides.NewValues(units.Dp(RadiusFull)) // rounded-full
	s.Background = colors.Uniform(ColorDanger)        // bg-danger
	// Note: absolute positioning (-top-[3px] -right-[5px]) would need positioning system
}

// Background stylers
func StyleMainBackground(s *styles.Style) {
	s.Direction = styles.Column
	s.Background = colors.Uniform(ColorGrayLightest) // #f8f8f8 (matching frontend)
}

// Select-none equivalent for emoji and icons
func StyleSelectNone(s *styles.Style) {
	// Note: text selection control would need additional implementation
	// This is a placeholder for the select-none class behavior
}

// ====================================================================================
// Stat Card Stylers (migrated from app/styles.go)
// ====================================================================================

// StyleStatCard creates a colored stat card with customizable background color
// Used for dashboard statistics (matches nishiki-frontend stat card pattern)
func StyleStatCard(cardColor color.RGBA) func(*styles.Style) {
	return func(s *styles.Style) {
		s.Direction = styles.Column
		s.Align.Items = styles.Center
		s.Background = colors.Uniform(cardColor)
		s.Border.Radius = sides.NewValues(units.Dp(RadiusDefault)) // rounded
		s.Padding.Set(units.Dp(Spacing4))
		s.Gap.Set(units.Dp(Spacing1))
		s.Min.X.Set(100, units.UnitDp)
	}
}

// ====================================================================================
// Hover State Stylers (migrated from app/styles.go)
// ====================================================================================

// StyleHoverPrimary applies primary color hover effect
func StyleHoverPrimary(s *styles.Style) {
	s.Cursor = cursors.Pointer
	// TODO: Add hover color change when Cogent Core supports it
}

// StyleHoverDanger applies danger color hover effect
func StyleHoverDanger(s *styles.Style) {
	s.Cursor = cursors.Pointer
	// TODO: Add hover color change when Cogent Core supports it
}

// StyleHoverGrayLight applies gray light hover effect
func StyleHoverGrayLight(s *styles.Style) {
	s.Cursor = cursors.Pointer
	s.Background = colors.Uniform(ColorGrayLight)
}
