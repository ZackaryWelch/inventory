package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/nishiki/backend-go/app/http/httputil"
)

func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			path := r.URL.Path
			raw := r.URL.RawQuery

			// Wrap response writer to capture status and size
			rw := httputil.NewResponseWriter(w)

			// Process request
			next.ServeHTTP(rw, r)

			// Calculate latency
			latency := time.Since(start)

			// Build full path
			if raw != "" {
				path = path + "?" + raw
			}

			// Get client IP
			clientIP := r.RemoteAddr
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				clientIP = forwarded
			} else if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
				clientIP = realIP
			}

			// Log request details
			logger.Info("HTTP Request",
				slog.String("method", r.Method),
				slog.String("path", path),
				slog.Int("status", rw.Status()),
				slog.Duration("latency", latency),
				slog.String("ip", clientIP),
				slog.String("user-agent", r.UserAgent()),
				slog.Int("size", rw.Size()),
			)
		})
	}
}

func ErrorHandlingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Wrap response writer to check if written
			rw := httputil.NewResponseWriter(w)

			next.ServeHTTP(rw, r)

			// If status is 5xx and no body was written, write a generic error
			if rw.Status() >= 500 && !rw.Written() {
				logger.Error("Request error",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.Int("status", rw.Status()),
				)

				httputil.JSON(rw, http.StatusInternalServerError, map[string]string{
					"error": "internal server error",
					"code":  "INTERNAL_ERROR",
				})
			}
		})
	}
}
