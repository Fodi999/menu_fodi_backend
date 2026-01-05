package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// RecipeCatalog represents a structured recipe from catalog (NOT user-generated)
type RecipeCatalog struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CanonicalName    string         `gorm:"column:canonicalName;type:varchar(255);not null;uniqueIndex" json:"canonicalName"`
	LocalName        string         `gorm:"column:localName;type:varchar(255);not null" json:"localName"` // DEPRECATED: use Name* fields
	Title            string         `gorm:"column:title;type:varchar(255);not null" json:"title"` // Primary title (unified, typically Polish)
	
	// Multilingual names (like Ingredient model)
	NamePl           *string        `gorm:"column:name_pl;type:varchar(255)" json:"namePl,omitempty"`
	NameEn           *string        `gorm:"column:name_en;type:varchar(255)" json:"nameEn,omitempty"`
	NameRu           *string        `gorm:"column:name_ru;type:varchar(255)" json:"nameRu,omitempty"`
	
	// Multilingual descriptions
	DescriptionPl    *string        `gorm:"column:description_pl;type:text" json:"descriptionPl,omitempty"`
	DescriptionEn    *string        `gorm:"column:description_en;type:text" json:"descriptionEn,omitempty"`
	DescriptionRu    *string        `gorm:"column:description_ru;type:text" json:"descriptionRu,omitempty"`
	
	Country          string         `gorm:"type:varchar(100);not null;index" json:"country"`
	Region           *string        `gorm:"type:varchar(100)" json:"region,omitempty"`
	Category         string         `gorm:"type:varchar(50);not null;index" json:"category"`   // appetizer, main, dessert, soup, salad
	Difficulty          string         `gorm:"type:varchar(20);not null;index" json:"difficulty"` // easy, medium, hard
	TimeMinutes         int            `gorm:"column:timeMinutes;not null;index" json:"timeMinutes"`
	Servings            int            `gorm:"not null;default:1" json:"servings"` // Always 1 (base portion), use servingsMultiplier for scaling
	PortionWeightGrams  *int           `gorm:"column:portionWeightGrams" json:"portionWeightGrams,omitempty"` // Total weight of one serving in grams
	Steps               datatypes.JSON `gorm:"type:jsonb;not null;default:'[]'" json:"steps"`                           // [{"step":1,"instruction":"..."}] - DEPRECATED: use Steps* fields
	
	// Multilingual steps (cooking instructions)
	StepsPl          datatypes.JSON `gorm:"column:steps_pl;type:jsonb" json:"stepsPl,omitempty"`
	StepsEn          datatypes.JSON `gorm:"column:steps_en;type:jsonb" json:"stepsEn,omitempty"`
	StepsRu          datatypes.JSON `gorm:"column:steps_ru;type:jsonb" json:"stepsRu,omitempty"`
	
	NutritionProfile datatypes.JSON `gorm:"column:nutritionProfile;type:jsonb;default:'{}'" json:"nutritionProfile"` // {"type":"balanced","calories":450}
	Source           datatypes.JSON `gorm:"type:jsonb;not null" json:"source"`                                       // {"type":"cookbook","reference":"..."}
	CreatedAt        time.Time      `gorm:"column:createdAt;not null;default:now()" json:"createdAt"`
	UpdatedAt        time.Time      `gorm:"column:updatedAt;not null;default:now()" json:"updatedAt"`

	// Associations
	Ingredients []CatalogIngredient `gorm:"foreignKey:RecipeID" json:"ingredients,omitempty"`
	Allergens   []Allergen          `gorm:"many2many:RecipeAllergen;joinForeignKey:RecipeID;joinReferences:AllergenID" json:"allergens,omitempty"`
	DietTags    []DietTag           `gorm:"many2many:RecipeDietTag;joinForeignKey:RecipeID;joinReferences:DietTagID" json:"dietTags,omitempty"`
	RecipeSteps []RecipeStep        `gorm:"foreignKey:RecipeID;references:ID" json:"-"` // New relational steps (not exported to JSON by default)
}

