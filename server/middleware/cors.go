package middleware

import (
	"net/http"
	"os"
	"strings"
)

var allowedOrigins []string

func init() {
	// Get allowed origins from environment variable or use defaults
	originsEnv := os.Getenv("ALLOWED_ORIGINS")
	if originsEnv != "" {
		allowedOrigins = strings.Split(originsEnv, ",")
		for i, o := range allowedOrigins {
			allowedOrigins[i] = strings.TrimSpace(o)
		}
	} else {
		// Production default only - set ALLOWED_ORIGINS env var for development
		allowedOrigins = []string{
			"https://mini-games.duckdns.org",
		}
	}
}

// GetAllowedOrigins returns the list of allowed origins for use by other packages
func GetAllowedOrigins() []string {
	return allowedOrigins
}

// IsOriginAllowed checks if the given origin is in the allowed list
func IsOriginAllowed(origin string) bool {
	for _, o := range allowedOrigins {
		if strings.EqualFold(origin, o) {
			return true
		}
	}
	return false
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Check if origin is allowed
		if IsOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
