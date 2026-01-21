package service

import (
	"fmt"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	notificationService "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/notifications/service"
	"gorm.io/gorm"
)

// ============================================================================
// EXPIRY CHECKER - ПРАВИЛЬНАЯ АРХИТЕКТУРА
// ============================================================================
// Цель: проверка продуктов и создание ТОЛЬКО expiry notifications
// Запускается: CRON (1× в день, например 08:00)
//
// Flow:
//   daysLeft = 2  → INFO (summary only)
//   daysLeft = 1  → WARNING (⚠️ скоро испортится)
//   daysLeft <= 0 → CRITICAL (⛔ срочно использовать)
//   daysLeft < -3 → cleanup (удаляем + history event)
//
// ============================================================================

// FridgeEvaluationResult результат оценки продукта
type FridgeEvaluationResult struct {
	Status   models.FridgeItemStatus
	DaysLeft *int
}

// EvaluateFridgeItem вычисляет статус и дни до истечения срока
// ❗ ЦЕНТРАЛЬНАЯ ФУНКЦИЯ - вся логика определения срока здесь
func EvaluateFridgeItem(item *models.FridgeItem) FridgeEvaluationResult {
	if item.ExpiresAt == nil {
		return FridgeEvaluationResult{
			Status:   models.FridgeItemStatusFresh,
			DaysLeft: nil,
		}
	}

	// Нормализуем даты до начала дня (без времени)
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	expiry := time.Date(item.ExpiresAt.Year(), item.ExpiresAt.Month(), item.ExpiresAt.Day(), 0, 0, 0, 0, item.ExpiresAt.Location())

	// Вычисляем разницу в днях
	daysLeft := int(expiry.Sub(today).Hours() / 24)

	var status models.FridgeItemStatus
	if daysLeft < 0 {
		status = models.FridgeItemStatusExpired
	} else {
		status = models.FridgeItemStatusFresh
	}

	return FridgeEvaluationResult{
		Status:   status,
		DaysLeft: &daysLeft,
	}
}

// ============================================================================
// NOTIFICATION CREATION (использует НОВЫЙ сервис)
// ============================================================================

// CheckAndNotifyExpiringItems проверяет все продукты и создает уведомления
// ✅ ИСПОЛЬЗУЕТСЯ: CRON job (1× в день)
// ❌ НЕ ИСПОЛЬЗУЕТСЯ: GET endpoints (они мутируют состояние)
func CheckAndNotifyExpiringItems(db *gorm.DB, userID string) error {
	// Инициализируем notification service
	notifSvc := notificationService.NewNotificationService(db)

	var items []models.FridgeItem

	// Получаем fresh items с expiration date
	err := db.Preload("Ingredient").
		Where("user_id = ? AND status = ? AND expires_at IS NOT NULL",
			userID, models.FridgeItemStatusFresh).
		Find(&items).Error

	if err != nil {
		return fmt.Errorf("failed to fetch fridge items: %w", err)
	}

	if len(items) == 0 {
		return nil // Нет продуктов для проверки
	}

	fmt.Printf("🔍 [EXPIRY CHECK] Checking %d items for user %s\n", len(items), userID)

	// Группируем по уровню важности для summary
	var criticalItems []models.FridgeItem
	var warningItems []models.FridgeItem
	var infoItems []models.FridgeItem

	for _, item := range items {
		if item.Ingredient == nil {
			continue // Skip items without ingredient data
		}

		result := EvaluateFridgeItem(&item)
		if result.DaysLeft == nil {
			continue // No expiration date
		}

		daysLeft := *result.DaysLeft

		// Обновляем статус если просрочен
		if result.Status == models.FridgeItemStatusExpired && item.Status != models.FridgeItemStatusExpired {
			db.Model(&item).Updates(map[string]interface{}{
				"status":    result.Status,
				"days_left": result.DaysLeft,
			})
		}

		// Группируем по level
		switch {
		case daysLeft <= 0:
			criticalItems = append(criticalItems, item)
		case daysLeft == 1:
			warningItems = append(warningItems, item)
		case daysLeft == 2:
			infoItems = append(infoItems, item)
		}
	}

	// ✅ СОЗДАЁМ УВЕДОМЛЕНИЯ через новый сервис
	if len(criticalItems) > 0 {
		if err := createCriticalNotifications(notifSvc, userID, criticalItems); err != nil {
			fmt.Printf("❌ Failed to create critical notifications: %v\n", err)
		}
	}

	if len(warningItems) > 0 {
		if err := createWarningNotifications(notifSvc, userID, warningItems); err != nil {
			fmt.Printf("❌ Failed to create warning notifications: %v\n", err)
		}
	}

	if len(infoItems) > 0 {
		// INFO - только summary, не спамим отдельными
		fmt.Printf("ℹ️  %d products expiring in 2 days (info level - summary only)\n", len(infoItems))
	}

	return nil
}

