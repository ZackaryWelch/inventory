# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Nishiki Backend Go is a comprehensive inventory management system built with Go, implementing Clean Architecture and Domain-Driven Design principles. It provides RESTful APIs for managing collections of various object types (food, books, video games, music, board games) with hierarchical organization through containers and categories, plus space management and bulk import features.

## Development Commands

### Build and Run
```bash
# Build the application
go build .

# Run locally (requires MongoDB)
go run main.go

# Format all Go code
gofmt -w .

# Run tests
go test ./...

# Run with race detection
go test -race ./...
```

### Docker Development
```bash
# Start all services (MongoDB + backend)
docker compose up --build

# Start only MongoDB for local development
docker compose up mongodb -d

# Clean up
docker compose down -v
```

### Dependencies
```bash
# Download dependencies
go mod download

# Clean up dependencies
go mod tidy

# Verify dependencies
go mod verify
```

## Architecture Overview

This project follows Clean Architecture with clear separation of concerns:

### Domain Layer (`domain/`)
- **Entities**: Core business objects with rich behavior
  - `User`: System users with Authentik integration
  - `Group`: Shared access groups with member management
  - `Collection`: Inventory collections with object type constraints
  - `Container`: Storage containers within collections for organization
  - `Object`: Inventory items with flexible properties and tags
  - `Category`: Object categorization system for all object types
- **Repository Interfaces**: Data access contracts
  - `CollectionRepository`: Collection CRUD and query operations
  - `ContainerRepository`: Container operations within collections
  - `CategoryRepository`: Category operations for object classification
- **Service Interfaces**: External service contracts
  - `AuthService`: Authentication and user provisioning
- **Use Cases**: Business logic orchestration
  - Group creation and management
  - Collection lifecycle (create, read, update, delete)
  - Container management within collections
  - Object management within containers
  - Category-based organization
  - Bulk import functionality for various object types
  - Space planning and organization algorithms

### Application Layer (`app/`)
- **Configuration**: TOML-based config with Viper
- **Dependency Injection**: Container-based DI system
- **HTTP Transport**: Gin-based REST API
  - Controllers for auth, groups, users, containers, and foods
  - Authentication and logging middleware
  - Request/response models and validation
  - CORS support for frontend integration
  - Error handling middleware

### Infrastructure Layer (`external/`)
- **MongoDB Adapter**: Database connectivity and transactions
- **Repository Implementations**: MongoDB-based data access
  - `CollectionMongoRepository`: Collection operations with embedded containers/objects
  - `ContainerMongoRepository`: Container operations within collections
  - `CategoryMongoRepository`: Category operations for object classification
- **Authentik Service**: OIDC token validation and user provisioning
- **Organization Service**: Space calculation and auto-organization algorithms
- **Mocks**: Generated mocks for testing (auth service, repositories)

## Key Features

### Authentication & Authorization
- **Authentik OIDC Integration**: JWT token validation via JWKS
- **Multiple OAuth Clients**: Support for multiple frontend applications (web, mobile, etc.)
  - Each client has its own client_id, client_secret, and redirect_url
  - Backend automatically routes OAuth flows based on redirect_uri parameter
  - Clients are matched by redirect_uri during token exchange
  - OIDC config endpoint requires client_id query parameter
- **User Provisioning**: Automatic user creation from OIDC claims
- **Group-based Authorization**: Access control via group membership
- **Middleware**: Authentication required for all API endpoints

### Database Design
- **MongoDB Collections**: `users`, `groups`, `collections`, `categories`
- **Embedded Documents**: 
  - Containers embedded within collections
  - Objects embedded within containers
  - Hierarchical structure: Collection > Container > Object
- **Indexes**: Optimized queries for common access patterns
- **Transactions**: Consistency for multi-document operations
- **Object Types**: Support for food, book, videogame, music, boardgame, and general objects
- **Categories**: Universal categorization system for all object types

### Domain Logic
- **Rich Domain Models**: Business rules enforced at entity level
- **Value Object Validation**: Type-safe domain concepts
- **Use Case Pattern**: Clear business operation boundaries
- **Error Handling**: Domain-specific error types

## API Design

### RESTful Endpoints

