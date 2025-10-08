package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Загрузка переменных окружения
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Получаем DATABASE_URL
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	// Подключаемся к базе данных
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("✅ Connected to database")

	// Выполняем SQL команды для переименования столбцов
	sqls := []string{
		`ALTER TABLE "OrderItem" RENAME COLUMN "orderId" TO "order_id"`,
		`ALTER TABLE "OrderItem" RENAME COLUMN "productId" TO "product_id"`,
	}

	for _, sql := range sqls {
		log.Printf("📝 Executing: %s", sql)
		if err := db.Exec(sql).Error; err != nil {
			// Игнорируем ошибки, если столбцы уже переименованы
			log.Printf("⚠️ Warning: %v (this is OK if columns are already renamed)", err)
		} else {
			log.Printf("✅ Success!")
		}
	}

	log.Println("🎉 Migration completed!")
}
