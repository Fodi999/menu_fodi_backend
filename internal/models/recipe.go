package models

import (
	"time"

	"gorm.io/datatypes"
)

// Recipe represents a culinary recipe posted by a user
type Recipe struct {
	ID            string  `json:"id" gorm:"type:varchar(255);primaryKey"`
	CanonicalName *string `json:"canonicalName,omitempty" gorm:"column:canonicalName;type:varchar(255)"` // Optional for user recipes
	LocalName     string  `json:"localName" gorm:"column:localName;type:varchar(255);not null;default:''"`
	Title         string  `json:"title" gorm:"column:title;type:varchar(255);not null"`
	Description   string  `json:"description" gorm:"column:description;type:text"`
	ImageUrl      string  `json:"imageUrl" gorm:"column:imageUrl;type:text"`
	ImagePublicId string  `json:"imagePublicId,omitempty" gorm:"column:imagePublicId;type:text"` // Cloudinary public ID

	// Recipe Metadata (shared with catalog recipes)
	Country     string         `json:"country" gorm:"column:country;type:varchar(100);not null"`
	Category    string         `json:"category" gorm:"column:category;type:varchar(50);not null"`
	Difficulty  string         `json:"difficulty" gorm:"column:difficulty;type:varchar(20);not null"`
	TimeMinutes int            `json:"timeMinutes" gorm:"column:timeMinutes;not null"`
	Servings    int            `json:"servings" gorm:"column:servings;not null;default:1"`
	Source      datatypes.JSON `json:"source" gorm:"column:source;type:jsonb;not null;default:'{\"type\":\"manual\"}'"`
	Status      string         `json:"status" gorm:"column:status;type:varchar(20);not null;default:'draft'"` // draft, published, archived

	AuthorID string `json:"authorId" gorm:"column:author_id;type:varchar(255);not null;index"`
	Author   User   `json:"author" gorm:"foreignKey:AuthorID;references:ID"`

	// Nutrition & Metrics
	GrossWeight *int     `json:"grossWeight,omitempty" gorm:"column:gross_weight"` // Брутто (г)
	NetWeight   *int     `json:"netWeight,omitempty" gorm:"column:net_weight"`     // Нетто (г)
	Calories    *int     `json:"calories,omitempty" gorm:"column:calories"`        // ккал
	Protein     *float64 `json:"protein,omitempty" gorm:"column:protein"`          // Белки (г)
	Fats        *float64 `json:"fats,omitempty" gorm:"column:fats"`                // Жиры (г)
	Carbs       *float64 `json:"carbs,omitempty" gorm:"column:carbs"`              // Углеводы (г)
	RecipeYield *int     `json:"yield,omitempty" gorm:"column:yield"`              // Выход (г)
	Cost        *float64 `json:"cost,omitempty" gorm:"column:cost"`                // Цена (PLN)

	// ChefTokens System
	TokensReward *int `json:"tokensReward,omitempty" gorm:"column:tokens_reward;default:10"` // Награда за создание
	ViewsCount   int  `json:"viewsCount" gorm:"column:views_count;default:0"`                // Просмотры
	TokensEarned int  `json:"tokensEarned" gorm:"column:tokens_earned;default:0"`            // Заработано токенов

	CreatedAt time.Time `json:"createdAt" gorm:"column:createdAt;autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updatedAt;autoUpdateTime"`
}

// TableName specifies the table name for Recipe model
func (Recipe) TableName() string {
	return "Recipe"
}
