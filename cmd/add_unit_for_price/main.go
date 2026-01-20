package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using environment variables")
	}

	// Get DATABASE_URL
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("❌ DATABASE_URL is not set")
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	log.Println("✅ Connected to database")

	// Run migration
	log.Println("📝 Adding unit_for_price column...")
	
	if err := db.Exec(`
		ALTER TABLE user_fridge_price_history
		ADD COLUMN IF NOT EXISTS unit_for_price TEXT;
	`).Error; err != nil {
		log.Fatalf("❌ Failed to add column: %v", err)
	}

	log.Println("✅ Column unit_for_price added")

	// Add comment
	if err := db.Exec(`
		COMMENT ON COLUMN user_fridge_price_history.unit_for_price IS 'Unit of measurement for the price (kg, l, pcs, g, ml)';
	`).Error; err != nil {
		log.Printf("⚠️  Failed to add comment: %v", err)
	}

	// Update existing records with default values
	log.Println("📝 Updating existing records with default unit_for_price...")
	
	// Set 'kg' for small prices (likely normalized to grams)
	result := db.Exec(`
		UPDATE user_fridge_price_history
		SET unit_for_price = 'kg'
		WHERE unit_for_price IS NULL
		  AND price_per_unit < 1;
	`)
	if result.Error != nil {
		log.Printf("⚠️  Failed to update kg records: %v", result.Error)
	} else {
		log.Printf("✅ Updated %d records with unit_for_price = 'kg'", result.RowsAffected)
	}

	// Set 'pcs' for larger prices (likely per piece)
	result = db.Exec(`
		UPDATE user_fridge_price_history
		SET unit_for_price = 'pcs'
		WHERE unit_for_price IS NULL
		  AND price_per_unit >= 1;
	`)
	if result.Error != nil {
		log.Printf("⚠️  Failed to update pcs records: %v", result.Error)
	} else {
		log.Printf("✅ Updated %d records with unit_for_price = 'pcs'", result.RowsAffected)
	}

	// Check results
	var count int64
	db.Table("user_fridge_price_history").Where("unit_for_price IS NOT NULL").Count(&count)
	fmt.Printf("📊 Total price history records with unit_for_price: %d\n", count)

	log.Println("✨ Migration completed successfully!")
}
