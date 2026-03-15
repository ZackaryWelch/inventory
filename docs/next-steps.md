# Next Steps Plan

## Immediate Bug Fixes

### ~~B1. Key normalization in bulk import~~ ✅ Done

`toSnakeCase` exported as `ToSnakeCase`. `NormalizeRowKeys()` added to `TypeInferenceService` and called in `Execute` before `CoerceRow`. All three distribution paths now store properties with normalized snake_case keys. Reserved columns list moved from hardcoded package vars into `TypeInferenceService` (configurable via `[import] reserved_columns` in `app.toml`).

### ~~B2. `resolveNameField` not used in default/automatic paths~~ ✅ Done

Both the default and `executeAutomaticDistribution` paths now call `resolveNameField(item, req.NameColumn)`.

---

## Backend — Phase 7: Object Search & Export

### ~~7a. Search/filter query params on `GET /accounts/{id}/collections/{id}/objects`~~ ✅ Done

Optional query params: `?q=` (name contains, case-insensitive), `?tag=` (repeatable, all must match), `?container_id=`, `?property[key]=value` (substring match).

`GetCollectionObjectsRequest` now has `Query`, `Tags []string`, `ContainerID *ContainerID`, `PropertyFilters map[string]string`. Filtering is in-memory after loading the collection. 11 unit tests in `get_collection_objects_usecase_test.go` cover all filter combinations and access control.

### ~~7b. Export endpoint — `GET /accounts/{id}/collections/{id}/export`~~ ✅ Done

Fixed fields (`name`, `description`, `quantity`, `unit`, `tags`, `expires_at`) always emitted in that order. Display names auto-derived from snake_case → Title Case; schema can override per key. Property columns follow: schema definitions whose keys aren't fixed fields, or all unique property keys sorted alpha if no schema.

`ExportCollectionUseCase` in `domain/usecases/`. `CollectionController.ExportCollection` responds with `Content-Type: text/csv`, `Content-Disposition: attachment; filename="{collection_name}.csv"`. Route: `GET /accounts/{id}/collections/{collection_id}/export`.

### 7c. Group operations backend

Implement the missing HTTP endpoints that the MCP stubs advertise:
- `PUT /groups/{id}` — update group name (new use case wrapping Authentik)
- `DELETE /groups/{id}` — delete group (new use case)

These unblock the three MCP stubs (`update_group`, `delete_group`, and eventually `join_group`).

---

## MCP — Phase 8: Search, Export & Quality

### ~~8a. New tool: `search_objects`~~ ✅ Done

Input: `collection_id`, `query?`, `tags?`, `container_id?`, `property_filters?`. Delegates to the updated `GetCollectionObjectsUseCase`. Returns `{ count, objects[] }`.

### ~~8b. New tool: `export_collection`~~ ✅ Done

Input: `collection_id`, `format?` (`"csv"` | `"json"`, default `"csv"`). CSV uses `ExportCollectionUseCase` (schema-ordered columns). JSON calls `GetCollectionObjectsUseCase` and returns objects array via `jsonResult`. A `textResult` helper was added to `server.go` for plain-text MCP responses.

### 8c. New prompt: `migrate_schema`

A prompt that helps the user review and update a collection's schema after an import — shows inferred types, asks for corrections, then calls `update_collection_schema`.

### 8d. New prompt: `find_by_property`

A prompt for natural-language property searches: "find all electronics for sale" → `search_objects` with `property_filters: {"for_sale": "true"}`.

### 8e. Unblock group stubs

Once 7c group endpoints exist, replace the three `errorResult(fmt.Errorf("backend missing..."))` stubs with real implementations.

---

## Frontend — Phase 9: Typed Rendering

### ~~9a. Add `PropertySchema` to frontend `Collection` type~~ ✅ Done

`PropertySchema = response.PropertySchemaResponse` and `PropertyDefinition = response.PropertyDefinitionResponse` added to the type alias block in `app/gio_app.go`.

### ~~9b. Property renderer — `frontend/app/property_renderers.go`~~ ✅ Done

```go
func RenderPropertyValue(key string, value interface{}, defs []PropertyDefinition) string
```

