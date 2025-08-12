# Go Backend API Specification for Frontend Integration

This document outlines the complete API specification required by the frontend application. The Go backend should implement these endpoints to ensure full compatibility with the refactored frontend code.

## Authentication & Headers

All API endpoints (except `/health`) require authentication via Bearer token:
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

## Base URL
All endpoints are relative to the configured `NEXT_PUBLIC_API_BASE_URL` environment variable.

---

## 🔐 Authentication Endpoints

### GET /auth/me
Get current authenticated user information.

**Response 200:**
```json
{
  "id": "string",
  "name": "string", 
  "email": "string"
}
```

**Response 401:** Unauthorized
**Response 500:** Internal server error

---

## 👥 Group Endpoints

### GET /groups
Fetch all groups the authenticated user is a member of.

**Response 200:**
```json
[
  {
    "id": "string",
    "name": "string"
  }
]
```

**Response 401:** Unauthorized
**Response 500:** Internal server error

### POST /groups
Create a new group.

**Request Body:**
```json
{
  "name": "string"
}
```

**Response 201:**
```json
{
  "id": "string",
  "name": "string"
}
```

**Response 400:** Bad request (validation errors)
**Response 401:** Unauthorized
**Response 500:** Internal server error

---

## 📦 Container Endpoints

### GET /groups/{groupId}/containers
Fetch all containers for a specific group.

**Path Parameters:**
- `groupId` (string, required): The group ID

**Response 200:**
```json
[
  {
    "id": "string",
    "name": "string",
    "group": {
      "id": "string",
      "name": "string"
    },
    "foods": [
      {
        "id": "string",
        "name": "string",
        "quantity": number | null,
        "category": "string",
        "unit": "string" | null,
        "expiry": "2024-12-31T23:59:59Z" | null
      }
    ]
  }
]
```

**Response 401:** Unauthorized
**Response 403:** User not member of group
**Response 404:** Group not found
**Response 500:** Internal server error

### POST /containers
Create a new container.

**Request Body:**
```json
{
  "name": "string",
  "groupId": "string"
}
```

**Response 201:**
```json
{
  "id": "string",
  "name": "string",
  "group": {
    "id": "string",
    "name": "string"
  },
  "foods": []
}
```

**Response 400:** Bad request (validation errors)
**Response 401:** Unauthorized
**Response 403:** User not member of group
**Response 404:** Group not found
**Response 500:** Internal server error

---

## 🍎 Food Endpoints

### POST /foods
Add a new food item to a container.

**Request Body:**
```json
{
  "name": "string",
  "quantity": number | null,
  "category": "string",
  "unit": "string" | null,
  "expiry": "2024-12-31T23:59:59Z" | null,
  "containerId": "string"
}
```

**Response 201:**
```json
{
  "id": "string",
  "name": "string",
  "quantity": number | null,
  "category": "string",
  "unit": "string" | null,
  "expiry": "2024-12-31T23:59:59Z" | null
}
```

**Response 400:** Bad request (validation errors)
**Response 401:** Unauthorized
**Response 403:** User does not have access to container
**Response 404:** Container not found
**Response 500:** Internal server error

### PUT /foods/{foodId}
Update an existing food item.

**Path Parameters:**
- `foodId` (string, required): The food ID

**Request Body (all fields optional):**
```json
{
  "name": "string",
  "quantity": number | null,
  "category": "string", 
  "unit": "string" | null,
  "expiry": "2024-12-31T23:59:59Z" | null
}
```

**Response 200:**
```json
{
  "id": "string",
  "name": "string",
  "quantity": number | null,
  "category": "string",
  "unit": "string" | null,
  "expiry": "2024-12-31T23:59:59Z" | null
}
```

**Response 400:** Bad request (validation errors)
**Response 401:** Unauthorized
**Response 403:** User does not have access to food item
**Response 404:** Food not found
**Response 500:** Internal server error

### DELETE /foods/{foodId}
Delete a food item.

**Path Parameters:**
- `foodId` (string, required): The food ID

**Response 204:** No content (successful deletion)
**Response 401:** Unauthorized
**Response 403:** User does not have access to food item
**Response 404:** Food not found
**Response 500:** Internal server error

---

## 🏥 Health Check Endpoint

### GET /health
Service health check (no authentication required).

**Response 200:**
```json
{
  "status": "ok",
  "timestamp": "2024-12-31T23:59:59Z"
}
```

---

## 🚧 Missing Endpoints (TODO)

The following endpoints are referenced in the frontend but not yet implemented. They should be added to complete the functionality:

### GET /groups/{groupId}/users
Fetch all users/members of a specific group.

**Path Parameters:**
- `groupId` (string, required): The group ID

**Response 200:**
```json
[
  {
    "id": "string",
    "name": "string",
    "email": "string"
  }
]
```

### GET /users/{userId}
Fetch user information by ID.

**Path Parameters:**
- `userId` (string, required): The user ID

**Response 200:**
```json
{
  "id": "string",
  "name": "string",
  "email": "string"
}
```

### POST /groups/join
Join a group using an invitation hash.

**Request Body:**
```json
{
  "invitationHash": "string"
}
```

**Response 200:**
```json
{
  "groupId": "string"
}
```

### GET /groups/{groupId}
Fetch single group information.

**Path Parameters:**
- `groupId` (string, required): The group ID

**Response 200:**
```json
{
  "id": "string",
  "name": "string"
}
```

---

## 📋 Implementation Notes

### Date Handling
- All date fields should be in ISO 8601 format with UTC timezone
- `null` values are allowed for optional date fields
- Frontend will handle date parsing and localization

### Validation Rules
- Group names: 1-100 characters, non-empty
- Container names: 1-100 characters, non-empty
- Food names: 1-100 characters, non-empty
- Food categories: predefined list (see frontend `containerMapping.ts`)
- Food quantities: positive numbers or null
- Food units: string or null

### Error Response Format
All error responses should follow this structure:
```json
{
  "error": "string",
  "message": "string",
  "timestamp": "2024-12-31T23:59:59Z"
}
```

### Authorization
- Users can only access groups they are members of
- Users can only access containers belonging to their groups
- Users can only modify foods in containers they have access to

### CORS
Ensure CORS is properly configured for the frontend domain(s).

### Database Considerations
- Foods are embedded within containers (as per current MongoDB schema)
- Group membership should be validated for all group-related operations
- Consider implementing soft deletes for audit trails

---

## 🔧 Frontend Configuration

The frontend expects these environment variables:
```bash
NEXT_PUBLIC_API_BASE_URL=http://localhost:3001
NEXT_PUBLIC_AUTHENTIK_URL=https://your-authentik-server.com
NEXT_PUBLIC_AUTHENTIK_CLIENT_ID=nishiki-frontend
```

The backend should be configured to accept requests from the frontend domain and validate JWT tokens from the configured Authentik instance.