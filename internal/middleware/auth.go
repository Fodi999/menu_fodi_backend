package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	authservice "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/auth/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
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
		
		// Debug: log first 50 chars of header (for debugging, remove sensitive data)
		if authHeader != "" {
			preview := authHeader
			if len(preview) > 50 {
				preview = preview[:50] + "..."
			}
			log.Printf("🔍 Auth header preview: %q", preview)
		}

		if authHeader == "" {
			log.Printf("❌ No Authorization header for %s %s", r.Method, r.URL.Path)
			utils.WriteError(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		// Proper Bearer token extraction
		var tokenString string
		
		// Check if header starts with "Bearer "
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			// Standard format: "Bearer <token>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 {
				headerPreview := authHeader
				if len(headerPreview) > 50 {
					headerPreview = headerPreview[:50] + "..."
				}
				log.Printf("❌ Invalid Authorization format: expected 'Bearer <token>', got %d parts. Header preview: %q", len(parts), headerPreview)
				utils.WriteError(w, http.StatusUnauthorized, "Invalid Authorization format: expected 'Bearer <token>'")
				return
			}
			tokenString = strings.TrimSpace(parts[1])
		} else {
			// Try to handle case where token comes without "Bearer " prefix
			// This shouldn't happen, but some clients might send it incorrectly
			tokenString = strings.TrimSpace(authHeader)
			log.Printf("⚠️ Authorization header doesn't start with 'Bearer ', treating entire header as token (length: %d)", len(tokenString))
		}
		
		log.Printf("🎫 Token extracted, length: %d", len(tokenString))
		
		// Validate token length (JWT tokens are typically 200+ characters)
		if len(tokenString) < 50 {
			tokenPreview := tokenString
			if len(tokenPreview) > 20 {
				tokenPreview = tokenPreview[:20] + "..."
			}
			log.Printf("⚠️ Token seems too short (%d chars). Expected JWT token length: 200+. Token preview: %q", len(tokenString), tokenPreview)
			log.Printf("⚠️ This might indicate: 1) Frontend sending partial token, 2) Token being truncated by proxy/CDN, 3) Wrong token format")
			// Don't fail here, let JWT validation handle it, but log warning
		}

		claims, err := authservice.ValidateToken(tokenString)
		if err != nil {
			log.Printf("❌ JWT validation failed for %s %s: %v", r.Method, r.URL.Path, err)
			utils.WriteError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// 🔒 КРИТИЧНО: Проверяем статус пользователя в БД
		// JWT может быть валидным, но пользователь мог быть заблокирован после выдачи токена
		userRepo := &database.UserRepository{}
		user, err := userRepo.FindByID(claims.Subject)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Printf("❌ User not found: %s for %s %s", claims.Subject, r.Method, r.URL.Path)
				utils.WriteError(w, http.StatusUnauthorized, "User not found")
				return
			}
			log.Printf("❌ Database error checking user status: %v", err)
			utils.WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		// Проверяем статус пользователя
		if user.Status != models.UserStatusActive {
			log.Printf("❌ User %s is not active (status: %s) for %s %s", claims.Subject, user.Status, r.Method, r.URL.Path)
			utils.WriteError(w, http.StatusForbidden, "Account is not active")
			return
		}

		log.Printf("✅ Auth OK for user %s (status: %s): %s %s", claims.Subject, user.Status, r.Method, r.URL.Path)

		// Добавляем данные пользователя в контекст
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		ctx = context.WithValue(ctx, "userID", claims.Subject) // Добавляем userID для удобства (string key для совместимости)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AdminMiddleware проверяет права администратора (admin или super_admin)
func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(UserContextKey).(*authservice.Claims)
		if !ok {
			utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// Allow both admin and super_admin
		if claims.Role != models.RoleAdmin && claims.Role != models.RoleSuperAdmin {
			utils.WriteError(w, http.StatusForbidden, "Admin access required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// SuperAdminMiddleware проверяет права супер администратора (только super_admin)
// Используется для критичных операций: назначение ролей, удаление пользователей
func SuperAdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(UserContextKey).(*authservice.Claims)
		if !ok {
			utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		if claims.Role != models.RoleSuperAdmin {
			log.Printf("❌ SuperAdmin required: User %s has role '%s'", claims.Subject, claims.Role)
			utils.WriteError(w, http.StatusForbidden, "Super admin access required")
			return
		}

		log.Printf("✅ SuperAdmin access granted for user %s", claims.Subject)
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

		// Optional: Check user status if user exists (don't block if not found)
		userRepo := &database.UserRepository{}
		user, err := userRepo.FindByID(claims.Subject)
		if err == nil && user != nil {
			// User exists - check status
			if user.Status != models.UserStatusActive {
				// User is blocked - continue without auth (don't block request)
				log.Printf("⚠️ OptionalAuth: User %s is not active (status: %s), continuing without auth", claims.Subject, user.Status)
				next.ServeHTTP(w, r)
				return
			}
		}
		// If user not found, continue anyway (optional auth)

		// Valid token → add to context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		ctx = context.WithValue(ctx, "userID", claims.Subject)
		log.Printf("✅ OptionalAuth: User %s authenticated", claims.Subject)

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
	userID, err := uuid.Parse(claims.Subject)
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

			log.Printf("✅ RequireRole(%s): Access granted for user %s", role, claims.Subject)
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
