# Nishiki Frontend Porting Plan

**Goal**: Port remaining functionality from `nishiki-frontend/` (React/Next.js) to `frontend/` (Go/Cogent Core WASM) with exact visual and functional parity, enabling removal of the old frontend directory and streamlining the project for future development.

**Target**: Achieve pixel-perfect style matching when both frontends are run in Firefox stateless mode, then deprecate `nishiki-frontend/`.

---

## Style Matching Protocol

### Critical: Visual Parity Requirements

The primary objective is **exact visual matching** between the React and Go frontends. All development must follow this protocol:

#### 1. Reference-Driven Development

**Before implementing any component**:
1. Run `nishiki-frontend` in Firefox stateless mode: `cd nishiki-frontend && npm run dev`
2. Navigate to the component/view you're implementing
3. Take screenshots of the component in different states (normal, hover, active, error)
4. Use Firefox DevTools to inspect exact CSS values:
   - Open Inspector (F12)
   - Select element
   - Note computed styles: padding, margin, font-size, line-height, colors, border-radius
   - Record exact pixel values and color hex codes

**During implementation**:
1. Run `frontend` in Firefox stateless mode: `cd frontend && ./bin/web && ./bin/serve`
2. Open both apps side-by-side in Firefox
3. Compare visually at every step
4. Adjust Go styles until pixel-perfect match achieved

**Verification**:
1. Screenshot both implementations
2. Overlay screenshots in image editor (use difference blend mode)
3. Iterate until no visible differences remain

#### 2. Design Token Mapping

All styling must use exact values from `nishiki-frontend/tailwind.config.ts` and `nishiki-frontend/src/styles/globals.css`.

**Color System** - Port to `frontend/app/styles.go`:

```go
// From tailwind.config.ts colors and globals.css CSS variables
var (
    // Primary colors
    ColorPrimaryLightest = color.RGBA{R: 214, G: 234, B: 231, A: 255} // #d6eae7
    ColorPrimaryLight    = color.RGBA{R: 149, G: 206, B: 198, A: 255} // #95cec6
    ColorPrimary         = color.RGBA{R: 106, G: 179, B: 171, A: 255} // #6ab3ab (--color-primary)
    ColorPrimaryDark     = color.RGBA{R: 85, G: 143, B: 137, A: 255}  // #558f89

    // Accent colors
    ColorAccent     = color.RGBA{R: 252, G: 216, B: 132, A: 255} // #fcd884 (--color-accent)
    ColorAccentDark = color.RGBA{R: 242, G: 192, B: 78, A: 255}  // #f2c04e

    // Danger colors
    ColorDanger     = color.RGBA{R: 205, G: 90, B: 90, A: 255}  // #cd5a5a (--color-danger)
    ColorDangerDark = color.RGBA{R: 184, G: 72, B: 72, A: 255}  // #b84848

    // Gray scale
    ColorGrayLightest = color.RGBA{R: 249, G: 250, B: 251, A: 255} // #f9fafb
    ColorGrayLight    = color.RGBA{R: 229, G: 231, B: 235, A: 255} // #e5e7eb
    ColorGray         = color.RGBA{R: 156, G: 163, B: 175, A: 255} // #9ca3af
    ColorGrayDark     = color.RGBA{R: 75, G: 85, B: 99, A: 255}   // #4b5563

    // Base colors
    ColorOverlay = color.RGBA{R: 0, G: 0, B: 0, A: 128}       // rgba(0, 0, 0, 0.5)
    ColorWhite   = color.RGBA{R: 255, G: 255, B: 255, A: 255} // #ffffff
    ColorBlack   = color.RGBA{R: 0, G: 0, B: 0, A: 255}       // #000000
)
```

**Typography Scale** - Exact pixel values from Tailwind:

```go
// Font sizes from tailwind.config.ts
const (
    FontSize2XS  = 10  // 0.625rem - text-2xs (custom)
    FontSizeXS   = 12  // 0.75rem - text-xs
    FontSizeSM   = 14  // 0.875rem - text-sm
    FontSizeBase = 16  // 1rem - text-base
    FontSizeLG   = 18  // 1.125rem - text-lg
    FontSizeXL   = 20  // 1.25rem - text-xl
    FontSize2XL  = 24  // 1.5rem - text-2xl
    FontSize3XL  = 30  // 1.875rem - text-3xl
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
    LineHeight12 = 12  // leading-3
    LineHeight16 = 16  // leading-4
    LineHeight20 = 20  // leading-5
    LineHeight24 = 24  // leading-6
    LineHeight28 = 28  // leading-7
)
```

