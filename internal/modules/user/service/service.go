package service

import (
	"errors"

	"github.com/google/uuid"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/user/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/user/repo"
)

var (
	ErrInvalidUpdateData = errors.New("invalid update data")
)

// UserService handles user business logic
type UserService interface {
	GetProfile(userID uuid.UUID) (*dto.UserProfileResponse, error)
	UpdateProfile(userID uuid.UUID, req dto.UpdateProfileRequest) error
	GetUserProgress(userID uuid.UUID) ([]dto.UserProgressResponse, error)
	GetDashboard(userID uuid.UUID) (*dto.DashboardResponse, error)
	GetAchievements(userID uuid.UUID) ([]dto.AchievementResponse, error)
	GetWallet(userID uuid.UUID) (*dto.WalletResponse, error)
}

type userService struct {
	repo repo.UserRepository
}

// NewUserService creates new user service
func NewUserService(repo repo.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) GetProfile(userID uuid.UUID) (*dto.UserProfileResponse, error) {
	profile, err := s.repo.GetProfile(userID)
	if err != nil {
		if errors.Is(err, repo.ErrProfileNotFound) {
			// Create default profile if not exists
			profile, err = s.repo.CreateProfile(userID)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return &dto.UserProfileResponse{
		UserID:           profile.UserID,
		Name:             profile.Name,
		Email:            profile.Email,
		Level:            profile.Level,
		Stars:            profile.Stars,
		XP:               profile.XP,
		Role:             profile.Role,
		Language:         profile.Language,
		AvatarURL:        profile.AvatarURL,
		CompletedCourses: profile.CompletedCourses,
		WalletBalance:    profile.WalletBalance,
	}, nil
}

func (s *userService) UpdateProfile(userID uuid.UUID, req dto.UpdateProfileRequest) error {
	// Validate input
	if req.Name == "" && req.AvatarURL == "" && req.Language == "" {
		return ErrInvalidUpdateData
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.AvatarURL != "" {
		updates["avatar_url"] = req.AvatarURL
	}
	if req.Language != "" {
		updates["language"] = req.Language
	}

	return s.repo.UpdateProfile(userID, updates)
}

func (s *userService) GetUserProgress(userID uuid.UUID) ([]dto.UserProgressResponse, error) {
	progress, err := s.repo.GetUserProgress(userID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.UserProgressResponse, len(progress))
	for i, p := range progress {
		// Calculate progress percentage
		progressPct := float64(0)
		if p.TotalLessons > 0 {
			progressPct = (float64(p.CompletedLessons) / float64(p.TotalLessons)) * 100
		}

		responses[i] = dto.UserProgressResponse{
			ID:               p.ID,
			UserID:           p.UserID,
			CourseID:         p.CourseID,
			CompletedLessons: p.CompletedLessons,
			TotalLessons:     p.TotalLessons,
			Progress:         progressPct,
			LastAccessedAt:   p.LastAccessedAt,
		}
	}

	return responses, nil
}

func (s *userService) GetDashboard(userID uuid.UUID) (*dto.DashboardResponse, error) {
	// Get user profile
	profile, err := s.repo.GetProfile(userID)
	if err != nil {
		if errors.Is(err, repo.ErrProfileNotFound) {
			profile, err = s.repo.CreateProfile(userID)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// Calculate XP progress to next level
	nextLevelXP := profile.Level * 500
	progressToNextLevel := float64(0)
	if nextLevelXP > 0 {
		progressToNextLevel = (float64(profile.XP) / float64(nextLevelXP)) * 100
	}

	// Get total published courses (optional, return 0 if error)
	totalCourses, _ := s.repo.GetTotalPublishedCourses()

	// Get recent course progress (optional, return empty list if error)
	courseProgress, _ := s.repo.GetRecentCourseProgress(userID, 5)
	if courseProgress == nil {
		courseProgress = []dto.CourseProgressInfo{}
	}

	// Get recent quizzes as activity (optional, return empty list if error)
	recentActivity, _ := s.repo.GetRecentQuizzes(userID, 5)
	if recentActivity == nil {
		recentActivity = []dto.ActivityInfo{}
	}

	// Get course recommendations (optional, return empty list if error)
	maxLevel := profile.Level + 2
	recommendations, _ := s.repo.GetCourseRecommendations(userID, profile.Language, maxLevel, 3)
	if recommendations == nil {
		recommendations = []dto.RecommendationInfo{}
	}

	// Get recent wallet transactions (optional, return empty list if error)
	recentTransactions, _ := s.repo.GetRecentTransactions(userID, 5)
	if recentTransactions == nil {
		recentTransactions = []dto.TransactionInfo{}
	}

	// Get active recipes (optional, return empty list if error)
	activeRecipes, _ := s.repo.GetActiveRecipes(userID, 3)
	if activeRecipes == nil {
		activeRecipes = []dto.ActiveRecipeInfo{}
	}

	return &dto.DashboardResponse{
		Profile: dto.UserProfileInfo{
			Level:            profile.Level,
			Stars:            profile.Stars,
			XP:               profile.XP,
			CompletedCourses: profile.CompletedCourses,
			WalletBalance:    profile.WalletBalance,
			Name:             profile.Name,
			AvatarURL:        profile.AvatarURL,
			Language:         profile.Language,
		},
		ProgressToNextLevel: progressToNextLevel,
		NextLevelXP:         nextLevelXP,
		TotalCourses:        totalCourses,
		CourseProgress:      courseProgress,
		RecentActivity:      recentActivity,
		Recommendations:     recommendations,
		RecentTransactions:  recentTransactions,
		ActiveRecipes:       activeRecipes,
	}, nil
}

func (s *userService) GetAchievements(userID uuid.UUID) ([]dto.AchievementResponse, error) {
	achievements, err := s.repo.GetUserAchievements(userID)
	if err != nil {
		// Return empty list instead of error if table doesn't exist
		return []dto.AchievementResponse{}, nil
	}
	if achievements == nil {
		return []dto.AchievementResponse{}, nil
	}
	return achievements, nil
}

func (s *userService) GetWallet(userID uuid.UUID) (*dto.WalletResponse, error) {
	// Get user profile for balance
	profile, err := s.repo.GetProfile(userID)
	if err != nil {
		if errors.Is(err, repo.ErrProfileNotFound) {
			profile, err = s.repo.CreateProfile(userID)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// Get recent transactions (optional)
	transactions, _ := s.repo.GetRecentTransactions(userID, 100)
	if transactions == nil {
		transactions = []dto.TransactionInfo{}
	}

	// Calculate earnings and spending
	totalEarned := float64(0)
	totalSpent := float64(0)
	lastTransaction := profile.UpdatedAt

	for _, txn := range transactions {
		if txn.Type == "credit" || txn.Type == "earned" {
			totalEarned += txn.Amount
		} else if txn.Type == "debit" || txn.Type == "spent" {
			totalSpent += txn.Amount
		}
		if txn.CreatedAt.After(lastTransaction) {
			lastTransaction = txn.CreatedAt
		}
	}

	return &dto.WalletResponse{
		UserID:          userID,
		Balance:         profile.WalletBalance,
		Currency:        "tokens",
		LastTransaction: lastTransaction,
		TotalEarned:     totalEarned,
		TotalSpent:      totalSpent,
		Earnings: dto.WalletEarnings{
			CoursesCompleted: totalEarned * 0.5,
			QuizzesCompleted: totalEarned * 0.3,
			Bonuses:          totalEarned * 0.2,
			Referrals:        0,
		},
		Spending: dto.WalletSpending{
			CourseEnrollments: totalSpent * 0.6,
			PremiumFeatures:   totalSpent * 0.3,
			Rewards:           totalSpent * 0.1,
		},
		TransactionCount: len(transactions),
	}, nil
}

// Helper to convert models.UserProfile to DTO
func profileToDTO(p *models.UserProfile) *dto.UserProfileResponse {
	return &dto.UserProfileResponse{
		UserID:           p.UserID,
		Name:             p.Name,
		Email:            p.Email,
		Level:            p.Level,
		Stars:            p.Stars,
		XP:               p.XP,
		Role:             p.Role,
		Language:         p.Language,
		AvatarURL:        p.AvatarURL,
		CompletedCourses: p.CompletedCourses,
		WalletBalance:    p.WalletBalance,
	}
}
