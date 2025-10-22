package app

import (
	"image/color"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/cursors"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/sides"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/text/rich"
	"cogentcore.org/core/text/text"
)

// Font weight constants mapping from old styles to new rich weights
const (
	WeightNormal   = rich.Weights(3) // normal
	WeightMedium   = rich.Weights(4) // medium
	WeightSemiBold = rich.Weights(5) // semibold
	WeightBold     = rich.Weights(6) // bold
)

// Text alignment constants mapping from old styles to new text aligns
const (
	AlignStart  = text.Aligns(0) // start
	AlignCenter = text.Aligns(1) // center
	AlignEnd    = text.Aligns(2) // end
)

// Design system matching nishiki-frontend exactly
// Color palette from globals.css with exact hex values
var (
	// Primary colors (matching --color-primary-* from globals.css)
	ColorPrimaryLightest = color.RGBA{R: 230, G: 242, B: 241, A: 255} // #e6f2f1
	ColorPrimaryLight    = color.RGBA{R: 171, G: 212, B: 207, A: 255} // #abd4cf
	ColorPrimary         = color.RGBA{R: 106, G: 179, B: 171, A: 255} // #6ab3ab
	ColorPrimaryDark     = color.RGBA{R: 95, G: 161, B: 154, A: 255}  // #5fa19a

	// Accent colors (matching --color-accent-*)
	ColorAccent     = color.RGBA{R: 252, G: 216, B: 132, A: 255} // #fcd884
	ColorAccentDark = color.RGBA{R: 241, G: 197, B: 96, A: 255}  // #f1c560

	// Danger colors (matching --color-danger-*)
	ColorDanger     = color.RGBA{R: 205, G: 90, B: 90, A: 255} // #cd5a5a
	ColorDangerDark = color.RGBA{R: 185, G: 81, B: 81, A: 255} // #b95151

	// Gray scale (matching --color-gray-*)
	ColorGrayLightest = color.RGBA{R: 248, G: 248, B: 248, A: 255} // #f8f8f8
	ColorGrayLight    = color.RGBA{R: 238, G: 238, B: 238, A: 255} // #eeeeee
	ColorGray         = color.RGBA{R: 189, G: 189, B: 189, A: 255} // #bdbdbd
	ColorGrayDark     = color.RGBA{R: 119, G: 119, B: 119, A: 255} // #777777

	// Base colors (matching --color-white/black)
	ColorWhite   = color.RGBA{R: 255, G: 255, B: 255, A: 255} // #ffffff
	ColorBlack   = color.RGBA{R: 34, G: 34, B: 34, A: 255}    // #222222
	ColorOverlay = color.RGBA{R: 0, G: 0, B: 0, A: 64}        // rgba(0, 0, 0, 0.25)

	// Semantic color aliases for text
	ColorTextPrimary   = ColorBlack    // Primary text color
	ColorTextSecondary = ColorGrayDark // Secondary/muted text color
)

// Button styles matching nishiki-frontend Button component variants exactly
// Base: 'text-base min-w-[70px] rounded inline-flex items-center justify-center gap-2.5'
func StyleButtonBase(s *styles.Style) {
	s.Font.Size = units.Dp(16)                      // text-base
	s.Min.X.Set(70, units.UnitDp)                   // min-w-[70px]
	s.Border.Radius = sides.NewValues(units.Dp(10)) // rounded
	s.Display = styles.Flex                         // inline-flex
	s.Align.Items = styles.Center                   // items-center
	s.Justify.Content = styles.Center               // justify-center
	s.Gap.Set(units.Dp(10))                         // gap-2.5 (10px)
	s.Cursor = cursors.Pointer
}

// Legacy method wrappers removed - use StyleButton* functions directly

// Updated Button Variants (from nishiki-frontend Button.tsx)
// variant: 'primary': 'bg-primary text-white enabled:hover:bg-primary-dark disabled:opacity-50'
func StyleButtonPrimary(s *styles.Style) {
	StyleButtonBase(s)
	s.Background = colors.Uniform(ColorPrimary) // bg-primary
	s.Color = colors.Uniform(ColorWhite)        // text-white
}

// variant: 'danger': 'bg-danger text-white enabled:hover:bg-danger-dark disabled:opacity-50'
func StyleButtonDanger(s *styles.Style) {
	StyleButtonBase(s)
	s.Background = colors.Uniform(ColorDanger) // bg-danger
	s.Color = colors.Uniform(ColorWhite)       // text-white
}

// variant: 'cancel': 'bg-transparent text-black hover:bg-gray-light'
func StyleButtonCancel(s *styles.Style) {
	StyleButtonBase(s)
	s.Background = colors.Uniform(color.RGBA{R: 0, G: 0, B: 0, A: 0}) // bg-transparent
	s.Color = colors.Uniform(ColorBlack)                              // text-black
}

// Not in nishiki-frontend but used in Go frontend
func StyleButtonAccent(s *styles.Style) {
	StyleButtonBase(s)
	s.Background = colors.Uniform(ColorAccent) // bg-accent
	s.Color = colors.Uniform(ColorBlack)       // text-black
}

// variant: 'ghost': 'bg-transparent'
func StyleButtonGhost(s *styles.Style) {
	StyleButtonBase(s)
	s.Background = colors.Uniform(color.RGBA{R: 0, G: 0, B: 0, A: 0}) // bg-transparent
}

// Button Sizes
// size: 'sm': 'h-8 px-7' (32px height, 28px horizontal padding)
func StyleButtonSm(s *styles.Style) {
	s.Min.Y.Set(32, units.UnitDp)
	s.Padding.Set(units.Dp(0), units.Dp(28))
}

// size: 'md': 'h-10 px-12' (40px height, 48px horizontal padding)
func StyleButtonMd(s *styles.Style) {
	s.Min.Y.Set(40, units.UnitDp)
	s.Padding.Set(units.Dp(0), units.Dp(48))
}

// size: 'lg': 'h-12 px-12' (48px height, 48px horizontal padding)
func StyleButtonLg(s *styles.Style) {
	s.Min.Y.Set(48, units.UnitDp)
	s.Padding.Set(units.Dp(0), units.Dp(48))
}

// size: 'icon': 'h-12 w-12 min-w-0' (48px square, no min width)
func StyleButtonIcon(s *styles.Style) {
	StyleButtonBase(s)
	s.Min.Y.Set(48, units.UnitDp)
	s.Min.X.Set(48, units.UnitDp)
	s.Background = colors.Uniform(ColorGrayLight)
}

// Card styles matching nishiki-frontend Card component exactly (bg-white rounded)
func StyleCard(s *styles.Style) {
	s.Background = colors.Uniform(ColorWhite)       // bg-white
	s.Border.Radius = sides.NewValues(units.Dp(10)) // rounded (DEFAULT = 0.625rem = 10px from tailwind.config.ts)
	s.Margin.Bottom = units.Dp(8)                   // mb-2 (matching React FoodCard and GroupCard spacing)
}

// Card layout patterns from nishiki-frontend Tailwind classes
// Pattern: Card + className="flex justify-between gap-2"
func StyleCardFlexBetween(s *styles.Style) {
	StyleCard(s) // Apply card base styles
	s.Display = styles.Flex // Ensure flex display
	s.Direction = styles.Row
	s.Justify.Content = styles.SpaceBetween
	s.Gap.Set(units.Dp(8)) // gap-2
}

// Pattern: className="flex grow gap-4 items-center pl-4 py-2" (member/container card content)
func StyleCardContentGrow(s *styles.Style) {
	s.Display = styles.Flex                                            // flex
	s.Direction = styles.Row
	s.Grow.Set(1, 0)                                                   // grow
	s.Gap.Set(units.Dp(16))                                            // gap-4
	s.Align.Items = styles.Center                                      // items-center
	s.Padding.Set(units.Dp(8), units.Dp(0), units.Dp(8), units.Dp(16)) // pl-4 py-2 (16px left, 8px top/bottom)
	s.Cursor = cursors.Pointer                                         // Make clickable like React Link
}

