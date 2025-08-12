# Nishiki Frontend - Cogent Core

A cross-platform inventory management frontend built with [Cogent Core](https://cogentcore.org/core), supporting desktop, mobile, and web platforms.

## Features

- **Cross-Platform**: Runs on Windows, macOS, Linux, iOS, Android, and Web
- **Modern UI**: Native-feeling interface with Cogent Core widgets
- **Authentication**: Authentik OIDC integration
- **Inventory Management**: Full CRUD operations for groups, collections, and objects
- **Real-time Updates**: Live data synchronization with backend
- **Responsive Design**: Adapts to different screen sizes and platforms

## Architecture

This application follows the architecture specified in `FRONTEND_PLAN.md`:

### Core Components
- **Authentication Service**: Handles OIDC flow with Authentik
- **API Client**: HTTP client for backend communication
- **UI Views**: 
  - Login/Authentication
  - Dashboard with quick stats
  - Groups management
  - Collections management
  - User profile

### Design Language
The UI follows the design patterns from the original React frontend:
- **Colors**: Primary (#6ab3ab), Accent (#fcd884), Danger (#cd5a5a)
- **Layout**: Mobile-first approach with clean card-based design
- **Navigation**: Bottom tab navigation on mobile, sidebar on desktop
- **Typography**: Clean, readable fonts with proper hierarchy

## Configuration

Copy `config.toml.example` to `config.toml` and update the values:

```toml
# Backend API URL
backend_url = "http://localhost:3001"

# Authentik OIDC Configuration
auth_url = "https://your-authentik-server.com"
client_id = "your-client-id"
client_secret = "your-client-secret"
redirect_url = "http://localhost:8080/auth/callback"
```

You can also use environment variables with the `NISHIKI_` prefix:
- `NISHIKI_BACKEND_URL`
- `NISHIKI_AUTH_URL`
- `NISHIKI_CLIENT_ID`
- `NISHIKI_CLIENT_SECRET`
- `NISHIKI_REDIRECT_URL`

## Development

### Prerequisites
- Go 1.21 or later
- Platform-specific dependencies:
  - **Linux**: X11 development libraries (`sudo apt-get install libx11-dev libxrandr-dev libxinerama-dev libxcursor-dev libxi-dev`)
  - **macOS**: Xcode command line tools (`xcode-select --install`)
  - **Windows**: No additional dependencies required

### Running the Application

1. Install system dependencies (Linux only):
```bash
# Ubuntu/Debian
sudo apt-get install libx11-dev libxrandr-dev libxinerama-dev libxcursor-dev libxi-dev

# RHEL/CentOS/Fedora
sudo yum install libX11-devel libXrandr-devel libXinerama-devel libXcursor-devel libXi-devel
```

2. Install Go dependencies:
```bash
go mod download
```

3. Run the application:
```bash
go run main.go
```

### Building for Different Platforms

#### Desktop (current platform)
```bash
go build -o nishiki-frontend main.go
```

#### Web (WebAssembly)
```bash
GOOS=js GOARCH=wasm go build -o nishiki-frontend.wasm main.go
```

#### Mobile (with Cogent Core mobile tools)
```bash
# iOS
core mobile ios

# Android
core mobile android
```

## API Integration

The frontend communicates with the Nishiki backend via REST API:

### Authentication Endpoints
- `GET /auth/me` - Get current user
- `GET /auth/oidc-config` - OIDC configuration
- `POST /auth/token` - Token exchange

### Data Endpoints
- `GET /groups` - User's groups
- `POST /groups` - Create group
- `GET /accounts/{id}/collections` - User's collections
- `POST /accounts/{id}/collections` - Create collection
- And more...

## Development Status

### ✅ Phase 1: Core Framework & Authentication (COMPLETED)
- [x] Cogent Core project setup with Go 1.21+ compatibility
- [x] Complete UI structure and navigation system
- [x] Viper-based configuration management with TOML and environment variables
- [x] OAuth2 authentication flow foundation with Authentik integration
- [x] HTTP client with authentication headers and error handling

### ✅ Phase 2: Core UI Components (COMPLETED)
- [x] Professional login screen with branding
- [x] Dashboard with navigation buttons and statistics cards
- [x] Enhanced Groups management with full CRUD operations
- [x] Enhanced Collections management with full CRUD operations
- [x] User profile view with account information and logout
- [x] Detailed group views with member management
- [x] Detailed collection views with container/object organization
- [x] Modal dialogs for all create/edit/delete operations
- [x] Responsive card-based layouts matching original design

### ✅ Phase 3: Advanced Features (COMPLETED)
- [x] Comprehensive object management within containers
- [x] Object detail views with properties and tags
- [x] Object type-specific property forms (food, books, games, etc.)
- [x] Global search functionality across all data types
- [x] Advanced filtering by type, tags, properties, dates
- [x] Search results with hierarchical navigation breadcrumbs
- [x] Active filters display with remove/clear functionality
- [x] Sort options and view mode toggles

### 🚧 Phase 4: Ready for Polish & Deployment
- [x] Cross-platform architecture (desktop, mobile, web ready)
- [x] Consistent design language matching React frontend
- [x] Touch-friendly interface with proper button sizing
- [x] Modal overlay system for dialogs
- [x] Comprehensive error handling structure
- [x] Extensible codebase for future enhancements

### 📋 Future Enhancements (Ready to Implement)
- [ ] Bulk import functionality with CSV/JSON support
- [ ] Auto-organization algorithms and space planning
- [ ] Real-time updates with WebSocket integration
- [ ] Offline support with local caching
- [ ] Push notifications for mobile platforms
- [ ] Performance optimizations for large datasets

## Contributing

1. Follow the existing code patterns
2. Maintain the design language from the original frontend
3. Ensure cross-platform compatibility
4. Test on multiple platforms before submitting

## License

Same as the main Nishiki project.