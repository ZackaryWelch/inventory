# MCP Elicitation: Interactive Tool Input for Claude

Elicitation allows MCP servers to request structured user input mid-task via an interactive dialog, rather than returning errors or making assumptions.

## What Is Elicitation?

**Traditional tool flow:**
```
User invokes tool → MCP processes → returns result or error
```

**With elicitation:**
```
User invokes tool → MCP detects missing/ambiguous info → Claude shows interactive form → 
user fills fields → MCP receives structured ElicitationResult → completes tool
```

This creates a conversation-like experience where tools can ask clarifying questions inline.

---

## Core Concepts

### When to Elicit vs. When Not To

**Use elicitation when:**
- The user's intent is clear, but a required field is missing (e.g., expiry date for a perishable item)
- Multiple interpretations exist and the user must disambiguate (e.g., "which kitchen container?")
- User confirmation is needed for a potentially destructive action
- The tool can't reasonably infer the value from context

**Don't elicit when:**
- The MCP server can resolve the value via API calls (fetch related resources and infer)
- The information is already in the conversation context
- Elicitation would be blocking on every use (poor UX—fallback to defaults or ask Claude instead)
- The field has a sensible default

### Elicitation Lifecycle

1. **Detection**: Tool handler detects missing/ambiguous info
2. **Elicitation Request**: Server returns `Elicitation` object with form schema
3. **User Input**: Claude presents interactive dialog; user fills and submits
4. **Result Handling**: MCP server receives `ElicitationResult` with structured data
5. **Tool Completion**: Tool runs with complete inputs; returns final result

---

## Implementation Pattern

### Step 1: Define Elicitation Schemas

An elicitation request specifies what fields to ask for:

```go
type Elicitation struct {
    ElicitationID string      // Unique ID for this elicitation
    Prompt        string      // Question shown to user
    Fields        []Field     // Form fields to populate
}

type Field struct {
    Name        string       // Field key (must match your tool's input field)
    Type        string       // "text", "date", "number", "select", "multiselect"
    Label       string       // Display label shown to user
    Required    bool         // Is this field required?
    Description string       // Help text
    Options     []string     // For "select"/"multiselect"
    Default     interface{}  // Default value
}
```

### Step 2: Modify Tool Handlers to Detect When Elicitation Is Needed

In your tool handler, check for missing or ambiguous inputs early:

```go
func (s *Server) handleCreateObject(params map[string]interface{}) (interface{}, error) {
    // Parse the request
    var req CreateObjectRequest
    // ... unmarshal params into req ...

    // Check if this is a follow-up to an elicitation
    if req.ElicitationResult != nil {
        // User filled the form; use those values
        if expiryStr, ok := req.ElicitationResult["expires_at"]; ok {
            req.ExpiresAt = expiryStr
        }
        // Continue to normal processing with complete data
    } else if req.ExpiresAt == "" {
        // First call, and expiry is missing—elicit it
        return &Elicitation{
            ElicitationID: "create_object_expiry",
            Prompt: "This item appears to be perishable. When does it expire?",
            Fields: []Field{
                {
                    Name:        "expires_at",
                    Type:        "date",
                    Label:       "Expiration Date",
                    Required:    true,
                    Description: "Leave blank if this item doesn't expire",
                },
            },
        }, nil
    }

    // Normal flow: call backend with complete data
    return s.createObjectInBackend(req)
}
```

### Step 3: Handle ElicitationResult Hook

The server-level `ElicitationResult` hook intercepts elicitation responses:

```go
server.OnElicitationResult(func(ctx context.Context, result *ElicitationResult) error {
    // result.ElicitationID tells you which elicitation this responds to
    // result.Values is the user's form submission
    // Re-invoke the original tool with the populated data
    return nil
})
```

---

## Design Principles

### 1. Keep Elicitations Focused

Ask for one logical grouping of related fields, not a dozen separate questions.

**Good:**
```
Prompt: "Help me confirm the receipt items"
Fields: [item_name, quantity, unit, expiry_date]
```

**Bad:**
```
Prompt: "Add more info"
Fields: [name, description, color, size, weight, location, category, notes, ...]
```

