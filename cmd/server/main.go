package main

import (
	"log"
	"runtime/debug"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/app"
)

func main() {
	// 🔥 GLOBAL PANIC RECOVERY - CRITICAL FOR DEBUGGING
	defer func() {
		if r := recover(); r != nil {
			log.Printf("🚨 GLOBAL PANIC RECOVERED: %v\n", r)
			log.Printf("📍 Stack trace:\n%s\n", debug.Stack())
			log.Fatal("Application crashed")
		}
	}()

	// Create and run application
	application, err := app.New()
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
