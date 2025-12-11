package backend
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Event struct {
	Type      string                 `json:"type"`
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

func main() {
	r := chi.NewRouter()

	// WebSocket endpoint
	r.Get("/ws", handleWebSocket)

	// Test endpoint to simulate token operations
	r.Post("/test/allocate", simulateAllocate)
	r.Post("/test/spend", simulateSpend)

	// Serve test HTML
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "websocket_test.html")
	})

	log.Println("🚀 Test WebSocket server started on :8080")
	log.Println("📡 WebSocket endpoint: ws://localhost:8080/ws")
	log.Println("🌐 Open http://localhost:8080 in your browser")
	
	http.ListenAndServe(":8080", r)
}

var clients = make(map[*websocket.Conn]bool)

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	clients[conn] = true
	log.Printf("✅ Client connected (total: %d)", len(clients))

	// Send welcome message
	welcome := Event{
		Type:      "connection",
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"status":  "connected",
			"message": "WebSocket connection established",
		},
	}
	conn.WriteJSON(welcome)

	// Read messages from client
	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			delete(clients, conn)
			log.Printf("❌ Client disconnected (total: %d)", len(clients))
			break
		}

		log.Printf("Received from client: %v", msg)

		// Handle subscribe command
		if action, ok := msg["action"].(string); ok && action == "subscribe" {
			response := Event{
				Type:      "subscription_confirmed",
				Timestamp: time.Now().Unix(),
				Data: map[string]interface{}{
					"events": msg["events"],
					"status": "subscribed",
				},
			}
			conn.WriteJSON(response)
		}
	}
}

func broadcast(event Event) {
	eventJSON, _ := json.Marshal(event)
	for conn := range clients {
		err := conn.WriteMessage(websocket.TextMessage, eventJSON)
		if err != nil {
			log.Printf("Error broadcasting: %v", err)
			conn.Close()
			delete(clients, conn)
		}
	}
}

func simulateAllocate(w http.ResponseWriter, r *http.Request) {
	// Simulate treasury allocation
	event := Event{
		Type:      "treasury_allocate",
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"balance":   999990000,
			"amount":    100,
			"user_id":   "user_123",
			"operation": "allocate",
		},
	}
	broadcast(event)

	// User token balance event
	userEvent := Event{
		Type:      "token_balance_update",
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"user_id":        "user_123",
			"balance_before": 0,
			"balance_after":  100,
			"amount":         100,
			"reason":         "allocated_from_treasury",
			"type":           "earn",
		},
	}
	broadcast(userEvent)

	log.Println("📤 Broadcasted allocation events")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "Allocation event broadcasted"})
}

func simulateSpend(w http.ResponseWriter, r *http.Request) {
	// Simulate token spending
	event := Event{
		Type:      "token_spend",
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"user_id":        "user_123",
			"balance_before": 100,
			"balance_after":  50,
			"amount":         50,
			"reason":         "tokens_spent",
			"type":           "spend",
		},
	}
	broadcast(event)

	// Treasury return event
	treasuryEvent := Event{
		Type:      "treasury_spend",
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"balance":   999990050,
			"amount":    50,
			"user_id":   "user_123",
			"operation": "return",
		},
	}
	broadcast(treasuryEvent)

	log.Println("📤 Broadcasted spending events")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "Spend event broadcasted"})
}
