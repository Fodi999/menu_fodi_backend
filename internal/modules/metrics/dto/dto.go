package dto

type BusinessMetrics struct {
	BusinessID        string  `json:"businessId"`
	TokenSymbol       string  `json:"tokenSymbol"`
	CurrentPrice      float64 `json:"currentPrice"`
	InitialPrice      float64 `json:"initialPrice"`
	PriceChange       float64 `json:"priceChange"`
	TotalSupply       int64   `json:"totalSupply"`
	TokensSold        int64   `json:"tokensSold"`
	TokensAvailable   int64   `json:"tokensAvailable"`
	MarketCap         float64 `json:"marketCap"`
	TotalInvestors    int     `json:"totalInvestors"`
	TotalInvested     float64 `json:"totalInvested"`
	TotalReturned     float64 `json:"totalReturned"`
	NetInflow         float64 `json:"netInflow"`
	AvgInvestment     float64 `json:"avgInvestment"`
	TotalBuyTx        int64   `json:"totalBuyTransactions"`
	TotalSellTx       int64   `json:"totalSellTransactions"`
	BuyVolume         float64 `json:"buyVolume"`
	SellVolume        float64 `json:"sellVolume"`
	NetVolume         float64 `json:"netVolume"`
	DailyActiveUsers  int     `json:"dailyActiveUsers"`
	WeeklyActiveUsers int     `json:"weeklyActiveUsers"`
	ROI               float64 `json:"roi"`
	AvgInvestorROI    float64 `json:"avgInvestorROI"`
	TokenVelocity     float64 `json:"tokenVelocity"`
	InvestorGrowth    float64 `json:"investorGrowth"`
	PriceVolatility   float64 `json:"priceVolatility"`
}
