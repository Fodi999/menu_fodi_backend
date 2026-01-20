package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Ingredient struct {
	ID     string  `gorm:"column:id"`
	Name   string  `gorm:"column:name"`
	NamePL *string `gorm:"column:name_pl"`
	NameEN *string `gorm:"column:name_en"`
	NameRU *string `gorm:"column:name_ru"`
	Unit   string  `gorm:"column:unit"`
}

func (Ingredient) TableName() string {
	return "Ingredient"
}

func normalizeSearchQuery(s string) string {
	s = strings.ToLower(s)
	replacements := map[rune]rune{
		'ą': 'a', 'ć': 'c', 'ę': 'e', 'ł': 'l',
		'ń': 'n', 'ó': 'o', 'ś': 's', 'ź': 'z', 'ż': 'z',
	}
	result := []rune(s)
	for i, r := range result {
		if replacement, ok := replacements[r]; ok {
			result[i] = replacement
		}
	}
	return string(result)
}

func main() {
	// Load .env
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

	queries := []string{"масло", "мол", "сыр"}

	for _, query := range queries {
		fmt.Printf("\n🔍 Testing search: '%s'\n", query)
		fmt.Println(strings.Repeat("=", 60))

		normalizedQuery := normalizeSearchQuery(query) + "%"

		// Search with LIMIT 20 (current API behavior)
		var ingredients []Ingredient
		result := db.
			Where("LOWER(name) LIKE ? OR LOWER(name_pl) LIKE ? OR LOWER(name_en) LIKE ? OR LOWER(name_ru) LIKE ?",
				normalizedQuery, normalizedQuery, normalizedQuery, normalizedQuery).
			Order("name ASC").
			Limit(20).
			Find(&ingredients)

		if result.Error != nil {
			log.Fatal(result.Error)
		}

		fmt.Printf("✅ API returns (LIMIT 20): %d results\n", len(ingredients))
		for i, ing := range ingredients {
			nameRU := "nil"
			if ing.NameRU != nil {
				nameRU = *ing.NameRU
			}
			fmt.Printf("   %2d. %-30s (ru: %-20s unit: %s)\n", i+1, ing.Name, nameRU, ing.Unit)
		}

		// Count total without LIMIT
		var count int64
		db.Model(&Ingredient{}).
			Where("LOWER(name) LIKE ? OR LOWER(name_pl) LIKE ? OR LOWER(name_en) LIKE ? OR LOWER(name_ru) LIKE ?",
				normalizedQuery, normalizedQuery, normalizedQuery, normalizedQuery).
			Count(&count)

		fmt.Printf("\n📊 Total in database: %d\n", count)

		if count > 20 {
			fmt.Printf("⚠️  WARNING: Missing %d results! (API limit: 20)\n", count-20)
		}
	}
}
