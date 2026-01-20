package models

import "time"

// UserFridgeItem модель холодильника домашнего повара (HOME_CHEF) - MVP версия
// Простая структура без лишних полей
type UserFridgeItem struct {
	ID           string  `gorm:"primaryKey;type:uuid;default:gen_random_uuid();column:id" json:"id"`
	UserID       string  `gorm:"type:uuid;not null;column:user_id;index" json:"userId"`
	IngredientID string  `gorm:"type:uuid;not null;column:ingredient_id;index" json:"ingredientId"` // Обязательная связь с каталогом
	Quantity     float64 `gorm:"not null;column:quantity" json:"quantity"`                          // Числовое значение (например, 500)
	Unit         string  `gorm:"not null;column:unit" json:"unit"`                                  // "g", "ml", "szt" - единица измерения

	// Current price cache (denormalized from history for performance)
	// Source of truth: user_fridge_price_history table
	CurrentPricePerUnit  *float64   `gorm:"column:current_price_per_unit" json:"currentPricePerUnit,omitempty"`
	CurrentPriceCurrency string     `gorm:"column:current_price_currency;default:'PLN'" json:"currentPriceCurrency,omitempty"`
	PriceUpdatedAt       *time.Time `gorm:"column:price_updated_at" json:"priceUpdatedAt,omitempty"`

	// Date tracking
	ArrivedAt time.Time  `gorm:"column:arrived_at;not null;default:CURRENT_TIMESTAMP;index:,sort:desc" json:"arrivedAt"` // Когда продукт попал в холодильник (автоматически)
	ExpiresAt *time.Time `gorm:"column:expires_at;index" json:"expiresAt,omitempty"`                                     // Дата истечения срока (nullable, может вычисляться автоматически)
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`

	// Relations
	User       *User       `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	Ingredient *Ingredient `gorm:"foreignKey:IngredientID;references:ID" json:"ingredient,omitempty"`
}