**Spacing System** - From Tailwind spacing scale:

```go
// Spacing values in dp (4px base unit)
// Tailwind: 0, 0.5, 1, 1.5, 2, 2.5, 3, 3.5, 4, 4.5, 5, 6, 7, 8, 9, 10, 11, 12, 14, 16, 18, 20, 24, etc.
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
```

**Border Radius** - From tailwind.config.ts:

```go
// Border radius values from tailwind.config.ts
// IMPORTANT: Cogent Core v0.3.12 requires sides.NewValues(units.Dp(X))
import "cogentcore.org/core/styles/sides"

const (
    RadiusXS      = 2   // 0.125rem - rounded-xs
    RadiusSM      = 4   // 0.25rem - rounded-sm
    RadiusDefault = 10  // 0.625rem - rounded (custom default)
    RadiusMD      = 6   // 0.375rem - rounded-md
    RadiusLG      = 8   // 0.5rem - rounded-lg
    RadiusXL      = 12  // 0.75rem - rounded-xl
    Radius2XL     = 16  // 1rem - rounded-2xl
    Radius3XL     = 24  // 1.5rem - rounded-3xl
    RadiusFull    = 9999 // rounded-full (fully rounded)
)

// Usage in style functions:
func StyleCard(s *styles.Style) {
    s.Border.Radius = sides.NewValues(units.Dp(RadiusDefault)) // 10px rounded
    s.Background = colors.Uniform(ColorWhite)
    s.Padding.Set(units.Dp(Spacing4)) // 16px padding
}

func StyleButton(s *styles.Style) {
    s.Border.Radius = sides.NewValues(units.Dp(RadiusFull)) // fully rounded
    s.Padding.Set(units.Dp(Spacing3), units.Dp(Spacing6)) // py-3 px-6
}
```

**Shadows** - From tailwind.config.ts:

```go
// Box shadow definitions
// Note: Cogent Core may have limited shadow support - verify capabilities

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
```

#### 3. Component-Level Style Matching

**MobileLayout Pattern** (from `nishiki-frontend/src/components/layouts/MobileLayout.tsx`):

```typescript
// React/Tailwind reference
<div className="flex min-h-screen flex-col bg-gray-lightest">
  <div className="flex flex-col gap-2 px-4 pt-6 pb-16"> {/* Main content */}
    {children}
  </div>
</div>
```

**Exact Go translation**:

```go
// frontend/app/styles.go
func StyleMobileLayoutContainer(s *styles.Style) {
    s.Display = styles.Flex
    s.Direction = styles.Column
    s.Min.Y.Set(100, units.UnitVh) // min-h-screen
    s.Background = colors.Uniform(ColorGrayLightest)
}

func StyleMobileLayoutContent(s *styles.Style) {
    s.Display = styles.Flex
    s.Direction = styles.Column
    s.Gap.Set(units.Dp(8)) // gap-2 (8px)
    s.Padding.Set(
        units.Dp(24), // pt-6 (24px)
        units.Dp(16), // px-4 (16px)
        units.Dp(64), // pb-16 (64px)
        units.Dp(16), // px-4 (16px)
    )
}
```

**Card Component Pattern** (from shadcn/ui):

```typescript
// React/Tailwind reference
<div className="rounded-lg border bg-white text-black shadow-sm p-4">
  {content}
</div>
```

**Exact Go translation**:

```go
func StyleCard(s *styles.Style) {
    s.Border.Radius = sides.NewValues(units.Dp(8)) // rounded-lg
    s.Border.Width = sides.NewValues(units.Dp(1))
    s.Border.Color = colors.Uniform(ColorGrayLight)
    s.Background = colors.Uniform(ColorWhite)
    s.Color = colors.Uniform(ColorBlack)
    s.Padding.Set(units.Dp(16)) // p-4
    // Note: shadow-sm may need custom implementation in Cogent Core
}
```

**Button Variants** (from `nishiki-frontend/src/components/ui/Button.tsx`):

```typescript
// Primary button: bg-primary text-white hover:bg-primary-dark
// Danger button: bg-danger text-white hover:bg-danger-dark
// Accent button: bg-accent text-black hover:bg-accent-dark
```

