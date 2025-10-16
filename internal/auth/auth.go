package auth

import (
	"log"
	"net/http"
	"os"
)

const (
	APIKeyHeader = "X-API-Key"
	APIKeyEnv    = "GLOBAL_API_KEY"
)

func ValidateAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get(APIKeyHeader)
		if !isValidAPIKey(apiKey) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isValidAPIKey checks if the provided API key matches the expected key from environment variables
func isValidAPIKey(apiKey string) bool {
	validAPIKey := os.Getenv(APIKeyEnv)
	if apiKey != validAPIKey {
		log.Printf("Unauthorized access attempt with API key: %s", apiKey)
		return false
	}
	return true
}

// AddAPIKeyToRequest adds the API key to the request headers for internal service communication
func AddAPIKeyToRequest(req *http.Request) {
	apiKey := os.Getenv(APIKeyEnv)
	if apiKey != "" {
		req.Header.Set(APIKeyHeader, apiKey)
	}
}
