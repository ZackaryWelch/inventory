package theme

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

// NishikiTheme extends material.Theme with custom colors and styling
type NishikiTheme struct {
	*material.Theme

	// Custom semantic colors
	Primary       color.NRGBA
	PrimaryDark   color.NRGBA
	Accent        color.NRGBA
	Danger        color.NRGBA
	DangerDark    color.NRGBA

	// Background colors
	Background    color.NRGBA
	Surface       color.NRGBA

	// Text colors
	TextPrimary   color.NRGBA
	TextSecondary color.NRGBA

	// UI element colors
	Border        color.NRGBA
	Overlay       color.NRGBA
}

// NewTheme creates a customized Nishiki theme based on material design
func NewTheme() *NishikiTheme {
	// Start with default material theme and register Go fonts
	th := material.NewTheme()

	// Register the embedded Go fonts (required for WASM)
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	// Customize the base theme
	th.Fg = ColorTextPrimary        // Default text color
	th.Bg = ColorWhite              // Default background
	th.ContrastBg = ColorPrimary    // Primary color for important elements
	th.ContrastFg = ColorWhite      // Text on primary color

	// Create our extended theme
	nishikiTheme := &NishikiTheme{
		Theme:         th,
		Primary:       ColorPrimary,
		PrimaryDark:   ColorPrimaryDark,
		Accent:        ColorAccent,
		Danger:        ColorDanger,
		DangerDark:    ColorDangerDark,
		Background:    ColorWhite,
		Surface:       ColorGrayLightest,
		TextPrimary:   ColorTextPrimary,
		TextSecondary: ColorTextSecondary,
		Border:        ColorGrayLight,
		Overlay:       ColorOverlay,
	}

	return nishikiTheme
}

// ButtonStyle returns common button styling
type ButtonStyle struct {
	BackgroundColor color.NRGBA
	TextColor       color.NRGBA
	CornerRadius    unit.Dp
	Inset           unit.Dp
}

// PrimaryButton returns styling for primary action buttons
func (t *NishikiTheme) PrimaryButton() ButtonStyle {
	return ButtonStyle{
		BackgroundColor: t.Primary,
		TextColor:       ColorWhite,
		CornerRadius:    unit.Dp(RadiusDefault),
		Inset:           unit.Dp(Spacing4),
	}
}

// DangerButton returns styling for destructive action buttons
func (t *NishikiTheme) DangerButton() ButtonStyle {
	return ButtonStyle{
		BackgroundColor: t.Danger,
		TextColor:       ColorWhite,
		CornerRadius:    unit.Dp(RadiusDefault),
		Inset:           unit.Dp(Spacing4),
	}
}

// AccentButton returns styling for accent buttons
func (t *NishikiTheme) AccentButton() ButtonStyle {
	return ButtonStyle{
		BackgroundColor: t.Accent,
		TextColor:       ColorBlack,
		CornerRadius:    unit.Dp(RadiusDefault),
		Inset:           unit.Dp(Spacing4),
	}
}

// CancelButton returns styling for cancel buttons
func (t *NishikiTheme) CancelButton() ButtonStyle {
	return ButtonStyle{
		BackgroundColor: ColorGrayLightest,
		TextColor:       ColorBlack,
		CornerRadius:    unit.Dp(RadiusDefault),
		Inset:           unit.Dp(Spacing4),
	}
}

// LoginButton returns styling for the Authentik login button
func (t *NishikiTheme) LoginButton() ButtonStyle {
	return ButtonStyle{
		BackgroundColor: ColorBlue600,
		TextColor:       ColorWhite,
		CornerRadius:    unit.Dp(RadiusMD),
		Inset:           unit.Dp(Spacing4),
	}
}

// TextStyle returns common text styling configurations
type TextStyle struct {
	Size   unit.Sp
	Color  color.NRGBA
	Weight font.Weight
}

// H1 returns heading level 1 styling
func (t *NishikiTheme) H1() TextStyle {
	return TextStyle{
		Size:   unit.Sp(FontSize3XL),
		Color:  t.TextPrimary,
		Weight: font.Bold,
	}
}

// H2 returns heading level 2 styling
func (t *NishikiTheme) H2() TextStyle {
	return TextStyle{
		Size:   unit.Sp(FontSize2XL),
		Color:  t.TextPrimary,
		Weight: font.Bold,
	}
}

// H3 returns heading level 3 styling
func (t *NishikiTheme) H3() TextStyle {
	return TextStyle{
		Size:   unit.Sp(FontSizeXL),
		Color:  t.TextPrimary,
		Weight: font.SemiBold,
	}
}

// Body returns body text styling
func (t *NishikiTheme) Body() TextStyle {
	return TextStyle{
		Size:   unit.Sp(FontSizeBase),
		Color:  t.TextPrimary,
		Weight: font.Normal,
	}
}

// BodySecondary returns secondary body text styling
func (t *NishikiTheme) BodySecondary() TextStyle {
	return TextStyle{
		Size:   unit.Sp(FontSizeBase),
		Color:  t.TextSecondary,
		Weight: font.Normal,
	}
}

// Small returns small text styling
func (t *NishikiTheme) Small() TextStyle {
	return TextStyle{
		Size:   unit.Sp(FontSizeSM),
		Color:  t.TextSecondary,
		Weight: font.Normal,
	}
}

// CardStyle returns styling for card containers
type CardStyle struct {
	BackgroundColor color.NRGBA
	CornerRadius    unit.Dp
	Padding         unit.Dp
}

// Card returns default card styling
func (t *NishikiTheme) Card() CardStyle {
	return CardStyle{
		BackgroundColor: t.Surface,
		CornerRadius:    unit.Dp(RadiusDefault),
		Padding:         unit.Dp(Spacing4),
	}
}
