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

// GetRecipeByIDOrCanonical - получает рецепт по UUID или canonical_name
// Используется для гибкого поиска: сначала UUID, потом canonical_name
func (r *RecipeRepository) GetRecipeByIDOrCanonical(
	ctx context.Context,
	identifier string,
) (*models.RecipeCatalog, error) {
	var recipe models.RecipeCatalog

	// Try to parse as UUID first
	if parsedUUID, err := uuid.Parse(identifier); err == nil {
		// It's a valid UUID
		err = r.db.WithContext(ctx).
			Preload("Ingredients").
			Preload("Ingredients.Ingredient").
			Where("id = ?", parsedUUID).
			First(&recipe).Error

		if err == nil {
			return &recipe, nil
		}
		if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to get recipe by UUID: %w", err)
		}
	}

	// Not a UUID or not found, try canonical_name
	err := r.db.WithContext(ctx).
		Preload("Ingredients").
		Preload("Ingredients.Ingredient").
		Where("\"canonicalName\" = ?", identifier).
		First(&recipe).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("recipe not found: %s", identifier)
		}
		return nil, fmt.Errorf("failed to get recipe by canonical_name: %w", err)
	}

	return &recipe, nil
}