### 2. Provide Helpful Defaults and Suggestions

When possible, pre-fill fields with reasonable values:

```go
Fields: []Field{
    {
        Name:    "container_id",
        Type:    "select",
        Label:   "Which container?",
        Options: containerNames,  // Pre-computed from API
        Default: mostRecentContainer, // Smart default
    },
}
```

### 3. Use Contextual Help Text

The `Description` field should guide users:

```go
{
    Name:        "expires_at",
    Type:        "date",
    Label:       "Expiration Date",
    Description: "Format: YYYY-MM-DD. Leave blank if non-perishable.",
}
```

### 4. Fail Gracefully If Elicitation Is Rejected

If the user cancels the elicitation dialog, the MCP server should handle it:

```go
if req.ElicitationCancelled {
    return &ErrorResponse{
        Message: "Cancelled. Please provide the missing information in your next message.",
    }, nil
}
```

---

## Nishiki-Specific Use Cases

### Use Case 1: Ambiguous Container Selection

**Scenario:** User says "Add milk to the fridge" but there are multiple fridges (kitchen, garage).

```go
func (s *Server) handleCreateObject(params map[string]interface{}) (interface{}, error) {
    var req CreateObjectRequest
    // ... parse params ...

    // If container_id is ambiguous, elicit
    candidates := s.findContainersByName(req.ContainerName)
    if len(candidates) > 1 {
        return &Elicitation{
            ElicitationID: fmt.Sprintf("resolve_container_%s", req.ContainerName),
            Prompt:        fmt.Sprintf("Found %d containers matching '%s'. Which one?", 
                           len(candidates), req.ContainerName),
            Fields: []Field{
                {
                    Name:     "container_id",
                    Type:     "select",
                    Label:    "Container",
                    Required: true,
                    Options:  formatContainerOptions(candidates),
                    Default:  candidates[0].ID, // Most recently used
                },
            },
        }, nil
    }

    if len(candidates) == 0 {
        // No match—ask the user to create or choose from all containers
        allContainers := s.listAllContainers()
        return &Elicitation{
            ElicitationID: "choose_container_from_all",
            Prompt:        fmt.Sprintf("No container named '%s'. Choose from existing:", req.ContainerName),
            Fields: []Field{
                {
                    Name:     "container_id",
                    Type:     "select",
                    Label:    "Container",
                    Options:  formatContainerOptions(allContainers),
                },
            },
        }, nil
    }

    // Exactly one match—proceed
    req.ContainerID = candidates[0].ID
    return s.createObjectInBackend(req)
}

func formatContainerOptions(containers []Container) []string {
    var opts []string
    for _, c := range containers {
        opts = append(opts, fmt.Sprintf("%s (in %s)", c.Name, c.ParentCollection))
    }
    return opts
}
```

### Use Case 2: Expiry Date Confirmation for Perishables

**Scenario:** User adds milk without specifying an expiry date. The MCP infers it's likely perishable.

```go
func (s *Server) handleCreateObject(params map[string]interface{}) (interface{}, error) {
    var req CreateObjectRequest
    // ... parse params ...

    // If expiry_at is missing and this looks perishable, elicit
    if req.ExpiresAt == "" && s.isProbablyPerishable(req.ObjectName) {
        return &Elicitation{
            ElicitationID: "confirm_expiry",
            Prompt:        fmt.Sprintf("'%s' looks perishable. When does it expire?", req.ObjectName),
            Fields: []Field{
                {
                    Name:        "expires_at",
                    Type:        "date",
                    Label:       "Expiration Date",
                    Required:    false,
                    Description: "Leave blank if it doesn't expire or you're unsure",
                },
            },
        }, nil
    }

    return s.createObjectInBackend(req)
}

func (s *Server) isProbablyPerishable(name string) bool {
    perishableKeywords := []string{"milk", "yogurt", "bread", "cheese", "meat", "eggs"}
    lower := strings.ToLower(name)
    for _, kw := range perishableKeywords {
        if strings.Contains(lower, kw) {
            return true
        }
    }
    return false
}
```

### Use Case 3: Bulk Import Item Confirmation

