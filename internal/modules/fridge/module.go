package fridge

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/fridge/service"
	fridgehttp "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/fridge/transport/http"
)

// Module представляет модуль холодильника для HOME_CHEF пользователей
type Module struct {
	handlers *fridgehttp.FridgeHandlers
}

// NewModule создает новый модуль холодильника
func NewModule(db *gorm.DB) *Module {
	// Инициализируем репозитории
	fridgeRepo := database.NewUserFridgeRepository(db)
	ingredientRepo := &database.IngredientRepository{}

	// Инициализируем сервис
	svc := service.NewFridgeService(db, fridgeRepo, ingredientRepo)

	// Инициализируем handlers
	handlers := fridgehttp.NewFridgeHandlers(svc)

	return &Module{
		handlers: handlers,
	}
}

// RegisterRoutes регистрирует маршруты холодильника
func (m *Module) RegisterRoutes(r chi.Router, jwtMiddleware func(http.Handler) http.Handler) {
	r.Route("/fridge", func(r chi.Router) {
		// Требуется аутентификация + роль HOME_CHEF
		r.Use(jwtMiddleware)
		r.Use(middleware.RequireRole(models.RoleHomeChef))

		// Операции с продуктами
		r.Get("/items", m.handlers.GetUserItems)              // GET /api/fridge/items - список продуктов
		r.Post("/items", m.handlers.AddItem)                  // POST /api/fridge/items - добавить продукт
		r.Patch("/items/{id}", m.handlers.UpdateItemQuantity) // PATCH /api/fridge/items/{id} - обновить количество
		r.Delete("/items/{id}", m.handlers.DeleteItem)        // DELETE /api/fridge/items/{id} - удалить продукт

		// Операции с ценами (event sourcing)
		r.Post("/items/{id}/price", m.handlers.AddPrice)               // POST /api/fridge/items/{id}/price - добавить событие изменения цены
		r.Get("/items/{id}/price/history", m.handlers.GetPriceHistory) // GET /api/fridge/items/{id}/price/history - история изменения цен

		// Add missing ingredients from recipe
		r.Post("/add-missing", m.handlers.AddMissingIngredients) // POST /api/fridge/add-missing - добавить недостающие ингредиенты рецепта
	})
}
