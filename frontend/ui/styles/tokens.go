package styles

import (
	"image/color"

	"cogentcore.org/core/text/rich"
	"cogentcore.org/core/text/text"
)

// ====================================================================================
// DESIGN TOKENS - Exact values from nishiki-frontend
// Source: tailwind.config.ts and globals.css
// ====================================================================================

// ====================================================================================
// Color System - From nishiki-frontend/src/styles/globals.css and tailwind.config.ts
// ====================================================================================

var (
	// Primary colors (matching --color-primary-* from globals.css)
	ColorPrimaryLightest = color.RGBA{R: 214, G: 234, B: 231, A: 255} // #d6eae7
	ColorPrimaryLight    = color.RGBA{R: 149, G: 206, B: 198, A: 255} // #95cec6
	ColorPrimary         = color.RGBA{R: 106, G: 179, B: 171, A: 255} // #6ab3ab (--color-primary)
	ColorPrimaryDark     = color.RGBA{R: 85, G: 143, B: 137, A: 255}  // #558f89

	// Accent colors (matching --color-accent-*)
	ColorAccent     = color.RGBA{R: 252, G: 216, B: 132, A: 255} // #fcd884 (--color-accent)
	ColorAccentDark = color.RGBA{R: 242, G: 192, B: 78, A: 255}  // #f2c04e

	// Danger colors (matching --color-danger-*)
	ColorDanger     = color.RGBA{R: 205, G: 90, B: 90, A: 255}  // #cd5a5a (--color-danger)
	ColorDangerDark = color.RGBA{R: 184, G: 72, B: 72, A: 255}  // #b84848

	// Gray scale (matching --color-gray-*)
	ColorGrayLightest = color.RGBA{R: 249, G: 250, B: 251, A: 255} // #f9fafb
	ColorGrayLight    = color.RGBA{R: 229, G: 231, B: 235, A: 255} // #e5e7eb
	ColorGray         = color.RGBA{R: 156, G: 163, B: 175, A: 255} // #9ca3af
	ColorGrayDark     = color.RGBA{R: 75, G: 85, B: 99, A: 255}   // #4b5563

	// Base colors (matching --color-white/black)
	ColorWhite       = color.RGBA{R: 255, G: 255, B: 255, A: 255} // #ffffff
	ColorBlack       = color.RGBA{R: 0, G: 0, B: 0, A: 255}       // #000000
	ColorOverlay     = color.RGBA{R: 0, G: 0, B: 0, A: 128}       // rgba(0, 0, 0, 0.5)
	ColorTransparent = color.RGBA{R: 0, G: 0, B: 0, A: 0}         // transparent
)

// ====================================================================================
// Typography Scale - From tailwind.config.ts
// ====================================================================================

// Font sizes from tailwind.config.ts
const (
	FontSize2XS  = 10 // 0.625rem - text-2xs (custom)
	FontSizeXS   = 12 // 0.75rem - text-xs
	FontSizeSM   = 14 // 0.875rem - text-sm
	FontSizeBase = 16 // 1rem - text-base
	FontSizeLG   = 18 // 1.125rem - text-lg
	FontSizeXL   = 20 // 1.25rem - text-xl
	FontSize2XL  = 24 // 1.5rem - text-2xl
	FontSize3XL  = 30 // 1.875rem - text-3xl
)

// Line heights (leading)
const (
	LineHeightNone    = 1.0   // leading-none
	LineHeightTight   = 1.25  // leading-tight
	LineHeightSnug    = 1.375 // leading-snug
	LineHeightNormal  = 1.5   // leading-normal
	LineHeightRelaxed = 1.625 // leading-relaxed
	LineHeightLoose   = 2.0   // leading-loose
)

// Specific line heights in pixels (for precise matching)
const (
	LineHeight12 = 12 // leading-3
	LineHeight16 = 16 // leading-4
	LineHeight20 = 20 // leading-5
	LineHeight24 = 24 // leading-6
	LineHeight28 = 28 // leading-7
)

// Font weight constants mapping from old styles to new rich weights
const (
	WeightNormal   = rich.Weights(3) // normal
	WeightMedium   = rich.Weights(4) // medium
	WeightSemiBold = rich.Weights(5) // semibold
	WeightBold     = rich.Weights(6) // bold
)

// ====================================================================================
// Spacing System - From Tailwind spacing scale
// Tailwind: 0, 0.5, 1, 1.5, 2, 2.5, 3, 3.5, 4, 4.5, 5, 6, 7, 8, 9, 10, 11, 12, etc.
// ====================================================================================

// Spacing values in dp (4px base unit)
const (
	Spacing0   = 0   // 0
	Spacing1   = 4   // 0.25rem
	Spacing2   = 8   // 0.5rem
	Spacing3   = 12  // 0.75rem
	Spacing4   = 16  // 1rem
	Spacing4_5 = 18  // 1.125rem (custom)
	Spacing5   = 20  // 1.25rem
	Spacing6   = 24  // 1.5rem
	Spacing8   = 32  // 2rem
	Spacing10  = 40  // 2.5rem
	Spacing12  = 48  // 3rem
	Spacing16  = 64  // 4rem
	Spacing18  = 72  // 4.5rem (custom)
	Spacing20  = 80  // 5rem
	Spacing24  = 96  // 6rem
)

// ====================================================================================
// Border Radius - From tailwind.config.ts
// ====================================================================================

// Border radius values from tailwind.config.ts
// IMPORTANT: Cogent Core v0.3.12 requires sides.NewValues(units.Dp(X))
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

// ====================================================================================
// Box Shadows - From tailwind.config.ts
// Note: Cogent Core may have limited shadow support - verify capabilities
// ====================================================================================

const (
	ShadowAround = "0 0 8px 4px rgba(0, 0, 0, 0.1)" // custom shadow-around
)

// Standard Tailwind shadows (if needed)
const (
	ShadowSM      = "0 1px 2px 0 rgba(0, 0, 0, 0.05)"
	ShadowDefault = "0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px -1px rgba(0, 0, 0, 0.1)"
	ShadowMD      = "0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -2px rgba(0, 0, 0, 0.1)"
	ShadowLG      = "0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -4px rgba(0, 0, 0, 0.1)"
)

// ====================================================================================
// Text Alignment Constants
// ====================================================================================

// Text alignment constants mapping from old styles to new text aligns
const (
	AlignStart  = text.Aligns(0) // start
	AlignCenter = text.Aligns(1) // center
	AlignEnd    = text.Aligns(2) // end
)

// ====================================================================================
// Helper Functions for Common Unit Conversions
// ====================================================================================

// Note: Use units.Dp(), units.Vw(), units.Vh() directly from cogentcore.org/core/styles/units
// These provide the correct API for Cogent Core v0.3.12
