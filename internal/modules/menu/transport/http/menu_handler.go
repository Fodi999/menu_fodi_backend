
package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/menu/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ============================================================================
// HTTP HANDLER: Menu (Kitchen Pipeline)
// ============================================================================

type MenuHandler struct {
	service *service.MenuService
}

func NewMenuHandler(svc *service.MenuService) *MenuHandler {
	return &MenuHandler{service: svc}
}

// ============================================================================
// ENDPOINTS
// ============================================================================

// GetTodayMenu - GET /api/menu/today
// Returns all recipes user wants to cook today
func (h *MenuHandler) GetTodayMenu(w http.ResponseWriter, r *http.Request) {
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		utils.RespondError(w, http.StatusUnauthorized, "unauthorized", "user ID not found")
		return
	}
	userID := userIDPtr.String() // Convert UUID to string
	
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "pl"
	}
	
	items, err := h.service.GetTodayMenu(r.Context(), userID, lang)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "failed to get menu", err.Error())
		return
	}
	
	utils.RespondJSON(w, http.StatusOK, items)
}

// AddToMenu - POST /api/menu/today
// Add recipe to today's menu
func (h *MenuHandler) AddToMenu(w http.ResponseWriter, r *http.Request) {
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		utils.RespondError(w, http.StatusUnauthorized, "unauthorized", "user ID not found")
		return
	}
	userID := userIDPtr.String() // Convert UUID to string
	
	// Parse request body
	var req models.AddToMenuRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid request", err.Error())
		return
	}
	
	// Parse recipe ID (UUID or canonical_name)
	recipeID, err := uuid.Parse(req.RecipeID)
	if err != nil {
		// TODO: Support canonical_name lookup
		utils.RespondError(w, http.StatusBadRequest, "invalid recipe_id", "must be a valid UUID")
		return
	}
	
	// Default servings
	servings := req.Servings
	if servings == 0 {
		servings = 1
	}
	
	// Add to menu
	var notes *string
	if req.Notes != "" {
		notes = &req.Notes
	}
	
	item, err := h.service.AddToMenu(r.Context(), userID, recipeID, servings, notes)
	if err != nil {
		// Check if it's InsufficientIngredientsError
		var insufficientErr *service.InsufficientIngredientsError
		if errors.As(err, &insufficientErr) {
			// 409 Conflict: Cannot cook - missing ingredients
			utils.RespondJSON(w, http.StatusConflict, map[string]interface{}{
				"error":              "insufficient_ingredients",
				"message":            "Cannot add to menu: missing ingredients in fridge",
				"missing_ingredients": insufficientErr.MissingIngredients,
			})
			return
		}
		
		// Other errors
		utils.RespondError(w, http.StatusInternalServerError, "failed to add to menu", err.Error())
		return
	}
	
	utils.RespondJSON(w, http.StatusCreated, item)
}

// StartCooking - POST /api/menu/{id}/start
// Start cooking a menu item
func (h *MenuHandler) StartCooking(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid menu item ID", err.Error())
		return
	}
	
	// Parse optional servings adjustment
	var req models.StartCookingRequest
	_ = json.NewDecoder(r.Body).Decode(&req) // Ignore error, servings is optional
	
	var servings *int
	if req.Servings > 0 {
		servings = &req.Servings
	}
	
	// Start cooking
	if err := h.service.StartCooking(r.Context(), itemID, servings); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "failed to start cooking", err.Error())
		return
	}
	
	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Started cooking",
	})
}

// CompleteCooking - POST /api/menu/{id}/complete
// Complete cooking a menu item
func (h *MenuHandler) CompleteCooking(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid menu item ID", err.Error())
		return
	}
	
	// Parse optional actual servings
	var req models.CompleteCookingRequest
	_ = json.NewDecoder(r.Body).Decode(&req) // Ignore error, optional
	
	var actualServings *int
	if req.ActualServings > 0 {
		actualServings = &req.ActualServings
	}
	
	// Complete cooking
	if err := h.service.CompleteCooking(r.Context(), itemID, actualServings); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "failed to complete cooking", err.Error())
		return
	}
	
	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Cooking completed",
	})
}

// CancelMenuItem - POST /api/menu/{id}/cancel
// Cancel a menu item
func (h *MenuHandler) CancelMenuItem(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid menu item ID", err.Error())
		return
	}
	
	if err := h.service.CancelMenuItem(r.Context(), itemID); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "failed to cancel menu item", err.Error())
		return
	}
	
	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Menu item cancelled",
	})
}

// DeleteMenuItem - DELETE /api/menu/{id}
// Delete a menu item
func (h *MenuHandler) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid menu item ID", err.Error())
		return
	}
	
	if err := h.service.DeleteMenuItem(r.Context(), itemID); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "failed to delete menu item", err.Error())
		return
	}
	
	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Menu item deleted",
	})
}

// GetHistory - GET /api/menu/history?limit=30
// Get completed menu items (history)
func (h *MenuHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		utils.RespondError(w, http.StatusUnauthorized, "unauthorized", "user ID not found in context")
		return
	}
	
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "en"
	}
	
	// Parse optional limit parameter
	limit := 30 // default
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}
	
	history, err := h.service.GetHistory(r.Context(), userID.String(), lang, limit)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "failed to get history", err.Error())
		return
	}
	
	utils.RespondJSON(w, http.StatusOK, history)
}
