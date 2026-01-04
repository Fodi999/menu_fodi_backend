package repo

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/user/dto"
)

var (
	ErrProfileNotFound = errors.New("user profile not found")
	ErrCourseNotFound  = errors.New("course not found")
)

// UserRepository handles user data operations
type UserRepository interface {
	// Profile operations
	GetProfile(userID uuid.UUID) (*models.UserProfile, error)
	CreateProfile(userID uuid.UUID) (*models.UserProfile, error)
	UpdateProfile(userID uuid.UUID, updates map[string]interface{}) error

	// Progress operations
	GetUserProgress(userID uuid.UUID) ([]models.UserProgress, error)
	GetUserCertificates(userID uuid.UUID) ([]models.Certificate, error)

	// Dashboard aggregations
	GetTotalPublishedCourses() (int64, error)
	GetRecentCourseProgress(userID uuid.UUID, limit int) ([]dto.CourseProgressInfo, error)
	GetRecentQuizzes(userID uuid.UUID, limit int) ([]dto.ActivityInfo, error)
	GetCourseRecommendations(userID uuid.UUID, language string, maxLevel int, limit int) ([]dto.RecommendationInfo, error)
	GetRecentTransactions(userID uuid.UUID, limit int) ([]dto.TransactionInfo, error)
	GetActiveRecipes(userID uuid.UUID, limit int) ([]dto.ActiveRecipeInfo, error)

	// Achievements
	GetUserAchievements(userID uuid.UUID) ([]dto.AchievementResponse, error)

	// Settings
	GetUserByID(userID uuid.UUID) (*models.User, error)
	UpdateSettings(userID uuid.UUID, settings models.UserSettings) error
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates new user repository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetProfile(userID uuid.UUID) (*models.UserProfile, error) {
	var profile models.UserProfile
	err := r.db.Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}
	
	// Get the actual role from User table (source of truth)
	var user models.User
	if err := r.db.Where("id = ?", userID.String()).First(&user).Error; err == nil {
		profile.Role = user.Role // Update with current role from User table
	}
	
	return &profile, nil
}

func (r *userRepository) CreateProfile(userID uuid.UUID) (*models.UserProfile, error) {
	// First, get the user to extract name, email, and role
	var user models.User
	if err := r.db.Where("id = ?", userID.String()).First(&user).Error; err != nil {
		return nil, err
	}

	profile := &models.UserProfile{
		UserID:        userID,
		Name:          user.Name,
		Email:         user.Email,
		Role:          user.Role, // Set role from User table (source of truth)
		Level:         1,
		Stars:         0,
		XP:            0,
		WalletBalance: 0,
	}
	err := r.db.Create(profile).Error
	if err != nil {
		return nil, err
	}
	return profile, nil
}

func (r *userRepository) UpdateProfile(userID uuid.UUID, updates map[string]interface{}) error {
	return r.db.Model(&models.UserProfile{}).
		Where("user_id = ?", userID).
		Updates(updates).Error
}

func (r *userRepository) GetUserProgress(userID uuid.UUID) ([]models.UserProgress, error) {
	var progress []models.UserProgress
	err := r.db.Where("user_id = ?", userID).
		Order("last_accessed_at DESC").
		Find(&progress).Error
	return progress, err
}

func (r *userRepository) GetUserCertificates(userID uuid.UUID) ([]models.Certificate, error) {
	var certificates []models.Certificate
	err := r.db.Where("user_id = ?", userID).
		Order("issued_at DESC").
		Find(&certificates).Error
	return certificates, err
}

func (r *userRepository) GetTotalPublishedCourses() (int64, error) {
	var count int64
	err := r.db.Model(&models.Course{}).
		Where("status = ?", "published").
		Count(&count).Error
	return count, err
}

