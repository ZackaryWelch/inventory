# Next Steps Plan

## Immediate Bug Fixes

### B1. Key normalization in bulk import (critical)

`InferSchema` stores property definitions with snake_case `Key` (e.g. `date_purchased`) but the data rows still use original CSV headers (`Date Purchased`). `CoerceRow` looks up `def.Key` in the row and finds nothing, so no coercion happens and no typed values are stored.

**Fix — `backend/domain/services/type_inference.go`:**

Add a `NormalizeRowKeys(row map[string]interface{}) map[string]interface{}` method that rewrites every key to `toSnakeCase(key)`. Call it in `BulkImportCollectionUseCase.Execute` **before** `CoerceRow`, and also normalize keys when building `properties` maps in all three distribution paths.

**Fix — `backend/domain/usecases/bulk_import_collection_usecase.go`:**

In `executeLocationDistribution` (and the default/target paths), store properties with normalized keys:
```go
properties[services.ToSnakeCase(key)] = value
```

Export `toSnakeCase` as `ToSnakeCase` from the services package.

### B2. `resolveNameField` not used in default/automatic paths

The `default` and `automatic` distribution paths still hardcode `item["name"]` instead of calling `resolveNameField(item, req.NameColumn)`. The `Electronic_Supplies.csv` uses `Name` (capital N), which works by coincidence but `title` / `item` columns won't.

**Fix — `bulk_import_collection_usecase.go`:** Replace the two hardcoded `item["name"]` lookups with `resolveNameField(item, req.NameColumn)`.

---

## Backend — Phase 7: Object Search & Export

### 7a. Search/filter query params on `GET /accounts/{id}/collections/{id}/objects`

Add optional query parameters: `?q=` (name contains), `?tag=`, `?container_id=`, `?property[key]=value`.

- `GetCollectionObjectsRequest` gains `Query`, `Tags []string`, `ContainerID *ContainerID`, `PropertyFilters map[string]string`
- Repository-level filter (done in-memory on the loaded container objects for now; MongoDB filter later)

### 7b. Export endpoint — `GET /accounts/{id}/collections/{id}/export`

Returns objects as CSV. Columns: `name`, `description`, `quantity`, `unit`, `tags`, then one column per property definition (using `DisplayName` as header). If no schema, uses all unique property keys sorted alphabetically.

New use case: `ExportCollectionUseCase`. Controller method `ExportCollection`. Response: `Content-Type: text/csv`, `Content-Disposition: attachment; filename="{collection_name}.csv"`.

### 7c. Group operations backend

Implement the missing HTTP endpoints that the MCP stubs advertise:
- `PUT /groups/{id}` — update group name (new use case wrapping Authentik)
- `DELETE /groups/{id}` — delete group (new use case)

These unblock the three MCP stubs (`update_group`, `delete_group`, and eventually `join_group`).

---

## MCP — Phase 8: Search, Export & Quality

### 8a. New tool: `search_objects`

```
Input: collection_id, query? (name substring), tags? ([]string), container_id?, property_filters? (map[string]string)
```

Calls the new filtered `GetCollectionObjectsUseCase` and returns matching objects with container context.

### 8b. New tool: `export_collection`

```
Input: collection_id, format? ("csv"|"json", default "csv")
```

Returns the export as a string in the result. For CSV, the column order follows the collection's `PropertySchema.Definitions` (display names as headers). Useful for MCP-driven data pipelines.

### 8c. New prompt: `migrate_schema`

A prompt that helps the user review and update a collection's schema after an import — shows inferred types, asks for corrections, then calls `update_collection_schema`.

### 8d. New prompt: `find_by_property`

A prompt for natural-language property searches: "find all electronics for sale" → `search_objects` with `property_filters: {"for_sale": "true"}`.

### 8e. Unblock group stubs

Once 7c group endpoints exist, replace the three `errorResult(fmt.Errorf("backend missing..."))` stubs with real implementations.

---

## Frontend — Phase 9: Typed Rendering

### 9a. Add `PropertySchema` to frontend `Collection` type

`types/user.go` re-exports `response.CollectionResponse` which now includes `PropertySchema *PropertySchemaResponse`. Add corresponding type aliases:

```go
type PropertySchema = response.PropertySchemaResponse
type PropertyDefinition = response.PropertyDefinitionResponse
```

### 9b. Property renderer — `frontend/app/property_renderers.go` (new file)

```go
func RenderPropertyValue(key string, value interface{}, schema *types.PropertySchema) string
```

Dispatch table by `PropertyType`:

| Type | Rendering |
|------|-----------|
| `currency` | `fmt.Sprintf("$%.2f", v)` — right-aligned |
| `date` | `time.Parse(RFC3339, v)` → `Jan 2, 2006` |
| `bool` | `"Yes"` / `"No"` (or checkbox widget) |
| `url` | Display name = last path segment; open via `js.Global().Get("window").Call("open", url)` |
| `numeric` | Right-aligned, no trailing zeros |
| `grouped_text` | Styled chip label |
| `text` (default) | Plain string |

### 9c. Update object property rendering — `frontend/app/objects_ui.go`

When rendering an object's properties, look up the collection's `PropertySchema` and call `RenderPropertyValue` for each key. Fall back to plain string rendering when schema is nil or key not found.

### 9d. Grouped-text filter chips

For every `grouped_text` property in the schema, collect unique values across all loaded objects. Render a horizontal filter bar above the object list with clickable chips (one per unique value). Active filter chips narrow the displayed objects.

State: `activeGroupedTextFilters map[string]string` on `GioApp` (property key → selected value, empty = all).

### 9e. Import dialog improvements — `frontend/app/import_dialog.go`

Add a **Column Mapping** section to the preview dialog (rendered after the data preview):

- **Name column** dropdown: auto-selected to `Name`/`Title`/`Item` if found; otherwise user picks
- **Location column** dropdown: auto-selected to `Location` if found; shows "(none — use automatic)" option
- **Infer types** checkbox: on by default
- **Type preview table**: one row per column with header name, inferred type (shown after a quick client-side inference pass), and an editable dropdown to override

After column mapping, the Execute button sends the correct `name_column`, `location_column`, and `infer_schema` fields.

### 9f. Schema editor dialog

A new dialog accessible from the Collection detail view (gear icon or "Edit Schema" button):

- Table view of current schema definitions (key, display name, type, required)
- Inline type dropdowns per row
- Add/remove rows
- Save calls `PUT /accounts/{id}/collections/{id}/schema`

---

## Priority Order

```
B1 (key normalization)       ← breaks schema inference
B2 (resolveNameField)        ← breaks Name-column imports
7b (export endpoint)         ← enables MCP export tool
8b (export_collection tool)  ← closes the read/write loop in MCP
7a (search filter)           ← high MCP value
8a (search_objects tool)     ← uses 7a
9a–9c (typed rendering)      ← immediately visible frontend value
9d (grouped-text filters)    ← enhances browsing
9e (import dialog)           ← makes import self-service
7c + 8e (group ops)          ← removes stubs
9f (schema editor)           ← power-user feature
8c–8d (new prompts)          ← quality of life for MCP users
```
