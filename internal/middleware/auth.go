package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	authservice "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/auth/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/google/uuid"
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
		log.Printf("🔐 AuthMiddleware: %s %s", r.Method, r.URL.Path)
		
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Printf("❌ No Authorization header for %s %s", r.Method, r.URL.Path)
			utils.WriteError(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := authservice.ValidateToken(tokenString)
		if err != nil {
			log.Printf("❌ Invalid token for %s %s: %v", r.Method, r.URL.Path, err)
			utils.WriteError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		log.Printf("✅ Auth OK for user %s: %s %s", claims.UserID, r.Method, r.URL.Path)

		// Добавляем данные пользователя в контекст
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AdminMiddleware проверяет права администратора
func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(UserContextKey).(*authservice.Claims)
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

// GetUserID извлекает UUID пользователя из JWT claims в контексте
func GetUserID(r *http.Request) *uuid.UUID {
	claims, ok := r.Context().Value(UserContextKey).(*authservice.Claims)
	if !ok || claims == nil {
		return nil
	}

	// Parse UserID string to UUID
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil
	}

	return &userID
}