**Exact Go translations**:

```go
func StyleButtonPrimary(s *styles.Style) {
    s.Background = colors.Uniform(ColorPrimary)
    s.Color = colors.Uniform(ColorWhite)
    s.Border.Radius = sides.NewValues(units.Dp(RadiusFull))
    s.Padding.Set(units.Dp(12), units.Dp(24)) // py-3 px-6
    s.Font.Weight = styles.WeightSemiBold
    s.Font.Size = units.Dp(FontSizeSM) // 14px

    // Hover state (use Cogent Core state styling)
    s.State.BackgroundHover = colors.Uniform(ColorPrimaryDark)
}

func StyleButtonDanger(s *styles.Style) {
    s.Background = colors.Uniform(ColorDanger)
    s.Color = colors.Uniform(ColorWhite)
    s.Border.Radius = sides.NewValues(units.Dp(RadiusFull))
    s.Padding.Set(units.Dp(12), units.Dp(24))
    s.Font.Weight = styles.WeightSemiBold
    s.Font.Size = units.Dp(FontSizeSM)
    s.State.BackgroundHover = colors.Uniform(ColorDangerDark)
}

func StyleButtonAccent(s *styles.Style) {
    s.Background = colors.Uniform(ColorAccent)
    s.Color = colors.Uniform(ColorBlack) // Important: accent uses black text
    s.Border.Radius = sides.NewValues(units.Dp(RadiusFull))
    s.Padding.Set(units.Dp(12), units.Dp(24))
    s.Font.Weight = styles.WeightSemiBold
    s.Font.Size = units.Dp(FontSizeSM)
    s.State.BackgroundHover = colors.Uniform(ColorAccentDark)
}

func StyleButtonCancel(s *styles.Style) {
    s.Background = colors.Uniform(ColorGrayLight)
    s.Color = colors.Uniform(ColorGrayDark)
    s.Border.Radius = sides.NewValues(units.Dp(RadiusFull))
    s.Padding.Set(units.Dp(12), units.Dp(24))
    s.Font.Weight = styles.WeightSemiBold
    s.Font.Size = units.Dp(FontSizeSM)
    s.State.BackgroundHover = colors.Uniform(ColorGray)
}
```

#### 4. Animation Matching

From `tailwind.config.ts` keyframes and animations:

```javascript
// Tailwind animations to port
fadeIn: 'fadeIn 0.2s ease-out'
fadeOut: 'fadeOut 0.15s ease-in'
slideInFromBottom: 'slideInFromBottom 0.3s ease-out'
slideOutToBottom: 'slideOutToBottom 0.2s ease-in'
slideInFromRight: 'slideInFromRight 0.3s ease-out'
slideOutToRight: 'slideOutToRight 0.2s ease-in'
scaleIn: 'scaleIn 0.2s ease-out'
```

**Go implementation strategy**:
- Check Cogent Core animation capabilities
- May need to use CSS animations via style injection
- Or implement using Cogent Core's animation system
- Match exact timing functions and durations

---

## Architecture Refactoring

### Current State Issues

The `frontend/app/` directory currently has monolithic files that violate separation of concerns:
- `app_methods.go` (574 lines) - mixes UI rendering, API calls, and event handlers
- `collections_ui.go` (464 lines) - collection-specific UI
- `objects_ui.go` (582 lines) - object-specific UI
- `search_filter.go` (422 lines) - search functionality
- `ui_management.go` (569 lines) - general UI management
- `styles.go` (1500+ lines) - all styling

### Target Architecture

Reorganize to match `nishiki-frontend` structure for maintainability:

