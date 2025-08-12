# AI Agent Guide: Migrating React TypeScript Apps to Cogent Core

## Overview for AI Agents

This guide helps AI agents systematically migrate React TypeScript applications to Cogent Core (Go). The migration involves translating component-based UI patterns to Cogent Core's Plan-based system while preserving application logic and state management.

## Pre-Migration Checklist

```yaml
Requirements:
  - Go 1.21+ installed
  - Cogent Core: go get cogentcore.org/core@latest
  - Tool: go install cogentcore.org/core/cmd/core@latest
  
Source Analysis:
  - Map all React components to identify hierarchy
  - Document state management pattern (Redux/Context/Zustand)
  - List all API endpoints and data models
  - Identify routing structure
  - Note any special libraries (charts, maps, etc.)
```

## Core Concept Mappings

### React → Cogent Core Translation Table

| React Concept | Cogent Core Equivalent | Migration Notes |
|--------------|------------------------|-----------------|
| `Component` | `type MyWidget struct` | Embed `core.Frame` or specific widget |
| `useState` | Struct fields + `Update()` | State lives in struct fields |
| `useEffect` | `OnShow()` or goroutines | Lifecycle managed differently |
| `props` | Struct fields or method params | Pass data via struct initialization |
| `JSX` | `MakeUI(*tree.Plan)` | Declarative UI in Go code |
| `Context API` | Embedded structs or interfaces | Share state via composition |
| `Redux/Zustand` | Single app struct | Centralized state in main struct |
| `React Router` | `core.Tabs` or `core.Frame` switching | Navigation via widget updates |
| `CSS Modules` | `styles.Style` | Styling via Go structs |
| `onClick` | `OnClick(func(e events.Event))` | Event handlers as Go functions |

## Step-by-Step Migration Process

### Step 1: Analyze React Component Structure

```typescript
// React Component Example
interface TodoProps {
  items: TodoItem[]
  onAdd: (item: TodoItem) => void
}

const TodoList: React.FC<TodoProps> = ({ items, onAdd }) => {
  const [newItem, setNewItem] = useState("")
  
  return (
    <div className="todo-container">
      <input 
        value={newItem}
        onChange={(e) => setNewItem(e.target.value)}
      />
      <button onClick={() => onAdd({text: newItem})}>Add</button>
      {items.map(item => (
        <TodoItem key={item.id} item={item} />
      ))}
    </div>
  )
}
```

### Step 2: Create Cogent Core Structure

```go
// Cogent Core Equivalent
package main

import (
    "cogentcore.org/core/core"
    "cogentcore.org/core/events"
    "cogentcore.org/core/tree"
)

type TodoList struct {
    core.Frame
    Items    []TodoItem
    NewItem  string
}

func (t *TodoList) MakeUI(p *tree.Plan) {
    tree.AddChild(p, func(w *core.Frame) {
        w.Style.Direction = styles.Column
        
        // Input field
        tree.AddChild(p, func(in *core.TextField) {
            in.SetText(t.NewItem)
            in.OnChange(func(e events.Event) {
                t.NewItem = in.Text()
            })
        })
        
        // Add button
        tree.AddChild(p, func(btn *core.Button) {
            btn.SetText("Add")
            btn.OnClick(func(e events.Event) {
                t.Items = append(t.Items, TodoItem{Text: t.NewItem})
                t.NewItem = ""
                t.Update()
            })
        })
        
        // Items list
        for _, item := range t.Items {
            tree.AddChild(p, func(w *TodoItemWidget) {
                w.Item = item
            })
        }
    })
}
```

## Common Pattern Translations

### 1. State Management

```typescript
// React: Multiple useState hooks
const [user, setUser] = useState<User | null>(null)
const [loading, setLoading] = useState(false)
const [error, setError] = useState<string>("")
```

```go
// Cogent Core: Struct fields
type AppState struct {
    core.Frame
    User    *User
    Loading bool
    Error   string
}

// Update UI after state change
func (a *AppState) SetUser(user *User) {
    a.User = user
    a.Update() // Triggers UI refresh
}
```

