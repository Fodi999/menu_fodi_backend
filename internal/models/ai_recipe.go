package models

import (
	"time"

	"github.com/google/uuid"
)

// AIGeneratedRecipe represents a completed recipe created via AI Chef Mentor
type AIGeneratedRecipe struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SessionID   uuid.UUID `gorm:"type:uuid;index" json:"sessionId"`                    // Link to chef_mentor_session
	UserID      *uuid.UUID `gorm:"type:uuid;index" json:"userId,omitempty"`            // Creator (can be null for anonymous)
	
	// Recipe Data
	Title       string  `gorm:"type:varchar(255);not null" json:"title"`
	Description string  `gorm:"type:text" json:"description,omitempty"`
	Category    string  `gorm:"type:varchar(50)" json:"category,omitempty"`            // sushi, ramen, desserts, etc.
	Difficulty  string  `gorm:"type:varchar(20)" json:"difficulty,omitempty"`          // easy, intermediate, hard
	Language    string  `gorm:"type:varchar(5);default:'ua'" json:"language"`
	
	// Ingredients & Steps (JSONB)
	Ingredients JSONB `gorm:"type:jsonb;not null" json:"ingredients"`                  // [{name, amount, unit, gross, net}]
	Steps       JSONB `gorm:"type:jsonb" json:"steps,omitempty"`                       // [step1, step2, ...]
	
	// Nutrition Data (JSONB)
	Nutrition   JSONB `gorm:"type:jsonb" json:"nutrition"`                             // {calories, protein, fats, carbs}
	
	// Metrics
	Cost        float64 `gorm:"type:decimal(10,2)" json:"cost"`                        // Estimated cost in UAH
	Yield       int     `gorm:"default:0" json:"yield"`                                // Total yield in grams
	GrossWeight int     `gorm:"default:0" json:"grossWeight"`                          // Gross weight in grams
	NetWeight   int     `gorm:"default:0" json:"netWeight"`                            // Net weight in grams
	Time        int     `gorm:"default:0" json:"time"`                                 // Cooking time in minutes
	Portions    int     `gorm:"default:1" json:"portions"`                             // Number of servings
	
	// Publishing & Sharing
	IsPublic    bool   `gorm:"default:false" json:"isPublic"`                          // Public in marketplace
	ShareURL    string `gorm:"type:varchar(100);unique" json:"shareUrl,omitempty"`     // Unique share link
	
	// Analytics
	ViewsCount     int `gorm:"default:0" json:"viewsCount"`
	LikesCount     int `gorm:"default:0" json:"likesCount"`
	DownloadsCount int `gorm:"default:0" json:"downloadsCount"`
	
	// Timestamps
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName sets table name
func (AIGeneratedRecipe) TableName() string {
	return "ai_generated_recipes"
}

// RecipeLike represents a user liking a recipe
type RecipeLike struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RecipeID  uuid.UUID `gorm:"type:uuid;not null;index" json:"recipeId"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName sets table name
func (RecipeLike) TableName() string {
	return "recipe_likes"
}
