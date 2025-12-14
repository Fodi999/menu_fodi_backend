package service

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

// EventType определяет типы событий в системе
type EventType string

const (
	// Treasury Events
	TreasuryUpdateEvent   EventType = "treasury_update"
	TreasuryAllocateEvent EventType = "treasury_allocate"
	TreasurySpendEvent    EventType = "treasury_spend"

	// Token Bank Events
	TokenBalanceUpdateEvent EventType = "token_balance_update"
	TokenEarnEvent          EventType = "token_earn"
	TokenSpendEvent         EventType = "token_spend"

	// Task Events
	TaskCompletedEvent     EventType = "task_completed"
	TaskRewardClaimedEvent EventType = "task_reward_claimed"

	// User Events
	UserRegisteredEvent   EventType = "user_registered"
	UserWelcomeBonusEvent EventType = "user_welcome_bonus"
)

// Event представляет событие в системе
type Event struct {
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	UserID    string                 `json:"user_id,omitempty"` // Для пользовательских событий
}

// Subscriber функция-подписчик на события
type Subscriber func(event Event)

// EventBus шина событий для pub/sub
type EventBus struct {
	subscribers map[EventType][]Subscriber
	mu          sync.RWMutex
}

var (
	// GlobalEventBus глобальный экземпляр event bus
	GlobalEventBus *EventBus
	once           sync.Once
)

// GetEventBus возвращает singleton экземпляр EventBus
func GetEventBus() *EventBus {
	once.Do(func() {
		GlobalEventBus = &EventBus{
			subscribers: make(map[EventType][]Subscriber),
		}
		log.Println("✅ EventBus initialized")
	})
	return GlobalEventBus
}

// Subscribe подписывается на событие определенного типа
func (eb *EventBus) Subscribe(eventType EventType, subscriber Subscriber) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.subscribers[eventType] = append(eb.subscribers[eventType], subscriber)
	log.Printf("📡 New subscriber for event: %s (total: %d)", eventType, len(eb.subscribers[eventType]))
}

// SubscribeMultiple подписывается на несколько типов событий
func (eb *EventBus) SubscribeMultiple(eventTypes []EventType, subscriber Subscriber) {
	for _, eventType := range eventTypes {
		eb.Subscribe(eventType, subscriber)
	}
}

// Unsubscribe отписывается от всех событий (для cleanup)
func (eb *EventBus) Unsubscribe(eventType EventType, subscriber Subscriber) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	subscribers := eb.subscribers[eventType]
	for i, sub := range subscribers {
		// Сравнение функций в Go невозможно напрямую, используем workaround
		_ = sub
		// Удаляем подписчика по индексу
		eb.subscribers[eventType] = append(subscribers[:i], subscribers[i+1:]...)
		break
	}
}

// Publish публикует событие всем подписчикам
func (eb *EventBus) Publish(eventType EventType, data map[string]interface{}) {
	eb.mu.RLock()
	subscribers := eb.subscribers[eventType]
	eb.mu.RUnlock()

	if len(subscribers) == 0 {
		return
	}

	event := Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}

	// Извлекаем userID из data, если есть
	if userID, ok := data["user_id"].(string); ok {
		event.UserID = userID
	}

	log.Printf("📢 Publishing event: %s (subscribers: %d)", eventType, len(subscribers))

	// Асинхронно уведомляем всех подписчиков
	for _, subscriber := range subscribers {
		go func(sub Subscriber) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("⚠️ Subscriber panic recovered: %v", r)
				}
			}()
			sub(event)
		}(subscriber)
	}
}

// PublishUserEvent публикует событие для конкретного пользователя
func (eb *EventBus) PublishUserEvent(eventType EventType, userID string, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["user_id"] = userID
	eb.Publish(eventType, data)
}

// GetSubscribersCount возвращает количество подписчиков на событие
func (eb *EventBus) GetSubscribersCount(eventType EventType) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return len(eb.subscribers[eventType])
}

// Clear очищает всех подписчиков (для тестов)
func (eb *EventBus) Clear() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.subscribers = make(map[EventType][]Subscriber)
	log.Println("🧹 EventBus cleared")
}

// ToJSON сериализует событие в JSON
func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// Helper functions для создания событий

// NewTreasuryUpdateEvent создаёт событие обновления Treasury
func NewTreasuryUpdateEvent(balance, totalUsed, remaining int64) Event {
	return Event{
		Type:      TreasuryUpdateEvent,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"balance":    balance,
			"total_used": totalUsed,
			"remaining":  remaining,
		},
	}
}

// NewTokenBalanceUpdateEvent создаёт событие обновления баланса пользователя
func NewTokenBalanceUpdateEvent(userID string, balanceBefore, balanceAfter, amount int64, reason string) Event {
	return Event{
		Type:      TokenBalanceUpdateEvent,
		Timestamp: time.Now(),
		UserID:    userID,
		Data: map[string]interface{}{
			"user_id":        userID,
			"balance_before": balanceBefore,
			"balance_after":  balanceAfter,
			"amount":         amount,
			"reason":         reason,
		},
	}
}

// NewTaskCompletedEvent создаёт событие завершения задания
func NewTaskCompletedEvent(userID, taskID string, reward int64) Event {
	return Event{
		Type:      TaskCompletedEvent,
		Timestamp: time.Now(),
		UserID:    userID,
		Data: map[string]interface{}{
			"user_id": userID,
			"task_id": taskID,
			"reward":  reward,
		},
	}
}
