# Cogent Core Styling Guide: Syncing from React/Tailwind

This guide explains how to translate styles from the React frontend (which uses Tailwind CSS) to the Cogent Core Go frontend.

## Table of Contents
- [Understanding the Difference](#understanding-the-difference)
- [Core Concepts](#core-concepts)
- [Translation Rules](#translation-rules)
- [Common Patterns](#common-patterns)
- [Debugging Strategies](#debugging-strategies)
- [Best Practices](#best-practices)

---

## Understanding the Difference

### React Frontend: DOM + CSS
- **Rendering**: HTML DOM elements styled with CSS
- **Layout**: Browser's CSS layout engine (Flexbox, Grid)
- **Styles**: Tailwind utility classes compiled to CSS
- **Inspection**: Browser DevTools shows computed CSS

### Cogent Core Frontend: Canvas Rendering
- **Rendering**: Direct canvas drawing (like a game engine)
- **Layout**: Cogent Core's custom layout engine
- **Styles**: Go `styles.Style` struct with layout properties
- **Inspection**: No DOM - must understand layout from code

**Key Insight**: You cannot inspect Cogent Core layouts with browser DevTools because there are no HTML elements. The entire UI is rendered to a `<canvas>` element.

---

## Core Concepts

### 1. The Style Struct

Cogent Core uses a `styles.Style` struct instead of CSS:

```go
type Style struct {
    // Layout
    Display   Displays      // flex, grid, etc.
    Direction Directions    // row, column
    Wrap      bool

    // Alignment (AlignSet has Content, Items, Self)
    Justify   AlignSet      // Main axis alignment
    Align     AlignSet      // Cross axis alignment

    // Sizing
    Min       math32.Vector2  // Minimum size
    Max       math32.Vector2  // Maximum size
    Grow      math32.Vector2  // Flex grow factors

    // Spacing
    Padding   sides.Values
    Margin    sides.Values
    Gap       math32.Vector2

    // Visual
    Background color.RGBA
    Color      color.RGBA
    Border     BorderStyle
    // ... and more
}
```

### 2. Flexbox Mapping

**CSS Flexbox** → **Cogent Core**

| CSS Property | Cogent Core Property | Notes |
|--------------|---------------------|-------|
| `display: flex` | `s.Display = styles.Flex` | Enable flex layout |
| `flex-direction: row` | `s.Direction = styles.Row` | Default direction |
| `flex-direction: column` | `s.Direction = styles.Column` | Vertical layout |
| `justify-content: center` | `s.Justify.Content = styles.Center` | Main axis |
| `align-items: center` | `s.Align.Items = styles.Center` | Cross axis |
| `flex-grow: 1` | `s.Grow.Set(1, 0)` | Grow on X axis only |
| `gap: 8px` | `s.Gap.Set(units.Dp(8))` | Space between items |

**Important**: In Cogent Core:
- **Row layout** (default): Main axis = horizontal, Cross axis = vertical
- **Column layout**: Main axis = vertical, Cross axis = horizontal
- Therefore:
  - In Row: `Justify` = horizontal, `Align` = vertical
  - In Column: `Justify` = vertical, `Align` = horizontal

### 3. Sizing Units

**Tailwind/CSS** → **Cogent Core**

| CSS Unit | Cogent Core | Example |
|----------|-------------|---------|
| `px` | `units.Dp(N)` | `s.Padding.Set(units.Dp(16))` |
| `rem` | `units.Dp(N * 16)` | 1rem = 16px |
| `%` | `units.UnitPct` | `s.Min.X.Set(50, units.UnitPct)` |
| `vw` | `units.UnitVw` | `s.Min.X.Set(100, units.UnitVw)` |
| `vh` | `units.UnitVh` | `s.Min.Y.Set(100, units.UnitVh)` |
| `auto` / grow | `s.Grow.Set(1, 0)` | Preferred over viewport units |

**Avoid Overusing Viewport Units**: Instead of `Min.X.Set(100, units.UnitVw)`, prefer `Grow.Set(1, 0)` which allows the element to fill its parent naturally.

---

## Translation Rules

### Step 1: Identify the Element Hierarchy

In React:
```tsx
<div className="flex items-center justify-center h-screen">
  <div className="transform flex flex-col items-center justify-center">
    <div className="w-32 h-26 mb-20">
      <LogoVerticalPrimary />
    </div>
    <div className="w-full max-w-sm">
      <button className="w-full px-4 py-3">Sign in</button>
    </div>
  </div>
</div>
```

In Cogent Core:
```go
// Outer container
container := core.NewFrame(parent)
container.Styler(StyleLoginContainer)

// Inner content
content := core.NewFrame(container)
content.Styler(StyleLoginContent)

// Logo
logo := core.NewFrame(content)
logo.Styler(StyleLoginLogo)

// Button container
btnContainer := core.NewFrame(content)
btnContainer.Styler(StyleLoginButtonContainer)
```

### Step 2: Map Tailwind Classes to Style Properties

**Example: Login Container**

React:
```tsx
<div className="flex items-center justify-center h-screen">
```

Cogent Core:
```go
func StyleLoginContainer(s *styles.Style) {
    s.Display = styles.Flex           // flex
    s.Justify.Content = styles.Center // justify-center (horizontal in Row)
    s.Align.Items = styles.Center     // items-center (vertical in Row)
    s.Min.Y.Set(100, units.UnitVh)    // h-screen
    s.Grow.Set(1, 1)                  // Grow to fill parent
}
```

**Example: Column Layout**

React:
```tsx
<div className="pt-6 px-4 pb-2 flex flex-col gap-2">
```

Cogent Core:
```go
func StyleContentColumn(s *styles.Style) {
    s.Display = styles.Flex                        // flex
    s.Direction = styles.Column                    // flex-col
    s.Padding.Set(
        units.Dp(24),  // pt-6 (6 * 4 = 24px)
        units.Dp(16),  // px-4 (4 * 4 = 16px)
        units.Dp(8),   // pb-2 (2 * 4 = 8px)
        units.Dp(16),  // px-4
    )
    s.Gap.Set(units.Dp(8))  // gap-2 (2 * 4 = 8px)
}
```

### Step 3: Handle Background Colors and Borders

**Colors**: Define in `tokens.go`, reference in style functions

```go
// In tokens.go
const (
    ColorPrimary = color.RGBA{R: 106, G: 179, B: 171, A: 255} // #6ab3ab
    ColorGrayLightest = color.RGBA{R: 248, G: 248, B: 248, A: 255} // #f8f8f8
)

// In style function
func StyleMainBackground(s *styles.Style) {
    s.Background = colors.Uniform(ColorGrayLightest)
}
```

**Borders**: Use `sides.NewValues()` for radius

```go
// Tailwind: rounded (0.625rem = 10px)
s.Border.Radius = sides.NewValues(units.Dp(10))

// Tailwind: rounded-md (0.375rem = 6px)
s.Border.Radius = sides.NewValues(units.Dp(6))

// Tailwind: rounded-full
s.Border.Radius = sides.NewValues(units.Dp(9999))
```

---

## Common Patterns

### Pattern 1: Full-Screen Centered Content (Login Page)

**React**:
```tsx
<body className="min-h-screen bg-primary-lightest">
  <div className="flex items-center justify-center h-screen">
    <div className="flex flex-col items-center justify-center">
      {/* Content */}
    </div>
  </div>
</body>
```

**Cogent Core**:
```go
// Body styling
func StyleMainBackground(s *styles.Style) {
    s.Min.Y.Set(100, units.UnitVh)
    s.Min.X.Set(100, units.UnitVw)
    s.Background = colors.Uniform(ColorGrayLightest)
    // NO Display or Direction - let children control layout
}

// Main container
func StyleLoginContainer(s *styles.Style) {
    s.Display = styles.Flex
    s.Justify.Content = styles.Center
    s.Align.Items = styles.Center
    s.Min.Y.Set(100, units.UnitVh)
    s.Grow.Set(1, 1)  // KEY: Fill parent
}

// Content wrapper
func StyleLoginContent(s *styles.Style) {
    s.Display = styles.Flex
    s.Direction = styles.Column
    s.Align.Items = styles.Center
    s.Justify.Content = styles.Center
}
```

### Pattern 2: Mobile Layout Container

**React** (`MobileLayout.tsx`):
```tsx
<div className="pt-12 pb-16 min-h-screen">
  <Header />
  {children}
  <BottomMenu />
</div>
```

**Cogent Core**:
```go
func StyleMobileLayoutContainer(s *styles.Style) {
    s.Padding.Set(
        units.Dp(48),  // pt-12 (12 * 4 = 48px)
        units.Dp(0),
        units.Dp(64),  // pb-16 (16 * 4 = 64px)
        units.Dp(0),
    )
    s.Min.Y.Set(100, units.UnitVh)
    // NO Display, Direction, or Background!
    // Background is on Body, not here
}
```

**Critical Mistake to Avoid**:
```go
// ❌ WRONG - Don't add flex/direction/background to container
func StyleMobileLayoutContainer(s *styles.Style) {
    s.Display = styles.Flex              // NO!
    s.Direction = styles.Column          // NO!
    s.Background = colors.Uniform(...)   // NO! (on Body)
}
```

### Pattern 3: Content Area with Padding

**React**:
```tsx
<div className="pt-6 px-4 pb-2 flex flex-col gap-2">
  {/* Cards or content */}
</div>
```

**Cogent Core**:
```go
func StyleContentColumn(s *styles.Style) {
    s.Display = styles.Flex
    s.Direction = styles.Column
    s.Padding.Set(units.Dp(24), units.Dp(16), units.Dp(8), units.Dp(16))
    s.Gap.Set(units.Dp(8))
}
```

### Pattern 4: Buttons with Variants

**React** (Tailwind variants):
```tsx
// Primary: "bg-primary text-white enabled:hover:bg-primary-dark"
// Danger: "bg-danger text-white"
```

**Cogent Core**:
```go
// Base styles
func StyleButtonBase(s *styles.Style) {
    s.Font.Size = units.Dp(16)  // text-base
    s.Border.Radius = sides.NewValues(units.Dp(10))
    s.Display = styles.Flex
    s.Align.Items = styles.Center
    s.Justify.Content = styles.Center
    s.Gap.Set(units.Dp(10))
    s.Cursor = cursors.Pointer
}

// Primary variant
func StyleButtonPrimary(s *styles.Style) {
    StyleButtonBase(s)
    s.Background = colors.Uniform(ColorPrimary)
    s.Color = colors.Uniform(ColorWhite)
}

// Danger variant
func StyleButtonDanger(s *styles.Style) {
    StyleButtonBase(s)
    s.Background = colors.Uniform(ColorDanger)
    s.Color = colors.Uniform(ColorWhite)
}
```

---

## Debugging Strategies

### 1. Compare Visual Outputs

Since you can't inspect Cogent Core with DevTools:

```bash
# Capture screenshots of both versions
python3 scripts/selenium-screenshot.py http://localhost:3000 /tmp/react.png 5
python3 scripts/selenium-screenshot.py http://localhost:3002 /tmp/go.png 5

# Compare side-by-side
```

### 2. Understand the HTML Structure First

**Always start by examining the React HTML**:

```bash
curl -s http://localhost:3000/page | grep -A 20 "<body"
```

This shows you:
- Element hierarchy
- Tailwind classes applied
- Default HTML block-level vs inline behavior

### 3. Check Cogent Core Documentation

```bash
# Check available style properties
go doc cogentcore.org/core/styles Style

# Check alignment options
go doc cogentcore.org/core/styles Aligns

# Check units
go doc cogentcore.org/core/styles/units
```

### 4. Think in Layout Phases

Cogent Core calculates layout in phases:
1. **Size allocation**: Min/Max constraints applied
2. **Flex distribution**: Grow factors calculated
3. **Alignment**: Justify and Align positioning
4. **Rendering**: Drawing to canvas

If something isn't centering:
- Check parent has `Display = Flex`
- Check parent has appropriate Justify/Align
- Check child has `Grow` set if it should fill space
- Check sizing constraints (Min/Max)

---

## Best Practices

### 1. Organize Styles by Component

**File Structure**:
```
ui/styles/
├── tokens.go       # Colors, typography, spacing constants
├── layouts.go      # Container layouts (MobileLayout, etc.)
├── components.go   # Reusable components (buttons, cards)
└── utilities.go    # Utility styles (backgrounds, helpers)
```

### 2. Match Tailwind Naming in Comments

Always include the original Tailwind class in comments:

```go
// className="flex items-center justify-center h-screen"
func StyleLoginContainer(s *styles.Style) {
    s.Display = styles.Flex           // flex
    s.Justify.Content = styles.Center // justify-center
    s.Align.Items = styles.Center     // items-center
    s.Min.Y.Set(100, units.UnitVh)    // h-screen
}
```

This makes it easy to:
- Verify the translation is correct
- Update when React styles change
- Understand the intent

### 3. Use Spacing Constants from Tailwind

Define spacing scale matching Tailwind:

```go
// In tokens.go - Tailwind spacing scale (multiply by 4 for px)
const (
    Spacing0   = 0    // 0px
    Spacing1   = 4    // 0.25rem = 4px
    Spacing2   = 8    // 0.5rem = 8px
    Spacing4   = 16   // 1rem = 16px
    Spacing6   = 24   // 1.5rem = 24px
    Spacing12  = 48   // 3rem = 48px
    Spacing16  = 64   // 4rem = 64px
    Spacing20  = 80   // 5rem = 80px
)

// Usage
s.Padding.Set(units.Dp(Spacing6), units.Dp(Spacing4))  // pt-6 px-4
```

### 4. Don't Set Background Color Twice

**React Pattern**:
```tsx
<body className="bg-primary-lightest">  {/* Background here */}
  <div className="pt-12 pb-16">         {/* NO background here */}
    {children}
  </div>
</body>
```

**Cogent Core Pattern**:
```go
// ✅ Correct - Background on Body only
func StyleMainBackground(s *styles.Style) {
    s.Background = colors.Uniform(ColorGrayLightest)
}

func StyleMobileLayoutContainer(s *styles.Style) {
    // NO s.Background here!
}
```

### 5. Prefer Grow Over Viewport Units

**Instead of**:
```go
s.Min.X.Set(100, units.UnitVw)  // Forces full viewport width
s.Min.Y.Set(100, units.UnitVh)  // Forces full viewport height
```

**Prefer**:
```go
s.Grow.Set(1, 1)  // Grows to fill parent naturally
s.Min.Y.Set(100, units.UnitVh)  // Use vh only for height constraints
```

This is more flexible and handles responsive layouts better.

### 6. Test Iteratively

When porting a complex layout:

1. **Start with the outermost container**
2. **Verify it renders correctly** (screenshot comparison)
3. **Add the next level of children**
4. **Verify again**
5. **Repeat**

Don't try to port an entire page at once - work from outside-in, one level at a time.

---

## Example: Complete Login Page Port

### React Source

```tsx
// body in layout.tsx
<body className="min-h-screen bg-primary-lightest antialiased font-outfit">
  {/* LoginPage.tsx */}
  <div className="flex items-center justify-center h-screen">
    <div className="transform flex flex-col items-center justify-center">
      <LogoVerticalPrimary className="w-32 h-26 mb-20" />

      {/* AuthentikLoginButton.tsx */}
      <div className="w-full max-w-sm">
        <button className="w-full flex items-center justify-center px-4 py-3
                          border border-transparent rounded-md shadow-sm
                          text-base font-medium text-white bg-blue-600">
          Sign in with Authentik
        </button>
        <div className="mt-4 text-center">
          <p className="text-sm text-gray-600">
            Secure authentication powered by Authentik
          </p>
        </div>
      </div>
    </div>
  </div>
</body>
```

### Cogent Core Port

```go
// Step 1: Define styles (ui/styles/)

// Body background
func StyleMainBackground(s *styles.Style) {
    s.Min.Y.Set(100, units.UnitVh)
    s.Min.X.Set(100, units.UnitVw)
    s.Background = colors.Uniform(ColorGrayLightest)
}

// Outer container: flex items-center justify-center h-screen
func StyleLoginContainer(s *styles.Style) {
    s.Display = styles.Flex
    s.Justify.Content = styles.Center
    s.Align.Items = styles.Center
    s.Min.Y.Set(100, units.UnitVh)
    s.Grow.Set(1, 1)
}

// Inner content: flex flex-col items-center justify-center
func StyleLoginContent(s *styles.Style) {
    s.Display = styles.Flex
    s.Direction = styles.Column
    s.Align.Items = styles.Center
    s.Justify.Content = styles.Center
}

// Logo: w-32 h-26 mb-20
func StyleLoginLogo(s *styles.Style) {
    s.Display = styles.Flex
    s.Align.Items = styles.Center
    s.Justify.Content = styles.Center
    s.Min.X.Set(128, units.UnitDp)   // w-32 (32 * 4)
    s.Min.Y.Set(104, units.UnitDp)   // h-26 (26 * 4)
    s.Margin.Bottom = units.Dp(80)    // mb-20 (20 * 4)
}

// Button container: w-full max-w-sm
func StyleLoginButtonContainer(s *styles.Style) {
    s.Direction = styles.Column
    s.Max.X.Set(384, units.UnitDp)    // max-w-sm (24rem)
}

// Login button: w-full px-4 py-3 rounded-md bg-blue-600 text-white
func StyleButtonLogin(s *styles.Style) {
    s.Font.Size = units.Dp(16)
    s.Border.Radius = sides.NewValues(units.Dp(6))
    s.Display = styles.Flex
    s.Align.Items = styles.Center
    s.Justify.Content = styles.Center
    s.Gap.Set(units.Dp(10))
    s.Cursor = cursors.Pointer
    s.Grow.Set(1, 0)  // w-full within container
    s.Padding.Set(units.Dp(12), units.Dp(16))
    s.Background = colors.Uniform(ColorBlue600)
    s.Color = colors.Uniform(ColorWhite)
    s.Font.Weight = WeightMedium
}

// Subtitle: mt-4 text-center text-sm text-gray-600
func StyleLoginSubtitle(s *styles.Style) {
    s.Font.Size = units.Dp(14)
    s.Color = colors.Uniform(ColorGray600)
    s.Text.Align = AlignCenter
    s.Margin.Top = units.Dp(16)
}

// Step 2: Build UI structure (app/app_methods.go)

func (app *App) createMainUI(b *core.Body) {
    b.Styler(appstyles.StyleMainBackground)

    app.mainContainer = core.NewFrame(b)
    app.mainContainer.Styler(appstyles.StyleMainContainer)

    if !app.isSignedIn {
        app.showLoginView()
    }
}

func (app *App) showLoginView() {
    app.mainContainer.DeleteChildren()

    // Override container styling for login
    app.mainContainer.Styler(appstyles.StyleLoginContainer)

    // Login content
    loginContent := core.NewFrame(app.mainContainer)
    loginContent.Styler(appstyles.StyleLoginContent)

    // Logo
    logo := core.NewFrame(loginContent)
    logo.Styler(appstyles.StyleLoginLogo)
    logoText := core.NewText(logo).SetText("NISHIKI")
    logoText.Styler(appstyles.StyleAppTitle)

    // Button container
    buttonContainer := core.NewFrame(loginContent)
    buttonContainer.Styler(appstyles.StyleLoginButtonContainer)

    // Login button
    loginBtn := core.NewButton(buttonContainer)
    loginBtn.SetText("Sign in with Authentik")
    loginBtn.SetIcon(icons.Login)
    loginBtn.Styler(appstyles.StyleButtonLogin)
    loginBtn.OnClick(func(e events.Event) {
        app.handleLogin()
    })

    // Subtitle
    subtitle := core.NewText(buttonContainer)
    subtitle.SetText("Secure authentication powered by Authentik")
    subtitle.Styler(appstyles.StyleLoginSubtitle)

    app.mainContainer.Update()
}
```

---

## Troubleshooting Checklist

When layouts don't match:

- [ ] **Did you check the React HTML structure first?**
- [ ] **Are you setting Display = Flex where needed?**
- [ ] **Are Justify and Align targeting the right axes for your Direction?**
- [ ] **Did you set Grow instead of viewport width units?**
- [ ] **Is the background color on Body, not the container?**
- [ ] **Are you removing Direction from containers that shouldn't have it?**
- [ ] **Did you check parent constraints (Min/Max) aren't preventing growth?**
- [ ] **Are you using the correct spacing constants from tokens.go?**
- [ ] **Did you capture screenshots to compare visually?**

---

## Conclusion

Porting React/Tailwind styles to Cogent Core requires:

1. **Understanding canvas-based rendering** vs DOM/CSS
2. **Mapping Tailwind classes** to Cogent Core style properties
3. **Respecting the layout hierarchy** and not adding unnecessary properties
4. **Using Grow instead of viewport units** where appropriate
5. **Iterative testing** with visual comparison

Always refer back to the React source as the source of truth, and remember that Cogent Core's layout engine is powerful but different from CSS - understand the concepts, don't just blindly translate.
