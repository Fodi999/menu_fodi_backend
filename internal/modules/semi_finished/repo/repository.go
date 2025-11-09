package repo

import (
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"gorm.io/gorm"
)

// SemiFinishedRepository handles database operations for semi-finished products
type SemiFinishedRepository struct {
	db *gorm.DB
}

// NewSemiFinishedRepository creates a new repository instance
func NewSemiFinishedRepository(db *gorm.DB) *SemiFinishedRepository {
	return &SemiFinishedRepository{db: db}
}

// GetAll retrieves all semi-finished products
func (r *SemiFinishedRepository) GetAll() ([]models.SemiFinished, error) {
	var semiFinished []models.SemiFinished
	if err := r.db.Omit("Ingredients").Order("created_at DESC").Find(&semiFinished).Error; err != nil {
		return nil, err
	}
	return semiFinished, nil
}

// GetByID retrieves a semi-finished product by ID
func (r *SemiFinishedRepository) GetByID(id string) (*models.SemiFinished, error) {
	var sf models.SemiFinished
	if err := r.db.First(&sf, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &sf, nil
}

// Create creates a new semi-finished product
func (r *SemiFinishedRepository) Create(sf *models.SemiFinished) error {
	return r.db.Create(sf).Error
}

// Update updates an existing semi-finished product
func (r *SemiFinishedRepository) Update(sf *models.SemiFinished) error {
	return r.db.Save(sf).Error
}

// Delete deletes a semi-finished product
func (r *SemiFinishedRepository) Delete(id string) error {
	return r.db.Delete(&models.SemiFinished{}, "id = ?", id).Error
}

// AddIngredient adds an ingredient to a semi-finished product
func (r *SemiFinishedRepository) AddIngredient(ingredient *models.SemiFinishedIngredient) error {
	return r.db.Create(ingredient).Error
}

// DeleteIngredients deletes all ingredients for a semi-finished product
func (r *SemiFinishedRepository) DeleteIngredients(semiFinishedID string) error {
	return r.db.Where("semi_finished_id = ?", semiFinishedID).Delete(&models.SemiFinishedIngredient{}).Error
}

// CheckNameExists checks if a semi-finished product with this name already exists
func (r *SemiFinishedRepository) CheckNameExists(name string, excludeID string) (bool, error) {
	var count int64
	query := r.db.Model(&models.SemiFinished{}).Where("LOWER(name) = LOWER(?)", name)
	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// CheckIngredientExists checks if an ingredient exists
func (r *SemiFinishedRepository) CheckIngredientExists(ingredientID string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Ingredient{}).Where("id = ?", ingredientID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
