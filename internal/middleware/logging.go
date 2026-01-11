package middleware

import (
	"log"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logging logs HTTP requests with method, path, status code, and duration
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the response writer to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // default
		}

		// Call the next handler
		next.ServeHTTP(wrapped, r)

		// Log request details
		duration := time.Since(start)
		log.Printf(
			"%s %s %d %v",
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration,
		)
	})
}