// Pattern: className="flex grow flex-col gap-3 pl-4 py-2" (group card content)
func StyleCardContentColumn(s *styles.Style) {
	s.Display = styles.Flex                                            // flex
	s.Direction = styles.Column
	s.Grow.Set(1, 0)                                                   // grow
	s.Gap.Set(units.Dp(12))                                            // gap-3
	s.Padding.Set(units.Dp(8), units.Dp(0), units.Dp(8), units.Dp(16)) // pl-4 py-2 (16px left, 8px top/bottom)
	s.Cursor = cursors.Pointer                                         // Make clickable like React Link
}

// Icon circle pattern: className="flex items-center justify-center bg-accent rounded-full w-11 h-11"
func StyleIconCircleAccent(s *styles.Style) {
	s.Background = colors.Uniform(ColorAccent)        // bg-accent
	s.Border.Radius = sides.NewValues(units.Dp(9999)) // rounded-full
	s.Min.X.Set(44, units.UnitDp)                     // w-11 (44px)
	s.Min.Y.Set(44, units.UnitDp)                     // h-11 (44px)
	s.Display = styles.Flex                           // flex
	s.Align.Items = styles.Center                     // items-center
	s.Justify.Content = styles.Center                 // justify-center
}

// Duplicate section removed - definitions already exist above

// Input Variants (from nishiki-frontend Input.tsx)
// Base: 'flex w-full text-base focus:outline-none disabled:cursor-not-allowed'
func StyleInputBase(s *styles.Style) {
	s.Display = styles.Flex        // flex
	s.Min.X.Set(100, units.UnitEw) // w-full
	s.Font.Size = units.Dp(16)     // text-base
	s.Cursor = cursors.Text
}

// variant: 'rounded': 'rounded-full bg-white border border-gray px-6 py-4 placeholder:text-gray focus:ring-2 focus:ring-primary-dark focus:border-transparent'
func StyleInputRounded(s *styles.Style) {
	StyleInputBase(s)
	s.Border.Radius = sides.NewValues(units.Dp(9999)) // rounded-full
	s.Background = colors.Uniform(ColorWhite)         // bg-white
	s.Border.Style.Set(styles.BorderSolid)            // border
	s.Border.Width.Set(units.Dp(1))                   // border
	s.Border.Color.Set(colors.Uniform(ColorGray))     // border-gray
	s.Padding.Set(units.Dp(16), units.Dp(24))         // px-6 py-4 (24px, 16px)
}

// SearchInput pattern: relative container with icon
// Icon positioning: className="absolute top-4 left-6"
func StyleSearchIcon(s *styles.Style) {
	// Note: absolute positioning not available in v0.3.12, use layout instead
	s.Color = colors.Uniform(ColorGray)
}

// Input with left padding for icon: className="pl-12"
func StyleSearchInputWithIcon(s *styles.Style) {
	StyleInputRounded(s)
	s.Padding.Left = units.Dp(48) // pl-12
}

// Badge Variants (from nishiki-frontend Badge.tsx)
// Base: 'inline-flex items-center rounded-full text-sm h-6 px-3.5'
func StyleBadgeBase(s *styles.Style) {
	s.Display = styles.Flex                           // inline-flex
	s.Align.Items = styles.Center                     // items-center
	s.Border.Radius = sides.NewValues(units.Dp(9999)) // rounded-full
	s.Font.Size = units.Dp(14)                        // text-sm
	s.Min.Y.Set(24, units.UnitDp)                     // h-6 (24px)
	s.Padding.Set(units.Dp(0), units.Dp(14))          // px-3.5 (14px)
}

// variant: 'light': 'bg-primary-light'
func StyleBadgeLight(s *styles.Style) {
	StyleBadgeBase(s)
	s.Background = colors.Uniform(ColorPrimaryLight) // bg-primary-light
}

// variant: 'lightest': 'bg-primary-lightest'
func StyleBadgeLightest(s *styles.Style) {
	StyleBadgeBase(s)
	s.Background = colors.Uniform(ColorPrimaryLightest) // bg-primary-lightest
}

// variant: 'outline': 'bg-gray-lightest border border-primary text-primary'
func StyleBadgeOutline(s *styles.Style) {
	StyleBadgeBase(s)
	s.Background = colors.Uniform(ColorGrayLightest) // bg-gray-lightest
	s.Border.Style.Set(styles.BorderSolid)           // border
	s.Border.Width.Set(units.Dp(1))                  // border
	s.Border.Color.Set(colors.Uniform(ColorPrimary)) // border-primary
	s.Color = colors.Uniform(ColorPrimary)           // text-primary
}

// Legacy Food Card Pattern removed - use StyleFoodCardContainer instead

// Emoji food circle: className="bg-white w-10 h-10 rounded-full flex items-center justify-center border border-primary select-none text-2xl"
func StyleFoodEmojiCircle(s *styles.Style) {
	s.Background = colors.Uniform(ColorWhite)         // bg-white
	s.Min.X.Set(40, units.UnitDp)                     // w-10
	s.Min.Y.Set(40, units.UnitDp)                     // h-10
	s.Border.Radius = sides.NewValues(units.Dp(9999)) // rounded-full
	s.Display = styles.Flex                           // flex
	s.Align.Items = styles.Center                     // items-center
	s.Justify.Content = styles.Center                 // justify-center
	s.Border.Style.Set(styles.BorderSolid)            // border
	s.Border.Width.Set(units.Dp(1))                   // border
	s.Border.Color.Set(colors.Uniform(ColorPrimary))  // border-primary
	s.Font.Size = units.Dp(24)                        // text-2xl
}

// Text patterns from frontend
// className="text-xs text-gray-dark flex items-center gap-1 my-1.5"
func StyleTextXsGrayWithIcon(s *styles.Style) {
	s.Font.Size = units.Dp(12)              // text-xs
	s.Color = colors.Uniform(ColorGrayDark) // text-gray-dark
	s.Display = styles.Flex                 // flex
	s.Align.Items = styles.Center           // items-center
	s.Gap.Set(units.Dp(4))                  // gap-1
	s.Margin.Set(units.Dp(6), units.Dp(0))  // my-1.5
}

// className="text-sm flex items-center gap-1"
func StyleTextSmWithIcon(s *styles.Style) {
	s.Font.Size = units.Dp(14)    // text-sm
	s.Display = styles.Flex       // flex
	s.Align.Items = styles.Center // items-center
	s.Gap.Set(units.Dp(4))        // gap-1
}

// className="ml-auto" (time display)
func StyleTextAutoRight(s *styles.Style) {
	s.Margin.Left = units.Dp(-1) // ml-auto equivalent (grow to push right)
}

// Grid patterns
// className="grid grid-cols-2 gap-6"
func StyleGrid2Cols(s *styles.Style) {
	s.Display = styles.Grid // grid
	s.Gap.Set(units.Dp(24)) // gap-6
	// Note: grid-cols-2 would need CSS Grid implementation in Cogent Core
}

// Form patterns
// className="flex flex-col gap-4" (common drawer body)
func StyleDrawerBody(s *styles.Style) {
	s.Display = styles.Flex     // flex
	s.Direction = styles.Column // flex-col
	s.Gap.Set(units.Dp(16))     // gap-4
}

// className="flex flex-wrap gap-1.5 whitespace-nowrap"
func StyleFlexWrap(s *styles.Style) {
	s.Display = styles.Flex // flex
	s.Wrap = true           // flex-wrap
	s.Gap.Set(units.Dp(6))  // gap-1.5
	// Note: whitespace-nowrap would need text wrapping control
}

// Filter button pattern with dot
// className="absolute -top-[3px] -right-[5px] w-2 h-2 rounded-full bg-danger"
func StyleFilterDot(s *styles.Style) {
	s.Min.X.Set(8, units.UnitDp)                      // w-2
	s.Min.Y.Set(8, units.UnitDp)                      // h-2
	s.Border.Radius = sides.NewValues(units.Dp(9999)) // rounded-full
	s.Background = colors.Uniform(ColorDanger)        // bg-danger
	// Note: absolute positioning (-top-[3px] -right-[5px]) would need positioning system
}

// Category selection patterns
// className="w-6 aspect-square rounded-full border border-primary flex items-center justify-center"
func StyleCategoryIcon(s *styles.Style) {
	s.Min.X.Set(24, units.UnitDp)                     // w-6
	s.Min.Y.Set(24, units.UnitDp)                     // aspect-square (24px)
	s.Border.Radius = sides.NewValues(units.Dp(9999)) // rounded-full
	s.Border.Style.Set(styles.BorderSolid)            // border
	s.Border.Width.Set(units.Dp(1))                   // border
	s.Border.Color.Set(colors.Uniform(ColorPrimary))  // border-primary
	s.Display = styles.Flex                           // flex
	s.Align.Items = styles.Center                     // items-center
	s.Justify.Content = styles.Center                 // justify-center
}

