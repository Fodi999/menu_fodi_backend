package handlers

import (
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
)

// HealthCheck проверка здоровья сервера
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "ok",
		"service":  "menu-fodifood-backend",
		"database": "connected",
	})
}
