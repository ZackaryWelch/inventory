package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/nishiki/backend-go/app/http/httputil"
	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/services"
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

func (m *AuthMiddleware) RequireAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				m.logger.Warn("Missing Authorization header")
				httputil.Error(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			// Remove "Bearer " prefix
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == authHeader {
				m.logger.Warn("Invalid Authorization header format")
				httputil.Error(w, http.StatusUnauthorized, "invalid authorization header format")
				return
			}

			// Validate token
			claims, err := m.authService.ValidateToken(r.Context(), token)
			if err != nil {
				m.logger.Warn("Token validation failed", slog.Any("error", err))
				httputil.Error(w, http.StatusUnauthorized, "invalid token")
				return
			}

			// Try to get existing user
			user, err := m.authService.GetUserFromClaims(r.Context(), claims)
			if err != nil {
				// User doesn't exist, create new user
				user, err = m.authService.CreateUserFromClaims(r.Context(), claims)
				if err != nil {
					m.logger.Error("Failed to create user from claims", slog.Any("error", err))
					httputil.Error(w, http.StatusInternalServerError, "failed to process user")
					return
				}

				m.logger.Info("Created new user from token",
					slog.String("user_id", user.ID().String()),
					slog.String("username", user.Username().String()))
			}

			// Store user, claims, and token in context
			r = httputil.SetContextValue(r, httputil.AuthUserKey, user)
			r = httputil.SetContextValue(r, httputil.AuthClaimsKey, claims)
			r = httputil.SetContextValue(r, httputil.AuthTokenKey, token)

			next.ServeHTTP(w, r)
		})
	}
}

func (m *AuthMiddleware) OptionalAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == authHeader {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := m.authService.ValidateToken(r.Context(), token)
			if err != nil {
				m.logger.Debug("Optional auth token validation failed", slog.Any("error", err))
				next.ServeHTTP(w, r)
				return
			}

			user, err := m.authService.GetUserFromClaims(r.Context(), claims)
			if err != nil {
				m.logger.Debug("Optional auth user lookup failed", slog.Any("error", err))
				next.ServeHTTP(w, r)
				return
			}

			r = httputil.SetContextValue(r, httputil.AuthUserKey, user)
			r = httputil.SetContextValue(r, httputil.AuthClaimsKey, claims)
			r = httputil.SetContextValue(r, httputil.AuthTokenKey, token)
			next.ServeHTTP(w, r)
		})
	}
}

// GetCurrentUser extracts the authenticated user from the request context
func GetCurrentUser(r *http.Request) (*entities.User, bool) {
	user := httputil.GetContextValue(r, httputil.AuthUserKey)
	if user == nil {
		return nil, false
	}

	authUser, ok := user.(*entities.User)
	if !ok {
		return nil, false
	}

	return authUser, true
}

// GetCurrentClaims extracts the auth claims from the request context
func GetCurrentClaims(r *http.Request) (*services.AuthClaims, bool) {
	claims := httputil.GetContextValue(r, httputil.AuthClaimsKey)
	if claims == nil {
		return nil, false
	}

	authClaims, ok := claims.(*services.AuthClaims)
	if !ok {
		return nil, false
	}

	return authClaims, true
}

// GetCurrentToken extracts the auth token from the request context
func GetCurrentToken(r *http.Request) (string, bool) {
	token := httputil.GetContextValue(r, httputil.AuthTokenKey)
	if token == nil {
		return "", false
	}

	authToken, ok := token.(string)
	if !ok {
		return "", false
	}

	return authToken, true
}
