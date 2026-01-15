package service

import (
	"fmt"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"gorm.io/gorm"
)

// FridgeServiceV2 сервис для работы с холодильником (v2 с поддержкой expiry)
type FridgeServiceV2 interface {
	// Items
	GetItems(userID string) ([]models.FridgeItem, error)
	AddItem(userID string, req AddFridgeItemRequest) (*models.FridgeItem, error)
	UpdateItem(itemID string, userID string, req UpdateFridgeItemRequest) (*models.FridgeItem, error)
	DeleteItem(itemID string, userID string) error
	DiscardItem(itemID string, userID string) error
	
	// Auto-checks
	CheckAndNotifyExpiring(userID string) error
}

type fridgeServiceV2 struct {
	db *gorm.DB
}

func NewFridgeServiceV2(db *gorm.DB) FridgeServiceV2 {
	return &fridgeServiceV2{
		db: db,
	}
}

// AddFridgeItemRequest запрос на добавление продукта
type AddFridgeItemRequest struct {
	IngredientID string     `json:"ingredientId" binding:"required"`
	Quantity     float64    `json:"quantity" binding:"required,gt=0"`
	Unit         string     `json:"unit" binding:"required"`
	ExpiresAt    *time.Time `json:"expiresAt"`
	PriceTotal   float64    `json:"priceTotal"`
}

// UpdateFridgeItemRequest запрос на обновление продукта
type UpdateFridgeItemRequest struct {
	Quantity  *float64   `json:"quantity" binding:"omitempty,gt=0"`
	ExpiresAt *time.Time `json:"expiresAt"`
	PriceTotal *float64  `json:"priceTotal"`
}

// GetItems получить все продукты пользователя с автопроверкой
func (s *fridgeServiceV2) GetItems(userID string) ([]models.FridgeItem, error) {
	var items []models.FridgeItem
	
	err := s.db.Preload("Ingredient").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&items).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch fridge items: %w", err)
	}

	// Обновляем статус и daysLeft для каждого продукта
	freshItems := make([]models.FridgeItem, 0, len(items))
	for i := range items {
		result := EvaluateFridgeItem(&items[i])
		items[i].Status = result.Status
		items[i].DaysLeft = result.DaysLeft

		// Сохраняем изменения в БД если статус изменился
		if items[i].Status == models.FridgeItemStatusExpired {
			s.db.Model(&items[i]).Updates(map[string]interface{}{
				"status":    result.Status,
				"days_left": result.DaysLeft,
			})
			// ❌ НЕ добавляем expired продукты в результат
			continue
		}

		// ✅ Добавляем только fresh/ok продукты
		freshItems = append(freshItems, items[i])
	}

	// Проверяем и создаем уведомления (1 раз в день)
	go s.CheckAndNotifyExpiring(userID)

	return freshItems, nil
}

// AddItem добавить продукт в холодильник
func (s *fridgeServiceV2) AddItem(userID string, req AddFridgeItemRequest) (*models.FridgeItem, error) {
	// Проверяем существование ингредиента
	var ingredient models.Ingredient
	if err := s.db.First(&ingredient, "id = ?", req.IngredientID).Error; err != nil {
		return nil, fmt.Errorf("ingredient not found: %w", err)
	}

	// Создаем новый item
	item := models.FridgeItem{
		UserID:       userID,
		IngredientID: req.IngredientID,
		Quantity:     req.Quantity,
		Unit:         req.Unit,
		ExpiresAt:    req.ExpiresAt,
		PriceTotal:   req.PriceTotal,
		Status:       models.FridgeItemStatusFresh,
	}

	// Вычисляем начальный статус
	result := EvaluateFridgeItem(&item)
	item.Status = result.Status
	item.DaysLeft = result.DaysLeft

	if err := s.db.Create(&item).Error; err != nil {
		return nil, fmt.Errorf("failed to create fridge item: %w", err)
	}

	// Загружаем связанный ингредиент
	s.db.Preload("Ingredient").First(&item, "id = ?", item.ID)

	// Создаем уведомление если нужно
	if result.DaysLeft != nil {
		go CreateExpiryNotification(s.db, &item, *result.DaysLeft)
	}

	return &item, nil
}

// UpdateItem обновить продукт
func (s *fridgeServiceV2) UpdateItem(itemID string, userID string, req UpdateFridgeItemRequest) (*models.FridgeItem, error) {
	var item models.FridgeItem
	
	if err := s.db.Where("id = ? AND user_id = ?", itemID, userID).First(&item).Error; err != nil {
		return nil, fmt.Errorf("fridge item not found: %w", err)
	}

	// Обновляем поля
	updates := make(map[string]interface{})
	if req.Quantity != nil {
		updates["quantity"] = *req.Quantity
	}
	if req.ExpiresAt != nil {
		updates["expires_at"] = *req.ExpiresAt
	}
	if req.PriceTotal != nil {
		updates["price_total"] = *req.PriceTotal
	}

	if len(updates) > 0 {
		if err := s.db.Model(&item).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update fridge item: %w", err)
		}
	}

	// Пересчитываем статус
	s.db.First(&item, "id = ?", itemID)
	result := EvaluateFridgeItem(&item)
	s.db.Model(&item).Updates(map[string]interface{}{
		"status":    result.Status,
		"days_left": result.DaysLeft,
	})

	// Загружаем обновленные данные
	s.db.Preload("Ingredient").First(&item, "id = ?", itemID)

	return &item, nil
}

// DeleteItem удалить продукт из холодильника
func (s *fridgeServiceV2) DeleteItem(itemID string, userID string) error {
	result := s.db.Where("id = ? AND user_id = ?", itemID, userID).Delete(&models.FridgeItem{})
	
	if result.Error != nil {
		return fmt.Errorf("failed to delete fridge item: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("fridge item not found")
	}

	return nil
}

// DiscardItem выбросить продукт (мягкое удаление)
func (s *fridgeServiceV2) DiscardItem(itemID string, userID string) error {
	var item models.FridgeItem
	
	if err := s.db.Where("id = ? AND user_id = ?", itemID, userID).First(&item).Error; err != nil {
		return fmt.Errorf("fridge item not found: %w", err)
	}

	// Меняем статус на discarded
	if err := s.db.Model(&item).Update("status", models.FridgeItemStatusDiscarded).Error; err != nil {
		return fmt.Errorf("failed to discard item: %w", err)
	}

	fmt.Printf("🗑️  Item %s discarded (loss: %.2f PLN)\n", item.ID, item.PriceTotal)

	return nil
}

// CheckAndNotifyExpiring проверяет продукты и создает уведомления
func (s *fridgeServiceV2) CheckAndNotifyExpiring(userID string) error {
	return CheckAndNotifyExpiringItems(s.db, userID)
}

// SetDB устанавливает подключение к БД (для инжекции)
func (s *fridgeServiceV2) SetDB(db *gorm.DB) {
	s.db = db
}
