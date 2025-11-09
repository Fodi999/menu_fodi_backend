package dto

// LeaderboardEntry represents a single leaderboard entry
type LeaderboardEntry struct {
	Rank             int     `json:"rank"`
	UserID           string  `json:"userId"`
	Name             string  `json:"name"`
	AvatarURL        string  `json:"avatarUrl"`
	Level            int     `json:"level"`
	XP               int     `json:"xp"`
	Stars            int     `json:"stars"`
	WalletBalance    float64 `json:"walletBalance"`
	CompletedCourses int     `json:"completedCourses"`
	RecipeCount      int     `json:"recipeCount"`
	AchievementCount int     `json:"achievementCount"`
}

// LeaderboardResponse response with leaderboard data
type LeaderboardResponse struct {
	SortBy      string             `json:"sortBy"`
	Total       int                `json:"total"`
	Leaderboard []LeaderboardEntry `json:"leaderboard"`
}