func (RecipeCatalog) TableName() string {
	return "Recipe" // Uses same table as migration 035
}

// GetLocalizedName returns recipe name in the requested language
// Falls back to English if requested language is not available
func (r *RecipeCatalog) GetLocalizedName(lang string) string {
	switch lang {
	case "ru":
		if r.NameRu != nil && *r.NameRu != "" {
			return *r.NameRu
		}
	case "pl":
		if r.NamePl != nil && *r.NamePl != "" {
			return *r.NamePl
		}
	case "en":
		if r.NameEn != nil && *r.NameEn != "" {
			return *r.NameEn
		}
	}
	
	// Fallback chain: EN -> PL -> RU -> CanonicalName
	if r.NameEn != nil && *r.NameEn != "" {
		return *r.NameEn
	}
	if r.NamePl != nil && *r.NamePl != "" {
		return *r.NamePl
	}
	if r.NameRu != nil && *r.NameRu != "" {
		return *r.NameRu
	}
	return r.CanonicalName
}

// GetLocalizedDescription returns recipe description in the requested language
// Falls back to English if requested language is not available
func (r *RecipeCatalog) GetLocalizedDescription(lang string) string {
	switch lang {
	case "ru":
		if r.DescriptionRu != nil && *r.DescriptionRu != "" {
			return *r.DescriptionRu
		}
	case "pl":
		if r.DescriptionPl != nil && *r.DescriptionPl != "" {
			return *r.DescriptionPl
		}
	case "en":
		if r.DescriptionEn != nil && *r.DescriptionEn != "" {
			return *r.DescriptionEn
		}
	}
	
	// Fallback chain: EN -> PL -> RU -> empty
	if r.DescriptionEn != nil && *r.DescriptionEn != "" {
		return *r.DescriptionEn
	}
	if r.DescriptionPl != nil && *r.DescriptionPl != "" {
		return *r.DescriptionPl
	}
	if r.DescriptionRu != nil && *r.DescriptionRu != "" {
		return *r.DescriptionRu
	}
	return ""
}

// GetLocalizedSteps returns recipe steps in the requested language
// Prefers RecipeSteps table (new relational structure) over JSONB columns (legacy)
func (r *RecipeCatalog) GetLocalizedSteps(lang string) []string {
	// Priority 1: Use new RecipeSteps table if available
	if len(r.RecipeSteps) > 0 {
		steps := GetStepsForRecipe(r.RecipeSteps, lang)
		if len(steps) > 0 {
			return steps
		}
		
		// Fallback to English if requested language not found
		if lang != "en" {
			steps = GetStepsForRecipe(r.RecipeSteps, "en")
			if len(steps) > 0 {
				return steps
			}
		}
		
		// Fallback to Polish
		if lang != "pl" {
			steps = GetStepsForRecipe(r.RecipeSteps, "pl")
			if len(steps) > 0 {
				return steps
			}
		}
		
		// Fallback to Russian
		if lang != "ru" {
			steps = GetStepsForRecipe(r.RecipeSteps, "ru")
			if len(steps) > 0 {
				return steps
			}
		}
	}
	
	// Priority 2: Fallback to legacy JSONB columns
	return r.getLegacySteps(lang)
}

