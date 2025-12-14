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

	// 4. Формируем ответ
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

		result = append(result, models.FridgeItemListResponse{
			ID:       item.ID,
			Name:     item.Ingredient.Name,
			Quantity: item.Quantity,
			Unit:     item.Unit,
			DaysLeft: daysLeft,
			Status:   models.GetFridgeItemStatus(daysLeft),
		})
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
		result = append(result, models.FridgeItemListResponse{
			ID:       item.ID,
			Name:     item.Ingredient.Name,
			Quantity: item.Quantity,
			Unit:     item.Unit,
			DaysLeft: daysLeft,
			Status:   models.GetFridgeItemStatus(daysLeft),
		})
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
