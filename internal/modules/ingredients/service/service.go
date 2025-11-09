package service

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ingredients/dto"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IngredientsService - сервис управления ингредиентами
type IngredientsService struct {
	repo *database.IngredientRepository
	db   *gorm.DB
}

// NewIngredientsService создает новый сервис
func NewIngredientsService(db *gorm.DB) *IngredientsService {
	return &IngredientsService{
		repo: &database.IngredientRepository{},
		db:   db,
	}
}

// GetAll получение всех ингредиентов со складскими остатками
func (s *IngredientsService) GetAll() ([]models.StockItem, error) {
	return s.repo.FindAll()
}

// GetByID получение одного ингредиента
func (s *IngredientsService) GetByID(id string) (*models.StockItem, error) {
	return s.repo.FindByID(id)
}

// Create создание нового ингредиента
func (s *IngredientsService) Create(req *dto.CreateIngredientRequest) (*models.StockItem, error) {
	// Проверяем обязательные поля
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.Unit == "" {
		return nil, fmt.Errorf("unit is required")
	}

	// 🔍 Автоматическое определение единицы измерения
	if autoUnit := s.detectDefaultUnit(req.Name); autoUnit != "" {
		req.Unit = autoUnit
	}

	// Создаём ингредиент
	ingredient := &models.Ingredient{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Unit:      req.Unit,
		CreatedAt: time.Now(),
	}

	// Генерируем уникальный номер партии
	batchNumber := s.generateBatchNumber(req.Name)

	// Создаём складскую запись
	stockItem := &models.StockItem{
		ID:           uuid.New().String(),
		IngredientID: ingredient.ID,
		Quantity:     req.Quantity,
		UpdatedAt:    time.Now(),
		BatchNumber:  &batchNumber,
	}

	// Присваиваем опциональные поля
	if req.BruttoWeight > 0 {
		stockItem.BruttoWeight = &req.BruttoWeight
	}
	if req.NettoWeight > 0 {
		stockItem.NettoWeight = &req.NettoWeight
	}
	if req.WastePercentage >= 0 {
		stockItem.WastePercentage = &req.WastePercentage
	}
	if req.ExpiryDays > 0 {
		stockItem.ExpiryDays = &req.ExpiryDays
	}
	if req.Supplier != "" {
		stockItem.Supplier = &req.Supplier
	}
	if req.Category != "" {
		stockItem.Category = &req.Category
	}
	if req.PriceBrutto > 0 {
		stockItem.PriceBrutto = &req.PriceBrutto
	}
	if req.PriceNetto > 0 {
		stockItem.PriceNetto = &req.PriceNetto
	}
	if req.PricePerUnit > 0 {
		stockItem.PricePerUnit = &req.PricePerUnit
	}

	log.Printf("📦 StockItem before save (Batch: %s): %+v\n", batchNumber, stockItem)

	// Сохраняем в БД
	if err := s.repo.CreateIngredient(ingredient, stockItem); err != nil {
		return nil, fmt.Errorf("failed to create ingredient: %w", err)
	}

	// Создаем запись о начальном поступлении
	movementQuantity := req.Quantity
	if movementQuantity == 0 && req.BruttoWeight > 0 {
		movementQuantity = req.BruttoWeight
	}
	if movementQuantity == 0 && req.NettoWeight > 0 {
		movementQuantity = req.NettoWeight
	}

	if movementQuantity > 0 {
		movement := &models.StockMovement{
			ID:          uuid.New().String(),
			StockItemID: stockItem.ID,
			Type:        "addition",
			Quantity:    movementQuantity,
			PriceBrutto: stockItem.PriceBrutto,
			PriceNetto:  stockItem.PriceNetto,
			CreatedAt:   time.Now(),
		}
		note := "Начальное поступление"
		movement.Note = &note

		if err := s.db.Create(movement).Error; err != nil {
			log.Printf("⚠️ Failed to create stock movement: %v", err)
		} else {
			log.Printf("✅ Created stock movement: %s for %.2f units", movement.ID, movementQuantity)
		}
	}

	stockItem.Ingredient = ingredient
	return stockItem, nil
}