// className="bg-white w-8 h-8 rounded-full flex items-center justify-center border border-primary select-none"
func StyleCategoryIconLarge(s *styles.Style) {
	s.Background = colors.Uniform(ColorWhite)         // bg-white
	s.Min.X.Set(32, units.UnitDp)                     // w-8
	s.Min.Y.Set(32, units.UnitDp)                     // h-8
	s.Border.Radius = sides.NewValues(units.Dp(9999)) // rounded-full
	s.Display = styles.Flex                           // flex
	s.Align.Items = styles.Center                     // items-center
	s.Justify.Content = styles.Center                 // justify-center
	s.Border.Style.Set(styles.BorderSolid)            // border
	s.Border.Width.Set(units.Dp(1))                   // border
	s.Border.Color.Set(colors.Uniform(ColorPrimary))  // border-primary
}

// Old style methods removed - use StyleX functions directly:
// StyleMainContainer, StyleMainHeader, StyleNavHeader, StyleIconCircleAccent,
// StyleFoodCardContainer, StyleFoodCardEmojiCircle, StyleH1/H2/H3, etc.

// Old IconSize type removed - use StyleIconSizeX and StyleIconColor functions instead

// Old mobile layout creation methods removed - use StyleMainContainer, StyleMainHeader, etc.

// Food categories matching React frontend exactly
var FoodCategories = map[string]FoodCategory{
	"unselected":       {Name: "Unselected", Emoji: "🥣"},
	"beverage":         {Name: "Beverage", Emoji: "☕️"},
	"dairy":            {Name: "Dairy", Emoji: "🥛"},
	"eggs":             {Name: "Egg", Emoji: "🥚"},
	"fatsAndOils":      {Name: "Fat & Oil", Emoji: "🫒"},
	"fruits":           {Name: "Fruit", Emoji: "🍎"},
	"vegetables":       {Name: "Vegetable", Emoji: "🥗"},
	"legumes":          {Name: "Legume", Emoji: "🫘"},
	"nutsAndSeeds":     {Name: "Nut & Seed", Emoji: "🥜"},
	"meat":             {Name: "Meat", Emoji: "🥩"},
	"desserts":         {Name: "Dessert", Emoji: "🍰"},
	"soup":             {Name: "Soup", Emoji: "🍜"},
	"seafoods":         {Name: "Seafood", Emoji: "🍣"},
	"convenienceMeals": {Name: "Convenience Meal", Emoji: "🥡"},
	"seasoning":        {Name: "Seasoning", Emoji: "🧂"},
	"alcohol":          {Name: "Alcohol", Emoji: "🍺"},
	"other":            {Name: "Other", Emoji: "🍽️"},
}

type FoodCategory struct {
	Name  string
	Emoji string
}

// Common Styler functions to centralize styling logic

// Layout stylers matching nishiki-frontend mobile-first design
func StyleContentColumn(s *styles.Style) {
	s.Direction = styles.Column
	s.Grow.Set(1, 1)
	s.Padding.Set(units.Dp(24), units.Dp(16), units.Dp(8), units.Dp(16)) // pt-6 px-4 pb-2 (matching React GroupsPage exactly)
	s.Gap.Set(units.Dp(8))                                               // gap-2 (matching frontend)
}

func StyleMainContainer(s *styles.Style) {
	s.Direction = styles.Column
	s.Grow.Set(1, 1)
	s.Padding.Set(units.Dp(48), units.Dp(0), units.Dp(64), units.Dp(0)) // pt-12 pb-16 (mobile layout)
	s.Min.Y.Set(100, units.UnitVh)                                      // min-h-screen
	s.Background = colors.Uniform(ColorGrayLightest)                     // bg-gray-lightest (matching React app background)
}

func StyleCenteredContainer(s *styles.Style) {
	s.Direction = styles.Column
	s.Align.Items = styles.Center
	s.Justify.Content = styles.Center
	s.Grow.Set(1, 1)
	s.Gap.Set(units.Dp(32))
	s.Padding.Set(units.Dp(24))
}

func StyleHeaderRow(s *styles.Style) {
	s.Direction = styles.Row
	s.Align.Items = styles.Center
	s.Justify.Content = styles.SpaceBetween // SpaceBetween for left/center/right elements (matching React Header)
	s.Background = colors.Uniform(ColorWhite)
	s.Min.Y.Set(48, units.UnitDp)  // h-12 (48px)
	s.Min.X.Set(100, units.UnitVw) // w-full
	s.Padding.Set(units.Dp(0), units.Dp(16)) // px-4 for proper edge spacing
	// Note: Position styling not available in v0.3.12, will use normal flow
}

func StyleActionsRow(s *styles.Style) {
	s.Direction = styles.Row
	s.Gap.Set(units.Dp(12))
	s.Align.Items = styles.Center
	s.Margin.Bottom = units.Dp(16)
}

// ActionBar - matches React's h-12 w-full flex items-center justify-end pattern  
func StyleActionBar(s *styles.Style) {
	s.Display = styles.Flex
	s.Min.Y.Set(48, units.UnitDp)    // h-12 (48px)
	s.Min.X.Set(100, units.UnitEw)   // w-full
	s.Align.Items = styles.Center    // items-center
	s.Justify.Content = styles.End   // justify-end (matching React GroupsPage)
}

func StyleGridContainer(s *styles.Style) {
	s.Display = styles.Grid
	s.Gap.Set(units.Dp(16))
	s.Wrap = true
}

// Text stylers matching Tailwind typography scale
func StyleAppTitle(s *styles.Style) {
	s.Font.Size = units.Dp(24) // text-2xl
	s.Font.Weight = WeightBold
	s.Color = colors.Uniform(ColorPrimary)
}

func StyleSubtitle(s *styles.Style) {
	s.Font.Size = units.Dp(16) // text-base
	s.Color = colors.Uniform(ColorGrayDark)
}

func StyleSectionTitle(s *styles.Style) {
	s.Font.Size = units.Dp(20) // text-xl (matching H2 component)
	s.Font.Weight = WeightSemiBold
	s.Color = colors.Uniform(ColorBlack)
}

func StyleCardTitle(s *styles.Style) {
	s.Font.Size = units.Dp(18) // text-lg (matching group card titles)
	s.Font.Weight = WeightSemiBold
	s.Margin.Bottom = units.Dp(2) // Simulate React's "leading-6" line height spacing
}

func StyleSmallText(s *styles.Style) {
	s.Font.Size = units.Dp(12) // text-xs
	s.Color = colors.Uniform(ColorGrayDark)
}

// Background stylers
func StyleMainBackground(s *styles.Style) {
	s.Direction = styles.Column
	s.Background = colors.Uniform(ColorGrayLightest) // #f8f8f8 (matching frontend)
}

// Navigation and interactive element stylers
func StyleNavButton(s *styles.Style) {
	s.Background = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(10)) // rounded
	s.Padding.Set(units.Dp(16))
	s.Gap.Set(units.Dp(8))
	s.Min.X.Set(120, units.UnitDp)
	s.Border.Style.Set(styles.BorderSolid)
	s.Border.Width.Set(units.Dp(1))
	s.Border.Color.Set(colors.Uniform(ColorGrayLight))
}

func StyleUserButton(s *styles.Style) {
	s.Background = colors.Uniform(color.RGBA{R: 240, G: 240, B: 240, A: 255})
	s.Border.Radius = sides.NewValues(units.Dp(10)) // rounded
	s.Padding.Set(units.Dp(8), units.Dp(16))
}

func StyleBackButton(s *styles.Style) {
	s.Background = colors.Uniform(color.RGBA{R: 240, G: 240, B: 240, A: 255})
	s.Border.Radius = sides.NewValues(units.Dp(9999)) // rounded-full
	s.Padding.Set(units.Dp(8))
}

// Card and container stylers
func StyleStatCard(cardColor color.RGBA) func(*styles.Style) {
	return func(s *styles.Style) {
		s.Direction = styles.Column
		s.Align.Items = styles.Center
		s.Background = colors.Uniform(cardColor)
		s.Border.Radius = sides.NewValues(units.Dp(10)) // rounded
		s.Padding.Set(units.Dp(16))
		s.Gap.Set(units.Dp(4))
		s.Min.X.Set(100, units.UnitDp)
	}
}

