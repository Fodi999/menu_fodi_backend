package menu

import (
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/menu/repository"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/menu/service"
	menuhttp "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/menu/transport/http"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// ============================================================================
// MODULE: Menu (Kitchen Pipeline)
// Purpose: Single source of truth for what user wants to cook TODAY
// ============================================================================

type MenuModule struct {
	menuHandler *menuhttp.MenuHandler
}

func NewMenuModule(db *gorm.DB) *MenuModule {
	// Layer 1: Repository (data access)
	menuRepo := repository.NewMenuRepository(db)
	
	// Layer 2: Service (business logic)
	menuService := service.NewMenuService(menuRepo)
	
	// Layer 3: Handler (HTTP transport)
	menuHandler := menuhttp.NewMenuHandler(menuService)
	
	return &MenuModule{
		menuHandler: menuHandler,
	}
}

// RegisterRoutes - регистрация маршрутов
func (m *MenuModule) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Route("/menu", func(r chi.Router) {
		// All menu endpoints require authentication
		r.Use(authMiddleware)
		
		// Today's menu
		r.Get("/today", m.menuHandler.GetTodayMenu)       // GET /api/menu/today - get today's menu
		r.Post("/today", m.menuHandler.AddToMenu)         // POST /api/menu/today - add recipe to menu
		
		// Menu item actions
		r.Post("/{id}/start", m.menuHandler.StartCooking)      // POST /api/menu/{id}/start - start cooking
		r.Post("/{id}/complete", m.menuHandler.CompleteCooking) // POST /api/menu/{id}/complete - complete cooking
		r.Post("/{id}/cancel", m.menuHandler.CancelMenuItem)    // POST /api/menu/{id}/cancel - cancel menu item
		r.Delete("/{id}", m.menuHandler.DeleteMenuItem)         // DELETE /api/menu/{id} - delete menu item
	})
}
