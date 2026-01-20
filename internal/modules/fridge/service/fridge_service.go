package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/fridge/dto"
	notificationService "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/notifications/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// FridgeService сервис для работы с холодильником
type FridgeService struct {
	db                  *gorm.DB
	fridgeRepo          *database.UserFridgeRepository
	ingredientRepo      *database.IngredientRepository
	notificationService notificationService.NotificationService
}

// NewFridgeService создает новый экземпляр сервиса
func NewFridgeService(db *gorm.DB, fridgeRepo *database.UserFridgeRepository, ingredientRepo *database.IngredientRepository) *FridgeService {
	return &FridgeService{
		db:                  db,
		fridgeRepo:          fridgeRepo,
		ingredientRepo:      ingredientRepo,
		notificationService: notificationService.NewNotificationService(db),
	}
}

// AddItem добавляет продукт в холодильник
func (s *FridgeService) AddItem(userID string, req models.CreateFridgeItemRequest) (*models.FridgeItemResponse, error) {
	// 1. Получаем ингредиент из каталога
	ingredient, err := s.ingredientRepo.GetIngredientByID(req.IngredientID)
	if err != nil {
		return nil, fmt.Errorf("ingredient not found: %w", err)
	}

	// 2. НОВАЯ ЛОГИКА: Всегда создаём новую запись (отдельную партию)
	// Убрали UNIQUE constraint - теперь можно иметь несколько партий одного продукта:
	//   - с разными датами поступления (arrived_at)
	//   - с разными сроками годности (expires_at)
	//   - с разными ценами (price history)
	// Это правильная модель холодильника: купил молоко сегодня + купил молоко вчера = 2 записи

	// 3. Создаем новую запись (новая партия)
	arrivedAt := time.Now()

	// Вычисляем expires_at автоматически, если не задано явно
	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		expiresAt = req.ExpiresAt
	} else if ingredient.DefaultShelfLifeDays != nil {
		t := arrivedAt.AddDate(0, 0, *ingredient.DefaultShelfLifeDays)
		expiresAt = &t
	}

	// Создаем запись в холодильнике
	item := &models.UserFridgeItem{
		UserID:       userID,
		IngredientID: req.IngredientID,
		Quantity:     req.Quantity,
		Unit:         ingredient.Unit,
		ArrivedAt:    arrivedAt,
		ExpiresAt:    expiresAt,
	}

	if err := s.fridgeRepo.Create(item); err != nil {
		return nil, fmt.Errorf("failed to create fridge item: %w", err)
	}

	// 5. Если указана цена, добавляем событие в историю (event sourcing)
	if req.PriceInput != nil {
		normalized, err := s.normalizePrice(req.PriceInput.Value, req.PriceInput.Per, ingredient.Unit)
		if err != nil {
			logger.Error("failed to normalize price",
				zap.Float64("value", req.PriceInput.Value),
				zap.String("per", req.PriceInput.Per),
				zap.String("unit", ingredient.Unit),
				zap.Error(err))
			return nil, fmt.Errorf("invalid price input: %w", err)
		}

		// Добавляем первое событие цены с исходной ценой и единицей измерения
		priceReq := models.AddPriceRequest{
			PricePerUnit: req.PriceInput.Value, // ✅ Сохраняем ИСХОДНУЮ цену (не нормализованную)
			UnitForPrice: req.PriceInput.Per,   // ✅ Сохраняем единицу измерения (kg, l, pcs)
			Currency:     "PLN",
			Source:       "manual",
		}

		if err := s.AddPrice(userID, item.ID, priceReq); err != nil {
			// Логируем ошибку с деталями
			logger.Error("failed to add initial price",
				zap.String("user_id", userID),
				zap.String("item_id", item.ID),
				zap.Float64("original_price", req.PriceInput.Value),
				zap.String("unit_for_price", req.PriceInput.Per),
				zap.Float64("normalized", normalized),
				zap.Error(err))
			// НЕ фейлим весь запрос из-за цены - продукт уже создан
		}
	}

	// 6. Создаём системное уведомление о добавлении продукта
	s.createItemAddedNotification(userID, item, ingredient)

	// 7. Формируем ответ
	return s.buildFridgeItemResponse(item, ingredient), nil
}

