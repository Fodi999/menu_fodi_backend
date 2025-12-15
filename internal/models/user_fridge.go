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
	
	// Current price cache (denormalized from history for performance)
	// Source of truth: user_fridge_price_history table
	CurrentPricePerUnit  *float64   `gorm:"column:current_price_per_unit" json:"currentPricePerUnit,omitempty"`
	CurrentPriceCurrency string     `gorm:"column:current_price_currency;default:'PLN'" json:"currentPriceCurrency,omitempty"`
	PriceUpdatedAt       *time.Time `gorm:"column:price_updated_at" json:"priceUpdatedAt,omitempty"`
	
	// Date tracking
	ArrivedAt time.Time  `gorm:"column:arrived_at;not null;default:CURRENT_TIMESTAMP;index:,sort:desc" json:"arrivedAt"` // Когда продукт попал в холодильник (автоматически)
	ExpiresAt *time.Time `gorm:"column:expires_at;index" json:"expiresAt,omitempty"`                                      // Дата истечения срока (nullable, может вычисляться автоматически)
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`

	// Relations
	User       *User       `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	Ingredient *Ingredient `gorm:"foreignKey:IngredientID;references:ID" json:"ingredient,omitempty"`
}

// UserFridgePriceHistory история изменения цен (event sourcing)
type UserFridgePriceHistory struct {
	ID               string    `gorm:"primaryKey;type:text;default:gen_random_uuid()::text;column:id" json:"id"`
	UserFridgeItemID string    `gorm:"type:text;not null;column:user_fridge_item_id;index" json:"userFridgeItemId"`
	PricePerUnit     float64   `gorm:"not null;column:price_per_unit" json:"pricePerUnit"`
	Currency         string    `gorm:"not null;default:'PLN';column:currency" json:"currency"`
	Source           string    `gorm:"not null;default:'manual';column:source;index" json:"source"` // manual, receipt, estimate, market, ai
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime;index:,sort:desc" json:"createdAt"`

	// Relation
	UserFridgeItem *UserFridgeItem `gorm:"foreignKey:UserFridgeItemID;references:ID" json:"userFridgeItem,omitempty"`
}

// TableName указывает имя таблицы для GORM
func (UserFridgePriceHistory) TableName() string {
	return "user_fridge_price_history"
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
	ExpiresAt    *time.Time  `json:"expiresAt,omitempty"`              // Опциональная дата истечения (иначе auto-вычисляется)
}

// PriceInput структура для ввода цены от фронтенда
type PriceInput struct {
	Value float64 `json:"value" binding:"required,gt=0"` // 3.20
	Per   string  `json:"per" binding:"required"`        // "kg", "l", "szt"
}

// AddPriceRequest запрос на добавление события изменения цены
type AddPriceRequest struct {
	PricePerUnit float64 `json:"pricePerUnit" binding:"required,gt=0"` // Нормализованная цена за единицу
	Currency     string  `json:"currency" binding:"required"`          // PLN, EUR, USD
	Source       string  `json:"source" binding:"required"`            // manual, receipt, estimate, market, ai
}

// PriceHistoryResponse DTO для истории цен
type PriceHistoryResponse struct {
	ID           string    `json:"id"`
	PricePerUnit float64   `json:"pricePerUnit"`
	Currency     string    `json:"currency"`
	Source       string    `json:"source"` // manual, receipt, estimate, market, ai
	CreatedAt    time.Time `json:"createdAt"`
}

// FridgeItemResponse DTO для ответа API с расширенной информацией
type FridgeItemResponse struct {
	ID         string              `json:"id"`
	Ingredient IngredientShortInfo `json:"ingredient"`
	Quantity   float64             `json:"quantity"`
	ExpiresAt  string              `json:"expiresAt"` // ISO 8601 формат (или "" если нет)
	DaysLeft   *int                `json:"daysLeft"`  // Вычисляется на бэкенде (null если нет срока)
}

// IngredientShortInfo краткая информация об ингредиенте для ответа
type IngredientShortInfo struct {
	Name     string `json:"name"`
	Unit     string `json:"unit"`
	Category string `json:"category"`
}

// FridgeItemListResponse DTO для списка продуктов в холодильнике
type FridgeItemListResponse struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Category     string     `json:"category"` // protein, vegetable, dairy, grain, condiment, other
	Quantity     float64    `json:"quantity"`
	Unit         string     `json:"unit"`
	PricePerUnit *float64   `json:"pricePerUnit,omitempty"` // Цена за единицу (из кэша current_price_per_unit)
	TotalPrice   *float64   `json:"totalPrice,omitempty"`   // Вычисляется: quantity * pricePerUnit
	Currency     string     `json:"currency,omitempty"`     // PLN, EUR, USD
	ArrivedAt    time.Time  `json:"arrivedAt"`              // Когда продукт попал в холодильник
	ExpiresAt    *time.Time `json:"expiresAt,omitempty"`    // Когда испортится (может быть NULL)
	DaysLeft     *int       `json:"daysLeft,omitempty"`     // Дней до истечения (NULL если нет срока годности)
	Status       string     `json:"status"`                 // "fresh", "ok", "warning", "expired"
}

// GetFridgeItemStatus возвращает статус продукта на основе оставшихся дней
func GetFridgeItemStatus(daysLeft *int) string {
	if daysLeft == nil {
		return "fresh" // Нет срока годности - продукт свежий
	}
	if *daysLeft < 0 {
		return "expired"
	}
	if *daysLeft <= 2 {
		return "warning"
	}
	return "ok"
}
