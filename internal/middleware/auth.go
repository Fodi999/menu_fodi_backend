package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
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
		log.Printf("📋 Auth header present: %v, length: %d", authHeader != "", len(authHeader))
		
		if authHeader == "" {
			log.Printf("❌ No Authorization header for %s %s", r.Method, r.URL.Path)
			utils.WriteError(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		// Proper Bearer token extraction
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			log.Printf("❌ Invalid Authorization format (not 'Bearer <token>') for %s %s", r.Method, r.URL.Path)
			utils.WriteError(w, http.StatusUnauthorized, "Invalid Authorization format")
			return
		}
		
		tokenString := strings.TrimSpace(parts[1])
		log.Printf("🎫 Token extracted, length: %d", len(tokenString))
		
		claims, err := authservice.ValidateToken(tokenString)
		if err != nil {
			log.Printf("❌ JWT validation failed for %s %s: %v", r.Method, r.URL.Path, err)
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

		if claims.Role != models.RoleAdmin {
			utils.WriteError(w, http.StatusForbidden, "Admin access required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// OptionalAuthMiddleware checks JWT token if present, but doesn't require it
// If token is valid, adds user to context. If no token or invalid, continues without user.
func OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		
		// No token? Continue without auth
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Extract Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			// Invalid format? Continue without auth (don't block)
			next.ServeHTTP(w, r)
			return
		}
		
		tokenString := strings.TrimSpace(parts[1])
		if tokenString == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Validate token
		claims, err := authservice.ValidateToken(tokenString)
		if err != nil {
			// Invalid token? Continue without auth (don't block)
			log.Printf("⚠️ OptionalAuth: Invalid token, continuing without auth: %v", err)
			next.ServeHTTP(w, r)
			return
		}

		// Valid token → add to context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		ctx = context.WithValue(ctx, "userID", claims.UserID)
		log.Printf("✅ OptionalAuth: User %s authenticated", claims.UserID)
		
		next.ServeHTTP(w, r.WithContext(ctx))
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

// RequireRole создаёт middleware для проверки конкретной роли пользователя
// Используется для разделения доступа: home_chef, pro_chef, admin
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(UserContextKey).(*authservice.Claims)
			if !ok {
				log.Printf("❌ RequireRole(%s): No user in context", role)
				utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			if claims.Role != role {
				log.Printf("❌ RequireRole(%s): User has role '%s', access denied", role, claims.Role)
				utils.WriteError(w, http.StatusForbidden, "Access denied: insufficient permissions")
				return
			}

			log.Printf("✅ RequireRole(%s): Access granted for user %s", role, claims.UserID)
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserFromContext извлекает Claims пользователя из контекста
func GetUserFromContext(r *http.Request) *authservice.Claims {
	claims, ok := r.Context().Value(UserContextKey).(*authservice.Claims)
	if !ok {
		return nil
	}
	return claims
}
