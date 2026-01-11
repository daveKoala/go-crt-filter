package middleware

import "net/http"

// CustomHeaders adds custom headers to all responses
func CustomHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add custom headers that will be sent with every response
		w.Header().Set("X-Powered-By", "Go CT Scanner")
		w.Header().Set("X-API-Version", "v1.0.0")
		w.Header().Set("X-Request-Id", RequestIdGenerator())

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func RequestIdGenerator() string {
	return "hello world"
}
