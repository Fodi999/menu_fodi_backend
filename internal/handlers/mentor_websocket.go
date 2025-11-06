package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/ai"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// MentorClient представляет WebSocket клиента для AI Mentor чата
type MentorClient struct {
	Conn      *websocket.Conn
	Send      chan []byte
	UserID    string
	SessionID string
	Language  string
	LastSeen  time.Time
	once      sync.Once
}

// MentorChatMessage структура сообщения в AI Mentor чате
type MentorChatMessage struct {
	Type      string                 `json:"type"` // "user_message", "ai_response", "session_start", "error"
	Content   string                 `json:"content"`
	SessionID string                 `json:"sessionId"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

var (
	mentorClients     = make(map[*MentorClient]bool)
	mentorClientsLock sync.RWMutex

	mentorUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// Разрешаем подключения из любых источников для учебной платформы
			return true
		},
	}
)

// disconnect безопасно отключает mentor клиента
func (mc *MentorClient) disconnect() {
	mc.once.Do(func() {
		mentorClientsLock.Lock()
		delete(mentorClients, mc)
		totalClients := len(mentorClients)
		mentorClientsLock.Unlock()

		close(mc.Send)
		mc.Conn.Close()

		// Завершаем сессию в базе
		if mc.SessionID != "" {
			now := time.Now()
			database.DB.Model(&models.MentorSession{}).
				Where("id = ?", mc.SessionID).
				Update("ended_at", now)
		}

		log.Printf("[MENTOR] 🔌 Client disconnected (UserID: %s, Session: %s, remaining: %d)",
			mc.UserID, mc.SessionID, totalClients)
	})
}

// readPump читает сообщения от пользователя
func (mc *MentorClient) readPump() {
	defer mc.disconnect()

	mc.Conn.SetReadLimit(2048) // Увеличиваем лимит для текстовых сообщений
	mc.Conn.SetReadDeadline(time.Now().Add(120 * time.Second))
	mc.Conn.SetPongHandler(func(string) error {
		mc.Conn.SetReadDeadline(time.Now().Add(120 * time.Second))
		mc.LastSeen = time.Now()
		return nil
	})

	for {
		var msg MentorChatMessage
		err := mc.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[MENTOR] ❌ WebSocket error: %v", err)
			}
			break
		}

		mc.LastSeen = time.Now()

		// Обрабатываем сообщение пользователя
		if msg.Type == "user_message" {
			go mc.handleUserMessage(msg.Content)
		}
	}
}

// writePump отправляет сообщения клиенту
func (mc *MentorClient) writePump() {
	ticker := time.NewTicker(50 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-mc.Send:
			mc.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				mc.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				mc.disconnect()
				return
			}

			if err := mc.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("[MENTOR] ❌ Error sending message: %v", err)
				mc.disconnect()
				return
			}

		case <-ticker.C:
			mc.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := mc.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				mc.disconnect()
				return
			}
		}
	}
}

// handleUserMessage обрабатывает сообщение пользователя и генерирует AI ответ
func (mc *MentorClient) handleUserMessage(content string) {
	log.Printf("[MENTOR] 💬 User message (Session: %s): %s", mc.SessionID, content)

	// Сохраняем сообщение пользователя в БД
	userMsg := models.MentorMessage{
		SessionID: uuid.MustParse(mc.SessionID),
		Role:      "user",
		Content:   content,
	}
	database.DB.Create(&userMsg)

	// Обновляем счётчик сообщений
	database.DB.Model(&models.MentorSession{}).
		Where("id = ?", mc.SessionID).
		UpdateColumn("messages", database.DB.Raw("messages + 1"))

	// Вызываем AI
	mentor := ai.NewMentorChat(mc.Language)
	aiResponse, err := mentor.Ask(content)
	if err != nil {
		log.Printf("[MENTOR] ❌ AI error: %v", err)
		mc.sendErrorMessage("Failed to get AI response. Please try again.")
		return
	}

	// Сохраняем ответ AI в БД
	aiMsg := models.MentorMessage{
		SessionID: uuid.MustParse(mc.SessionID),
		Role:      "assistant",
		Content:   aiResponse,
	}
	database.DB.Create(&aiMsg)

	// Отправляем ответ клиенту
	response := MentorChatMessage{
		Type:      "ai_response",
		Content:   aiResponse,
		SessionID: mc.SessionID,
		Timestamp: time.Now(),
	}

	responseBytes, _ := json.Marshal(response)
	select {
	case mc.Send <- responseBytes:
		log.Printf("[MENTOR] ✅ AI response sent")
	default:
		log.Printf("[MENTOR] ⚠️ Client channel full")
	}
}

// sendErrorMessage отправляет сообщение об ошибке клиенту
func (mc *MentorClient) sendErrorMessage(errorMsg string) {
	response := MentorChatMessage{
		Type:      "error",
		Content:   errorMsg,
		SessionID: mc.SessionID,
		Timestamp: time.Now(),
	}
	responseBytes, _ := json.Marshal(response)
	select {
	case mc.Send <- responseBytes:
	default:
	}
}

// HandleMentorWebSocket обработчик WebSocket для AI Mentor Chat
func HandleMentorWebSocket(w http.ResponseWriter, r *http.Request) {
	// Получаем параметры из query string
	userID := r.URL.Query().Get("userId")
	language := r.URL.Query().Get("language")
	topic := r.URL.Query().Get("topic")

	if userID == "" {
		http.Error(w, "Missing userId parameter", http.StatusBadRequest)
		return
	}

	if language == "" {
		language = "pl" // default
	}

	log.Printf("[MENTOR] 📞 New mentor chat connection: UserID=%s, Lang=%s, Topic=%s", userID, language, topic)

	// Проверяем профиль пользователя
	var profile models.UserProfile
	if err := database.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		http.Error(w, "User profile not found", http.StatusNotFound)
		return
	}

	// Создаём новую сессию
	session := models.MentorSession{
		UserID:   uuid.MustParse(userID),
		Language: language,
		Topic:    topic,
	}
	if err := database.DB.Create(&session).Error; err != nil {
		http.Error(w, "Failed to create mentor session", http.StatusInternalServerError)
		return
	}

	// Upgrade WebSocket
	conn, err := mentorUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[MENTOR] ❌ WebSocket upgrade error: %v", err)
		return
	}

	client := &MentorClient{
		Conn:      conn,
		Send:      make(chan []byte, 256),
		UserID:    userID,
		SessionID: session.ID.String(),
		Language:  language,
		LastSeen:  time.Now(),
	}

	// Регистрируем клиента
	mentorClientsLock.Lock()
	mentorClients[client] = true
	totalClients := len(mentorClients)
	mentorClientsLock.Unlock()

	log.Printf("[MENTOR] ✅ Session started: %s (total active: %d)", session.ID, totalClients)

	// Отправляем приветственное сообщение
	welcomeMsg := MentorChatMessage{
		Type:      "session_start",
		Content:   client.getWelcomeMessage(profile.Name),
		SessionID: session.ID.String(),
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"studentName":  profile.Name,
			"studentLevel": profile.Level,
			"language":     language,
		},
	}
	welcomeBytes, _ := json.Marshal(welcomeMsg)
	client.Send <- welcomeBytes

	// Запускаем горутины
	go client.writePump()
	go client.readPump()
}

// getWelcomeMessage возвращает приветственное сообщение на нужном языке
func (mc *MentorClient) getWelcomeMessage(studentName string) string {
	messages := map[string]string{
		"pl": "Witaj, " + studentName + "! Jestem Twoim osobistym AI Mentor Kulinarny. W czym mogę Ci dzisiaj pomóc? Możesz zapytać o techniki gotowania, przepisy, składniki lub porady kulinarne.",

		"ua": "Вітаю, " + studentName + "! Я твій персональний AI Кулінарний Ментор. Чим можу тобі сьогодні допомогти? Можеш запитати про техніки приготування, рецепти, інгредієнти або кулінарні поради.",

		"en": "Welcome, " + studentName + "! I'm your personal AI Culinary Mentor. How can I help you today? You can ask about cooking techniques, recipes, ingredients, or culinary advice.",
	}

	if msg, ok := messages[mc.Language]; ok {
		return msg
	}
	return messages["pl"]
}

// GetMentorSessionHistory GET /api/mentor/history/{sessionId} - история сообщений сессии
func GetMentorSessionHistory(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "Missing sessionId parameter", http.StatusBadRequest)
		return
	}

	var messages []models.MentorMessage
	if err := database.DB.Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"data": map[string]interface{}{
			"sessionId": sessionID,
			"messages":  messages,
			"total":     len(messages),
		},
	})
}

// GetUserMentorSessions GET /api/user/{userId}/mentor/sessions - все сессии пользователя
func GetUserMentorSessions(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	if userID == "" {
		http.Error(w, "Missing userId parameter", http.StatusBadRequest)
		return
	}

	var sessions []models.MentorSession
	if err := database.DB.Where("user_id = ?", userID).
		Order("started_at DESC").
		Limit(20).
		Find(&sessions).Error; err != nil {
		http.Error(w, "Failed to fetch sessions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"data":   sessions,
	})
}
