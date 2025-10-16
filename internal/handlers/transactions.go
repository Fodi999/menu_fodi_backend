package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/gorilla/mux"
)

// 📊 GET /api/businesses/{id}/transactions
func GetBusinessTransactions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	businessID := vars["id"]

	db := database.GetDB()

	// Получаем все транзакции для бизнеса
	var transactions []models.Transaction
	query := db.Where("business_id = ?", businessID).Preload("Business").Order("created_at DESC")

	// Опциональный фильтр по типу транзакции
	txType := r.URL.Query().Get("type")
	if txType != "" {
		query = query.Where("tx_type = ?", txType)
	}

	// Пагинация
	limit := r.URL.Query().Get("limit")
	if limit != "" {
		query = query.Limit(parseLimit(limit, 50))
	}

	if err := query.Find(&transactions).Error; err != nil {
		log.Printf("[TRANSACTION] ❌ Error fetching business transactions: %v", err)
		http.Error(w, "Failed to fetch transactions", http.StatusInternalServerError)
		return
	}

	// Статистика
	var stats struct {
		TotalBuyTransactions  int64
		TotalSellTransactions int64
		TotalBuyAmount        float64
		TotalSellAmount       float64
		TotalTokensBought     int64
		TotalTokensSold       int64
	}

	db.Model(&models.Transaction{}).
		Where("business_id = ? AND tx_type = ?", businessID, "buy").
		Count(&stats.TotalBuyTransactions)

	db.Model(&models.Transaction{}).
		Where("business_id = ? AND tx_type = ?", businessID, "sell").
		Count(&stats.TotalSellTransactions)

	db.Model(&models.Transaction{}).
		Where("business_id = ? AND tx_type = ?", businessID, "buy").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.TotalBuyAmount)

	db.Model(&models.Transaction{}).
		Where("business_id = ? AND tx_type = ?", businessID, "sell").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.TotalSellAmount)

	db.Model(&models.Transaction{}).
		Where("business_id = ? AND tx_type = ?", businessID, "buy").
		Select("COALESCE(SUM(tokens), 0)").
		Scan(&stats.TotalTokensBought)

	db.Model(&models.Transaction{}).
		Where("business_id = ? AND tx_type = ?", businessID, "sell").
		Select("COALESCE(SUM(tokens), 0)").
		Scan(&stats.TotalTokensSold)

	log.Printf("[TRANSACTION] 📊 Fetched %d transactions for business %s", len(transactions), businessID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "✅ Business transactions fetched",
		"count":        len(transactions),
		"transactions": transactions,
		"stats": map[string]interface{}{
			"totalBuyTransactions":  stats.TotalBuyTransactions,
			"totalSellTransactions": stats.TotalSellTransactions,
			"totalBuyAmount":        stats.TotalBuyAmount,
			"totalSellAmount":       stats.TotalSellAmount,
			"totalTokensBought":     stats.TotalTokensBought,
			"totalTokensSold":       stats.TotalTokensSold,
			"netAmount":             stats.TotalBuyAmount - stats.TotalSellAmount,
			"netTokens":             stats.TotalTokensBought - stats.TotalTokensSold,
		},
	})
}