func StyleCardInfo(s *styles.Style) {
	s.Direction = styles.Column
	s.Gap.Set(units.Dp(12)) // gap-3 (matching frontend GroupCard flex-col gap-3)
	s.Grow.Set(1, 0)        // grow
}

func StyleStatsContainer(s *styles.Style) {
	s.Direction = styles.Column
	s.Gap.Set(units.Dp(12))
	s.Background = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(10)) // rounded
	s.Padding.Set(units.Dp(16))
}

func StyleStatsGrid(s *styles.Style) {
	s.Direction = styles.Row
	s.Gap.Set(units.Dp(16))
	s.Wrap = true
}

func StyleNavContainer(s *styles.Style) {
	s.Direction = styles.Row
	s.Gap.Set(units.Dp(12))
	s.Wrap = true
	s.Max.X.Set(512, units.UnitDp)  // max-w-lg constraint (matching React BottomMenu)
	s.Margin.Left = units.Dp(-1)    // mx-auto (center horizontally)
	s.Margin.Right = units.Dp(-1)
}

func StyleDevSection(s *styles.Style) {
	s.Direction = styles.Column
	s.Background = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(10)) // rounded
	s.Padding.Set(units.Dp(16))
	s.Gap.Set(units.Dp(12))
	s.Margin.Top = units.Dp(16)
}

// Text styling for specific contexts
func StyleStatValue(s *styles.Style) {
	s.Font.Size = units.Dp(24) // text-2xl
	s.Font.Weight = WeightBold
	s.Color = colors.Uniform(ColorWhite)
}

func StyleStatLabel(s *styles.Style) {
	s.Font.Size = units.Dp(14) // text-sm
	s.Color = colors.Uniform(ColorWhite)
}

func StyleStatsTitle(s *styles.Style) {
	s.Font.Size = units.Dp(18) // text-lg
	s.Font.Weight = WeightSemiBold
}

func StyleDevTitle(s *styles.Style) {
	s.Font.Size = units.Dp(16) // text-base
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
	s.Margin.Top = units.Dp(32)
}

func StyleLogoutButton(s *styles.Style) {
	s.Align.Self = styles.Start
	s.Margin.Top = units.Dp(16)
}

func StyleCreateButton(s *styles.Style) {
	s.Align.Self = styles.End
}

func StyleViewButton(s *styles.Style) {
	s.Background = colors.Uniform(ColorPrimary)
	s.Color = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(10)) // rounded
	s.Padding.Set(units.Dp(8), units.Dp(12))
	s.Gap.Set(units.Dp(4))
}

func StyleViewButtonAccent(s *styles.Style) {
	s.Background = colors.Uniform(ColorAccent)
	s.Color = colors.Uniform(ColorBlack)
	s.Border.Radius = sides.NewValues(units.Dp(10)) // rounded
	s.Padding.Set(units.Dp(8), units.Dp(12))
	s.Gap.Set(units.Dp(4))
}

func StyleClearCacheButton(s *styles.Style) {
	s.Background = colors.Uniform(ColorAccent)
	s.Color = colors.Uniform(ColorBlack)
	s.Border.Radius = sides.NewValues(units.Dp(10)) // rounded
	s.Padding.Set(units.Dp(8), units.Dp(12))
	s.Gap.Set(units.Dp(4))
	s.Align.Self = styles.Start
}

// Header-specific styles
func StyleHeaderLeftContainer(s *styles.Style) {
	s.Direction = styles.Row
	s.Align.Items = styles.Center
	s.Gap.Set(units.Dp(12))
}

// Description text styling (matching frontend text-sm text-gray-dark)
func StyleDescriptionText(s *styles.Style) {
	s.Font.Size = units.Dp(14) // text-sm
	s.Color = colors.Uniform(ColorGrayDark)
}

// ==================== COMPREHENSIVE STYLE FUNCTIONS ====================
// Added to eliminate all inline styles and ensure consistency

// Dialog and Modal Styles
func StyleDialogContainer(s *styles.Style) {
	s.Background = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(16)) // BorderRadiusLarge
	s.Padding.Set(units.Dp(24))
	s.Gap.Set(units.Dp(16))
	s.Direction = styles.Column
	s.Min.X.Set(400, units.UnitDp)
	s.Max.X.Set(500, units.UnitDp)
}

func StyleDialogTitle(s *styles.Style) {
	s.Font.Size = units.Dp(20) // text-xl
	s.Font.Weight = WeightSemiBold
}

func StyleDialogButtonRow(s *styles.Style) {
	s.Direction = styles.Row
	s.Gap.Set(units.Dp(12))
	s.Justify.Content = styles.End
}

// Search and Filter Styles
func StyleSearchContainer(s *styles.Style) {
	s.Direction = styles.Row
	s.Gap.Set(units.Dp(12))
	s.Align.Items = styles.Center
	s.Background = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(16)) // BorderRadiusLarge
	s.Padding.Set(units.Dp(16))
}

func StyleSearchField(s *styles.Style) {
	s.Grow.Set(1, 1)
	s.Border.Style.Set(styles.BorderNone)
}

func StyleFilterContainer(s *styles.Style) {
	s.Direction = styles.Column
	s.Background = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(16)) // BorderRadiusLarge
	s.Padding.Set(units.Dp(16))
	s.Gap.Set(units.Dp(16))
}

func StyleFilterRow(s *styles.Style) {
	s.Direction = styles.Row
	s.Wrap = true
	s.Gap.Set(units.Dp(12))
}

func StyleFilterLabel(s *styles.Style) {
	s.Font.Size = units.Dp(12) // text-xs
	s.Font.Weight = WeightSemiBold
	s.Color = colors.Uniform(ColorGrayDark)
}

func StyleDropdownButton(s *styles.Style) {
	s.Background = colors.Uniform(ColorGrayLightest)
	s.Border.Radius = sides.NewValues(units.Dp(8)) // BorderRadiusMedium
	s.Padding.Set(units.Dp(8), units.Dp(12))
	s.Gap.Set(units.Dp(4))
}

func StyleTagBadge(s *styles.Style) {
	s.Font.Size = units.Dp(10) // text-xs
	s.Background = colors.Uniform(ColorPrimaryLightest)
	s.Color = colors.Uniform(ColorPrimary)
	s.Padding.Set(units.Dp(4), units.Dp(8))
	s.Border.Radius = sides.NewValues(units.Dp(9999))
}

func StyleTagBadgeSecondary(s *styles.Style) {
	s.Font.Size = units.Dp(10) // text-xs
	s.Background = colors.Uniform(color.RGBA{R: 240, G: 240, B: 240, A: 255})
	s.Color = colors.Uniform(ColorGrayDark)
	s.Padding.Set(units.Dp(4), units.Dp(8))
	s.Border.Radius = sides.NewValues(units.Dp(9999))
}

// Object and Item Card Styles

func StyleObjectCardHeader(s *styles.Style) {
	s.Direction = styles.Row
	s.Align.Items = styles.Center
	s.Justify.Content = styles.SpaceBetween
}

func StyleObjectCardNameSection(s *styles.Style) {
	s.Direction = styles.Row
	s.Align.Items = styles.Center
	s.Gap.Set(units.Dp(8))
	s.Grow.Set(1, 0)
}

func StyleObjectCardName(s *styles.Style) {
	s.Font.Size = units.Dp(16) // text-base
	s.Font.Weight = WeightSemiBold
	s.Color = colors.Uniform(ColorBlack)
}

func StyleObjectCardDescription(s *styles.Style) {
	s.Font.Size = units.Dp(14) // text-sm
	s.Color = colors.Uniform(ColorGrayDark)
}

func StyleObjectCardActionsMenu(s *styles.Style) {
	s.Direction = styles.Row
	s.Gap.Set(units.Dp(4))
}

// Properties Display Styles
func StylePropertiesContainer(s *styles.Style) {
	s.Direction = styles.Column
	s.Gap.Set(units.Dp(4))
}

func StylePropertiesTitle(s *styles.Style) {
	s.Font.Size = units.Dp(12) // text-xs
	s.Font.Weight = WeightSemiBold
	s.Color = colors.Uniform(ColorGrayDark)
}

