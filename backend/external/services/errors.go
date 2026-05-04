package services

import "maps"

// AuthError is a structured error for auth-service operations. It carries a
// stable sentinel code and optional context data so callers (and log consumers)
// can inspect metadata without scraping it out of Error() strings. Use
// errors.Is() to match by code. Always derive copies with With()/Wrap() — never
// mutate a sentinel.
type AuthError struct {
	code  string
	data  map[string]any
	cause error
}

// Sentinel errors for the Authentik auth service.
var (
	ErrAuthConfigInvalid    = &AuthError{code: "auth config invalid"}
	ErrAuthentikUnreachable = &AuthError{code: "authentik unreachable"}
	ErrOIDCProviderInit     = &AuthError{code: "oidc provider init failed"}
)

func (e *AuthError) Error() string {
	if e.cause != nil {
		return e.code + ": " + e.cause.Error()
	}
	return e.code
}

func (e *AuthError) Is(target error) bool {
	t, ok := target.(*AuthError)
	if !ok {
		return false
	}
	return e.code == t.code
}

func (e *AuthError) Unwrap() error { return e.cause }

// Code returns the stable error code string.
func (e *AuthError) Code() string { return e.code }

// Data returns the optional structured context data.
func (e *AuthError) Data() map[string]any { return e.data }

// With returns a copy with the given data merged into any existing data.
// Incoming keys win on collision so callers can layer context.
func (e *AuthError) With(data map[string]any) *AuthError {
	cp := *e
	if e.data == nil {
		cp.data = data
	} else {
		cp.data = make(map[string]any, len(e.data)+len(data))
		maps.Copy(cp.data, e.data)
		maps.Copy(cp.data, data)
	}
	return &cp
}

// Wrap returns a copy with the given cause, preserving existing data.
func (e *AuthError) Wrap(cause error) *AuthError {
	cp := *e
	cp.cause = cause
	return &cp
}
