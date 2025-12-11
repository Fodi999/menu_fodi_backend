package ws

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: В продакшене настроить CORS правильно
		return true
	},
}

// WSHandler обрабатывает WebSocket подключения
type WSHandler struct {
	hub *Hub
}

// NewWSHandler создаёт новый WSHandler
func NewWSHandler(hub *Hub) *WSHandler {
	return &WSHandler{
		hub: hub,
	}
}

// HandleWebSocket обрабатывает WebSocket подключение
func (h *WSHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Извлекаем userID из контекста (если есть)
	userID := ""
	if uid := r.Context().Value("userID"); uid != nil {
		userID = uid.(string)
	}
	
	// Upgrade HTTP соединения в WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	
	// Создаём клиента
	client := NewClient(h.hub, conn, userID)
	
	// Регистрируем клиента в Hub
	h.hub.register <- client
	
	// Запускаем горутины для чтения и записи
	go client.WritePump()
	go client.ReadPump()
	
	log.Printf("🔌 WebSocket client connected: %s (user: %s)", client.ID, userID)
}

// HandleTreasuryWebSocket - специальный endpoint для админов (Treasury updates)
func (h *WSHandler) HandleTreasuryWebSocket(w http.ResponseWriter, r *http.Request) {
	// Проверка admin роли (упрощённая, в реальности используйте middleware)
	// TODO: Добавить admin middleware
	
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	
	client := NewClient(h.hub, conn, "")
	
	// Автоматически подписываем на Treasury события
	client.subscribeTo("treasury_update")
	client.subscribeTo("treasury_allocate")
	client.subscribeTo("treasury_spend")
	
	h.hub.register <- client
	
	go client.WritePump()
	go client.ReadPump()
	
	log.Printf("🏦 Treasury WebSocket client connected: %s", client.ID)
}

// HandleUserTokensWebSocket - endpoint для отслеживания токенов конкретного пользователя
func (h *WSHandler) HandleUserTokensWebSocket(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		http.Error(w, "userID required", http.StatusBadRequest)
		return
	}
	
	// TODO: Проверить, что текущий пользователь = userID или это админ
	
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	
	client := NewClient(h.hub, conn, userID)
	
	// Автоматически подписываем на token события этого пользователя
	client.subscribeTo("token_balance_update")
	client.subscribeTo("token_earn")
	client.subscribeTo("token_spend")
	client.subscribeTo("task_completed")
	client.subscribeTo("task_reward_claimed")
	
	h.hub.register <- client
	
	go client.WritePump()
	go client.ReadPump()
	
	log.Printf("💰 User tokens WebSocket client connected: %s (user: %s)", client.ID, userID)
}

// HandleStats возвращает статистику WebSocket подключений (HTTP endpoint)
func (h *WSHandler) HandleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	stats := h.hub.GetStats()
	totalClients := stats["total_clients"].(int)
	authenticatedUsers := stats["authenticated_users"].(int)
	
	// Simple JSON response
	response := `{"status":"ok","stats":{"total_clients":` + string(rune(totalClients)) + `,"authenticated_users":` + string(rune(authenticatedUsers)) + `}}`
	w.Write([]byte(response))
}