```
frontend/
├── app/
│   ├── app.go                      # App struct and initialization only
│   ├── config_desktop.go           # Desktop config (keep)
│   ├── config_wasm.go              # WASM config (keep)
│   └── router.go                   # NEW: View routing logic
│
├── pkg/                            # NEW: Shared packages
│   ├── api/                        # API client layer
│   │   ├── auth/
│   │   │   └── client.go           # Auth API calls
│   │   ├── groups/
│   │   │   └── client.go           # Groups API calls
│   │   ├── collections/
│   │   │   └── client.go           # Collections API calls
│   │   ├── containers/
│   │   │   └── client.go           # Containers API calls
│   │   ├── objects/
│   │   │   └── client.go           # Objects API calls
│   │   ├── categories/
│   │   │   └── client.go           # Categories API calls
│   │   └── common/
│   │       ├── client.go           # Shared HTTP client logic
│   │       ├── token_fetcher.go    # Auto token refresh
│   │       └── result.go           # Result type for error handling
│   │
│   ├── types/                      # Shared type definitions
│   │   ├── user.go
│   │   ├── group.go
│   │   ├── collection.go
│   │   ├── container.go
│   │   ├── object.go
│   │   ├── category.go
│   │   └── common.go
│   │
│   └── utils/                      # Utility functions
│       ├── validation.go
│       ├── formatting.go
│       └── helpers.go
│
├── ui/                             # NEW: UI layer
│   ├── components/                 # Reusable components
│   │   ├── badge.go
│   │   ├── button.go
│   │   ├── card.go
│   │   ├── checkbox.go
│   │   ├── dialog.go
│   │   ├── drawer.go
│   │   ├── dropdown.go
│   │   ├── form.go
│   │   ├── input.go
│   │   ├── label.go
│   │   ├── select.go
│   │   └── ...
│   │
│   ├── layouts/                    # Layout components
│   │   ├── mobile_layout.go       # Main mobile layout
│   │   ├── header.go              # Header component
│   │   └── bottom_menu.go         # Bottom navigation
│   │
│   ├── views/                      # Full page views
│   │   ├── login.go               # Login view
│   │   ├── dashboard.go           # Dashboard view
│   │   ├── groups.go              # Groups list view
│   │   ├── group_detail.go        # Single group view
│   │   ├── members.go             # Group members view
│   │   ├── collections.go         # Collections list view
│   │   ├── collection_detail.go   # Single collection view
│   │   ├── objects.go             # Objects list view
│   │   ├── object_detail.go       # Single object view
│   │   ├── search.go              # Search view
│   │   └── profile.go             # Profile view
│   │
│   └── styles/                     # Styling system
│       ├── tokens.go              # Design tokens (colors, spacing, etc.)
│       ├── components.go          # Component styles
│       ├── layouts.go             # Layout styles
│       └── utilities.go           # Utility styles
│
├── web/                            # Generated web assets (keep)
├── bin/                            # Build outputs (keep)
├── cmd/                            # Build commands (keep)
└── config/                         # Configuration (keep)
```

### Refactoring Steps

**Step 1: Extract API Clients**

Move API logic from `app_methods.go` to dedicated API clients:

```go
// Before: app_methods.go
func (app *App) fetchGroups() error {
    resp, err := app.makeAuthenticatedRequest("GET", "/groups", nil)
    // ... HTTP logic ...
}

// After: pkg/api/groups/client.go
package groups

type Client struct {
    baseURL    string
    httpClient *http.Client
    tokenFetcher func() (string, error)
}

func (c *Client) List() ([]types.Group, error) {
    // HTTP logic here
}

func (c *Client) Get(id string) (*types.Group, error) { }
func (c *Client) Create(req CreateGroupRequest) (*types.Group, error) { }
func (c *Client) Update(id string, req UpdateGroupRequest) (*types.Group, error) { }
func (c *Client) Delete(id string) error { }
```

**Step 2: Extract View Logic**

Move view rendering from `app_methods.go` to `ui/views/`:

```go
// Before: app_methods.go (200+ lines for showGroupsView)
func (app *App) showGroupsView() {
    // Massive function with UI construction
}

// After: ui/views/groups.go
package views

func RenderGroupsList(parent core.Widget, app *app.App) {
    // Header
    header := components.RenderHeader(parent, components.HeaderProps{
        Title: "Groups",
        ShowBack: true,
        OnBack: func() { app.ShowDashboardView() },
    })

    // Content
    content := core.NewFrame(parent)
    content.Styler(styles.StyleContentColumn)

    // Create button
    createBtn := components.RenderButton(content, components.ButtonProps{
        Text: "Create Group",
        Icon: icons.Add,
        Variant: components.ButtonPrimary,
        OnClick: func() { app.ShowCreateGroupDialog() },
    })

    // Groups list
    if len(app.Groups) == 0 {
        RenderEmptyState(content, "No groups found. Create your first group!")
    } else {
        for _, group := range app.Groups {
            RenderGroupCard(content, GroupCardProps{
                Group: group,
                OnClick: func() { app.ShowGroupDetail(group.ID) },
            })
        }
    }
}
```

