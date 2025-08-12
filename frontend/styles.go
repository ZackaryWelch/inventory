package main

import (
	"image/color"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
)

// Design system matching the React frontend exactly
// Color palette from globals.css
var (
	// Primary colors
	ColorPrimaryLightest = color.RGBA{R: 230, G: 242, B: 241, A: 255} // #e6f2f1
	ColorPrimaryLight    = color.RGBA{R: 171, G: 212, B: 207, A: 255} // #abd4cf
	ColorPrimary         = color.RGBA{R: 106, G: 179, B: 171, A: 255} // #6ab3ab
	ColorPrimaryDark     = color.RGBA{R: 95, G: 161, B: 154, A: 255}  // #5fa19a

	// Accent colors
	ColorAccent     = color.RGBA{R: 252, G: 216, B: 132, A: 255} // #fcd884
	ColorAccentDark = color.RGBA{R: 241, G: 197, B: 96, A: 255}  // #f1c560

	// Danger colors
	ColorDanger     = color.RGBA{R: 205, G: 90, B: 90, A: 255} // #cd5a5a
	ColorDangerDark = color.RGBA{R: 185, G: 81, B: 81, A: 255} // #b95151

	// Gray scale
	ColorGrayLightest = color.RGBA{R: 248, G: 248, B: 248, A: 255} // #f8f8f8
	ColorGrayLight    = color.RGBA{R: 238, G: 238, B: 238, A: 255} // #eeeeee
	ColorGray         = color.RGBA{R: 189, G: 189, B: 189, A: 255} // #bdbdbd
	ColorGrayDark     = color.RGBA{R: 119, G: 119, B: 119, A: 255} // #777777

	// Base colors
	ColorWhite   = color.RGBA{R: 255, G: 255, B: 255, A: 255} // #ffffff
	ColorBlack   = color.RGBA{R: 34, G: 34, B: 34, A: 255}    // #222222
	ColorOverlay = color.RGBA{R: 0, G: 0, B: 0, A: 64}        // rgba(0, 0, 0, 0.25)
)

// Button styles matching React frontend button variants
func (app *App) styleButtonPrimary(btn *core.Button) {
	btn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorPrimary)
		s.Color = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(units.Dp(12), units.Dp(24)) // md size: h-10 px-12
		s.Font.Size = units.Dp(16)                // text-base
		s.Font.Weight = styles.WeightMedium
		s.Min.X.Set(70, units.UnitDp) // min-w-[70px]
		s.Gap.Set(units.Dp(10))       // gap-2.5
	})
}

func (app *App) styleButtonDanger(btn *core.Button) {
	btn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorDanger)
		s.Color = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(units.Dp(12), units.Dp(24))
		s.Font.Size = units.Dp(16)
		s.Font.Weight = styles.WeightMedium
		s.Min.X.Set(70, units.UnitDp)
		s.Gap.Set(units.Dp(10))
	})
}

func (app *App) styleButtonAccent(btn *core.Button) {
	btn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorAccent)
		s.Color = colors.Uniform(ColorBlack)
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(units.Dp(12), units.Dp(24))
		s.Font.Size = units.Dp(16)
		s.Font.Weight = styles.WeightMedium
		s.Min.X.Set(70, units.UnitDp)
		s.Gap.Set(units.Dp(10))
	})
}

func (app *App) styleButtonCancel(btn *core.Button) {
	btn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(color.RGBA{R: 0, G: 0, B: 0, A: 0}) // transparent
		s.Color = colors.Uniform(ColorBlack)
		s.Border.Radius = styles.BorderRadiusMedium
		s.Padding.Set(units.Dp(12), units.Dp(24))
		s.Font.Size = units.Dp(16)
		s.Font.Weight = styles.WeightMedium
		s.Min.X.Set(70, units.UnitDp)
		s.Gap.Set(units.Dp(10))
	})
}

