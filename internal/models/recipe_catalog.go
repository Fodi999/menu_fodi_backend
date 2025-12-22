package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// RecipeCatalog represents a structured recipe from catalog (NOT user-generated)
type RecipeCatalog struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CanonicalName    string         `gorm:"column:canonicalName;type:varchar(255);not null;uniqueIndex" json:"canonicalName"`
	LocalName        string         `gorm:"column:localName;type:varchar(255);not null" json:"localName"`
	Country          string         `gorm:"type:varchar(100);not null;index" json:"country"`
	Region           *string        `gorm:"type:varchar(100)" json:"region,omitempty"`
	Category         string         `gorm:"type:varchar(50);not null;index" json:"category"`   // appetizer, main, dessert, soup, salad
	Difficulty       string         `gorm:"type:varchar(20);not null;index" json:"difficulty"` // easy, medium, hard
	TimeMinutes      int            `gorm:"column:timeMinutes;not null;index" json:"timeMinutes"`
	Servings         int            `gorm:"not null;default:1" json:"servings"` // Always 1 (base portion), use servingsMultiplier for scaling
	Steps            datatypes.JSON `gorm:"type:jsonb;not null;default:'[]'" json:"steps"`                           // [{"step":1,"instruction":"..."}]
	NutritionProfile datatypes.JSON `gorm:"column:nutritionProfile;type:jsonb;default:'{}'" json:"nutritionProfile"` // {"type":"balanced","calories":450}
	Source           datatypes.JSON `gorm:"type:jsonb;not null" json:"source"`                                       // {"type":"cookbook","reference":"..."}
	CreatedAt        time.Time      `gorm:"column:createdAt;not null;default:now()" json:"createdAt"`
	UpdatedAt        time.Time      `gorm:"column:updatedAt;not null;default:now()" json:"updatedAt"`

	// Associations
	Ingredients []CatalogIngredient `gorm:"foreignKey:RecipeID" json:"ingredients,omitempty"`
	Allergens   []Allergen          `gorm:"many2many:RecipeAllergen;joinForeignKey:RecipeID;joinReferences:AllergenID" json:"allergens,omitempty"`
	DietTags    []DietTag           `gorm:"many2many:RecipeDietTag;joinForeignKey:RecipeID;joinReferences:DietTagID" json:"dietTags,omitempty"`
}

func (RecipeCatalog) TableName() string {
	return "Recipe" // Uses same table as migration 035
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
