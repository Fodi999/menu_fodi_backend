package database

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
