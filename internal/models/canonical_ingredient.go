package models

import (
	"time"
)

// ============================================================================
// КАНОНИЧЕСКАЯ СИСТЕМА ПРОДУКТОВ
// Один реальный продукт = одна запись в CanonicalIngredient
// Все варианты названий = алиасы в IngredientAlias
// ============================================================================

// CanonicalIngredient - канонический продукт (один реальный продукт)
type CanonicalIngredient struct {
	ID                   string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	CanonicalKey         string    `gorm:"column:canonicalKey;type:varchar(255);not null;uniqueIndex" json:"canonicalKey"` // onion, garlic
	CanonicalName        string    `gorm:"column:canonicalName;type:varchar(255);not null" json:"canonicalName"`           // Лук репчатый
	Category             string    `gorm:"column:category;type:varchar(50);not null;index" json:"category"`                // fish, meat, vegetable
	NutritionGroup       string    `gorm:"column:nutritionGroup;type:varchar(50);not null" json:"nutritionGroup"`          // protein, carbohydrate
	BaseUnit             string    `gorm:"column:baseUnit;type:varchar(10);not null;default:'g'" json:"baseUnit"`          // g, ml, pcs
	DefaultShelfLifeDays *int      `gorm:"column:defaultShelfLifeDays" json:"defaultShelfLifeDays,omitempty"`
	DefaultPricePerUnit  *float64  `gorm:"column:defaultPricePerUnit;type:decimal(10,2)" json:"defaultPricePerUnit,omitempty"`
	ImageURL             *string   `gorm:"column:imageUrl;type:text" json:"imageUrl,omitempty"`
	Status               string    `gorm:"column:status;type:varchar(20);not null;default:'active';index" json:"status"` // active, archived
	CreatedAt            time.Time `gorm:"column:createdAt;not null;default:now()" json:"createdAt"`
	UpdatedAt            time.Time `gorm:"column:updatedAt;not null;default:now()" json:"updatedAt"`

	// Связи
	Aliases []IngredientAlias `gorm:"foreignKey:CanonicalIngredientID" json:"aliases,omitempty"`
}

func (CanonicalIngredient) TableName() string {
	return "CanonicalIngredient"
}

// Status constants
const (
	IngredientStatusActive   = "active"
	IngredientStatusArchived = "archived"
)

// IngredientAlias - алиас продукта (языковые варианты, синонимы, опечатки)
type IngredientAlias struct {
	ID                     string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	CanonicalIngredientID  string    `gorm:"column:canonicalIngredientId;type:uuid;not null;index" json:"canonicalIngredientId"`
	Name                   string    `gorm:"column:name;type:varchar(255);not null" json:"name"`                      // "лук", "Onion", "цибуля"
	NormalizedName         string    `gorm:"column:normalizedName;type:varchar(255);not null;uniqueIndex" json:"-"`   // "лук", "onion", "цибуля"
	Language               *string   `gorm:"column:language;type:varchar(10);index" json:"language,omitempty"`        // pl, en, ru, uk
	AliasType              string    `gorm:"column:aliasType;type:varchar(50);default:'synonym'" json:"aliasType"`    // primary, translation, synonym, typo
	CreatedAt              time.Time `gorm:"column:createdAt;not null;default:now()" json:"createdAt"`

	// Связь
	CanonicalIngredient *CanonicalIngredient `gorm:"foreignKey:CanonicalIngredientID" json:"canonicalIngredient,omitempty"`
}

func (IngredientAlias) TableName() string {
	return "IngredientAlias"
}

// AliasType constants
const (
	AliasTypePrimary     = "primary"     // Основное название
	AliasTypeTranslation = "translation" // Перевод
	AliasTypeSynonym     = "synonym"     // Синоним
	AliasTypeTypo        = "typo"        // Распространенная опечатка
)

// ============================================================================
// DTO для API
// ============================================================================

