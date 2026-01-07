package models

import "time"

// Culinary categories - for UI display (какой это продукт визуально)
const (
	CategoryFish      = "fish"      // Рыба и морепродукты
	CategoryMeat      = "meat"      // Мясо и птица
	CategoryEgg       = "egg"       // Яйца
	CategoryVegetable = "vegetable" // Овощи
	CategoryFruit     = "fruit"     // Фрукты и ягоды
	CategoryDairy     = "dairy"     // Молочные продукты
	CategoryGrain     = "grain"     // Крупы, макароны, хлеб
	CategoryCondiment = "condiment" // Специи, соусы, масла
	CategoryOther     = "other"     // Прочее
)

// Nutrition groups - for AI and analytics (какая роль в питании)
const (
	NutritionProtein      = "protein"      // Белковые продукты
	NutritionCarbohydrate = "carbohydrate" // Углеводные продукты
	NutritionFat          = "fat"          // Жиросодержащие продукты
	NutritionVegetable    = "vegetable"    // Овощи (некрахмалистые)
	NutritionFruit        = "fruit"        // Фрукты и ягоды
	NutritionDairy        = "dairy"        // Молочные продукты
	NutritionCondiment    = "condiment"    // Специи и приправы
	NutritionOther        = "other"        // Прочее
)

// Ingredient модель ингредиента - ОБЩИЙ КАТАЛОГ для всех пользователей
// Не содержит информации о пользователе, складе, партии или количестве
// Используется для автокомплита и как справочник
type Ingredient struct {
	ID                   string    `gorm:"primaryKey;column:id" json:"id"`
	Name                 string    `gorm:"column:name" json:"name"` // Legacy field, use name_pl
	NamePL               *string   `gorm:"column:name_pl" json:"namePl,omitempty"`
	NameEN               *string   `gorm:"column:name_en" json:"nameEn,omitempty"`
	NameRU               *string   `gorm:"column:name_ru" json:"nameRu,omitempty"`
	NormalizedValue      *string   `gorm:"column:normalized_value" json:"-"` // For search only
	Unit                 string    `gorm:"column:unit;not null" json:"unit"` // "g", "ml", "pcs"
	Category             string    `gorm:"column:category;not null" json:"category"` // Culinary category (UI)
	NutritionGroup       string    `gorm:"column:nutrition_group;not null" json:"nutritionGroup"` // Nutritional grouping (AI, analytics)
	DefaultShelfLifeDays *int      `gorm:"column:defaultShelfLifeDays" json:"defaultShelfLifeDays,omitempty"`
	DefaultPricePerUnit  *float64  `gorm:"column:defaultPricePerUnit" json:"defaultPricePerUnit,omitempty"`
	AutoTranslated       bool      `gorm:"column:auto_translated;default:false" json:"autoTranslated"`
	CreatedAt            time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
}

// TableName указывает имя таблицы для GORM
func (Ingredient) TableName() string {
	return "Ingredient"
}

// GetName returns ingredient name for specified language
// Falls back to PL if requested language not available
func (i *Ingredient) GetName(lang string) string {
	switch lang {
	case "en":
		if i.NameEN != nil && *i.NameEN != "" {
			return *i.NameEN
		}
	case "ru":
		if i.NameRU != nil && *i.NameRU != "" {
			return *i.NameRU
		}
	case "pl":
		if i.NamePL != nil && *i.NamePL != "" {
			return *i.NamePL
		}
	}

	// Fallback to legacy Name field or NamePL
	if i.NamePL != nil && *i.NamePL != "" {
		return *i.NamePL
	}
	return i.Name
}

// IngredientBasicInfo - lightweight DTO for ingredient data in API responses
// Contains all language fields for instant frontend language switching
type IngredientBasicInfo struct {
	ID       string  `json:"id"`
	Key      string  `json:"key"`      // Normalized key for search (deprecated, use id)
	Name     string  `json:"name"`     // Legacy field (same as name_pl)
	NamePL   *string `json:"name_pl"`  // Polish name
	NameEN   *string `json:"name_en"`  // English name
	NameRU   *string `json:"name_ru"`  // Russian name
	Category string  `json:"category"` // protein, vegetable, dairy, grain, condiment, other
	Unit     string  `json:"unit"`     // g, ml, szt
}

// NewIngredientBasicInfo creates DTO from Ingredient model
func NewIngredientBasicInfo(ing *Ingredient) *IngredientBasicInfo {
	if ing == nil {
		return nil
	}
	return &IngredientBasicInfo{
		ID:       ing.ID,
		Key:      ing.ID, // Using ID as key for consistency
		Name:     ing.GetName("pl"),
		NamePL:   ing.NamePL,
		NameEN:   ing.NameEN,
		NameRU:   ing.NameRU,
		Category: ing.Category,
		Unit:     ing.Unit,
	}
}

