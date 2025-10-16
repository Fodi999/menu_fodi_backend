package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
)

type HintRequest struct {
	Question string `json:"question"`
}

// HintHandler обрабатывает запросы на получение подсказок
func HintHandler(w http.ResponseWriter, r *http.Request) {
	var req HintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "Invalid request format",
		})
		return
	}

	if req.Question == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "Question is required",
		})
		return
	}

	// Поиск продуктов по вопросу
	var products []models.Product
	question := strings.ToLower(req.Question)
	
	if err := database.DB.Where("LOWER(name) LIKE ?", "%"+question+"%").
		Or("LOWER(category) LIKE ?", "%"+question+"%").
		Limit(5).
		Find(&products).Error; err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Failed to search products",
		})
		return
	}

	// Формируем подсказку
	hint := "Вот что я нашел по вашему запросу:"
	if len(products) == 0 {
		hint = "К сожалению, ничего не найдено. Попробуйте другой запрос."
	} else {
		for i, p := range products {
			hint += fmt.Sprintf("\n%d. %s - %.2f сом", i+1, p.Name, p.Price)
		}
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
		"data": map[string]interface{}{
			"hint":              hint,
			"suggested_products": products,
		},
	})
}