func StylePropertyRow(s *styles.Style) {
	s.Direction = styles.Row
	s.Justify.Content = styles.SpaceBetween
}

func StylePropertyKey(s *styles.Style) {
	s.Font.Size = units.Dp(12) // text-xs
	s.Color = colors.Uniform(ColorGrayDark)
}

func StylePropertyValue(s *styles.Style) {
	s.Font.Size = units.Dp(12) // text-xs
	s.Font.Weight = WeightMedium
}

func StylePropertyRowHighlight(s *styles.Style) {
	s.Direction = styles.Row
	s.Justify.Content = styles.SpaceBetween
	s.Padding.Set(units.Dp(8))
	s.Background = colors.Uniform(ColorGrayLightest)
	s.Border.Radius = sides.NewValues(units.Dp(8)) // BorderRadiusMedium
}

// Breadcrumb Navigation Styles
func StyleBreadcrumbContainer(s *styles.Style) {
	s.Background = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(16)) // BorderRadiusLarge
	s.Padding.Set(units.Dp(16))
}

func StyleBreadcrumbRow(s *styles.Style) {
	s.Direction = styles.Row
	s.Align.Items = styles.Center
	s.Gap.Set(units.Dp(8))
}

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

func StyleCollectionCardHeader(s *styles.Style) {
	s.Direction = styles.Row
	s.Align.Items = styles.Center
	s.Justify.Content = styles.SpaceBetween
}

func StyleCollectionTitleSection(s *styles.Style) {
	s.Direction = styles.Row
	s.Align.Items = styles.Center
	s.Gap.Set(units.Dp(8))
	s.Grow.Set(1, 0)
}

func StyleCollectionTitleContainer(s *styles.Style) {
	s.Direction = styles.Column
	s.Gap.Set(units.Dp(2))
}

func StyleCollectionName(s *styles.Style) {
	s.Font.Size = units.Dp(16) // text-base
	s.Font.Weight = WeightSemiBold
	s.Color = colors.Uniform(ColorBlack)
}

func StyleCollectionType(s *styles.Style) {
	s.Font.Size = units.Dp(12) // text-xs
	s.Color = colors.Uniform(ColorGrayDark)
}

func StyleStatsRow(s *styles.Style) {
	s.Direction = styles.Row
	s.Gap.Set(units.Dp(16))
	s.Justify.Content = styles.SpaceBetween
	s.Background = colors.Uniform(ColorGrayLightest)
	s.Border.Radius = sides.NewValues(units.Dp(8)) // BorderRadiusMedium
	s.Padding.Set(units.Dp(12))
}

func StyleStatColumn(s *styles.Style) {
	s.Direction = styles.Column
	s.Align.Items = styles.Center
	s.Gap.Set(units.Dp(2))
}

func StyleStatValuePrimary(s *styles.Style) {
	s.Font.Size = units.Dp(18) // text-lg
	s.Font.Weight = WeightBold
	s.Color = colors.Uniform(ColorPrimary)
}

func StyleStatValueAccent(s *styles.Style) {
	s.Font.Size = units.Dp(18) // text-lg
	s.Font.Weight = WeightBold
	s.Color = colors.Uniform(ColorAccent)
}

// Container Card Styles

func StyleContainerInfoSection(s *styles.Style) {
	s.Direction = styles.Row
	s.Align.Items = styles.Center
	s.Gap.Set(units.Dp(12))
	s.Grow.Set(1, 0)
	s.Cursor = cursors.Pointer
}

func StyleContainerDetails(s *styles.Style) {
	s.Direction = styles.Column
	s.Gap.Set(units.Dp(4))
}

func StyleContainerName(s *styles.Style) {
	s.Font.Size = units.Dp(16) // text-base
	s.Font.Weight = WeightSemiBold
}

func StyleContainerDescription(s *styles.Style) {
	s.Font.Size = units.Dp(14) // text-sm
	s.Color = colors.Uniform(ColorGrayDark)
}

func StyleContainerCount(s *styles.Style) {
	s.Font.Size = units.Dp(12) // text-xs
	s.Color = colors.Uniform(ColorGrayDark)
}

func StyleContainerActionsMenu(s *styles.Style) {
	s.Direction = styles.Row
	s.Gap.Set(units.Dp(8))
}

// Action Button Styles
func StyleActionButtonEdit(s *styles.Style) {
	s.Background = colors.Uniform(ColorAccent)
	s.Color = colors.Uniform(ColorBlack)
	s.Border.Radius = sides.NewValues(units.Dp(9999))
	s.Padding.Set(units.Dp(6))
}

func StyleActionButtonDelete(s *styles.Style) {
	s.Background = colors.Uniform(ColorDanger)
	s.Color = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(9999))
	s.Padding.Set(units.Dp(6))
}

func StyleActionButtonInvite(s *styles.Style) {
	s.Background = colors.Uniform(ColorPrimary)
	s.Color = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(9999))
	s.Padding.Set(units.Dp(8))
}

func StyleActionButtonRemove(s *styles.Style) {
	s.Background = colors.Uniform(ColorDanger)
	s.Color = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(9999))
	s.Padding.Set(units.Dp(6))
}

func StyleActionButtonLarge(s *styles.Style) {
	s.Background = colors.Uniform(ColorPrimary)
	s.Color = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(16)) // BorderRadiusLarge
	s.Padding.Set(units.Dp(12), units.Dp(16))
	s.Gap.Set(units.Dp(8))
}

func StyleActionButtonLargeAccent(s *styles.Style) {
	s.Background = colors.Uniform(ColorAccent)
	s.Color = colors.Uniform(ColorBlack)
	s.Border.Radius = sides.NewValues(units.Dp(16)) // BorderRadiusLarge
	s.Padding.Set(units.Dp(12), units.Dp(16))
	s.Gap.Set(units.Dp(8))
}

func StyleActionButtonLargeDanger(s *styles.Style) {
	s.Background = colors.Uniform(ColorDanger)
	s.Color = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(16)) // BorderRadiusLarge
	s.Padding.Set(units.Dp(12), units.Dp(16))
	s.Gap.Set(units.Dp(8))
}

// Grid and Layout Styles
func StyleObjectsGrid(s *styles.Style) {
	s.Direction = styles.Row
	s.Wrap = true
	s.Gap.Set(units.Dp(16))
}

func StyleCollectionsGrid(s *styles.Style) {
	s.Direction = styles.Row
	s.Wrap = true
	s.Gap.Set(units.Dp(16))
}

func StyleViewModeRow(s *styles.Style) {
	s.Direction = styles.Row
	s.Gap.Set(units.Dp(8))
	s.Justify.Content = styles.End
}

func StyleViewModeButtonActive(s *styles.Style) {
	s.Background = colors.Uniform(ColorPrimary)
	s.Color = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(8)) // BorderRadiusMedium
	s.Padding.Set(units.Dp(8))
}

func StyleViewModeButtonInactive(s *styles.Style) {
	s.Background = colors.Uniform(color.RGBA{R: 240, G: 240, B: 240, A: 255})
	s.Border.Radius = sides.NewValues(units.Dp(8)) // BorderRadiusMedium
	s.Padding.Set(units.Dp(8))
}

// Search Result Styles

func StyleSearchResultContent(s *styles.Style) {
	s.Direction = styles.Column
	s.Gap.Set(units.Dp(4))
	s.Grow.Set(1, 0)
}

func StyleSearchResultTitle(s *styles.Style) {
	s.Font.Weight = WeightSemiBold
}

func StyleSearchResultDescription(s *styles.Style) {
	s.Font.Size = units.Dp(12) // text-xs
	s.Color = colors.Uniform(ColorGrayDark)
}

func StyleSearchResultPath(s *styles.Style) {
	s.Font.Size = units.Dp(10) // text-xs
	s.Color = colors.Uniform(ColorPrimary)
}

func StyleSearchResultAction(s *styles.Style) {
	s.Background = colors.Uniform(ColorPrimary)
	s.Color = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(8)) // BorderRadiusMedium
	s.Padding.Set(units.Dp(6), units.Dp(12))
	s.Gap.Set(units.Dp(4))
}