// CanonicalIngredientDTO - полная информация о продукте со всеми алиасами
type CanonicalIngredientDTO struct {
	ID                   string                  `json:"id"`
	CanonicalKey         string                  `json:"canonicalKey"`
	CanonicalName        string                  `json:"canonicalName"`
	Category             string                  `json:"category"`
	NutritionGroup       string                  `json:"nutritionGroup"`
	BaseUnit             string                  `json:"baseUnit"`
	DefaultShelfLifeDays *int                    `json:"defaultShelfLifeDays,omitempty"`
	DefaultPricePerUnit  *float64                `json:"defaultPricePerUnit,omitempty"`
	ImageURL             *string                 `json:"imageUrl,omitempty"`
	Status               string                  `json:"status"`
	Names                map[string]string       `json:"names"` // {"pl": "Cebula", "en": "Onion", "ru": "Лук"}
	AllAliases           []IngredientAliasSimple `json:"aliases,omitempty"`
	CreatedAt            time.Time               `json:"createdAt"`
	UpdatedAt            time.Time               `json:"updatedAt"`
}

// IngredientAliasSimple - упрощенная информация об алиасе
type IngredientAliasSimple struct {
	Name      string  `json:"name"`
	Language  *string `json:"language,omitempty"`
	AliasType string  `json:"aliasType"`
}

// IngredientSearchResult - результат поиска для автокомплита
type IngredientSearchResult struct {
	ID            string  `json:"id"`
	CanonicalKey  string  `json:"canonicalKey"`
	DisplayName   string  `json:"displayName"` // Название на запрошенном языке
	Category      string  `json:"category"`
	Unit          string  `json:"unit"`
	MatchedAlias  string  `json:"matchedAlias,omitempty"` // Какой алиас совпал с поиском
}

// ============================================================================
// Хелперы
// ============================================================================

// GetNameForLanguage возвращает название продукта на указанном языке
func (ci *CanonicalIngredient) GetNameForLanguage(lang string) string {
	if len(ci.Aliases) == 0 {
		return ci.CanonicalName
	}

	// Ищем primary alias для языка
	for _, alias := range ci.Aliases {
		if alias.Language != nil && *alias.Language == lang && alias.AliasType == AliasTypePrimary {
			return alias.Name
		}
	}

	// Ищем любой translation для языка
	for _, alias := range ci.Aliases {
		if alias.Language != nil && *alias.Language == lang && alias.AliasType == AliasTypeTranslation {
			return alias.Name
		}
	}

	// Fallback на каноническое имя
	return ci.CanonicalName
}

// ToDTO конвертирует CanonicalIngredient в DTO
func (ci *CanonicalIngredient) ToDTO() *CanonicalIngredientDTO {
	dto := &CanonicalIngredientDTO{
		ID:                   ci.ID,
		CanonicalKey:         ci.CanonicalKey,
		CanonicalName:        ci.CanonicalName,
		Category:             ci.Category,
		NutritionGroup:       ci.NutritionGroup,
		BaseUnit:             ci.BaseUnit,
		DefaultShelfLifeDays: ci.DefaultShelfLifeDays,
		DefaultPricePerUnit:  ci.DefaultPricePerUnit,
		ImageURL:             ci.ImageURL,
		Status:               ci.Status,
		Names:                make(map[string]string),
		AllAliases:           []IngredientAliasSimple{},
		CreatedAt:            ci.CreatedAt,
		UpdatedAt:            ci.UpdatedAt,
	}

	// Собираем названия по языкам
	if ci.Aliases != nil {
		for _, alias := range ci.Aliases {
			// Добавляем в AllAliases
			dto.AllAliases = append(dto.AllAliases, IngredientAliasSimple{
				Name:      alias.Name,
				Language:  alias.Language,
				AliasType: alias.AliasType,
			})

			// Заполняем Names для primary/translation
			if alias.Language != nil && (alias.AliasType == AliasTypePrimary || alias.AliasType == AliasTypeTranslation) {
				if _, exists := dto.Names[*alias.Language]; !exists {
					dto.Names[*alias.Language] = alias.Name
				}
			}
		}
	}

	return dto
}
