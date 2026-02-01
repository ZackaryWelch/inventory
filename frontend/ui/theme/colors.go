package theme

import (
	"image/color"
)

// ====================================================================================
// DESIGN TOKENS - Migrated from Cogent Core styles to Gio
// Source: ui/styles/tokens.go (originally from nishiki-frontend)
// ====================================================================================

// ====================================================================================
// Color System - Using color.NRGBA for Gio (non-premultiplied alpha)
// ====================================================================================

var (
	// Primary colors
	ColorPrimaryLightest = color.NRGBA{R: 230, G: 242, B: 241, A: 255} // #e6f2f1
	ColorPrimaryLight    = color.NRGBA{R: 171, G: 212, B: 207, A: 255} // #abd4cf
	ColorPrimary         = color.NRGBA{R: 106, G: 179, B: 171, A: 255} // #6ab3ab
	ColorPrimaryDark     = color.NRGBA{R: 85, G: 143, B: 137, A: 255}  // #558f89

	// Accent colors
	ColorAccent     = color.NRGBA{R: 252, G: 216, B: 132, A: 255} // #fcd884
	ColorAccentDark = color.NRGBA{R: 242, G: 192, B: 78, A: 255}  // #f2c04e

	// Danger colors
	ColorDanger     = color.NRGBA{R: 205, G: 90, B: 90, A: 255} // #cd5a5a
	ColorDangerDark = color.NRGBA{R: 184, G: 72, B: 72, A: 255} // #b84848

	// Gray scale
	ColorGrayLightest = color.NRGBA{R: 249, G: 250, B: 251, A: 255} // #f9fafb
	ColorGrayLight    = color.NRGBA{R: 229, G: 231, B: 235, A: 255} // #e5e7eb
	ColorGray         = color.NRGBA{R: 156, G: 163, B: 175, A: 255} // #9ca3af
	ColorGrayDark     = color.NRGBA{R: 75, G: 85, B: 99, A: 255}    // #4b5563

	// Base colors
	ColorWhite       = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // #ffffff
	ColorBlack       = color.NRGBA{R: 0, G: 0, B: 0, A: 255}       // #000000
	ColorOverlay     = color.NRGBA{R: 0, G: 0, B: 0, A: 128}       // rgba(0, 0, 0, 0.5)
	ColorTransparent = color.NRGBA{R: 0, G: 0, B: 0, A: 0}         // transparent

	// Text colors (semantic aliases)
	ColorTextPrimary   = ColorGrayDark // Default text color
	ColorTextSecondary = ColorGray     // Secondary/muted text color

	// Blue colors (for Authentik login button)
	ColorBlue600 = color.NRGBA{R: 37, G: 99, B: 235, A: 255}  // #2563eb - blue-600
	ColorBlue700 = color.NRGBA{R: 29, G: 78, B: 216, A: 255}  // #1d4ed8 - blue-700
	ColorGray600 = color.NRGBA{R: 75, G: 85, B: 99, A: 255}   // #4b5563 - gray-600
)

// ====================================================================================
// Typography Scale
// ====================================================================================

// Font sizes (in sp - scale-independent pixels)
const (
	FontSize2XS  = 10 // 0.625rem - text-2xs
	FontSizeXS   = 12 // 0.75rem - text-xs
	FontSizeSM   = 14 // 0.875rem - text-sm
	FontSizeBase = 16 // 1rem - text-base
	FontSizeLG   = 18 // 1.125rem - text-lg
	FontSizeXL   = 20 // 1.25rem - text-xl
	FontSize2XL  = 24 // 1.5rem - text-2xl
	FontSize3XL  = 30 // 1.875rem - text-3xl
)

// Line heights
const (
	LineHeightNone    = 1.0   // leading-none
	LineHeightTight   = 1.25  // leading-tight
	LineHeightSnug    = 1.375 // leading-snug
	LineHeightNormal  = 1.5   // leading-normal
	LineHeightRelaxed = 1.625 // leading-relaxed
	LineHeightLoose   = 2.0   // leading-loose
)

// ====================================================================================
// Spacing System (in dp - density-independent pixels)
// ====================================================================================

const (
	Spacing0   = 0  // 0
	Spacing1   = 4  // 0.25rem
	Spacing2   = 8  // 0.5rem
	Spacing2_5 = 10 // 0.625rem (custom)
	Spacing3   = 12 // 0.75rem
	Spacing4   = 16 // 1rem
	Spacing4_5 = 18 // 1.125rem
	Spacing5   = 20 // 1.25rem
	Spacing6   = 24 // 1.5rem
	Spacing8   = 32 // 2rem
	Spacing10  = 40 // 2.5rem
	Spacing12  = 48 // 3rem
	Spacing16  = 64 // 4rem
	Spacing18  = 72 // 4.5rem
	Spacing20  = 80 // 5rem
	Spacing24  = 96 // 6rem
)

// ====================================================================================
// Border Radius (in dp)
// ====================================================================================

const (
	RadiusXS      = 2    // 0.125rem - rounded-xs
	RadiusSM      = 4    // 0.25rem - rounded-sm
	RadiusDefault = 10   // 0.625rem - rounded (custom default)
	RadiusMD      = 6    // 0.375rem - rounded-md
	RadiusLG      = 8    // 0.5rem - rounded-lg
	RadiusXL      = 12   // 0.75rem - rounded-xl
	Radius2XL     = 16   // 1rem - rounded-2xl
	Radius3XL     = 24   // 1.5rem - rounded-3xl
	RadiusFull    = 9999 // rounded-full (fully rounded)
)