// Filter Badge and Active Filter Styles
func StyleActiveFiltersContainer(s *styles.Style) {
	s.Direction = styles.Column
	s.Background = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(16)) // BorderRadiusLarge
	s.Padding.Set(units.Dp(16))
	s.Gap.Set(units.Dp(12))
}

func StyleActiveFiltersRow(s *styles.Style) {
	s.Direction = styles.Row
	s.Wrap = true
	s.Gap.Set(units.Dp(8))
}

func StyleActiveFilterBadge(s *styles.Style) {
	s.Direction = styles.Row
	s.Align.Items = styles.Center
	s.Gap.Set(units.Dp(4))
	s.Background = colors.Uniform(ColorPrimaryLightest)
	s.Border.Radius = sides.NewValues(units.Dp(9999))
	s.Padding.Set(units.Dp(6), units.Dp(12))
}

func StyleActiveFilterText(s *styles.Style) {
	s.Font.Size = units.Dp(12) // text-xs
	s.Color = colors.Uniform(ColorPrimary)
}

func StyleActiveFilterRemove(s *styles.Style) {
	s.Background = colors.Uniform(ColorPrimary)
	s.Color = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(9999))
	s.Padding.Set(units.Dp(2))
	s.Font.Size = units.Dp(10) // text-xs
}

func StyleClearFiltersButton(s *styles.Style) {
	s.Background = colors.Uniform(color.RGBA{R: 240, G: 240, B: 240, A: 255})
	s.Border.Radius = sides.NewValues(units.Dp(8)) // BorderRadiusMedium
	s.Padding.Set(units.Dp(8), units.Dp(12))
	s.Align.Self = styles.Start
}

// Object Type Selection and Property Styles
func StyleObjectTypeContainer(s *styles.Style) {
	s.Direction = styles.Row
	s.Wrap = true
	s.Gap.Set(units.Dp(8))
}

func StyleObjectTypeButtonSelected(s *styles.Style) {
	s.Background = colors.Uniform(ColorPrimary)
	s.Color = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(8)) // BorderRadiusMedium
	s.Padding.Set(units.Dp(8), units.Dp(12))
}

func StyleObjectTypeButtonUnselected(s *styles.Style) {
	s.Background = colors.Uniform(color.RGBA{R: 240, G: 240, B: 240, A: 255})
	s.Color = colors.Uniform(ColorBlack)
	s.Border.Radius = sides.NewValues(units.Dp(8)) // BorderRadiusMedium
	s.Padding.Set(units.Dp(8), units.Dp(12))
}

func StylePropertiesFormContainer(s *styles.Style) {
	s.Direction = styles.Column
	s.Gap.Set(units.Dp(8))
	s.Background = colors.Uniform(ColorGrayLightest)
	s.Border.Radius = sides.NewValues(units.Dp(8)) // BorderRadiusMedium
	s.Padding.Set(units.Dp(12))
}

// Title and Text Styles by Context
func StyleFilterTitle(s *styles.Style) {
	s.Font.Size = units.Dp(18) // text-lg
	s.Font.Weight = WeightSemiBold
}

func StyleFilterSubtitle(s *styles.Style) {
	s.Font.Size = units.Dp(14) // text-sm
	s.Font.Weight = WeightSemiBold
}

func StyleSectionSubtitle(s *styles.Style) {
	s.Font.Size = units.Dp(16) // text-base
	s.Font.Weight = WeightSemiBold
}

func StyleObjectTitle(s *styles.Style) {
	s.Font.Size = units.Dp(20) // text-xl
	s.Font.Weight = WeightBold
}

func StyleMoreText(s *styles.Style) {
	s.Font.Size = units.Dp(10) // text-xs
	s.Color = colors.Uniform(ColorGrayDark)
}

// Complex Layout and Grid Styles
func StyleSearchSection(s *styles.Style) {
	s.Direction = styles.Row
	s.Gap.Set(units.Dp(8))
	s.Align.Items = styles.Center
}

func StyleSearchFieldContainer(s *styles.Style) {
	s.Min.X.Set(200, units.UnitDp)
}

func StyleAddSection(s *styles.Style) {
	s.Direction = styles.Row
	s.Gap.Set(units.Dp(8))
}

func StyleActionsSplit(s *styles.Style) {
	s.Direction = styles.Row
	s.Gap.Set(units.Dp(12))
	s.Justify.Content = styles.SpaceBetween
}

// Overlay and Modal Background
func StyleOverlayBackground(s *styles.Style) {
	s.Background = colors.Uniform(ColorOverlay) // Semi-transparent black
	s.Display = styles.Flex
	s.Align.Items = styles.Center
	s.Justify.Content = styles.Center
}

// Advanced Dialog Styles
func StyleAdvancedDialogContainer(s *styles.Style) {
	s.Background = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(16)) // BorderRadiusLarge
	s.Padding.Set(units.Dp(24))
	s.Gap.Set(units.Dp(16))
	s.Direction = styles.Column
	s.Min.X.Set(500, units.UnitDp)
	s.Max.X.Set(600, units.UnitDp)
	s.Max.Y.Set(500, units.UnitDp)
}

func StyleTagsSection(s *styles.Style) {
	s.Direction = styles.Column
	s.Gap.Set(units.Dp(8))
}

func StyleTagsGrid(s *styles.Style) {
	s.Direction = styles.Row
	s.Wrap = true
	s.Gap.Set(units.Dp(8))
}

func StyleTagButton(s *styles.Style) {
	s.Background = colors.Uniform(color.RGBA{R: 240, G: 240, B: 240, A: 255})
	s.Border.Radius = sides.NewValues(units.Dp(9999))
	s.Padding.Set(units.Dp(6), units.Dp(12))
}

func StyleExpiryContainer(s *styles.Style) {
	s.Direction = styles.Row
	s.Gap.Set(units.Dp(8))
	s.Align.Items = styles.Center
}

func StyleExpiryField(s *styles.Style) {
	s.Min.X.Set(60, units.UnitDp)
}

// Object Detail View Styles
func StyleObjectDetailHeader(s *styles.Style) {
	s.Direction = styles.Row
	s.Align.Items = styles.Center
	s.Gap.Set(units.Dp(12))
}

func StyleObjectDetailIcon(s *styles.Style) {
	s.Font.Size = units.Dp(24) // text-2xl for larger icons
}

func StyleObjectDetailTitleContainer(s *styles.Style) {
	s.Direction = styles.Column
	s.Gap.Set(units.Dp(2))
}

// Additional Frontend Patterns from nishiki-frontend

// Drawer/Modal Patterns (from Drawer.tsx)
// DrawerOverlay: className="fixed inset-0 z-50 bg-overlay"
func StyleDrawerOverlay(s *styles.Style) {
	// Note: fixed positioning and z-index would need positioning system
	s.Background = colors.Uniform(ColorOverlay) // bg-overlay
	s.Min.X.Set(100, units.UnitVw)              // inset-0 (full width)
	s.Min.Y.Set(100, units.UnitVh)              // inset-0 (full height)
}

// DrawerContent base: className="fixed z-50 bg-white flex flex-col"
func StyleDrawerContent(s *styles.Style) {
	// Note: fixed positioning and z-index would need positioning system
	s.Background = colors.Uniform(ColorWhite) // bg-white
	s.Display = styles.Flex                   // flex
	s.Direction = styles.Column               // flex-col
}

// Drawer bottom variant: className="inset-x-0 bottom-0 rounded-t max-h-[calc(100vh-2rem)]"
func StyleDrawerBottom(s *styles.Style) {
	StyleDrawerContent(s)
	s.Border.Radius.Top = units.Dp(10) // rounded-t
	s.Max.Y.Set(-32, units.UnitVh)     // max-h-[calc(100vh-2rem)] (100vh - 32px)
}

// Drawer right variant: className="inset-y-0 right-0 h-full w-5/6 max-w-sm"
func StyleDrawerRight(s *styles.Style) {
	StyleDrawerContent(s)
	s.Min.Y.Set(100, units.UnitVh)   // h-full
	s.Min.X.Set(83.33, units.UnitVw) // w-5/6 (83.33%)
	s.Max.X.Set(384, units.UnitDp)   // max-w-sm (384px)
}

