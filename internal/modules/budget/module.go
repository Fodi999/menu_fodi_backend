package budget

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	httphandlers "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/budget/transport/http"
)

type Module struct {
	handler *httphandlers.BudgetHandler
}

func NewModule(db *gorm.DB) *Module {
	repo := database.NewWeeklyBudgetRepository(db)
	handler := httphandlers.NewBudgetHandler(repo)

	return &Module{handler: handler}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Route("/api/budget", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Get("/current", m.handler.GetCurrentWeekBudget) // GET /api/budget/current
		r.Get("/weekly", m.handler.GetWeeklyBudgets)      // GET /api/budget/weekly?weeks=4
		r.Get("/stats", m.handler.GetBudgetStats)         // GET /api/budget/stats
		r.Get("/week", m.handler.GetBudgetForWeek)        // GET /api/budget/week?date=2025-12-22
		r.Put("/plan", m.handler.SetPlannedBudget)        // PUT /api/budget/plan
	})
}
