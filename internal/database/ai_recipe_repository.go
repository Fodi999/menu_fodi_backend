package database

import (
	"encoding/json"
	"fmt"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/google/uuid"
)

// AIRecipeRepository handles database operations for AI-generated recipes
type AIRecipeRepository struct{}

// NewAIRecipeRepository creates a new repository instance
func NewAIRecipeRepository() *AIRecipeRepository {
	return &AIRecipeRepository{}
}

// SaveRecipe saves a completed recipe from Chef Mentor session
func (r *AIRecipeRepository) SaveRecipe(recipe *models.AIGeneratedRecipe) error {
	return DB.Create(recipe).Error
}

// GetRecipeByID retrieves a recipe by ID
func (r *AIRecipeRepository) GetRecipeByID(recipeID string) (*models.AIGeneratedRecipe, error) {
	var recipe models.AIGeneratedRecipe
	
	if err := DB.Where("id = ?", recipeID).First(&recipe).Error; err != nil {
		return nil, err
	}
	
	return &recipe, nil
}

// GetUserRecipes retrieves all recipes created by a user
func (r *AIRecipeRepository) GetUserRecipes(userID uuid.UUID, limit int, offset int) ([]models.AIGeneratedRecipe, error) {
	var recipes []models.AIGeneratedRecipe
	
	query := DB.Where("user_id = ?", userID).
		Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}
	
	if err := query.Find(&recipes).Error; err != nil {
		return nil, err
	}
	
	return recipes, nil
}

// FindSimilarRecipes finds recipes with matching ingredients
func (r *AIRecipeRepository) FindSimilarRecipes(ingredients []string, limit int) ([]models.AIGeneratedRecipe, error) {
	var recipes []models.AIGeneratedRecipe
	
	// Use PostgreSQL JSONB operators to find matching ingredients
	// This query finds recipes where ANY ingredient name matches
	query := DB.Where("ingredients @> ?", buildIngredientsQuery(ingredients)).
		Where("is_public = ?", true).
		Order("created_at DESC").
		Limit(limit)
	
	if err := query.Find(&recipes).Error; err != nil {
		return nil, err
	}
	
	return recipes, nil
}

// GetPublicRecipes retrieves all public recipes (marketplace)
func (r *AIRecipeRepository) GetPublicRecipes(category string, limit int, offset int) ([]models.AIGeneratedRecipe, error) {
	var recipes []models.AIGeneratedRecipe
	
	query := DB.Where("is_public = ?", true)
	
	if category != "" {
		query = query.Where("category = ?", category)
	}
	
	query = query.Order("created_at DESC").
		Limit(limit).
		Offset(offset)
	
	if err := query.Find(&recipes).Error; err != nil {
		return nil, err
	}
	
	return recipes, nil
}

// PublishRecipe makes a recipe public
func (r *AIRecipeRepository) PublishRecipe(recipeID string, shareURL string) error {
	return DB.Model(&models.AIGeneratedRecipe{}).
		Where("id = ?", recipeID).
		Updates(map[string]interface{}{
			"is_public": true,
			"share_url": shareURL,
		}).Error
}

// IncrementViews increments recipe view count
func (r *AIRecipeRepository) IncrementViews(recipeID string) error {
	return DB.Model(&models.AIGeneratedRecipe{}).
		Where("id = ?", recipeID).
		Update("views_count", DB.Raw("views_count + 1")).Error
}

// IncrementDownloads increments recipe download count
func (r *AIRecipeRepository) IncrementDownloads(recipeID string) error {
	return DB.Model(&models.AIGeneratedRecipe{}).
		Where("id = ?", recipeID).
		Update("downloads_count", DB.Raw("downloads_count + 1")).Error
}

// LikeRecipe adds a like to a recipe
func (r *AIRecipeRepository) LikeRecipe(recipeID uuid.UUID, userID uuid.UUID) error {
	// Check if already liked
	var existingLike models.RecipeLike
	err := DB.Where("recipe_id = ? AND user_id = ?", recipeID, userID).First(&existingLike).Error
	
	if err == nil {
		// Already liked
		return nil
	}
	
	// Create like
	like := models.RecipeLike{
		RecipeID: recipeID,
		UserID:   userID,
	}
	
	if err := DB.Create(&like).Error; err != nil {
		return err
	}
	
	// Increment likes count
	return DB.Model(&models.AIGeneratedRecipe{}).
		Where("id = ?", recipeID).
		Update("likes_count", DB.Raw("likes_count + 1")).Error
}

// UnlikeRecipe removes a like from a recipe
func (r *AIRecipeRepository) UnlikeRecipe(recipeID uuid.UUID, userID uuid.UUID) error {
	// Delete like
	if err := DB.Where("recipe_id = ? AND user_id = ?", recipeID, userID).
		Delete(&models.RecipeLike{}).Error; err != nil {
		return err
	}
	
	// Decrement likes count
	return DB.Model(&models.AIGeneratedRecipe{}).
		Where("id = ?", recipeID).
		Update("likes_count", DB.Raw("GREATEST(0, likes_count - 1)")).Error
}

