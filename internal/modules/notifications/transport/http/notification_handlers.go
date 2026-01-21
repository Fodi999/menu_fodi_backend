package http

import (
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/notifications/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/go-chi/chi/v5"
)

type NotificationHandlers struct {
	service service.NotificationService
}

func NewNotificationHandlers(svc service.NotificationService) *NotificationHandlers {
	return &NotificationHandlers{service: svc}
}

// ============================================================================
// API ENDPOINTS - ПРАВИЛЬНАЯ АРХИТЕКТУРА
// ============================================================================

// GetNotifications GET /api/notifications - получить уведомления по уровням
// Возвращает: { critical: [], warning: [], info: [] }
func (h *NotificationHandlers) GetNotifications(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	groups, err := h.service.GetNotificationsByLevel(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch notifications")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, groups)
}

// GetUnreadCount GET /api/notifications/unread-count - количество непрочитанных
// Возвращает: { critical: 1, warning: 2, info: 0, total: 3 }
// ❗ total = critical + warning (info НЕ считается для badge)
func (h *NotificationHandlers) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	count, err := h.service.GetUnreadCount(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to count unread")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, count)
}

// MarkAsRead PATCH /api/notifications/:id/read - пометить как прочитанное
func (h *NotificationHandlers) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	notificationID := chi.URLParam(r, "id")

	if err := h.service.MarkAsRead(notificationID, userID); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Notification marked as read",
	})
}

// MarkAllAsRead POST /api/notifications/read-all - пометить все как прочитанные
func (h *NotificationHandlers) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	if err := h.service.MarkAllAsRead(userID); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "All notifications marked as read",
	})
}

// ResolveNotification POST /api/notifications/:id/resolve - пометить как решённое
// Используется когда пользователь использовал продукт или выбросил
func (h *NotificationHandlers) ResolveNotification(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	notificationID := chi.URLParam(r, "id")

	if err := h.service.ResolveNotification(notificationID, userID); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Notification resolved",
	})
}
