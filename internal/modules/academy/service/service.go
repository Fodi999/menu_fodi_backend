package service

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/academy/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/academy/repo"
	"github.com/google/uuid"
)

const (
	QuizQuestionCount = 10 // Количество вопросов в тесте
	TokensPerStar     = 10 // ChefTokens за 1 звезду
)

// AcademyService бизнес-логика академии
type AcademyService struct {
	repo    repo.AcademyRepository
	certSvc *CertificateService
}

// NewAcademyService создает новый сервис
func NewAcademyService(repo repo.AcademyRepository) *AcademyService {
	return &AcademyService{
		repo:    repo,
		certSvc: NewCertificateService(),
	}
}

// ============================================================================
// Courses
// ============================================================================

// GetCourses возвращает список курсов с фильтрацией
func (s *AcademyService) GetCourses(filters *dto.CourseFilters) ([]*dto.CourseResponse, error) {
	courses, err := s.repo.GetCourses(filters)
	if err != nil {
		return nil, err
	}

	response := make([]*dto.CourseResponse, len(courses))
	for i, course := range courses {
		response[i] = &dto.CourseResponse{
			ID:          course.ID,
			Title:       course.Title,
			Description: course.Description,
			Category:    course.Category,
			Level:       course.Level,
			Language:    course.Language,
			Duration:    course.Duration,
			IsPublished: course.IsPublished,
			CreatedAt:   course.CreatedAt,
		}
	}

	return response, nil
}

// GetCourse возвращает детали курса
func (s *AcademyService) GetCourse(courseID uuid.UUID) (*dto.CourseResponse, error) {
	course, err := s.repo.GetCourseByID(courseID)
	if err != nil {
		return nil, err
	}

	return &dto.CourseResponse{
		ID:          course.ID,
		Title:       course.Title,
		Description: course.Description,
		Category:    course.Category,
		Level:       course.Level,
		Language:    course.Language,
		Duration:    course.Duration,
		IsPublished: course.IsPublished,
		CreatedAt:   course.CreatedAt,
	}, nil
}

// ============================================================================
// Lessons
// ============================================================================

// GetCourseLessons возвращает уроки курса
func (s *AcademyService) GetCourseLessons(courseID uuid.UUID) ([]*dto.LessonResponse, error) {
	lessons, err := s.repo.GetCourseLessons(courseID)
	if err != nil {
		return nil, err
	}

	response := make([]*dto.LessonResponse, len(lessons))
	for i, lesson := range lessons {
		response[i] = &dto.LessonResponse{
			ID:          lesson.ID,
			CourseID:    lesson.CourseID,
			Title:       lesson.Title,
			Content:     lesson.Content,
			VideoURL:    lesson.VideoURL,
			Order:       lesson.Order,
			Duration:    lesson.Duration,
			IsPublished: lesson.IsPublished,
			CreatedAt:   lesson.CreatedAt,
		}
	}

	return response, nil
}

// GetLesson возвращает детали урока
func (s *AcademyService) GetLesson(lessonID uuid.UUID) (*dto.LessonResponse, error) {
	lesson, err := s.repo.GetLessonByID(lessonID)
	if err != nil {
		return nil, err
	}

	return &dto.LessonResponse{
		ID:          lesson.ID,
		CourseID:    lesson.CourseID,
		Title:       lesson.Title,
		Content:     lesson.Content,
		VideoURL:    lesson.VideoURL,
		Order:       lesson.Order,
		Duration:    lesson.Duration,
		IsPublished: lesson.IsPublished,
		CreatedAt:   lesson.CreatedAt,
	}, nil
}

// ============================================================================
// Quiz
// ============================================================================

// GetQuiz возвращает случайные вопросы для теста
func (s *AcademyService) GetQuiz(courseID uuid.UUID) ([]*dto.QuizQuestionResponse, error) {
	allQuestions, err := s.repo.GetQuizQuestions(courseID)
	if err != nil {
		return nil, err
	}

	// Выбираем случайные вопросы
	questionCount := QuizQuestionCount
	if len(allQuestions) < questionCount {
		questionCount = len(allQuestions)
	}

	// Перемешиваем
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(allQuestions), func(i, j int) {
		allQuestions[i], allQuestions[j] = allQuestions[j], allQuestions[i]
	})

	selectedQuestions := allQuestions[:questionCount]

	// Конвертируем в DTO (без правильного ответа)
	response := make([]*dto.QuizQuestionResponse, len(selectedQuestions))
	for i, q := range selectedQuestions {
		response[i] = &dto.QuizQuestionResponse{
			ID:          q.ID,
			CourseID:    q.CourseID,
			Question:    q.Question,
			Options:     q.Options,
			Explanation: q.Explanation,
		}
	}

	return response, nil
}