// UserFridgePriceHistory история изменения цен (event sourcing)
type UserFridgePriceHistory struct {
	ID               string    `gorm:"primaryKey;type:text;default:gen_random_uuid()::text;column:id" json:"id"`
	UserFridgeItemID string    `gorm:"type:text;not null;column:user_fridge_item_id;index" json:"userFridgeItemId"`
	PricePerUnit     float64   `gorm:"not null;column:price_per_unit" json:"pricePerUnit"`       // Цена за единицу (может быть нормализована)
	UnitForPrice     string    `gorm:"type:text;column:unit_for_price" json:"unitForPrice"`      // Единица измерения цены (kg, l, pcs, g, ml)
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
	PricePerUnit float64 `json:"pricePerUnit" binding:"required,gt=0"` // Цена за единицу (может быть нормализована или исходная)
	UnitForPrice string  `json:"unitForPrice"`                         // Единица измерения цены (kg, l, pcs, g, ml)
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

// PriceAnalysis анализ динамики цены для "умной кухни"
//
// 💡 SMART KITCHEN FEATURE - Price Trend Analysis
// Используйте для отображения пользователю:
//   - Мини-график изменения цены (trend: "up"/"down"/"stable")
//   - Бейдж "⬆ +15%" или "⬇ -10%" рядом с ценой
//   - Алерт "Цена выросла на 20% за неделю"
//
// Example response:
//
//	{
//	  "trend": "up",           // "up" | "down" | "stable"
//	  "percentChange": 15.05,  // +15.05% (положительное = подорожало)
//	  "lastPrice": 0.00581,    // Текущая цена
//	  "previousPrice": 0.00505 // Предыдущая цена
//	}
//
// Frontend implementation ideas:
//   - trend === "up" → показать 🔴 красный бейдж "⬆ +15%"
//   - trend === "down" → показать 🟢 зелёный бейдж "⬇ -10%"
//   - trend === "stable" → скрыть или показать серый "≈ 0%"
//   - Использовать historyCount для проверки надёжности (2 записи = минимум)
type PriceAnalysis struct {
	Trend         string    `json:"trend"`         // "up", "down", "stable"
	PercentChange float64   `json:"percentChange"` // +15.5 (подорожало на 15.5%) или -10.2 (подешевело на 10.2%)
	LastPrice     float64   `json:"lastPrice"`     // Текущая цена за единицу
	PreviousPrice float64   `json:"previousPrice"` // Предыдущая цена за единицу
	LastUpdated   time.Time `json:"lastUpdated"`   // Когда была последняя цена
	HistoryCount  int       `json:"historyCount"`  // Количество записей в истории
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
//
// ⚠️ API CONTRACT - MONEY RULES (не изменять без согласования с фронтом):
//  1. PricePerUnit: normalized price per base unit (g/ml/szt) with HIGH PRECISION
//     Example: 4.00 PLN/kg → 0.004 PLN/g (stored as-is for calculations)
//  2. TotalPrice: ALWAYS ROUNDED to 2 decimal places by backend
//     Example: 3560g * 0.00581 PLN/g = 20.68 PLN (NOT 20.6836)
//  3. Frontend MUST NOT recalculate totalPrice from quantity * pricePerUnit
//  4. Wartość lodówki = SUM(totalPrice) where all values already rounded
//
// 🌍 MULTILINGUAL SUPPORT (NEW):
//  5. Ingredient: Full ingredient object with all language fields (name, name_pl, name_en, name_ru)
//  6. Frontend can switch languages instantly without re-fetching data
//  7. Search works in any language, autocomplete uses same data structure
type FridgeItemListResponse struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`     // DEPRECATED: Use Ingredient.GetName(lang) instead
	Category string  `json:"category"` // protein, vegetable, dairy, grain, condiment, other
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`

	// 🌍 MULTILINGUAL INGREDIENT DATA - MUST HAVE for production
	Ingredient *IngredientBasicInfo `json:"ingredient"` // Full ingredient with all translations

	// PRICE FIELDS - See API CONTRACT above ⬆️
	PricePerUnit *float64 `json:"pricePerUnit,omitempty"` // Normalized price per base unit (high precision, for reference only)
	TotalPrice   *float64 `json:"totalPrice,omitempty"`   // ALWAYS rounded to 2 decimals - NEVER recalculate on frontend!
	Currency     string   `json:"currency,omitempty"`     // PLN, EUR, USD

	// SMART KITCHEN - PRICE ANALYTICS
	PriceAnalysis *PriceAnalysis `json:"priceAnalysis,omitempty"` // Trend analysis: "⬆ +15%" or "⬇ -10%" for mini-charts

	// DATE FIELDS
	ArrivedAt time.Time  `json:"arrivedAt"`           // Когда продукт попал в холодильник
	ExpiresAt *time.Time `json:"expiresAt,omitempty"` // Когда испортится (может быть NULL)
	DaysLeft  *int       `json:"daysLeft,omitempty"`  // Дней до истечения (NULL если нет срока годности)
	Status    string     `json:"status"`              // "fresh", "ok", "warning", "expired"
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

// PriceInfo - информация о цене продукта (из user_fridge_price_history)
type PriceInfo struct {
	Value float64 `json:"value"` // 6.3 (цена за единицу)
	Per   string  `json:"per"`   // "kg", "l", "pcs" (единица измерения цены)
}

// ComputedPrice - вычисленные значения стоимости
type ComputedPrice struct {
	UnitPrice float64 `json:"unitPrice"` // Цена за 1 базовую единицу (g/ml/pcs)
	TotalCost float64 `json:"totalCost"` // Общая стоимость (quantity × unitPrice)
}

// IngredientInfo - базовая информация об ингредиенте для API response
// ✅ ПРАВИЛЬНЫЙ ПОДХОД: Отдаём ВСЕ переводы, frontend выбирает нужный
type IngredientInfo struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`             // Legacy field (fallback)
	NamePL *string `json:"namePl,omitempty"` // Польское название
	NameEN *string `json:"nameEn,omitempty"` // Английское название
	NameRU *string `json:"nameRu,omitempty"` // Русское название
	Unit   string  `json:"unit"`             // g, ml, pcs
}

// CurrentPriceInfo - информация о текущей цене
type CurrentPriceInfo struct {
	Value     float64    `json:"value"`               // 12.34
	Per       string     `json:"per"`                 // kg, l, pcs
	Currency  string     `json:"currency"`            // PLN, EUR, USD
	UpdatedAt *time.Time `json:"updatedAt,omitempty"` // когда обновлена
}

// FridgeItemResponseV2 - новая версия DTO с ценами и категориями
// Контракт API: /api/fridge/items
type FridgeItemResponseV2 struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`              // ✅ ОБРАТНАЯ СОВМЕСТИМОСТЬ: локализованное имя (то же что ingredient.name)
	Ingredient   IngredientInfo    `json:"ingredient"`        // Информация об ингредиенте (новый формат)
	CategoryKey  string            `json:"categoryKey"`       // fish, meat, egg, dairy, etc. (stable key, НЕ зависит от языка)
	Quantity     float64           `json:"quantity"`          // 2000
	Unit         string            `json:"unit"`              // g, ml, pcs
	ExpiresAt    *time.Time        `json:"expiresAt,omitempty"` // ISO 8601
	DaysLeft     *int              `json:"daysLeft,omitempty"`  // Вычисленное на backend
	CurrentPrice *CurrentPriceInfo `json:"currentPrice,omitempty"` // Текущая цена (если есть)

	// Deprecated: используйте currentPrice
	Price    *PriceInfo     `json:"price,omitempty"`    // Старое поле для обратной совместимости
	Computed *ComputedPrice `json:"computed,omitempty"` // Вычисленная стоимость
}

