# Authentik Setup Guide for Nishiki

This guide helps you configure Authentik to work with the Nishiki application.

## Prerequisites

- Running Authentik server instance
- Admin access to Authentik
- Nishiki application URLs (frontend and backend)

## 1. Create OAuth2/OIDC Provider

1. **Log into Authentik Admin Interface**
2. **Navigate to Applications → Providers**
3. **Click "Create" and select "OAuth2/OpenID Provider"**

### Provider Configuration

```yaml
Name: nishiki-provider
Authorization flow: default-authentication-flow (or your custom flow)
Client type: Confidential
Client ID: nishiki-backend
Client Secret: [Generate secure secret - save this for .env]
Redirect URIs: 
  - http://localhost:3000/auth/callback
  - http://localhost:3000/auth/silent-callback
  - https://your-domain.com/auth/callback (for production)
Signing Key: [Select your signing key]
```

### Advanced Settings

```yaml
Scopes: 
  - openid
  - email
  - profile
  - groups (create if not exists)
  
Subject mode: Based on the User's hashed ID
Include claims in id_token: ✓ (checked)
Token validity: 
  - Access Token: 10 minutes
  - Refresh Token: 30 days
```

## 2. Create Application

1. **Navigate to Applications → Applications**
2. **Click "Create"**

### Application Configuration

```yaml
Name: Nishiki
Slug: nishiki
Provider: nishiki-provider (select the provider created above)
Launch URL: http://localhost:3000
Icon: [Upload Nishiki logo if desired]
```

## 3. Create Frontend Provider (Optional)

If you want separate client credentials for frontend:

1. **Create another OAuth2/OIDC Provider**

### Frontend Provider Configuration

```yaml
Name: nishiki-frontend-provider
Client type: Public
Client ID: nishiki-frontend
Redirect URIs:
  - http://localhost:3000/auth/callback
  - http://localhost:3000/auth/silent-callback
  - https://your-domain.com/auth/callback (for production)
```

## 4. Configure User Groups (Optional)

Create groups for role-based access control:

1. **Navigate to Directory → Groups**
2. **Create groups as needed:**
   - `nishiki-admin` - Full admin access
   - `nishiki-user` - Regular user access
   - `nishiki-readonly` - Read-only access

## 5. Environment Configuration

Create `.env` file in the project root:

```bash
# Authentik Configuration
AUTHENTIK_URL=https://your-authentik-server.com
AUTHENTIK_CLIENT_ID=nishiki-backend
AUTHENTIK_CLIENT_SECRET=your-generated-client-secret

# Frontend Configuration
NEXT_PUBLIC_AUTHENTIK_URL=https://your-authentik-server.com
NEXT_PUBLIC_AUTHENTIK_CLIENT_ID=nishiki-frontend
NEXT_PUBLIC_APP_URL=http://localhost:3000
```

## 6. Test Configuration

### Backend Endpoints

Test that these URLs are accessible:

- `https://your-authentik-server.com/application/o/nishiki/jwks/` - JWKS endpoint
- `https://your-authentik-server.com/application/o/nishiki/.well-known/openid_configuration` - OIDC discovery

### Frontend Flow

1. Start the application: `./start-all-services.sh`
2. Navigate to `http://localhost:3000/login`
3. Click "Sign in with Authentik"
4. Should redirect to Authentik login
5. After login, should redirect back to application

## 7. Production Configuration

For production deployment:

### Update Redirect URIs

Update the provider redirect URIs to include your production domain:

```yaml
Redirect URIs:
  - https://your-production-domain.com/auth/callback
  - https://your-production-domain.com/auth/silent-callback
```

### Environment Variables

```bash
AUTHENTIK_URL=https://your-authentik-server.com
NEXT_PUBLIC_AUTHENTIK_URL=https://your-authentik-server.com
NEXT_PUBLIC_APP_URL=https://your-production-domain.com
```

### Security Considerations

- Use HTTPS for all production URLs
- Generate strong, unique client secrets
- Configure appropriate token expiration times
- Enable CORS for your frontend domain in Authentik
- Review and limit scopes to minimum required

## Troubleshooting

### Common Issues

1. **CORS Errors**: Ensure frontend domain is added to Authentik CORS settings
2. **Invalid Redirect URI**: Verify redirect URIs match exactly (including trailing slashes)
3. **Token Validation Errors**: Check JWKS endpoint accessibility and client secret
4. **User Creation Fails**: Verify required claims (sub, email, preferred_username) are included

### Debug Endpoints

- Backend health: `http://localhost:3001/auth/health`
- Get current user: `http://localhost:3001/auth/me` (requires valid token)
- Frontend auth status: Check browser localStorage for OIDC tokens

### Logs

Check logs for authentication issues:
- Backend: `docker compose logs backend-main`
- Frontend: Browser console for OIDC client errors
- Authentik: Check Authentik admin logs for authentication events

## Additional Resources

- [Authentik OAuth2 Provider Documentation](https://docs.goauthentik.io/docs/add-secure-apps/providers/oauth2)
- [OIDC Client TS Documentation](https://github.com/authts/oidc-client-ts)
- [OAuth 2.0 Authorization Code Flow](https://oauth.net/2/grant-types/authorization-code/)