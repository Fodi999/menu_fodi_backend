package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Server
	HTTPPort string
	Env      string // dev, staging, prod

	// Database
	DatabaseURL string

	// JWT
	JWTSecret string

	// Groq AI
	GroqAPIKey string

	// SMTP (optional)
	SMTPHost string
	SMTPPort string
	SMTPUser string
	SMTPPass string

	// Admin
	AdminEmail string

	// Cloudinary (optional)
	CloudinaryURL string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file (ignore error if it doesn't exist - env vars might be set directly)
	_ = godotenv.Load()

	cfg := &Config{
		HTTPPort:      getEnv("HTTP_PORT", "8080"),
		Env:           getEnv("APP_ENV", "dev"),
		DatabaseURL:   getEnv("DATABASE_URL", ""),
		JWTSecret:     getEnv("JWT_SECRET", ""),
		GroqAPIKey:    getEnv("GROQ_API_KEY", ""),
		SMTPHost:      getEnv("SMTP_HOST", ""),
		SMTPPort:      getEnv("SMTP_PORT", "587"),
		SMTPUser:      getEnv("SMTP_USER", ""),
		SMTPPass:      getEnv("SMTP_PASS", ""),
		AdminEmail:    getEnv("ADMIN_EMAIL", "admin@chefacademy.com"),
		CloudinaryURL: getEnv("CLOUDINARY_URL", ""),
	}

	// Validate required fields
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an integer environment variable with a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// IsDev returns true if running in development mode
func (c *Config) IsDev() bool {
	return c.Env == "dev" || c.Env == "development"
}

// IsProd returns true if running in production mode
func (c *Config) IsProd() bool {
	return c.Env == "prod" || c.Env == "production"
}
