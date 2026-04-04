# Suggested Commands for Nishiki Development

## Backend (run from `backend/` directory)

```bash
# Run locally (requires MongoDB and Authentik)
go run main.go

# Run with Docker (from repo root)
docker compose up --build

# Generate mocks (REQUIRED before running tests)
go generate ./domain/...

# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run integration tests
go test ./test/integration/...

# Run tests for specific package
go test ./domain/usecases

# Format code
gofmt -w .

# Lint
golangci-lint run
```

## Frontend (run from `frontend/` directory)

```bash
# Build for WebAssembly (outputs to gio-web/)
go run cmd/web/main.go

# Serve the WASM build locally
go run cmd/serve/main.go

# Build native desktop binary
go build ./cmd/desktop/

# Run tests
go test ./...

# Format code
gofmt -w .
```

## Docker (from repo root)

```bash
docker compose up --build
docker compose up
docker compose down
```

## Config Files
- Backend: `backend/app.toml` (copy from `backend/app.toml.example`)
- Frontend: `frontend/config.toml` (desktop) or `frontend/app/config/config.toml` (WASM embed)