// DrawerHeader: className="h-12 shrink-0 px-4 border-b border-gray-light grow-1 relative flex items-center"
func StyleDrawerHeader(s *styles.Style) {
	s.Min.Y.Set(48, units.UnitDp)                      // h-12
	s.Padding.Set(units.Dp(0), units.Dp(16))           // px-4
	s.Border.Style.Bottom = styles.BorderSolid         // border-b
	s.Border.Width.Bottom = units.Dp(1)                // border-b
	s.Border.Color.Set(colors.Uniform(ColorGrayLight)) // border-gray-light
	s.Display = styles.Flex                            // flex
	s.Align.Items = styles.Center                      // items-center
	// Note: shrink-0, grow-1, relative positioning would need layout system
}

// DrawerBody: className="p-4 overflow-y-auto max-h-full"
func StyleDrawerBodyContent(s *styles.Style) {
	s.Padding.Set(units.Dp(16)) // p-4
	// Note: overflow-y-auto and max-h-full would need scrolling system
}

// DrawerFooter: className="mt-auto px-4 pt-2 pb-6 flex justify-end gap-4"
func StyleDrawerFooter(s *styles.Style) {
	s.Padding.Set(units.Dp(8), units.Dp(16), units.Dp(24), units.Dp(16)) // px-4 pt-2 pb-6
	s.Display = styles.Flex                                              // flex
	s.Justify.Content = styles.End                                       // justify-end
	s.Gap.Set(units.Dp(16))                                              // gap-4
	// Note: mt-auto would need margin auto system
}

// DrawerTitle: className="text-2xl"
func StyleDrawerTitle(s *styles.Style) {
	s.Font.Size = units.Dp(24) // text-2xl
}

// DrawerClose button: className="absolute right-0 h-full px-4"
func StyleDrawerCloseButton(s *styles.Style) {
	s.Min.Y.Set(100, units.UnitEh)           // h-full
	s.Padding.Set(units.Dp(0), units.Dp(16)) // px-4
	// Note: absolute right-0 positioning would need positioning system
}

// Advanced Layout Patterns
// Aspect ratio square for buttons and containers
func StyleAspectSquare(s *styles.Style) {
	// For square aspect ratio, set min width equal to min height
	// This assumes height is already set on the element
	s.Min.X.Set(100, units.UnitEh) // aspect-square equivalent
}

// Flex grow utilities
func StyleGrow(s *styles.Style) {
	s.Grow.Set(1, 0) // grow (flex-grow: 1)
}

func StyleGrowFull(s *styles.Style) {
	s.Grow.Set(1, 1) // grow with shrink
}

// Animation and interaction states
// Hover states for interactive elements
func StyleHoverPrimary(s *styles.Style) {
	// Note: Hover states would need hover system implementation
	// This represents hover:bg-primary-dark pattern
}

func StyleHoverDanger(s *styles.Style) {
	// Note: Hover states would need hover system implementation
	// This represents hover:bg-danger-dark pattern
}

func StyleHoverGrayLight(s *styles.Style) {
	// Note: Hover states would need hover system implementation
	// This represents hover:bg-gray-light pattern
}

// Navigation patterns
// Header with actions: className="h-16 px-4 border-b border-gray-light flex items-center justify-between"
func StyleNavHeader(s *styles.Style) {
	s.Min.Y.Set(64, units.UnitDp)                      // h-16
	s.Padding.Set(units.Dp(0), units.Dp(16))           // px-4
	s.Border.Style.Bottom = styles.BorderSolid         // border-b
	s.Border.Width.Bottom = units.Dp(1)                // border-b
	s.Border.Color.Set(colors.Uniform(ColorGrayLight)) // border-gray-light
	s.Display = styles.Flex                            // flex
	s.Align.Items = styles.Center                      // items-center
	s.Justify.Content = styles.SpaceBetween            // justify-between
}

// Navigation button grids: className="grid grid-cols-2 gap-4 w-full"
func StyleNavGrid(s *styles.Style) {
	s.Display = styles.Grid        // grid
	s.Gap.Set(units.Dp(16))        // gap-4
	s.Min.X.Set(100, units.UnitEw) // w-full
	// Note: grid-cols-2 would need CSS Grid implementation
}

// Search container patterns
// Search bar with margin: className="mb-2"
func StyleSearchBar(s *styles.Style) {
	s.Margin.Bottom = units.Dp(8) // mb-2
}

// Empty state patterns
// className="text-center p-8 text-gray-dark"
func StyleEmptyState(s *styles.Style) {
	s.Text.Align = text.Center              // text-center
	s.Padding.Set(units.Dp(32))             // p-8
	s.Color = colors.Uniform(ColorGrayDark) // text-gray-dark
}

// Icon System (from nishiki-frontend Icon.tsx)
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

// FilterBadge patterns (from FilterBadge.tsx)
// Badge with custom padding: className="pl-1 pr-0 gap-0"
func StyleFilterBadge(s *styles.Style) {
	StyleBadgeBase(s)
	s.Padding.Left = units.Dp(4)  // pl-1
	s.Padding.Right = units.Dp(0) // pr-0
	s.Gap.Set(units.Dp(0))        // gap-0
}

// FilterBadge icon circle: className="bg-white w-4 h-4 rounded-full p-[3.5px] mr-1 flex items-center justify-center"
func StyleFilterBadgeIconCircle(s *styles.Style) {
	s.Background = colors.Uniform(ColorWhite)         // bg-white
	s.Min.X.Set(16, units.UnitDp)                     // w-4
	s.Min.Y.Set(16, units.UnitDp)                     // h-4
	s.Border.Radius = sides.NewValues(units.Dp(9999)) // rounded-full
	s.Padding.Set(units.Dp(3.5))                      // p-[3.5px]
	s.Margin.Right = units.Dp(4)                      // mr-1
	s.Display = styles.Flex                           // flex
	s.Align.Items = styles.Center                     // items-center
	s.Justify.Content = styles.Center                 // justify-center
}

// FilterBadge emoji circle: className="bg-white w-4 h-4 rounded-full p-[3px] mr-1 flex items-center justify-center text-2xs select-none"
func StyleFilterBadgeEmojiCircle(s *styles.Style) {
	s.Background = colors.Uniform(ColorWhite)         // bg-white
	s.Min.X.Set(16, units.UnitDp)                     // w-4
	s.Min.Y.Set(16, units.UnitDp)                     // h-4
	s.Border.Radius = sides.NewValues(units.Dp(9999)) // rounded-full
	s.Padding.Set(units.Dp(3))                        // p-[3px]
	s.Margin.Right = units.Dp(4)                      // mr-1
	s.Display = styles.Flex                           // flex
	s.Align.Items = styles.Center                     // items-center
	s.Justify.Content = styles.Center                 // justify-center
	s.Font.Size = units.Dp(10)                        // text-2xs (10px)
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

// Header System (from Header.tsx)
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
	s.Min.Y.Set(100, units.UnitEh) // h-full
	s.Min.X.Set(48, units.UnitDp)  // aspect-square (48px for h-12)
	s.Padding.Left = units.Dp(16)  // pl-4
	s.Display = styles.Flex        // flex
	s.Align.Items = styles.Center  // items-center
}

// HeaderMenuCircleButton: className="h-full px-4 flex items-center"
func StyleHeaderMenuButton(s *styles.Style) {
	s.Min.Y.Set(100, units.UnitEh)           // h-full
	s.Padding.Set(units.Dp(0), units.Dp(16)) // px-4
	s.Display = styles.Flex                  // flex
	s.Align.Items = styles.Center            // items-center
}

// Typography System (from Typography components)
// H1: className="text-2xl"
func StyleH1(s *styles.Style) {
	s.Font.Size = units.Dp(24) // text-2xl
}

// H2: className="text-xl"
func StyleH2(s *styles.Style) {
	s.Font.Size = units.Dp(20) // text-xl
}

// H3: className="text-lg"
func StyleH3(s *styles.Style) {
	s.Font.Size = units.Dp(18) // text-lg
}

// Loading Patterns (from nishiki-frontend page.tsx and LoginPage.tsx)
// className="min-h-screen flex items-center justify-center"
func StyleLoadingScreen(s *styles.Style) {
	s.Min.Y.Set(100, units.UnitVh)    // min-h-screen
	s.Display = styles.Flex           // flex
	s.Align.Items = styles.Center     // items-center
	s.Justify.Content = styles.Center // justify-center
}

