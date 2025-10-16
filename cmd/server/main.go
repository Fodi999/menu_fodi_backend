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

	// Автоматическая миграция схемы базы данных
	if err := database.AutoMigrate(); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Инициализация WebSocket Hub для real-time уведомлений
	handlers.InitWebSocketHub()
	log.Println("✅ WebSocket Hub initialized")

	// Инициализация роутера
	router := mux.NewRouter()

	// Middleware
	router.Use(middleware.Logger)

	// Root health check для Koyeb и других платформ деплоя
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}).Methods("GET")

	// API Routes
	api := router.PathPrefix("/api").Subrouter()

	// Health check endpoint (публичный)
	api.HandleFunc("/health", handlers.HealthCheck).Methods("GET", "OPTIONS")

	// Hint endpoint для Elixir бота (с API key защитой)
	hintRouter := api.PathPrefix("/hint").Subrouter()
	hintRouter.Use(middleware.APIKeyMiddleware)
	hintRouter.HandleFunc("", handlers.HintHandler).Methods("POST", "OPTIONS")

	// Auth routes (публичные)
	api.HandleFunc("/auth/register", handlers.Register).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/login", handlers.Login).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/verify", handlers.VerifyTokenHandler).Methods("POST", "OPTIONS")

	// Protected routes (требуют JWT)
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.AuthMiddleware)

	// User routes
	protected.HandleFunc("/user/profile", handlers.GetProfile).Methods("GET", "OPTIONS")
	protected.HandleFunc("/user/profile", handlers.UpdateProfile).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/user/orders", handlers.GetUserOrders).Methods("GET", "OPTIONS")

	// Orders (публичный endpoint для создания заказа)
	api.HandleFunc("/orders", handlers.CreateOrder).Methods("POST", "OPTIONS")

	// Admin routes
	admin := protected.PathPrefix("/admin").Subrouter()
	admin.Use(middleware.AdminMiddleware)

	// Users
	admin.HandleFunc("/users", handlers.GetAllUsers).Methods("GET", "OPTIONS")
	admin.HandleFunc("/users/{id}", handlers.UpdateUser).Methods("PUT", "OPTIONS")
	admin.HandleFunc("/users/{id}", handlers.DeleteUser).Methods("DELETE", "OPTIONS")
	admin.HandleFunc("/users/update-role", handlers.UpdateUserRole).Methods("PATCH", "OPTIONS")

	// Orders
	admin.HandleFunc("/orders", handlers.GetAllOrders).Methods("GET", "OPTIONS")
	admin.HandleFunc("/orders/recent", handlers.GetRecentOrders).Methods("GET", "OPTIONS")
	admin.HandleFunc("/orders/{id}/status", handlers.UpdateOrderStatus).Methods("PUT", "OPTIONS")

	// Stats
	admin.HandleFunc("/stats", handlers.GetAdminStats).Methods("GET", "OPTIONS")

	// Ingredients
	admin.HandleFunc("/ingredients", handlers.GetAllIngredients).Methods("GET", "OPTIONS")
	admin.HandleFunc("/ingredients", handlers.CreateIngredient).Methods("POST", "OPTIONS")
	admin.HandleFunc("/ingredients/{id}", handlers.UpdateIngredient).Methods("PUT", "OPTIONS")
	admin.HandleFunc("/ingredients/{id}", handlers.DeleteIngredient).Methods("DELETE", "OPTIONS")
	admin.HandleFunc("/ingredients/{id}/movements", handlers.GetStockMovements).Methods("GET", "OPTIONS")

	// Semi-Finished Products (Полуфабрикаты)
	admin.HandleFunc("/semi-finished", handlers.GetSemiFinished).Methods("GET", "OPTIONS")
	admin.HandleFunc("/semi-finished", handlers.CreateSemiFinished).Methods("POST", "OPTIONS")
	admin.HandleFunc("/semi-finished/{id}", handlers.GetSemiFinishedByID).Methods("GET", "OPTIONS")
	admin.HandleFunc("/semi-finished/{id}", handlers.UpdateSemiFinished).Methods("PUT", "OPTIONS")
	admin.HandleFunc("/semi-finished/{id}", handlers.DeleteSemiFinished).Methods("DELETE", "OPTIONS")

	// Products
	admin.HandleFunc("/products", handlers.GetAllProducts).Methods("GET", "OPTIONS")
	admin.HandleFunc("/products", handlers.CreateProduct).Methods("POST", "OPTIONS")
	admin.HandleFunc("/products/{id}", handlers.GetProduct).Methods("GET", "OPTIONS")
	admin.HandleFunc("/products/{id}", handlers.UpdateProduct).Methods("PUT", "OPTIONS")
	admin.HandleFunc("/products/{id}", handlers.DeleteProduct).Methods("DELETE", "OPTIONS")

	// WebSocket для real-time уведомлений (вне всех middleware, проверка токена внутри хэндлера)
	router.HandleFunc("/api/admin/ws", handlers.HandleWebSocket)

	// Business routes (публичные)
	api.HandleFunc("/businesses", handlers.GetBusinesses).Methods("GET", "OPTIONS")
	api.HandleFunc("/businesses", handlers.CreateBusiness).Methods("POST", "OPTIONS") // ✅ Стандартный REST endpoint
	api.HandleFunc("/businesses/create", handlers.CreateBusiness).Methods("POST", "OPTIONS")
	api.HandleFunc("/businesses/{id}", handlers.GetBusinessByID).Methods("GET", "OPTIONS")
	api.HandleFunc("/businesses/{id}", handlers.UpdateBusiness).Methods("PUT", "OPTIONS")
	api.HandleFunc("/businesses/{id}", handlers.DeleteBusiness).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/businesses/{id}/permanent", handlers.PermanentDeleteBusiness).Methods("DELETE", "OPTIONS")

	// Business Token routes (публичные)
	api.HandleFunc("/businesses/{id}/tokens", handlers.GetBusinessTokens).Methods("GET", "OPTIONS")
	api.HandleFunc("/businesses/{id}/tokens", handlers.CreateBusinessToken).Methods("POST", "OPTIONS")
	api.HandleFunc("/businesses/{id}/tokens/mint", handlers.MintBusinessTokens).Methods("POST", "OPTIONS")
	api.HandleFunc("/businesses/{id}/tokens/burn", handlers.BurnBusinessTokens).Methods("POST", "OPTIONS")
	api.HandleFunc("/businesses/{id}/tokens/recalculate-price", handlers.RecalculateTokenPrice).Methods("POST", "OPTIONS")

	// Business Subscription routes (инвестиции)
	api.HandleFunc("/businesses/{id}/subscribe", handlers.SubscribeToBusiness).Methods("POST", "OPTIONS")
	api.HandleFunc("/businesses/{id}/unsubscribe", handlers.UnsubscribeFromBusiness).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/businesses/{id}/subscribers", handlers.GetBusinessSubscribers).Methods("GET", "OPTIONS")
	api.HandleFunc("/users/{id}/subscriptions", handlers.GetUserSubscriptions).Methods("GET", "OPTIONS")
	api.HandleFunc("/subscriptions/stats", handlers.GetSubscriptionStats).Methods("GET", "OPTIONS")

	// Transaction routes (аналитика и история)
	api.HandleFunc("/businesses/{id}/transactions", handlers.GetBusinessTransactions).Methods("GET", "OPTIONS")
	api.HandleFunc("/users/{id}/transactions", handlers.GetUserTransactions).Methods("GET", "OPTIONS")
	api.HandleFunc("/transactions/analytics", handlers.GetTransactionAnalytics).Methods("GET", "OPTIONS")

	// Metrics routes (AI-метрики)
	api.HandleFunc("/metrics/{businessId}", handlers.GetBusinessMetrics).Methods("GET", "OPTIONS")

	// CORS настройки
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001", "https://menu-fodifood.vercel.app", "http://localhost:4000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-API-Key", "X-User-ID"},
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
