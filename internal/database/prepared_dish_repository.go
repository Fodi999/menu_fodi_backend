package database

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
)

type PreparedDishRepository struct {
	db *gorm.DB
}

func NewPreparedDishRepository(db *gorm.DB) *PreparedDishRepository {
	return &PreparedDishRepository{db: db}
}

// PreparedDishFilters for filtering prepared dishes
type PreparedDishFilters struct {
	Category      string // Filter by recipe category (via JOIN)
	AvailableOnly bool   // Only dishes with portions > 0
	ExpiredOnly   bool   // Only expired dishes
}

// Create creates a new prepared dish record
func (r *PreparedDishRepository) Create(dish *models.PreparedDish) error {
	// PostgreSQL: user_id is TEXT, recipe_id is UUID
	result := r.db.Exec(`
		INSERT INTO prepared_dishes (user_id, recipe_id, portions_available, portions_initial, prepared_at, expires_at, source)
		VALUES (?, ?::uuid, ?, ?, ?, ?, ?)
		RETURNING id::text, created_at, updated_at
	`, dish.UserID, dish.RecipeID, dish.PortionsAvailable, dish.PortionsInitial, dish.PreparedAt, dish.ExpiresAt, dish.Source)

	if result.Error != nil {
		return fmt.Errorf("failed to create prepared dish: %w", result.Error)
	}

	// Fetch the created record to get generated ID and timestamps
	return r.db.Raw(`
		SELECT id::text as id, user_id, recipe_id::text as recipe_id, 
		       portions_available, portions_initial, prepared_at, expires_at, source,
		       created_at, updated_at
		FROM prepared_dishes
		WHERE user_id = ? AND recipe_id = ?::uuid AND prepared_at = ?
		ORDER BY created_at DESC
		LIMIT 1
	`, dish.UserID, dish.RecipeID, dish.PreparedAt).Scan(dish).Error
}

// GetByUserID returns all prepared dishes for a user
func (r *PreparedDishRepository) GetByUserID(userID string) ([]models.PreparedDish, error) {
	var dishes []models.PreparedDish

	err := r.db.Raw(`
		SELECT id::text as id, user_id, recipe_id::text as recipe_id,
		       portions_available, portions_initial, prepared_at, expires_at, source,
		       created_at, updated_at
		FROM prepared_dishes
		WHERE user_id = ?
		ORDER BY prepared_at DESC
	`, userID).Scan(&dishes).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get prepared dishes: %w", err)
	}

	// Load recipe details for each dish
	for i := range dishes {
		var recipe models.RecipeCatalog
		err := r.db.Where("id = ?", dishes[i].RecipeID).First(&recipe).Error
		if err != nil {
			return nil, fmt.Errorf("failed to load recipe for dish %s: %w", dishes[i].ID, err)
		}
		dishes[i].Recipe = &recipe
	}

	return dishes, nil
}

// GetByUserIDWithFilters returns prepared dishes with optional filters
func (r *PreparedDishRepository) GetByUserIDWithFilters(userID string, filters PreparedDishFilters) ([]models.PreparedDish, error) {
	var dishes []models.PreparedDish

	// Build query with JOIN to Recipe table for category filtering
	query := r.db.
		Table("prepared_dishes pd").
		Joins("INNER JOIN \"Recipe\" r ON r.id = pd.recipe_id").
		Where("pd.user_id = ?", userID)

	// Apply filters
	if filters.Category != "" {
		query = query.Where("r.category = ?", filters.Category)
	}

	if filters.AvailableOnly {
		query = query.Where("pd.portions_available > 0")
	}

	if filters.ExpiredOnly {
		query = query.Where("pd.expires_at IS NOT NULL AND pd.expires_at < NOW()")
	}

	// Select fields and order by prepared_at DESC
	err := query.
		Select(`pd.id::text as id, pd.user_id, pd.recipe_id::text as recipe_id,
		        pd.portions_available, pd.portions_initial, pd.prepared_at, 
		        pd.expires_at, pd.source, pd.created_at, pd.updated_at`).
		Order("pd.prepared_at DESC").
		Scan(&dishes).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get prepared dishes with filters: %w", err)
	}

	// Load recipe details for each dish
	for i := range dishes {
		var recipe models.RecipeCatalog
		err := r.db.Where("id = ?", dishes[i].RecipeID).First(&recipe).Error
		if err != nil {
			return nil, fmt.Errorf("failed to load recipe for dish %s: %w", dishes[i].ID, err)
		}
		dishes[i].Recipe = &recipe
	}

	return dishes, nil
}

// FindByID finds a prepared dish by ID
func (r *PreparedDishRepository) FindByID(dishID string) (*models.PreparedDish, error) {
	var dish models.PreparedDish

	err := r.db.Raw(`
		SELECT id::text as id, user_id, recipe_id::text as recipe_id,
		       portions_available, portions_initial, prepared_at, expires_at, source,
		       created_at, updated_at
		FROM prepared_dishes
		WHERE id = ?::uuid
	`, dishID).Scan(&dish).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find prepared dish: %w", err)
	}

	// Load recipe details
	var recipe models.RecipeCatalog
	err = r.db.Where("id = ?", dish.RecipeID).First(&recipe).Error
	if err != nil {
		return nil, fmt.Errorf("failed to load recipe for dish: %w", err)
	}
	dish.Recipe = &recipe

	return &dish, nil
}

// ConsumePortions reduces available portions and returns updated dish
func (r *PreparedDishRepository) ConsumePortions(dishID string, portions int) (*models.PreparedDish, error) {
	if portions <= 0 {
		return nil, fmt.Errorf("portions must be positive")
	}

	// Update portions_available with check constraint
	result := r.db.Exec(`
		UPDATE prepared_dishes
		SET portions_available = portions_available - ?,
		    updated_at = NOW()
		WHERE id = ?::uuid 
		  AND portions_available >= ?
	`, portions, dishID, portions)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to consume portions: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("insufficient portions available or dish not found")
	}

	// Fetch updated dish
	return r.FindByID(dishID)
}

// Delete removes a prepared dish
func (r *PreparedDishRepository) Delete(dishID string) error {
	result := r.db.Exec("DELETE FROM prepared_dishes WHERE id = ?::uuid", dishID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete prepared dish: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("prepared dish not found")
	}
	return nil
}
