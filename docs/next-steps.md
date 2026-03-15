# Next Steps

## Active Work

### Frontend — Schema editor dialog

A new dialog accessible from the Collection detail view (gear icon or "Edit Schema" button):

- Table view of current schema definitions (key, display name, type, required)
- Inline type dropdowns per row
- Add/remove rows
- Save calls `PUT /accounts/{id}/collections/{id}/schema`

---

## Potential Future Work

### Group ownership & member management

Currently any authenticated user can call `PUT/DELETE /groups/{id}` and `POST/DELETE /groups/{id}/users/{user_id}`. A group owner concept would let us restrict those operations to the group creator. Authentik doesn't natively model "owner", so this would require storing ownership in MongoDB or using a group attribute.

### Group join via invite

`POST /groups/join` returns 501. Authentik has no built-in invitation token system, so this needs an invitation table in MongoDB (token → group_id + expiry), a `POST /groups/{id}/invite` endpoint to generate tokens, and the join endpoint to validate and redeem them.

### Import: per-column type override UI

The import dialog deferred a type preview table that lets users override the inferred type per column before executing. The backend already supports the schema; it's a frontend-only addition to `import_dialog.go`.

### Audit logging via Authentik Events API

The Authentik `events` API (`/api/v3/events/`) records logins, group changes, and user modifications. Group owners could view when members logged in/out and when the group was modified. Relevant endpoints: `events_list`, `events_retrieve`. This is read-only and doesn't require any new infrastructure.

### Fine-grained permissions (RBAC)

Current model: you own what you created or what's shared to your group. If per-user roles within a group are needed later (e.g. read-only member vs. editor), Authentik's `rbac` API and `core_users_me_retrieve` provide the building blocks. Defer until there's a concrete use case.

### MCP: `join_group` tool

Blocked on the invite infrastructure above. Once token-based invites exist, the stub in `tools.go` can be replaced with a real implementation.

### MCP: `migrate_schema` / `find_by_property` prompts — type preview table integration

Once the import dialog gains per-column type overrides, the `migrate_schema` prompt could suggest corrections in a format directly pasteable into that UI.

### Events / activity feed in the frontend

Surface recent inventory changes (objects added/removed, imports, schema updates) in the UI. Could use a MongoDB change stream or a lightweight events table.
