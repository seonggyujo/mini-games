package middleware

import "net/http"

// Security adds security headers to all responses
func Security(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// Enable XSS filter in browsers
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Prevent caching of sensitive data
		w.Header().Set("Cache-Control", "no-store")

		next.ServeHTTP(w, r)
	})
}