Takes `[]PropertyDefinition` (no pointer) — nil-safe via `range`. Dispatches by `def.Type`:

| Type | Rendering |
|------|-----------|
| `currency` | Symbol from `CurrencyCode` (USD→`$`, EUR→`€`, etc.) + `%.2f` |
| `date` | Parses RFC3339 or `2006-01-02` → `Jan 2, 2006` |
| `bool` | `"Yes"` / `"No"` (handles bool, string, numeric) |
| `url` | Last path segment via `path.Base` |
| `numeric` | `strconv.FormatFloat` with no trailing zeros |
| `grouped_text` / `text` | Plain string |

Helpers: `findPropertyDef`, `toFloat`.

### ~~9c. Update object property rendering — `frontend/app/collection_detail_view.go`~~ ✅ Done

`renderObjectCard` now renders a **Properties** section between Tags and action buttons. `renderObjectProperties` orders keys schema-first then remaining keys. `propertyDisplayName` uses schema `DisplayName` or falls back to snake_case → Title Case.

### ~~9d. Grouped-text filter chips~~ ✅ Done

`collectGroupedTextValues()` scans loaded objects for unique values per `grouped_text` property in the schema. `renderGroupedTextFilters()` renders a row of chips per property above the search field in the objects column (label + "All" chip + one chip per unique value). `matchesGroupedTextFilters()` applies active filters in `renderObjectsList`. `activeGroupedTextFilters map[string]string` on `GioApp` resets to `nil` on collection change or back navigation. Chips are lazily allocated via `getGroupedTextChipButton(key string)`.

### ~~9e. Import dialog improvements — `frontend/app/import_dialog.go`~~ ✅ Done

A **Column Mapping** section is now rendered after the data preview:

- **Name column** chips: auto-selected on load via `detectNameColumn`; user can override
- **Location column** chips: auto-selected if a `location` column exists; `(none)` chip sets `importLocationColumn = nil` (no sentinel string — `nil` means automatic distribution)
- **Infer property types** checkbox: on by default (`widget.Bool`, initialised to `true` on file load)

`executeImport` reads `ga.importNameColumn`, `ga.importLocationColumn` (`*string`), and `importInferSchemaCheck.Value` instead of re-detecting. Note: the **Type preview table** (per-column type override) was deferred — name/location/infer coverage handles the primary self-service use case.

### ~~9f. Object delete fix + dual delete endpoints~~ ✅ Done

`ObjectResponse` now includes `container_id` (populated from the embedded container context). Two delete endpoints exist:

- `DELETE /accounts/{id}/collections/{collection_id}/containers/{container_id}/objects/{object_id}` — explicit container in path (`RemoveObjectFromContainer`)
- `DELETE /accounts/{id}/objects/{object_id}` — container_id optional query param; if omitted, use case calls new `ContainerRepository.FindByObjectID` to look it up

`GetCollectionObjectsUseCase` now returns `[]ObjectWithContainerID` (Object + ContainerID) so objects always carry their container context through the pipeline. All `NewObjectResponse` call sites updated. Frontend `handleObjectDelete` now reads `obj.ContainerID` from the response.

### 9g. Schema editor dialog

A new dialog accessible from the Collection detail view (gear icon or "Edit Schema" button):

- Table view of current schema definitions (key, display name, type, required)
- Inline type dropdowns per row
- Add/remove rows
- Save calls `PUT /accounts/{id}/collections/{id}/schema`

---

## Priority Order

```
~~B1 (key normalization)~~         ✅ done
~~B2 (resolveNameField)~~          ✅ done
~~7b (export endpoint)~~           ✅ done
~~8b (export_collection tool)~~    ✅ done
~~7a (search filter)~~             ✅ done  (11 unit tests)
~~8a (search_objects tool)~~       ✅ done
~~9a–9c (typed rendering)~~        ✅ done
~~9d (grouped-text filters)~~      ✅ done
~~9e (import dialog)~~             ✅ done  (type preview table deferred)
~~9f (object delete + container_id in response)~~ ✅ done
7c + 8e (group ops)                ← removes stubs
9g (schema editor)                 ← power-user feature
8c–8d (new prompts)                ← quality of life for MCP users
```
