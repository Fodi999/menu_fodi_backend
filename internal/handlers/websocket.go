package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/auth"
	"github.com/gorilla/websocket"
)

const (
	// Время ожидания для записи сообщения
	writeWait = 10 * time.Second
	// Время ожидания чтения следующего pong сообщения от клиента
	pongWait = 60 * time.Second
	// Период отправки ping клиенту (должен быть меньше pongWait)
	pingPeriod = (pongWait * 9) / 10
	// Максимальный размер сообщения
	maxMessageSize = 512
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: true,
		HandshakeTimeout:  5 * time.Second,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")

			// Белый список разрешённых доменов
			allowed := []string{
				"http://localhost:3000",       // Локальная разработка Next.js
				"http://localhost:8080",       // Локальная разработка Backend
				"https://admin.fodifood.ru",   // Продакшен админка
				"https://menu.fodifood.ru",    // Продакшен меню
				"https://fodifood.vercel.app", // Vercel deployment
			}

			// В режиме разработки разрешаем все localhost
			if origin == "" {
				return true // WebSocket из того же домена
			}

			for _, o := range allowed {
				if origin == o {
					return true
				}
			}

			log.Printf("🚫 Blocked WebSocket origin: %s", origin)
			return false
		},
	}

	// Хранилище активных WebSocket клиентов
	clients     = make(map[*Client]bool)
	clientsLock sync.RWMutex
)

// InitWebSocketHub инициализирует и запускает WebSocket Hub
// Должна быть вызвана при старте сервера
func InitWebSocketHub() {
	log.Println("📡 Starting WebSocket Hub...")
	// Hub уже инициализирован через глобальные переменные
	// Можно добавить дополнительную логику инициализации если нужно
}

// Client представляет WebSocket клиента
type Client struct {
	Conn     *websocket.Conn
	Send     chan []byte
	LastSeen time.Time // Метка последней активности для мониторинга
	once     sync.Once // Гарантирует однократное отключение
}

// WebSocketMessage структура сообщения
type WebSocketMessage struct {
	Type string      `json:"type"` // "new_order", "order_updated", "order_cancelled"
	Data interface{} `json:"data"`
}

// disconnect безопасно отключает клиента (вызывается только один раз)
func (c *Client) disconnect() {
	c.once.Do(func() {
		// Удаляем из реестра клиентов
		clientsLock.Lock()
		delete(clients, c)
		totalClients := len(clients)
		clientsLock.Unlock()

		// Закрываем канал отправки (остановит writePump если еще работает)
		close(c.Send)

		// Закрываем WebSocket соединение
		c.Conn.Close()

		log.Printf("🔌 Client disconnected (was active: %v ago, remaining: %d)", 
			time.Since(c.LastSeen), totalClients)
	})
}

// validateAdminToken проверяет токен администратора с использованием JWT
func validateAdminToken(token string) bool {
	if token == "" {
		return false
	}

	// Валидируем JWT токен
	claims, err := auth.ValidateToken(token)
	if err != nil {
		log.Printf("🚫 Invalid JWT token: %v", err)
		return false
	}

	// Проверяем, что пользователь является администратором
	if claims.Role != "admin" {
		log.Printf("🚫 User %s is not an admin (role: %s)", claims.Email, claims.Role)
		return false
	}

	log.Printf("✅ Admin authenticated: %s (ID: %s)", claims.Email, claims.UserID)
	return true
}

// readPump читает сообщения от WebSocket клиента
func (c *Client) readPump() {
	defer c.disconnect()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		c.LastSeen = time.Now() // Обновляем метку при получении pong
		return nil
	})

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("❌ WebSocket error: %v", err)
			}
			break
		}
		// Обновляем метку активности при получении любого сообщения
		c.LastSeen = time.Now()
	}
}

// writePump отправляет сообщения WebSocket клиенту
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Канал закрыт
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				c.disconnect()
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("❌ Error sending message: %v", err)
				c.disconnect()
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.disconnect()
				return
			}
		}
	}
}

