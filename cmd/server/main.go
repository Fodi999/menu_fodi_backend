package main

import (
	"log"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/app"
)

func main() {
	// Create and run application
	application, err := app.New()
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
