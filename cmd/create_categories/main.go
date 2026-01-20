package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type IngredientCategory struct {
	Key       string `gorm:"primaryKey;column:key"`
	Icon      string `gorm:"column:icon;not null"`
	SortOrder int    `gorm:"column:sort_order;not null"`
	LabelPL   string `gorm:"column:label_pl;not null"`
	LabelEN   string `gorm:"column:label_en;not null"`
	LabelRU   string `gorm:"column:label_ru;not null"`
}

func (IngredientCategory) TableName() string {
	return "ingredient_categories"
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Create table
	fmt.Println("📦 Creating ingredient_categories table...")
	if err := db.AutoMigrate(&IngredientCategory{}); err != nil {
		log.Fatal(err)
	}

	// Insert categories
	categories := []IngredientCategory{
		{Key: "all", Icon: "🧊", SortOrder: 0, LabelPL: "Wszystko", LabelEN: "All", LabelRU: "Все"},
		{Key: "fish", Icon: "🐟", SortOrder: 1, LabelPL: "Ryby", LabelEN: "Fish", LabelRU: "Рыба"},
		{Key: "meat", Icon: "🥩", SortOrder: 2, LabelPL: "Mięso", LabelEN: "Meat", LabelRU: "Мясо"},
		{Key: "egg", Icon: "🥚", SortOrder: 3, LabelPL: "Jajka", LabelEN: "Eggs", LabelRU: "Яйца"},
		{Key: "dairy", Icon: "🥛", SortOrder: 4, LabelPL: "Nabiał", LabelEN: "Dairy", LabelRU: "Молочные"},
		{Key: "vegetable", Icon: "🥕", SortOrder: 5, LabelPL: "Warzywa", LabelEN: "Vegetables", LabelRU: "Овощи"},
		{Key: "fruit", Icon: "🍎", SortOrder: 6, LabelPL: "Owoce", LabelEN: "Fruits", LabelRU: "Фрукты"},
		{Key: "grain", Icon: "🌾", SortOrder: 7, LabelPL: "Zboża", LabelEN: "Grains", LabelRU: "Крупы"},
		{Key: "condiment", Icon: "🧂", SortOrder: 8, LabelPL: "Przyprawy", LabelEN: "Condiments", LabelRU: "Приправы"},
		{Key: "other", Icon: "📦", SortOrder: 9, LabelPL: "Inne", LabelEN: "Other", LabelRU: "Другое"},
	}

	fmt.Println("📝 Inserting category data...")
	for _, cat := range categories {
		result := db.FirstOrCreate(&cat, IngredientCategory{Key: cat.Key})
		if result.Error != nil {
			log.Printf("Error inserting %s: %v", cat.Key, result.Error)
		} else {
			fmt.Printf("✅ %s - %s (%s)\n", cat.Icon, cat.LabelEN, cat.Key)
		}
	}

	// Verify
	var count int64
	db.Model(&IngredientCategory{}).Count(&count)
	fmt.Printf("\n📊 Total categories in database: %d\n", count)

	// Show all
	var all []IngredientCategory
	db.Order("sort_order ASC").Find(&all)
	fmt.Println("\n📋 Categories (sorted):")
	for _, cat := range all {
		fmt.Printf("   %d. %s %s (key: %s) - PL:%s EN:%s RU:%s\n",
			cat.SortOrder, cat.Icon, cat.LabelEN, cat.Key,
			cat.LabelPL, cat.LabelEN, cat.LabelRU)
	}
}
