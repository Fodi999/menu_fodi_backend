
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================================================
// REPOSITORY: User Menu Items (Kitchen Pipeline)
// Single source of truth for what user wants to cook
// ============================================================================

type MenuRepository struct {
	db *gorm.DB
}

func NewMenuRepository(db *gorm.DB) *MenuRepository {
	return &MenuRepository{db: db}
}

// ============================================================================
// CORE METHODS
// ============================================================================

// GetTodayMenu - получить меню на сегодня (все статусы кроме cancelled)
func (r *MenuRepository) GetTodayMenu(ctx context.Context, userID string) ([]models.UserMenuItem, error) {
	var items []models.UserMenuItem
	
	err := r.db.WithContext(ctx).
		Preload("Recipe").                    // Load full recipe data
		Preload("Recipe.Ingredients").        // Load recipe ingredients
		Preload("Recipe.Ingredients.Ingredient"). // Load ingredient details
		Where("user_id = ? AND planned_for = ? AND status != ?", 
			userID, time.Now().Format("2006-01-02"), models.MenuItemCancelled).
		Order("created_at ASC").
		Find(&items).Error
	
	return items, err
}

// GetMenuItem - получить один пункт меню по ID
func (r *MenuRepository) GetMenuItem(ctx context.Context, itemID uuid.UUID) (*models.UserMenuItem, error) {
	var item models.UserMenuItem
	
	err := r.db.WithContext(ctx).
		Preload("Recipe").
		Preload("Recipe.Ingredients").
		Preload("Recipe.Ingredients.Ingredient").
		First(&item, "id = ?", itemID).Error
	
	if err != nil {
		return nil, err
	}
	
	return &item, nil
}

// AddToMenu - добавить рецепт в меню
func (r *MenuRepository) AddToMenu(ctx context.Context, item *models.UserMenuItem) error {
	// Check if recipe already in menu for today
	var existing models.UserMenuItem
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND recipe_id = ? AND planned_for = ? AND status IN (?, ?)", 
			item.UserID, item.RecipeID, item.PlannedFor, 
			models.MenuItemPlanned, models.MenuItemCooking).
		First(&existing).Error
	
	if err == nil {
		return fmt.Errorf("recipe already in menu for today")
	}
	
	if err != gorm.ErrRecordNotFound {
		return err
	}
	
	// Create new menu item
	return r.db.WithContext(ctx).Create(item).Error
}

// UpdateServings - обновить количество порций
func (r *MenuRepository) UpdateServings(ctx context.Context, itemID uuid.UUID, servings int) error {
	return r.db.WithContext(ctx).
		Model(&models.UserMenuItem{}).
		Where("id = ?", itemID).
		Update("servings", servings).Error
}

// UpdateNotes - обновить заметки
func (r *MenuRepository) UpdateNotes(ctx context.Context, itemID uuid.UUID, notes string) error {
	return r.db.WithContext(ctx).
		Model(&models.UserMenuItem{}).
		Where("id = ?", itemID).
		Update("notes", notes).Error
}

// StartCooking - начать готовить (переход planned → cooking)
func (r *MenuRepository) StartCooking(ctx context.Context, itemID uuid.UUID) error {
	now := time.Now()
	
	return r.db.WithContext(ctx).
		Model(&models.UserMenuItem{}).
		Where("id = ? AND status = ?", itemID, models.MenuItemPlanned).
		Updates(map[string]interface{}{
			"status":             models.MenuItemCooking,
			"started_cooking_at": now,
		}).Error
}

// CompleteCooking - завершить готовку (переход cooking → completed)
func (r *MenuRepository) CompleteCooking(ctx context.Context, itemID uuid.UUID) error {
	now := time.Now()
	
	return r.db.WithContext(ctx).
		Model(&models.UserMenuItem{}).
		Where("id = ? AND status = ?", itemID, models.MenuItemCooking).
		Updates(map[string]interface{}{
			"status":       models.MenuItemCompleted,
			"completed_at": now,
		}).Error
}

// CancelMenuItem - отменить приготовление
func (r *MenuRepository) CancelMenuItem(ctx context.Context, itemID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.UserMenuItem{}).
		Where("id = ?", itemID).
		Update("status", models.MenuItemCancelled).Error
}

// DeleteMenuItem - удалить из меню (hard delete)
func (r *MenuRepository) DeleteMenuItem(ctx context.Context, itemID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Delete(&models.UserMenuItem{}, "id = ?", itemID).Error
}

// ============================================================================
// ANALYTICS & STATS
// ============================================================================

// GetMenuStats - статистика меню
func (r *MenuRepository) GetMenuStats(ctx context.Context, userID string) (map[string]int, error) {
	type StatusCount struct {
		Status models.MenuItemStatus
		Count  int
	}
	
	var stats []StatusCount
	err := r.db.WithContext(ctx).
		Model(&models.UserMenuItem{}).
		Select("status, COUNT(*) as count").
		Where("user_id = ? AND planned_for = ?", userID, time.Now().Format("2006-01-02")).
		Group("status").
		Scan(&stats).Error
	
	if err != nil {
		return nil, err
	}
	
	result := make(map[string]int)
	for _, s := range stats {
		result[string(s.Status)] = s.Count
	}
	
	return result, nil
}

// GetHistory - история приготовлений
func (r *MenuRepository) GetHistory(ctx context.Context, userID string, limit int) ([]models.UserMenuItem, error) {
	var items []models.UserMenuItem
	
	err := r.db.WithContext(ctx).
		Preload("Recipe").
		Where("user_id = ? AND status = ?", userID, models.MenuItemCompleted).
		Order("completed_at DESC").
		Limit(limit).
		Find(&items).Error
	
	return items, err
}