**Step 3: Component Library**

Create reusable components in `ui/components/`:

```go
// ui/components/button.go
package components

type ButtonVariant int
const (
    ButtonPrimary ButtonVariant = iota
    ButtonDanger
    ButtonAccent
    ButtonCancel
)

type ButtonProps struct {
    Text     string
    Icon     icons.Icon
    Variant  ButtonVariant
    Size     ButtonSize
    OnClick  func()
    Disabled bool
}

func RenderButton(parent core.Widget, props ButtonProps) *core.Button {
    btn := core.NewButton(parent).SetText(props.Text)

    if props.Icon != "" {
        btn.SetIcon(props.Icon)
    }

    // Apply variant styles
    switch props.Variant {
    case ButtonPrimary:
        btn.Styler(styles.StyleButtonPrimary)
    case ButtonDanger:
        btn.Styler(styles.StyleButtonDanger)
    case ButtonAccent:
        btn.Styler(styles.StyleButtonAccent)
    case ButtonCancel:
        btn.Styler(styles.StyleButtonCancel)
    }

    // Apply size styles
    switch props.Size {
    case ButtonSM:
        btn.Styler(styles.StyleButtonSM)
    case ButtonMD:
        btn.Styler(styles.StyleButtonMD)
    case ButtonLG:
        btn.Styler(styles.StyleButtonLG)
    }

    if props.OnClick != nil {
        btn.OnClick(func(e events.Event) {
            if !props.Disabled {
                props.OnClick()
            }
        })
    }

    return btn
}
```

**Step 4: Styling Organization**

Split `styles.go` into logical modules:

```go
// ui/styles/tokens.go - Design tokens only
package styles

import (
    "image/color"
    "cogentcore.org/core/styles/units"
)

// Colors
var (
    ColorPrimary = color.RGBA{R: 106, G: 179, B: 171, A: 255}
    // ... all color constants
)

// Typography
const (
    FontSize2XS = 10
    // ... all font sizes
)

// Spacing
const (
    Spacing0 = 0
    // ... all spacing values
)

// Border Radius
const (
    RadiusXS = 2
    // ... all radius values
)
```

```go
// ui/styles/components.go - Component-specific styles
package styles

func StyleButtonPrimary(s *styles.Style) { }
func StyleButtonDanger(s *styles.Style) { }
func StyleCard(s *styles.Style) { }
func StyleCardHeader(s *styles.Style) { }
// ... all component styles
```

```go
// ui/styles/layouts.go - Layout-specific styles
package styles

func StyleMobileLayoutContainer(s *styles.Style) { }
func StyleMobileLayoutContent(s *styles.Style) { }
func StyleHeaderRow(s *styles.Style) { }
// ... all layout styles
```

---

## Feature Implementation Priority

### Phase 1: Foundation (Week 1-2)

**Goal**: Establish architecture and component library

1. **Refactor existing code**:
   - Create `pkg/api/` structure with existing API calls
   - Move types to `pkg/types/`
   - Split `styles.go` into `ui/styles/` modules
   - Create `ui/components/` with existing components

2. **Build core component library** (matching shadcn/ui):
   - Button (all variants)
   - Card (header, content, footer)
   - Input (text, number, date)
   - Label
   - Dialog (base, confirm, form)
   - Drawer (base, selection)
   - Checkbox
   - Select/Dropdown

3. **Implement layouts**:
   - MobileLayout (header + content + bottom menu)
   - Header component (back button, title, actions)
   - BottomMenu component (navigation tabs)

**Acceptance Criteria**:
- All components visually match nishiki-frontend in Firefox
- No inline styles - all styling through style functions
- Component library documented with usage examples

### Phase 2: Groups Management (Week 3)

**Goal**: Complete groups feature parity

1. **Groups List View** (enhance existing):
   - Port `nishiki-frontend/src/components/pages/GroupsPage.tsx`
   - Create group dialog with validation
   - Delete group confirmation
   - Pull-to-refresh functionality

2. **Group Detail View** (new):
   - Port `nishiki-frontend/src/components/pages/GroupSinglePage.tsx`
   - Group information display
   - Member count and list preview
   - Edit/delete actions
   - Leave group option

