package http

import (
	"encoding/json"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	authservice "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/auth/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/recipes_admin/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/recipes_admin/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/go-chi/chi/v5"
)

type RecipeAdminHandlers struct {
	service *service.RecipeAdminService
}

func NewRecipeAdminHandlers() *RecipeAdminHandlers {
	return &RecipeAdminHandlers{
		service: service.NewRecipeAdminService(),
	}
}

// CreateDraft - POST /api/admin/recipes
// МИНИМАЛЬНАЯ валидация, создает draft
func (h *RecipeAdminHandlers) CreateDraft(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	// Get authenticated user
	claims, ok := r.Context().Value(middleware.UserContextKey).(*authservice.Claims)
	if !ok || claims.Subject == "" {
		utils.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Create draft recipe
	recipe, err := h.service.CreateDraft(claims.Subject, &req)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create draft recipe")
		return
	}

	response := dto.CreateRecipeResponse{
		ID:            recipe.ID,
		LocalName:     recipe.LocalName,
		CanonicalName: recipe.CanonicalName,
		Status:        recipe.Status,
		Category:      recipe.Category,
		Difficulty:    recipe.Difficulty,
		AuthorID:      recipe.AuthorID,
		CreatedAt:     recipe.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"status": "success",
		"data":   response,
	})
}

// UpdateDraft - PATCH /api/admin/recipes/{id}
// Обновление draft рецепта (partial updates, включая ingredients/steps)
func (h *RecipeAdminHandlers) UpdateDraft(w http.ResponseWriter, r *http.Request) {
	recipeID := chi.URLParam(r, "id")

	var req dto.UpdateRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	recipe, err := h.service.UpdateDraft(recipeID, &req)
	if err != nil {
		if err.Error() == "recipe not found" {
			utils.RespondWithError(w, http.StatusNotFound, "Recipe not found")
			return
		}
		if err.Error() == "can only update draft recipes" {
			utils.RespondWithError(w, http.StatusForbidden, "Can only update draft recipes")
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update recipe")
		return
	}

	response := dto.UpdateRecipeResponse{
		ID:        recipe.ID,
		Title:     recipe.Title,
		Status:    recipe.Status,
		UpdatedAt: recipe.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   response,
	})
}

// Publish - POST /api/admin/recipes/{id}/publish
// ПОЛНАЯ валидация перед публикацией
func (h *RecipeAdminHandlers) Publish(w http.ResponseWriter, r *http.Request) {
	recipeID := chi.URLParam(r, "id")

	var req dto.PublishRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	recipe, warnings, err := h.service.Publish(recipeID, &req)
	if err != nil {
		if err.Error() == "recipe not found" {
			utils.RespondWithError(w, http.StatusNotFound, "Recipe not found")
			return
		}
		if err.Error() == "at least 1 ingredient required for publishing" ||
			err.Error() == "at least 1 step required for publishing" {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to publish recipe")
		return
	}

	response := dto.PublishRecipeResponse{
		ID:               recipe.ID,
		Title:            recipe.Title,
		Status:           recipe.Status,
		PublishedAt:      recipe.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		IngredientsCount: len(req.Ingredients),
		StepsCount:       len(req.Steps),
		Warnings:         warnings,
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   response,
	})
}

// Archive - POST /api/admin/recipes/{id}/archive
func (h *RecipeAdminHandlers) Archive(w http.ResponseWriter, r *http.Request) {
	recipeID := chi.URLParam(r, "id")

	if err := h.service.Archive(recipeID); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to archive recipe")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Recipe archived successfully",
	})
}

// GetDrafts - GET /api/admin/recipes/drafts
func (h *RecipeAdminHandlers) GetDrafts(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*authservice.Claims)
	if !ok || claims.Subject == "" {
		utils.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	recipes, err := h.service.GetDrafts(claims.Subject)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch drafts")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   recipes,
		"meta": map[string]interface{}{
			"total": len(recipes),
		},
	})
}
