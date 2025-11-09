package service

import "gorm.io/gorm"

// HealthService - сервис health check
type HealthService struct {
	db *gorm.DB
}

// NewHealthService создает новый сервис
func NewHealthService(db *gorm.DB) *HealthService {
	return &HealthService{db: db}
}

// CheckHealth проверяет здоровье системы
func (s *HealthService) CheckHealth() (string, error) {
	// Проверка подключения к БД
	sqlDB, err := s.db.DB()
	if err != nil {
		return "disconnected", err
	}
	
	if err := sqlDB.Ping(); err != nil {
		return "disconnected", err
	}
	
	return "connected", nil
}
