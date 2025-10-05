package handlers

import (
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
)

// AdminStats статистика для админ-панели
type AdminStats struct {
	TotalUsers    int64 `json:"totalUsers"`
	TotalOrders   int64 `json:"totalOrders"`
	TotalProducts int64 `json:"totalProducts"`
	Revenue       float64 `json:"revenue"`
}

// GetAdminStats получение статистики (только для админа)
func GetAdminStats(w http.ResponseWriter, r *http.Request) {
	var stats AdminStats
	
	// Подсчет пользователей
	var userCount int64
	database.DB.Table("User").Count(&userCount)
	stats.TotalUsers = userCount
	
	// Подсчет заказов
	var orderCount int64
	database.DB.Table("Order").Count(&orderCount)
	stats.TotalOrders = orderCount
	
	// Подсчет продуктов
	var productCount int64
	database.DB.Table("Product").Count(&productCount)
	stats.TotalProducts = productCount
	
	// Подсчет выручки
	var revenue float64
	database.DB.Table("Order").
		Where("status != ?", "cancelled").
		Select("COALESCE(SUM(total), 0)").
		Row().Scan(&revenue)
	stats.Revenue = revenue
	
	utils.RespondWithJSON(w, http.StatusOK, stats)
}

// RecentOrder недавний заказ для админ-панели
type RecentOrder struct {
	ID        string  `json:"id"`
	UserEmail string  `json:"userEmail"`
	Status    string  `json:"status"`
	Total     float64 `json:"total"`
	CreatedAt string  `json:"createdAt"`
}

// GetRecentOrders получение недавних заказов (только для админа)
func GetRecentOrders(w http.ResponseWriter, r *http.Request) {
	var orders []RecentOrder
	
	// Получаем последние 10 заказов с email пользователей
	database.DB.Raw(`
		SELECT 
			o.id,
			u.email as user_email,
			o.status,
			o.total,
			o."createdAt"
		FROM "Order" o
		LEFT JOIN "User" u ON o."userId" = u.id
		ORDER BY o."createdAt" DESC
		LIMIT 10
	`).Scan(&orders)
	
	utils.RespondWithJSON(w, http.StatusOK, orders)
}
