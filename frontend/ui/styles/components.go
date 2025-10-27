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
// COMPONENT STYLES - Matching nishiki-frontend components exactly
// ====================================================================================

// ====================================================================================
// Button Component Styles (from nishiki-frontend/src/components/ui/Button.tsx)
// ====================================================================================

// Base: 'text-base min-w-[70px] rounded inline-flex items-center justify-center gap-2.5'
func StyleButtonBase(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeBase)             // text-base
	// Removed Min.X constraint - let buttons size naturally based on content
	s.Border.Radius = sides.NewValues(units.Dp(RadiusDefault)) // rounded
	s.Display = styles.Flex                          // inline-flex
	s.Align.Items = styles.Center                    // items-center
	s.Justify.Content = styles.Center                // justify-center
	s.Gap.Set(units.Dp(Spacing2_5))                  // gap-2.5 (10px)
	s.Cursor = cursors.Pointer
	s.Text.WhiteSpace = text.WrapNever               // Button text should never wrap
}

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
// size: 'sm': 'px-7' (28px horizontal padding) - height determined by padding
func StyleButtonSm(s *styles.Style) {
	s.Padding.Set(units.Dp(Spacing2), units.Dp(28))
}

// size: 'md': 'px-12' (48px horizontal padding) - height determined by padding
func StyleButtonMd(s *styles.Style) {
	s.Padding.Set(units.Dp(10), units.Dp(48)) // 10px = 2.5 spacing
	s.Min.X.Set(100, units.UnitPw) // w-full (parent width) - buttons should be full width in column layouts
}

// size: 'lg': 'px-12' (48px horizontal padding) - height determined by padding
func StyleButtonLg(s *styles.Style) {
	s.Padding.Set(units.Dp(Spacing3), units.Dp(48))
}

// size: 'icon': square button - size determined by padding
func StyleButtonIcon(s *styles.Style) {
	StyleButtonBase(s)
	s.Padding.Set(units.Dp(Spacing3))
	s.Background = colors.Uniform(ColorGrayLight)
}

// Login button - matching React AuthentikLoginButton.tsx blue button
// className="w-full flex items-center justify-center px-4 py-3 border border-transparent rounded-md shadow-sm text-base font-medium text-white bg-blue-600 hover:bg-blue-700"
func StyleButtonLogin(s *styles.Style) {
	// Apply base button styles
	s.Font.Size = units.Dp(FontSizeBase)                  // text-base
	s.Border.Radius = sides.NewValues(units.Dp(RadiusMD)) // rounded-md
	s.Display = styles.Flex                               // flex
	s.Align.Items = styles.Center                         // items-center
	s.Justify.Content = styles.Center                     // justify-center
	s.Gap.Set(units.Dp(Spacing2_5))                       // gap-2.5 (10px)
	s.Cursor = cursors.Pointer

	// Login-specific styles - natural sizing based on content
	s.Padding.Set(units.Dp(12), units.Dp(16))             // py-3 px-4 (12px, 16px)
	s.Background = colors.Uniform(ColorBlue600)           // bg-blue-600
	s.Color = colors.Uniform(ColorWhite)                  // text-white
	s.Font.Weight = WeightMedium                          // font-medium
	// Note: shadow-sm and hover:bg-blue-700 would need additional Cogent Core support
}

// ====================================================================================
// Card Component Styles (from nishiki-frontend/src/components/ui/Card.tsx)
// ====================================================================================

// Card styles matching nishiki-frontend Card component exactly (bg-white rounded)
func StyleCard(s *styles.Style) {
	s.Background = colors.Uniform(ColorWhite)              // bg-white
	s.Border.Radius = sides.NewValues(units.Dp(RadiusDefault)) // rounded (DEFAULT = 0.625rem = 10px)
	s.Margin.Bottom = units.Dp(Spacing2)                   // mb-2 (matching React FoodCard and GroupCard spacing)
	s.Min.X.Set(100, units.UnitPw)                         // w-full (parent width) - CRITICAL: cards must span full width
}

func StyleProfileCard(s *styles.Style) {
	StyleCard(s)                             // Apply base card styles
	s.Direction = styles.Column              // Stack fields vertically
	s.Padding.Set(units.Dp(Spacing4))        // p-4 for spacing inside card
	s.Gap.Set(units.Dp(Spacing2))            // gap-2 between fields
}

