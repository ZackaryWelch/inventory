# Backend Implementation Completion Plan

## Current Status
The backend has been substantially refactored from food-specific to general inventory management, but several compilation errors and implementation gaps remain.

## Critical Issues Requiring Immediate Attention

### 1. Compilation Errors
**Status**: Active compilation failures preventing build

**Issues**:
- BulkImport response handling using gin.H instead of typed responses
- Missing request/response type definitions
- Import path inconsistencies
- Use case constructor signature mismatches

**Next Steps**:
1. Fix BulkImport methods to use `response.BulkImportResponse` instead of gin.H
2. Update DeleteObject to use `response.DeleteObjectResponse` instead of gin.H
3. Fix import paths and package references
4. Resolve use case constructor calls

### 2. Missing Implementation - Collection Bulk Import
**Status**: Endpoint exists but not properly implemented

**Current Issue**:
- `BulkImportToCollection` method calls wrong use case
- No proper container selection logic for collection imports
- Response handling incomplete

**Required Fix**:
- Create collection-level bulk import use case OR
- Modify existing use case to handle collection-level imports
- Implement container selection/creation logic
- Fix response handling

### 3. Go Vet Errors
**Status**: Unknown - needs to be run and addressed

**Action Required**:
```bash
go vet ./...
```
Fix any issues found before proceeding.

### 4. Unit Test Failures
**Status**: Unknown - needs testing after compilation fixes

**Existing Test Files**:
- `domain/usecases/get_groups_usecase_test.go`
- `domain/usecases/create_container_usecase_test.go`
- `app/http/controllers/user_controller_test.go`
- `external/services/authentik_auth_service_test.go`

**Action Required**:
```bash
go test ./...
```
Update test files to match new use cases and controller signatures.

## Specific Code Fixes Needed

### 1. Object Controller Response Types
**File**: `app/http/controllers/object_controller.go`

**Lines 519-524** - Fix BulkImport response:
```go
// Current (incorrect):
c.JSON(http.StatusOK, gin.H{
    "imported": resp.Imported,
    "failed":   resp.Failed,
    "total":    resp.Total,
    "errors":   resp.Errors,
})

// Should be:
c.JSON(http.StatusOK, response.BulkImportResponse{
    Imported: resp.Imported,
    Failed:   resp.Failed,
    Total:    resp.Total,
    Errors:   resp.Errors,
})
```

**Line 402** - Fix DeleteObject response:
```go
// Current (incorrect):
c.JSON(http.StatusOK, gin.H{"success": resp.Success})

// Should be:
c.JSON(http.StatusOK, response.DeleteObjectResponse{
    Success: resp.Success,
})
```

**Lines 600-626** - Fix BulkImportToCollection:
- Implement proper collection-level bulk import logic
- Use correct use case and request/response types
- Add container selection logic

### 2. Missing Response Constructor
**File**: `app/http/response/object_response.go`

Add missing constructor:
```go
func NewDeleteObjectResponse(success bool) DeleteObjectResponse {
    return DeleteObjectResponse{
        Success: success,
    }
}
```

### 3. Import Path Issues
**Verify and fix import paths**:
- Current: `github.com/nishiki/backend-go/...`  
- Should be: `github.com/nishiki-tech/nishiki-backend/...`

### 4. Use Case Request Type Issues
**File**: `app/http/controllers/object_controller.go`

**Line 592** - Wrong request type:
```go
// Current:
ucReq := usecases.BulkImportRequest{...}

// Should be:
ucReq := usecases.BulkImportObjectsRequest{...}
```

## Repository Implementation Gaps

### 1. Collection Repository Enhancements
**File**: `external/repositories/collection_mongo_repository.go`

**Missing Methods**:
- Enhanced object querying across containers
- Better error handling for nested operations
- Transaction support for bulk operations

### 2. Container Repository Query Improvements
**File**: `external/repositories/container_mongo_repository.go`

**Needed Improvements**:
- Optimize object retrieval queries
- Add container capacity/organization logic
- Better validation for container operations

## Dependency Injection Updates

### 1. Container Constructor Signatures
**File**: `app/container/container.go`

**Issues**:
- Use case constructors may have incorrect signatures
- Missing repository dependencies
- Authentication service wiring

**Action Required**:
1. Review all use case constructor calls
2. Ensure all required dependencies are passed
3. Verify interface implementations

## API Specification Completeness

### 1. Missing Endpoints
Review if these endpoints are needed:
- `POST /accounts/{id}/collections/{id}/containers/{id}/objects` - Add object to specific container
- `GET /accounts/{id}/containers/{id}/objects` - Get objects in specific container
- `PUT /accounts/{id}/collections/{id}/organize` - Collection-specific organization

### 2. Route Configuration
**File**: `app/http/routes/routes.go`

**Verify**:
- All controller methods are properly routed
- Path parameters match controller expectations
- HTTP methods are correct

## Testing Requirements

### 1. Unit Tests to Update
- Update existing tests to use new use cases
- Fix constructor calls in test files
- Add tests for new object management use cases

### 2. Integration Tests Needed
- End-to-end API testing
- Authentication middleware testing
- Database transaction testing
- Error handling validation

### 3. Test Commands
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detection
go test -race ./...
```

## Development Environment Setup

### 1. Required Services
- MongoDB instance running
- Authentik OIDC provider configured
- Environment variables or config file setup

### 2. Build Commands
```bash
# Build application
go build .

# Run locally
go run main.go

# Format code
gofmt -w .

# Check for issues
go vet ./...
```

## Completion Checklist

### Immediate (Session 1)
- [ ] Fix compilation errors in object_controller.go
- [ ] Create missing response constructors  
- [ ] Fix import paths throughout codebase
- [ ] Run `go vet ./...` and fix issues
- [ ] Run `go test ./...` and fix failing tests

### Follow-up (Session 2)
- [ ] Implement proper BulkImportToCollection logic
- [ ] Enhance repository implementations
- [ ] Update dependency injection container
- [ ] Complete API route configuration
- [ ] Add comprehensive error handling

### Verification (Session 3)
- [ ] Full compilation without errors
- [ ] All tests passing
- [ ] API endpoints functional via testing
- [ ] Authentication flow working
- [ ] Database operations tested

## Success Criteria

1. **No compilation errors**: `go build .` succeeds
2. **Clean vet check**: `go vet ./...` passes
3. **Tests pass**: `go test ./...` succeeds
4. **Server starts**: Application runs without crashes
5. **API functional**: Basic CRUD operations work via HTTP testing
6. **Authentication working**: OIDC integration functional

## Post-Completion
Once backend is stable, frontend development can proceed using the API endpoints documented in FRONTEND_PLAN.md.