package repo

import (
	"errors"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/academy/dto"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrCourseNotFound     = errors.New("course not found")
	ErrLessonNotFound     = errors.New("lesson not found")
	ErrAlreadyEnrolled    = errors.New("already enrolled in this course")
	ErrNotEnrolled        = errors.New("not enrolled in this course")
	ErrCourseNotCompleted = errors.New("course not completed yet")
	ErrCertificateExists  = errors.New("certificate already exists")
	ErrNoQuizQuestions    = errors.New("no quiz questions found")
)

// AcademyRepository интерфейс для работы с данными academy
type AcademyRepository interface {
	// Courses
	GetCourses(filters *dto.CourseFilters) ([]*models.Course, error)
	GetCourseByID(courseID uuid.UUID) (*models.Course, error)

	// Lessons
	GetCourseLessons(courseID uuid.UUID) ([]*models.Lesson, error)
	GetLessonByID(lessonID uuid.UUID) (*models.Lesson, error)

	// Quiz
	GetQuizQuestions(courseID uuid.UUID) ([]*models.QuizQuestion, error)
	CreateUserQuiz(quiz *models.UserQuiz) error

	// Progress & Enrollment
	CheckEnrollment(userID, courseID uuid.UUID) (bool, error)
	CreateEnrollment(userID, courseID uuid.UUID) error
	GetUserProgress(userID, courseID uuid.UUID) (*models.UserProgress, error)
	CreateOrUpdateProgress(progress *models.UserProgress) error
	CompleteLesson(userID, lessonID uuid.UUID) error

	// Certificate
	GetCertificate(userID, courseID uuid.UUID) (*models.Certificate, error)
	CreateCertificate(cert *models.Certificate) error
	GetUserCertificates(userID uuid.UUID) ([]*models.Certificate, error)

	// User Profile
	GetUserProfile(userID uuid.UUID) (*models.UserProfile, error)
	UpdateUserProfile(profile *models.UserProfile) error
	AddWalletReward(userID uuid.UUID, amount float64, description string, relatedID uuid.UUID) error
}

type academyRepository struct {
	db *gorm.DB
}

// NewAcademyRepository создает новый репозиторий
func NewAcademyRepository(db *gorm.DB) AcademyRepository {
	return &academyRepository{db: db}
}

// ============================================================================
// Courses
// ============================================================================

func (r *academyRepository) GetCourses(filters *dto.CourseFilters) ([]*models.Course, error) {
	var courses []*models.Course
	query := r.db.Where("is_published = ?", true).Order("level ASC, created_at DESC")

	if filters != nil {
		if filters.Language != "" {
			query = query.Where("language = ?", filters.Language)
		}
		if filters.Category != "" {
			query = query.Where("category = ?", filters.Category)
		}
		if filters.Level > 0 {
			query = query.Where("level = ?", filters.Level)
		}
	}

	if err := query.Find(&courses).Error; err != nil {
		return nil, err
	}

	return courses, nil
}

func (r *academyRepository) GetCourseByID(courseID uuid.UUID) (*models.Course, error) {
	var course models.Course
	if err := r.db.Where("id = ?", courseID).First(&course).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCourseNotFound
		}
		return nil, err
	}
	return &course, nil
}

// ============================================================================
// Lessons
// ============================================================================

func (r *academyRepository) GetCourseLessons(courseID uuid.UUID) ([]*models.Lesson, error) {
	var lessons []*models.Lesson
	if err := r.db.Where("course_id = ? AND is_published = ?", courseID, true).
		Order("\"order\" ASC").
		Find(&lessons).Error; err != nil {
		return nil, err
	}
	return lessons, nil
}

func (r *academyRepository) GetLessonByID(lessonID uuid.UUID) (*models.Lesson, error) {
	var lesson models.Lesson
	if err := r.db.Where("id = ?", lessonID).First(&lesson).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLessonNotFound
		}
		return nil, err
	}
	return &lesson, nil
}

// ============================================================================
// Quiz
// ============================================================================

func (r *academyRepository) GetQuizQuestions(courseID uuid.UUID) ([]*models.QuizQuestion, error) {
	var questions []*models.QuizQuestion
	if err := r.db.Where("course_id = ?", courseID).Find(&questions).Error; err != nil {
		return nil, err
	}

	if len(questions) == 0 {
		return nil, ErrNoQuizQuestions
	}

	return questions, nil
}