// Card layout patterns from nishiki-frontend Tailwind classes
// Pattern: Card + className="flex justify-between gap-2"
func StyleCardFlexBetween(s *styles.Style) {
	StyleCard(s)                             // Apply card base styles
	s.Display = styles.Flex                  // Ensure flex display
	s.Direction = styles.Row
	s.Justify.Content = styles.SpaceBetween
	s.Gap.Set(units.Dp(Spacing2))            // gap-2
}

// Pattern: className="flex grow gap-4 items-center pl-4 py-2" (member/container card content)
func StyleCardContentGrow(s *styles.Style) {
	s.Display = styles.Flex                  // flex
	s.Direction = styles.Row
	s.Grow.Set(1, 0)                         // grow
	s.Gap.Set(units.Dp(Spacing4))            // gap-4
	s.Align.Items = styles.Center            // items-center
	s.Padding.Set(units.Dp(Spacing2), units.Dp(0), units.Dp(Spacing2), units.Dp(Spacing4)) // pl-4 py-2
	s.Cursor = cursors.Pointer               // Make clickable like React Link
}

// Pattern: className="flex grow flex-col gap-3 pl-4 py-2" (group card content)
func StyleCardContentColumn(s *styles.Style) {
	s.Display = styles.Flex                  // flex
	s.Direction = styles.Column
	s.Grow.Set(1, 0)                         // grow
	s.Gap.Set(units.Dp(Spacing3))            // gap-3
	s.Padding.Set(units.Dp(Spacing2), units.Dp(0), units.Dp(Spacing2), units.Dp(Spacing4)) // pl-4 py-2
	s.Cursor = cursors.Pointer               // Make clickable like React Link
}

func StyleCardInfo(s *styles.Style) {
	s.Direction = styles.Column
	s.Gap.Set(units.Dp(Spacing3))            // gap-3 (matching frontend GroupCard flex-col gap-3)
	s.Grow.Set(1, 0)                         // grow
}

// ====================================================================================
// Input Component Styles (from nishiki-frontend/src/components/ui/Input.tsx)
// ====================================================================================

// Base: 'flex w-full text-base focus:outline-none disabled:cursor-not-allowed'
func StyleInputBase(s *styles.Style) {
	s.Display = styles.Flex           // flex
	s.Min.X.Set(100, units.UnitPw)    // w-full (parent width)
	s.Font.Size = units.Dp(FontSizeBase) // text-base
	s.Cursor = cursors.Text
}

