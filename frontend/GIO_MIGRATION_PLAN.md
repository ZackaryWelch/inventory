# Plan: Migrate Frontend from Cogent Core to Gio

## Overview
Migrate the Nishiki inventory management frontend from Cogent Core v0.3.12 to Gio UI framework for improved performance, smaller WASM binary size, and more active ecosystem.

## Phase 1: Project Setup & Dependencies
1. **Update go.mod dependencies**
   - Remove `cogentcore.org/core v0.3.12`
   - Add `gioui.org@latest` and `gioui.org/x@latest` (for extended widgets)
   - Keep existing dependencies (oauth2, viper, API clients)

2. **Update build tooling**
   - Modify `cmd/web/main.go` for Gio WASM compilation
   - Update `bin/web` script to use Gio's build flags: `-target js -tags nowayland`
   - Update `cmd/webmain/main.go` for Gio's window/app lifecycle
   - Create new index.html for Gio (uses different WASM loader than Cogent Core)

## Phase 2: Core Architecture Migration
3. **App structure refactoring** (`app/app.go`)
   - Replace `*core.Body` with `*app.Window` (Gio's window type)
   - Replace `*core.Frame` containers with `layout.Context` patterns
   - Migrate from retained-mode (Cogent Core) to immediate-mode (Gio) UI paradigm
   - Add `*material.Theme` for Material Design components
   - Implement event loop with `app.Main()` and window event handling

4. **State management redesign**
   - Implement invalidation system (Gio requires explicit window redraws)
   - Add `window.Invalidate()` calls after state changes
   - Consider adding operation channel for async updates

## Phase 3: Styling System Migration
5. **Replace centralized styling** (`ui/styles/`)
   - Migrate `tokens.go` colors to Gio color.NRGBA values
   - Convert `components.go` style functions to Gio widget configuration
   - Replace Cogent Core's `Styler` pattern with Gio's widget parameters
   - Implement custom drawing for unique visual elements

6. **Theming approach**
   - Use `material.Theme` from `gioui.org/widget/material`
   - Create theme customization layer for brand colors
   - Implement dark mode support with theme switching

## Phase 4: Widget Migration
7. **Map Cogent Core widgets to Gio equivalents**
   - `core.Body` → `app.Window` with event loop
   - `core.Frame` → `layout.Flex`, `layout.Stack`, or custom layouts
   - `core.Button` → `material.Button` with `widget.Clickable` state
   - `core.Text` → `material.Label` or `widget.Label`
   - `core.TextField` → `material.Editor` with `widget.Editor` state
   - `core.List` → `layout.List` with custom item rendering
   - Dialogs → Modal overlays with `layout.Stack`

8. **Event handling migration**
   - Replace `OnClick(func(e events.Event))` with `widget.Clickable` state checks
   - Migrate from event callbacks to immediate-mode event checking
   - Update pointer, key, and touch event handling

## Phase 5: View Layer Migration
9. **Migrate UI views** (migrate in order of complexity)
   - **Login view** (`app_methods.go:showLoginView`)
   - **Dashboard view** (`app_methods.go:showDashboardView`)
   - **Collections view** (`collections_ui.go`)
   - **Objects view** (`objects_ui.go`)
   - **Containers view** (`containers_ui.go`)
   - **Import view** (`import_ui.go`)
   - **Groups management** (`ui_management.go`)

10. **Migrate UI helpers** (`ui_helpers.go`)
    - Rewrite `showDialog()` for Gio modal system
    - Recreate `createTextField()`, `createSearchField()`, etc. as Gio layout functions
    - Implement `createSectionHeader()`, `createFlexRow()` with Gio layouts

11. **Migrate reusable components** (`ui/components/`)
    - `button.go` → Material button variants
    - `card.go` → Custom card layout with Gio primitives
    - `badge.go` → Custom badge widget
    - `icon.go` → Use Gio's icon system or embed SVGs

## Phase 6: Layout System Migration
12. **Migrate layouts** (`ui/layouts/`)
    - `header.go` → Gio flex layout with app bar
    - `mobile_layout.go` → Responsive layout with Gio constraints
    - `bottom_menu.go` → Bottom navigation bar with Gio

13. **Implement responsive design**
    - Use `gtx.Constraints` for responsive breakpoints
    - Create mobile/desktop layout switching based on viewport size
    - Ensure touch-friendly sizing on mobile

## Phase 7: Integration & Platform-Specific Code
14. **Authentication flow** (`auth_service.go`)
    - Update OAuth redirect handling for Gio
    - Verify `syscall/js` compatibility (should work unchanged)
    - Test token storage and retrieval in browser localStorage

15. **Platform-specific builds**
    - Keep `//go:build js && wasm` tags for WebAssembly
    - Create `//go:build !wasm` for future desktop builds
    - Ensure config loading works across platforms

## Phase 8: Testing & Optimization
16. **Testing strategy**
    - Unit test layout functions
    - Integration test API client interactions
    - Manual testing of all user flows in browser
    - Test on mobile browsers (iOS Safari, Chrome Mobile)

17. **Performance optimization**
    - Profile WASM binary size (expect 50-70% reduction vs Cogent Core)
    - Optimize render performance with layout caching
    - Minimize memory allocations in render loop
    - Use `op.Defer()` for expensive operations

## Phase 9: Documentation & Deployment
18. **Update documentation**
    - Update `CLAUDE.md` with Gio-specific patterns
    - Document Gio immediate-mode UI patterns
    - Update build instructions
    - Create migration notes for future developers

19. **Deployment preparation**
    - Update nginx/server config if needed (likely unchanged)
    - Test WASM loading and execution
    - Verify all API endpoints work with new frontend
    - Create rollback plan

## Key Differences: Cogent Core vs Gio

### Paradigm Shift
- **Cogent Core**: Retained-mode (create widgets once, update via methods)
- **Gio**: Immediate-mode (rebuild UI every frame based on state)

### Code Pattern Changes
```go
// BEFORE (Cogent Core)
button := core.NewButton(parent).SetText("Click")
button.OnClick(func(e events.Event) {
    // Handle click
})

// AFTER (Gio)
var buttonState widget.Clickable
if material.Button(th, &buttonState, "Click").Layout(gtx) {
    // Handle click
}
```

### Benefits of Migration
- **Smaller binaries**: Gio WASM ~2-3MB vs Cogent Core ~8-10MB
- **Better performance**: Immediate-mode rendering is faster for frequent updates
- **Active ecosystem**: Gio has larger community and more examples
- **Mobile-first**: Better touch support and mobile rendering
- **Future-proof**: Easier path to native mobile (Android/iOS) apps

### Challenges
- **Learning curve**: Immediate-mode requires different mental model
- **More manual work**: Less built-in widgets, more custom drawing
- **State management**: Need explicit invalidation and redraw logic
- **No built-in router**: Must implement view navigation manually

## Estimated Effort
- **Phase 1-2**: 2-3 days (setup, architecture)
- **Phase 3-4**: 3-4 days (styling, widgets)
- **Phase 5**: 5-7 days (view migration)
- **Phase 6**: 2-3 days (layouts)
- **Phase 7-9**: 3-4 days (integration, testing, docs)
- **Total**: ~15-21 days for complete migration

## Risk Mitigation
- Create feature branch for migration
- Migrate incrementally (one view at a time)
- Keep old code for reference during migration
- Test thoroughly on mobile browsers
- Have backend team on standby for API issues

## Architecture Comparison

### Current Structure (Cogent Core)
```
app/
  app.go                    - App struct with *core.Body, *core.Frame
  app_methods.go            - View rendering functions
  ui_helpers.go             - Dialog/form helpers using core widgets

ui/
  styles/                   - Centralized Styler functions
  components/               - Reusable Cogent Core widgets
  layouts/                  - Layout components
```

### Proposed Structure (Gio)
```
app/
  app.go                    - App struct with *app.Window, *material.Theme
  render.go                 - Main render loop (immediate-mode)
  views/
    login.go                - Login view layout function
    dashboard.go            - Dashboard view layout function
    collections.go          - Collections view layout function
    objects.go              - Objects view layout function
    containers.go           - Containers view layout function

ui/
  theme/                    - Color palette, typography (material.Theme customization)
  widgets/                  - Custom Gio widgets (card, badge, etc.)
  layouts/                  - Layout helper functions

state/
  widgets.go                - Widget state structs (clickables, editors, etc.)
```

## Additional Considerations

### State Management
Gio's immediate-mode paradigm requires all widget state to be stored explicitly:
- `widget.Clickable` for buttons
- `widget.Editor` for text fields
- `widget.Bool` for checkboxes
- Custom state structs for complex components

Store these in the App struct or view-specific state structs.

### Navigation
Implement view routing manually:
```go
type ViewID int

const (
    ViewLogin ViewID = iota
    ViewDashboard
    ViewCollections
    // ...
)

type App struct {
    currentView ViewID
    // ...
}

func (app *App) render(gtx layout.Context) {
    switch app.currentView {
    case ViewLogin:
        app.renderLoginView(gtx)
    case ViewDashboard:
        app.renderDashboardView(gtx)
    // ...
    }
}
```

### Async Operations
Handle API calls and background tasks with channels:
```go
type Operation struct {
    Type string
    Data interface{}
}

type App struct {
    ops chan Operation
    // ...
}

// In goroutine
go func() {
    data, err := api.FetchData()
    app.ops <- Operation{Type: "data_loaded", Data: data}
    window.Invalidate()
}()

// In event loop
for {
    select {
    case e := <-window.Events():
        // Handle window events
    case op := <-app.ops:
        // Handle async operations
    }
}
```

## Resources
- Gio documentation: https://gioui.org/doc
- Gio examples: https://git.sr.ht/~eliasnaur/gio-example
- Material Design widgets: https://gioui.org/widget/material
- Community chat: https://gophers.slack.com #gio
