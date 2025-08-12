# Missing Authentication Endpoints for Frontend Integration

## Overview

The nishiki-backend-go has a solid authentication infrastructure but is missing two critical proxy endpoints that the frontend requires for OIDC authentication flow. These endpoints exist in the TypeScript backend and are essential for avoiding CORS issues when integrating with Authentik.

## Current State

✅ **What's Working:**
- JWT token validation with Authentik JWKS
- Authentication middleware (`auth_middleware.go`)
- User provisioning and context handling
- Protected route enforcement
- `/auth/me` endpoint for user info

❌ **What's Missing:**
- `/auth/oidc-config` - OIDC discovery proxy endpoint
- `/auth/token` - Token exchange proxy endpoint

## Required Endpoints

### 1. `/auth/oidc-config` (GET)

**Purpose:** Proxy OIDC discovery configuration from Authentik to avoid CORS issues.

**Location:** Add to `app/http/controllers/auth_controller.go`

**Implementation Requirements:**
```go
// GET /auth/oidc-config
func (ac *AuthController) GetOIDCConfig(c *gin.Context) {
    // 1. Fetch OIDC discovery from Authentik
    discoveryURL := fmt.Sprintf("%s/application/o/%s/.well-known/openid-configuration", 
        config.AuthentikURL, config.ProviderName)
    
    // 2. Make HTTP request to Authentik discovery endpoint
    response, err := http.Get(discoveryURL)
    // Handle SSL verification based on config.SkipTLSVerification
    
    // 3. Parse the JSON response
    var oidcConfig map[string]interface{}
    json.Unmarshal(responseBody, &oidcConfig)
    
    // 4. Replace token_endpoint with our proxy
    backendURL := os.Getenv("BACKEND_URL") // e.g., "http://localhost:3001"
    oidcConfig["token_endpoint"] = fmt.Sprintf("%s/auth/token", backendURL)
    
    // 5. Return modified config with CORS headers
    c.Header("Access-Control-Allow-Origin", "*")
    c.JSON(200, oidcConfig)
}
```

**Route Registration:** Add to `app/http/routes/routes.go`:
```go
// Add to auth group (no auth middleware needed)
authGroup.GET("/oidc-config", authController.GetOIDCConfig)
```

### 2. `/auth/token` (POST)

**Purpose:** Proxy token exchange requests to Authentik with proper credentials and CORS handling.

**Implementation Requirements:**
```go
// POST /auth/token
func (ac *AuthController) ProxyTokenExchange(c *gin.Context) {
    // 1. Get request body (authorization code, etc.)
    var requestBody map[string]interface{}
    c.ShouldBindJSON(&requestBody)
    
    // 2. Add client credentials from config
    requestBody["client_id"] = config.ClientID
    requestBody["client_secret"] = config.ClientSecret
    
    // 3. Forward to Authentik token endpoint
    authentikTokenURL := fmt.Sprintf("%s/application/o/%s/token/", 
        config.AuthentikURL, config.ProviderName)
    
    // 4. Make POST request to Authentik with form data
    // Handle SSL verification based on config.SkipTLSVerification
    
    // 5. Return Authentik response with CORS headers
    c.Header("Access-Control-Allow-Origin", "*")
    c.Header("Access-Control-Allow-Methods", "POST, OPTIONS")
    c.Header("Access-Control-Allow-Headers", "Content-Type")
    
    // Forward status code and response body from Authentik
    c.Data(response.StatusCode, "application/json", responseBody)
}
```

**Route Registration:** Add to `app/http/routes/routes.go`:
```go
// Add to auth group (no auth middleware needed)
authGroup.POST("/token", authController.ProxyTokenExchange)
```

## Implementation Details

### File Modifications Required

1. **`app/http/controllers/auth_controller.go`**
   - Add `GetOIDCConfig` method
   - Add `ProxyTokenExchange` method
   - Add HTTP client with TLS configuration support

2. **`app/http/routes/routes.go`**
   - Register new routes in auth group
   - Ensure no auth middleware on these endpoints

3. **Dependencies**
   - May need to add HTTP client utilities
   - Ensure proper error handling and logging

### Configuration Support

Use existing config values from `app.toml`:
```toml
[authentik]
url = "https://your-authentik-server"
provider_name = "nishiki"
client_id = "your-client-id"
client_secret = "your-client-secret"
skip_tls_verification = false
```

### Security Considerations

1. **CORS Headers:** Both endpoints need proper CORS configuration for frontend access
2. **Client Credentials:** Only the `/auth/token` endpoint should add client secrets
3. **TLS Verification:** Respect the `skip_tls_verification` config setting
4. **Error Handling:** Return appropriate error responses for failed Authentik calls

### Testing

After implementation, test with:
```bash
# Test OIDC config discovery
curl http://localhost:3001/auth/oidc-config

# Test token exchange (requires valid auth code)
curl -X POST http://localhost:3001/auth/token \
  -H "Content-Type: application/json" \
  -d '{"grant_type":"authorization_code","code":"...","redirect_uri":"..."}'
```

## Why These Endpoints Are Essential

The frontend uses `oidc-client-ts` which requires:
1. **OIDC Discovery:** To learn authorization/token/userinfo endpoints
2. **Token Exchange:** To convert authorization codes to JWT tokens

Without these proxy endpoints, the frontend cannot complete authentication due to CORS restrictions when calling Authentik directly from the browser.

These endpoints essentially make the Go backend compatible with the existing frontend authentication flow that currently works with the TypeScript backend.