// variant: 'rounded': 'rounded-full bg-white border border-gray px-6 py-4 placeholder:text-gray focus:ring-2 focus:ring-primary-dark focus:border-transparent'
func StyleInputRounded(s *styles.Style) {
	StyleInputBase(s)
	s.Border.Radius = sides.NewValues(units.Dp(RadiusFull)) // rounded-full
	s.Background = colors.Uniform(ColorWhite)               // bg-white CRITICAL: White background for inputs
	s.Color = colors.Uniform(ColorBlack)                    // CRITICAL: Black text color for visibility
	s.Border.Style.Set(styles.BorderSolid)                  // border
	s.Border.Width.Set(units.Dp(1))                         // border
	s.Border.Color.Set(colors.Uniform(ColorGray))           // border-gray
	s.Padding.Set(units.Dp(Spacing4), units.Dp(Spacing6))   // px-6 py-4 (24px, 16px)
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

// ====================================================================================
// Badge Component Styles (from nishiki-frontend/src/components/ui/Badge.tsx)
// ====================================================================================

// Base: 'inline-flex items-center rounded-full text-sm h-6 px-3.5'
func StyleBadgeBase(s *styles.Style) {
	s.Display = styles.Flex                           // inline-flex
	s.Align.Items = styles.Center                     // items-center
	s.Border.Radius = sides.NewValues(units.Dp(RadiusFull)) // rounded-full
	s.Font.Size = units.Dp(FontSizeSM)                // text-sm
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
	s.Background = colors.Uniform(ColorGrayLightest)    // bg-gray-lightest
	s.Border.Style.Set(styles.BorderSolid)              // border
	s.Border.Width.Set(units.Dp(1))                     // border
	s.Border.Color.Set(colors.Uniform(ColorPrimary))    // border-primary
	s.Color = colors.Uniform(ColorPrimary)              // text-primary
}

func StyleTagBadge(s *styles.Style) {
	s.Font.Size = units.Dp(FontSize2XS)                 // text-xs
	s.Background = colors.Uniform(ColorPrimaryLightest)
	s.Color = colors.Uniform(ColorPrimary)
	s.Padding.Set(units.Dp(Spacing1), units.Dp(Spacing2)) // py-1 px-2
	s.Border.Radius = sides.NewValues(units.Dp(RadiusFull))
}

func StyleTagBadgeSecondary(s *styles.Style) {
	s.Font.Size = units.Dp(FontSize2XS)                 // text-xs
	s.Background = colors.Uniform(color.RGBA{R: 240, G: 240, B: 240, A: 255})
	s.Color = colors.Uniform(ColorGrayDark)
	s.Padding.Set(units.Dp(Spacing1), units.Dp(Spacing2)) // py-1 px-2
	s.Border.Radius = sides.NewValues(units.Dp(RadiusFull))
}

// ====================================================================================
// Icon Circle Patterns
// ====================================================================================

// Icon circle pattern: className="flex items-center justify-center bg-accent rounded-full w-11 h-11"
func StyleIconCircleAccent(s *styles.Style) {
	s.Background = colors.Uniform(ColorAccent)              // bg-accent
	s.Border.Radius = sides.NewValues(units.Dp(RadiusFull)) // rounded-full
	s.Min.X.Set(44, units.UnitDp)                           // w-11 (44px)
	s.Min.Y.Set(44, units.UnitDp)                           // h-11 (44px)
	s.Display = styles.Flex                                 // flex
	s.Align.Items = styles.Center                           // items-center
	s.Justify.Content = styles.Center                       // justify-center
}

// Emoji food circle: className="bg-white w-10 h-10 rounded-full flex items-center justify-center border border-primary select-none text-2xl"
func StyleFoodEmojiCircle(s *styles.Style) {
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
}

// className="w-6 aspect-square rounded-full border border-primary flex items-center justify-center"
func StyleCategoryIcon(s *styles.Style) {
	s.Min.X.Set(24, units.UnitDp)                           // w-6
	s.Min.Y.Set(24, units.UnitDp)                           // aspect-square (24px)
	s.Border.Radius = sides.NewValues(units.Dp(RadiusFull)) // rounded-full
	s.Border.Style.Set(styles.BorderSolid)                  // border
	s.Border.Width.Set(units.Dp(1))                         // border
	s.Border.Color.Set(colors.Uniform(ColorPrimary))        // border-primary
	s.Display = styles.Flex                                 // flex
	s.Align.Items = styles.Center                           // items-center
	s.Justify.Content = styles.Center                       // justify-center
}

// className="bg-white w-8 h-8 rounded-full flex items-center justify-center border border-primary select-none"
func StyleCategoryIconLarge(s *styles.Style) {
	s.Background = colors.Uniform(ColorWhite)               // bg-white
	s.Min.X.Set(32, units.UnitDp)                           // w-8
	s.Min.Y.Set(32, units.UnitDp)                           // h-8
	s.Border.Radius = sides.NewValues(units.Dp(RadiusFull)) // rounded-full
	s.Display = styles.Flex                                 // flex
	s.Align.Items = styles.Center                           // items-center
	s.Justify.Content = styles.Center                       // justify-center
	s.Border.Style.Set(styles.BorderSolid)                  // border
	s.Border.Width.Set(units.Dp(1))                         // border
	s.Border.Color.Set(colors.Uniform(ColorPrimary))        // border-primary
}

// ====================================================================================
// Dialog Component Styles
// ====================================================================================

func StyleDialogContainer(s *styles.Style) {
	s.Background = colors.Uniform(ColorWhite)
	s.Border.Radius = sides.NewValues(units.Dp(Radius2XL)) // BorderRadiusLarge
	s.Padding.Set(units.Dp(Spacing6))
	s.Gap.Set(units.Dp(Spacing4))
	s.Direction = styles.Column
	s.Min.X.Set(400, units.UnitDp)
	s.Max.X.Set(500, units.UnitDp)
}

func StyleDialogTitle(s *styles.Style) {
	s.Font.Size = units.Dp(FontSizeXL) // text-xl
	s.Font.Weight = WeightSemiBold
}

func StyleDialogButtonRow(s *styles.Style) {
	s.Direction = styles.Row
	s.Gap.Set(units.Dp(Spacing3))
	s.Justify.Content = styles.End
}

// ====================================================================================
// Drawer Component Styles (from nishiki-frontend/src/components/ui/Drawer.tsx)
// ====================================================================================

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
	s.Border.Radius.Top = units.Dp(RadiusDefault) // rounded-t
	s.Max.Y.Set(-32, units.UnitVh)                // max-h-[calc(100vh-2rem)] (100vh - 32px)
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
	s.Padding.Set(units.Dp(0), units.Dp(Spacing4))     // px-4
	s.Border.Style.Bottom = styles.BorderSolid         // border-b
	s.Border.Width.Bottom = units.Dp(1)                // border-b
	s.Border.Color.Set(colors.Uniform(ColorGrayLight)) // border-gray-light
	s.Display = styles.Flex                            // flex
	s.Align.Items = styles.Center                      // items-center
	// Note: shrink-0, grow-1, relative positioning would need layout system
}