3. **Members Management** (new):
   - Port `nishiki-frontend/src/components/pages/MembersPage.tsx`
   - Member list with roles
   - Invite member dialog
   - Remove member confirmation
   - Copy invite link functionality

4. **Join Group Flow** (new):
   - Join by invite hash (`/groups/join/[hash]`)
   - Invite link generation
   - Success/error states

**Files to Reference**:
- `nishiki-frontend/src/features/groups/`
- `nishiki-frontend/src/lib/api/group/client/`

**Acceptance Criteria**:
- Can create, view, edit, delete groups
- Can invite and remove members
- Can join groups via invite link
- Visual match with React version

### Phase 3: Collections & Containers (Week 4-5)

**Goal**: Complete collections feature parity

1. **Collections List View** (enhance existing):
   - Port collection card design exactly
   - Object type icons (food, book, videogame, music, boardgame)
   - Collection stats (item count, last updated)
   - Create collection dialog with object type selection

2. **Collection Detail View** (new):
   - Container list within collection
   - Collection statistics
   - Edit collection (name, description, object type)
   - Delete collection with confirmation

3. **Container Management** (new):
   - Create container dialog (name, description, space)
   - Container card with space visualization
   - Edit container
   - Delete container with warning if contains objects
   - Move container (reorder)

4. **Space Management** (new):
   - Space calculation display
   - Space availability indicators
   - Auto-organize functionality (calls backend `/organize`)

**Files to Reference**:
- `nishiki-frontend/src/lib/api/container/client/`
- Container-related components from features

**Acceptance Criteria**:
- Can create, view, edit, delete collections
- Can manage containers within collections
- Space visualization matches React version
- Object type icons display correctly

### Phase 4: Objects/Foods Management (Week 6-7)

**Goal**: Complete objects feature parity

1. **Objects List View** (enhance existing):
   - Port `nishiki-frontend/src/components/pages/FoodsPage.tsx`
   - Filter by category, expiration, tags
   - Search within container
   - Sort options (name, date, expiration)
   - Grid/list view toggle

2. **Object Creation/Editing** (new):
   - Multi-step form (or single form with sections):
     - Basic info (name, quantity, unit)
     - Category selection with autocomplete
     - Tags (add/remove)
     - Expiration date picker
     - Notes
   - Form validation matching Zod schemas
   - Image upload (if backend supports)

3. **Object Detail View** (new):
   - Full object information display
   - Expiration warnings (color-coded)
   - Edit/delete actions
   - Move to different container option

4. **Bulk Import** (new):
   - CSV/JSON upload
   - Field mapping interface
   - Import preview table
   - Validation errors display
   - Confirm import

**Files to Reference**:
- `nishiki-frontend/src/features/foods/`
- `nishiki-frontend/src/lib/api/food/client/`

**Acceptance Criteria**:
- Can create, view, edit, delete objects
- Filtering and search work correctly
- Expiration warnings display
- Bulk import functional

### Phase 5: Search & Categories (Week 8)

**Goal**: Global search and category management

1. **Global Search** (enhance existing):
   - Search across all collections
   - Filter by object type, category, tags
   - Recent searches (localStorage)
   - Search suggestions
   - Results grouped by collection

2. **Categories Management** (new):
   - List all categories
   - Create category (name, icon, color)
   - Edit category
   - Delete category (warn if in use)
   - Category hierarchy (if backend supports)

**Files to Reference**:
- Search components from `nishiki-frontend/src/components/`
- `nishiki-frontend/src/lib/api/` category endpoints

**Acceptance Criteria**:
- Global search finds objects across collections
- Categories CRUD functional
- Category filtering works in object lists

### Phase 6: Polish & Navigation (Week 9)

**Goal**: Bottom navigation, animations, final touches

1. **Bottom Navigation Menu**:
   - Port `nishiki-frontend/src/components/parts/BottomMenu.tsx`
   - Home, Groups, Collections, Search, Profile tabs
   - Active state indicators
   - Icon + label layout

2. **Animations**:
   - Port Tailwind animations to Cogent Core
   - Drawer slide-in/out
   - Dialog fade-in/out
   - Page transitions

3. **Error Handling & Loading States**:
   - Network error displays
   - Loading skeletons for lists
   - Toast notifications (if Cogent Core supports)
   - Retry mechanisms

4. **Profile Enhancements**:
   - User settings (if backend supports)
   - Notification preferences
   - Theme toggle (if applicable)

