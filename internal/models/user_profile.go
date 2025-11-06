package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// UserProfile расширенный профиль ученика в Culinary Academy
type UserProfile struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID           uuid.UUID      `gorm:"type:uuid;not null;unique" json:"userId"` // связь с основной таблицей User
	Name             string         `gorm:"type:varchar(255);not null" json:"name"`
	Email            string         `gorm:"type:varchar(255);not null;unique" json:"email"`
	AvatarURL        string         `gorm:"type:text" json:"avatarUrl"`
	Level            int            `gorm:"default:1" json:"level"`                               // уровень 1-10
	Stars            int            `gorm:"default:0" json:"stars"`                               // награды за тесты
	Role             string         `gorm:"type:varchar(50);default:'student'" json:"role"`       // student, mentor, admin
	Language         string         `gorm:"type:varchar(5);default:'pl'" json:"language"`         // pl, ua, en
	XP               int            `gorm:"default:0" json:"xp"`                                  // опыт (experience points)
	CompletedCourses int            `gorm:"default:0" json:"completedCourses"`                    // количество пройденных курсов
	FavoriteRecipes  pq.StringArray `gorm:"type:text[]" json:"favoriteRecipes"`                   // UUID рецептов
	WalletBalance    float64        `gorm:"type:decimal(10,2);default:0.00" json:"walletBalance"` // баланс ChefToken
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName задаёт имя таблицы
func (UserProfile) TableName() string {
	return "UserProfile"
}

// PersonalRecipe личный рецепт ученика
type PersonalRecipe struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null" json:"userId"`
	Title       string         `gorm:"type:varchar(255);not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Ingredients pq.StringArray `gorm:"type:text[]" json:"ingredients"`
	Steps       pq.StringArray `gorm:"type:text[]" json:"steps"`
	ImageURL    string         `gorm:"type:text" json:"imageUrl"`
	Category    string         `gorm:"type:varchar(100)" json:"category"`  // sushi, sashimi, maki, etc
	Difficulty  string         `gorm:"type:varchar(50)" json:"difficulty"` // easy, medium, hard
	CookingTime int            `gorm:"default:0" json:"cookingTime"`       // в минутах
	Servings    int            `gorm:"default:1" json:"servings"`
	Rating      float64        `gorm:"type:decimal(3,1);default:0.0" json:"rating"`  // AI рейтинг 0.0-10.0
	IsPublic    bool           `gorm:"default:false" json:"isPublic"`                // публичный для marketplace
	Price       float64        `gorm:"type:decimal(10,2);default:0.00" json:"price"` // цена в ChefToken
	Purchases   int            `gorm:"default:0" json:"purchases"`                   // количество покупок
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName задаёт имя таблицы
func (PersonalRecipe) TableName() string {
	return "PersonalRecipe"
}

// Certificate сертификат об окончании курса
type Certificate struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null" json:"userId"`
	CourseID   uuid.UUID `gorm:"type:uuid;not null" json:"courseId"`
	CourseName string    `gorm:"type:varchar(255);not null" json:"courseName"`
	UserName   string    `gorm:"type:varchar(255);not null" json:"userName"`
	Level      int       `gorm:"not null" json:"level"`              // уровень на момент получения
	Stars      int       `gorm:"not null" json:"stars"`              // звёзды за курс
	PDFURL     string    `gorm:"type:text" json:"pdfUrl"`            // ссылка на PDF
	Signature  string    `gorm:"type:varchar(255)" json:"signature"` // подпись "Dima Fomin AI Academy"
	IssuedAt   time.Time `gorm:"autoCreateTime" json:"issuedAt"`
}

// TableName задаёт имя таблицы
func (Certificate) TableName() string {
	return "Certificate"
}

// UserProgress прогресс ученика по курсам и урокам
type UserProgress struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID           uuid.UUID  `gorm:"type:uuid;not null" json:"userId"`
	CourseID         uuid.UUID  `gorm:"type:uuid;not null" json:"courseId"`
	LessonID         uuid.UUID  `gorm:"type:uuid" json:"lessonId"` // null если общий прогресс курса
	CompletedLessons int        `gorm:"default:0" json:"completedLessons"`
	TotalLessons     int        `gorm:"default:0" json:"totalLessons"`
	QuizScore        int        `gorm:"default:0" json:"quizScore"` // процент 0-100
	StarsEarned      int        `gorm:"default:0" json:"starsEarned"`
	IsCompleted      bool       `gorm:"default:false" json:"isCompleted"`
	LastAccessedAt   time.Time  `gorm:"autoUpdateTime" json:"lastAccessedAt"`
	CompletedAt      *time.Time `json:"completedAt,omitempty"`
}

// TableName задаёт имя таблицы
func (UserProgress) TableName() string {
	return "UserProgress"
}

// WalletTransaction транзакция ChefToken
type WalletTransaction struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null" json:"userId"`
	Amount      float64   `gorm:"type:decimal(10,2);not null" json:"amount"` // положительное = получение, отрицательное = трата
	Type        string    `gorm:"type:varchar(50);not null" json:"type"`     // reward, purchase, refund, bonus
	Description string    `gorm:"type:text" json:"description"`
	RelatedID   uuid.UUID `gorm:"type:uuid" json:"relatedId,omitempty"` // ID курса, рецепта и т.д.
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName задаёт имя таблицы
func (WalletTransaction) TableName() string {
	return "WalletTransaction"
}

// MarketPurchase покупка рецепта в marketplace
type MarketPurchase struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BuyerID   uuid.UUID `gorm:"type:uuid;not null" json:"buyerId"`
	SellerID  uuid.UUID `gorm:"type:uuid;not null" json:"sellerId"`
	RecipeID  uuid.UUID `gorm:"type:uuid;not null" json:"recipeId"`
	Price     float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName задаёт имя таблицы
func (MarketPurchase) TableName() string {
	return "MarketPurchase"
}

// Achievement достижение ученика (badge)
type Achievement struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code        string    `gorm:"type:varchar(100);unique;not null" json:"code"` // knife_master, fusion_pro
	Title       string    `gorm:"type:varchar(255);not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	IconURL     string    `gorm:"type:text" json:"iconUrl"`
	Category    string    `gorm:"type:varchar(50)" json:"category"` // skill, course, recipe, special
	Requirement string    `gorm:"type:text" json:"requirement"`     // "Complete 5 courses", "Create 10 recipes"
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName задаёт имя таблицы
func (Achievement) TableName() string {
	return "Achievement"
}

// UserAchievement связь пользователя с достижением
type UserAchievement struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;not null" json:"userId"`
	AchievementID uuid.UUID `gorm:"type:uuid;not null" json:"achievementId"`
	UnlockedAt    time.Time `gorm:"autoCreateTime" json:"unlockedAt"`
}

// TableName задаёт имя таблицы
func (UserAchievement) TableName() string {
	return "UserAchievement"
}
