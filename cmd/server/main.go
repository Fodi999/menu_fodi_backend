package main

import (
	"log"
	"net/http"
	"os"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/handlers"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	// Загрузка переменных окружения
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Подключение к базе данных
	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Инициализация роутера
	router := mux.NewRouter()

	// Middleware
	router.Use(middleware.Logger)

	// API Routes
	api := router.PathPrefix("/api").Subrouter()

	// Auth routes (публичные)
	api.HandleFunc("/auth/register", handlers.Register).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/login", handlers.Login).Methods("POST", "OPTIONS")

	// Protected routes (требуют JWT)
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.AuthMiddleware)

	// User routes
	protected.HandleFunc("/user/profile", handlers.GetProfile).Methods("GET", "OPTIONS")
	protected.HandleFunc("/user/profile", handlers.UpdateProfile).Methods("PUT", "OPTIONS")

	// Admin routes
	admin := protected.PathPrefix("/admin").Subrouter()
	admin.Use(middleware.AdminMiddleware)
	admin.HandleFunc("/users", handlers.GetAllUsers).Methods("GET", "OPTIONS")
	admin.HandleFunc("/users/{id}", handlers.UpdateUser).Methods("PUT", "OPTIONS")
	admin.HandleFunc("/users/{id}", handlers.DeleteUser).Methods("DELETE", "OPTIONS")
	admin.HandleFunc("/orders", handlers.GetAllOrders).Methods("GET", "OPTIONS")
	admin.HandleFunc("/orders/recent", handlers.GetRecentOrders).Methods("GET", "OPTIONS")
	admin.HandleFunc("/stats", handlers.GetAdminStats).Methods("GET", "OPTIONS")

	// CORS настройки
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://menu-fodifood.vercel.app"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	// Запуск сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