// DrawerBody: className="p-4 overflow-y-auto max-h-full"
func StyleDrawerBody(s *styles.Style) {
	s.Display = styles.Flex                         // flex
	s.Direction = styles.Column                     // flex-col
	s.Gap.Set(units.Dp(Spacing4))                   // gap-4
}

func StyleDrawerBodyContent(s *styles.Style) {
	s.Padding.Set(units.Dp(Spacing4)) // p-4
	// Note: overflow-y-auto and max-h-full would need scrolling system
}

// DrawerFooter: className="mt-auto px-4 pt-2 pb-6 flex justify-end gap-4"
func StyleDrawerFooter(s *styles.Style) {
	s.Padding.Set(units.Dp(Spacing2), units.Dp(Spacing4), units.Dp(Spacing6), units.Dp(Spacing4)) // px-4 pt-2 pb-6
	s.Display = styles.Flex                      // flex
	s.Justify.Content = styles.End               // justify-end
	s.Gap.Set(units.Dp(Spacing4))                // gap-4
	// Note: mt-auto would need margin auto system
}

// DrawerTitle: className="text-2xl"
func StyleDrawerTitle(s *styles.Style) {
	s.Font.Size = units.Dp(FontSize2XL) // text-2xl
}

// DrawerClose button: className="absolute right-0 h-full px-4"
func StyleDrawerCloseButton(s *styles.Style) {
	s.Min.Y.Set(100, units.UnitPh)                 // h-full (parent height)
	s.Padding.Set(units.Dp(0), units.Dp(Spacing4)) // px-4
	// Note: absolute right-0 positioning would need positioning system
}

// ====================================================================================
// Dropdown Menu Component Styles
// ====================================================================================

// DropdownMenu trigger button: Button variant="ghost" className="w-12"
func StyleDropdownMenuTrigger(s *styles.Style) {
	StyleButtonGhost(s)           // Apply ghost button variant
	s.Min.X.Set(48, units.UnitDp) // w-12
}

// DropdownMenuContent: className="z-50 min-w-64 overflow-hidden rounded bg-white text-black shadow-around"
func StyleDropdownMenuContent(s *styles.Style) {
	// Note: z-50 would need z-index system
	s.Min.X.Set(256, units.UnitDp)                          // min-w-64 (256px)
	s.Border.Radius = sides.NewValues(units.Dp(RadiusDefault)) // rounded
	s.Background = colors.Uniform(ColorWhite)               // bg-white
	s.Color = colors.Uniform(ColorBlack)                    // text-black
	// Note: overflow-hidden and shadow-around would need additional systems
}

// DropdownMenuItem with proper styling
func StyleDropdownMenuItem(s *styles.Style) {
	s.Padding.Set(units.Dp(Spacing2), units.Dp(Spacing3)) // Standard menu item padding
	s.Cursor = cursors.Pointer
	s.Min.X.Set(100, units.UnitPw) // w-full (parent width) for proper clickable area
}

func StyleDropdownButton(s *styles.Style) {
	s.Background = colors.Uniform(ColorGrayLightest)
	s.Border.Radius = sides.NewValues(units.Dp(RadiusLG)) // BorderRadiusMedium
	s.Padding.Set(units.Dp(Spacing2), units.Dp(Spacing3))
	s.Gap.Set(units.Dp(Spacing1))
}
