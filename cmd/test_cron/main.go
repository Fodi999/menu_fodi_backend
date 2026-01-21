package main

import (
	"fmt"
	"log"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/cron"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/config"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/logger"
	"go.uber.org/zap"
)

func main() {
	fmt.Println("🧪 CRON TEST UTILITY")
	fmt.Println("==================")
	fmt.Println("This will run the daily expiry check immediately")
	fmt.Println()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := logger.Init(cfg.Env); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	logger.Info("🔌 Connecting to database...")

	if err := database.Init(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	logger.Info("✅ Database connected", zap.String("database", "neon_postgresql"))

	checker := cron.NewFridgeExpiryChecker(database.DB)

	fmt.Println()
	logger.Info("🚀 Running expiry check NOW (manual trigger)...")
	checker.RunNow()

	fmt.Println()
	logger.Info("✅ Test completed successfully")
	logger.Info("💡 To see notifications, check: GET /api/notifications")
}
