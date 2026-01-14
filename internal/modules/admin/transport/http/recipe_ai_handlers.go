package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/admin/service"
	authservice "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/auth/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/go-chi/chi/v5"
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

	// ✅ CRITICAL FIX: Читаем язык из Accept-Language заголовка
	if req.Language == "" {
		acceptLang := r.Header.Get("Accept-Language")
		req.Language = normalizeLang(acceptLang)
		fmt.Printf("🌐 Language from Accept-Language: %s → %s\n", acceptLang, req.Language)
	} else {
		fmt.Printf("🌐 Language from body: %s\n", req.Language)
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

	// ✅ CRITICAL FIX: Читаем язык из Accept-Language заголовка
	if req.Language == "" {
		acceptLang := r.Header.Get("Accept-Language")
		req.Language = normalizeLang(acceptLang)
		fmt.Printf("🌐 Language from Accept-Language: %s → %s\n", acceptLang, req.Language)
	} else {
		fmt.Printf("🌐 Language from body: %s\n", req.Language)
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

// SaveEditedRecipe - POST /api/admin/recipes/save
// Сохраняет отредактированный пользователем рецепт
func (h *AdminHandlers) SaveEditedRecipe(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("🚨 PANIC in SaveEditedRecipe: %v\n", r)
			utils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
	}()

	// Получаем userID из контекста
	claims, ok := r.Context().Value(middleware.UserContextKey).(*authservice.Claims)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}
	userID := claims.UserID

	// Парсим запрос
	var req service.SaveEditedRecipeRequest
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
	if len(req.Steps) == 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "steps are required")
	}

	fmt.Printf("💾 SaveEditedRecipe: title='%s', ingredients=%d, steps=%d\n",
		req.Title, len(req.Ingredients), len(req.Steps))

	// Сохраняем через service
	recipe, err := h.service.SaveEditedRecipe(req, userID)
	if err != nil {
		errMsg := err.Error()

		// 🎯 УМНАЯ ОБРАБОТКА КОНФЛИКТА: Проверяем на дубликат названия
		if strings.Contains(errMsg, "already exists") {
			fmt.Printf("⚠️  Recipe name conflict detected: '%s'\n", req.Title)

			// 🌍 Генерируем мульти-язычные альтернативные названия через AI
			multilingualSuggestions, aiErr := h.service.GenerateMultilingualTitles(req.Title, req.Language)
			if aiErr != nil {
				fmt.Printf("⚠️  Failed to generate multilingual suggestions: %v\n", aiErr)
				// Fallback: простые варианты только на основном языке
				multilingualSuggestions = map[string][]string{
					req.Language: {
						req.Title + " (домашний рецепт)",
						req.Title + " (авторский)",
						req.Title + " на сковороде",
					},
				}
			}

			// Извлекаем canonical name из ошибки
			canonicalName := strings.ToLower(strings.ReplaceAll(req.Title, " ", "_"))

			// Возвращаем 409 с мульти-язычными подсказками
			utils.RespondWithJSON(w, http.StatusConflict, map[string]interface{}{
				"success": false,
				"code":    "RECIPE_NAME_EXISTS",
				"message": "Рецепт с таким названием уже существует",
				"conflict": map[string]interface{}{
					"canonicalName": canonicalName,
					"originalTitle": req.Title,
				},
				"suggestions": multilingualSuggestions, // Теперь это map[string][]string
			})
			return
		}

		fmt.Printf("❌ SaveEditedRecipe failed: %v\n", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to save recipe")
		return
	}

	fmt.Printf("✅ Recipe saved: %s [%s]\n", recipe.Title, recipe.ID)
	utils.RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Recipe saved successfully",
		"data":    recipe,
	})
}

// UpdateRecipe - PUT /api/admin/recipes/{id}
// Обновляет существующий рецепт
func (h *AdminHandlers) UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("🚨 PANIC in UpdateRecipe: %v\n", r)
			utils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
	}()

	// Получаем recipeID из URL
	recipeID := r.URL.Path[len("/api/admin/recipes/"):]
	if recipeID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "recipe ID is required")
		return
	}

	// Парсим запрос
	var req service.UpdateRecipeRequest
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
	if len(req.Steps) == 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "steps are required")
		return
	}

	fmt.Printf("🔄 UpdateRecipe: id=%s, title='%s'\n", recipeID, req.Title)

	// Обновляем через service
	recipe, err := h.service.UpdateRecipe(recipeID, req)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "recipe not found" {
			utils.RespondWithError(w, http.StatusNotFound, errMsg)
			return
		}
		fmt.Printf("❌ UpdateRecipe failed: %v\n", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update recipe")
		return
	}

	fmt.Printf("✅ Recipe updated: %s [%s]\n", recipe.Title, recipe.ID)
	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Recipe updated successfully",
		"data":    recipe,
	})
}

// DeleteRecipe - удалить рецепт (DELETE /api/admin/recipes/:id)
func (h *AdminHandlers) DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("🚨 PANIC in DeleteRecipe: %v\n", r)
			utils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
	}()

	// Получаем recipeID из URL
	recipeID := chi.URLParam(r, "id")
	if recipeID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "recipe ID is required")
		return
	}

	fmt.Printf("🗑️  DeleteRecipe: id=%s\n", recipeID)

	// Удаляем через service
	if err := h.service.DeleteRecipe(recipeID); err != nil {
		errMsg := err.Error()
		if errMsg == "recipe not found" {
			utils.RespondWithError(w, http.StatusNotFound, "Recipe not found")
			return
		}
		fmt.Printf("❌ DeleteRecipe failed: %v\n", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete recipe")
		return
	}

	fmt.Printf("✅ Recipe deleted: %s\n", recipeID)
	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Recipe deleted successfully",
	})
}
