package dto

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Course DTOs
// ============================================================================

// CourseFilters - фильтры для списка курсов
type CourseFilters struct {
	Language string `json:"language"` // Фильтр по языку
	Category string `json:"category"` // Фильтр по категории
	Level    int    `json:"level"`    // Фильтр по уровню
}

// CourseResponse - ответ с деталями курса
type CourseResponse struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Level       int       `json:"level"`
	Language    string    `json:"language"`
	Duration    int       `json:"duration"` // Длительность в минутах
	IsPublished bool      `json:"isPublished"`
	CreatedAt   time.Time `json:"createdAt"`
}

// ============================================================================
// Lesson DTOs
// ============================================================================

// LessonResponse - ответ с деталями урока
type LessonResponse struct {
	ID          uuid.UUID `json:"id"`
	CourseID    uuid.UUID `json:"courseId"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	VideoURL    string    `json:"videoUrl,omitempty"`
	Order       int       `json:"order"`
	Duration    int       `json:"duration"` // Длительность в минутах
	IsPublished bool      `json:"isPublished"`
	CreatedAt   time.Time `json:"createdAt"`
}

// ============================================================================
// Quiz DTOs
// ============================================================================

// QuizQuestionResponse - вопрос теста
type QuizQuestionResponse struct {
	ID            uuid.UUID `json:"id"`
	CourseID      uuid.UUID `json:"courseId"`
	Question      string    `json:"question"`
	Options       []string  `json:"options"`     // Варианты ответов
	CorrectAnswer int       `json:"-"`           // Не возвращаем клиенту
	Explanation   string    `json:"explanation"` // Объяснение правильного ответа
}

// QuizRequest - запрос на прохождение теста
type QuizRequest struct {
	CourseID uuid.UUID `json:"courseId"`
	Answers  []int     `json:"answers"` // Индексы выбранных ответов
}

// QuizResultResponse - результат теста
type QuizResultResponse struct {
	Score          int `json:"score"`          // Процент правильных ответов
	CorrectAnswers int `json:"correctAnswers"` // Количество правильных ответов
	TotalQuestions int `json:"totalQuestions"` // Всего вопросов
	StarsEarned    int `json:"starsEarned"`    // Заработано звезд
	Reward         int `json:"reward"`         // Награда в ChefTokens
}

// ============================================================================
// Progress DTOs
// ============================================================================

// EnrollRequest - запрос на запись на курс
type EnrollRequest struct {
	CourseID uuid.UUID `json:"courseId"`
}

// CompleteLessonRequest - запрос на завершение урока
type CompleteLessonRequest struct {
	LessonID uuid.UUID `json:"lessonId"`
}

// UserProgressResponse - прогресс пользователя по курсу
type UserProgressResponse struct {
	UserID           uuid.UUID  `json:"userId"`
	CourseID         uuid.UUID  `json:"courseId"`
	CompletedLessons int        `json:"completedLessons"`
	TotalLessons     int        `json:"totalLessons"`
	QuizScore        int        `json:"quizScore"`
	StarsEarned      int        `json:"starsEarned"`
	IsCompleted      bool       `json:"isCompleted"`
	CompletedAt      *time.Time `json:"completedAt,omitempty"`
	LastAccessedAt   time.Time  `json:"lastAccessedAt"`
}

// ============================================================================
// Certificate DTOs
// ============================================================================

// CertificateRequest - запрос на генерацию сертификата
type CertificateRequest struct {
	CourseID uuid.UUID `json:"courseId"`
}

// CertificateResponse - сертификат пользователя
type CertificateResponse struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"userId"`
	CourseID   uuid.UUID `json:"courseId"`
	CourseName string    `json:"courseName"`
	UserName   string    `json:"userName"`
	Level      int       `json:"level"`
	Stars      int       `json:"stars"`
	PDFURL     string    `json:"pdfUrl"`
	Signature  string    `json:"signature"`
	IssuedAt   time.Time `json:"issuedAt"`
}

// CertificateListResponse - список сертификатов
type CertificateListResponse struct {
	Certificates []*CertificateResponse `json:"certificates"`
	Total        int                    `json:"total"`
}

// ============================================================================
// Enrollment DTOs
// ============================================================================

// EnrollmentResponse - информация о записи на курс
type EnrollmentResponse struct {
	ID         uuid.UUID       `json:"id"`
	UserID     uuid.UUID       `json:"userId"`
	CourseID   uuid.UUID       `json:"courseId"`
	EnrolledAt time.Time       `json:"enrolledAt"`
	Course     *CourseResponse `json:"course,omitempty"`
}
