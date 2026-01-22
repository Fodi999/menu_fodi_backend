package database

import (
	"log"
	"os"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

// Init initializes database connection
func Init(dsn string) error {
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",
			SingularTable: false,
			NoLowerCase:   false,
		},
		// Отключаем prepared statements для предотвращения конфликтов в connection pool
		PrepareStmt: false,
	})

	if err != nil {
		return err
	}

	// Настраиваем connection pool для стабильной работы с PostgreSQL
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	// Максимум открытых соединений (для Neon Serverless рекомендуется 10-20)
	sqlDB.SetMaxOpenConns(15)

	// Максимум idle соединений
	sqlDB.SetMaxIdleConns(5)

	// Максимальное время жизни соединения (предотвращает "stale connections")
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Максимальное время простоя соединения
	sqlDB.SetConnMaxIdleTime(2 * time.Minute)

	log.Println("✅ Database connection pool configured: maxOpen=15, maxIdle=5, maxLifetime=5m")

	return nil
}

// Connect подключается к PostgreSQL базе данных (deprecated: use Init)
func Connect() error {
	dsn := os.Getenv("DATABASE_URL")
	return Init(dsn)
}

// GetDB возвращает экземпляр базы данных
func GetDB() *gorm.DB {
	return DB
}

// AutoMigrate выполняет автоматическую миграцию схемы базы данных
func AutoMigrate() error {
	log.Println("🔄 Starting database schema migration...")

	// Выполняем миграцию для всех моделей
	err := DB.AutoMigrate(
		&models.User{},
		&models.Ingredient{},
		&models.SemiFinished{},
		&models.SemiFinishedIngredient{},
		&models.Product{},
		&models.ProductIngredient{},
		&models.ProductSemiFinished{},
		&models.Order{},
		&models.OrderItem{},
		&models.Business{},
		&models.BusinessToken{},
		&models.BusinessSubscription{},
		&models.Transaction{},
		// Culinary Academy & User Dashboard
		&models.UserProfile{},
		&models.PersonalRecipe{},
		&models.Certificate{},
		&models.UserProgress{},
		&models.WalletTransaction{},
		&models.MarketPurchase{},
		&models.Course{},
		&models.Lesson{},
		&models.QuizQuestion{},
		&models.UserQuiz{},
		&models.MentorSession{},
		&models.MentorMessage{},
		&models.Achievement{},
		&models.UserAchievement{},
		// Marketplace Evolution
		&models.RecipePurchase{},
		&models.RecipeReview{},
		// AI Chef Mentor (persistent sessions)
		&models.ChefMentorSession{},
		&models.ChefMentorMessage{},
		// AI Generated Recipes (Culinary OS)
		&models.AIGeneratedRecipe{},
		&models.RecipeLike{},
		// Recipe Social Feed
		&models.RecipePost{},
		&models.PostComment{},
		&models.PostLike{},
		// User Fridge (HOME_CHEF ingredients management)
		&models.UserFridgeItem{},
		// Token Bank (admin management)
		&models.TokenBank{},
		// Kitchen Pipeline (user menu)
		&models.UserMenuItem{},
	)

	if err != nil {
		log.Printf("❌ Migration failed: %v", err)
		return err
	}

	log.Println("✅ Database schema migration completed successfully")
	return nil
}
