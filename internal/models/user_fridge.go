package models

import "time"

// UserFridgeItem модель холодильника домашнего повара (HOME_CHEF) - MVP версия
// Простая структура без лишних полей
type UserFridgeItem struct {
	ID           string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid();column:id" json:"id"`
	UserID       string    `gorm:"type:uuid;not null;column:user_id;index" json:"userId"`
	IngredientID string    `gorm:"type:uuid;not null;column:ingredient_id;index" json:"ingredientId"` // Обязательная связь с каталогом
	Quantity     float64   `gorm:"not null;column:quantity" json:"quantity"`                           // Числовое значение (например, 500)
	Unit         string    `gorm:"not null;column:unit" json:"unit"`                                   // "g", "ml", "pcs" - копия из каталога
	ExpiresAt    time.Time `gorm:"not null;column:expires_at;index" json:"expiresAt"`                  // Дата истечения срока
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`

	// Relations
	User       *User       `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	Ingredient *Ingredient `gorm:"foreignKey:IngredientID;references:ID" json:"ingredient,omitempty"`
}

// TableName указывает имя таблицы для GORM
func (UserFridgeItem) TableName() string {
	return "user_fridge_items"
}

// CreateFridgeItemRequest запрос на добавление продукта в холодильник
type CreateFridgeItemRequest struct {
	IngredientID string  `json:"ingredientId" binding:"required"` // UUID из каталога
	Quantity     float64 `json:"quantity" binding:"required,gt=0"` // Количество (должно быть > 0)
}

// FridgeItemResponse DTO для ответа API с расширенной информацией
type FridgeItemResponse struct {
	ID         string              `json:"id"`
	Ingredient IngredientShortInfo `json:"ingredient"`
	Quantity   float64             `json:"quantity"`
	ExpiresAt  string              `json:"expiresAt"` // ISO 8601 формат
	DaysLeft   int                 `json:"daysLeft"`  // Вычисляется на бэкенде
}

// IngredientShortInfo краткая информация об ингредиенте для ответа
type IngredientShortInfo struct {
	Name     string `json:"name"`
	Unit     string `json:"unit"`
	Category string `json:"category"`
}

// FridgeItemListResponse DTO для списка продуктов в холодильнике
type FridgeItemListResponse struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
	DaysLeft int     `json:"daysLeft"`
	Status   string  `json:"status"` // "ok", "warning", "critical"
}

// GetStatus возвращает статус продукта на основе оставшихся дней
func GetFridgeItemStatus(daysLeft int) string {
	if daysLeft < 0 {
		return "expired"
	}
	if daysLeft <= 1 {
		return "critical"
	}
	if daysLeft <= 3 {
		return "warning"
	}
	return "ok"
}
