package models

import (
	"time"
)

// Recipe represents a culinary recipe posted by a user
type Recipe struct {
	ID          string    `json:"id" gorm:"type:varchar(255);primaryKey"`
	Title       string    `json:"title" gorm:"type:varchar(255);not null"`
	Description string    `json:"description" gorm:"type:text"`
	ImageUrl    string    `json:"imageUrl" gorm:"type:varchar(500)"`
	AuthorID    string    `json:"authorId" gorm:"type:varchar(255);not null;index"`
	Author      User      `json:"author" gorm:"foreignKey:AuthorID;references:ID"`
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for Recipe model
func (Recipe) TableName() string {
	return "Recipe"
}
