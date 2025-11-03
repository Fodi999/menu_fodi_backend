package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Course курс в Culinary Academy
type Course struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title       string    `gorm:"type:varchar(255);not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	ImageURL    string    `gorm:"type:text" json:"imageUrl"`
	Level       int       `gorm:"not null" json:"level"`                         // 1-10 сложность
	Category    string    `gorm:"type:varchar(100)" json:"category"`             // sushi, sashimi, knife-skills, etc
	Duration    int       `gorm:"default:0" json:"duration"`                     // общая продолжительность в минутах
	LessonsCount int      `gorm:"default:0" json:"lessonsCount"`                 // количество уроков
	Language    string    `gorm:"type:varchar(5);default:'pl'" json:"language"`  // pl, ua, en
	IsPublished bool      `gorm:"default:false" json:"isPublished"`
	Instructor  string    `gorm:"type:varchar(255)" json:"instructor"`           // имя инструктора
	Stars       int       `gorm:"default:0" json:"stars"`                        // награда за прохождение
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName задаёт имя таблицы
func (Course) TableName() string {
	return "Course"
}

// Lesson урок внутри курса
type Lesson struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CourseID    uuid.UUID      `gorm:"type:uuid;not null" json:"courseId"`
	Title       string         `gorm:"type:varchar(255);not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	VideoURL    string         `gorm:"type:text" json:"videoUrl"`
	Order       int            `gorm:"not null" json:"order"`                     // порядок урока в курсе
	Duration    int            `gorm:"default:0" json:"duration"`                 // длительность в минутах
	Content     string         `gorm:"type:text" json:"content"`                  // текстовый контент
	Steps       pq.StringArray `gorm:"type:text[]" json:"steps"`                  // шаги выполнения
	IsPublished bool           `gorm:"default:false" json:"isPublished"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName задаёт имя таблицы
func (Lesson) TableName() string {
	return "Lesson"
}

// QuizQuestion вопрос теста
type QuizQuestion struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CourseID   uuid.UUID      `gorm:"type:uuid;not null" json:"courseId"`
	Question   string         `gorm:"type:text;not null" json:"question"`
	Options    pq.StringArray `gorm:"type:text[];not null" json:"options"`       // варианты ответа
	CorrectAnswer int         `gorm:"not null" json:"-"`                         // индекс правильного ответа (не показывать)
	Explanation string        `gorm:"type:text" json:"explanation"`              // объяснение ответа
	Difficulty  string        `gorm:"type:varchar(50);default:'medium'" json:"difficulty"` // easy, medium, hard
	Language    string        `gorm:"type:varchar(5);default:'pl'" json:"language"`
	CreatedAt   time.Time     `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName задаёт имя таблицы
func (QuizQuestion) TableName() string {
	return "QuizQuestion"
}

// UserQuiz результат прохождения теста
type UserQuiz struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null" json:"userId"`
	CourseID    uuid.UUID      `gorm:"type:uuid;not null" json:"courseId"`
	Score       int            `gorm:"not null" json:"score"`                     // процент правильных ответов
	TotalQuestions int         `gorm:"not null" json:"totalQuestions"`
	CorrectAnswers int         `gorm:"not null" json:"correctAnswers"`
	Answers     pq.Int32Array  `gorm:"type:integer[]" json:"answers"`             // индексы выбранных ответов
	StarsEarned int            `gorm:"default:0" json:"starsEarned"`
	CompletedAt time.Time      `gorm:"autoCreateTime" json:"completedAt"`
}

// TableName задаёт имя таблицы
func (UserQuiz) TableName() string {
	return "UserQuiz"
}

// MentorSession сессия с AI ментором
type MentorSession struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"userId"`
	Messages  int       `gorm:"default:0" json:"messages"`                     // количество сообщений в сессии
	Topic     string    `gorm:"type:varchar(255)" json:"topic"`                // тема обсуждения
	Language  string    `gorm:"type:varchar(5);default:'pl'" json:"language"`
	StartedAt time.Time `gorm:"autoCreateTime" json:"startedAt"`
	EndedAt   *time.Time `json:"endedAt,omitempty"`
}

// TableName задаёт имя таблицы
func (MentorSession) TableName() string {
	return "MentorSession"
}

// MentorMessage сообщение в чате с AI ментором
type MentorMessage struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SessionID uuid.UUID `gorm:"type:uuid;not null" json:"sessionId"`
	Role      string    `gorm:"type:varchar(20);not null" json:"role"`        // "user" или "assistant"
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName задаёт имя таблицы
func (MentorMessage) TableName() string {
	return "MentorMessage"
}
