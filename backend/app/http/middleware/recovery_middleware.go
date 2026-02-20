package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/nishiki/backend-go/app/http/httputil"
)

// RecoveryMiddleware returns a middleware that recovers from panics
func RecoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the panic with stack trace
					logger.Error("Panic recovered",
						slog.Any("error", err),
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
						slog.String("stack", string(debug.Stack())),
					)

					// Return 500 Internal Server Error
					httputil.Error(w, http.StatusInternalServerError, "internal server error")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
