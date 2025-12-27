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
		allowedOrigins = []string{
			"https://mini-games.duckdns.org",
			"http://localhost:3000",
			"http://localhost:8080",
		}
	}
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		for _, o := range allowedOrigins {
			if strings.EqualFold(origin, o) {
				allowed = true
				break
			}
		}

		if allowed {
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
