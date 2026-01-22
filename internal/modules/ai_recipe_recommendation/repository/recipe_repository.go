package repository

import (
	"context"
	"fmt"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================================================
// RECIPE REPOSITORY - единственный источник правды для Recipe
// ============================================================================

type RecipeRepository struct {
	db *gorm.DB
}

func NewRecipeRepository(db *gorm.DB) *RecipeRepository {
	return &RecipeRepository{db: db}
}

// GetRecipeWithRelations - получает рецепт со ВСЕМИ связями
// Это единственный метод для получения полных данных рецепта
func (r *RecipeRepository) GetRecipeWithRelations(
	ctx context.Context,
	recipeID uuid.UUID,
) (*models.RecipeCatalog, error) {
	var recipe models.RecipeCatalog

	err := r.db.WithContext(ctx).
		Preload("Ingredients").           // RecipeIngredient[]
		Preload("Ingredients.Ingredient"). // Ingredient details
		First(&recipe, "id = ?", recipeID).
		Error

	if err != nil {
		return nil, fmt.Errorf("failed to get recipe: %w", err)
	}

	return &recipe, nil
}

// GetAllRecipes - получает все рецепты из каталога со связями
func (r *RecipeRepository) GetAllRecipes(ctx context.Context) ([]models.RecipeCatalog, error) {
	var recipes []models.RecipeCatalog

	err := r.db.WithContext(ctx).
		Preload("Ingredients").
		Preload("Ingredients.Ingredient").
		Find(&recipes).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get recipes: %w", err)
	}

	return recipes, nil
}

// GetRecipesByIDs - получает несколько рецептов по ID
func (r *RecipeRepository) GetRecipesByIDs(
	ctx context.Context,
	recipeIDs []uuid.UUID,
) ([]models.RecipeCatalog, error) {
	var recipes []models.RecipeCatalog

	err := r.db.WithContext(ctx).
		Preload("Ingredients").
		Preload("Ingredients.Ingredient").
		Where("id IN ?", recipeIDs).
		Find(&recipes).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get recipes by IDs: %w", err)
	}

	return recipes, nil
}