// GetUserItems возвращает список продуктов пользователя
// lang: preferred language for ingredient names ("ru", "pl", "en")
func (s *FridgeService) GetUserItems(userID string, lang string) ([]models.FridgeItemListResponse, error) {
	// Default language if not provided
	if lang == "" {
		lang = "pl"
	}

	// Автоматически очищаем просроченные продукты перед возвратом списка
	if err := s.cleanupExpiredItems(userID); err != nil {
		logger.Warn("failed to cleanup expired items",
			zap.String("user_id", userID),
			zap.Error(err))
		// Не блокируем весь запрос из-за ошибки очистки
	}

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
		status := models.GetFridgeItemStatus(daysLeft)

		// ❌ НЕ отдаём expired продукты в основной список
		if status == "expired" {
			continue
		}

		totalPrice := s.calculateTotalPrice(item.Quantity, item.CurrentPricePerUnit)

		response := models.FridgeItemListResponse{
			ID:         item.ID,
			Name:       item.Ingredient.GetName(lang), // 🌍 Use localized name
			Category:   item.Ingredient.Category,      // Добавляем категорию для группировки
			Quantity:   item.Quantity,
			Unit:       item.Unit,
			Ingredient: models.NewIngredientBasicInfo(item.Ingredient), // 🌍 Full multilingual data
			ArrivedAt:  item.ArrivedAt,                                 // Дата поступления в холодильник
			ExpiresAt:  item.ExpiresAt,                                 // Срок годности
			DaysLeft:   daysLeft,
			Status:     status,
		}

		// Добавляем цену только если она есть (из кэша current_price_*)
		if item.CurrentPricePerUnit != nil {
			response.PricePerUnit = item.CurrentPricePerUnit // Цена за единицу
			response.TotalPrice = totalPrice                 // Общая стоимость
			response.Currency = item.CurrentPriceCurrency    // Валюта

			// SMART KITCHEN: Добавляем анализ динамики цены
			priceAnalysis, err := s.CalculatePriceTrend(item.ID)
			if err != nil {
				// Не критично, просто логируем и продолжаем без аналитики
				logger.Warn("failed to calculate price trend",
					zap.String("item_id", item.ID),
					zap.Error(err))
			} else if priceAnalysis != nil {
				response.PriceAnalysis = priceAnalysis
			}
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
	// 1. Проверяем, что продукт принадлежит пользователю И получаем данные
	item, err := s.fridgeRepo.GetByID(itemID)
	if err != nil {
		return fmt.Errorf("fridge item not found: %w", err)
	}

	if item.UserID != userID {
		return errors.New("access denied: item does not belong to user")
	}

	// 2. Удаляем продукт
	if err := s.fridgeRepo.Delete(itemID); err != nil {
		return err
	}

	// 3. Создаём уведомление ПОСЛЕ успешного удаления
	s.createItemDeletedNotification(userID, item)

	return nil
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
		status := models.GetFridgeItemStatus(daysLeft)

		// ❌ НЕ отдаём expired продукты
		if status == "expired" {
			continue
		}

		totalPrice := s.calculateTotalPrice(item.Quantity, item.CurrentPricePerUnit)

		response := models.FridgeItemListResponse{
			ID:        item.ID,
			Name:      item.Ingredient.Name,
			Category:  item.Ingredient.Category, // Добавляем категорию
			Quantity:  item.Quantity,
			Unit:      item.Unit,
			ArrivedAt: item.ArrivedAt, // Дата поступления
			DaysLeft:  daysLeft,
			Status:    status,
		}

		// Добавляем цену только если она есть (из кэша current_price_*)
		if item.CurrentPricePerUnit != nil {
			response.PricePerUnit = item.CurrentPricePerUnit // Цена за единицу
			response.TotalPrice = totalPrice                 // Общая стоимость
			response.Currency = item.CurrentPriceCurrency    // Валюта
		}

		result = append(result, response)
	}

	return result, nil
}

