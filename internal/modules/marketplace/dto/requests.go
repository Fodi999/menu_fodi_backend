package dto

import (
	"time"

	"github.com/google/uuid"
)

// MarketplaceFilters represents filters for marketplace search
type MarketplaceFilters struct {
	Category   string  `json:"category,omitempty"`
	Difficulty string  `json:"difficulty,omitempty"`
	MaxPrice   float64 `json:"maxPrice,omitempty"`
	MinRating  float64 `json:"minRating,omitempty"`
	SortBy     string  `json:"sortBy,omitempty"` // popular, newest, rating, price
	Limit      int     `json:"limit,omitempty"`
}

// RecipeWithAuthor represents a marketplace recipe with author info
type RecipeWithAuthor struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"userId"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Category     string    `json:"category"`
	Difficulty   string    `json:"difficulty"`
	Price        float64   `json:"price"`
	Rating       float64   `json:"rating"`
	Purchases    int       `json:"purchases"`
	ImageURL     string    `json:"imageUrl"`
	IsPublic     bool      `json:"isPublic"`
	CreatedAt    time.Time `json:"createdAt"`
	
	// Author info
	AuthorName   string  `json:"authorName"`
	AuthorLevel  int     `json:"authorLevel"`
	AuthorAvatar string  `json:"authorAvatar"`
	
	// Review info
	ReviewCount  int     `json:"reviewCount"`
	AvgReview    float64 `json:"avgReview"`
}

// MarketplaceResponse represents marketplace recipes list
type MarketplaceResponse struct {
	Recipes []RecipeWithAuthor `json:"recipes"`
	Total   int                `json:"total"`
}

// PurchaseRequest represents recipe purchase request
type PurchaseRequest struct {
	RecipeID string `json:"recipeId" validate:"required"`
	BuyerID  string `json:"buyerId" validate:"required"`
}

// PurchaseResponse represents purchase result
type PurchaseResponse struct {
	PurchaseID      string  `json:"purchaseId"`
	Recipe          string  `json:"recipe"`
	Price           float64 `json:"price"`
	Commission      float64 `json:"commission"`
	SellerReceived  float64 `json:"sellerReceived"`
	BuyerBalance    float64 `json:"buyerBalance"`
}

// UserPurchase represents a user's recipe purchase
type UserPurchase struct {
	PurchaseID  string    `json:"purchaseId"`
	RecipeID    uuid.UUID `json:"recipeId"`
	Title       string    `json:"title"`
	Category    string    `json:"category"`
	ImageURL    string    `json:"imageUrl"`
	Price       float64   `json:"price"`
	PurchasedAt time.Time `json:"purchasedAt"`
}

// SellerStats represents seller statistics
type SellerStats struct {
	SellerID       string  `json:"sellerId"`
	TotalSales     int     `json:"totalSales"`
	TotalRevenue   float64 `json:"totalRevenue"`
	AverageRating  float64 `json:"averageRating"`
	TopRecipe      string  `json:"topRecipe"`
	TopRecipeSales int     `json:"topRecipeSales"`
}

// LeaderboardEntry represents a chef in the leaderboard
type LeaderboardEntry struct {
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

// LeaderboardResponse represents leaderboard data
type LeaderboardResponse struct {
	Leaders []LeaderboardEntry `json:"leaders"`
	Total   int                `json:"total"`
	SortBy  string             `json:"sortBy"`
}

// UploadImageResponse represents image upload response
type UploadImageResponse struct {
	Success   bool   `json:"success"`
	URL       string `json:"url"`
	SecureURL string `json:"secureUrl"`
	PublicID  string `json:"publicId"`
	Message   string `json:"message,omitempty"`
}
