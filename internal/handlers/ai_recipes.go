package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// GetMyRecipesHandler retrieves user's AI-generated recipes
// GET /api/ai/recipes/my
func GetMyRecipesHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Get userID from auth context
	userIDStr := r.URL.Query().Get("userId")
	if userIDStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "User ID required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Pagination
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit == 0 {
		limit = 10
	}

	repo := database.NewAIRecipeRepository()
	recipes, err := repo.GetUserRecipes(userID, limit, offset)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch recipes")
		return
	}

	total, _ := repo.CountUserRecipes(userID)

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"recipes": recipes,
			"total":   total,
			"limit":   limit,
			"offset":  offset,
		},
	})
}

// GetRecipeByIDHandler retrieves a single recipe
// GET /api/ai/recipes/{id}
func GetRecipeByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	recipeID := vars["id"]

	repo := database.NewAIRecipeRepository()
	recipe, err := repo.GetRecipeByID(recipeID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Recipe not found")
		return
	}

	// Increment view count
	repo.IncrementViews(recipeID)

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   recipe,
	})
}

// FindSimilarRecipesHandler finds recipes with similar ingredients
// GET /api/ai/recipes/similar?ingredients=rice,eel,avocado
func FindSimilarRecipesHandler(w http.ResponseWriter, r *http.Request) {
	ingredientsStr := r.URL.Query().Get("ingredients")
	if ingredientsStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Ingredients parameter required")
		return
	}

	// Parse ingredients
	ingredients := []string{}
	for _, ing := range r.URL.Query()["ingredients"] {
		ingredients = append(ingredients, ing)
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 5
	}

	repo := database.NewAIRecipeRepository()
	recipes, err := repo.FindSimilarRecipes(ingredients, limit)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to find similar recipes")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"recipes":     recipes,
			"ingredients": ingredients,
			"count":       len(recipes),
		},
	})
}

// GetMarketplaceRecipesHandler retrieves all public recipes
// GET /api/ai/recipes/marketplace
func GetMarketplaceRecipesHandler(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit == 0 {
		limit = 20
	}

	repo := database.NewAIRecipeRepository()
	recipes, err := repo.GetPublicRecipes(category, limit, offset)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch marketplace recipes")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"recipes":  recipes,
			"category": category,
			"limit":    limit,
			"offset":   offset,
		},
	})
}

// PublishRecipeHandler makes a recipe public
// POST /api/ai/recipes/{id}/publish
func PublishRecipeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	recipeID := vars["id"]

	repo := database.NewAIRecipeRepository()

	// Generate share URL
	shareURL := "recipe-" + recipeID[:8]

	if err := repo.PublishRecipe(recipeID, shareURL); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to publish recipe")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "success",
		"message":  "Recipe published successfully",
		"shareUrl": shareURL,
	})
}

// LikeRecipeHandler adds a like to a recipe
// POST /api/ai/recipes/{id}/like
func LikeRecipeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	recipeID, err := uuid.Parse(vars["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid recipe ID")
		return
	}

	// TODO: Get userID from auth context
	userIDStr := r.URL.Query().Get("userId")
	if userIDStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "User ID required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	repo := database.NewAIRecipeRepository()
	if err := repo.LikeRecipe(recipeID, userID); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to like recipe")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Recipe liked",
	})
}

// GetTopRecipesHandler retrieves most popular recipes
// GET /api/ai/recipes/top?sort=views|likes|downloads
func GetTopRecipesHandler(w http.ResponseWriter, r *http.Request) {
	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "likes"
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 10
	}

	repo := database.NewAIRecipeRepository()
	recipes, err := repo.GetTopRecipes(sortBy, limit)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch top recipes")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"recipes": recipes,
			"sortBy":  sortBy,
			"limit":   limit,
		},
	})
}

// SearchRecipesHandler searches recipes by title
// GET /api/ai/recipes/search?q=sushi
func SearchRecipesHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Search query required")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
	}

	repo := database.NewAIRecipeRepository()
	recipes, err := repo.SearchRecipes(query, limit)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to search recipes")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"recipes": recipes,
			"query":   query,
			"count":   len(recipes),
		},
	})
}

// UpdateRecipeImageHandler updates recipe image URL
// POST /api/ai/recipes/{id}/image
func UpdateRecipeImageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	recipeIDStr := vars["id"]

	recipeID, err := uuid.Parse(recipeIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid recipe ID")
		return
	}

	// Parse request body
	var req struct {
		ImageURL string `json:"imageUrl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.ImageURL == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Image URL is required")
		return
	}

	// Update recipe in database
	repo := database.NewAIRecipeRepository()
	recipe, err := repo.GetRecipeByID(recipeID.String())
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Recipe not found")
		return
	}

	// Update image URL
	recipe.ImageURL = req.ImageURL
	if err := database.DB.Model(recipe).Update("image_url", req.ImageURL).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update recipe")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "success",
		"message":  "Recipe image updated",
		"imageUrl": req.ImageURL,
	})
}
