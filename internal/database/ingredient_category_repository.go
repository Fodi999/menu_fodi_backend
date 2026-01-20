package database

import (
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
)

// IngredientCategoryRepository handles ingredient_categories table operations
type IngredientCategoryRepository struct{}

// GetAll returns all categories sorted by sort_order
func (r *IngredientCategoryRepository) GetAll() ([]models.IngredientCategory, error) {
	var categories []models.IngredientCategory
	result := DB.Order("sort_order ASC").Find(&categories)
	if result.Error != nil {
		return nil, result.Error
	}
	return categories, nil
}

// GetByKey returns a single category by key
func (r *IngredientCategoryRepository) GetByKey(key string) (*models.IngredientCategory, error) {
	var category models.IngredientCategory
	result := DB.Where("key = ?", key).First(&category)
	if result.Error != nil {
		return nil, result.Error
	}
	return &category, nil
}
