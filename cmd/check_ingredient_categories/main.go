package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Проверяем категории в Ingredient
	var ingredients []models.Ingredient
	if err := db.Select("id", "name", "name_pl", "category").
		Where("name_pl IN ?", []string{"Sól", "Łosoś", "Kasza gryczana", "Kefir", "Jaja", "Olej roślinny"}).
		Find(&ingredients).Error; err != nil {
		log.Fatal("Failed to query ingredients:", err)
	}

	fmt.Println("\n🔍 Ingredient Categories in Database:")
	fmt.Println("=====================================")
	for _, ing := range ingredients {
		namePL := ""
		if ing.NamePL != nil {
			namePL = *ing.NamePL
		}
		fmt.Printf("📦 %s (id: %s)\n", namePL, ing.ID)
		fmt.Printf("   category: '%s'\n\n", ing.Category)
	}

	// Считаем по категориям
	type CategoryCount struct {
		Category string
		Count    int64
	}
	var counts []CategoryCount
	db.Model(&models.Ingredient{}).
		Select("category, COUNT(*) as count").
		Group("category").
		Order("count DESC").
		Find(&counts)

	fmt.Println("\n📊 Category Distribution:")
	fmt.Println("=========================")
	for _, c := range counts {
		fmt.Printf("%s: %d ingredients\n", c.Category, c.Count)
	}
}