**Scenario:** User provides a receipt. The MCP parses it into structured items and asks for confirmation before importing.

```go
func (s *Server) handleBulkImport(params map[string]interface{}) (interface{}, error) {
    var req BulkImportRequest
    // ... parse params ...

    // If this is the initial request, parse and elicit confirmation
    if req.ElicitationResult == nil {
        items := s.parseReceiptDescription(req.Description)

        // Build confirmable items
        var itemFields []string
        for i, item := range items {
            itemFields = append(itemFields, 
                fmt.Sprintf("%s | qty: %v %s | expires: %s",
                    item.Name, item.Quantity, item.Unit, item.ExpiresAt))
        }

        return &Elicitation{
            ElicitationID: "confirm_bulk_items",
            Prompt:        "Review parsed items before importing:",
            Fields: []Field{
                {
                    Name:        "confirmed_items",
                    Type:        "multiselect",
                    Label:       "Items to Import",
                    Required:    true,
                    Options:     itemFields,
                    Default:     itemFields, // All checked by default
                    Description: "Uncheck any items you want to skip",
                },
                {
                    Name:        "target_container",
                    Type:        "select",
                    Label:       "Import to Container",
                    Required:    true,
                    Options:     s.listAllContainerNames(),
                    Description: "All items go to this container",
                },
            },
        }, nil
    }

    // User confirmed—extract selections and proceed
    selectedIndices := req.ElicitationResult["confirmed_items"].([]int)
    targetContainer := req.ElicitationResult["target_container"].(string)

    items := s.parseReceiptDescription(req.Description)
    filteredItems := make([]Item, 0)
    for _, idx := range selectedIndices {
        filteredItems = append(filteredItems, items[idx])
    }

    return s.importItemsToContainer(filteredItems, targetContainer)
}
```

### Use Case 4: Destructive Action Confirmation

**Scenario:** User wants to delete a collection. Elicit confirmation to prevent accidents.

```go
func (s *Server) handleDeleteCollection(params map[string]interface{}) (interface{}, error) {
    var req DeleteCollectionRequest
    // ... parse params ...

    // If not yet confirmed, elicit
    if req.ElicitationResult == nil {
        collection := s.getCollection(req.CollectionID)
        itemCount := s.countObjectsInCollection(req.CollectionID)

        return &Elicitation{
            ElicitationID: fmt.Sprintf("confirm_delete_collection_%s", req.CollectionID),
            Prompt:        fmt.Sprintf("Delete collection '%s' with %d items? This cannot be undone.", 
                           collection.Name, itemCount),
            Fields: []Field{
                {
                    Name:        "confirm_delete",
                    Type:        "select",
                    Label:       "Confirm Deletion",
                    Required:    true,
                    Options:     []string{"No, cancel", "Yes, delete permanently"},
                    Default:     "No, cancel",
                },
            },
        }, nil
    }

    // Check confirmation
    if req.ElicitationResult["confirm_delete"] != "Yes, delete permanently" {
        return &ToolResponse{
            Content: "Deletion cancelled.",
        }, nil
    }

    return s.deleteCollectionInBackend(req.CollectionID)
}
```

### Use Case 5: Collection Selection for Sharing

**Scenario:** User wants to "share my book collection with the family group" but has multiple collections.

```go
func (s *Server) handleShareCollection(params map[string]interface{}) (interface{}, error) {
    var req ShareCollectionRequest
    // ... parse params ...

    // If collection is ambiguous, elicit
    candidates := s.findCollectionsByType(req.ObjectType) // "books"
    if len(candidates) > 1 {
        return &Elicitation{
            ElicitationID: "choose_collection_to_share",
            Prompt:        fmt.Sprintf("You have %d book collections. Which one to share?", len(candidates)),
            Fields: []Field{
                {
                    Name:     "collection_id",
                    Type:     "select",
                    Label:    "Collection",
                    Required: true,
                    Options:  formatCollectionOptions(candidates),
                },
            },
        }, nil
    }

    if len(candidates) == 1 {
        req.CollectionID = candidates[0].ID
    }

    return s.shareCollectionWithGroup(req)
}
```

---

## UX Best Practices for Nishiki

