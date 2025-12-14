package database

import (
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/google/uuid"
)

// UserFridgeRepository репозиторий для работы с холодильником домашнего повара
type UserFridgeRepository struct{}

// GetUserFridgeItems возвращает все продукты из холодильника пользователя
func (r *UserFridgeRepository) GetUserFridgeItems(userID string) ([]models.UserFridgeItem, error) {
	var items []models.UserFridgeItem
	result := DB.
		Preload("Ingredient").
		Where(`"userId" = ?`, userID).
		Order(`"createdAt" DESC`).
		Find(&items)

	if result.Error != nil {
		return nil, result.Error
	}
	return items, nil
}

// GetByID находит продукт по ID
func (r *UserFridgeRepository) GetByID(id string) (*models.UserFridgeItem, error) {
	var item models.UserFridgeItem
	result := DB.
		Preload("Ingredient").
		First(&item, "id = ?", id)

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

	result := DB.Create(item)
	return result.Error
}

// Update обновляет продукт
func (r *UserFridgeRepository) Update(item *models.UserFridgeItem) error {
	result := DB.Save(item)
	return result.Error
}

// Delete удаляет продукт из холодильника
func (r *UserFridgeRepository) Delete(id string) error {
	result := DB.Delete(&models.UserFridgeItem{}, "id = ?", id)
	return result.Error
}

// GetExpiringSoon возвращает продукты с истекающим сроком годности
func (r *UserFridgeRepository) GetExpiringSoon(userID string, daysThreshold int) ([]models.UserFridgeItem, error) {
	var items []models.UserFridgeItem

	query := `
		SELECT * FROM "UserFridgeItem"
		WHERE "userId" = ?
		AND "expiryDate" IS NOT NULL
		AND "expiryDate" <= NOW() + INTERVAL '? days'
		ORDER BY "expiryDate" ASC
	`

	result := DB.Raw(query, userID, daysThreshold).Scan(&items)
	if result.Error != nil {
		return nil, result.Error
	}

	return items, nil
}
