package service

import (
	"fmt"
	"strings"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"gorm.io/gorm"
)

// HintService - сервис подсказок
type HintService struct {
	db *gorm.DB
}

// NewHintService создает новый сервис
func NewHintService(db *gorm.DB) *HintService {
	return &HintService{db: db}
}

// SearchProducts ищет продукты по запросу
func (s *HintService) SearchProducts(question string) ([]models.Product, error) {
	var products []models.Product
	query := strings.ToLower(question)

	err := s.db.Where("LOWER(name) LIKE ?", "%"+query+"%").
		Or("LOWER(category) LIKE ?", "%"+query+"%").
		Limit(5).
		Find(&products).Error

	if err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	return products, nil
}

// GenerateHint генерирует текстовую подсказку
func (s *HintService) GenerateHint(products []models.Product) string {
	if len(products) == 0 {
		return "К сожалению, ничего не найдено. Попробуйте другой запрос."
	}

	hint := "Вот что я нашел по вашему запросу:"
	for i, p := range products {
		hint += fmt.Sprintf("\n%d. %s - %.2f сом", i+1, p.Name, p.Price)
	}

	return hint
}
