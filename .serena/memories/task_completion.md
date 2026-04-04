# Task Completion Checklist

When completing a coding task in Nishiki, run the following:

## Backend Changes (in `backend/`)

1. **If domain interfaces changed** — regenerate mocks first:
   ```bash
   go generate ./domain/...
   ```

2. **Format**:
   ```bash
   gofmt -w .
   ```

3. **Lint**:
   ```bash
   golangci-lint run
   ```

4. **Test**:
   ```bash
   go test ./...
   ```

## Frontend Changes (in `frontend/`)

1. **Format**:
   ```bash
   gofmt -w .
   ```

2. **Test**:
   ```bash
   go test ./...
   ```

3. **Build check** (WASM):
   ```bash
   go run cmd/web/main.go
   ```

## Notes
- No linter configured for frontend (only `gofmt`)
- Integration tests require running infrastructure (MongoDB, Authentik) — skip unless specifically testing integration
- MCP stubs (join_group, update_group, delete_group) are intentionally incomplete — backend returns 501
