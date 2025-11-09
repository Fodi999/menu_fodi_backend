package leaderboard

import (
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/leaderboard/service"
	httphandlers "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/leaderboard/transport/http"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type Module struct {
	handlers *httphandlers.LeaderboardHandlers
}

func NewModule(db *gorm.DB) *Module {
	lbService := service.NewLeaderboardService(db)
	handlers := httphandlers.NewLeaderboardHandlers(lbService)

	return &Module{
		handlers: handlers,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, _ func(next http.Handler) http.Handler) {
	r.Get("/leaderboard", m.handlers.GetLeaderboard)
}