#### Core System
- `GET /health` - Service health check
- `GET /auth/me` - Current user information

#### Group Management
- `GET /groups` - User's groups
- `POST /groups` - Create new group
- `GET /groups/{id}` - Get specific group
- `GET /groups/{id}/users` - Group users
- `POST /groups/join` - Join existing group

#### Account & Collection Management
- `GET /accounts/{id}` - Get account details
- `GET /accounts/{id}/groups` - Get user's groups
- `POST /accounts/{id}/groups` - Create new group
- `PUT /accounts/{id}/groups/{id}` - Update group
- `DELETE /accounts/{id}/groups/{id}` - Delete group
- `POST /accounts/{id}/groups/{id}/invite` - Invite users to group
- `GET /accounts/{id}/collections` - Get user's collections
- `POST /accounts/{id}/collections` - Create collection
- `GET /accounts/{id}/collections/{id}` - Get collection details
- `PUT /accounts/{id}/collections/{id}` - Update collection
- `DELETE /accounts/{id}/collections/{id}` - Delete collection

#### Container Management (within Collections)
- `GET /accounts/{id}/collections/{id}/containers` - Get collection containers
- `POST /accounts/{id}/collections/{id}/containers` - Create container
- `GET /accounts/{id}/collections/{id}/containers/{id}` - Get container details
- `PUT /accounts/{id}/collections/{id}/containers/{id}` - Update container
- `DELETE /accounts/{id}/collections/{id}/containers/{id}` - Delete container

#### Object Management (within Containers)
- `GET /accounts/{id}/collections/{id}/containers/{id}/objects` - Get container objects
- `POST /accounts/{id}/objects` - Add object to container
- `PUT /accounts/{id}/objects/{id}` - Update object properties
- `DELETE /accounts/{id}/objects/{id}` - Delete object

#### Bulk Operations & Organization
- `POST /accounts/{id}/import` - Bulk import (creates new collection)
- `POST /accounts/{id}/collections/{id}/import` - Bulk import to collection
- `POST /accounts/{id}/organize` - Auto-organize collections using space algorithms

#### Categories
- `GET /categories` - Get all categories
- `POST /categories` - Create category
- `PUT /categories/{id}` - Update category
- `DELETE /categories/{id}` - Delete category

### Request/Response Patterns
- JSON request/response bodies
- Proper HTTP status codes (200, 201, 400, 401, 403, 404, 500)
- Structured error responses
- Input validation at multiple layers

## Configuration

### TOML Configuration (`app.toml`)
```toml
[server]
port = 3001
debug = true
[server.tls]
enabled = true
cert_file = "./certs/server.crt"
key_file = "./certs/server.key"

[database]
host = "localhost"           # MongoDB host
port = 27017                 # MongoDB port
username = "root"            # MongoDB username (leave empty for no auth)
password = "password"        # MongoDB password (leave empty for no auth)
auth_source = "admin"        # Authentication database (usually "admin" for root users)
database = "nishiki"         # Application database name
timeout = 10                 # Connection timeout in seconds

# Alternative: use a complete URI (overrides individual fields above)
# uri = "mongodb://username:password@localhost:27017/?authSource=admin"

[auth]
authentik_url = "https://your-authentik-server.com"  # Update with your Authentik URL
jwks_cache_duration = 300
allow_self_signed = false  # Set to true for development with self-signed certs
api_token = "your-authentik-api-token"  # Required for group/user management

# Multiple OAuth clients support - add one entry for each frontend/client
[[auth.clients]]
provider_name = "nishiki"  # Authentik provider/application name
client_id = "your-web-client-id"
client_secret = "your-web-client-secret"
redirect_url = "http://localhost:3000/auth/callback"

[[auth.clients]]
provider_name = "nishiki-mobile"  # Different Authentik application
client_id = "your-mobile-client-id"
client_secret = "your-mobile-client-secret"
redirect_url = "myapp://oauth/callback"

[logging]
level = "info"
seq_endpoint = "https://your-seq-server.com"  # Optional Seq logging endpoint
seq_api_key = "your-seq-api-key"              # Optional Seq API key
```

