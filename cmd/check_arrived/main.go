package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type UserFridgeItem struct {
	ID        string `gorm:"column:id"`
	ArrivedAt string `gorm:"column:arrived_at"`
	ExpiresAt *string `gorm:"column:expires_at"`
}

func (UserFridgeItem) TableName() string {
	return "user_fridge_items"
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgresql://neondb_owner:npg_dz4Gl8ZhPLbX@ep-soft-mud-a-agon8wu3-pooler.eu-central-1.aws.neon.tech/neondb?sslmode=require"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}

	var items []UserFridgeItem
	result := db.Limit(5).Find(&items)
	if result.Error != nil {
		log.Fatal("Query failed:", result.Error)
	}

	fmt.Printf("Found %d items:\n", len(items))
	for _, item := range items {
		fmt.Printf("ID: %s, ArrivedAt: %s, ExpiresAt: %v\n", item.ID, item.ArrivedAt, item.ExpiresAt)
	}
}