// SubmitQuiz проверяет ответы и начисляет награды
func (s *AcademyService) SubmitQuiz(userID uuid.UUID, request *dto.QuizRequest) (*dto.QuizResultResponse, error) {
	// Получаем вопросы
	questions, err := s.repo.GetQuizQuestions(request.CourseID)
	if err != nil {
		return nil, err
	}

	// Подсчитываем правильные ответы
	correctCount := 0
	for i, answer := range request.Answers {
		if i < len(questions) && answer == questions[i].CorrectAnswer {
			correctCount++
		}
	}

	totalQuestions := len(request.Answers)
	score := (correctCount * 100) / totalQuestions

	// Рассчитываем звезды (0-5)
	stars := calculateStars(score)

	// Сохраняем результат
	userQuiz := &models.UserQuiz{
		UserID:         userID,
		CourseID:       request.CourseID,
		Score:          score,
		TotalQuestions: totalQuestions,
		CorrectAnswers: correctCount,
		Answers:        convertToInt32Array(request.Answers),
		StarsEarned:    stars,
		CompletedAt:    time.Now(),
	}

	if err := s.repo.CreateUserQuiz(userQuiz); err != nil {
		return nil, err
	}

	// Обновляем профиль пользователя
	profile, err := s.repo.GetUserProfile(userID)
	if err == nil {
		profile.Stars += stars
		profile.XP += score
		s.repo.UpdateUserProfile(profile)

		// Награда в ChefTokens
		if stars > 0 {
			reward := float64(stars * TokensPerStar)
			description := fmt.Sprintf("Quiz completion reward: %d stars", stars)
			s.repo.AddWalletReward(userID, reward, description, request.CourseID)
		}
	}

	// Обновляем прогресс по курсу
	progress, _ := s.repo.GetUserProgress(userID, request.CourseID)
	if progress != nil {
		progress.QuizScore = score
		progress.StarsEarned = stars

		// Проверяем завершение курса
		if progress.CompletedLessons >= progress.TotalLessons && score >= 50 {
			progress.IsCompleted = true
			now := time.Now()
			progress.CompletedAt = &now
		}

		s.repo.CreateOrUpdateProgress(progress)
	}

	return &dto.QuizResultResponse{
		Score:          score,
		CorrectAnswers: correctCount,
		TotalQuestions: totalQuestions,
		StarsEarned:    stars,
		Reward:         stars * TokensPerStar,
	}, nil
}

// ============================================================================
// Progress & Enrollment
// ============================================================================

// EnrollInCourse записывает пользователя на курс
func (s *AcademyService) EnrollInCourse(userID, courseID uuid.UUID) error {
	// Проверяем, не записан ли уже
	enrolled, err := s.repo.CheckEnrollment(userID, courseID)
	if err != nil {
		return err
	}

	if enrolled {
		return repo.ErrAlreadyEnrolled
	}

	// Проверяем, что курс существует
	_, err = s.repo.GetCourseByID(courseID)
	if err != nil {
		return err
	}

	return s.repo.CreateEnrollment(userID, courseID)
}

// CompleteLesson отмечает урок как завершенный
func (s *AcademyService) CompleteLesson(userID, lessonID uuid.UUID) error {
	return s.repo.CompleteLesson(userID, lessonID)
}

// GetUserProgress возвращает прогресс пользователя
func (s *AcademyService) GetUserProgress(userID, courseID uuid.UUID) (*dto.UserProgressResponse, error) {
	progress, err := s.repo.GetUserProgress(userID, courseID)
	if err != nil {
		return nil, err
	}

	if progress == nil {
		return nil, repo.ErrNotEnrolled
	}

	return &dto.UserProgressResponse{
		UserID:           progress.UserID,
		CourseID:         progress.CourseID,
		CompletedLessons: progress.CompletedLessons,
		TotalLessons:     progress.TotalLessons,
		QuizScore:        progress.QuizScore,
		StarsEarned:      progress.StarsEarned,
		IsCompleted:      progress.IsCompleted,
		CompletedAt:      progress.CompletedAt,
		LastAccessedAt:   progress.LastAccessedAt,
	}, nil
}

