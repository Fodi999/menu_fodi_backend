package models

import "time"

// UserFridgeItem модель холодильника домашнего повара (HOME_CHEF)
// Более простая структура чем StockItem, без партий и поставщиков
type UserFridgeItem struct {
	ID           string     `gorm:"primaryKey;column:id" json:"id"`
	UserID       string     `gorm:"type:uuid;not null;column:userId" json:"userId"`
	IngredientID *string    `gorm:"type:uuid;column:ingredientId" json:"ingredientId,omitempty"` // Опциональная связь с каталогом
	Name         string     `gorm:"not null;column:name" json:"name"`
	Quantity     string     `gorm:"column:quantity" json:"quantity"`     // "500 g", "2 l" - произвольный формат
	Price        *float64   `gorm:"column:price" json:"price,omitempty"` // Цена за покупку (опционально)
	PurchasedAt  *time.Time `gorm:"column:purchasedAt" json:"purchasedAt,omitempty"`
	ExpiryDate   *time.Time `gorm:"column:expiryDate" json:"expiryDate,omitempty"`
	CreatedAt    time.Time  `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time  `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`
	DeletedAt    *time.Time `gorm:"column:deletedAt;index" json:"deletedAt,omitempty"` // Soft delete для аналитики и истории

	// Relations
	User       *User       `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	Ingredient *Ingredient `gorm:"foreignKey:IngredientID;references:ID" json:"ingredient,omitempty"`
}

// TableName указывает имя таблицы для GORM
func (UserFridgeItem) TableName() string {
	return "UserFridgeItem"
}

// CreateFridgeItemRequest запрос на добавление продукта в холодильник
type CreateFridgeItemRequest struct {
	IngredientID *string    `json:"ingredientId,omitempty"` // Можно выбрать из каталога
	Name         string     `json:"name"`                   // Или указать вручную
	Quantity     string     `json:"quantity"`
	Price        *float64   `json:"price,omitempty"`
	PurchasedAt  *time.Time `json:"purchasedAt,omitempty"`
	ExpiryDate   *time.Time `json:"expiryDate,omitempty"`
}

// UpdateFridgeItemRequest запрос на обновление продукта
type UpdateFridgeItemRequest struct {
	Name       *string    `json:"name,omitempty"`
	Quantity   *string    `json:"quantity,omitempty"`
	Price      *float64   `json:"price,omitempty"`
	ExpiryDate *time.Time `json:"expiryDate,omitempty"`
}

// FridgeItemResponse DTO для ответа API
type FridgeItemResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Quantity    string     `json:"quantity"`
	Price       *float64   `json:"price,omitempty"`
	PurchasedAt *time.Time `json:"purchasedAt,omitempty"`
	ExpiryDate  *time.Time `json:"expiryDate,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// ToResponse преобразует UserFridgeItem в FridgeItemResponse
func (f *UserFridgeItem) ToResponse() *FridgeItemResponse {
	return &FridgeItemResponse{
		ID:          f.ID,
		Name:        f.Name,
		Quantity:    f.Quantity,
		Price:       f.Price,
		PurchasedAt: f.PurchasedAt,
		ExpiryDate:  f.ExpiryDate,
		CreatedAt:   f.CreatedAt,
		UpdatedAt:   f.UpdatedAt,
	}
}