// cleanupExpiredItems автоматически удаляет просроченные продукты и создает события потерь
func (s *FridgeService) cleanupExpiredItems(userID string) error {
	// Находим все просроченные продукты пользователя
	var expiredItems []models.UserFridgeItem
	err := s.db.Where("user_id = ? AND expires_at IS NOT NULL AND expires_at < NOW()", userID).
		Preload("Ingredient").
		Find(&expiredItems).Error

	if err != nil {
		return fmt.Errorf("failed to find expired items: %w", err)
	}

	if len(expiredItems) == 0 {
		return nil // Нет просроченных продуктов
	}

	logger.Info("cleaning up expired items",
		zap.String("user_id", userID),
		zap.Int("count", len(expiredItems)))

	// Обрабатываем каждый просроченный продукт
	for _, item := range expiredItems {
		// Рассчитываем стоимость потери
		cost := 0.0
		pricePerUnit := 0.0
		currency := item.CurrentPriceCurrency // Уже string, не pointer

		if item.CurrentPricePerUnit != nil {
			cost = item.Quantity * (*item.CurrentPricePerUnit)
			pricePerUnit = *item.CurrentPricePerUnit
		}

		// Вычисляем сколько дней продукт пролежал в холодильнике
		daysInFridge := int(time.Since(item.ArrivedAt).Hours() / 24)

		// Форматируем даты в ISO 8601
		expiryDateStr := ""
		if item.ExpiresAt != nil {
			expiryDateStr = item.ExpiresAt.Format(time.RFC3339)
		}

		// Создаем событие потери в истории
		metadata := models.ExpiredItemMetadata{
			IngredientID:   item.IngredientID,
			IngredientName: item.Ingredient.Name,
			Quantity:       item.Quantity,
			Unit:           item.Unit,
			Cost:           cost,
			PricePerUnit:   pricePerUnit,
			Currency:       currency,
			ExpiryDate:     expiryDateStr,
			ArrivedAt:      item.ArrivedAt.Format(time.RFC3339),
			DaysInFridge:   daysInFridge,
			Reason:         "expiry_date_passed",
			Context:        "auto_cleanup_on_list",
		}

		// Сериализуем metadata в JSON
		metadataJSON, err := json.Marshal(metadata)
		if err != nil {
			logger.Error("failed to marshal metadata",
				zap.String("item_id", item.ID),
				zap.Error(err))
			continue
		}

		historyEvent := models.HistoryEvent{
			ID:         uuid.NewString(), // Generate UUID for id field
			UserID:     userID,
			EventType:  models.EventTypeExpired,
			SourceType: models.SourceTypeAuto,
			SourceID:   &item.ID,
			Metadata:   metadataJSON,
		}

		if err := s.db.Create(&historyEvent).Error; err != nil {
			logger.Error("failed to create expired event",
				zap.String("item_id", item.ID),
				zap.Error(err))
			continue // Продолжаем обработку остальных
		}

		// Удаляем просроченный продукт из холодильника
		if err := s.db.Delete(&item).Error; err != nil {
			logger.Error("failed to delete expired item",
				zap.String("item_id", item.ID),
				zap.Error(err))
			continue
		}

		logger.Info("expired item removed",
			zap.String("item_id", item.ID),
			zap.String("name", item.Ingredient.Name),
			zap.Float64("cost", cost))
	}

	return nil
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
func (s *FridgeService) calculateDaysLeft(expiresAt *time.Time) *int {
	if expiresAt == nil {
		return nil // Нет срока годности
	}
	duration := time.Until(*expiresAt)
	days := int(duration.Hours() / 24)
	return &days
}

// normalizePrice нормализует цену к единице измерения в БД (всегда g/ml/pcs)
// Цена ВСЕГДА вводится за крупные единицы: kg, l, pcs (НЕ за g/ml!)
func (s *FridgeService) normalizePrice(value float64, per string, unit string) (float64, error) {
	switch per {
	case "kg":
		if unit != "g" {
			return 0, fmt.Errorf("unit mismatch: cannot convert price per kg to unit %s", unit)
		}
		// 3.20 PLN / kg → 0.0032 PLN / g
		return value / 1000, nil

	case "g":
		// TEMPORARY: Support "g" unit until frontend is updated
		// TODO: Frontend should send "per": "kg" with value/1000
		if unit != "g" {
			return 0, fmt.Errorf("unit mismatch: price per g requires unit g, got %s", unit)
		}
		return value, nil

	case "l":
		if unit != "ml" {
			return 0, fmt.Errorf("unit mismatch: cannot convert price per l to unit %s", unit)
		}
		// 2.50 PLN / l → 0.0025 PLN / ml
		return value / 1000, nil

	case "ml":
		// TEMPORARY: Support "ml" unit until frontend is updated
		// TODO: Frontend should send "per": "l" with value/1000
		if unit != "ml" {
			return 0, fmt.Errorf("unit mismatch: price per ml requires unit ml, got %s", unit)
		}
		return value, nil

	case "pcs", "szt", "шт":
		if unit != "pcs" && unit != "szt" {
			return 0, fmt.Errorf("unit mismatch: price per pcs requires unit pcs, got %s", unit)
		}
		// 1.00 PLN / pcs → 1.00 PLN / pcs (без изменений)
		return value, nil

	default:
		return 0, fmt.Errorf("invalid price unit: %s (must be: kg/g for grams, l/ml for ml, pcs/szt for pieces)", per)
	}
}

// round2 округляет число до 2 знаков после запятой (для денег)
// Пример: 10.350000000000001 → 10.35
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// calculateTotalPrice вычисляет общую стоимость продукта
// Всегда округляет до 2 знаков после запятой (правило денег)
func (s *FridgeService) calculateTotalPrice(quantity float64, pricePerUnit *float64) *float64 {
	if pricePerUnit == nil {
		return nil
	}
	// Вычисляем и сразу округляем до 2 знаков
	total := round2(quantity * (*pricePerUnit))
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
		// Маппируем ошибку БД на доменную ошибку
		if err.Error() == "record not found" {
			return ErrNotFound // 404
		}
		return fmt.Errorf("failed to get fridge item: %w", err)
	}

	if item.UserID != userID {
		return ErrForbidden // 403
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
		return fmt.Errorf("%w: %s (allowed: manual, receipt, estimate, market, ai)",
			ErrInvalidSource, req.Source) // 400
	}

	// 3. Добавляем событие в историю (обернуто в транзакцию)
	// Используем транзакцию для атомарности: history INSERT + cache UPDATE
	if err := s.fridgeRepo.InsertPriceHistory(itemID, req.PricePerUnit, req.UnitForPrice, req.Currency, req.Source); err != nil {
		// Детальное логирование для отладки
		return fmt.Errorf("failed to insert price history (itemID=%s, price=%.8f, unit=%s, currency=%s, source=%s): %w",
			itemID, req.PricePerUnit, req.UnitForPrice, req.Currency, req.Source, err)
	}

	// 4. Обновляем кэш текущей цены (денормализация для производительности)
	// Защита от NULL: если price был NULL, теперь устанавливаем значение
	if err := s.fridgeRepo.UpdateCurrentPrice(itemID, req.PricePerUnit, req.Currency); err != nil {
		return fmt.Errorf("failed to update current price cache: %w", err)
	}

	return nil
}