func (r *userRepository) GetRecentCourseProgress(userID uuid.UUID, limit int) ([]dto.CourseProgressInfo, error) {
	var results []dto.CourseProgressInfo

	err := r.db.Table("user_progress").
		Select("user_progress.course_id, courses.name as course_name, "+
			"user_progress.completed_lessons, user_progress.total_lessons, "+
			"user_progress.progress, user_progress.last_accessed_at as last_accessed").
		Joins("JOIN courses ON courses.id = user_progress.course_id").
		Where("user_progress.user_id = ?", userID).
		Order("user_progress.last_accessed_at DESC").
		Limit(limit).
		Scan(&results).Error

	return results, err
}

func (r *userRepository) GetRecentQuizzes(userID uuid.UUID, limit int) ([]dto.ActivityInfo, error) {
	var results []dto.ActivityInfo

	err := r.db.Table("user_quizzes").
		Select("'quiz' as type, courses.name as course, "+
			"user_quizzes.stars_earned as stars, user_quizzes.score, "+
			"user_quizzes.completed_at as timestamp").
		Joins("JOIN courses ON courses.id = user_quizzes.course_id").
		Where("user_quizzes.user_id = ?", userID).
		Order("user_quizzes.completed_at DESC").
		Limit(limit).
		Scan(&results).Error

	return results, err
}

func (r *userRepository) GetCourseRecommendations(userID uuid.UUID, language string, maxLevel int, limit int) ([]dto.RecommendationInfo, error) {
	var courses []struct {
		ID          uuid.UUID
		Title       string
		Description string
		Level       int
		ImageURL    string
	}

	err := r.db.Table("courses").
		Select("id, title, description, level, image_url").
		Where("language = ? AND level <= ? AND status = ?", language, maxLevel, "published").
		Order("level ASC").
		Limit(limit).
		Scan(&courses).Error

	if err != nil {
		return nil, err
	}

	// Get user's current level for match calculation
	var userLevel int
	r.db.Model(&models.UserProfile{}).
		Select("level").
		Where("user_id = ?", userID).
		Scan(&userLevel)

	results := make([]dto.RecommendationInfo, len(courses))
	for i, course := range courses {
		match := 100
		if course.Level > userLevel {
			match = 85
		}
		results[i] = dto.RecommendationInfo{
			CourseID:    course.ID,
			Title:       course.Title,
			Description: course.Description,
			Level:       course.Level,
			Match:       match,
			ImageURL:    course.ImageURL,
		}
	}

	return results, nil
}

func (r *userRepository) GetRecentTransactions(userID uuid.UUID, limit int) ([]dto.TransactionInfo, error) {
	var results []dto.TransactionInfo

	err := r.db.Table("wallet_transactions").
		Select("id, amount, type, created_at").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Scan(&results).Error

	return results, err
}

func (r *userRepository) GetActiveRecipes(userID uuid.UUID, limit int) ([]dto.ActiveRecipeInfo, error) {
	var results []dto.ActiveRecipeInfo

	err := r.db.Table("personal_recipes").
		Select("id, name, progress, status").
		Where("user_id = ? AND status IN (?)", userID, []string{"active", "in_progress"}).
		Order("updated_at DESC").
		Limit(limit).
		Scan(&results).Error

	return results, err
}

func (r *userRepository) GetUserAchievements(userID uuid.UUID) ([]dto.AchievementResponse, error) {
	var results []dto.AchievementResponse

	err := r.db.Table("user_achievements").
		Select("achievements.id, achievements.code, achievements.title, "+
			"achievements.description, achievements.icon_url, achievements.category, "+
			"user_achievements.unlocked_at").
		Joins("JOIN achievements ON achievements.id = user_achievements.achievement_id").
		Where("user_achievements.user_id = ?", userID).
		Order("user_achievements.unlocked_at DESC").
		Scan(&results).Error

	return results, err
}

// GetUserByID retrieves user by ID
func (r *userRepository) GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", userID.String()).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// UpdateSettings updates user settings
func (r *userRepository) UpdateSettings(userID uuid.UUID, settings models.UserSettings) error {
	return r.db.Model(&models.User{}).
		Where("id = ?", userID.String()).
		Update("settings", settings).Error
}