**Acceptance Criteria**:
- Bottom navigation works and matches design
- Animations smooth and match React version
- Error states handled gracefully

---

## Testing & Verification

### Visual Regression Testing

After each component/view implementation:

1. **Screenshot Comparison**:
   ```bash
   # Terminal 1: Run React frontend
   cd nishiki-frontend && npm run dev

   # Terminal 2: Run Go frontend
   cd frontend && ./bin/web && ./bin/serve

   # Take screenshots in Firefox of both versions
   # Compare using image diff tool
   ```

2. **Responsive Testing**:
   - Test at mobile viewport (375x667 - iPhone SE)
   - Test at tablet viewport (768x1024)
   - Test at desktop viewport (1920x1080)

3. **State Testing**:
   - Normal state
   - Hover state
   - Active/focused state
   - Disabled state
   - Error state
   - Loading state

### Functional Testing

For each feature:

1. **Happy Path**:
   - Create flow
   - Read/view flow
   - Update flow
   - Delete flow

2. **Edge Cases**:
   - Empty states
   - Network failures
   - Invalid input
   - Concurrent operations

3. **Integration Testing**:
   - Full user workflows (e.g., create group → add members → create collection → add objects)
   - Cross-feature interactions

---

## Cleanup & Deprecation Plan

### Pre-Deprecation Checklist

Before removing `nishiki-frontend/`:

- [ ] All views from nishiki-frontend implemented in frontend
- [ ] All components visually match (verified via screenshots)
- [ ] All API endpoints used
- [ ] All user workflows functional
- [ ] Error handling equivalent or better
- [ ] Loading states implemented
- [ ] Animations implemented (or consciously omitted)
- [ ] Documentation updated

### Deprecation Steps

1. **Create feature parity matrix**:
   ```markdown
   | Feature                  | React | Go  | Notes                    |
   |--------------------------|-------|-----|--------------------------|
   | Groups List              | ✓     | ✓   | Match confirmed          |
   | Groups Detail            | ✓     | ✓   | Match confirmed          |
   | Groups Members           | ✓     | ✓   | Match confirmed          |
   | Collections List         | ✓     | ✓   | Match confirmed          |
   | Collections Detail       | ✓     | ✓   | Match confirmed          |
   | Containers CRUD          | ✓     | ✓   | Match confirmed          |
   | Objects List             | ✓     | ✓   | Match confirmed          |
   | Objects CRUD             | ✓     | ✓   | Match confirmed          |
   | Global Search            | ✓     | ✓   | Match confirmed          |
   | Categories Management    | ✓     | ✓   | Match confirmed          |
   | Bottom Navigation        | ✓     | ✓   | Match confirmed          |
   | Authentication Flow      | ✓     | ✓   | Match confirmed          |
   | Profile View             | ✓     | ✓   | Match confirmed          |
   ```

2. **Archive nishiki-frontend**:
   ```bash
   # Create archive branch
   git checkout -b archive/nishiki-frontend-react
   git push origin archive/nishiki-frontend-react

   # Remove from main branch
   git checkout master
   git rm -rf nishiki-frontend/
   git commit -m "Remove React frontend (archived in archive/nishiki-frontend-react)

   Feature parity achieved with Go/WASM frontend.
   All functionality migrated to frontend/ directory.

   See frontend/PORTING_PLAN.md for migration details."
   ```

3. **Update project documentation**:
   - Update root `README.md` to remove React frontend references
   - Update `CLAUDE.md` to remove nishiki-frontend context
   - Ensure `frontend/CLAUDE.md` is comprehensive

4. **Clean up configuration files**:
   - Remove React-specific configs (if any at root level)
   - Remove `package.json` if only used for nishiki-frontend
   - Clean up Docker compose if React-specific services exist

---

## Code Style Guidelines for Claude Agents

### Style Function Naming

**Pattern**: `Style[Component][Variant][Modifier]`

**Examples**:
- `StyleButtonPrimary` - Primary button variant
- `StyleButtonPrimaryLarge` - Primary button, large size
- `StyleCardHeader` - Card header section
- `StyleTextSectionTitle` - Section title text style
- `StyleInputError` - Input in error state

**Rules**:
- Always use `Style` prefix
- Use PascalCase
- Be specific but concise
- Group related styles together in source file

### Color Usage

