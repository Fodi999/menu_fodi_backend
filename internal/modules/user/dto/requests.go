package dto

import (
	"time"

	"github.com/google/uuid"
)

// UpdateProfileRequest represents profile update request
type UpdateProfileRequest struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
	Language  string `json:"language"`
}

// UserProfileResponse represents user profile response
type UserProfileResponse struct {
	UserID           uuid.UUID `json:"userId"`
	Name             string    `json:"name"`
	Email            string    `json:"email"`
	Level            int       `json:"level"`
	Stars            int       `json:"stars"`
	XP               int       `json:"xp"`
	Role             string    `json:"role"`
	Language         string    `json:"language"`
	AvatarURL        string    `json:"avatarUrl"`
	CompletedCourses int       `json:"completedCourses"`
	WalletBalance    float64   `json:"walletBalance"`
}

// UserProgressResponse represents user course progress
type UserProgressResponse struct {
	ID               uuid.UUID `json:"id"`
	UserID           uuid.UUID `json:"userId"`
	CourseID         uuid.UUID `json:"courseId"`
	CompletedLessons int       `json:"completedLessons"`
	TotalLessons     int       `json:"totalLessons"`
	Progress         float64   `json:"progress"`
	LastAccessedAt   time.Time `json:"lastAccessedAt"`
}

// DashboardResponse represents user dashboard data
type DashboardResponse struct {
	Profile             UserProfileInfo      `json:"profile"`
	ProgressToNextLevel float64              `json:"progressToNextLevel"`
	NextLevelXP         int                  `json:"nextLevelXP"`
	TotalCourses        int64                `json:"totalCourses"`
	CourseProgress      []CourseProgressInfo `json:"courseProgress"`
	RecentActivity      []ActivityInfo       `json:"recentActivity"`
	Recommendations     []RecommendationInfo `json:"recommendations"`
	RecentTransactions  []TransactionInfo    `json:"recentTransactions"`
	ActiveRecipes       []ActiveRecipeInfo   `json:"activeRecipes"`

	// 🆕 Smart Kitchen Integration
	FridgeSummary *FridgeSummary `json:"fridgeSummary"` // Статистика холодильника
	ChefTokens    int            `json:"chefTokens"`    // Chef Tokens баланс
}

// UserProfileInfo for dashboard
type UserProfileInfo struct {
	Level            int     `json:"level"`
	Stars            int     `json:"stars"`
	XP               int     `json:"xp"`
	CompletedCourses int     `json:"completedCourses"`
	WalletBalance    float64 `json:"walletBalance"`
	Name             string  `json:"name"`
	AvatarURL        string  `json:"avatarUrl"`
	Language         string  `json:"language"`
}

// CourseProgressInfo represents course progress
type CourseProgressInfo struct {
	CourseID         uuid.UUID `json:"courseId"`
	CourseName       string    `json:"courseName"`
	CompletedLessons int       `json:"completedLessons"`
	TotalLessons     int       `json:"totalLessons"`
	Progress         float64   `json:"progress"`
	LastAccessed     time.Time `json:"lastAccessed"`
}

// ActivityInfo represents recent user activity
type ActivityInfo struct {
	Type      string    `json:"type"`
	Course    string    `json:"course"`
	Stars     int       `json:"stars,omitempty"`
	Score     int       `json:"score,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// RecommendationInfo represents course recommendation
type RecommendationInfo struct {
	CourseID    uuid.UUID `json:"courseId"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Level       int       `json:"level"`
	Match       int       `json:"match"`
	ImageURL    string    `json:"imageUrl"`
}

// TransactionInfo represents wallet transaction
type TransactionInfo struct {
	ID        uuid.UUID `json:"id"`
	Amount    float64   `json:"amount"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"createdAt"`
}

// ActiveRecipeInfo represents active recipe in kitchen
type ActiveRecipeInfo struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Progress int       `json:"progress"`
	Status   string    `json:"status"`
}

// FridgeSummary represents fridge statistics for dashboard
type FridgeSummary struct {
	TotalItems    int     `json:"totalItems"`    // Общее количество продуктов
	CriticalItems int     `json:"criticalItems"` // Продукты с истекающим сроком (≤2 дня)
	WarningItems  int     `json:"warningItems"`  // Продукты требующие внимания (3-5 дней)
	TotalValue    float64 `json:"totalValue"`    // Общая стоимость продуктов (PLN)
	PotentialLoss float64 `json:"potentialLoss"` // Потенциальная потеря (продукты critical)
	Currency      string  `json:"currency"`      // Валюта (PLN)
}

// AchievementResponse represents user achievement
type AchievementResponse struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	IconURL     string    `json:"iconUrl"`
	Category    string    `json:"category"`
	UnlockedAt  time.Time `json:"unlockedAt"`
}

// WalletResponse represents user wallet information
type WalletResponse struct {
	UserID           uuid.UUID      `json:"userId"`
	Balance          float64        `json:"balance"`
	Currency         string         `json:"currency"`
	LastTransaction  time.Time      `json:"lastTransaction"`
	TotalEarned      float64        `json:"totalEarned"`
	TotalSpent       float64        `json:"totalSpent"`
	Earnings         WalletEarnings `json:"earnings"`
	Spending         WalletSpending `json:"spending"`
	TransactionCount int            `json:"transactionCount"`
}

// WalletEarnings breakdown
type WalletEarnings struct {
	CoursesCompleted float64 `json:"coursesCompleted"`
	QuizzesCompleted float64 `json:"quizzesCompleted"`
	Bonuses          float64 `json:"bonuses"`
	Referrals        float64 `json:"referrals"`
}

// WalletSpending breakdown
type WalletSpending struct {
	CourseEnrollments float64 `json:"courseEnrollments"`
	PremiumFeatures   float64 `json:"premiumFeatures"`
	Rewards           float64 `json:"rewards"`
}
