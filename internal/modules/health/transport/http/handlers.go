package http

import (
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/health/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/httpx"
)

// HealthHandlers обрабатывает health check запросы
type HealthHandlers struct {
	service *service.HealthService
}

// NewHealthHandlers создает новые хендлеры
func NewHealthHandlers(service *service.HealthService) *HealthHandlers {
	return &HealthHandlers{service: service}
}

// HealthCheck godoc
// @Summary Health check
// @Description Проверка здоровья сервера и БД
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} dto.HealthResponse
// @Router /health [get]
func (h *HealthHandlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	dbStatus, _ := h.service.CheckHealth()
	
	httpx.Success(w, map[string]interface{}{
		"status": "ok",
		"data": map[string]interface{}{
			"service":  "menu-fodifood-backend",
			"database": dbStatus,
		},
	})
}