func (app *App) styleButtonIcon(btn *core.Button) {
	btn.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorGrayLight)
		s.Border.Radius = styles.BorderRadiusFull
		s.Padding.Set(units.Dp(12)) // h-12 w-12
		s.Min.X.Set(48, units.UnitDp)
		s.Min.Y.Set(48, units.UnitDp)
	})
}

// Card styles matching React frontend
func (app *App) styleCard(frame *core.Frame) {
	frame.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusMedium // rounded (default)
		s.Margin.Bottom = units.Dp(8)               // mb-2
	})
}

// Mobile layout styles matching React frontend
func (app *App) styleMobileLayout(container *core.Frame) {
	container.Styler(func(s *styles.Style) {
		s.Padding.Set(units.Dp(48), units.Dp(0), units.Dp(64), units.Dp(0)) // pt-12 pb-16
		s.Min.Y.Set(100, units.UnitVh)
	})
}

// Header styles (fixed top-0 z-40 w-full h-12)
func (app *App) styleHeader(header *core.Frame) {
	header.Styler(func(s *styles.Style) {
		s.Position = styles.PositionFixed
		s.Top = units.Dp(0)
		s.Z = 40
		s.Min.X.Set(100, units.UnitVw)
		s.Min.Y.Set(48, units.UnitDp)
		s.Background = colors.Uniform(ColorWhite)
		s.Display = styles.Flex
		s.Align.Items = styles.Center
		s.Justify.Content = styles.Center
	})
}

// Bottom menu styles (fixed bottom-0 left-0 z-40 w-full h-16)
func (app *App) styleBottomMenu(menu *core.Frame) {
	menu.Styler(func(s *styles.Style) {
		s.Position = styles.PositionFixed
		s.Bottom = units.Dp(0)
		s.Left = units.Dp(0)
		s.Z = 40
		s.Min.X.Set(100, units.UnitVw)
		s.Min.Y.Set(64, units.UnitDp)
		s.Background = colors.Uniform(ColorWhite)
		s.Border.Style.Top = styles.BorderSolid
		s.Border.Width.Top = units.Dp(1)
		s.Border.Color.Top = ColorGrayLight
	})
}

// Count badge styles (w-8 h-8 rounded-full bg-accent)
func (app *App) styleCountBadge(badge *core.Frame) {
	badge.Styler(func(s *styles.Style) {
		s.Min.X.Set(32, units.UnitDp)
		s.Min.Y.Set(32, units.UnitDp)
		s.Border.Radius = styles.BorderRadiusFull
		s.Background = colors.Uniform(ColorAccent)
		s.Display = styles.Flex
		s.Justify.Content = styles.Center
		s.Align.Items = styles.Center
	})
}

// Food card specific styles
func (app *App) styleFoodCard(card *core.Frame) {
	card.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorWhite)
		s.Border.Radius = styles.BorderRadiusMedium
		s.Margin.Bottom = units.Dp(8) // mb-2
		s.Min.X.Set(100, units.UnitVw)
		s.Display = styles.Flex
	})
}

// Food emoji circle (bg-white w-10 h-10 rounded-full border border-primary)
func (app *App) styleFoodEmojiCircle(circle *core.Frame) {
	circle.Styler(func(s *styles.Style) {
		s.Background = colors.Uniform(ColorWhite)
		s.Min.X.Set(40, units.UnitDp)
		s.Min.Y.Set(40, units.UnitDp)
		s.Border.Radius = styles.BorderRadiusFull
		s.Border.Style.Set(styles.BorderSolid)
		s.Border.Width.Set(units.Dp(1))
		s.Border.Color.Set(ColorPrimary)
		s.Display = styles.Flex
		s.Align.Items = styles.Center
		s.Justify.Content = styles.Center
		s.Font.Size = units.Dp(24) // text-2xl for emoji
	})
}

// Text styles matching React frontend
func (app *App) styleTextLarge(text *core.Text) {
	text.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(18)  // text-lg
		s.Font.Family = "leading-6" // leading-6
	})
}

