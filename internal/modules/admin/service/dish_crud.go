package service

import (
	"fmt"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ===========================
// Dish CRUD DTOs
// ===========================

// UpdateDishRequest - запрос на обновление блюда
type UpdateDishRequest struct {
	Title       *string  `json:"title"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price"`
	Margin      *float64 `json:"margin"`
}

// GetDishesParams - параметры для фильтрации блюд
type GetDishesParams struct {
	Status   *string // draft, approved, published
	RecipeID *string // фильтр по рецепту
	Limit    int
	Offset   int
}

// ===========================
// Service Interface Extension (добавить в service.go)
// ===========================

// ApproveDish(dishID, adminID string) error
// PublishDish(dishID, adminID string) error
// UpdateDish(dishID string, req UpdateDishRequest, adminID string) (*models.Dish, error)
// GetDishes(params GetDishesParams) ([]models.Dish, int64, error)
// GetDishByID(dishID string) (*models.Dish, error)

// ===========================
// Implementation
// ===========================

// ApproveDish утверждает блюдо (draft → approved)
// Только админ может утверждать блюда
func (s *adminService) ApproveDish(dishID, adminID string) error {
	var dish models.Dish
	if err := s.db.First(&dish, "id = ?", dishID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("dish not found")
		}
		return fmt.Errorf("failed to load dish: %w", err)
	}

	// Проверка: можно утвердить только draft
	if dish.Status != models.DishStatusDraft {
		return fmt.Errorf("only draft dishes can be approved (current status: %s)", dish.Status)
	}

	oldStatus := dish.Status
	approvedAt := time.Now()
	
	// Обновляем статус
	updates := map[string]interface{}{
		"status":      models.DishStatusApproved,
		"approved_by": adminID,
		"approved_at": approvedAt,
		"updated_at":  time.Now(),
	}

	if err := s.db.Model(&dish).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to approve dish: %w", err)
	}

	// Логируем событие
	s.logDishEvent(adminID, dishID, "dish_approved", map[string]interface{}{
		"old_status": oldStatus,
		"new_status": models.DishStatusApproved,
		"approved_at": approvedAt.Format(time.RFC3339),
	})

	logger.Info("Dish approved",
		zap.String("dish_id", dishID),
		zap.String("admin_id", adminID),
		zap.String("old_status", string(oldStatus)),
	)

	return nil
}

// PublishDish публикует блюдо (approved → published)
// Перед публикацией проверяет доступность ингредиентов
func (s *adminService) PublishDish(dishID, adminID string) error {
	var dish models.Dish
	if err := s.db.Preload("Recipe").First(&dish, "id = ?", dishID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("dish not found")
		}
		return fmt.Errorf("failed to load dish: %w", err)
	}

	// Проверка: можно публиковать только approved
	if dish.Status != models.DishStatusApproved {
		return fmt.Errorf("only approved dishes can be published (current status: %s)", dish.Status)
	}

	// TODO: Проверка доступности ингредиентов
	// canCook, _, err := s.checkRecipeAvailability(dish.RecipeID)
	// if err != nil {
	//     return fmt.Errorf("failed to check availability: %w", err)
	// }
	// if !canCook {
	//     return fmt.Errorf("cannot publish: ingredients are not available")
	// }

	oldStatus := dish.Status
	
	// Обновляем статус и доступность
	updates := map[string]interface{}{
		"status":       models.DishStatusPublished,
		"is_available": true, // По умолчанию доступно
		"updated_at":   time.Now(),
	}

	if err := s.db.Model(&dish).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to publish dish: %w", err)
	}

	// Логируем событие
	s.logDishEvent(adminID, dishID, "dish_published", map[string]interface{}{
		"old_status":   oldStatus,
		"new_status":   models.DishStatusPublished,
		"recipe_id":    dish.RecipeID.String(),
		"recipe_title": dish.Recipe.Title,
	})

	logger.Info("Dish published",
		zap.String("dish_id", dishID),
		zap.String("admin_id", adminID),
		zap.String("recipe_id", dish.RecipeID.String()),
	)

	return nil
}

// UnpublishDish снимает блюдо с публикации (published → approved)
func (s *adminService) UnpublishDish(dishID, adminID string) error {
	var dish models.Dish
	if err := s.db.First(&dish, "id = ?", dishID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("dish not found")
		}
		return fmt.Errorf("failed to load dish: %w", err)
	}

	// Проверка: можно снять с публикации только published
	if dish.Status != models.DishStatusPublished {
		return fmt.Errorf("only published dishes can be unpublished (current status: %s)", dish.Status)
	}

	oldStatus := dish.Status
	
	// Обновляем статус
	updates := map[string]interface{}{
		"status":     models.DishStatusApproved,
		"updated_at": time.Now(),
	}

	if err := s.db.Model(&dish).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to unpublish dish: %w", err)
	}

	// Логируем событие
	s.logDishEvent(adminID, dishID, "dish_unpublished", map[string]interface{}{
		"old_status": oldStatus,
		"new_status": models.DishStatusApproved,
		"reason":     "admin_action",
	})

	logger.Info("Dish unpublished",
		zap.String("dish_id", dishID),
		zap.String("admin_id", adminID),
	)

	return nil
}

