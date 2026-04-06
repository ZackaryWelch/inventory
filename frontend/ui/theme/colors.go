package theme

import (
	"image/color"
)

// ====================================================================================
// DESIGN TOKENS - Migrated from Cogent Core styles to Gio
// Source: ui/styles/tokens.go (originally from nishiki-frontend)
// ====================================================================================

// ====================================================================================
// Color Palettes
// ====================================================================================

// Palette holds a complete set of semantic colors for a theme variant.
type Palette struct {
	// Primary colors
	PrimaryLightest color.NRGBA
	PrimaryLight    color.NRGBA
	Primary         color.NRGBA
	PrimaryDark     color.NRGBA

	// Accent colors
	Accent     color.NRGBA
	AccentDark color.NRGBA

	// Danger colors
	Danger     color.NRGBA
	DangerDark color.NRGBA

	// Background & surface
	Background color.NRGBA
	Surface    color.NRGBA
	SurfaceAlt color.NRGBA // Alternate surface (cards on surface)

	// Text colors
	TextPrimary   color.NRGBA
	TextSecondary color.NRGBA

	// Border & overlay
	Border  color.NRGBA
	Overlay color.NRGBA
}

// DarkPalette is the default dark color scheme.
var DarkPalette = Palette{
	PrimaryLightest: color.NRGBA{R: 30, G: 60, B: 57, A: 255},    // #1e3c39
	PrimaryLight:    color.NRGBA{R: 60, G: 120, B: 114, A: 255},  // #3c7872
	Primary:         color.NRGBA{R: 106, G: 179, B: 171, A: 255}, // #6ab3ab
	PrimaryDark:     color.NRGBA{R: 85, G: 143, B: 137, A: 255},  // #558f89

	Accent:     color.NRGBA{R: 252, G: 216, B: 132, A: 255}, // #fcd884
	AccentDark: color.NRGBA{R: 242, G: 192, B: 78, A: 255},  // #f2c04e

	Danger:     color.NRGBA{R: 220, G: 100, B: 100, A: 255}, // #dc6464
	DangerDark: color.NRGBA{R: 184, G: 72, B: 72, A: 255},   // #b84848

	Background: color.NRGBA{R: 24, G: 27, B: 32, A: 255}, // #181b20
	Surface:    color.NRGBA{R: 32, G: 36, B: 42, A: 255}, // #20242a
	SurfaceAlt: color.NRGBA{R: 42, G: 46, B: 54, A: 255}, // #2a2e36

	TextPrimary:   color.NRGBA{R: 230, G: 233, B: 240, A: 255}, // #e6e9f0
	TextSecondary: color.NRGBA{R: 160, G: 168, B: 180, A: 255}, // #a0a8b4

	Border:  color.NRGBA{R: 55, G: 60, B: 70, A: 255}, // #373c46
	Overlay: color.NRGBA{R: 0, G: 0, B: 0, A: 160},    // rgba(0,0,0,0.63)
}

// LightPalette is the light color scheme.
var LightPalette = Palette{
	PrimaryLightest: color.NRGBA{R: 230, G: 242, B: 241, A: 255}, // #e6f2f1
	PrimaryLight:    color.NRGBA{R: 171, G: 212, B: 207, A: 255}, // #abd4cf
	Primary:         color.NRGBA{R: 106, G: 179, B: 171, A: 255}, // #6ab3ab
	PrimaryDark:     color.NRGBA{R: 85, G: 143, B: 137, A: 255},  // #558f89

	Accent:     color.NRGBA{R: 252, G: 216, B: 132, A: 255}, // #fcd884
	AccentDark: color.NRGBA{R: 242, G: 192, B: 78, A: 255},  // #f2c04e

	Danger:     color.NRGBA{R: 205, G: 90, B: 90, A: 255}, // #cd5a5a
	DangerDark: color.NRGBA{R: 184, G: 72, B: 72, A: 255}, // #b84848

	Background: color.NRGBA{R: 245, G: 247, B: 250, A: 255}, // #f5f7fa
	Surface:    color.NRGBA{R: 255, G: 255, B: 255, A: 255}, // #ffffff
	SurfaceAlt: color.NRGBA{R: 249, G: 250, B: 251, A: 255}, // #f9fafb

	TextPrimary:   color.NRGBA{R: 30, G: 35, B: 45, A: 255},    // #1e232d
	TextSecondary: color.NRGBA{R: 100, G: 110, B: 125, A: 255}, // #646e7d

	Border:  color.NRGBA{R: 210, G: 215, B: 225, A: 255}, // #d2d7e1
	Overlay: color.NRGBA{R: 0, G: 0, B: 0, A: 128},       // rgba(0,0,0,0.5)
}

