package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
)

// FridgeService сервис для работы с холодильником
type FridgeService struct {
	fridgeRepo     *database.UserFridgeRepository
	ingredientRepo *database.IngredientRepository
}

// NewFridgeService создает новый экземпляр сервиса
func NewFridgeService(fridgeRepo *database.UserFridgeRepository, ingredientRepo *database.IngredientRepository) *FridgeService {
	return &FridgeService{
		fridgeRepo:     fridgeRepo,
		ingredientRepo: ingredientRepo,
	}
}

// AddItem добавляет продукт в холодильник
func (s *FridgeService) AddItem(userID string, req models.CreateFridgeItemRequest) (*models.FridgeItemResponse, error) {
	// 1. Получаем ингредиент из каталога
	ingredient, err := s.ingredientRepo.GetIngredientByID(req.IngredientID)
	if err != nil {
		return nil, fmt.Errorf("ingredient not found: %w", err)
	}

	// 2. Вычисляем expires_at на основе defaultShelfLifeDays
	expiresAt := s.calculateExpiresAt(ingredient.DefaultShelfLifeDays)

	// 3. Создаем запись в холодильнике
	item := &models.UserFridgeItem{
		UserID:       userID,
		IngredientID: req.IngredientID,
		Quantity:     req.Quantity,
		Unit:         ingredient.Unit, // Копируем unit из каталога
		ExpiresAt:    expiresAt,
	}

	if err := s.fridgeRepo.Create(item); err != nil {
		return nil, fmt.Errorf("failed to create fridge item: %w", err)
	}

	// 4. Если указана цена, добавляем событие в историю (event sourcing)
	if req.PriceInput != nil {
		normalized, err := s.normalizePrice(req.PriceInput.Value, req.PriceInput.Per, ingredient.Unit)
		if err != nil {
			return nil, fmt.Errorf("invalid price input: %w", err)
		}

		// Добавляем первое событие цены
		priceReq := models.AddPriceRequest{
			PricePerUnit: normalized,
			Currency:     "PLN",
			Source:       "manual",
		}
		
		if err := s.AddPrice(userID, item.ID, priceReq); err != nil {
			// Не фейлим весь запрос из-за цены, просто логируем
			fmt.Printf("warning: failed to add initial price: %v\n", err)
		}
	}

	// 5. Формируем ответ
	return s.buildFridgeItemResponse(item, ingredient), nil
}

// GetUserItems возвращает список продуктов пользователя
func (s *FridgeService) GetUserItems(userID string) ([]models.FridgeItemListResponse, error) {
	items, err := s.fridgeRepo.GetUserFridgeItems(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get fridge items: %w", err)
	}

	result := make([]models.FridgeItemListResponse, 0, len(items))
	for _, item := range items {
		if item.Ingredient == nil {
			continue
		}

		daysLeft := s.calculateDaysLeft(item.ExpiresAt)
		totalPrice := s.calculateTotalPrice(item.Quantity, item.CurrentPricePerUnit)

		response := models.FridgeItemListResponse{
			ID:       item.ID,
			Name:     item.Ingredient.Name,
			Category: item.Ingredient.Category, // Добавляем категорию для группировки
			Quantity: item.Quantity,
			Unit:     item.Unit,
			DaysLeft: daysLeft,
			Status:   models.GetFridgeItemStatus(daysLeft),
		}

		// Добавляем цену только если она есть (из кэша current_price_*)
		if item.CurrentPricePerUnit != nil {
			response.PricePerUnit = item.CurrentPricePerUnit // Цена за единицу
			response.TotalPrice = totalPrice                  // Общая стоимость
			response.Currency = item.CurrentPriceCurrency     // Валюта
		}

		result = append(result, response)
	}

	return result, nil
}

// GetItemByID возвращает информацию о конкретном продукте
func (s *FridgeService) GetItemByID(itemID string) (*models.FridgeItemResponse, error) {
	item, err := s.fridgeRepo.GetByID(itemID)
	if err != nil {
		return nil, fmt.Errorf("fridge item not found: %w", err)
	}

	if item.Ingredient == nil {
		return nil, errors.New("ingredient not loaded")
	}

	return s.buildFridgeItemResponse(item, item.Ingredient), nil
}

// DeleteItem удаляет продукт из холодильника
func (s *FridgeService) DeleteItem(itemID string, userID string) error {
	// Проверяем, что продукт принадлежит пользователю
	item, err := s.fridgeRepo.GetByID(itemID)
	if err != nil {
		return fmt.Errorf("fridge item not found: %w", err)
	}

	if item.UserID != userID {
		return errors.New("access denied: item does not belong to user")
	}

	return s.fridgeRepo.Delete(itemID)
}

// GetExpiringSoon возвращает продукты с истекающим сроком
func (s *FridgeService) GetExpiringSoon(userID string, days int) ([]models.FridgeItemListResponse, error) {
	items, err := s.fridgeRepo.GetExpiringSoon(userID, days)
	if err != nil {
		return nil, fmt.Errorf("failed to get expiring items: %w", err)
	}

	result := make([]models.FridgeItemListResponse, 0, len(items))
	for _, item := range items {
		if item.Ingredient == nil {
			continue
		}

		daysLeft := s.calculateDaysLeft(item.ExpiresAt)
		totalPrice := s.calculateTotalPrice(item.Quantity, item.CurrentPricePerUnit)

		response := models.FridgeItemListResponse{
			ID:       item.ID,
			Name:     item.Ingredient.Name,
			Category: item.Ingredient.Category, // Добавляем категорию
			Quantity: item.Quantity,
			Unit:     item.Unit,
			DaysLeft: daysLeft,
			Status:   models.GetFridgeItemStatus(daysLeft),
		}

		// Добавляем цену только если она есть (из кэша current_price_*)
		if item.CurrentPricePerUnit != nil {
			response.PricePerUnit = item.CurrentPricePerUnit // Цена за единицу
			response.TotalPrice = totalPrice                  // Общая стоимость
			response.Currency = item.CurrentPriceCurrency     // Валюта
		}

		result = append(result, response)
	}

	return result, nil
}

