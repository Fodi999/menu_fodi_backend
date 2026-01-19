package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/cron"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/config"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// App represents the application
type App struct {
	config      *config.Config
	db          *gorm.DB
	server      *http.Server
	cronChecker *cron.FridgeExpiryChecker
}

// New creates a new application instance
func New() (*App, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	if err := logger.Init(cfg.Env); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	logger.Info("🚀 Starting Chef Academy Backend", zap.String("env", cfg.Env))

	// Initialize database
	if err := database.Init(cfg.DatabaseURL); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	logger.Info("✅ Database connected successfully")

	// Initialize CRON jobs for fridge expiry checks
	cronChecker := cron.NewFridgeExpiryChecker(database.DB)
	cronChecker.Start()
	logger.Info("⏰ CRON jobs initialized - Daily fridge expiry checks at 08:00 UTC")

	app := &App{
		config:      cfg,
		db:          database.DB,
		cronChecker: cronChecker,
	}

	// Setup routes using modular DDD architecture
	logger.Info("📦 Using MODULAR routes (DDD architecture)")
	router := app.setupModularRoutes()

	// Create HTTP server
	app.server = &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return app, nil
}

// Run starts the application
func (a *App) Run() error {
	// Start server in a goroutine
	go func() {
		logger.Info("🌐 Server starting", zap.String("port", a.config.HTTPPort))
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("🛑 Shutting down server...")

	// Stop CRON jobs
	if a.cronChecker != nil {
		a.cronChecker.Stop()
		logger.Info("⏰ CRON jobs stopped")
	}

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
		return err
	}

	logger.Info("✅ Server stopped gracefully")
	logger.Sync()

	return nil
}
