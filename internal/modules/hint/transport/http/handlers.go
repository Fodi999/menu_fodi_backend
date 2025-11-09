package http

import (
	"encoding/json"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/hint/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/hint/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/httpx"
)

// HintHandlers обработчики подсказок
type HintHandlers struct {
	service *service.HintService
}

// NewHintHandlers создает новый обработчик
func NewHintHandlers(srv *service.HintService) *HintHandlers {
	return &HintHandlers{service: srv}
}

// GetHint обрабатывает запрос на получение подсказки
// @Summary Get product hints
// @Description Search products and get hints based on user question
// @Tags Hints
// @Accept json
// @Produce json
// @Param request body dto.HintRequest true "User question"
// @Success 200 {object} dto.HintResponse
// @Failure 400 {object} httpx.ErrorResponse
// @Router /api/hint [post]
func (h *HintHandlers) GetHint(w http.ResponseWriter, r *http.Request) {
	var req dto.HintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Invalid request format")
		return
	}

	if req.Question == "" {
		httpx.BadRequest(w, "Question is required")
		return
	}

	// Поиск продуктов
	products, err := h.service.SearchProducts(req.Question)
	if err != nil {
		httpx.InternalError(w, "Failed to search products")
		return
	}

	// Генерация подсказки
	hint := h.service.GenerateHint(products)

	httpx.Success(w, dto.HintResponse{
		Hint:              hint,
		SuggestedProducts: products,
	})
}
