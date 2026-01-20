package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Ingredient struct {
	ID       string  `gorm:"column:id"`
	Name     string  `gorm:"column:name"`
	NameRU   *string `gorm:"column:name_ru"`
	Category string  `gorm:"column:category"`
	Unit     string  `gorm:"column:unit"`
}

func (Ingredient) TableName() string {
	return "Ingredient"
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

	searchNames := []string{"Jajka", "Яйца", "Olive oil", "Масло", "Sól", "Соль", "Ryż", "Рис", "Łosoś", "Лосось"}

	fmt.Println("🔍 Checking ingredient categories:")
	fmt.Println("=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=")

	for _, searchName := range searchNames {
		var ingredient Ingredient
		result := db.Where("name LIKE ? OR name_pl LIKE ? OR name_en LIKE ? OR name_ru LIKE ?",
			"%"+searchName+"%", "%"+searchName+"%", "%"+searchName+"%", "%"+searchName+"%").
			First(&ingredient)

		if result.Error == nil {
			nameRU := "nil"
			if ingredient.NameRU != nil {
				nameRU = *ingredient.NameRU
			}
			fmt.Printf("\n✅ Found: %s\n", searchName)
			fmt.Printf("   Name: %s\n", ingredient.Name)
			fmt.Printf("   NameRU: %s\n", nameRU)
			fmt.Printf("   Category: %s\n", ingredient.Category)
			fmt.Printf("   Unit: %s\n", ingredient.Unit)
		} else {
			fmt.Printf("\n❌ Not found: %s\n", searchName)
		}
	}
}