// UpdateDish обновляет блюдо
// Можно редактировать только draft и approved блюда
func (s *adminService) UpdateDish(dishID string, req UpdateDishRequest, adminID string) (*models.Dish, error) {
	var dish models.Dish
	if err := s.db.First(&dish, "id = ?", dishID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("dish not found")
		}
		return nil, fmt.Errorf("failed to load dish: %w", err)
	}

	// Проверка: можно редактировать только draft и approved
	if !dish.IsEditable() {
		return nil, fmt.Errorf("cannot edit dish with status: %s (only draft and approved can be edited)", dish.Status)
	}

	// Сохраняем старые значения для логирования
	oldValues := map[string]interface{}{
		"title":       dish.Title,
		"description": dish.Description,
		"price":       dish.Price,
		"margin":      dish.Margin,
	}

	// Обновляем поля
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Price != nil {
		if *req.Price < 0 {
			return nil, fmt.Errorf("price must be positive")
		}
		updates["price"] = *req.Price
	}
	if req.Margin != nil {
		if *req.Margin < 0 || *req.Margin > 100 {
			return nil, fmt.Errorf("margin must be between 0 and 100")
		}
		updates["margin"] = *req.Margin
	}

	if err := s.db.Model(&dish).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update dish: %w", err)
	}

	// Перезагружаем для возврата
	if err := s.db.Preload("Recipe").First(&dish, "id = ?", dishID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload dish: %w", err)
	}

	// Логируем событие
	newValues := map[string]interface{}{
		"title":       dish.Title,
		"description": dish.Description,
		"price":       dish.Price,
		"margin":      dish.Margin,
	}

	s.logDishEvent(adminID, dishID, "dish_updated", map[string]interface{}{
		"old_values": oldValues,
		"new_values": newValues,
	})

	logger.Info("Dish updated",
		zap.String("dish_id", dishID),
		zap.String("admin_id", adminID),
	)

	return &dish, nil
}

// GetDishes возвращает список блюд с фильтрацией
func (s *adminService) GetDishes(params GetDishesParams) ([]models.Dish, int64, error) {
	var dishes []models.Dish
	var total int64

	// Базовый запрос
	query := s.db.Model(&models.Dish{})

	// Фильтры
	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}
	if params.RecipeID != nil {
		query = query.Where("recipe_id = ?", *params.RecipeID)
	}

	// Подсчёт общего количества
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count dishes: %w", err)
	}

	// Получение данных с пагинацией
	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	// Загружаем с связями
	query = query.
		Preload("Recipe").
		Preload("Creator").
		Preload("Approver").
		Order("created_at DESC")

	if err := query.Find(&dishes).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to load dishes: %w", err)
	}

	return dishes, total, nil
}

// GetDishByID возвращает блюдо по ID
func (s *adminService) GetDishByID(dishID string) (*models.Dish, error) {
	var dish models.Dish
	
	err := s.db.
		Preload("Recipe").
		Preload("Recipe.Ingredients").
		Preload("Recipe.Ingredients.Ingredient").
		Preload("Creator").
		Preload("Approver").
		First(&dish, "id = ?", dishID).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("dish not found")
		}
		return nil, fmt.Errorf("failed to load dish: %w", err)
	}

	return &dish, nil
}

// DeleteDish удаляет блюдо
// Можно удалять только draft блюда
func (s *adminService) DeleteDish(dishID, adminID string) error {
	var dish models.Dish
	if err := s.db.First(&dish, "id = ?", dishID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("dish not found")
		}
		return fmt.Errorf("failed to load dish: %w", err)
	}

	// Проверка: можно удалять только draft
	if dish.Status != models.DishStatusDraft {
		return fmt.Errorf("only draft dishes can be deleted (current status: %s)", dish.Status)
	}

	if err := s.db.Delete(&dish).Error; err != nil {
		return fmt.Errorf("failed to delete dish: %w", err)
	}

	// Логируем событие
	s.logDishEvent(adminID, dishID, "dish_deleted", map[string]interface{}{
		"title":      dish.Title,
		"status":     dish.Status,
		"recipe_id":  dish.RecipeID.String(),
	})

	logger.Info("Dish deleted",
		zap.String("dish_id", dishID),
		zap.String("admin_id", adminID),
	)

	return nil
}
