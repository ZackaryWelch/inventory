package httputil

import "net/http"

// Middleware is a function that wraps an http.Handler
type Middleware func(http.Handler) http.Handler

// Chain chains multiple middleware functions together
// The first middleware in the list will be the outermost (first to execute)
func Chain(middlewares ...Middleware) Middleware {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}

// HandlerFunc converts an http.HandlerFunc to http.Handler for use with middleware
func HandlerFunc(fn http.HandlerFunc) http.Handler {
	return fn
}

// WrapHandler wraps a handler with middleware and returns an http.HandlerFunc
func WrapHandler(h http.Handler, middlewares ...Middleware) http.HandlerFunc {
	chain := Chain(middlewares...)
	return chain(h).ServeHTTP
}