### 2. API Calls & Effects

```typescript
// React: useEffect for API calls
useEffect(() => {
    fetch('/api/users')
        .then(res => res.json())
        .then(data => setUsers(data))
}, [])
```

```go
// Cogent Core: Use OnShow or init method
func (a *App) OnShow() {
    go func() {
        users, err := fetchUsers()
        if err != nil {
            a.Error = err.Error()
        } else {
            a.Users = users
        }
        a.AsyncUpdate() // Thread-safe update
    }()
}

func fetchUsers() ([]User, error) {
    resp, err := http.Get("/api/users")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var users []User
    json.NewDecoder(resp.Body).Decode(&users)
    return users, nil
}
```

### 3. Conditional Rendering

```typescript
// React: Conditional JSX
return (
    <>
        {loading && <Spinner />}
        {error && <ErrorMessage text={error} />}
        {user && <UserProfile user={user} />}
    </>
)
```

```go
// Cogent Core: Conditional in MakeUI
func (a *App) MakeUI(p *tree.Plan) {
    if a.Loading {
        tree.AddChild(p, func(w *core.Spinner) {})
        return
    }
    
    if a.Error != "" {
        tree.AddChild(p, func(w *core.Text) {
            w.SetText(a.Error)
            w.Style.Color = colors.Scheme.Error.Base
        })
    }
    
    if a.User != nil {
        tree.AddChild(p, func(w *UserProfile) {
            w.User = a.User
        })
    }
}
```

### 4. Forms & Input Handling

```typescript
// React: Controlled form
const [formData, setFormData] = useState({
    name: '',
    email: ''
})

const handleSubmit = (e: FormEvent) => {
    e.preventDefault()
    api.submitForm(formData)
}
```

```go
// Cogent Core: Form handling
type FormData struct {
    Name  string
    Email string
}

type FormWidget struct {
    core.Frame
    Data FormData
}

func (f *FormWidget) MakeUI(p *tree.Plan) {
    tree.AddChild(p, func(w *core.Form) {
        w.Struct = &f.Data
        w.OnChange(func(e events.Event) {
            // Auto-updates f.Data fields
        })
    })
    
    tree.AddChild(p, func(btn *core.Button) {
        btn.SetText("Submit")
        btn.OnClick(func(e events.Event) {
            f.submitForm()
        })
    })
}
```

### 5. Lists & Iteration

```typescript
// React: Mapping arrays
items.map((item, index) => (
    <Card key={item.id}>
        <h3>{item.title}</h3>
        <p>{item.description}</p>
    </Card>
))
```

```go
// Cogent Core: Range loops
func (a *App) MakeUI(p *tree.Plan) {
    for i, item := range a.Items {
        item := item // Capture for closure
        tree.AddChild(p, func(w *core.Frame) {
            w.Style.Border.Radius = styles.BorderRadiusMedium
            
            tree.AddChild(p, func(title *core.Text) {
                title.SetText(item.Title)
                title.Style.Font.Weight = styles.WeightBold
            })
            
            tree.AddChild(p, func(desc *core.Text) {
                desc.SetText(item.Description)
            })
        })
    }
}
```

## Migration Patterns for Common Libraries

### React Router → Cogent Core Navigation

```go
// Instead of routes, use widget switching
type App struct {
    core.Frame
    CurrentView string
}

func (a *App) MakeUI(p *tree.Plan) {
    // Navigation bar
    tree.AddChild(p, func(w *core.Toolbar) {
        w.Maker(func(p *tree.Plan) {
            tree.AddChild(p, func(btn *core.Button) {
                btn.SetText("Home").OnClick(func(e events.Event) {
                    a.CurrentView = "home"
                    a.Update()
                })
            })
        })
    })
    
    // View switching
    switch a.CurrentView {
    case "home":
        tree.AddChild(p, func(w *HomeView) {})
    case "profile":
        tree.AddChild(p, func(w *ProfileView) {})
    }
}
```