// ============================================================================
// Certificate
// ============================================================================

// GenerateCertificate генерирует PDF-сертификат
func (s *AcademyService) GenerateCertificate(userID, courseID uuid.UUID) (*dto.CertificateResponse, error) {
	// Проверяем, есть ли уже сертификат
	existingCert, err := s.repo.GetCertificate(userID, courseID)
	if err != nil {
		return nil, err
	}

	if existingCert != nil {
		// Сертификат уже существует
		return &dto.CertificateResponse{
			ID:         existingCert.ID,
			UserID:     existingCert.UserID,
			CourseID:   existingCert.CourseID,
			CourseName: existingCert.CourseName,
			UserName:   existingCert.UserName,
			Level:      existingCert.Level,
			Stars:      existingCert.Stars,
			PDFURL:     existingCert.PDFURL,
			Signature:  existingCert.Signature,
			IssuedAt:   existingCert.IssuedAt,
		}, nil
	}

	// Проверяем прогресс
	progress, err := s.repo.GetUserProgress(userID, courseID)
	if err != nil {
		return nil, err
	}

	if progress == nil {
		return nil, repo.ErrNotEnrolled
	}

	if !progress.IsCompleted {
		return nil, repo.ErrCourseNotCompleted
	}

	// Получаем данные курса и профиля
	course, err := s.repo.GetCourseByID(courseID)
	if err != nil {
		return nil, err
	}

	profile, err := s.repo.GetUserProfile(userID)
	if err != nil {
		return nil, err
	}

	// Генерируем PDF
	certData := CertificateData{
		StudentName:    profile.Name,
		CourseName:     course.Title,
		Level:          profile.Level,
		Stars:          progress.StarsEarned,
		CompletionDate: *progress.CompletedAt,
		Language:       profile.Language,
		QuizScore:      progress.QuizScore,
	}

	pdfPath, err := s.certSvc.GenerateCertificate(certData)
	if err != nil {
		return nil, err
	}

	// Сохраняем в базу
	certificate := &models.Certificate{
		UserID:     userID,
		CourseID:   courseID,
		CourseName: course.Title,
		UserName:   profile.Name,
		Level:      profile.Level,
		Stars:      progress.StarsEarned,
		PDFURL:     pdfPath,
		Signature:  "Chef Dima Fomin - Culinary Academy AI",
		IssuedAt:   time.Now(),
	}

	if err := s.repo.CreateCertificate(certificate); err != nil {
		return nil, err
	}

	return &dto.CertificateResponse{
		ID:         certificate.ID,
		UserID:     certificate.UserID,
		CourseID:   certificate.CourseID,
		CourseName: certificate.CourseName,
		UserName:   certificate.UserName,
		Level:      certificate.Level,
		Stars:      certificate.Stars,
		PDFURL:     certificate.PDFURL,
		Signature:  certificate.Signature,
		IssuedAt:   certificate.IssuedAt,
	}, nil
}

// GetUserCertificates возвращает все сертификаты пользователя
func (s *AcademyService) GetUserCertificates(userID uuid.UUID) (*dto.CertificateListResponse, error) {
	certs, err := s.repo.GetUserCertificates(userID)
	if err != nil {
		return nil, err
	}

	response := make([]*dto.CertificateResponse, len(certs))
	for i, cert := range certs {
		response[i] = &dto.CertificateResponse{
			ID:         cert.ID,
			UserID:     cert.UserID,
			CourseID:   cert.CourseID,
			CourseName: cert.CourseName,
			UserName:   cert.UserName,
			Level:      cert.Level,
			Stars:      cert.Stars,
			PDFURL:     cert.PDFURL,
			Signature:  cert.Signature,
			IssuedAt:   cert.IssuedAt,
		}
	}

	return &dto.CertificateListResponse{
		Certificates: response,
		Total:        len(response),
	}, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// calculateStars рассчитывает количество звезд по проценту
func calculateStars(score int) int {
	if score >= 90 {
		return 5
	} else if score >= 80 {
		return 4
	} else if score >= 70 {
		return 3
	} else if score >= 60 {
		return 2
	} else if score >= 50 {
		return 1
	}
	return 0
}

// convertToInt32Array конвертирует []int в []int32
func convertToInt32Array(arr []int) []int32 {
	result := make([]int32, len(arr))
	for i, v := range arr {
		result[i] = int32(v)
	}
	return result
}
