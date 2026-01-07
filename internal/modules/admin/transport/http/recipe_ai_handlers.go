package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	authservice "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/auth/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/admin/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
)

// CreateRecipeWithAI - POST /api/admin/recipes/create-ai
// Создает рецепт через AI (сохраняет в БД)
func (h *AdminHandlers) CreateRecipeWithAI(w http.ResponseWriter, r *http.Request) {
	// 🛡️ Защита от panic
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("🚨 PANIC in CreateRecipeWithAI: %v\n", r)
			utils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
	}()

	// Получаем claims из контекста (ПРАВИЛЬНЫЙ способ через UserContextKey)
	claims, ok := r.Context().Value(middleware.UserContextKey).(*authservice.Claims)
	if !ok || claims == nil {
		fmt.Printf("❌ Claims not found in context\n")
		utils.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}
	
	userID := claims.UserID
	if userID == "" {
		utils.RespondWithError(w, http.StatusUnauthorized, "User ID is empty")
		return
	}

	fmt.Printf("✅ UserID from context: %s\n", userID)

	// Парсим запрос
	var req service.CreateRecipeAIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Printf("❌ Invalid request body: %v\n", err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Валидация
	if req.Title == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "title is required")
		return
	}
	if len(req.Ingredients) == 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "ingredients are required")
		return
	}
	if req.RawCookingText == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "rawCookingText is required")
		return
	}

	fmt.Printf("🎯 CreateRecipeWithAI: title='%s', ingredients=%d, user=%s\n",
		req.Title, len(req.Ingredients), userID)

	// Создаем рецепт через AI
	recipe, err := h.service.CreateRecipeWithAI(req, userID)
	if err != nil {
		// Проверяем тип ошибки для правильного HTTP кода
		errMsg := err.Error()
		if errMsg == "recipe with similar name already exists" {
			utils.RespondWithError(w, http.StatusConflict, errMsg)
			return
		}
		if errMsg == "AI generation failed" || errMsg == "failed to parse AI JSON" {
			utils.RespondWithError(w, http.StatusUnprocessableEntity, "AI could not process recipe")
			return
		}
		fmt.Printf("❌ CreateRecipeWithAI failed: %v\n", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create recipe")
		return
	}

	fmt.Printf("✅ Recipe created via AI: %s [%s]\n", recipe.Title, recipe.ID)
	utils.RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Recipe created via AI",
		"data":    recipe,
	})
}

// PreviewRecipeWithAI - POST /api/admin/recipes/preview-ai
// Генерирует AI-рецепт БЕЗ сохранения (preview mode)
func (h *AdminHandlers) PreviewRecipeWithAI(w http.ResponseWriter, r *http.Request) {
	// 🛡️ Защита от panic
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("🚨 PANIC in PreviewRecipeWithAI: %v\n", r)
			utils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
	}()

	// Парсим запрос
	var req service.CreateRecipeAIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Printf("❌ Invalid request body: %v\n", err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Валидация
	if req.Title == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "title is required")
		return
	}
	if len(req.Ingredients) == 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "ingredients are required")
		return
	}
	if req.RawCookingText == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "rawCookingText is required")
		return
	}

	fmt.Printf("🔍 PreviewRecipeWithAI: title='%s', ingredients=%d\n", req.Title, len(req.Ingredients))

	// Генерируем preview через AI
	preview, err := h.service.PreviewRecipeWithAI(req)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "AI generation failed" || errMsg == "failed to parse AI JSON" {
			utils.RespondWithError(w, http.StatusUnprocessableEntity, "AI could not process recipe")
			return
		}
		fmt.Printf("❌ PreviewRecipeWithAI failed: %v\n", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to generate preview")
		return
	}

	fmt.Printf("✅ Preview generated: %d steps, %d min\n", len(preview.Steps), preview.TimeMinutes)
	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Recipe preview generated",
		"data":    preview,
	})
}