// 📊 GET /api/users/{id}/transactions
func GetUserTransactions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	db := database.GetDB()

	// Получаем все транзакции пользователя (как отправитель или получатель)
	var transactions []models.Transaction
	query := db.Where("from_user = ? OR to_user = ?", userID, userID).
		Preload("Business").
		Order("created_at DESC")

	// Опциональный фильтр по типу транзакции
	txType := r.URL.Query().Get("type")
	if txType != "" {
		query = query.Where("tx_type = ?", txType)
	}

	// Опциональный фильтр по бизнесу
	businessID := r.URL.Query().Get("businessId")
	if businessID != "" {
		query = query.Where("business_id = ?", businessID)
	}

	// Пагинация
	limit := r.URL.Query().Get("limit")
	if limit != "" {
		query = query.Limit(parseLimit(limit, 50))
	}

	if err := query.Find(&transactions).Error; err != nil {
		log.Printf("[TRANSACTION] ❌ Error fetching user transactions: %v", err)
		http.Error(w, "Failed to fetch transactions", http.StatusInternalServerError)
		return
	}

	// Статистика
	var stats struct {
		TotalBuyAmount    float64
		TotalSellAmount   float64
		TotalTokensBought int64
		TotalTokensSold   int64
		TotalInvested     float64
		TotalReturned     float64
	}

	// Транзакции покупки (пользователь покупает токены)
	db.Model(&models.Transaction{}).
		Where("from_user = ? AND tx_type = ?", userID, "buy").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.TotalBuyAmount)

	db.Model(&models.Transaction{}).
		Where("from_user = ? AND tx_type = ?", userID, "buy").
		Select("COALESCE(SUM(tokens), 0)").
		Scan(&stats.TotalTokensBought)

	// Транзакции продажи (пользователь продает токены)
	db.Model(&models.Transaction{}).
		Where("to_user = ? AND tx_type = ?", userID, "sell").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.TotalSellAmount)

	db.Model(&models.Transaction{}).
		Where("to_user = ? AND tx_type = ?", userID, "sell").
		Select("COALESCE(SUM(tokens), 0)").
		Scan(&stats.TotalTokensSold)

	stats.TotalInvested = stats.TotalBuyAmount
	stats.TotalReturned = stats.TotalSellAmount

	log.Printf("[TRANSACTION] 📊 Fetched %d transactions for user %s", len(transactions), userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "✅ User transactions fetched",
		"count":        len(transactions),
		"transactions": transactions,
		"stats": map[string]interface{}{
			"totalTokensBought": stats.TotalTokensBought,
			"totalTokensSold":   stats.TotalTokensSold,
			"totalInvested":     stats.TotalInvested,
			"totalReturned":     stats.TotalReturned,
			"netProfit":         stats.TotalReturned - stats.TotalInvested,
			"netTokens":         stats.TotalTokensBought - stats.TotalTokensSold,
		},
	})
}

// 📈 GET /api/transactions/analytics
func GetTransactionAnalytics(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()

	businessID := r.URL.Query().Get("businessId")
	if businessID == "" {
		http.Error(w, "businessId query parameter required", http.StatusBadRequest)
		return
	}

	// Получаем транзакции с группировкой по дате
	var dailyStats []struct {
		Date        string  `json:"date"`
		BuyCount    int64   `json:"buyCount"`
		SellCount   int64   `json:"sellCount"`
		BuyAmount   float64 `json:"buyAmount"`
		SellAmount  float64 `json:"sellAmount"`
		BuyTokens   int64   `json:"buyTokens"`
		SellTokens  int64   `json:"sellTokens"`
	}

	// Группируем по дате за последние 30 дней
	query := `
		SELECT 
			DATE(created_at) as date,
			COUNT(CASE WHEN tx_type = 'buy' THEN 1 END) as buy_count,
			COUNT(CASE WHEN tx_type = 'sell' THEN 1 END) as sell_count,
			COALESCE(SUM(CASE WHEN tx_type = 'buy' THEN amount END), 0) as buy_amount,
			COALESCE(SUM(CASE WHEN tx_type = 'sell' THEN amount END), 0) as sell_amount,
			COALESCE(SUM(CASE WHEN tx_type = 'buy' THEN tokens END), 0) as buy_tokens,
			COALESCE(SUM(CASE WHEN tx_type = 'sell' THEN tokens END), 0) as sell_tokens
		FROM "Transaction"
		WHERE business_id = ?
			AND created_at >= NOW() - INTERVAL '30 days'
		GROUP BY DATE(created_at)
		ORDER BY date DESC
	`

	if err := db.Raw(query, businessID).Scan(&dailyStats).Error; err != nil {
		log.Printf("[TRANSACTION] ❌ Error fetching analytics: %v", err)
		http.Error(w, "Failed to fetch analytics", http.StatusInternalServerError)
		return
	}

	log.Printf("[TRANSACTION] 📈 Fetched analytics for business %s: %d days", businessID, len(dailyStats))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "✅ Transaction analytics fetched",
		"businessId": businessID,
		"period":     "30 days",
		"data":       dailyStats,
	})
}

// Вспомогательная функция для парсинга лимита
func parseLimit(limitStr string, defaultLimit int) int {
	var limit int
	if _, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil || limit <= 0 || limit > 1000 {
		return defaultLimit
	}
	return limit
}
