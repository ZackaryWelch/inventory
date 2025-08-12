# Nishiki Backend Go

A Go rewrite of the Nishiki food inventory management backend, built with Clean Architecture principles and Domain-Driven Design.

## Features

- **Clean Architecture**: Separation of concerns with domain, application, and infrastructure layers
- **Domain-Driven Design**: Rich domain models with business logic
- **Authentik OIDC**: Authentication and authorization
- **MongoDB**: Document database with aggregation support
- **Gin Framework**: High-performance HTTP router
- **HTTPS Support**: TLS encryption for secure communication
- **Docker**: Containerized deployment
- **Structured Logging**: JSON logging with optional Seq integration

## Architecture

```
├── domain/           # Business logic layer
│   ├── entities/     # Domain entities (User, Group, Container, Food)
│   ├── repositories/ # Repository interfaces
│   ├── services/     # Domain service interfaces
│   └── usecases/     # Business use cases
├── app/              # Application layer
│   ├── config/       # Configuration management
│   ├── container/    # Dependency injection
│   └── http/         # HTTP transport layer
└── external/         # Infrastructure layer
    ├── adapters/     # Database adapters
    ├── repositories/ # Repository implementations
    └── services/     # External service implementations
```

## Quick Start

### Prerequisites

- Go 1.23+
- Docker and Docker Compose
- Authentik server (for authentication)

### Development Setup

1. **Clone and setup**:
```bash
git clone <repo-url>
cd nishiki-backend-go
cp .env.example .env
# Edit .env with your Authentik configuration
```

2. **Start with Docker Compose**:
```bash
docker compose up --build
```

3. **Or run locally**:
```bash
# Start MongoDB
docker compose up mongodb -d

# Install dependencies
go mod download

# Run the application
go run main.go
```

### Configuration

The application uses TOML configuration with environment variable overrides:

```toml
[server]
port = 3001
debug = true
[server.tls]
enabled = true
cert_file = "./certs/server.crt"
key_file = "./certs/server.key"

[database]
uri = "mongodb://localhost:27017"
database = "nishiki"
timeout = 10

[auth]
authentik_url = "https://your-authentik-server.com"
client_id = "nishiki-backend"
client_secret = "your-client-secret"
```

## API Endpoints

### Authentication
- `GET /auth/me` - Get current user info

### Groups
- `GET /groups` - List user's groups
- `POST /groups` - Create new group
- `GET /groups/{id}/containers` - List containers in group

### Containers
- `POST /containers` - Create new container

### Foods
- `POST /foods` - Add food to container
- `PUT /foods/{id}` - Update food item
- `DELETE /foods/{id}` - Remove food item

### Health
- `GET /health` - Service health check

## Domain Model

### Core Entities

- **User**: Represents a system user with Authentik integration
- **Group**: Food storage group that users can join
- **Container**: Storage container within a group
- **Food**: Individual food items with expiration tracking

### Value Objects

- **Quantity**: Numeric quantities with validation
- **Unit**: Measurement units (kg, l, etc.)
- **Expiry**: Expiration dates with business logic
- **FoodName, GroupName, etc.**: Validated string types

### Business Rules

- Users must be group members to access group resources
- Food quantities must be non-negative
- Expiry dates must be after 1970-01-01
- Group and container names have length constraints

## Development

### Building
```bash
go build .
```

### Testing
```bash
go test ./...
```

### Linting
```bash
golangci-lint run
```

### Formatting
```bash
gofmt -w .
```

## Deployment

### Docker
```bash
docker build -t nishiki-backend-go .
docker run -p 3001:3001 nishiki-backend-go
```

### Docker Compose
```bash
docker compose up --build -d
```

### Environment Variables

All configuration can be overridden with environment variables using the prefix `NISHIKI_`:

- `NISHIKI_SERVER_PORT=3001`
- `NISHIKI_DATABASE_URI=mongodb://...`
- `NISHIKI_AUTH_AUTHENTIK_URL=https://...`

## Security

- HTTPS-only in production
- JWT token validation via Authentik JWKS
- Group-based authorization
- Input validation at domain level
- Parameterized database queries

## Monitoring

- Structured JSON logging
- Health check endpoint
- Optional Seq log aggregation
- Request/response logging middleware

## Contributing

1. Follow Go conventions and patterns
2. Maintain clean architecture boundaries
3. Write comprehensive tests
4. Use structured logging
5. Validate all inputs at domain level

## License

[License information]