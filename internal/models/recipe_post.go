package models

import (
	"time"

	"github.com/google/uuid"
)

// RecipePost represents a recipe post on the social feed
type RecipePost struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID   uuid.UUID `gorm:"type:uuid;not null;index;constraint:OnDelete:CASCADE" json:"userId"`   // Author
	RecipeID uuid.UUID `gorm:"type:uuid;not null;index;constraint:OnDelete:CASCADE" json:"recipeId"` // Link to ai_generated_recipes

	// Post content
	Title       string `gorm:"type:varchar(255);not null" json:"title"`
	Description string `gorm:"type:text" json:"description,omitempty"`
	ImageURL    string `gorm:"type:varchar(500)" json:"imageUrl,omitempty"`

	// Recipe metrics (denormalized for performance)
	GrossWeight  int     `gorm:"default:0" json:"grossWeight,omitempty"`
	NetWeight    int     `gorm:"default:0" json:"netWeight,omitempty"`
	Calories     int     `gorm:"default:0" json:"calories,omitempty"`
	Protein      float64 `gorm:"type:decimal(10,2)" json:"protein,omitempty"`
	Fats         float64 `gorm:"type:decimal(10,2)" json:"fats,omitempty"`
	Carbs        float64 `gorm:"type:decimal(10,2)" json:"carbs,omitempty"`
	Yield        int     `gorm:"default:0" json:"yield,omitempty"`
	Cost         float64 `gorm:"type:decimal(10,2)" json:"cost,omitempty"`
	TokensReward int     `gorm:"default:0" json:"tokensReward,omitempty"` // Reward for views

	// Social metrics
	ViewsCount    int     `gorm:"default:0" json:"viewsCount"`
	LikesCount    int     `gorm:"default:0" json:"likesCount,omitempty"`
	CommentsCount int     `gorm:"default:0" json:"commentsCount,omitempty"`
	TokensEarned  float64 `gorm:"type:decimal(10,2);default:0" json:"tokensEarned"` // Total earned from views

	// Timestamps
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relations (not stored in DB, loaded via joins)
	Author *User              `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"author,omitempty"`
	Recipe *AIGeneratedRecipe `gorm:"foreignKey:RecipeID;constraint:OnDelete:CASCADE" json:"recipe,omitempty"`
}

// TableName sets table name
func (RecipePost) TableName() string {
	return "recipe_posts"
}

// PostComment represents a comment on a recipe post
type PostComment struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PostID    uuid.UUID `gorm:"type:uuid;not null;index" json:"postId"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`

	// Relations
	Author *User `gorm:"foreignKey:UserID" json:"author,omitempty"`
}

// TableName sets table name
func (PostComment) TableName() string {
	return "post_comments"
}

// PostLike represents a like on a recipe post
type PostLike struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PostID    uuid.UUID `gorm:"type:uuid;not null;index" json:"postId"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName sets table name
func (PostLike) TableName() string {
	return "post_likes"
}