**Always use named constants**:
```go
// CORRECT
s.Background = colors.Uniform(ColorPrimary)

// INCORRECT
s.Background = colors.Uniform(color.RGBA{R: 106, G: 179, B: 171, A: 255})
```

### Spacing Usage

**Always use spacing constants**:
```go
// CORRECT
s.Padding.Set(units.Dp(Spacing4), units.Dp(Spacing6))

// INCORRECT
s.Padding.Set(units.Dp(16), units.Dp(24))
```

### Border Radius (Cogent Core v0.3.12)

**MUST use `sides.NewValues()`**:
```go
// CORRECT
import "cogentcore.org/core/styles/sides"
s.Border.Radius = sides.NewValues(units.Dp(RadiusDefault))

// INCORRECT (will not compile)
s.Border.Radius = units.Dp(10)
```

### Component Function Signatures

**Use props pattern for complex components**:
```go
type ButtonProps struct {
    Text     string
    Icon     icons.Icon
    Variant  ButtonVariant
    OnClick  func()
    Disabled bool
}

func RenderButton(parent core.Widget, props ButtonProps) *core.Button {
    // Implementation
}
```

**Use simple signature for trivial components**:
```go
func RenderDivider(parent core.Widget) *core.Frame {
    divider := core.NewFrame(parent)
    divider.Styler(StyleDivider)
    return divider
}
```

### Import Organization

**Standard order**:
```go
import (
    // Standard library
    "fmt"
    "net/http"

    // Third-party
    "cogentcore.org/core/core"
    "cogentcore.org/core/styles"
    "cogentcore.org/core/styles/units"
    "cogentcore.org/core/styles/sides"

    // Internal - domain
    "github.com/yourorg/inventory/frontend/pkg/types"

    // Internal - current package related
    "github.com/yourorg/inventory/frontend/ui/components"
    "github.com/yourorg/inventory/frontend/ui/styles"
)
```

---

## Reference Files Quick Access

When implementing specific features, reference these files from `nishiki-frontend/`:

**Authentication**:
- `src/lib/auth/authentikAuth.ts` - OIDC client logic
- `src/app/auth/callback/page.tsx` - Callback handler
- `src/app/auth/logout/page.tsx` - Logout flow
- `src/app/login/page.tsx` - Login page

**Groups**:
- `src/components/pages/GroupsPage.tsx` - Groups list UI
- `src/components/pages/GroupSinglePage.tsx` - Group detail UI
- `src/components/pages/MembersPage.tsx` - Members list UI
- `src/features/groups/` - Group feature logic
- `src/lib/api/group/client/groupApiClient.client.ts` - Groups API

**Collections & Containers**:
- `src/lib/api/container/client/containerApiClient.client.ts` - Container API
- Container components (search in features)

**Objects/Foods**:
- `src/components/pages/FoodsPage.tsx` - Objects list UI
- `src/features/foods/` - Foods feature logic
- `src/lib/api/food/client/foodApiClient.client.ts` - Foods API

**UI Components (shadcn/ui)**:
- `src/components/ui/Button.tsx` - Button component
- `src/components/ui/Card.tsx` - Card component
- `src/components/ui/Dialog.tsx` - Dialog component
- `src/components/ui/Drawer.tsx` - Drawer component
- `src/components/ui/Form.tsx` - Form component
- `src/components/ui/Input/` - Input variants
- `src/components/ui/Select/` - Select variants

**Layouts**:
- `src/components/layouts/MobileLayout.tsx` - Main layout
- `src/components/parts/Header.tsx` - Header component
- `src/components/parts/BottomMenu.tsx` - Bottom navigation
- `src/components/parts/BottomMenuLink.tsx` - Nav link component

**Styling**:
- `tailwind.config.ts` - Design token source of truth
- `src/styles/globals.css` - CSS custom properties

---

## Success Criteria

The porting is complete when:

1. **Visual Parity**: Screenshots of Go frontend and React frontend in Firefox are pixel-identical
2. **Functional Parity**: All user workflows from React version work in Go version
3. **Code Quality**: Go code follows clean architecture, no monolithic files
4. **Documentation**: All features documented, component usage examples provided
5. **Testing**: Critical paths tested, edge cases handled
6. **Performance**: WASM build loads quickly, UI is responsive
7. **Maintainability**: Future developers can easily add features following established patterns

At this point, `nishiki-frontend/` can be safely archived and removed from the active codebase.
