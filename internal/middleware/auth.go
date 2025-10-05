package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/auth"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
)

type contextKey string

const UserContextKey contextKey = "user"

// Logger middleware для логирования запросов
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware проверяет JWT токен
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.WriteError(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			utils.WriteError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// Добавляем данные пользователя в контекст
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AdminMiddleware проверяет права администратора
func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(UserContextKey).(*auth.Claims)
		if !ok {
			utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		if claims.Role != "admin" {
			utils.WriteError(w, http.StatusForbidden, "Admin access required")
			return
		}

		next.ServeHTTP(w, r)
	})
}
