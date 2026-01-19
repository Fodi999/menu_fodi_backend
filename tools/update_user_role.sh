#!/bin/bash

# Обходной путь: используем SQL через миграцию на production
# Так как у нас нет прямого доступа к БД, создадим Go скрипт

cat > /tmp/update_role.go << 'GOEOF'
package main

import (
	"fmt"
	"log"
	"os"
	
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}
	
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	
	// Update user role
	result := db.Exec("UPDATE users SET role = 'admin' WHERE email = 'fodi85@gmail.ru'")
	if result.Error != nil {
		log.Fatalf("Update failed: %v", result.Error)
	}
	
	fmt.Printf("✅ Updated %d rows\n", result.RowsAffected)
	
	// Verify
	var user struct {
		Email string
		Role  string
	}
	db.Raw("SELECT email, role FROM users WHERE email = 'fodi85@gmail.ru'").Scan(&user)
	fmt.Printf("User: %s, Role: %s\n", user.Email, user.Role)
}
GOEOF

echo "Go script created at /tmp/update_role.go"
echo ""
echo "To run on production:"
echo "1. Copy this script to your server"
echo "2. Set DATABASE_URL environment variable"
echo "3. Run: go run update_role.go"
