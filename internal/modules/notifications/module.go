package notifications

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/notifications/service"
	notificationhttp "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/notifications/transport/http"
)

// Module represents the notifications module
type Module struct {
	handlers *notificationhttp.NotificationHandlers
}

// NewModule creates a new notifications module
func NewModule(db *gorm.DB) *Module {
	// Initialize service
	svc := service.NewNotificationService(db)

	// Initialize handlers
	handlers := notificationhttp.NewNotificationHandlers(svc)

	return &Module{
		handlers: handlers,
	}
}

// RegisterRoutes registers notification routes
func (m *Module) RegisterRoutes(r chi.Router, jwtMiddleware func(http.Handler) http.Handler) {
	r.Route("/notifications", func(r chi.Router) {
		// All routes require authentication
		r.Use(jwtMiddleware)

		// GET /api/notifications - get notifications grouped by level
		// Returns: { critical: [], warning: [], info: [] }
		r.Get("/", m.handlers.GetNotifications)

		// GET /api/notifications/unread-count - get unread counts
		// Returns: { critical: 1, warning: 2, info: 0, total: 3 }
		r.Get("/unread-count", m.handlers.GetUnreadCount)

		// PATCH /api/notifications/{id}/read - mark notification as read
		r.Patch("/{id}/read", m.handlers.MarkAsRead)

		// POST /api/notifications/read-all - mark all notifications as read
		r.Post("/read-all", m.handlers.MarkAllAsRead)

		// POST /api/notifications/{id}/resolve - mark notification as resolved
		// Used when user takes action (uses product or discards)
		r.Post("/{id}/resolve", m.handlers.ResolveNotification)
	})
}