// StockItem модель складских остатков - ДЛЯ PRO_CHEF (рестораны/бизнес)
// Содержит детальную информацию о партиях, поставщиках, ценах
type StockItem struct {
	ID              string      `gorm:"primaryKey;column:id" json:"id"`
	IngredientID    string      `gorm:"column:ingredientId" json:"ingredientId"`
	Quantity        float64     `gorm:"column:quantity" json:"quantity"`
	UpdatedAt       time.Time   `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`
	BatchNumber     *string     `gorm:"column:batchNumber" json:"batchNumber,omitempty"`
	BruttoWeight    *float64    `gorm:"column:bruttoWeight" json:"bruttoWeight,omitempty"`
	NettoWeight     *float64    `gorm:"column:nettoWeight" json:"nettoWeight,omitempty"`
	WastePercentage *float64    `gorm:"column:wastePercentage" json:"wastePercentage,omitempty"`
	ExpiryDays      *int        `gorm:"column:expiryDays" json:"expiryDays,omitempty"`
	Supplier        *string     `gorm:"column:supplier" json:"supplier,omitempty"`
	Category        *string     `gorm:"column:category" json:"category,omitempty"`
	PriceBrutto     *float64    `gorm:"column:priceBrutto" json:"priceBrutto,omitempty"`
	PriceNetto      *float64    `gorm:"column:priceNetto" json:"priceNetto,omitempty"`
	PricePerUnit    *float64    `gorm:"column:pricePerUnit" json:"pricePerUnit,omitempty"` // Цена за единицу (кг/л/шт)
	Ingredient      *Ingredient `gorm:"foreignKey:IngredientID;references:ID" json:"ingredient,omitempty"`
}

// TableName указывает имя таблицы для GORM
func (StockItem) TableName() string {
	return "StockItem"
}

// StockMovement модель движения товаров на складе
type StockMovement struct {
	ID          string    `gorm:"primaryKey;column:id" json:"id"`
	StockItemID string    `gorm:"column:stockItemId" json:"stockItemId"`
	Type        string    `gorm:"column:type" json:"type"` // "in" (поступление) или "out" (расход)
	Quantity    float64   `gorm:"column:quantity" json:"quantity"`
	PriceBrutto *float64  `gorm:"column:priceBrutto" json:"priceBrutto,omitempty"`
	PriceNetto  *float64  `gorm:"column:priceNetto" json:"priceNetto,omitempty"`
	Note        *string   `gorm:"column:note" json:"note,omitempty"`
	CreatedAt   time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
}

// TableName указывает имя таблицы для GORM
func (StockMovement) TableName() string {
	return "StockMovement"
}

// CreateIngredientRequest запрос на создание ингредиента
type CreateIngredientRequest struct {
	Name            string  `json:"name"`
	Unit            string  `json:"unit"`
	Quantity        float64 `json:"quantity"`
	BruttoWeight    float64 `json:"bruttoWeight"`
	NettoWeight     float64 `json:"nettoWeight"`
	WastePercentage float64 `json:"wastePercentage"`
	ExpiryDays      int     `json:"expiryDays"`
	Supplier        string  `json:"supplier"`
	Category        string  `json:"category"`
	PriceBrutto     float64 `json:"priceBrutto"`
	PriceNetto      float64 `json:"priceNetto"`
	PricePerUnit    float64 `json:"pricePerUnit"` // Цена за единицу (кг/л/шт)
}

// UpdateIngredientRequest запрос на обновление ингредиента
type UpdateIngredientRequest struct {
	Name            string  `json:"name"`
	Unit            string  `json:"unit"`
	Quantity        float64 `json:"quantity"`
	BruttoWeight    float64 `json:"bruttoWeight"`
	NettoWeight     float64 `json:"nettoWeight"`
	WastePercentage float64 `json:"wastePercentage"`
	ExpiryDays      int     `json:"expiryDays"`
	Supplier        string  `json:"supplier"`
	Category        string  `json:"category"`
	PriceBrutto     float64 `json:"priceBrutto"`
	PriceNetto      float64 `json:"priceNetto"`
	PricePerUnit    float64 `json:"pricePerUnit"` // Цена за единицу (кг/л/шт)
}

// IngredientResponse DTO для ответа API (плоская структура для frontend)
type IngredientResponse struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Unit            string    `json:"unit"`
	BatchNumber     *string   `json:"batchNumber,omitempty"`
	Category        *string   `json:"category,omitempty"`
	Supplier        *string   `json:"supplier,omitempty"`
	BruttoWeight    *float64  `json:"bruttoWeight,omitempty"`
	NettoWeight     *float64  `json:"nettoWeight,omitempty"`
	WastePercentage *float64  `json:"wastePercentage,omitempty"`
	ExpiryDays      *int      `json:"expiryDays,omitempty"`
	PriceBrutto     *float64  `json:"priceBrutto,omitempty"`
	PriceNetto      *float64  `json:"priceNetto,omitempty"`
	PricePerUnit    *float64  `json:"pricePerUnit,omitempty"` // Цена за единицу (кг/л/шт)
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// ToResponse преобразует StockItem в IngredientResponse
func (s *StockItem) ToResponse() *IngredientResponse {
	if s.Ingredient == nil {
		return nil
	}
	return &IngredientResponse{
		ID:              s.ID,
		Name:            s.Ingredient.Name,
		Unit:            s.Ingredient.Unit,
		BatchNumber:     s.BatchNumber,
		Category:        s.Category,
		Supplier:        s.Supplier,
		BruttoWeight:    s.BruttoWeight,
		NettoWeight:     s.NettoWeight,
		WastePercentage: s.WastePercentage,
		ExpiryDays:      s.ExpiryDays,
		PriceBrutto:     s.PriceBrutto,
		PriceNetto:      s.PriceNetto,
		PricePerUnit:    s.PricePerUnit, // Добавляем pricePerUnit в ответ
		CreatedAt:       s.Ingredient.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	}
}
