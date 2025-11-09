package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	authservice "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/auth/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type RecipeHandlers struct{}

func NewRecipeHandlers() *RecipeHandlers {
	return &RecipeHandlers{}
}

func (h *RecipeHandlers) GetAllPosts(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	var recipes []models.Recipe
	result := db.Preload("Author").Order("created_at DESC").Find(&recipes)
	if result.Error != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch recipes")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{"status": "success", "data": recipes})
}

func (h *RecipeHandlers) GetUserPosts(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	db := database.GetDB()

	var user models.User
	if err := db.First(&user, "id = ?", userID).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	var posts []models.RecipePost
	result := db.Where("user_id = ?", userID).Order("created_at DESC").Find(&posts)
	if result.Error != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch user posts")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{"status": "success", "data": posts})
}

func (h *RecipeHandlers) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title        string   `json:"title"`
		Description  string   `json:"description"`
		ImageUrl     string   `json:"imageUrl"`
		GrossWeight  *int     `json:"grossWeight"`
		NetWeight    *int     `json:"netWeight"`
		Calories     *int     `json:"calories"`
		Protein      *float64 `json:"protein"`
		Fats         *float64 `json:"fats"`
		Carbs        *float64 `json:"carbs"`
		RecipeYield  *int     `json:"yield"`
		Cost         *float64 `json:"cost"`
		TokensReward *int     `json:"tokensReward"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	claims, ok := r.Context().Value(middleware.UserContextKey).(*authservice.Claims)
	if !ok || claims.UserID == "" {
		utils.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}
	authorID := claims.UserID

	if input.Title == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Title is required")
		return
	}

	db := database.GetDB()
	var author models.User
	if err := db.First(&author, "id = ?", authorID).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Author not found")
		return
	}

	recipe := models.Recipe{
		ID:           uuid.New().String(),
		Title:        input.Title,
		Description:  input.Description,
		ImageUrl:     input.ImageUrl,
		AuthorID:     authorID,
		GrossWeight:  input.GrossWeight,
		NetWeight:    input.NetWeight,
		Calories:     input.Calories,
		Protein:      input.Protein,
		Fats:         input.Fats,
		Carbs:        input.Carbs,
		RecipeYield:  input.RecipeYield,
		Cost:         input.Cost,
		TokensReward: input.TokensReward,
		ViewsCount:   0,
		TokensEarned: 0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := db.Create(&recipe).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create recipe")
		return
	}

	db.Preload("Author").First(&recipe, "id = ?", recipe.ID)
	utils.RespondWithJSON(w, http.StatusCreated, map[string]interface{}{"status": "success", "data": recipe})
}

func (h *RecipeHandlers) UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	recipeID := chi.URLParam(r, "id")
	db := database.GetDB()

	var recipe models.Recipe
	if err := db.First(&recipe, "id = ?", recipeID).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Recipe not found")
		return
	}

	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		ImageUrl    string `json:"imageUrl"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	if input.Title != "" {
		recipe.Title = input.Title
	}
	if input.Description != "" {
		recipe.Description = input.Description
	}
	if input.ImageUrl != "" {
		recipe.ImageUrl = input.ImageUrl
	}
	recipe.UpdatedAt = time.Now()

	if err := db.Save(&recipe).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update recipe")
		return
	}

	db.Preload("Author").First(&recipe, "id = ?", recipe.ID)
	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{"status": "success", "data": recipe})
}

func (h *RecipeHandlers) DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	recipeID := chi.URLParam(r, "id")
	db := database.GetDB()

	var recipe models.Recipe
	if err := db.First(&recipe, "id = ?", recipeID).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Recipe not found")
		return
	}

	if err := db.Delete(&recipe).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete recipe")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{"status": "success", "message": "Recipe deleted successfully"})
}

func (h *RecipeHandlers) IncrementRecipeView(w http.ResponseWriter, r *http.Request) {
	recipeID := chi.URLParam(r, "id")
	db := database.GetDB()

	var recipe models.Recipe
	if err := db.Preload("Author").First(&recipe, "id = ?", recipeID).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Recipe not found")
		return
	}

	recipe.ViewsCount++
	if recipe.ViewsCount%10 == 0 {
		recipe.TokensEarned++
	}

	if err := db.Save(&recipe).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update views")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"viewsCount":   recipe.ViewsCount,
			"tokensEarned": recipe.TokensEarned,
		},
	})
}
