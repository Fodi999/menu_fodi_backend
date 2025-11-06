package models

import (
	"time"
)

// RecipePurchase покупка рецепта на маркетплейсе
type RecipePurchase struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RecipeID   string    `gorm:"type:uuid;not null" json:"recipeId"`
	BuyerID    string    `gorm:"type:uuid;not null" json:"buyerId"`
	SellerID   string    `gorm:"type:uuid;not null" json:"sellerId"`
	Price      float64   `gorm:"type:numeric(10,2);not null" json:"price"`
	Commission float64   `gorm:"type:numeric(10,2);default:0" json:"commission"` // 10% комиссия платформы
	NetAmount  float64   `gorm:"type:numeric(10,2);not null" json:"netAmount"`   // seller получает
	CreatedAt  time.Time `gorm:"default:now()" json:"createdAt"`

	// Связи
	Recipe PersonalRecipe `gorm:"foreignKey:RecipeID;references:ID" json:"recipe,omitempty"`
	Buyer  User           `gorm:"foreignKey:BuyerID;references:ID" json:"buyer,omitempty"`
	Seller User           `gorm:"foreignKey:SellerID;references:ID" json:"seller,omitempty"`
}

// TableName указывает имя таблицы
func (RecipePurchase) TableName() string {
	return "RecipePurchase"
}

// RecipeReview отзыв покупателя о рецепте
type RecipeReview struct {
	ID            string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RecipeID      string    `gorm:"type:uuid;not null" json:"recipeId"`
	UserID        string    `gorm:"type:uuid;not null" json:"userId"`
	Rating        float64   `gorm:"type:numeric(3,1);check:rating >= 0 AND rating <= 10" json:"rating"`
	Comment       string    `gorm:"type:text" json:"comment"`
	WouldBuyAgain bool      `gorm:"default:true" json:"wouldBuyAgain"`
	CreatedAt     time.Time `gorm:"default:now()" json:"createdAt"`
	UpdatedAt     time.Time `gorm:"default:now()" json:"updatedAt"`

	// Связи
	Recipe PersonalRecipe `gorm:"foreignKey:RecipeID;references:ID" json:"recipe,omitempty"`
	User   User           `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

// TableName указывает имя таблицы
func (RecipeReview) TableName() string {
	return "RecipeReview"
}

// MarketStats статистика продавца на маркетплейсе
type MarketStats struct {
	SellerID       string  `json:"sellerId"`
	TotalSales     int     `json:"totalSales"`
	TotalRevenue   float64 `json:"totalRevenue"`
	AverageRating  float64 `json:"averageRating"`
	TopRecipe      string  `json:"topRecipe"`
	TopRecipeSales int     `json:"topRecipeSales"`
}

// ChefLeaderboardEntry представляет запись в глобальном рейтинге поваров
type ChefLeaderboardEntry struct {
	Rank             int     `json:"rank"`
	UserID           string  `json:"userId"`
	Name             string  `json:"name"`
	AvatarURL        string  `json:"avatarUrl"`
	Level            int     `json:"level"`
	TotalXP          int     `json:"totalXp"`
	Language         string  `json:"language"`
	TotalSales       int     `json:"totalSales"`
	TotalRevenue     float64 `json:"totalRevenue"`
	AverageRating    float64 `json:"averageRating"`
	RecipeCount      int     `json:"recipeCount"`
	AchievementCount int     `json:"achievementCount"`
}

// LeaderboardResponse формат ответа API лидерборда
type LeaderboardResponse struct {
	Leaders []ChefLeaderboardEntry `json:"leaders"`
	Total   int                    `json:"total"`
	SortBy  string                 `json:"sortBy"`
}
