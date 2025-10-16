package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/services"
	"github.com/gorilla/mux"
)

var subscriptionService = services.NewSubscriptionService()

// SubscribeRequest запрос на подписку (инвестицию)
type SubscribeRequest struct {
	TokensAmount int64 `json:"tokensAmount"`
}

// 💰 POST /api/businesses/{id}/subscribe
func SubscribeToBusiness(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	businessID := vars["id"]

	// TODO: Получить userID из JWT токена после аутентификации
	// Временно используем тестовый ID
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User ID required (X-User-ID header)", http.StatusUnauthorized)
		return
	}

	var input SubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Валидация
	if input.TokensAmount <= 0 {
		http.Error(w, "Tokens amount must be positive", http.StatusBadRequest)
		return
	}

	// Создание подписки (покупка токенов)
	subscription, transaction, err := subscriptionService.Subscribe(userID, businessID, input.TokensAmount)
	if err != nil {
		log.Printf("[SUBSCRIPTION] ❌ Error subscribing: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "✅ Successfully subscribed to business",
		"subscription": subscription,
		"transaction":  transaction,
	})
}

// ❌ DELETE /api/businesses/{id}/unsubscribe
func UnsubscribeFromBusiness(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	businessID := vars["id"]

	// TODO: Получить userID из JWT токена после аутентификации
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User ID required (X-User-ID header)", http.StatusUnauthorized)
		return
	}

	// Удаление подписки (продажа токенов обратно)
	if err := subscriptionService.Unsubscribe(userID, businessID); err != nil {
		log.Printf("[SUBSCRIPTION] ❌ Error unsubscribing: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "✅ Successfully unsubscribed from business",
	})
}

// 📋 GET /api/users/{id}/subscriptions
func GetUserSubscriptions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	// Опционально: проверить, что запрашивающий имеет право видеть подписки этого пользователя
	// (либо это сам пользователь, либо админ)

	subscriptions, err := subscriptionService.GetUserSubscriptions(userID)
	if err != nil {
		log.Printf("[SUBSCRIPTION] ❌ Error fetching subscriptions: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "✅ User subscriptions fetched",
		"count":         len(subscriptions),
		"subscriptions": subscriptions,
	})
}

// 📊 GET /api/businesses/{id}/subscribers
func GetBusinessSubscribers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	businessID := vars["id"]

	subscribers, err := subscriptionService.GetBusinessSubscribers(businessID)
	if err != nil {
		log.Printf("[SUBSCRIPTION] ❌ Error fetching subscribers: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Рассчитываем общую инвестированную сумму
	var totalInvested float64
	var totalTokensSold int64
	for _, sub := range subscribers {
		totalInvested += sub.Invested
		totalTokensSold += sub.TokensOwned
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":        "✅ Business subscribers fetched",
		"subscriberCount": len(subscribers),
		"totalInvested":  totalInvested,
		"totalTokensSold": totalTokensSold,
		"subscribers":    subscribers,
	})
}

// 📊 GET /api/subscriptions/stats
func GetSubscriptionStats(w http.ResponseWriter, r *http.Request) {
	businessID := r.URL.Query().Get("businessId")
	userID := r.URL.Query().Get("userId")

	if businessID == "" || userID == "" {
		http.Error(w, "Both businessId and userId query parameters required", http.StatusBadRequest)
		return
	}

	subscription, err := subscriptionService.GetSubscriptionStats(userID, businessID)
	if err != nil {
		log.Printf("[SUBSCRIPTION] ❌ Error fetching subscription stats: %v", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "✅ Subscription stats fetched",
		"subscription": subscription,
	})
}
