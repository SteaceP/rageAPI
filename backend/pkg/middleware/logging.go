package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func LoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Start timer
			start := time.Now()

			// Create a custom response writer to capture status code
			crw := &customResponseWriter{
				ResponseWriter: w,
				status:         http.StatusOK, // Default status
			}

			// Call the next handler
			next.ServeHTTP(crw, r)

			// Log request details
			logger.Info("HTTP Request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", crw.status),
				zap.Duration("latency", time.Since(start)),
				zap.String("remote_addr", r.RemoteAddr),
			)
		})
	}
}

// customResponseWriter wraps http.ResponseWriter to capture status code
type customResponseWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader captures the HTTP status code and writes it to the ResponseWriter.
func (crw *customResponseWriter) WriteHeader(status int) {
	crw.status = status
	crw.ResponseWriter.WriteHeader(status)
}
