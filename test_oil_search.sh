#!/bin/bash

# Test oil/butter search in database
echo "=== Testing 'масло' search in database ==="
echo ""

# Run Go code to query database
go run <<'EOF'
package main

import (
	"fmt"
	"log"
	"os"
	
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/joho/godotenv"
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

func main() {
	// Load .env
	godotenv.Load()
	
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	
	var ingredients []Ingredient
	query := "масло"
	normalizedQuery := query + "%"
	
	// Search with LIMIT
	result := db.
		Where("LOWER(name) LIKE ? OR LOWER(name_pl) LIKE ? OR LOWER(name_en) LIKE ? OR LOWER(name_ru) LIKE ?",
			normalizedQuery, normalizedQuery, normalizedQuery, normalizedQuery).
		Order("name ASC").
		Limit(20).
		Find(&ingredients)
	
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	
	fmt.Printf("✅ Found %d ingredients with LIMIT 20:\n", len(ingredients))
	for i, ing := range ingredients {
		nameRU := "nil"
		if ing.NameRU != nil {
			nameRU = *ing.NameRU
		}
		fmt.Printf("%d. %s (ru: %s, unit: %s)\n", i+1, ing.Name, nameRU, ing.Unit)
	}
	
	// Count total without LIMIT
	var count int64
	db.Model(&Ingredient{}).
		Where("LOWER(name) LIKE ? OR LOWER(name_pl) LIKE ? OR LOWER(name_en) LIKE ? OR LOWER(name_ru) LIKE ?",
			normalizedQuery, normalizedQuery, normalizedQuery, normalizedQuery).
		Count(&count)
	
	fmt.Printf("\n📊 Total in database: %d\n", count)
	
	if count > 20 {
		fmt.Printf("⚠️  WARNING: API returns only 20 results, but DB has %d!\n", count)
	}
}
EOF