// GetPriceHistory возвращает историю изменения цен
func (s *FridgeService) GetPriceHistory(userID string, itemID string) ([]models.PriceHistoryResponse, error) {
	// 1. Проверяем доступ
	item, err := s.fridgeRepo.GetByID(itemID)
	if err != nil {
		if err.Error() == "record not found" {
			return nil, ErrNotFound // 404
		}
		return nil, fmt.Errorf("failed to get fridge item: %w", err)
	}

	if item.UserID != userID {
		return nil, ErrForbidden // 403
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

// CalculatePriceTrend анализирует динамику изменения цены для "умной кухни"
// Возвращает nil если недостаточно данных (< 2 записей в истории)
func (s *FridgeService) CalculatePriceTrend(itemID string) (*models.PriceAnalysis, error) {
	// Получаем историю цен (отсортирована по created_at DESC)
	history, err := s.fridgeRepo.GetPriceHistory(itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get price history: %w", err)
	}

	// Нужно минимум 2 записи для анализа тренда
	if len(history) < 2 {
		return nil, nil // Не ошибка, просто недостаточно данных
	}

	// Берём последние 2 записи
	last := history[0]     // Самая свежая
	previous := history[1] // Предыдущая

	// Считаем процент изменения: ((last - previous) / previous) * 100
	percentChange := ((last.PricePerUnit - previous.PricePerUnit) / previous.PricePerUnit) * 100

	// Определяем тренд
	var trend string
	const stableThreshold = 5.0 // ±5% считается стабильной ценой
	switch {
	case percentChange > stableThreshold:
		trend = "up"
	case percentChange < -stableThreshold:
		trend = "down"
	default:
		trend = "stable"
	}

	return &models.PriceAnalysis{
		Trend:         trend,
		PercentChange: round2(percentChange),
		LastPrice:     last.PricePerUnit,
		PreviousPrice: previous.PricePerUnit,
		LastUpdated:   last.CreatedAt,
		HistoryCount:  len(history),
	}, nil
}

// UpdateItemQuantity обновляет количество продукта в холодильнике
func (s *FridgeService) UpdateItemQuantity(userID string, itemID string, newQuantity float64) error {
	// 1. Проверяем, что продукт существует и принадлежит пользователю
	item, err := s.fridgeRepo.GetByID(itemID)
	if err != nil {
		if err.Error() == "record not found" {
			return ErrNotFound // 404
		}
		return fmt.Errorf("failed to get fridge item: %w", err)
	}

	if item.UserID != userID {
		return ErrForbidden // 403
	}

	// 2. Проверяем срок годности перед обновлением
	if item.ExpiresAt != nil && item.ExpiresAt.Before(time.Now()) {
		// Продукт просрочен - не обновляем, а удаляем с созданием события
		logger.Info("attempted to update expired item, removing instead",
			zap.String("item_id", itemID),
			zap.String("ingredient_name", item.Ingredient.Name))

		// Запускаем очистку просроченных продуктов для этого пользователя
		if err := s.cleanupExpiredItems(userID); err != nil {
			logger.Warn("cleanup failed during update", zap.Error(err))
		}

		return ErrNotFound // Продукт больше не существует
	}

	// 3. Обновляем количество
	item.Quantity = newQuantity
	if err := s.fridgeRepo.Update(item); err != nil {
		return fmt.Errorf("failed to update quantity: %w", err)
	}

	return nil
}

// AddMissingFromRecipe adds missing recipe ingredients to user's fridge
// Calculates diff: if ingredient exists but insufficient → add difference, if missing → add full amount
func (s *FridgeService) AddMissingFromRecipe(userID string, recipeID string) (*dto.AddMissingResult, error) {
	// 1. Load recipe ingredients
	var recipe models.RecipeCatalog
	err := s.db.
		Preload("Ingredients").
		Preload("Ingredients.Ingredient").
		Where("id::text = ?", recipeID).
		First(&recipe).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("recipe not found: %s", recipeID)
		}
		return nil, fmt.Errorf("failed to load recipe: %w", err)
	}

	// 2. Load user's fridge items
	var fridgeItems []models.UserFridgeItem
	err = s.db.
		Where("user_id::text = ?", userID).
		Preload("Ingredient").
		Find(&fridgeItems).Error
	if err != nil {
		return nil, fmt.Errorf("failed to load fridge items: %w", err)
	}

	// 3. Build fridge map: ingredientID → fridgeItem
	fridgeMap := make(map[string]*models.UserFridgeItem)
	for i := range fridgeItems {
		fridgeMap[fridgeItems[i].IngredientID] = &fridgeItems[i]
	}

	// 4. Calculate diff and add missing/insufficient ingredients
	result := &dto.AddMissingResult{
		Added:   0,
		Skipped: 0,
		Items:   []dto.AddedItem{},
	}

	for _, recipeIng := range recipe.Ingredients {
		requiredQty := recipeIng.Quantity
		requiredUnit := recipeIng.Unit

		fridgeItem, exists := fridgeMap[recipeIng.IngredientID]

		if exists && fridgeItem.Unit == requiredUnit && fridgeItem.Quantity >= requiredQty {
			// Already have enough
			result.Skipped++
			continue
		}

		// Calculate how much to add
		var addQty float64
		if exists && fridgeItem.Unit == requiredUnit {
			// Exists but insufficient → add difference
			addQty = requiredQty - fridgeItem.Quantity
			fridgeItem.Quantity = requiredQty
			if err := s.fridgeRepo.Update(fridgeItem); err != nil {
				logger.Log.Error("Failed to update fridge item",
					zap.Error(err),
					zap.String("itemId", fridgeItem.ID),
				)
				continue
			}
		} else {
			// Missing or different unit → add full amount
			addQty = requiredQty

			// Get ingredient details
			ingredient, err := s.ingredientRepo.GetIngredientByID(recipeIng.IngredientID)
			if err != nil {
				logger.Log.Error("Failed to get ingredient",
					zap.Error(err),
					zap.String("ingredientId", recipeIng.IngredientID),
				)
				continue
			}

			// Create new fridge item
			newItem := &models.UserFridgeItem{
				UserID:       userID,
				IngredientID: recipeIng.IngredientID,
				Quantity:     requiredQty,
				Unit:         requiredUnit,
				ArrivedAt:    time.Now(),
				ExpiresAt:    nil, // Will be calculated if ingredient has defaultShelfLifeDays
			}

			// Auto-calculate expiry if ingredient has shelf life
			if ingredient.DefaultShelfLifeDays != nil {
				expiresAt := newItem.ArrivedAt.AddDate(0, 0, *ingredient.DefaultShelfLifeDays)
				newItem.ExpiresAt = &expiresAt
			}

			if err := s.fridgeRepo.Create(newItem); err != nil {
				logger.Log.Error("Failed to create fridge item",
					zap.Error(err),
					zap.String("ingredientId", recipeIng.IngredientID),
				)
				continue
			}
		}

		// Add to result
		result.Added++
		result.Items = append(result.Items, dto.AddedItem{
			IngredientID:  recipeIng.IngredientID,
			Name:          recipeIng.Ingredient.Name,
			AddedQuantity: addQty,
			Unit:          requiredUnit,
		})
	}

	return result, nil
}

