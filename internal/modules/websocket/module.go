package websocket

import (
	"log"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/websocket/transport/ws"
	"github.com/go-chi/chi/v5"
)

// Module содержит WebSocket функциональность
type Module struct {
	hub     *ws.Hub
	handler *ws.WSHandler
}

// NewModule создаёт новый WebSocket модуль
func NewModule() *Module {
	// Создаём Hub
	hub := ws.NewHub()

	// Запускаем Hub в горутине
	go hub.Run()

	// Создаём handler
	handler := ws.NewWSHandler(hub)

	log.Println("📡 WebSocket module initialized")

	return &Module{
		hub:     hub,
		handler: handler,
	}
}

// RegisterRoutes регистрирует WebSocket endpoints
func (m *Module) RegisterRoutes(r chi.Router) {
	log.Println("🔌 Registering WebSocket routes...")

	// Основной WebSocket endpoint (для всех событий)
	r.Get("/ws", m.handler.HandleWebSocket)

	// Специализированные endpoints
	r.Get("/ws/treasury", m.handler.HandleTreasuryWebSocket)          // Admin Treasury updates
	r.Get("/ws/tokens/{userID}", m.handler.HandleUserTokensWebSocket) // User tokens updates

	// HTTP endpoint для статистики
	r.Get("/ws/stats", m.handler.HandleStats)

	log.Println("✅ WebSocket routes registered:")
	log.Println("   GET /ws - General WebSocket connection")
	log.Println("   GET /ws/treasury - Treasury updates (admin)")
	log.Println("   GET /ws/tokens/{userID} - User token updates")
	log.Println("   GET /ws/stats - WebSocket statistics")
}

// GetHub возвращает Hub для использования в других модулях
func (m *Module) GetHub() *ws.Hub {
	return m.hub
}
