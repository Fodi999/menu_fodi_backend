package database

import (
	"fmt"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserSavedRecipeRepository handles database operations for saved recipes
type UserSavedRecipeRepository struct {
	db *gorm.DB
}

// NewUserSavedRecipeRepository creates a new repository instance
func NewUserSavedRecipeRepository(db *gorm.DB) *UserSavedRecipeRepository {
	return &UserSavedRecipeRepository{db: db}
}

// SaveRecipe saves a recipe for a user (upsert)
func (r *UserSavedRecipeRepository) SaveRecipe(userID, recipeID string, servings int, source string) (*models.UserSavedRecipe, error) {
	// Check if recipe already exists
	var existingRecipe models.UserSavedRecipe
	result := r.db.Where("user_id = ? AND recipe_id = ?::uuid", userID, recipeID).First(&existingRecipe)

	if result.Error == nil {
		// Recipe exists, update it
		existingRecipe.Servings = servings
		existingRecipe.Source = source
		existingRecipe.SavedAt = time.Now()

		if err := r.db.Save(&existingRecipe).Error; err != nil {
			return nil, fmt.Errorf("failed to update saved recipe: %w", err)
		}

		return &existingRecipe, nil
	}

	// Recipe doesn't exist, create new one
	if result.Error != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing recipe: %w", result.Error)
	}

	savedRecipe := &models.UserSavedRecipe{
		ID:       uuid.New().String(),
		UserID:   userID,
		RecipeID: recipeID,
		Servings: servings,
		Source:   source,
		SavedAt:  time.Now(),
	}

	if err := r.db.Create(savedRecipe).Error; err != nil {
		return nil, fmt.Errorf("failed to create saved recipe: %w", err)
	}

	return savedRecipe, nil
}

// GetSavedRecipes retrieves all saved recipes for a user with recipe details
func (r *UserSavedRecipeRepository) GetSavedRecipes(userID string) ([]models.UserSavedRecipe, error) {
	var savedRecipes []models.UserSavedRecipe

	// Query saved recipes - use explicit SELECT to cast UUID to text
	result := r.db.
		Select("id::text as id, user_id, recipe_id::text as recipe_id, servings, source, saved_at").
		Where("user_id = ?", userID).
		Order("saved_at DESC").
		Find(&savedRecipes)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get saved recipes: %w", result.Error)
	}

	// Manually load recipe details for each saved recipe
	for i := range savedRecipes {
		var recipe models.RecipeCatalog
		if err := r.db.Where("id = ?::uuid", savedRecipes[i].RecipeID).First(&recipe).Error; err != nil {
			// If recipe not found, skip it but don't fail the whole query
			if err != gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("failed to load recipe %s: %w", savedRecipes[i].RecipeID, err)
			}
			continue
		}
		savedRecipes[i].Recipe = &recipe
	}

	return savedRecipes, nil
}

// GetSavedRecipe retrieves a single saved recipe
func (r *UserSavedRecipeRepository) GetSavedRecipe(userID, recipeID string) (*models.UserSavedRecipe, error) {
	var savedRecipe models.UserSavedRecipe

	result := r.db.
		Preload("Recipe").
		Where("user_id = ? AND recipe_id = ?::uuid", userID, recipeID).
		First(&savedRecipe)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get saved recipe: %w", result.Error)
	}

	return &savedRecipe, nil
}

// DeleteSavedRecipe removes a saved recipe
func (r *UserSavedRecipeRepository) DeleteSavedRecipe(userID, recipeID string) error {
	result := r.db.
		Where("user_id = ? AND recipe_id = ?::uuid", userID, recipeID).
		Delete(&models.UserSavedRecipe{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete saved recipe: %w", result.Error)
	}

	return nil
}

// GetSavedRecipeIDs returns array of recipe IDs saved by user (for excluding from recommendations)
func (r *UserSavedRecipeRepository) GetSavedRecipeIDs(userID string) ([]string, error) {
	var recipeIDs []string

	result := r.db.Model(&models.UserSavedRecipe{}).
		Where("user_id = ?", userID).
		Pluck("recipe_id", &recipeIDs)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get saved recipe IDs: %w", result.Error)
	}

	return recipeIDs, nil
}

// FindSavedRecipe finds a saved recipe by user ID and recipe ID
func (r *UserSavedRecipeRepository) FindSavedRecipe(userID, recipeID string) (*models.UserSavedRecipe, error) {
	var savedRecipe models.UserSavedRecipe

	result := r.db.Where("user_id = ? AND recipe_id = ?::uuid", userID, recipeID).First(&savedRecipe)

	if result.Error == gorm.ErrRecordNotFound {
		return nil, nil // Not found, return nil without error
	}

	if result.Error != nil {
		return nil, fmt.Errorf("failed to find saved recipe: %w", result.Error)
	}

	return &savedRecipe, nil
}

// SavedRecipeFilters represents filter options for saved recipes
type SavedRecipeFilters struct {
	Category    string // salad, main, soup, pizza, sushi, dessert, breakfast, drink
	Country     string
	Difficulty  string // easy, medium, hard
	CookedOnly  bool   // Only show cooked recipes
	UncokedOnly bool   // Only show recipes not yet cooked
}

// GetSavedRecipesWithFilters retrieves saved recipes with filters via JOIN
func (r *UserSavedRecipeRepository) GetSavedRecipesWithFilters(userID string, filters SavedRecipeFilters) ([]models.UserSavedRecipe, error) {
	var savedRecipes []models.UserSavedRecipe

	// Build query with JOIN to Recipe table (single source of truth for category)
	query := r.db.
		Table("user_saved_recipes usr").
		Select("usr.id::text as id, usr.user_id, usr.recipe_id::text as recipe_id, usr.servings, usr.source, usr.saved_at, usr.cooked_at").
		Joins("INNER JOIN \"Recipe\" r ON r.id = usr.recipe_id::uuid").
		Where("usr.user_id = ?", userID)

	// Apply filters from Recipe table (not from saved_recipes!)
	if filters.Category != "" {
		query = query.Where("r.category = ?", filters.Category)
	}
	if filters.Country != "" {
		query = query.Where("r.country = ?", filters.Country)
	}
	if filters.Difficulty != "" {
		query = query.Where("r.difficulty = ?", filters.Difficulty)
	}

	// Filter by cooked status
	if filters.CookedOnly {
		query = query.Where("usr.cooked_at IS NOT NULL")
	}
	if filters.UncokedOnly {
		query = query.Where("usr.cooked_at IS NULL")
	}

	// Order by save date
	query = query.Order("usr.saved_at DESC")

	result := query.Find(&savedRecipes)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get filtered saved recipes: %w", result.Error)
	}

	// Manually load recipe details for each saved recipe
	for i := range savedRecipes {
		var recipe models.RecipeCatalog
		if err := r.db.Where("id = ?::uuid", savedRecipes[i].RecipeID).First(&recipe).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("failed to load recipe %s: %w", savedRecipes[i].RecipeID, err)
			}
			continue
		}
		savedRecipes[i].Recipe = &recipe
	}

	return savedRecipes, nil
}