// HandleWebSocket обработчик WebSocket соединений
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Printf("[WS] 📞 New WebSocket connection attempt from: %s", r.RemoteAddr)
	log.Printf("[WS] 📍 Origin: %s", r.Header.Get("Origin"))

	// Проверяем авторизацию (только для админов)
	token := r.URL.Query().Get("token")
	if token == "" {
		authHeader := r.Header.Get("Authorization")
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}
	}

	if token == "" {
		log.Printf("[WS] ❌ No token provided")
		http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
		return
	}

	if !validateAdminToken(token) {
		log.Printf("[WS] ❌ Invalid admin token")
		log.Printf("[WS] ❌ Token rejected, cannot upgrade WS for %s", r.RemoteAddr)
		http.Error(w, "Unauthorized: Invalid admin credentials", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS] ❌ WebSocket upgrade error: %v", err)
		return
	}

	// Включаем компрессию для снижения трафика
	conn.EnableWriteCompression(true)

	// Отправляем приветственное сообщение сразу по соединению (до запуска горутин)
	welcomeMsg := WebSocketMessage{
		Type: "connected",
		Data: map[string]string{
			"message": "Connected to admin order notifications",
			"status":  "ready",
		},
	}
	conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err := conn.WriteJSON(welcomeMsg); err != nil {
		log.Printf("[WS] ⚠️ Could not send welcome message: %v", err)
		conn.Close()
		return
	}
	log.Printf("[WS] 📤 Welcome message sent to admin")

	client := &Client{
		Conn:     conn,
		Send:     make(chan []byte, 256),
		LastSeen: time.Now(), // Инициализируем метку активности
	}

	// Регистрируем нового клиента
	clientsLock.Lock()
	clients[client] = true
	totalClients := len(clients)
	clientsLock.Unlock()

	log.Printf("[WS] ✅ Admin connected: %s (total active: %d)", r.RemoteAddr, totalClients)

	// Запускаем горутины для чтения и записи
	go client.writePump()
	go client.readPump()
}

// BroadcastOrderNotification отправляет уведомление всем подключенным клиентам
func BroadcastOrderNotification(messageType string, data interface{}) {
	message := WebSocketMessage{
		Type: messageType,
		Data: data,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("❌ Error marshaling broadcast message: %v", err)
		return
	}

	var toRemove []*Client
	sentCount := 0

	// Читаем под RLock
	clientsLock.RLock()
	for client := range clients {
		select {
		case client.Send <- messageBytes:
			sentCount++
		default:
			// Канал переполнен, клиент не успевает обрабатывать сообщения
			log.Printf("⚠️ Client channel full, scheduling removal")
			toRemove = append(toRemove, client)
		}
	}
	clientsLock.RUnlock()

	// Безопасно удаляем проблемных клиентов под обычным Lock
	if len(toRemove) > 0 {
		for _, c := range toRemove {
			c.disconnect()
		}
		log.Printf("🗑️ Removed %d unresponsive clients", len(toRemove))
	}

	log.Printf("📡 Broadcast sent: type=%s to %d clients", messageType, sentCount)
}

// GetActiveConnectionsCount возвращает количество активных подключений
func GetActiveConnectionsCount() int {
	clientsLock.RLock()
	defer clientsLock.RUnlock()
	return len(clients)
}

// CloseAllConnections закрывает все WebSocket соединения (для graceful shutdown)
func CloseAllConnections() {
	clientsLock.RLock()
	clientsCopy := make([]*Client, 0, len(clients))
	for client := range clients {
		clientsCopy = append(clientsCopy, client)
	}
	totalClients := len(clientsCopy)
	clientsLock.RUnlock()

	log.Printf("🔌 Closing all WebSocket connections (%d clients)", totalClients)

	for _, client := range clientsCopy {
		// Отправляем сообщение о закрытии
		client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
		client.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, "Server shutdown"))
		
		// Используем централизованный disconnect
		client.disconnect()
	}

	log.Printf("✅ All WebSocket connections closed")
}
