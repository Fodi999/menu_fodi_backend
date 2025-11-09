package service

import (
	"fmt"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/stats/dto"
	"gorm.io/gorm"
)

type StatsService struct {
	db *gorm.DB
}

func NewStatsService(db *gorm.DB) *StatsService {
	return &StatsService{db: db}
}

func (s *StatsService) GetAdminStats() (*dto.AdminStats, error) {
	var stats dto.AdminStats

	if err := s.db.Table("User").Count(&stats.TotalUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	if err := s.db.Table("Order").Count(&stats.TotalOrders).Error; err != nil {
		return nil, fmt.Errorf("failed to count orders: %w", err)
	}

	if err := s.db.Table("Product").Count(&stats.TotalProducts).Error; err != nil {
		return nil, fmt.Errorf("failed to count products: %w", err)
	}

	if err := s.db.Table("Order").
		Where("status != ?", "cancelled").
		Select("COALESCE(SUM(total), 0)").
		Row().Scan(&stats.Revenue); err != nil {
		return nil, fmt.Errorf("failed to calculate revenue: %w", err)
	}

	return &stats, nil
}

func (s *StatsService) GetRecentOrders(limit int) ([]dto.RecentOrder, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	var orders []dto.RecentOrder

	result := s.db.Raw(`
		SELECT 
			o.id,
			u.email as user_email,
			o.status,
			o.total,
			o.created_at
		FROM "Order" o
		LEFT JOIN "User" u ON o.user_id = u.id
		ORDER BY o.created_at DESC
		LIMIT ?
	`, limit).Scan(&orders)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch recent orders: %w", result.Error)
	}

	return orders, nil
}
