package httputil

import (
	"context"
	"net/http"
)

// ContextKey is a type for context keys to avoid collisions
type ContextKey string

const (
	// AuthUserKey is the context key for the authenticated user
	AuthUserKey ContextKey = "auth_user"
	// AuthClaimsKey is the context key for the auth claims
	AuthClaimsKey ContextKey = "auth_claims"
	// AuthTokenKey is the context key for the auth token
	AuthTokenKey ContextKey = "auth_token"
)

// SetContextValue returns a new request with the value added to its context
func SetContextValue(r *http.Request, key ContextKey, value interface{}) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), key, value))
}

// GetContextValue retrieves a value from the request context
func GetContextValue(r *http.Request, key ContextKey) interface{} {
	return r.Context().Value(key)
}