// createItemAddedNotification создаёт системное уведомление о добавлении продукта
func (s *FridgeService) createItemAddedNotification(userID string, item *models.UserFridgeItem, ingredient *models.Ingredient) {
	// Получаем название продукта (предпочтительно на польском)
	ingredientName := ingredient.Name
	if ingredient.NamePL != nil && *ingredient.NamePL != "" {
		ingredientName = *ingredient.NamePL
	}

	// Форматируем количество
	quantityStr := fmt.Sprintf("%.1f %s", item.Quantity, item.Unit)

	// Формируем сообщение
	message := fmt.Sprintf("%s добавлен в холодильник (%s)", ingredientName, quantityStr)

	// Формируем meta информацию
	metaJSON, _ := json.Marshal(map[string]interface{}{
		"fridgeItemId": item.ID,
		"ingredientId": item.IngredientID,
		"quantity":     item.Quantity,
		"unit":         item.Unit,
	})
	metaStr := string(metaJSON)

	// Создаём уведомление
	notification := &models.Notification{
		UserID:  userID,
		Type:    models.NotificationTypeFridge,
		Level:   models.NotificationLevelInfo,
		Title:   "Продукт добавлен в холодильник",
		Message: message,
		Meta:    &metaStr,
	}

	// Сохраняем в БД (не фейлим если не получилось)
	if err := s.notificationService.Create(notification); err != nil {
		logger.Warn("failed to create item added notification",
			zap.String("user_id", userID),
			zap.String("item_id", item.ID),
			zap.Error(err))
	} else {
		logger.Info("notification created",
			zap.String("user_id", userID),
			zap.String("item_name", ingredientName))
	}
}

