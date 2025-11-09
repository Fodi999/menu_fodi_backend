package dto

type AdminStats struct {
	TotalUsers    int64   `json:"totalUsers"`
	TotalOrders   int64   `json:"totalOrders"`
	TotalProducts int64   `json:"totalProducts"`
	Revenue       float64 `json:"revenue"`
}

type RecentOrder struct {
	ID        string  `json:"id"`
	UserEmail string  `json:"userEmail"`
	Status    string  `json:"status"`
	Total     float64 `json:"total"`
	CreatedAt string  `json:"createdAt"`
}
