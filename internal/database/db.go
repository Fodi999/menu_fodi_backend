package database

import (
	"log"
	"os"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

// Connect подключается к PostgreSQL базе данных
func Connect() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",
			SingularTable: false,
			NoLowerCase:   false, // ✅ пусть GORM сам сопоставляет имена
		},
	})

	if err != nil {
		return err
	}

	log.Println("✅ Connected to PostgreSQL database")

	// Автомиграция моделей (опционально, так как у нас уже есть Prisma схема)
	// err = DB.AutoMigrate(&models.User{})
	// if err != nil {
	// 	return err
	// }

	return nil
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
	)

	if err != nil {
		log.Printf("❌ Migration failed: %v", err)
		return err
	}

	log.Println("✅ Database schema migration completed successfully")
	return nil
}
