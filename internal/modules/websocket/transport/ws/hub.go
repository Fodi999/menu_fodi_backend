package ws

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/websocket/service"
	"github.com/gorilla/websocket"
)

// Client представляет WebSocket клиента
type Client struct {
	ID     string
	UserID string
	Hub    *Hub
	Conn   *websocket.Conn
	Send   chan []byte
	
	// Subscription filters
	SubscribedEvents map[service.EventType]bool
	mu               sync.RWMutex
}

// Hub управляет WebSocket клиентами и broadcasts
type Hub struct {
	// Зарегистрированные клиенты
	clients map[*Client]bool
	
	// Клиенты по userID для таргетированных сообщений
	userClients map[string][]*Client
	
	// Канал для регистрации клиентов
	register chan *Client
	
	// Канал для отключения клиентов
	unregister chan *Client
	
	// Канал для broadcast сообщений
	broadcast chan []byte
	
	// EventBus интеграция
	eventBus *service.EventBus
	
	mu sync.RWMutex
}

// NewHub создаёт новый Hub
func NewHub() *Hub {
	hub := &Hub{
		clients:     make(map[*Client]bool),
		userClients: make(map[string][]*Client),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan []byte, 256),
		eventBus:    service.GetEventBus(),
	}
	
	// Подписываемся на все события EventBus
	hub.subscribeToEvents()
	
	return hub
}

// Run запускает Hub (должен быть запущен в горутине)
func (h *Hub) Run() {
	log.Println("🚀 WebSocket Hub started")
	
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)
			
		case client := <-h.unregister:
			h.unregisterClient(client)
			
		case message := <-h.broadcast:
			h.broadcastToAll(message)
		}
	}
}

// registerClient регистрирует нового клиента
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	h.clients[client] = true
	
	// Добавляем в map по userID
	if client.UserID != "" {
		h.userClients[client.UserID] = append(h.userClients[client.UserID], client)
		log.Printf("✅ Client registered: %s (user: %s, total: %d)", client.ID, client.UserID, len(h.clients))
	} else {
		log.Printf("✅ Anonymous client registered: %s (total: %d)", client.ID, len(h.clients))
	}
	
	// Отправляем welcome сообщение
	welcome := map[string]interface{}{
		"type":    "connection",
		"status":  "connected",
		"message": "WebSocket connection established",
		"client_id": client.ID,
	}
	welcomeJSON, _ := json.Marshal(welcome)
	client.Send <- welcomeJSON
}

// unregisterClient отключает клиента
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.Send)
		
		// Удаляем из userClients
		if client.UserID != "" {
			clients := h.userClients[client.UserID]
			for i, c := range clients {
				if c == client {
					h.userClients[client.UserID] = append(clients[:i], clients[i+1:]...)
					break
				}
			}
			if len(h.userClients[client.UserID]) == 0 {
				delete(h.userClients, client.UserID)
			}
		}
		
		log.Printf("❌ Client unregistered: %s (total: %d)", client.ID, len(h.clients))
	}
}

// broadcastToAll отправляет сообщение всем клиентам
func (h *Hub) broadcastToAll(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	for client := range h.clients {
		select {
		case client.Send <- message:
		default:
			// Канал заблокирован, пропускаем
			log.Printf("⚠️ Failed to send to client %s, channel blocked", client.ID)
		}
	}
}

// BroadcastToUser отправляет сообщение конкретному пользователю
func (h *Hub) BroadcastToUser(userID string, message []byte) {
	h.mu.RLock()
	clients := h.userClients[userID]
	h.mu.RUnlock()
	
	for _, client := range clients {
		select {
		case client.Send <- message:
		default:
			log.Printf("⚠️ Failed to send to user %s client %s", userID, client.ID)
		}
	}
}

// BroadcastEvent отправляет событие всем подписанным клиентам
func (h *Hub) BroadcastEvent(event service.Event) {
	eventJSON, err := event.ToJSON()
	if err != nil {
		log.Printf("❌ Failed to serialize event: %v", err)
		return
	}
	
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	// Если событие для конкретного пользователя, отправляем только ему
	if event.UserID != "" {
		clients := h.userClients[event.UserID]
		for _, client := range clients {
			if client.isSubscribedTo(event.Type) {
				select {
				case client.Send <- eventJSON:
				default:
					log.Printf("⚠️ Failed to send event %s to user %s", event.Type, event.UserID)
				}
			}
		}
		return
	}
	
	// Иначе broadcast всем подписанным клиентам
	for client := range h.clients {
		if client.isSubscribedTo(event.Type) {
			select {
			case client.Send <- eventJSON:
			default:
				log.Printf("⚠️ Failed to broadcast event %s to client %s", event.Type, client.ID)
			}
		}
	}
}