### Environment Variables
All configuration can be overridden with `NISHIKI_` prefixed environment variables:
- `NISHIKI_SERVER_PORT=3001`
- `NISHIKI_DATABASE_HOST=100.93.246.119`
- `NISHIKI_DATABASE_PORT=27017`
- `NISHIKI_DATABASE_USERNAME=root`
- `NISHIKI_DATABASE_PASSWORD=password`
- `NISHIKI_DATABASE_AUTH_SOURCE=admin`
- `NISHIKI_DATABASE_URI=mongodb://...` (alternative to individual fields)
- `NISHIKI_AUTH_AUTHENTIK_URL=https://...`
- `NISHIKI_AUTH_API_TOKEN=your-api-token`
- `NISHIKI_AUTH_ALLOW_SELF_SIGNED=true`
- `NISHIKI_LOGGING_SEQ_ENDPOINT=https://...`
- `NISHIKI_LOGGING_SEQ_API_KEY=your-api-key`

Note: OAuth client configuration (clients array) must be defined in the TOML file and cannot be overridden via environment variables.

## Testing Strategy

### Unit Tests
- Domain entity validation
- Use case business logic (with existing test files)
- Repository interface contracts
- Controller functionality
- Authentication service integration

### Integration Tests
- MongoDB repository implementations
- HTTP endpoint functionality
- Authentication middleware
- End-to-end API workflows

### Test Files Present
- `domain/usecases/get_groups_usecase_test.go`
- `domain/usecases/create_container_usecase_test.go`
- `app/http/controllers/user_controller_test.go`
- `external/services/authentik_auth_service_test.go`

### Test Organization
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./domain/entities
```

## Security Considerations

### Authentication
- HTTPS-only in production
- JWT token validation via Authentik JWKS
- Bearer token authentication
- Automatic user provisioning

### Authorization
- Group membership required for resource access
- User context propagated through middleware
- Domain-level access control

### Input Validation
- Request validation at HTTP layer
- Domain validation at entity level
- Type-safe value objects
- Parameterized database queries

## Deployment

### Docker
- Multi-stage build for optimized images
- Non-root user for security
- Health checks included
- TLS certificate support

### Environment
- MongoDB database required
- Authentik OIDC provider required
- Optional Seq logging endpoint
- HTTPS certificates for production

## Logging and Monitoring

### Structured Logging
- JSON format with Zap logger
- Request/response logging middleware
- Error tracking and debugging
- Configurable log levels

### Health Checks
- Database connectivity
- Service readiness
- Docker health checks
- Monitoring integration points

## Development Guidelines

### Code Style
- Follow Go conventions and idioms
- Use `gofmt` for consistent formatting
- Implement interfaces, not concrete types
- Prefer composition over inheritance

### Architecture Patterns
- Maintain clean architecture boundaries
- Domain logic in domain layer only
- Infrastructure details in external layer
- Dependency inversion principle

### Error Handling
- Domain-specific error types
- Proper error propagation
- Structured error responses
- Logging of error context

### Testing
- Write tests for all business logic
- Use table-driven tests where appropriate
- Mock external dependencies
- Test error conditions

## Important Files

### Entry Points
- `main.go` - Application entry point with graceful shutdown
- `app/container/container.go` - Dependency injection setup
- `app/http/routes/routes.go` - HTTP route configuration

### Configuration
- `app.toml` - Application configuration
- `docker-compose.yml` - Development environment (MongoDB commented out)
- `scripts/mongo-init.js` - MongoDB initialization script

### Domain Core
- `domain/entities/` - Core business entities
- `domain/usecases/` - Business use cases
- `domain/repositories/` - Data access interfaces

### Infrastructure
- `external/adapters/mongodb.go` - Database connectivity
- `external/repositories/` - MongoDB implementations
  - `container_mongo_repository.go` - Container data operations
  - `category_mongo_repository.go` - Category data operations
- `external/services/authentik_auth_service.go` - OIDC integration

### Testing & Mocks
- `mocks/` - Generated mock implementations
  - `mock_auth_service.go` - Authentication service mock
  - `mock_container_repository.go` - Container repository mock
  - `mock_category_repository.go` - Category repository mock

This architecture provides a robust, maintainable foundation for the food inventory management system with clear separation of concerns and comprehensive business logic implementation.