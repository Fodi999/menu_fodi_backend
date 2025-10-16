package main

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("🔧 Dropping foreign key constraint...")
	
	// Drop foreign key constraint
	if err := db.Exec(`ALTER TABLE "Business" DROP CONSTRAINT IF EXISTS "fk_Business_owner"`).Error; err != nil {
		log.Printf("⚠️  Warning dropping FK: %v", err)
	} else {
		log.Println("✅ Foreign key constraint dropped")
	}

	// Make owner_id nullable
	if err := db.Exec(`ALTER TABLE "Business" ALTER COLUMN "owner_id" DROP NOT NULL`).Error; err != nil {
		log.Printf("⚠️  Warning making owner_id nullable: %v", err)
	} else {
		log.Println("✅ owner_id is now nullable")
	}

	log.Println("✅ Migration completed successfully!")
}
