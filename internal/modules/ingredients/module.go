package ingredients

import (
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ingredients/service"
	httphandlers "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ingredients/transport/http"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type Module struct {
	handlers *httphandlers.IngredientsHandlers
}

func NewModule(db *gorm.DB) *Module {
	ingredientsService := service.NewIngredientsService(db)
	handlers := httphandlers.NewIngredientsHandlers(ingredientsService)

	return &Module{
		handlers: handlers,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, jwtMiddleware func(next http.Handler) http.Handler) {
	// 📖 CATALOG ROUTES - Справочник продуктов (для ВСЕХ авторизованных)
	r.Route("/catalog/ingredients", func(r chi.Router) {
		r.Use(jwtMiddleware)

		r.Get("/", m.handlers.ListIngredients) // Список с фильтрами (category, search)
		r.Get("/search", m.handlers.Search)    // Автокомплит поиска
	})

	// 📦 STOCK ROUTES - Управление складом (ТОЛЬКО pro_chef)
	r.Route("/stock", func(r chi.Router) {
		r.Use(jwtMiddleware)
		r.Use(middleware.RequireRole(models.RoleProChef))

		r.Get("/", m.handlers.GetAll)                          // Складские остатки (StockItem)
		r.Post("/", m.handlers.Create)                         // Добавить на склад
		r.Get("/{id}", m.handlers.GetOne)                      // Детали позиции
		r.Put("/{id}", m.handlers.Update)                      // Обновить остатки
		r.Delete("/{id}", m.handlers.Delete)                   // Удалить со склада
		r.Get("/{id}/movements", m.handlers.GetStockMovements) // История движений
	})
}