// subscribeToEvents подписывает Hub на все события EventBus
func (h *Hub) subscribeToEvents() {
	allEvents := []service.EventType{
		service.TreasuryUpdateEvent,
		service.TreasuryAllocateEvent,
		service.TreasurySpendEvent,
		service.TokenBalanceUpdateEvent,
		service.TokenEarnEvent,
		service.TokenSpendEvent,
		service.TaskCompletedEvent,
		service.TaskRewardClaimedEvent,
		service.UserRegisteredEvent,
		service.UserWelcomeBonusEvent,
	}
	
	for _, eventType := range allEvents {
		h.eventBus.Subscribe(eventType, func(event service.Event) {
			h.BroadcastEvent(event)
		})
	}
	
	log.Printf("📡 Hub subscribed to %d event types", len(allEvents))
}

// GetStats возвращает статистику Hub
func (h *Hub) GetStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	return map[string]interface{}{
		"total_clients":       len(h.clients),
		"authenticated_users": len(h.userClients),
	}
}

// ========================================
// Client methods
// ========================================

// NewClient создаёт нового WebSocket клиента
func NewClient(hub *Hub, conn *websocket.Conn, userID string) *Client {
	return &Client{
		ID:               generateClientID(),
		UserID:           userID,
		Hub:              hub,
		Conn:             conn,
		Send:             make(chan []byte, 256),
		SubscribedEvents: make(map[service.EventType]bool),
	}
}

// ReadPump читает сообщения от клиента
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()
	
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
		
		// Обработка команд от клиента
		c.handleMessage(message)
	}
}

// WritePump отправляет сообщения клиенту
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			
			if err := w.Close(); err != nil {
				return
			}
			
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage обрабатывает команды от клиента
func (c *Client) handleMessage(message []byte) {
	var cmd map[string]interface{}
	if err := json.Unmarshal(message, &cmd); err != nil {
		log.Printf("Invalid message from client %s: %v", c.ID, err)
		return
	}
	
	action, ok := cmd["action"].(string)
	if !ok {
		return
	}
	
	switch action {
	case "subscribe":
		// Подписка на события
		if events, ok := cmd["events"].([]interface{}); ok {
			for _, e := range events {
				if eventType, ok := e.(string); ok {
					c.subscribeTo(service.EventType(eventType))
				}
			}
		}
		
	case "unsubscribe":
		// Отписка от событий
		if events, ok := cmd["events"].([]interface{}); ok {
			for _, e := range events {
				if eventType, ok := e.(string); ok {
					c.unsubscribeFrom(service.EventType(eventType))
				}
			}
		}
		
	case "ping":
		// Ответ на ping
		pong := map[string]interface{}{
			"type": "pong",
			"timestamp": time.Now().Unix(),
		}
		pongJSON, _ := json.Marshal(pong)
		c.Send <- pongJSON
	}
}

// subscribeTo подписывает клиента на событие
func (c *Client) subscribeTo(eventType service.EventType) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.SubscribedEvents[eventType] = true
	log.Printf("📡 Client %s subscribed to %s", c.ID, eventType)
}

// unsubscribeFrom отписывает клиента от события
func (c *Client) unsubscribeFrom(eventType service.EventType) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.SubscribedEvents, eventType)
	log.Printf("📡 Client %s unsubscribed from %s", c.ID, eventType)
}

// isSubscribedTo проверяет подписку клиента на событие
func (c *Client) isSubscribedTo(eventType service.EventType) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	// Если не подписан ни на что, получает всё
	if len(c.SubscribedEvents) == 0 {
		return true
	}
	
	return c.SubscribedEvents[eventType]
}

// Helper functions

var clientIDCounter int64
var clientIDMutex sync.Mutex

func generateClientID() string {
	clientIDMutex.Lock()
	defer clientIDMutex.Unlock()
	clientIDCounter++
	return time.Now().Format("20060102150405") + "-" + string(rune(clientIDCounter))
}