func (app *App) styleTextSmall(text *core.Text) {
	text.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(12)              // text-xs
		s.Color = colors.Uniform(ColorGrayDark) // text-gray-dark
	})
}

func (app *App) styleTextMedium(text *core.Text) {
	text.Styler(func(s *styles.Style) {
		s.Font.Size = units.Dp(14) // text-sm
	})
}

// Icon sizing to match React frontend exactly
type IconSize int

const (
	IconSize2  IconSize = 8  // w-2 h-2
	IconSize3  IconSize = 12 // w-3 h-3
	IconSize4  IconSize = 16 // w-4 h-4
	IconSize5  IconSize = 20 // w-5 h-5
	IconSize6  IconSize = 24 // w-6 h-6
	IconSize8  IconSize = 32 // w-8 h-8
	IconSize10 IconSize = 40 // w-10 h-10
	IconSize12 IconSize = 48 // w-12 h-12
)

func (app *App) styleIcon(icon *core.Icon, size IconSize, clr color.RGBA) {
	icon.Styler(func(s *styles.Style) {
		s.Min.X.Set(size, units.UnitDp)
		s.Min.Y.Set(size, units.UnitDp)
		s.Color = colors.Uniform(clr)
	})
}

// Enhanced mobile layout that exactly matches React MobileLayout
func (app *App) createMobileLayoutWrapper() *core.Frame {
	wrapper := core.NewFrame(app.mainContainer)
	app.styleMobileLayout(wrapper)
	return wrapper
}

// Create header that matches React Header component
func (app *App) createMobileHeader(title string, showBack bool) *core.Frame {
	header := core.NewFrame(nil) // Will be positioned absolutely
	app.styleHeader(header)

	// Left section (back button)
	if showBack {
		leftContainer := core.NewFrame(header)
		leftContainer.Styler(func(s *styles.Style) {
			s.Position = styles.PositionAbsolute
			s.Left = units.Dp(0)
			s.Min.Y.Set(48, units.UnitDp)
		})

		backBtn := core.NewButton(leftContainer).SetIcon(icons.ArrowBack)
		backBtn.Styler(func(s *styles.Style) {
			s.Min.Y.Set(48, units.UnitDp)
			s.Aspect.Ratio = units.Dp(1)  // aspect-square
			s.Padding.Left = units.Dp(16) // pl-4
			s.Display = styles.Flex
			s.Align.Items = styles.Center
		})
		app.styleIcon(backBtn.Icon, IconSize4, ColorGrayDark)
	}

	// Center title
	if title != "" {
		titleText := core.NewText(header).SetText(title)
		titleText.Styler(func(s *styles.Style) {
			s.Font.Size = units.Dp(20)
			s.Font.Weight = styles.WeightSemiBold
		})
	}

	return header
}

// Bottom navigation matching React BottomMenu
func (app *App) createMobileBottomMenu() *core.Frame {
	bottomMenu := core.NewFrame(nil) // Will be positioned absolutely
	app.styleBottomMenu(bottomMenu)

	// Grid container (grid h-full max-w-lg grid-cols-3 mx-auto font-medium)
	gridContainer := core.NewFrame(bottomMenu)
	gridContainer.Styler(func(s *styles.Style) {
		s.Display = styles.Grid
		s.Min.Y.Set(64, units.UnitDp)
		s.Max.X.Set(units.Dp(512))            // max-w-lg (32rem = 512px)
		s.Margin.Set(units.Dp(0), units.Auto) // mx-auto
		s.Font.Weight = styles.WeightMedium
	})

	return bottomMenu
}

// Apply consistent spacing and layout
func (app *App) applyMobileSpacing(content *core.Frame) {
	content.Styler(func(s *styles.Style) {
		s.Padding.Set(units.Dp(24), units.Dp(16)) // pt-6 px-4
		s.Display = styles.Flex
		s.Direction = styles.Column
		s.Gap.Set(units.Dp(8)) // gap-2
	})
}

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
