package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/gorilla/mux"
)

// BusinessMetrics содержит AI-метрики для бизнеса
type BusinessMetrics struct {
	BusinessID string `json:"businessId"`

	// Token Metrics
	TokenSymbol     string  `json:"tokenSymbol"`
	CurrentPrice    float64 `json:"currentPrice"`
	InitialPrice    float64 `json:"initialPrice"`
	PriceChange     float64 `json:"priceChange"` // %
	TotalSupply     int64   `json:"totalSupply"`
	TokensSold      int64   `json:"tokensSold"`
	TokensAvailable int64   `json:"tokensAvailable"`
	MarketCap       float64 `json:"marketCap"` // TotalSupply × Price

	// Investment Metrics
	TotalInvestors int     `json:"totalInvestors"`
	TotalInvested  float64 `json:"totalInvested"`
	TotalReturned  float64 `json:"totalReturned"`
	NetInflow      float64 `json:"netInflow"` // TotalInvested - TotalReturned
	AvgInvestment  float64 `json:"avgInvestment"`

	// Transaction Metrics
	TotalBuyTx  int64   `json:"totalBuyTransactions"`
	TotalSellTx int64   `json:"totalSellTransactions"`
	BuyVolume   float64 `json:"buyVolume"`
	SellVolume  float64 `json:"sellVolume"`
	NetVolume   float64 `json:"netVolume"`

	// Activity Metrics
	DailyActiveUsers  int `json:"dailyActiveUsers"`  // За последние 24ч
	WeeklyActiveUsers int `json:"weeklyActiveUsers"` // За последние 7 дней

	// ROI Metrics
	ROI            float64 `json:"roi"`            // (CurrentPrice - InitialPrice) / InitialPrice × 100
	AvgInvestorROI float64 `json:"avgInvestorROI"` // Средний ROI всех инвесторов

	// Growth Metrics
	TokenVelocity   float64 `json:"tokenVelocity"`   // Transactions / Supply
	InvestorGrowth  float64 `json:"investorGrowth"`  // % прироста за неделю
	PriceVolatility float64 `json:"priceVolatility"` // Стандартное отклонение цены
}

// 📊 GET /api/metrics/{businessId}
func GetBusinessMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	businessID := vars["businessId"]

	db := database.GetDB()

	// 1️⃣ Получаем токен бизнеса
	var token models.BusinessToken
	if err := db.First(&token, "business_id = ?", businessID).Error; err != nil {
		http.Error(w, "Business token not found", http.StatusNotFound)
		return
	}

	// 2️⃣ Подсчитываем количество инвесторов и общий инвестиционный объём
	var investorCount int64
	var totalInvested float64

	db.Model(&models.BusinessSubscription{}).
		Where("business_id = ?", businessID).
		Count(&investorCount)

	db.Model(&models.BusinessSubscription{}).
		Where("business_id = ?", businessID).
		Select("COALESCE(SUM(invested), 0)").
		Scan(&totalInvested)

	// 3️⃣ Рассчитываем рыночную капитализацию и ROI
	marketCap := float64(token.TotalSupply) * token.Price
	roi := 0.0
	if totalInvested > 0 {
		roi = ((marketCap - totalInvested) / totalInvested) * 100.0
	}

	// 4️⃣ Средний ROI инвестора
	avgROI := 0.0
	if investorCount > 0 {
		avgROI = roi / float64(investorCount)
	}

	// 5️⃣ Дополнительные метрики (опционально)
	initialPrice := 19.0
	priceChange := ((token.Price - initialPrice) / initialPrice) * 100

	// Метрики транзакций
	var buyTxCount, sellTxCount int64
	var buyVolume, sellVolume float64
	var tokensSold int64

	db.Model(&models.Transaction{}).
		Where("business_id = ? AND tx_type = ?", businessID, "buy").
		Count(&buyTxCount)

	db.Model(&models.Transaction{}).
		Where("business_id = ? AND tx_type = ?", businessID, "sell").
		Count(&sellTxCount)

	db.Model(&models.Transaction{}).
		Where("business_id = ? AND tx_type = ?", businessID, "buy").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&buyVolume)

	db.Model(&models.Transaction{}).
		Where("business_id = ? AND tx_type = ?", businessID, "sell").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&sellVolume)

	db.Model(&models.Transaction{}).
		Where("business_id = ? AND tx_type = ?", businessID, "buy").
		Select("COALESCE(SUM(tokens), 0)").
		Scan(&tokensSold)

	// Активность пользователей
	var dailyActive, weeklyActive int64

	db.Raw(`
		SELECT COUNT(DISTINCT from_user) 
		FROM "Transaction" 
		WHERE business_id = ? 
			AND created_at >= NOW() - INTERVAL '24 hours'
	`, businessID).Scan(&dailyActive)

	db.Raw(`
		SELECT COUNT(DISTINCT from_user) 
		FROM "Transaction" 
		WHERE business_id = ? 
			AND created_at >= NOW() - INTERVAL '7 days'
	`, businessID).Scan(&weeklyActive)

	// Token velocity
	tokenVelocity := float64(0)
	if token.TotalSupply > 0 {
		tokenVelocity = float64(buyTxCount+sellTxCount) / float64(token.TotalSupply)
	}

	// Средний инвестмент
	avgInvestment := float64(0)
	if investorCount > 0 {
		avgInvestment = totalInvested / float64(investorCount)
	}

	metrics := BusinessMetrics{
		BusinessID:        businessID,
		TokenSymbol:       token.Symbol,
		CurrentPrice:      token.Price,
		InitialPrice:      initialPrice,
		PriceChange:       priceChange,
		TotalSupply:       token.TotalSupply,
		TokensSold:        tokensSold,
		TokensAvailable:   token.TotalSupply - tokensSold,
		MarketCap:         marketCap,
		TotalInvestors:    int(investorCount),
		TotalInvested:     totalInvested,
		TotalReturned:     0, // TODO: рассчитать из sell транзакций
		NetInflow:         totalInvested,
		AvgInvestment:     avgInvestment,
		TotalBuyTx:        buyTxCount,
		TotalSellTx:       sellTxCount,
		BuyVolume:         buyVolume,
		SellVolume:        sellVolume,
		NetVolume:         buyVolume - sellVolume,
		DailyActiveUsers:  int(dailyActive),
		WeeklyActiveUsers: int(weeklyActive),
		ROI:               roi,
		AvgInvestorROI:    avgROI,
		TokenVelocity:     tokenVelocity,
		InvestorGrowth:    0, // TODO: рассчитать прирост за неделю
		PriceVolatility:   0, // TODO: рассчитать волатильность
	}

	log.Printf("[METRICS] 📊 Calculated metrics for business %s: Price=$%.2f (%.1f%%), Investors=%d, MarketCap=$%.2f, ROI=%.1f%%",
		businessID, token.Price, priceChange, investorCount, marketCap, roi)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "✅ Business metrics calculated",
		"metrics": metrics,
	})
}