// createItemDeletedNotification создаёт уведомление об удалении продукта
func (s *FridgeService) createItemDeletedNotification(userID string, item *models.UserFridgeItem) {
	if item.Ingredient == nil {
		logger.Warn("cannot create delete notification: ingredient data missing",
			zap.String("item_id", item.ID),
			zap.String("user_id", userID))
		return
	}

	// Получаем польское название если доступно
	ingredientName := item.Ingredient.Name
	if item.Ingredient.NamePL != nil && *item.Ingredient.NamePL != "" {
		ingredientName = *item.Ingredient.NamePL
	}

	// Формат: "Czosnek удалён из холодильника (3.5 g)"
	message := fmt.Sprintf("%s удалён из холодильника (%.1f %s)",
		ingredientName,
		item.Quantity,
		item.Unit,
	)

	// Meta данные для уведомления
	meta := map[string]interface{}{
		"fridgeItemId": item.ID,
		"ingredientId": item.IngredientID,
		"quantity":     item.Quantity,
		"unit":         item.Unit,
		"action":       "deleted",
	}

	metaBytes, err := json.Marshal(meta)
	if err != nil {
		logger.Warn("failed to marshal notification meta",
			zap.String("item_id", item.ID),
			zap.Error(err))
		return
	}
	metaStr := string(metaBytes)

	// Создаём уведомление
	notification := &models.Notification{
		UserID:  userID,
		Type:    models.NotificationTypeFridge,
		Level:   models.NotificationLevelInfo,
		Title:   "Продукт удалён из холодильника",
		Message: message,
		Meta:    &metaStr,
	}

	// Не блокируем удаление при ошибке создания уведомления
	if err := s.notificationService.Create(notification); err != nil {
		logger.Warn("failed to create item deleted notification",
			zap.String("user_id", userID),
			zap.String("item_id", item.ID),
			zap.Error(err))
	} else {
		logger.Info("delete notification created",
			zap.String("user_id", userID),
			zap.String("item_name", ingredientName))
	}
}

