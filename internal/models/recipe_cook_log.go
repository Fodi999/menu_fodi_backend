package models

import (
	"time"

	"github.com/google/uuid"
)

// RecipeCookLog tracks when users cook recipes
type RecipeCookLog struct {
	ID                 uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID             string    `gorm:"column:userId;type:text;not null;index" json:"userId"`
	RecipeID           uuid.UUID `gorm:"column:recipeId;type:uuid;not null;index" json:"recipeId"`
	
	// Cooking details
	ServingsMultiplier float64   `gorm:"column:servingsMultiplier;type:decimal(10,2);not null;default:1.0" json:"servingsMultiplier"`
	CookedAt           time.Time `gorm:"column:cookedAt;type:timestamp with time zone;not null;default:now();index:,sort:desc" json:"cookedAt"`
	
	// Economy snapshot
	UsedValue          float64   `gorm:"column:usedValue;type:decimal(10,2);not null;default:0" json:"usedValue"`
	WasteRiskSaved     float64   `gorm:"column:wasteRiskSaved;type:decimal(10,2);not null;default:0" json:"wasteRiskSaved"`
	TotalRecipeCost    float64   `gorm:"column:totalRecipeCost;type:decimal(10,2);not null;default:0" json:"totalRecipeCost"`
	
	// Idempotency
	IdempotencyKey     *string   `gorm:"column:idempotencyKey;type:varchar(255);uniqueIndex" json:"idempotencyKey,omitempty"`
	
	// Metadata
	CreatedAt          time.Time `gorm:"column:createdAt;type:timestamp with time zone;not null;default:now()" json:"createdAt"`
	UpdatedAt          time.Time `gorm:"column:updatedAt;type:timestamp with time zone;not null;default:now()" json:"updatedAt"`
	
	// Relations
	User               User                   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Recipe             RecipeCatalog          `gorm:"foreignKey:RecipeID" json:"recipe,omitempty"`
	Ingredients        []RecipeCookIngredient `gorm:"foreignKey:CookLogID" json:"ingredients,omitempty"`
}

// TableName specifies the table name
func (RecipeCookLog) TableName() string {
	return "RecipeCookLog"
}

// RecipeCookIngredient tracks individual ingredients deducted when cooking
type RecipeCookIngredient struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CookLogID        uuid.UUID `gorm:"column:cookLogId;type:uuid;not null;index" json:"cookLogId"`
	IngredientID     string    `gorm:"column:ingredientId;type:text;not null;index" json:"ingredientId"`
	
	// Deduction details
	QuantityUsed     float64   `gorm:"column:quantityUsed;type:decimal(10,3);not null" json:"quantityUsed"`
	Unit             string    `gorm:"type:varchar(50);not null" json:"unit"`
	PricePerUnit     *float64  `gorm:"column:pricePerUnit;type:decimal(10,4)" json:"pricePerUnit,omitempty"`
	TotalCost        *float64  `gorm:"column:totalCost;type:decimal(10,2)" json:"totalCost,omitempty"`
	
	// Waste prevention tracking
	WasExpiringSoon  bool      `gorm:"column:wasExpiringSoon;not null;default:false" json:"wasExpiringSoon"`
	
	// Metadata
	CreatedAt        time.Time `gorm:"column:createdAt;type:timestamp with time zone;not null;default:now()" json:"createdAt"`
	
	// Relations
	CookLog          RecipeCookLog `gorm:"foreignKey:CookLogID" json:"cookLog,omitempty"`
	Ingredient       Ingredient    `gorm:"foreignKey:IngredientID" json:"ingredient,omitempty"`
}

// TableName specifies the table name
func (RecipeCookIngredient) TableName() string {
	return "RecipeCookIngredient"
}