### Material-UI → Cogent Core Material Design

```go
// Cogent Core has built-in Material Design 3
tree.AddChild(p, func(w *core.Button) {
    w.SetType(core.ButtonAction) // FAB style
    w.SetIcon(icons.Add)
    w.Style.Background = colors.Scheme.Primary.Base
})
```

### Axios/Fetch → Go HTTP Client

```go
// Create reusable API client
type APIClient struct {
    BaseURL string
    Client  *http.Client
}

func (c *APIClient) Get(path string, result interface{}) error {
    resp, err := c.Client.Get(c.BaseURL + path)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    return json.NewDecoder(resp.Body).Decode(result)
}
```

## File Structure Migration

```
React Project:                 Cogent Core Project:
src/                       →   /
  components/              →   widgets/
    TodoList.tsx           →     todo_list.go
    TodoItem.tsx           →     todo_item.go
  hooks/                   →   (embedded in widgets)
    useAuth.ts             →     auth.go
  services/               →   services/
    api.ts                →     api.go
  store/                  →   models/
    userSlice.ts          →     user.go
  App.tsx                 →   main.go
  index.tsx               →   main.go
```

## Testing Migration

```go
// Cogent Core testing
func TestTodoList(t *testing.T) {
    b := core.NewBody()
    
    todo := &TodoList{}
    b.AddChild(todo)
    
    b.RunMainWindow()
    
    // Simulate user interaction
    todo.NewItem = "Test Item"
    todo.AddItem()
    
    assert.Equal(t, 1, len(todo.Items))
}
```

## Deployment Configuration

```yaml
# .core.toml - Cogent Core config
[build]
name = "MyApp"
description = "Migrated from React"
icon = "icon.svg"

[build.web]
port = 8080
domain = "myapp.com"

[build.desktop]
width = 1200
height = 800
```

## Common Gotchas & Solutions

### 1. Async State Updates
**Issue**: Updating UI from goroutines
```go
// Wrong: Direct update from goroutine
go func() {
    data := fetchData()
    a.Data = data // Race condition!
    a.Update()
}()

// Correct: Use AsyncUpdate
go func() {
    data := fetchData()
    a.AsyncUpdate(func() {
        a.Data = data
    })
}()
```

### 2. Component Keys
**Issue**: React uses keys for list items
```go
// Cogent Core: Use stable references
type ItemWidget struct {
    core.Frame
    ID   string // Maintain ID for updates
    Item Item
}
```

### 3. CSS-in-JS
**Issue**: Styled-components or emotion
```go
// Cogent Core: Use styles.Style
func (w *MyWidget) Init() {
    w.Style.Background = colors.Scheme.Surface
    w.Style.Padding.Set(units.Dp(16))
    w.Style.Border.Radius = styles.BorderRadiusMedium
}
```

## Quick Reference Checklist

When migrating each React component:

- [ ] Convert TypeScript interfaces to Go structs
- [ ] Replace useState with struct fields
- [ ] Convert useEffect to OnShow() or init methods
- [ ] Translate JSX to MakeUI() method
- [ ] Replace event handlers with OnEvent methods
- [ ] Convert CSS to styles.Style configurations
- [ ] Update API calls to use Go's http package
- [ ] Replace React Router with view switching logic
- [ ] Convert props to struct embedding or fields
- [ ] Migrate tests to Cogent Core testing patterns

## Final Build Commands

```bash
# Development
core run .

# Build for web (WASM)
core build -o wasm .

# Build for desktop
core build -o darwin .  # macOS
core build -o windows . # Windows
core build -o linux .   # Linux

# Build for Android
core build -o android .
```

## Resources for AI Agents

- **Cogent Core Docs**: https://cogentcore.org/core/
- **Widget Examples**: Check docs for interactive examples
- **Migration Priority**: Start with leaf components, work up to root
- **State Management**: Prefer single app-level state struct
- **Performance**: Use `AsyncUpdate()` for heavy operations
- **Debugging**: Use `core.ProfileToggle()` for performance analysis