// Update обновление ингредиента
func (s *IngredientsService) Update(id string, req *dto.UpdateIngredientRequest) (*models.StockItem, error) {
	// Находим складской остаток
	stockItem, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("ingredient not found: %w", err)
	}

	// Обновляем ингредиент
	if req.Name != "" {
		stockItem.Ingredient.Name = req.Name
	}
	if req.Unit != "" {
		stockItem.Ingredient.Unit = req.Unit
	} else if req.Name != "" {
		// 🔍 Автоопределение единицы измерения
		if autoUnit := s.detectDefaultUnit(req.Name); autoUnit != "" {
			stockItem.Ingredient.Unit = autoUnit
		}
	}

	// Обновляем складские данные
	stockItem.Quantity = req.Quantity
	if req.BruttoWeight > 0 {
		stockItem.BruttoWeight = &req.BruttoWeight
	}
	if req.NettoWeight > 0 {
		stockItem.NettoWeight = &req.NettoWeight
	}
	if req.WastePercentage >= 0 {
		stockItem.WastePercentage = &req.WastePercentage
	}
	if req.ExpiryDays > 0 {
		stockItem.ExpiryDays = &req.ExpiryDays
	}
	if req.Supplier != "" {
		stockItem.Supplier = &req.Supplier
	}
	if req.Category != "" {
		stockItem.Category = &req.Category
	}
	if req.PriceBrutto > 0 {
		stockItem.PriceBrutto = &req.PriceBrutto
	}
	if req.PriceNetto > 0 {
		stockItem.PriceNetto = &req.PriceNetto
	}
	if req.PricePerUnit > 0 {
		stockItem.PricePerUnit = &req.PricePerUnit
	}
	stockItem.UpdatedAt = time.Now()

	// Сохраняем в БД
	if err := s.repo.UpdateIngredient(stockItem.Ingredient); err != nil {
		return nil, fmt.Errorf("failed to update ingredient: %w", err)
	}

	if err := s.repo.UpdateStockItem(stockItem); err != nil {
		return nil, fmt.Errorf("failed to update stock: %w", err)
	}

	return stockItem, nil
}

// Delete удаление ингредиента
func (s *IngredientsService) Delete(id string) error {
	return s.repo.DeleteStockItem(id)
}

// GetStockMovements получение истории движений товара
func (s *IngredientsService) GetStockMovements(stockItemID string) ([]models.StockMovement, error) {
	var movements []models.StockMovement
	result := s.db.
		Where("\"stockItemId\" = ?", stockItemID).
		Order("\"createdAt\" DESC").
		Limit(20).
		Find(&movements)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch stock movements: %w", result.Error)
	}

	return movements, nil
}

// detectDefaultUnit возвращает дефолтную единицу измерения
func (s *IngredientsService) detectDefaultUnit(name string) string {
	nameLower := strings.ToLower(name)

	defaultUnits := map[string]string{
		"мук":    "kg",
		"сахар":  "kg",
		"рис":    "kg",
		"круп":   "kg",
		"соль":   "kg",
		"вода":   "l",
		"масло":  "ml",
		"молок":  "l",
		"яйц":    "pcs",
		"лосос":  "kg",
		"сёмг":   "kg",
		"тунец":  "kg",
		"креве":  "kg",
		"угор":   "kg",
		"сыр":    "kg",
		"соус":   "ml",
		"уксус":  "ml",
		"нори":   "pcs",
		"васаби": "kg",
		"имбир":  "kg",
		"авока":  "pcs",
		"огуре":  "pcs",
	}

	for key, unit := range defaultUnits {
		if strings.Contains(nameLower, key) {
			log.Printf("⚙️ Автоматически установлена единица '%s' для ингредиента '%s'", unit, name)
			return unit
		}
	}

	return ""
}

// generateBatchNumber генерирует уникальный номер партии
func (s *IngredientsService) generateBatchNumber(ingredientName string) string {
	now := time.Now()

	// Берём первые 3 руны для поддержки UTF-8
	runes := []rune(ingredientName)
	prefix := ""
	if len(runes) >= 3 {
		prefix = string(runes[:3])
	} else {
		prefix = string(runes)
	}

	prefix = strings.ToUpper(strings.ReplaceAll(prefix, " ", ""))

	// Формат: КРЕ-20251006-020417
	return fmt.Sprintf("%s-%s-%s",
		prefix,
		now.Format("20060102"),
		now.Format("150405"))
}
