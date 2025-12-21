package database

import (
	"fmt"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserRecipeSessionRepository handles database operations for recipe sessions
type UserRecipeSessionRepository struct {
	db *gorm.DB
}

// NewUserRecipeSessionRepository creates a new repository instance
func NewUserRecipeSessionRepository(db *gorm.DB) *UserRecipeSessionRepository {
	return &UserRecipeSessionRepository{db: db}
}

// GetSession retrieves the session for a user
func (r *UserRecipeSessionRepository) GetSession(userID string) (*models.UserRecipeSession, error) {
	var session models.UserRecipeSession

	result := r.db.Where("user_id = ?", userID).First(&session)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Create a new session if not found
			return r.createSession(userID)
		}
		return nil, fmt.Errorf("failed to get session: %w", result.Error)
	}

	return &session, nil
}

// createSession creates a new session for a user
func (r *UserRecipeSessionRepository) createSession(userID string) (*models.UserRecipeSession, error) {
	session := &models.UserRecipeSession{
		UserID:            userID,
		ExcludedRecipeIDs: pq.StringArray{},
		UpdatedAt:         time.Now(),
	}

	result := r.db.Create(session)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create session: %w", result.Error)
	}

	return session, nil
}

// UpdateSession updates the session with new recipe ID and exclusions
func (r *UserRecipeSessionRepository) UpdateSession(userID, recipeID string, excludedIDs []string) error {
	// Convert to pq.StringArray
	excludedArray := pq.StringArray(excludedIDs)

	updates := map[string]interface{}{
		"last_recipe_id":      recipeID,
		"excluded_recipe_ids": excludedArray,
		"updated_at":          time.Now(),
	}

	result := r.db.Model(&models.UserRecipeSession{}).
		Where("user_id = ?", userID).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update session: %w", result.Error)
	}

	// If no rows were updated, create a new session
	if result.RowsAffected == 0 {
		session := &models.UserRecipeSession{
			UserID:            userID,
			LastRecipeID:      &recipeID,
			ExcludedRecipeIDs: excludedArray,
			UpdatedAt:         time.Now(),
		}

		result := r.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"last_recipe_id", "excluded_recipe_ids", "updated_at"}),
		}).Create(session)

		if result.Error != nil {
			return fmt.Errorf("failed to create session: %w", result.Error)
		}
	}

	return nil
}

// AddExcludedRecipe adds a recipe ID to the exclusion list
func (r *UserRecipeSessionRepository) AddExcludedRecipe(userID, recipeID string) error {
	session, err := r.GetSession(userID)
	if err != nil {
		return err
	}

	// Check if already excluded
	for _, id := range session.ExcludedRecipeIDs {
		if id == recipeID {
			return nil // Already excluded
		}
	}

	// Add to exclusion list
	session.ExcludedRecipeIDs = append(session.ExcludedRecipeIDs, recipeID)
	session.UpdatedAt = time.Now()

	result := r.db.Model(&models.UserRecipeSession{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"excluded_recipe_ids": session.ExcludedRecipeIDs,
			"updated_at":          session.UpdatedAt,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to add excluded recipe: %w", result.Error)
	}

	return nil
}

// ClearSession resets the session (clear exclusions, keep user_id)
func (r *UserRecipeSessionRepository) ClearSession(userID string) error {
	result := r.db.Model(&models.UserRecipeSession{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"last_recipe_id":      nil,
			"excluded_recipe_ids": pq.StringArray{},
			"updated_at":          time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to clear session: %w", result.Error)
	}

	return nil
}