// GetTopRecipes retrieves most popular recipes
func (r *AIRecipeRepository) GetTopRecipes(sortBy string, limit int) ([]models.AIGeneratedRecipe, error) {
	var recipes []models.AIGeneratedRecipe
	
	orderClause := "created_at DESC"
	switch sortBy {
	case "views":
		orderClause = "views_count DESC"
	case "likes":
		orderClause = "likes_count DESC"
	case "downloads":
		orderClause = "downloads_count DESC"
	}
	
	query := DB.Where("is_public = ?", true).
		Order(orderClause).
		Limit(limit)
	
	if err := query.Find(&recipes).Error; err != nil {
		return nil, err
	}
	
	return recipes, nil
}

// buildIngredientsQuery builds a JSONB query for ingredient matching
func buildIngredientsQuery(ingredients []string) string {
	// Build JSONB array for PostgreSQL query
	ingredientsJSON := "["
	for i, ing := range ingredients {
		ingredientsJSON += fmt.Sprintf(`{"name":"%s"}`, ing)
		if i < len(ingredients)-1 {
			ingredientsJSON += ","
		}
	}
	ingredientsJSON += "]"
	
	return ingredientsJSON
}

// CountUserRecipes counts total recipes for a user
func (r *AIRecipeRepository) CountUserRecipes(userID uuid.UUID) (int64, error) {
	var count int64
	
	if err := DB.Model(&models.AIGeneratedRecipe{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	
	return count, nil
}

// SearchRecipes searches recipes by title or ingredients
func (r *AIRecipeRepository) SearchRecipes(query string, limit int) ([]models.AIGeneratedRecipe, error) {
	var recipes []models.AIGeneratedRecipe
	
	searchQuery := fmt.Sprintf("%%%s%%", query)
	
	if err := DB.Where("is_public = ? AND (title ILIKE ? OR description ILIKE ?)", 
		true, searchQuery, searchQuery).
		Order("likes_count DESC, created_at DESC").
		Limit(limit).
		Find(&recipes).Error; err != nil {
		return nil, err
	}
	
	return recipes, nil
}

// ConvertRecipeDraftToAI converts a RecipeDraft to AIGeneratedRecipe
func ConvertRecipeDraftToAI(draft interface{}, sessionID uuid.UUID, userID *uuid.UUID, language string) (*models.AIGeneratedRecipe, error) {
	// Convert draft to JSON
	draftJSON, err := json.Marshal(draft)
	if err != nil {
		return nil, err
	}
	
	// Parse draft
	var draftMap map[string]interface{}
	if err := json.Unmarshal(draftJSON, &draftMap); err != nil {
		return nil, err
	}
	
	// Extract fields
	title, _ := draftMap["title"].(string)
	category, _ := draftMap["category"].(string)
	difficulty, _ := draftMap["difficulty"].(string)
	time, _ := draftMap["time"].(float64)
	portions, _ := draftMap["portions"].(float64)
	cost, _ := draftMap["cost"].(float64)
	yieldVal, _ := draftMap["yield"].(float64)
	calories, _ := draftMap["calories"].(float64)
	protein, _ := draftMap["protein"].(float64)
	fats, _ := draftMap["fats"].(float64)
	carbs, _ := draftMap["carbs"].(float64)
	
	// Build ingredients JSONB
	ingredientsJSON := models.JSONB{}
	if ingredients, ok := draftMap["ingredients"].([]interface{}); ok {
		for _, ing := range ingredients {
			if ingMap, ok := ing.(map[string]interface{}); ok {
				ingredientsJSON[fmt.Sprintf("%v", ingMap["name"])] = ingMap
			}
		}
	}
	
	// Build nutrition JSONB
	nutritionJSON := models.JSONB{
		"calories": calories,
		"protein":  protein,
		"fats":     fats,
		"carbs":    carbs,
	}
	
	// Build steps JSONB
	stepsJSON := models.JSONB{}
	if steps, ok := draftMap["steps"].([]interface{}); ok {
		for i, step := range steps {
			stepsJSON[fmt.Sprintf("step_%d", i+1)] = step
		}
	}
	
	return &models.AIGeneratedRecipe{
		SessionID:   sessionID,
		UserID:      userID,
		Title:       title,
		Category:    category,
		Difficulty:  difficulty,
		Language:    language,
		Ingredients: ingredientsJSON,
		Steps:       stepsJSON,
		Nutrition:   nutritionJSON,
		Cost:        cost,
		Yield:       int(yieldVal),
		Time:        int(time),
		Portions:    int(portions),
		IsPublic:    false,
	}, nil
}
