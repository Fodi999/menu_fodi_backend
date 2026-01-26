package admin

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/admin/service"
	httphandlers "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/admin/transport/http"
)

type Module struct {
	handlers *httphandlers.AdminHandlers
}

func NewModule() *Module {
	// Создаём экземпляры сервиса и политики
	adminService := service.NewAdminService()
	adminPolicy := service.NewAdminPolicy()

	// Инжектируем зависимости в handlers
	handlers := httphandlers.NewAdminHandlers(adminService, adminPolicy)

	return &Module{handlers: handlers}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler, adminMiddleware func(http.Handler) http.Handler, superAdminMiddleware func(http.Handler) http.Handler) {
	// PUBLIC ENDPOINTS - NO AUTH REQUIRED
	r.Route("/api/public", func(r chi.Router) {
		r.Get("/treasury", m.handlers.GetTreasuryInfo) // Public treasury info
	})

	// PUBLIC SSE ENDPOINT — БЕЗ АВТОРИЗАЦИИ (EventSource не может отправлять headers)
	r.Route("/treasury", func(r chi.Router) {
		r.Get("/stream", m.handlers.StreamTreasury) // SSE stream - публичный доступ
	})

	r.Route("/admin", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Use(adminMiddleware) // Требует admin или super_admin

		// Users (доступно admin + super_admin)
		r.Get("/users", m.handlers.GetAllUsers)
		r.Get("/users/stats", m.handlers.GetUsersStats)
		r.Get("/users/{id}", m.handlers.GetUserByID)     // ← ДОБАВЛЕНО: получить пользователя по ID
		r.Put("/users/{id}", m.handlers.UpdateUser)

		// CRITICAL: Super admin only operations
		r.With(superAdminMiddleware).Delete("/users/{id}", m.handlers.DeleteUser)            // Удаление пользователя
		r.With(superAdminMiddleware).Patch("/users/{id}/role", m.handlers.UpdateUserRole)    // Изменение ролей (RESTful)
		r.With(superAdminMiddleware).Patch("/users/update-role", m.handlers.UpdateUserRole)  // Legacy endpoint (deprecated)

		// Orders
		r.Get("/orders", m.handlers.GetAllOrders)
		r.Get("/orders/recent", m.handlers.GetRecentOrders)
		r.Put("/orders/{id}/status", m.handlers.UpdateOrderStatus)

		// Stats
		r.Get("/stats", m.handlers.GetAdminStats)

		// Dashboard
		r.Get("/dashboard", m.handlers.GetAdminDashboard)

		// Admin Profile
		r.Get("/profile", m.handlers.GetAdminProfile)

		// Token Bank
		r.Get("/token-bank", m.handlers.GetAllTokenBanks)
		r.Get("/token-bank/stats", m.handlers.GetTokenBankStats)
		r.Get("/token-bank/{userID}", m.handlers.GetUserTokenBank)
		r.Post("/token-bank/allocate", m.handlers.AllocateTokens)
		r.Post("/token-bank/revoke", m.handlers.RevokeTokens)
		r.Put("/token-bank/balance", m.handlers.SetTokenBalance)

		// Token Transactions History
		r.Get("/token-bank/transactions", m.handlers.GetAllTransactions)           // All transactions
		r.Get("/token-bank/transactions/{userID}", m.handlers.GetUserTransactions) // User-specific
		r.Get("/token-bank/transactions/filter", m.handlers.GetTransactionsByType) // With filters
		r.Get("/token-bank/transactions/stats", m.handlers.GetTransactionStats)    // Statistics

		// Treasury
		r.Get("/treasury", m.handlers.GetTreasuryInfo)
		r.Get("/treasury/stats", m.handlers.GetTreasuryStats)        // Detailed Treasury statistics
		r.Get("/token-bank/treasury", m.handlers.GetTreasuryBalance) // Simplified endpoint
		r.Post("/treasury/allocate", m.handlers.AllocateFromTreasury)

		// Ingredient Catalog Management
		r.Get("/ingredients", m.handlers.GetAllIngredients)          // GET /api/admin/ingredients - весь каталог
		r.Get("/ingredients/stats", m.handlers.GetIngredientsStats)  // GET /api/admin/ingredients/stats - статистика
		r.Get("/ingredients/suggest", m.handlers.SuggestIngredients) // GET /api/admin/ingredients/suggest?q=абр - autocomplete (fast, no AI)
		r.Post("/ingredients", m.handlers.CreateIngredient)          // POST /api/admin/ingredients - создать ингредиент (AI classification)
		r.Post("/ingredients/hint", m.handlers.IngredientHint)       // POST /api/admin/ingredients/hint - AI подсказка при конфликте
		r.Post("/ingredients/import", m.handlers.ImportIngredients)  // Bulk import catalog
		r.Delete("/ingredients/{id}", m.handlers.DeleteIngredient)   // DELETE /api/admin/ingredients/{id} - удалить ингредиент

		// Recipe Catalog Management
		r.Get("/recipes", m.handlers.GetAllRecipes)                        // GET /api/admin/recipes - каталог с фильтрацией
		r.Get("/recipes/filters/meta", m.handlers.GetRecipeFilterMetadata) // GET /api/admin/recipes/filters/meta - метаданные фильтров
		r.Get("/recipes/stats", m.handlers.GetRecipesStats)                // GET /api/admin/recipes/stats - статистика

		// AI Recipe Creation
		r.Post("/recipes/create-ai", m.handlers.CreateRecipeWithAI)   // POST /api/admin/recipes/create-ai - создать рецепт через AI
		r.Post("/recipes/preview-ai", m.handlers.PreviewRecipeWithAI) // POST /api/admin/recipes/preview-ai - preview без сохранения
		r.Post("/recipes/save", m.handlers.SaveEditedRecipe)          // POST /api/admin/recipes/save - сохранить отредактированный рецепт
		r.Put("/recipes/{id}", m.handlers.UpdateRecipe)               // PUT /api/admin/recipes/{id} - обновить существующий рецепт
		r.Delete("/recipes/{id}", m.handlers.DeleteRecipe)            // DELETE /api/admin/recipes/{id} - удалить рецепт

		// Recipe Image Management (Cloudinary)
		r.Post("/recipes/{id}/image", m.handlers.UploadRecipeImage)   // POST /api/admin/recipes/{id}/image - загрузить изображение
		r.Delete("/recipes/{id}/image", m.handlers.DeleteRecipeImage) // DELETE /api/admin/recipes/{id}/image - удалить изображение
	})
}
