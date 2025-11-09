package service

import (
	"log"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/metrics/dto"
)

type MetricsService struct{}

func NewMetricsService() *MetricsService {
	return &MetricsService{}
}

func (s *MetricsService) GetBusinessMetrics(businessID string) (*dto.BusinessMetrics, error) {
	db := database.GetDB()

	var token models.BusinessToken
	if err := db.First(&token, "business_id = ?", businessID).Error; err != nil {
		return nil, err
	}

	var investorCount int64
	var totalInvested float64

	db.Model(&models.BusinessSubscription{}).Where("business_id = ?", businessID).Count(&investorCount)
	db.Model(&models.BusinessSubscription{}).Where("business_id = ?", businessID).Select("COALESCE(SUM(invested), 0)").Scan(&totalInvested)

	marketCap := float64(token.TotalSupply) * token.Price
	roi := 0.0
	if totalInvested > 0 {
		roi = ((marketCap - totalInvested) / totalInvested) * 100.0
	}

	avgROI := 0.0
	if investorCount > 0 {
		avgROI = roi / float64(investorCount)
	}

	initialPrice := 19.0
	priceChange := ((token.Price - initialPrice) / initialPrice) * 100

	var buyTxCount, sellTxCount int64
	var buyVolume, sellVolume float64
	var tokensSold int64

	db.Model(&models.Transaction{}).Where("business_id = ? AND tx_type = ?", businessID, "buy").Count(&buyTxCount)
	db.Model(&models.Transaction{}).Where("business_id = ? AND tx_type = ?", businessID, "sell").Count(&sellTxCount)
	db.Model(&models.Transaction{}).Where("business_id = ? AND tx_type = ?", businessID, "buy").Select("COALESCE(SUM(amount), 0)").Scan(&buyVolume)
	db.Model(&models.Transaction{}).Where("business_id = ? AND tx_type = ?", businessID, "sell").Select("COALESCE(SUM(amount), 0)").Scan(&sellVolume)
	db.Model(&models.Transaction{}).Where("business_id = ? AND tx_type = ?", businessID, "buy").Select("COALESCE(SUM(tokens), 0)").Scan(&tokensSold)

	var dailyActive, weeklyActive int64
	db.Raw(`SELECT COUNT(DISTINCT from_user) FROM "Transaction" WHERE business_id = ? AND created_at >= NOW() - INTERVAL '24 hours'`, businessID).Scan(&dailyActive)
	db.Raw(`SELECT COUNT(DISTINCT from_user) FROM "Transaction" WHERE business_id = ? AND created_at >= NOW() - INTERVAL '7 days'`, businessID).Scan(&weeklyActive)

	tokenVelocity := float64(0)
	if token.TotalSupply > 0 {
		tokenVelocity = float64(buyTxCount+sellTxCount) / float64(token.TotalSupply)
	}

	avgInvestment := float64(0)
	if investorCount > 0 {
		avgInvestment = totalInvested / float64(investorCount)
	}

	metrics := &dto.BusinessMetrics{
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
		TotalReturned:     0,
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
		InvestorGrowth:    0,
		PriceVolatility:   0,
	}

	log.Printf("[METRICS] Calculated for business %s: Price=$%.2f, Investors=%d, MarketCap=$%.2f, ROI=%.1f%%",
		businessID, token.Price, investorCount, marketCap, roi)

	return metrics, nil
}
