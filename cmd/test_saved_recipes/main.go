package main

import (
	"log"
	"os"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load .env
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found")
	}

	// Get DATABASE_URL
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	// Connect to DB
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Enable SQL logging
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("✅ Connected to database")

	// Test user and recipe IDs
	userID := "407582be-59d5-4d21-873b-1a72d31b0d42"
	recipeID := "92691aae-c3af-427d-aaed-1408319f0a3c"

	// Check if user exists
	var user models.User
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		log.Printf("❌ User not found: %v", err)
	} else {
		log.Printf("✅ User found: %s", user.Email)
	}

	// Check if recipe exists
	var recipe models.RecipeCatalog
	if err := db.Where("id = ?", recipeID).First(&recipe).Error; err != nil {
		log.Printf("❌ Recipe not found: %v", err)
	} else {
		log.Printf("✅ Recipe found: %s", recipe.LocalName)
	}

	// Try to save recipe
	repo := database.NewUserSavedRecipeRepository(db)
	savedRecipe, err := repo.SaveRecipe(userID, recipeID, 2, "fridge")
	if err != nil {
		log.Printf("❌ Failed to save recipe: %v", err)
	} else {
		log.Printf("✅ Recipe saved successfully: %s", savedRecipe.ID)
	}

	// Try to fetch saved recipes
	savedRecipes, err := repo.GetSavedRecipes(userID)
	if err != nil {
		log.Printf("❌ Failed to get saved recipes: %v", err)
	} else {
		log.Printf("✅ Found %d saved recipes", len(savedRecipes))
		for _, sr := range savedRecipes {
			log.Printf("   - %s (servings: %d)", sr.RecipeID, sr.Servings)
		}
	}
}
