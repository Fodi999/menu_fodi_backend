package models

import "time"

// UserFridgeItem модель холодильника домашнего повара (HOME_CHEF) - MVP версия
// Простая структура без лишних полей
type UserFridgeItem struct {
	ID           string     `gorm:"primaryKey;type:uuid;default:gen_random_uuid();column:id" json:"id"`
	UserID       string     `gorm:"type:uuid;not null;column:user_id;index" json:"userId"`
	IngredientID string     `gorm:"type:uuid;not null;column:ingredient_id;index" json:"ingredientId"` // Обязательная связь с каталогом
	Quantity     float64    `gorm:"not null;column:quantity" json:"quantity"`                          // Числовое значение (например, 500)
	Unit         string     `gorm:"not null;column:unit" json:"unit"`                                  // "g", "ml", "szt" - единица измерения
	PricePerUnit *float64   `gorm:"column:price_per_unit" json:"pricePerUnit,omitempty"`               // Цена ЗА ЕДИНИЦУ (нормализованная: всегда за g/ml/szt)
	Currency     string     `gorm:"column:currency;default:'PLN'" json:"currency"`                     // PLN, EUR, USD
	ExpiresAt    *time.Time `gorm:"column:expires_at;index" json:"expiresAt,omitempty"`                // Дата истечения срока (nullable)
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`

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
	IngredientID string      `json:"ingredientId" binding:"required"`  // UUID из каталога
	Quantity     float64     `json:"quantity" binding:"required,gt=0"` // Количество (должно быть > 0)
	PriceInput   *PriceInput `json:"priceInput,omitempty"`             // Опциональная цена
}

// PriceInput структура для ввода цены от фронтенда
type PriceInput struct {
	Value float64 `json:"value" binding:"required,gt=0"` // 3.20
	Per   string  `json:"per" binding:"required"`        // "kg", "l", "szt"
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
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Category   string   `json:"category"` // protein, vegetable, dairy, grain, condiment, other
	Quantity   float64  `json:"quantity"`
	Unit       string   `json:"unit"`
	TotalPrice *float64 `json:"totalPrice,omitempty"` // Вычисляется: quantity * pricePerUnit
	Currency   string   `json:"currency,omitempty"`   // PLN, EUR, USD
	DaysLeft   int      `json:"daysLeft"`
	Status     string   `json:"status"` // "ok", "warning", "critical"
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

// UserFridgePriceHistory история изменений цен продуктов (event sourcing)
type UserFridgePriceHistory struct {
	ID                string    `gorm:"primaryKey;type:text;default:gen_random_uuid()::text;column:id" json:"id"`
	UserFridgeItemID  string    `gorm:"type:text;not null;column:user_fridge_item_id;index" json:"userFridgeItemId"`
	PricePerUnit      float64   `gorm:"not null;column:price_per_unit" json:"pricePerUnit"`
	Currency          string    `gorm:"not null;default:'PLN';column:currency" json:"currency"`
	Source            string    `gorm:"not null;default:'manual';column:source" json:"source"` // manual, receipt, estimate, market, ai
	CreatedAt         time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

// TableName указывает имя таблицы для GORM
func (UserFridgePriceHistory) TableName() string {
	return "user_fridge_price_history"
}

// AddPriceRequest запрос на добавление цены к продукту
type AddPriceRequest struct {
	PricePerUnit float64 `json:"pricePerUnit" binding:"required,gt=0"` // Нормализованная цена (за g/ml/szt)
	Currency     string  `json:"currency" binding:"required"`          // PLN, EUR, USD
	Source       string  `json:"source"`                               // manual (default), receipt, estimate, market, ai
}