// computeTotalCost вычисляет общую стоимость продукта в холодильнике
func (s *FridgeService) computeTotalCost(
	quantity float64,
	unit string,
	pricePerUnit float64,
	unitForPrice string,
) *models.ComputedPrice {

	// Нормализуем quantity к базовым единицам (g, ml, pcs)
	var quantityInBaseUnits float64
	switch unit {
	case "kg":
		quantityInBaseUnits = quantity * 1000 // kg → g
	case "g":
		quantityInBaseUnits = quantity
	case "l":
		quantityInBaseUnits = quantity * 1000 // l → ml
	case "ml":
		quantityInBaseUnits = quantity
	case "pcs":
		quantityInBaseUnits = quantity
	default:
		return nil // Неизвестная единица
	}

	// Нормализуем цену к базовым единицам
	var pricePerBaseUnit float64
	switch unitForPrice {
	case "kg":
		pricePerBaseUnit = pricePerUnit / 1000 // zł/kg → zł/g
	case "g":
		pricePerBaseUnit = pricePerUnit
	case "l":
		pricePerBaseUnit = pricePerUnit / 1000 // zł/l → zł/ml
	case "ml":
		pricePerBaseUnit = pricePerUnit
	case "pcs":
		pricePerBaseUnit = pricePerUnit // zł/pcs
	default:
		return nil
	}

	// Вычисляем общую стоимость
	totalCost := quantityInBaseUnits * pricePerBaseUnit

	// Округляем до 2 знаков после запятой
	totalCost = math.Round(totalCost*100) / 100
	pricePerBaseUnit = math.Round(pricePerBaseUnit*100000) / 100000 // Сохраняем точность для unitPrice

	return &models.ComputedPrice{
		UnitPrice: pricePerBaseUnit, // Цена за 1g/ml/pcs
		TotalCost: totalCost,         // Общая стоимость (округленная)
	}
}

