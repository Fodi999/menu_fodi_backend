package service

import (
	"fmt"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/leaderboard/dto"
	"gorm.io/gorm"
)

type LeaderboardService struct {
	db *gorm.DB
}

func NewLeaderboardService(db *gorm.DB) *LeaderboardService {
	return &LeaderboardService{db: db}
}

func (s *LeaderboardService) GetLeaderboard(sortBy string, limit int) (*dto.LeaderboardResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	if sortBy == "" {
		sortBy = "xp"
	}

	var profiles []models.UserProfile
	var orderClause string

	switch sortBy {
	case "tokens", "wallet":
		orderClause = "wallet_balance DESC"
	case "stars":
		orderClause = "stars DESC"
	case "level":
		orderClause = "level DESC, xp DESC"
	case "courses":
		orderClause = "completed_courses DESC"
	default:
		orderClause = "xp DESC"
		sortBy = "xp"
	}

	if err := s.db.Order(orderClause).Limit(limit).Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch leaderboard: %w", err)
	}

	entries := make([]dto.LeaderboardEntry, 0, len(profiles))
	for i, profile := range profiles {
		var recipeCount int64
		s.db.Model(&models.PersonalRecipe{}).Where("user_id = ?", profile.UserID).Count(&recipeCount)

		var achievementCount int64
		s.db.Model(&models.UserAchievement{}).Where("user_id = ?", profile.UserID).Count(&achievementCount)

		entry := dto.LeaderboardEntry{
			Rank:             i + 1,
			UserID:           profile.UserID.String(),
			Name:             profile.Name,
			AvatarURL:        profile.AvatarURL,
			Level:            profile.Level,
			XP:               profile.XP,
			Stars:            profile.Stars,
			WalletBalance:    profile.WalletBalance,
			CompletedCourses: profile.CompletedCourses,
			RecipeCount:      int(recipeCount),
			AchievementCount: int(achievementCount),
		}
		entries = append(entries, entry)
	}

	return &dto.LeaderboardResponse{
		SortBy:      sortBy,
		Total:       len(entries),
		Leaderboard: entries,
	}, nil
}
