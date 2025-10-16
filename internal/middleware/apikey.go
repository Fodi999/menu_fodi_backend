package middleware

import (
	"net/http"
	"os"

	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
)

// APIKeyMiddleware проверяет X-API-Key заголовок
func APIKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Пропускаем OPTIONS запросы для CORS
		if r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		apiKey := r.Header.Get("X-API-Key")
		expectedKey := os.Getenv("ELIXIR_API_KEY")

		if expectedKey == "" {
			expectedKey = "supersecret" // fallback для dev
		}

		if apiKey != expectedKey {
			utils.RespondWithJSON(w, http.StatusUnauthorized, map[string]interface{}{
				"status":  "error",
				"message": "Invalid or missing API key",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
