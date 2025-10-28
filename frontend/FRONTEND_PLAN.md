# Frontend Implementation Plan - Cogent Core

## Overview
Migrate from React/Next.js frontend to Cogent Core cross-platform GUI framework while maintaining the same user experience and functionality.

## Technology Stack
- **Framework**: Cogent Core (Go-based cross-platform GUI)
- **Authentication**: Authentik OIDC integration
- **API**: REST API calls to existing Go backend
- **Platform Support**: Desktop (Windows, macOS, Linux), Mobile (iOS, Android), Web

## Architecture Requirements

### Authentication System
- OIDC token-based authentication with Authentik
- Token storage and refresh handling
- Group-based access control validation
- User context management throughout the app

### Core Data Models
- **User**: From Authentik with group memberships
- **Group**: Shared access groups with member management
- **Collection**: Object type-constrained inventory collections
- **Container**: Storage organization within collections
- **Object**: Inventory items with flexible properties and tags
- **Category**: Classification system for objects

## UI Component Structure

### Main Navigation
- **Collection Grid**: Display user's collections with thumbnails
- **Group Management**: Create/join groups, manage members
- **User Profile**: Account settings and authentication

### Collection Management
- **Collection List/Grid**: Browse user's collections
- **Collection Detail**: View containers within collection
- **Collection Editor**: Create/edit collection properties
- **Collection Settings**: Access control, sharing, categories

### Container Management
- **Container List**: View containers within collection
- **Container Detail**: View objects within container
- **Container Editor**: Create/edit container properties
- **Space Planning**: Visual organization tools

### Object Management  
- **Object Grid**: Display objects with filtering/search
- **Object Detail**: View/edit object properties
- **Object Editor**: Create/edit objects with flexible properties
- **Tag Management**: Add/remove/filter by tags

### Import & Organization
- **Bulk Import**: JSON/CSV import with field mapping
- **Import Wizard**: Step-by-step import process
- **Auto-Organization**: Space optimization algorithms
- **Category Management**: Create/assign categories

## Key Features to Implement

### Phase 1: Core Framework & Authentication
1. **Cogent Core Project Setup**
   - Initialize new Cogent Core application
   - Configure build for multiple platforms
   - Set up project structure following Cogent Core patterns

2. **Authentication Integration**
   - OIDC authentication flow with Authentik
   - Token management (storage, refresh, validation)
   - Protected route/view handling
   - User context provider

3. **API Client Layer**
   - HTTP client for backend API calls
   - Request/response models matching backend
   - Error handling and retry logic
   - Authentication header management

### Phase 2: Core UI Components
1. **Navigation Framework**
   - Main navigation drawer/sidebar
   - Breadcrumb navigation
   - Search functionality
   - User profile menu

2. **Collection Management**
   - Collection grid view with thumbnails
   - Collection creation/editing forms
   - Collection settings and permissions
   - Delete/archive collections

3. **Container & Object Views**
   - Hierarchical navigation (Collection > Container > Object)
   - List and grid view modes
   - Filtering and search
   - Sorting options

### Phase 3: Advanced Features
1. **Object Management**
   - Object creation/editing with flexible properties
   - Tag management system
   - Bulk operations (edit, delete, move)
   - Object relationships

2. **Import System**
   - File upload handling (JSON, CSV)
   - Field mapping interface
   - Import preview and validation
   - Progress tracking for bulk imports

3. **Organization Tools**
   - Space planning visualization
   - Auto-organization algorithms
   - Container optimization
   - Category-based organization

### Phase 4: Polish & Mobile
1. **Responsive Design**
   - Mobile-optimized layouts
   - Touch-friendly interactions
   - Swipe gestures
   - Mobile navigation patterns

2. **Performance**
   - Lazy loading for large collections
   - Image optimization
   - Caching strategies
   - Background sync

## API Endpoints to Consume

### Authentication
- `GET /auth/me` - Current user info
- Group membership validation via token claims

### Groups
- `GET /groups` - User's groups
- `POST /groups` - Create group
- `POST /groups/join` - Join group
- `GET /groups/{id}/users` - Group members

### Collections
- `GET /accounts/{id}/collections` - User collections
- `POST /accounts/{id}/collections` - Create collection
- `GET /accounts/{id}/collections/{id}` - Collection details
- `PUT /accounts/{id}/collections/{id}` - Update collection
- `DELETE /accounts/{id}/collections/{id}` - Delete collection

### Containers
- `GET /accounts/{id}/collections/{id}/containers` - Collection containers
- `POST /accounts/{id}/collections/{id}/containers` - Create container
- `GET /accounts/{id}/collections/{id}/containers/{id}` - Container details
- `PUT /accounts/{id}/collections/{id}/containers/{id}` - Update container
- `DELETE /accounts/{id}/collections/{id}/containers/{id}` - Delete container

### Objects
- `GET /accounts/{id}/collections/{id}/objects` - Collection objects
- `POST /accounts/{id}/objects` - Create object
- `PUT /accounts/{id}/objects/{id}` - Update object
- `DELETE /accounts/{id}/objects/{id}` - Delete object
- `POST /accounts/{id}/import` - Bulk import
- `POST /accounts/{id}/collections/{id}/import` - Collection bulk import

### Organization
- `POST /accounts/{id}/organize` - Auto-organize collections

### Categories  
- `GET /categories` - All categories
- `POST /categories` - Create category
- `PUT /categories/{id}` - Update category
- `DELETE /categories/{id}` - Delete category

## Development Workflow

### Setup Phase
1. Install Cogent Core framework and tools
2. Create new Cogent Core project structure
3. Set up build configuration for target platforms
4. Configure development environment

### Implementation Order
1. Start with authentication and basic navigation
2. Implement collection browsing (read-only)
3. Add container and object viewing
4. Implement CRUD operations
5. Add import and organization features
6. Polish mobile experience

### Testing Strategy
- Unit tests for API client and data models
- Integration tests for authentication flow
- UI tests for critical user paths
- Cross-platform testing (desktop, mobile, web)
- Performance testing with large datasets

## Migration Considerations

### From React/Next.js
- Convert React components to Cogent Core widgets
- Migrate CSS styling to Cogent Core theming
- Replace React hooks with Cogent Core state management
- Convert API calls from fetch/axios to Cogent Core HTTP client

### Data Compatibility
- Maintain same API contract with backend
- Preserve user data and authentication flow
- Keep same URL patterns for deep linking
- Maintain feature parity with existing frontend

## Platform-Specific Features

### Desktop
- Keyboard shortcuts
- Multiple window support
- Native file system access
- System tray integration

### Mobile
- Touch gestures and interactions
- Camera integration for object photos
- Offline data caching
- Push notifications

### Web
- Progressive Web App features
- Web share API integration
- Local storage management
- Browser history integration