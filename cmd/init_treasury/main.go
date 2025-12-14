package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using environment variables")
	}

	// Initialize database connection
	fmt.Println("🔌 Connecting to database...")
	dsn := os.Getenv("DATABASE_URL")
	if err := database.Init(dsn); err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}
	fmt.Println("✅ Database connected successfully!")

	// Check if Treasury already exists
	repo := &database.TokenBankRepository{}
	treasury, err := repo.GetTreasury()

	if err == nil && treasury != nil {
		fmt.Println("ℹ️  Treasury already exists:")
		fmt.Printf("   💰 Balance: %d\n", treasury.Balance)
		fmt.Printf("   � Total Allocated: %d\n", treasury.TotalAllocated)
		fmt.Printf("   📉 Total Used: %d\n", treasury.TotalUsed)

		// Ask if user wants to reset
		fmt.Print("\n❓ Treasury exists. Reset to initial values? (yes/no): ")
		var response string
		fmt.Scanln(&response)

		if response != "yes" && response != "y" {
			fmt.Println("⏭️  Skipping initialization")
			return
		}
	}

	// Initialize Treasury with starting balance
	initialBalance := int64(1000000) // 1 million tokens

	fmt.Println("\n🚀 Initializing Treasury...")
	fmt.Printf("   Initial balance: %d tokens\n", initialBalance)

	// Direct SQL insert/update
	query := `
		INSERT INTO token_bank (user_id, balance, total_allocated, total_used, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (user_id) DO UPDATE 
		SET 
			balance = $2,
			total_allocated = $3,
			total_used = $4,
			updated_at = NOW()
	`

	// Use special Treasury user_id (empty string or special UUID)
	treasuryUserID := "00000000-0000-0000-0000-000000000000"

	result := database.DB.Exec(query, treasuryUserID, initialBalance, int64(0), int64(0))
	if result.Error != nil {
		log.Fatal("❌ Failed to initialize Treasury:", result.Error)
	}

	fmt.Println("✅ Treasury initialized successfully!")

	// Verify initialization
	treasury, err = repo.GetTreasury()
	if err != nil {
		log.Fatal("❌ Failed to verify Treasury:", err)
	}

	fmt.Println("\n📊 Treasury Status:")
	fmt.Printf("   💰 Випущено (Balance): %d\n", treasury.Balance)
	fmt.Printf("   🔄 Всього виділено (Total Allocated): %d\n", treasury.TotalAllocated)
	fmt.Printf("   🔒 Використано (Total Used): %d\n", treasury.TotalUsed)
	fmt.Printf("   ✅ Доступно (Available): %d\n", treasury.Balance)
	fmt.Println("\n🎉 Treasury is ready to use!")
}
