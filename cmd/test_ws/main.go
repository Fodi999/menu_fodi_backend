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

var clients = make(map[*websocket.Conn]bool)

func main() {
	r := chi.NewRouter()

	// WebSocket endpoint
	r.Get("/ws", handleWebSocket)

	// Test endpoints
	r.Post("/api/test/allocate", simulateAllocate)
	r.Post("/api/test/spend", simulateSpend)
	r.Post("/api/test/task-complete", simulateTaskComplete)

	// Serve test HTML
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "websocket_test.html")
	})

	log.Println("🚀 WebSocket Test Server Started")
	log.Println("📡 WebSocket: ws://localhost:8080/ws")
	log.Println("🌐 Test Page: http://localhost:8080")
	log.Println()
	log.Println("Test Commands:")
	log.Println("  curl -X POST http://localhost:8080/api/test/allocate")
	log.Println("  curl -X POST http://localhost:8080/api/test/spend")
	log.Println("  curl -X POST http://localhost:8080/api/test/task-complete")
	
	http.ListenAndServe(":8080", r)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("❌ Upgrade error:", err)
		return
	}
	defer func() {
		conn.Close()
		delete(clients, conn)
	}()

	clients[conn] = true
	log.Printf("✅ Client connected (total: %d)", len(clients))

	// Welcome message
	welcome := Event{
		Type:      "connection",
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"status":  "connected",
			"message": "Welcome to Token Bank WebSocket",
		},
	}
	conn.WriteJSON(welcome)

	// Read loop
	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("❌ Client disconnected (total: %d)", len(clients)-1)
			break
		}

		if action, ok := msg["action"].(string); ok {
			switch action {
			case "subscribe":
				response := Event{
					Type:      "subscription_confirmed",
					Timestamp: time.Now().Unix(),
					Data: map[string]interface{}{
						"events":  msg["events"],
						"status":  "subscribed",
						"message": "You will receive these events",
					},
				}
				conn.WriteJSON(response)
				log.Printf("📡 Client subscribed to: %v", msg["events"])
			case "ping":
				pong := Event{
					Type:      "pong",
					Timestamp: time.Now().Unix(),
					Data:      map[string]interface{}{"status": "alive"},
				}
				conn.WriteJSON(pong)
			}
		}
	}
}

func broadcast(event Event) {
	log.Printf("📤 Broadcasting event: %s to %d clients", event.Type, len(clients))
	
	eventJSON, _ := json.Marshal(event)
	for conn := range clients {
		err := conn.WriteMessage(websocket.TextMessage, eventJSON)
		if err != nil {
			log.Printf("⚠️ Error broadcasting to client: %v", err)
			conn.Close()
			delete(clients, conn)
		}
	}
}

func simulateAllocate(w http.ResponseWriter, r *http.Request) {
	log.Println("💰 Simulating token allocation...")

	// Treasury event
	treasuryEvent := Event{
		Type:      "treasury_allocate",
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"balance":   999990000,
			"amount":    100,
			"user_id":   "user_123",
			"operation": "allocate",
			"remaining": 999990000,
		},
	}
	broadcast(treasuryEvent)

	// User balance event
	time.Sleep(100 * time.Millisecond)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"message": "Allocation events broadcasted",
		"clients": len(clients),
	})
}

func simulateSpend(w http.ResponseWriter, r *http.Request) {
	log.Println("💸 Simulating token spending...")

	// User spend event
	spendEvent := Event{
		Type:      "token_spend",
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"user_id":        "user_123",
			"balance_before": 100,
			"balance_after":  50,
			"amount":         50,
			"reason":         "ai_request",
			"type":           "spend",
		},
	}
	broadcast(spendEvent)

	// Treasury return event
	time.Sleep(100 * time.Millisecond)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"message": "Spend events broadcasted",
		"clients": len(clients),
	})
}

func simulateTaskComplete(w http.ResponseWriter, r *http.Request) {
	log.Println("🎯 Simulating task completion...")

	// Task completed event
	taskEvent := Event{
		Type:      "task_completed",
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"user_id":   "user_123",
			"task_id":   "task_456",
			"task_name": "Complete first recipe",
			"reward":    50,
			"status":    "completed",
		},
	}
	broadcast(taskEvent)

	// Reward claimed event
	time.Sleep(100 * time.Millisecond)
	rewardEvent := Event{
		Type:      "task_reward_claimed",
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"user_id":        "user_123",
			"task_id":        "task_456",
			"reward":         50,
			"balance_before": 50,
			"balance_after":  100,
		},
	}
	broadcast(rewardEvent)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"message": "Task events broadcasted",
		"clients": len(clients),
	})
}