func (r *academyRepository) CreateUserQuiz(quiz *models.UserQuiz) error {
	return r.db.Create(quiz).Error
}

// ============================================================================
// Progress & Enrollment
// ============================================================================

func (r *academyRepository) CheckEnrollment(userID, courseID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.UserProgress{}).
		Where("user_id = ? AND course_id = ?", userID, courseID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *academyRepository) CreateEnrollment(userID, courseID uuid.UUID) error {
	progress := &models.UserProgress{
		UserID:   userID,
		CourseID: courseID,
	}
	return r.db.Create(progress).Error
}

func (r *academyRepository) GetUserProgress(userID, courseID uuid.UUID) (*models.UserProgress, error) {
	var progress models.UserProgress
	err := r.db.Where("user_id = ? AND course_id = ?", userID, courseID).
		First(&progress).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Прогресс еще не создан
		}
		return nil, err
	}

	return &progress, nil
}

func (r *academyRepository) CreateOrUpdateProgress(progress *models.UserProgress) error {
	// Пытаемся найти существующий прогресс
	var existing models.UserProgress
	err := r.db.Where("user_id = ? AND course_id = ?", progress.UserID, progress.CourseID).
		First(&existing).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Создаем новый
			return r.db.Create(progress).Error
		}
		return err
	}

	// Обновляем существующий
	return r.db.Model(&existing).Updates(map[string]interface{}{
		"completed_lessons": progress.CompletedLessons,
		"total_lessons":     progress.TotalLessons,
		"quiz_score":        progress.QuizScore,
		"stars_earned":      progress.StarsEarned,
		"is_completed":      progress.IsCompleted,
		"completed_at":      progress.CompletedAt,
	}).Error
}

func (r *academyRepository) CompleteLesson(userID, lessonID uuid.UUID) error {
	// Получаем урок
	lesson, err := r.GetLessonByID(lessonID)
	if err != nil {
		return err
	}

	// Получаем или создаем прогресс
	progress, err := r.GetUserProgress(userID, lesson.CourseID)
	if err != nil {
		return err
	}

	if progress == nil {
		// Создаем новый прогресс
		progress = &models.UserProgress{
			UserID:           userID,
			CourseID:         lesson.CourseID,
			CompletedLessons: 1,
		}
	} else {
		// Увеличиваем счетчик
		progress.CompletedLessons++
	}

	// Получаем общее количество уроков
	var totalLessons int64
	r.db.Model(&models.Lesson{}).
		Where("course_id = ? AND is_published = ?", lesson.CourseID, true).
		Count(&totalLessons)

	progress.TotalLessons = int(totalLessons)

	return r.CreateOrUpdateProgress(progress)
}

// ============================================================================
// Certificate
// ============================================================================

func (r *academyRepository) GetCertificate(userID, courseID uuid.UUID) (*models.Certificate, error) {
	var cert models.Certificate
	err := r.db.Where("user_id = ? AND course_id = ?", userID, courseID).
		First(&cert).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &cert, nil
}

func (r *academyRepository) CreateCertificate(cert *models.Certificate) error {
	return r.db.Create(cert).Error
}

func (r *academyRepository) GetUserCertificates(userID uuid.UUID) ([]*models.Certificate, error) {
	var certs []*models.Certificate
	if err := r.db.Where("user_id = ?", userID).
		Order("issued_at DESC").
		Find(&certs).Error; err != nil {
		return nil, err
	}
	return certs, nil
}

// ============================================================================
// User Profile
// ============================================================================

func (r *academyRepository) GetUserProfile(userID uuid.UUID) (*models.UserProfile, error) {
	var profile models.UserProfile
	if err := r.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *academyRepository) UpdateUserProfile(profile *models.UserProfile) error {
	return r.db.Save(profile).Error
}

func (r *academyRepository) AddWalletReward(userID uuid.UUID, amount float64, description string, relatedID uuid.UUID) error {
	// Обновляем баланс
	if err := r.db.Model(&models.UserProfile{}).
		Where("user_id = ?", userID).
		Update("wallet_balance", gorm.Expr("wallet_balance + ?", amount)).Error; err != nil {
		return err
	}

	// Создаем транзакцию
	transaction := &models.WalletTransaction{
		UserID:      userID,
		Amount:      amount,
		Type:        "reward",
		Description: description,
		RelatedID:   relatedID,
		CreatedAt:   time.Now(),
	}

	return r.db.Create(transaction).Error
}
