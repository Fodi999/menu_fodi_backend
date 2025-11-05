package models

import (
	"time"
)

// Recipe represents a culinary recipe posted by a user
type Recipe struct {
	ID           string    `json:"id" gorm:"type:varchar(255);primaryKey"`
	Title        string    `json:"title" gorm:"type:varchar(255);not null"`
	Description  string    `json:"description" gorm:"type:text"`
	ImageUrl     string    `json:"imageUrl" gorm:"type:varchar(500)"`
	AuthorID     string    `json:"authorId" gorm:"type:varchar(255);not null;index"`
	Author       User      `json:"author" gorm:"foreignKey:AuthorID;references:ID"`
	
	// Nutrition & Metrics
	GrossWeight  *int      `json:"grossWeight,omitempty" gorm:"column:gross_weight"`  // Брутто (г)
	NetWeight    *int      `json:"netWeight,omitempty" gorm:"column:net_weight"`      // Нетто (г)
	Calories     *int      `json:"calories,omitempty" gorm:"column:calories"`         // ккал
	Protein      *float64  `json:"protein,omitempty" gorm:"column:protein"`           // Белки (г)
	Fats         *float64  `json:"fats,omitempty" gorm:"column:fats"`                 // Жиры (г)
	Carbs        *float64  `json:"carbs,omitempty" gorm:"column:carbs"`               // Углеводы (г)
	RecipeYield  *int      `json:"yield,omitempty" gorm:"column:yield"`               // Выход (г)
	Cost         *float64  `json:"cost,omitempty" gorm:"column:cost"`                 // Цена (PLN)
	
	// ChefTokens System
	TokensReward *int      `json:"tokensReward,omitempty" gorm:"column:tokens_reward;default:10"` // Награда за создание
	ViewsCount   int       `json:"viewsCount" gorm:"column:views_count;default:0"`                 // Просмотры
	TokensEarned int       `json:"tokensEarned" gorm:"column:tokens_earned;default:0"`             // Заработано токенов
	
	CreatedAt    time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for Recipe model
func (Recipe) TableName() string {
	return "Recipe"
}