// className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"
func StyleLoadingSpinner(s *styles.Style) {
	s.Border.Radius = sides.NewValues(units.Dp(9999))     // rounded-full
	s.Min.Y.Set(48, units.UnitDp)                         // h-12
	s.Min.X.Set(48, units.UnitDp)                         // w-12
	s.Border.Style.Set(styles.BorderSolid)                // border
	s.Border.Width.Bottom = units.Dp(2)                   // border-b-2
	s.Border.Color.Set(colors.Uniform(ColorPrimary))      // border-primary (using primary instead of blue-600)
	s.Margin.Set(units.Dp(0), units.Dp(-1), units.Dp(16), units.Dp(-1)) // mx-auto mb-4
}

// LoadingSkeleton - matches React's loading card skeletons
func StyleLoadingSkeleton(s *styles.Style) {
	s.Background = colors.Uniform(color.RGBA{R: 229, G: 231, B: 235, A: 255}) // bg-gray-200
	s.Min.Y.Set(80, units.UnitDp)                         // h-20 (matching React skeleton height)
	s.Border.Radius = sides.NewValues(units.Dp(8))        // rounded-lg
	s.Margin.Bottom = units.Dp(8)                         // mb-2 (matching card spacing)
}

// LoginPage layout patterns
// className="flex items-center justify-center h-screen"
func StyleLoginContainer(s *styles.Style) {
	s.Display = styles.Flex           // flex
	s.Align.Items = styles.Center     // items-center
	s.Justify.Content = styles.Center // justify-center
	s.Min.Y.Set(100, units.UnitVh)    // h-screen
}

// className="transform flex flex-col items-center justify-center"
func StyleLoginContent(s *styles.Style) {
	s.Display = styles.Flex           // flex
	s.Direction = styles.Column       // flex-col
	s.Align.Items = styles.Center     // items-center
	s.Justify.Content = styles.Center // justify-center
}

// Logo patterns: className="w-32 h-26 mb-20"
func StyleLoginLogo(s *styles.Style) {
	s.Min.X.Set(128, units.UnitDp)      // w-32
	s.Min.Y.Set(104, units.UnitDp)      // h-26 (104px)
	s.Margin.Bottom = units.Dp(80)      // mb-20
}

// FoodCard System (from FoodCard.tsx)
// Complete FoodCard pattern: Card className="mb-2 w-full flex"
func StyleFoodCardContainer(s *styles.Style) {
	StyleCard(s)                   // Apply base card styles
	s.Margin.Bottom = units.Dp(8)  // mb-2
	s.Min.X.Set(100, units.UnitEw) // w-full
	s.Display = styles.Flex        // flex
}

// FoodCard button: className="flex grow gap-4 items-center text-left pl-4 py-2"
func StyleFoodCardButton(s *styles.Style) {
	s.Display = styles.Flex                                            // flex
	s.Grow.Set(1, 0)                                                   // grow
	s.Gap.Set(units.Dp(16))                                            // gap-4
	s.Align.Items = styles.Center                                      // items-center
	s.Text.Align = AlignStart                                          // text-left
	s.Padding.Set(units.Dp(8), units.Dp(0), units.Dp(8), units.Dp(16)) // pl-4 py-2
	s.Background = colors.Uniform(color.RGBA{R: 0, G: 0, B: 0, A: 0})  // transparent button
	s.Border.Style.Set(styles.BorderNone)                              // no border
	s.Cursor = cursors.Pointer
}

// FoodCard emoji figure: className="bg-white w-10 h-10 rounded-full flex items-center justify-center border border-primary select-none text-2xl"
func StyleFoodCardEmojiCircle(s *styles.Style) {
	s.Background = colors.Uniform(ColorWhite)         // bg-white
	s.Min.X.Set(40, units.UnitDp)                     // w-10
	s.Min.Y.Set(40, units.UnitDp)                     // h-10
	s.Border.Radius = sides.NewValues(units.Dp(9999)) // rounded-full
	s.Display = styles.Flex                           // flex
	s.Align.Items = styles.Center                     // items-center
	s.Justify.Content = styles.Center                 // justify-center
	s.Border.Style.Set(styles.BorderSolid)            // border
	s.Border.Width.Set(units.Dp(1))                   // border
	s.Border.Color.Set(colors.Uniform(ColorPrimary))  // border-primary
	s.Font.Size = units.Dp(24)                        // text-2xl
	// Note: select-none would need text selection control
}

// FoodCard content grow area: className="grow"
func StyleFoodCardContent(s *styles.Style) {
	s.Grow.Set(1, 0) // grow
}

// FoodCard container info: className="text-xs text-gray-dark flex items-center gap-1 my-1.5"
func StyleFoodCardContainerInfo(s *styles.Style) {
	s.Font.Size = units.Dp(12)              // text-xs
	s.Color = colors.Uniform(ColorGrayDark) // text-gray-dark
	s.Display = styles.Flex                 // flex
	s.Align.Items = styles.Center           // items-center
	s.Gap.Set(units.Dp(4))                  // gap-1
	s.Margin.Set(units.Dp(6), units.Dp(0))  // my-1.5
}

// FoodCard quantity info: className="text-sm flex items-center gap-1"
func StyleFoodCardQuantityInfo(s *styles.Style) {
	s.Font.Size = units.Dp(14)    // text-sm
	s.Display = styles.Flex       // flex
	s.Align.Items = styles.Center // items-center
	s.Gap.Set(units.Dp(4))        // gap-1
}

// FoodCard time display: className="ml-auto"
func StyleFoodCardTime(s *styles.Style) {
	// Note: ml-auto equivalent - push to right via parent layout
	s.Margin.Left = units.Dp(-1) // ml-auto
}

// DropdownMenu System (from DropdownMenu.tsx)
// DropdownMenu trigger button: Button variant="ghost" className="w-12"
func StyleDropdownMenuTrigger(s *styles.Style) {
	StyleButtonGhost(s)           // Apply ghost button variant
	s.Min.X.Set(48, units.UnitDp) // w-12
}

// DropdownMenuContent: className="z-50 min-w-64 overflow-hidden rounded bg-white text-black shadow-around"
func StyleDropdownMenuContent(s *styles.Style) {
	// Note: z-50 would need z-index system
	s.Min.X.Set(256, units.UnitDp)                  // min-w-64 (256px)
	s.Border.Radius = sides.NewValues(units.Dp(10)) // rounded
	s.Background = colors.Uniform(ColorWhite)       // bg-white
	s.Color = colors.Uniform(ColorBlack)            // text-black
	// Note: overflow-hidden and shadow-around would need additional systems
}

// DropdownMenuItem with proper styling
func StyleDropdownMenuItem(s *styles.Style) {
	s.Padding.Set(units.Dp(8), units.Dp(12)) // Standard menu item padding
	s.Cursor = cursors.Pointer
	s.Min.X.Set(100, units.UnitEw) // w-full for proper clickable area
}

// Quantity and unit display patterns
// Quantity container with icon: flex items-center gap-1
func StyleQuantityContainer(s *styles.Style) {
	s.Display = styles.Flex       // flex
	s.Align.Items = styles.Center // items-center
	s.Gap.Set(units.Dp(4))        // gap-1
}

// Time/date display styling
func StyleTimeDisplay(s *styles.Style) {
	s.Font.Size = units.Dp(14)              // text-sm (matching quantity)
	s.Color = colors.Uniform(ColorGrayDark) // subtle color for dates
}

// Select-none equivalent for emoji and icons
func StyleSelectNone(s *styles.Style) {
	// Note: text selection control would need additional implementation
	// This is a placeholder for the select-none class behavior
}

// Text alignment variants
func StyleTextLeft(s *styles.Style) {
	s.Text.Align = AlignStart // text-left
}

func StyleTextRight(s *styles.Style) {
	s.Text.Align = AlignEnd // text-right
}

func StyleTextCenter(s *styles.Style) {
	s.Text.Align = text.Center // text-center
}

// Margin auto utilities
func StyleMarginLeftAuto(s *styles.Style) {
	s.Margin.Left = units.Dp(-1) // ml-auto equivalent
}

func StyleMarginRightAuto(s *styles.Style) {
	s.Margin.Right = units.Dp(-1) // mr-auto equivalent
}

func StyleMarginAuto(s *styles.Style) {
	s.Margin.Left = units.Dp(-1) // mx-auto equivalent
	s.Margin.Right = units.Dp(-1)
}