// getLegacySteps returns steps from old JSONB columns (backward compatibility)
func (r *RecipeCatalog) getLegacySteps(lang string) []string {
	var stepsJSON datatypes.JSON
	
	switch lang {
	case "ru":
		if len(r.StepsRu) > 0 {
			stepsJSON = r.StepsRu
		}
	case "pl":
		if len(r.StepsPl) > 0 {
			stepsJSON = r.StepsPl
		}
	case "en":
		if len(r.StepsEn) > 0 {
			stepsJSON = r.StepsEn
		}
	}
	
	// Fallback chain: EN -> PL -> RU -> Steps (legacy)
	if len(stepsJSON) == 0 && len(r.StepsEn) > 0 {
		stepsJSON = r.StepsEn
	}
	if len(stepsJSON) == 0 && len(r.StepsPl) > 0 {
		stepsJSON = r.StepsPl
	}
	if len(stepsJSON) == 0 && len(r.StepsRu) > 0 {
		stepsJSON = r.StepsRu
	}
	if len(stepsJSON) == 0 {
		stepsJSON = r.Steps
	}
	
	// Parse JSONB array to []string
	var steps []string
	if len(stepsJSON) > 0 {
		// datatypes.JSON is just []byte, use json.Unmarshal
		if err := json.Unmarshal(stepsJSON, &steps); err == nil {
			return steps
		}
	}
	
	return []string{}
}

// CatalogIngredient represents ingredient requirement in a catalog recipe
type CatalogIngredient struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RecipeID      uuid.UUID `gorm:"column:recipeId;type:uuid;not null;index" json:"recipeId"`
	IngredientID  string    `gorm:"column:ingredientId;type:text;not null" json:"ingredientId"`                 // Changed to TEXT to match Ingredient.id type
	IngredientKey string    `gorm:"column:ingredientKey;type:varchar(255);not null;index" json:"ingredientKey"` // normalized key for matching
	Quantity      float64   `gorm:"type:decimal(10,2);not null" json:"quantity"`
	Unit          string    `gorm:"type:varchar(50);not null" json:"unit"`
	Optional      bool      `gorm:"default:false" json:"optional"`
	SortOrder     int       `gorm:"column:sortOrder;default:0" json:"sortOrder"`
	CreatedAt     time.Time `gorm:"column:createdAt;not null;default:now()" json:"createdAt"`

	// Runtime fields (not stored in database)
	InFridge       bool     `gorm:"-" json:"inFridge"`       // True if ingredient is available in user's fridge (sufficient quantity)
	FridgeQuantity *float64 `gorm:"-" json:"fridgeQuantity"` // How much of this ingredient is in fridge (nil if not in fridge)

	// Associations
	Ingredient Ingredient    `gorm:"foreignKey:IngredientID" json:"ingredient,omitempty"`
	Recipe     RecipeCatalog `gorm:"foreignKey:RecipeID" json:"-"`
}

func (CatalogIngredient) TableName() string {
	return "RecipeIngredient"
}

// Allergen represents food allergen
type Allergen struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"type:varchar(50);not null;uniqueIndex" json:"name"`
	DisplayName string    `gorm:"type:varchar(100);not null" json:"displayName"`
	Icon        *string   `gorm:"type:varchar(50)" json:"icon,omitempty"`
	CreatedAt   time.Time `gorm:"not null;default:now()" json:"createdAt"`
}

func (Allergen) TableName() string {
	return "Allergen"
}

// DietTag represents dietary restriction/preference
type DietTag struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"type:varchar(50);not null;uniqueIndex" json:"name"`
	DisplayName string    `gorm:"type:varchar(100);not null" json:"displayName"`
	Description *string   `gorm:"type:text" json:"description,omitempty"`
	CreatedAt   time.Time `gorm:"not null;default:now()" json:"createdAt"`
}

func (DietTag) TableName() string {
	return "DietTag"
}

// RecipeAllergen junction table
type RecipeAllergen struct {
	RecipeID   uuid.UUID `gorm:"type:uuid;not null;primaryKey" json:"recipeId"`
	AllergenID uuid.UUID `gorm:"type:uuid;not null;primaryKey" json:"allergenId"`
}

func (RecipeAllergen) TableName() string {
	return "RecipeAllergen"
}

// RecipeDietTag junction table
type RecipeDietTag struct {
	RecipeID  uuid.UUID `gorm:"type:uuid;not null;primaryKey" json:"recipeId"`
	DietTagID uuid.UUID `gorm:"type:uuid;not null;primaryKey" json:"dietTagId"`
}

func (RecipeDietTag) TableName() string {
	return "RecipeDietTag"
}
