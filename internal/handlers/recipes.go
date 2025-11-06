package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/auth"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// GetAllPosts retrieves all recipes for the main feed (public recipe feed)
// GET /api/posts
func GetAllPosts(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	var recipes []models.Recipe

	// Fetch all recipes with author information, ordered by creation date (newest first)
	result := db.Preload("Author").Order("created_at DESC").Find(&recipes)
	if result.Error != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch recipes")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   recipes,
	})
}

// GetUserPosts retrieves recipes posted by a specific user for their profile
// GET /api/users/{id}/posts
func GetUserPosts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	db := database.GetDB()

	// Validate that user exists
	var user models.User
	if err := db.First(&user, "id = ?", userID).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	var recipes []models.Recipe
	result := db.Preload("Author").Where("author_id = ?", userID).
		Order("created_at DESC").
		Find(&recipes)

	if result.Error != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch user recipes")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   recipes,
	})
}

// CreateRecipePost creates a new recipe for the public feed
// POST /api/recipes
func CreateRecipePost(w http.ResponseWriter, r *http.Request) {
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

	// Get authorID from JWT token (set by auth middleware)
	claims, ok := r.Context().Value(middleware.UserContextKey).(*auth.Claims)
	if !ok || claims.UserID == "" {
		utils.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}
	authorID := claims.UserID

	// Validate required fields
	if input.Title == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Title is required")
		return
	}

	db := database.GetDB()

	// Validate that author exists
	var author models.User
	if err := db.First(&author, "id = ?", authorID).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Author not found")
		return
	}

	// Create recipe with UUID and metrics
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

	// Reload recipe with author information
	db.Preload("Author").First(&recipe, "id = ?", recipe.ID)

	utils.RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"status": "success",
		"data":   recipe,
	})
}

// UpdateRecipePost updates an existing recipe
// PUT /api/recipes/{id}
func UpdateRecipePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	recipeID := vars["id"]

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

	// Update fields if provided
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

	// Reload with author
	db.Preload("Author").First(&recipe, "id = ?", recipe.ID)

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   recipe,
	})
}

// DeleteRecipePost deletes a recipe
// DELETE /api/recipes/{id}
func DeleteRecipePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	recipeID := vars["id"]

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

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Recipe deleted successfully",
	})
}

// IncrementRecipeView increments view count and awards tokens to author
// POST /api/recipes/{id}/view
func IncrementRecipeView(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	recipeID := vars["id"]

	db := database.GetDB()

	var recipe models.Recipe
	if err := db.Preload("Author").First(&recipe, "id = ?", recipeID).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Recipe not found")
		return
	}

	// Increment views
	recipe.ViewsCount++

	// Award 1 ChefToken for every 10 views
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