### 1. Minimize Friction in Common Paths

For "add milk to the fridge," don't elicit:
- The fridge (resolve via fuzzy matching + most-recent logic)
- The quantity unit (infer from item name)

Only elicit expiry if truly unclear. Consider smart defaults:

```go
// Instead of always eliciting, try:
if req.ContainerName == "fridge" && !req.HasExpiryInfo {
    // Default to 14 days for dairy-like items
    req.ExpiresAt = time.Now().AddDate(0, 0, 14)
}
```

### 2. Combine Related Decisions

Don't ask:
- "Which container?" → then "Which collection?"

Ask once with pre-filtered options:
```
"Where does this go?"
- Kitchen Fridge (in Food Collection)
- Workshop Cabinet (in Tools Collection)
```

### 3. Progressive Disclosure

For complex bulk operations, elicit in phases:

**Phase 1:** "Parse receipt? Here are the items I found: ..."
**Phase 2:** "Any edits? (optional elicitation for corrections)"
**Phase 3:** "Which containers for each item type?"

### 4. Offer Context in Prompts

```go
// Good
Prompt: "Found 3 fridges. Which one?"
Options: [
    "Kitchen Fridge (79% full, 1 expiring item)",
    "Garage Fridge (23% full)",
    "Workshop Fridge (empty)",
]

// Bad
Prompt: "Choose a container"
Options: ["fridge_001", "fridge_002", "fridge_003"]
```

---

## Error Handling

### Timeout or Cancellation

If the user cancels the elicitation dialog:

```go
if req.ElicitationCancelled {
    return &ToolResponse{
        IsError: true,
        Content: "Elicitation cancelled. Please provide the missing info in your next message, or try a simpler request.",
    }, nil
}
```

### Invalid Submission

If the user's form submission fails validation:

```go
if !isValidDate(req.ElicitationResult["expires_at"]) {
    return &Elicitation{
        ElicitationID: "confirm_expiry", // Same ID—re-ask
        Prompt:        "That date format didn't work. Try again:",
        Fields:        /* reset with error note */,
    }, nil
}
```

---

## Testing Elicitations

### In MCP Inspector

1. Start your MCP server in debug mode
2. Open MCP Inspector: `npx @modelcontextprotocol/inspector`
3. Connect to your server
4. Invoke a tool that triggers elicitation
5. Verify the `Elicitation` object structure matches the spec
6. Simulate user submission by constructing an `ElicitationResult`

### On claude-desktop

1. Add your server to `claude_desktop_config.json`
2. Restart Claude
3. Invoke a tool naturally ("add milk to the fridge")
4. Verify the interactive dialog renders
5. Fill the form and confirm

---

## Common Pitfalls

| Pitfall | Why It's Bad | Fix |
|---------|-------------|-----|
| Eliciting for every optional field | Creates dialog fatigue; breaks flow | Only elicit when truly ambiguous or for validation |
| Asking yes/no via multiselect | Confusing UX | Use "select" with two options |
| ElicitationID changes between attempts | Claude may not recognize follow-up | Keep same ID; increment retry counter if needed |
| No timeout or fallback | User may get stuck waiting | Provide sensible defaults or suggest alternatives |
| Elicitation message too long | Overwhelming form | Keep prompt to 1-2 sentences; use Description for help |
| Forgetting to handle ElicitationResult | Tool re-runs without the new data | Always check for ElicitationResult before eliciting again |

---

## Summary Checklist

- [ ] Identified scenarios where elicitation improves UX (ambiguous input, missing required fields, confirmation)
- [ ] Defined elicitation schemas (ElicitationID, prompt, fields with types and defaults)
- [ ] Modified tool handlers to detect elicitation triggers early
- [ ] Implemented ElicitationResult handling to merge user inputs with tool logic
- [ ] Set up context-aware defaults and options (pre-fetch related resources)
- [ ] Added error handling for cancellation, invalid input, and timeout
- [ ] Tested in MCP Inspector and claude-desktop
- [ ] Documented elicitation flows in tool descriptions
- [ ] Minimized friction in common paths (don't over-elicit)
- [ ] Created progressive disclosure for complex workflows
