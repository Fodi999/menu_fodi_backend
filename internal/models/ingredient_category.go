package models

// IngredientCategory represents a product category in the catalog
// Categories are reference data, not hardcoded enum
type IngredientCategory struct {
	Key       string `gorm:"primaryKey;column:key" json:"key"`              // Stable identifier (fish, meat, dairy)
	Icon      string `gorm:"column:icon;not null" json:"icon"`              // Emoji icon (🐟, 🥩, 🥛)
	SortOrder int    `gorm:"column:sort_order;not null" json:"sortOrder"`   // Display order in UI
	LabelPL   string `gorm:"column:label_pl;not null" json:"labelPl"`       // Polish label
	LabelEN   string `gorm:"column:label_en;not null" json:"labelEn"`       // English label
	LabelRU   string `gorm:"column:label_ru;not null" json:"labelRu"`       // Russian label
}

// TableName specifies the table name for GORM
func (IngredientCategory) TableName() string {
	return "ingredient_categories"
}

// GetLabel returns the label for the specified language
func (c *IngredientCategory) GetLabel(lang string) string {
	switch lang {
	case "pl":
		return c.LabelPL
	case "en":
		return c.LabelEN
	case "ru":
		return c.LabelRU
	default:
		return c.LabelPL // fallback to Polish
	}
}

// IngredientCategoryDTO is the API response format
// ✅ IMPORTANT: Frontend receives ONLY localized label, not all languages
type IngredientCategoryDTO struct {
	Key       string `json:"key"`       // fish, meat, dairy
	Label     string `json:"label"`     // Localized: Ryby, Fish, Рыба (based on Accept-Language)
	Icon      string `json:"icon"`      // 🐟
	SortOrder int    `json:"sortOrder"` // 1
}

// ToDTO converts model to DTO with localized label
func (c *IngredientCategory) ToDTO(lang string) *IngredientCategoryDTO {
	return &IngredientCategoryDTO{
		Key:       c.Key,
		Label:     c.GetLabel(lang),
		Icon:      c.Icon,
		SortOrder: c.SortOrder,
	}
}
