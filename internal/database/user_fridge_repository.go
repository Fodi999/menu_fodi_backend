package database

import (
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserFridgeRepository репозиторий для работы с холодильником домашнего повара
type UserFridgeRepository struct {
	db *gorm.DB
}

// NewUserFridgeRepository создает новый экземпляр репозитория
func NewUserFridgeRepository(db *gorm.DB) *UserFridgeRepository {
	return &UserFridgeRepository{db: db}
}

// GetUserFridgeItems возвращает все продукты из холодильника пользователя
func (r *UserFridgeRepository) GetUserFridgeItems(userID string) ([]models.UserFridgeItem, error) {
	var items []models.UserFridgeItem
	result := r.db.
		Preload("Ingredient").
		Where("user_id = ?", userID).
		Order("expires_at ASC"). // Сортировка по дате истечения
		Find(&items)

	if result.Error != nil {
		return nil, result.Error
	}
	return items, nil
}

// GetByID находит продукт по ID
func (r *UserFridgeRepository) GetByID(id string) (*models.UserFridgeItem, error) {
	var item models.UserFridgeItem
	result := r.db.
		Preload("Ingredient").
		Where("id = ?", id).
		First(&item)

	if result.Error != nil {
		return nil, result.Error
	}
	return &item, nil
}

// Create добавляет продукт в холодильник
func (r *UserFridgeRepository) Create(item *models.UserFridgeItem) error {
	// Генерируем UUID если не задан
	if item.ID == "" {
		item.ID = uuid.New().String()
	}

	result := r.db.Create(item)
	if result.Error != nil {
		return result.Error
	}

	// Загружаем связанный ингредиент
	return r.db.Preload("Ingredient").Where("id = ?", item.ID).First(item).Error
}

// Update обновляет продукт
func (r *UserFridgeRepository) Update(item *models.UserFridgeItem) error {
	result := r.db.Save(item)
	return result.Error
}

// Delete удаляет продукт из холодильника
func (r *UserFridgeRepository) Delete(id string) error {
	result := r.db.Delete(&models.UserFridgeItem{}, "id = ?", id)
	return result.Error
}

// GetExpiringSoon возвращает продукты с истекающим сроком годности
func (r *UserFridgeRepository) GetExpiringSoon(userID string, days int) ([]models.UserFridgeItem, error) {
	var items []models.UserFridgeItem

	expiryThreshold := time.Now().AddDate(0, 0, days)

	result := r.db.
		Preload("Ingredient").
		Where("user_id = ? AND expires_at <= ?", userID, expiryThreshold).
		Order("expires_at ASC").
		Find(&items)

	if result.Error != nil {
		return nil, result.Error
	}

	return items, nil
}

// GetExpired возвращает просроченные продукты
func (r *UserFridgeRepository) GetExpired(userID string) ([]models.UserFridgeItem, error) {
	var items []models.UserFridgeItem

	result := r.db.
		Preload("Ingredient").
		Where("user_id = ? AND expires_at < ?", userID, time.Now()).
		Order("expires_at DESC").
		Find(&items)

	if result.Error != nil {
		return nil, result.Error
	}

	return items, nil
}

// ===== PRICE HISTORY METHODS (Event Sourcing) =====

// InsertPriceHistory добавляет событие изменения цены в историю
func (r *UserFridgeRepository) InsertPriceHistory(itemID string, pricePerUnit float64, currency string, source string) error {
	history := models.UserFridgePriceHistory{
		// ID: не устанавливаем - используем DEFAULT gen_random_uuid()::text из БД
		UserFridgeItemID: itemID,
		PricePerUnit:     pricePerUnit,
		Currency:         currency,
		Source:           source,
	}

	result := r.db.Create(&history)
	return result.Error
}

// GetPriceHistory возвращает историю изменения цен для продукта
func (r *UserFridgeRepository) GetPriceHistory(itemID string) ([]models.UserFridgePriceHistory, error) {
	var history []models.UserFridgePriceHistory
	result := r.db.
		Where("user_fridge_item_id = ?", itemID).
		Order("created_at DESC").
		Find(&history)

	if result.Error != nil {
		return nil, result.Error
	}
	return history, nil
}

// UpdateCurrentPrice обновляет кэш текущей цены в основной таблице (денормализация)
func (r *UserFridgeRepository) UpdateCurrentPrice(itemID string, pricePerUnit float64, currency string) error {
	now := time.Now()
	result := r.db.Model(&models.UserFridgeItem{}).
		Where("id = ?", itemID).
		Updates(map[string]interface{}{
			"current_price_per_unit":  pricePerUnit,
			"current_price_currency": currency,
			"price_updated_at":       now,
		})

	return result.Error
}