// ActivePalette is the currently active color palette. Defaults to dark.
var ActivePalette = DarkPalette

// ====================================================================================
// Color Accessors - All UI code should use these variables.
// They are initialized from the active palette.
// ====================================================================================

var (
	// Primary colors
	ColorPrimaryLightest = ActivePalette.PrimaryLightest
	ColorPrimaryLight    = ActivePalette.PrimaryLight
	ColorPrimary         = ActivePalette.Primary
	ColorPrimaryDark     = ActivePalette.PrimaryDark

	// Accent colors
	ColorAccent     = ActivePalette.Accent
	ColorAccentDark = ActivePalette.AccentDark

	// Danger colors
	ColorDanger     = ActivePalette.Danger
	ColorDangerDark = ActivePalette.DangerDark

	// Background & surface
	ColorBackground = ActivePalette.Background
	ColorSurface    = ActivePalette.Surface
	ColorSurfaceAlt = ActivePalette.SurfaceAlt

	// Text colors (semantic aliases)
	ColorTextPrimary   = ActivePalette.TextPrimary
	ColorTextSecondary = ActivePalette.TextSecondary

	// Border & overlay
	ColorBorder  = ActivePalette.Border
	ColorOverlay = ActivePalette.Overlay

	// Base colors (palette-independent)
	ColorWhite       = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	ColorBlack       = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	ColorTransparent = color.NRGBA{R: 0, G: 0, B: 0, A: 0}

	// Blue colors (for Authentik login button)
	ColorBlue600 = color.NRGBA{R: 37, G: 99, B: 235, A: 255} // #2563eb
	ColorBlue700 = color.NRGBA{R: 29, G: 78, B: 216, A: 255} // #1d4ed8

	// Legacy aliases (kept for compatibility, map to palette)
	ColorGrayLightest = ActivePalette.SurfaceAlt
	ColorGrayLight    = ActivePalette.Border
	ColorGray         = ActivePalette.TextSecondary
	ColorGrayDark     = ActivePalette.TextPrimary
	ColorGray600      = ActivePalette.TextSecondary
)

// ApplyPalette updates all color variables from the given palette.
func ApplyPalette(p Palette) {
	ActivePalette = p

	ColorPrimaryLightest = p.PrimaryLightest
	ColorPrimaryLight = p.PrimaryLight
	ColorPrimary = p.Primary
	ColorPrimaryDark = p.PrimaryDark

	ColorAccent = p.Accent
	ColorAccentDark = p.AccentDark

	ColorDanger = p.Danger
	ColorDangerDark = p.DangerDark

	ColorBackground = p.Background
	ColorSurface = p.Surface
	ColorSurfaceAlt = p.SurfaceAlt

	ColorTextPrimary = p.TextPrimary
	ColorTextSecondary = p.TextSecondary

	ColorBorder = p.Border
	ColorOverlay = p.Overlay

	// Update legacy aliases
	ColorGrayLightest = p.SurfaceAlt
	ColorGrayLight = p.Border
	ColorGray = p.TextSecondary
	ColorGrayDark = p.TextPrimary
	ColorGray600 = p.TextSecondary
}

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
	Spacing0  = 0  // 0
	Spacing1  = 4  // 0.25rem
	Spacing2  = 8  // 0.5rem
	Spacing3  = 12 // 0.75rem
	Spacing4  = 16 // 1rem
	Spacing5  = 20 // 1.25rem
	Spacing6  = 24 // 1.5rem
	Spacing8  = 32 // 2rem
	Spacing10 = 40 // 2.5rem
	Spacing12 = 48 // 3rem
	Spacing16 = 64 // 4rem
	Spacing18 = 72 // 4.5rem
	Spacing20 = 80 // 5rem
	Spacing24 = 96 // 6rem
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
