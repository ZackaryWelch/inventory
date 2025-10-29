package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/services"
)

const (
	AuthUserKey   = "auth_user"
	AuthClaimsKey = "auth_claims"
	AuthTokenKey  = "auth_token"
)

type AuthMiddleware struct {
	authService services.AuthService
	logger      *slog.Logger
}

func NewAuthMiddleware(authService services.AuthService, logger *slog.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		logger:      logger,
	}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			m.logger.Warn("Missing Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			m.logger.Warn("Invalid Authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		// Validate token
		claims, err := m.authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			m.logger.Warn("Token validation failed", slog.Any("error", err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Try to get existing user
		user, err := m.authService.GetUserFromClaims(c.Request.Context(), claims)
		if err != nil {
			// User doesn't exist, create new user
			user, err = m.authService.CreateUserFromClaims(c.Request.Context(), claims)
			if err != nil {
				m.logger.Error("Failed to create user from claims", slog.Any("error", err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process user"})
				c.Abort()
				return
			}

			m.logger.Info("Created new user from token",
				slog.String("user_id", user.ID().String()),
				slog.String("username", user.Username().String()))
		}

		// Store user, claims, and token in context
		c.Set(AuthUserKey, user)
		c.Set(AuthClaimsKey, claims)
		c.Set(AuthTokenKey, token)

		c.Next()
	}
}

func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			c.Next()
			return
		}

		claims, err := m.authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			m.logger.Debug("Optional auth token validation failed", slog.Any("error", err))
			c.Next()
			return
		}

		user, err := m.authService.GetUserFromClaims(c.Request.Context(), claims)
		if err != nil {
			m.logger.Debug("Optional auth user lookup failed", slog.Any("error", err))
			c.Next()
			return
		}

		c.Set(AuthUserKey, user)
		c.Set(AuthClaimsKey, claims)
		c.Set(AuthTokenKey, token)
		c.Next()
	}
}

// GetCurrentUser extracts the authenticated user from the Gin context
func GetCurrentUser(c *gin.Context) (*entities.User, bool) {
	user, exists := c.Get(AuthUserKey)
	if !exists {
		return nil, false
	}

	authUser, ok := user.(*entities.User)
	if !ok {
		return nil, false
	}

	return authUser, true
}

// GetCurrentClaims extracts the auth claims from the Gin context
func GetCurrentClaims(c *gin.Context) (*services.AuthClaims, bool) {
	claims, exists := c.Get(AuthClaimsKey)
	if !exists {
		return nil, false
	}

	authClaims, ok := claims.(*services.AuthClaims)
	if !ok {
		return nil, false
	}

	return authClaims, true
}

// GetCurrentToken extracts the auth token from the Gin context
func GetCurrentToken(c *gin.Context) (string, bool) {
	token, exists := c.Get(AuthTokenKey)
	if !exists {
		return "", false
	}

	authToken, ok := token.(string)
	if !ok {
		return "", false
	}

	return authToken, true
}
