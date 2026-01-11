package service

import (
	"fmt"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
)

// FilterMetadata - метаданные для UI фильтров
type FilterMetadata struct {
	Categories   []string `json:"categories"`
	Difficulties []string `json:"difficulties"`
	TimeRanges   []int    `json:"timeRanges"`
	SourceTypes  []string `json:"sourceTypes"`
}

// GetRecipeFilterMetadata - возвращает доступные значения для фильтров
func (s *adminService) GetRecipeFilterMetadata() (*FilterMetadata, error) {
	meta := &FilterMetadata{
		TimeRanges:  []int{10, 20, 30, 60, 90, 120}, // предопределенные диапазоны
		SourceTypes: []string{"ai", "manual", "traditional"},
	}

	// Получаем уникальные категории из базы
	var categories []string
	if err := s.db.Model(&models.RecipeCatalog{}).
		Distinct("category").
		Order("category ASC").
		Pluck("category", &categories).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}
	meta.Categories = categories

	// Получаем уникальные уровни сложности
	var difficulties []string
	if err := s.db.Model(&models.RecipeCatalog{}).
		Distinct("difficulty").
		Order("difficulty ASC").
		Pluck("difficulty", &difficulties).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch difficulties: %w", err)
	}
	meta.Difficulties = difficulties

	return meta, nil
}