// createCriticalNotifications создаёт CRITICAL уведомления (daysLeft <= 0)
func createCriticalNotifications(svc notificationService.NotificationService, userID string, items []models.FridgeItem) error {
	for _, item := range items {
		result := EvaluateFridgeItem(&item)
		if result.DaysLeft == nil {
			continue
		}

		// Получаем название
		ingredientName := item.Ingredient.Name
		if item.Ingredient.NamePL != nil && *item.Ingredient.NamePL != "" {
			ingredientName = *item.Ingredient.NamePL
		}

		meta := models.FridgeNotificationMeta{
			FridgeItemID:   item.ID,
			IngredientID:   item.IngredientID,
			IngredientName: ingredientName,
			DaysLeft:       *result.DaysLeft,
			ExpiresAt:      item.ExpiresAt.Format(time.RFC3339),
			Quantity:       item.Quantity,
			Unit:           item.Unit,
			CategoryKey:    item.Ingredient.Category,
		}

		// Создаём через новый сервис (автоматическая unique_key защита)
		if err := svc.CreateExpiryNotification(userID, models.NotificationLevelCritical, meta); err != nil {
			fmt.Printf("❌ Failed to create critical notification for %s: %v\n", item.ID, err)
			continue
		}

		fmt.Printf("✅ [CRITICAL] Created notification for %s (days: %d)\n", ingredientName, *result.DaysLeft)
	}

	return nil
}

// createWarningNotifications создаёт WARNING уведомления (daysLeft = 1)
func createWarningNotifications(svc notificationService.NotificationService, userID string, items []models.FridgeItem) error {
	for _, item := range items {
		result := EvaluateFridgeItem(&item)
		if result.DaysLeft == nil {
			continue
		}

		// Получаем название
		ingredientName := item.Ingredient.Name
		if item.Ingredient.NamePL != nil && *item.Ingredient.NamePL != "" {
			ingredientName = *item.Ingredient.NamePL
		}

		meta := models.FridgeNotificationMeta{
			FridgeItemID:   item.ID,
			IngredientID:   item.IngredientID,
			IngredientName: ingredientName,
			DaysLeft:       *result.DaysLeft,
			ExpiresAt:      item.ExpiresAt.Format(time.RFC3339),
			Quantity:       item.Quantity,
			Unit:           item.Unit,
			CategoryKey:    item.Ingredient.Category,
		}

		if err := svc.CreateExpiryNotification(userID, models.NotificationLevelWarning, meta); err != nil {
			fmt.Printf("❌ Failed to create warning notification for %s: %v\n", item.ID, err)
			continue
		}

		fmt.Printf("✅ [WARNING] Created notification for %s (days: 1)\n", ingredientName)
	}

	return nil
}

// ============================================================================
// LEGACY SUPPORT (для обратной совместимости)
// ============================================================================

// CreateExpiryNotification - DEPRECATED: использует старую модель
// TODO: Удалить после миграции на новую систему
func CreateExpiryNotification(db *gorm.DB, item *models.FridgeItem, daysLeft int) error {
	// Redirect to new system
	svc := notificationService.NewNotificationService(db)

	if item.Ingredient == nil {
		return fmt.Errorf("ingredient data missing")
	}

	ingredientName := item.Ingredient.Name
	if item.Ingredient.NamePL != nil && *item.Ingredient.NamePL != "" {
		ingredientName = *item.Ingredient.NamePL
	}

	meta := models.FridgeNotificationMeta{
		FridgeItemID:   item.ID,
		IngredientID:   item.IngredientID,
		IngredientName: ingredientName,
		DaysLeft:       daysLeft,
		ExpiresAt:      item.ExpiresAt.Format(time.RFC3339),
		Quantity:       item.Quantity,
		Unit:           item.Unit,
		CategoryKey:    item.Ingredient.Category,
	}

	var level models.NotificationLevel
	switch {
	case daysLeft <= 0:
		level = models.NotificationLevelCritical
	case daysLeft == 1:
		level = models.NotificationLevelWarning
	case daysLeft == 2:
		level = models.NotificationLevelInfo
	default:
		return nil // No notification needed
	}

	return svc.CreateExpiryNotification(item.UserID, level, meta)
}

// GetNotificationLevel - DEPRECATED
// TODO: Remove after migration
func GetNotificationLevel(daysLeft int) models.NotificationLevel {
	switch {
	case daysLeft <= 0:
		return models.NotificationLevelCritical
	case daysLeft == 1:
		return models.NotificationLevelWarning
	case daysLeft == 2:
		return models.NotificationLevelInfo
	default:
		return ""
	}
}