// ===== PRIVATE HELPERS =====

// calculateExpiresAt вычисляет дату истечения срока на основе defaultShelfLifeDays
func (s *FridgeService) calculateExpiresAt(shelfLifeDays *int) *time.Time {
	days := 7 // Значение по умолчанию (неделя)
	if shelfLifeDays != nil && *shelfLifeDays > 0 {
		days = *shelfLifeDays
	}
	expiresAt := time.Now().AddDate(0, 0, days)
	return &expiresAt
}

// calculateDaysLeft вычисляет количество дней до истечения срока
func (s *FridgeService) calculateDaysLeft(expiresAt *time.Time) int {
	if expiresAt == nil {
		return 999 // Срок годности не указан
	}
	duration := time.Until(*expiresAt)
	days := int(duration.Hours() / 24)
	return days
}

// normalizePrice нормализует цену к единице измерения в БД (всегда g/ml/szt)
func (s *FridgeService) normalizePrice(value float64, per string, unit string) (float64, error) {
	switch per {
	case "kg":
		if unit != "g" {
			return 0, fmt.Errorf("unit mismatch: cannot convert price per kg to unit %s", unit)
		}
		// 3.20 PLN / kg → 0.0032 PLN / g
		return value / 1000, nil

	case "l":
		if unit != "ml" {
			return 0, fmt.Errorf("unit mismatch: cannot convert price per l to unit %s", unit)
		}
		// 2.50 PLN / l → 0.0025 PLN / ml
		return value / 1000, nil

	case "szt":
		if unit != "szt" {
			return 0, fmt.Errorf("unit mismatch: price per szt requires unit szt, got %s", unit)
		}
		// 1.00 PLN / szt → 1.00 PLN / szt (без изменений)
		return value, nil

	default:
		return 0, fmt.Errorf("unknown price unit: %s (supported: kg, l, szt)", per)
	}
}

// calculateTotalPrice вычисляет общую стоимость продукта
func (s *FridgeService) calculateTotalPrice(quantity float64, pricePerUnit *float64) *float64 {
	if pricePerUnit == nil {
		return nil
	}
	total := quantity * (*pricePerUnit)
	return &total
}

// buildFridgeItemResponse создает ответ для API
func (s *FridgeService) buildFridgeItemResponse(item *models.UserFridgeItem, ingredient *models.Ingredient) *models.FridgeItemResponse {
	daysLeft := s.calculateDaysLeft(item.ExpiresAt)

	expiresAtStr := ""
	if item.ExpiresAt != nil {
		expiresAtStr = item.ExpiresAt.Format("2006-01-02") // ISO 8601 формат (YYYY-MM-DD)
	}

	return &models.FridgeItemResponse{
		ID: item.ID,
		Ingredient: models.IngredientShortInfo{
			Name:     ingredient.Name,
			Unit:     ingredient.Unit,
			Category: ingredient.Category,
		},
		Quantity:  item.Quantity,
		ExpiresAt: expiresAtStr,
		DaysLeft:  daysLeft,
	}
}

// ===== PRICE HISTORY SERVICE METHODS (Event Sourcing) =====

// AddPrice добавляет событие изменения цены (event sourcing)
func (s *FridgeService) AddPrice(userID string, itemID string, req models.AddPriceRequest) error {
	// 1. Проверяем, что продукт существует и принадлежит пользователю
	item, err := s.fridgeRepo.GetByID(itemID)
	if err != nil {
		return fmt.Errorf("fridge item not found: %w", err)
	}

	if item.UserID != userID {
		return errors.New("access denied: item does not belong to user")
	}

	// 2. Валидация source
	validSources := map[string]bool{
		"manual":   true,
		"receipt":  true,
		"estimate": true,
		"market":   true,
		"ai":       true,
	}
	if !validSources[req.Source] {
		return fmt.Errorf("invalid source: %s (allowed: manual, receipt, estimate, market, ai)", req.Source)
	}

	// 3. Добавляем событие в историю
	if err := s.fridgeRepo.InsertPriceHistory(itemID, req.PricePerUnit, req.Currency, req.Source); err != nil {
		return fmt.Errorf("failed to insert price history: %w", err)
	}

	// 4. Обновляем кэш текущей цены (денормализация для производительности)
	if err := s.fridgeRepo.UpdateCurrentPrice(itemID, req.PricePerUnit, req.Currency); err != nil {
		return fmt.Errorf("failed to update current price: %w", err)
	}

	return nil
}

// GetPriceHistory возвращает историю изменения цен
func (s *FridgeService) GetPriceHistory(userID string, itemID string) ([]models.PriceHistoryResponse, error) {
	// 1. Проверяем доступ
	item, err := s.fridgeRepo.GetByID(itemID)
	if err != nil {
		return nil, fmt.Errorf("fridge item not found: %w", err)
	}

	if item.UserID != userID {
		return nil, errors.New("access denied: item does not belong to user")
	}

	// 2. Получаем историю
	history, err := s.fridgeRepo.GetPriceHistory(itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get price history: %w", err)
	}

	// 3. Преобразуем в DTO
	result := make([]models.PriceHistoryResponse, 0, len(history))
	for _, h := range history {
		result = append(result, models.PriceHistoryResponse{
			ID:           h.ID,
			PricePerUnit: h.PricePerUnit,
			Currency:     h.Currency,
			Source:       h.Source,
			CreatedAt:    h.CreatedAt,
		})
	}

	return result, nil
}
