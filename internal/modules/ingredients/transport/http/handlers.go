package http

import (
	"encoding/json"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ingredients/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ingredients/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/httpx"
	"github.com/go-chi/chi/v5"
)

// IngredientsHandlers обработчики ингредиентов
type IngredientsHandlers struct {
	service *service.IngredientsService
}

// NewIngredientsHandlers создает новый обработчик
func NewIngredientsHandlers(srv *service.IngredientsService) *IngredientsHandlers {
	return &IngredientsHandlers{service: srv}
}

// GetAll получение всех ингредиентов
// @Summary Get all ingredients
// @Description Get all ingredients with stock data
// @Tags Ingredients
// @Produce json
// @Success 200 {array} models.StockItem
// @Failure 500 {object} httpx.ErrorResponse
// @Security BearerAuth
// @Router /api/ingredients [get]
func (h *IngredientsHandlers) GetAll(w http.ResponseWriter, r *http.Request) {
	stockItems, err := h.service.GetAll()
	if err != nil {
		httpx.InternalError(w, "Failed to fetch ingredients")
		return
	}

	httpx.Success(w, stockItems)
}

// GetOne получение одного ингредиента
// @Summary Get ingredient by ID
// @Description Get single ingredient with stock data
// @Tags Ingredients
// @Produce json
// @Param id path string true "Ingredient ID"
// @Success 200 {object} models.StockItem
// @Failure 404 {object} httpx.ErrorResponse
// @Security BearerAuth
// @Router /api/ingredients/{id} [get]
func (h *IngredientsHandlers) GetOne(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	stockItem, err := h.service.GetByID(id)
	if err != nil {
		httpx.NotFound(w, "Ingredient not found")
		return
	}

	httpx.Success(w, stockItem)
}

// Create создание нового ингредиента
// @Summary Create ingredient
// @Description Create new ingredient with stock data
// @Tags Ingredients
// @Accept json
// @Produce json
// @Param request body dto.CreateIngredientRequest true "Ingredient data"
// @Success 201 {object} models.StockItem
// @Failure 400 {object} httpx.ErrorResponse
// @Security BearerAuth
// @Router /api/ingredients [post]
func (h *IngredientsHandlers) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Invalid request")
		return
	}

	stockItem, err := h.service.Create(&req)
	if err != nil {
		httpx.BadRequest(w, err.Error())
		return
	}

	httpx.Created(w, stockItem)
}

// Update обновление ингредиента
// @Summary Update ingredient
// @Description Update existing ingredient
// @Tags Ingredients
// @Accept json
// @Produce json
// @Param id path string true "Ingredient ID"
// @Param request body dto.UpdateIngredientRequest true "Updated data"
// @Success 200 {object} models.StockItem
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Security BearerAuth
// @Router /api/ingredients/{id} [put]
func (h *IngredientsHandlers) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Invalid request")
		return
	}

	stockItem, err := h.service.Update(id, &req)
	if err != nil {
		httpx.NotFound(w, err.Error())
		return
	}

	httpx.Success(w, stockItem)
}

// Delete удаление ингредиента
// @Summary Delete ingredient
// @Description Delete ingredient by ID
// @Tags Ingredients
// @Produce json
// @Param id path string true "Ingredient ID"
// @Success 200 {object} map[string]string
// @Failure 500 {object} httpx.ErrorResponse
// @Security BearerAuth
// @Router /api/ingredients/{id} [delete]
func (h *IngredientsHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.Delete(id); err != nil {
		httpx.InternalError(w, "Failed to delete ingredient")
		return
	}

	httpx.Success(w, map[string]string{"message": "Ingredient deleted successfully"})
}

// GetStockMovements получение истории движений
// @Summary Get stock movements
// @Description Get stock movement history for ingredient
// @Tags Ingredients
// @Produce json
// @Param id path string true "Stock Item ID"
// @Success 200 {array} models.StockMovement
// @Failure 500 {object} httpx.ErrorResponse
// @Security BearerAuth
// @Router /api/ingredients/{id}/movements [get]
func (h *IngredientsHandlers) GetStockMovements(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	movements, err := h.service.GetStockMovements(id)
	if err != nil {
		httpx.InternalError(w, "Failed to fetch stock movements")
		return
	}

	httpx.Success(w, movements)
}

// Search поиск ингредиентов по имени (АВТОКОМПЛИТ)
// @Summary Search ingredients (autocomplete)
// @Description Search ingredients by name - используется ВСЕМИ пользователями
// @Tags Ingredients
// @Produce json
// @Param query query string true "Search query (min 1 char)"
// @Success 200 {array} models.Ingredient
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Security BearerAuth
// @Router /api/ingredients/search [get]
func (h *IngredientsHandlers) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")

	// Валидация: минимум 1 символ
	if len(query) < 1 {
		httpx.BadRequest(w, "Query parameter must be at least 1 character")
		return
	}

	ingredients, err := h.service.Search(query)
	if err != nil {
		httpx.InternalError(w, "Failed to search ingredients")
		return
	}

	// Унифицированный формат ответа
	httpx.Success(w, map[string]interface{}{
		"items": ingredients,
		"count": len(ingredients),
	})
}

// ListIngredients список ингредиентов с фильтрацией
// @Summary List ingredients catalog
// @Description Get ingredients list with optional category filter and search
// @Tags Ingredients
// @Produce json
// @Param category query string false "Category filter: protein, vegetable, dairy, grain, condiment, other"
// @Param search query string false "Search by name prefix"
// @Success 200 {array} models.Ingredient
// @Failure 500 {object} httpx.ErrorResponse
// @Security BearerAuth
// @Router /api/ingredients [get]
func (h *IngredientsHandlers) ListIngredients(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	search := r.URL.Query().Get("search")

	ingredients, err := h.service.List(category, search)
	if err != nil {
		httpx.InternalError(w, "Failed to list ingredients")
		return
	}

	httpx.Success(w, map[string]interface{}{
		"items": ingredients,
		"count": len(ingredients),
	})
}