// GetUserItemsV2 возвращает продукты в холодильнике с ценами (новая версия API)
// lang - язык для локализации имён (pl, en, ru)
func (s *FridgeService) GetUserItemsV2(userID string, lang string) ([]models.FridgeItemResponseV2, error) {
	// Автоматически очищаем просроченные продукты
	if err := s.cleanupExpiredItems(userID); err != nil {
		logger.Warn("failed to cleanup expired items",
			zap.String("user_id", userID),
			zap.Error(err))
	}

	// Загружаем items с Preload ingredient
	items, err := s.fridgeRepo.GetUserFridgeItems(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get fridge items: %w", err)
	}

	// Fallback to pl if lang is empty
	if lang == "" {
		lang = "pl"
	}

	logger.Info("🌍 GetUserItemsV2 called",
		zap.String("user_id", userID),
		zap.String("requested_lang", lang))

	result := make([]models.FridgeItemResponseV2, 0, len(items))
	for _, item := range items {
		if item.Ingredient == nil {
			logger.Warn("item has no ingredient",
				zap.String("item_id", item.ID))
			continue
		}

		// 🔍 DEBUG: Проверяем наличие переводов
		if len(result) == 0 { // только для первого элемента
			hasRU := item.Ingredient.NameRU != nil && *item.Ingredient.NameRU != ""
			hasPL := item.Ingredient.NamePL != nil && *item.Ingredient.NamePL != ""
			hasEN := item.Ingredient.NameEN != nil && *item.Ingredient.NameEN != ""
			
			logger.Info("🔍 INGREDIENT TRANSLATIONS CHECK",
				zap.String("ingredient_id", item.Ingredient.ID),
				zap.String("legacy_name", item.Ingredient.Name),
				zap.Bool("has_name_ru", hasRU),
				zap.Bool("has_name_pl", hasPL),
				zap.Bool("has_name_en", hasEN))
			
			if hasRU && item.Ingredient.NameRU != nil {
				logger.Info("🔍 RU name value", zap.String("name_ru", *item.Ingredient.NameRU))
			}
		}

		daysLeft := s.calculateDaysLeft(item.ExpiresAt)
		status := models.GetFridgeItemStatus(daysLeft)

		// Не отдаём expired продукты
		if status == "expired" {
			continue
		}

		// ✅ ПРАВИЛО: categoryKey ТОЛЬКО из Ingredient.Category (stable key)
		categoryKey := item.Ingredient.Category
		if categoryKey == "" {
			categoryKey = "other" // fallback только если пусто
		}

		// Локализованное имя
		localizedName := item.Ingredient.GetName(lang)

		response := models.FridgeItemResponseV2{
			ID:   item.ID,
			Name: localizedName, // ✅ ОБРАТНАЯ СОВМЕСТИМОСТЬ: для старого фронта
			Ingredient: models.IngredientInfo{
				ID:   item.Ingredient.ID,
				Name: localizedName, // ✅ Локализация имени
				Unit: item.Ingredient.Unit,
			},
			CategoryKey: categoryKey, // ✅ STABLE KEY (не зависит от языка!)
			Quantity:    item.Quantity,
			Unit:        item.Unit,
			ExpiresAt:   item.ExpiresAt,
			DaysLeft:    daysLeft,
		}

		// ✅ ЦЕНА: всегда включаем если есть
		if item.CurrentPricePerUnit != nil && *item.CurrentPricePerUnit > 0 {
			response.CurrentPrice = &models.CurrentPriceInfo{
				Value:     *item.CurrentPricePerUnit,
				Per:       item.Unit, // цена за текущую единицу измерения
				Currency:  item.CurrentPriceCurrency,
				UpdatedAt: item.PriceUpdatedAt,
			}

			// Обратная совместимость: старые поля
			response.Price = &models.PriceInfo{
				Value: *item.CurrentPricePerUnit,
				Per:   item.Unit,
			}

			// Вычисляем стоимость
			computed := s.computeTotalCost(
				item.Quantity,
				item.Unit,
				*item.CurrentPricePerUnit,
				item.Unit,
			)
			if computed != nil {
				response.Computed = computed
			}
		}

		result = append(result, response)
	}

	return result, nil
}
