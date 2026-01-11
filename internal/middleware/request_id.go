package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/logger"
	"go.uber.org/zap"
)

// RequestIDMiddleware adds a unique request ID to each request
// If the client sends X-Request-ID header, it uses that
// Otherwise, it generates a new UUID
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client sent X-Request-ID
		requestID := r.Header.Get("X-Request-ID")

		// If not provided, generate a new UUID
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add to context for handlers to use
		ctx := context.WithValue(r.Context(), logger.RequestIDKey, requestID)

		// Add to response header so client can track the request
		w.Header().Set("X-Request-ID", requestID)

		// Log the incoming request with request ID
		logger.Log.Info("Incoming request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
			zap.String("request_id", requestID),
		)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID retrieves request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(logger.RequestIDKey).(string); ok {
		return id
	}
	return "unknown"
}
