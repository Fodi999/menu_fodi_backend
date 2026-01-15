package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"gorm.io/gorm"
)

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

// GetNotificationLevel определяет уровень уведомления по количеству дней
func GetNotificationLevel(daysLeft int) models.NotificationLevel {
	switch {
	case daysLeft < 0:
		return models.NotificationLevelCritical // Просрочено
	case daysLeft == 0:
		return models.NotificationLevelCritical // Истекает сегодня
	case daysLeft == 1:
		return models.NotificationLevelWarning // Завтра истекает
	case daysLeft >= 2 && daysLeft <= 3:
		return models.NotificationLevelInfo // Скоро истечет
	default:
		return "" // Не требует уведомления
	}
}

// CreateExpiryNotification создает уведомление о истечении срока
func CreateExpiryNotification(db *gorm.DB, item *models.FridgeItem, daysLeft int) error {
	level := GetNotificationLevel(daysLeft)
	if level == "" {
		return nil // Не требует уведомления
	}

	// Проверяем, не создано ли уже уведомление сегодня
	today := time.Now().Truncate(24 * time.Hour)
	var existingCount int64
	
	metaJSON, _ := json.Marshal(map[string]interface{}{
		"fridgeItemId": item.ID,
		"daysLeft":     daysLeft,
	})
	metaStr := string(metaJSON)

	err := db.Model(&models.Notification{}).
		Where("user_id = ? AND type IN (?, ?) AND created_at >= ?", 
			item.UserID, models.NotificationTypeFridge, models.NotificationTypeAI, today).
		Where("meta->>'fridgeItemId' = ?", item.ID).
		Count(&existingCount).Error

	if err != nil {
		return fmt.Errorf("failed to check existing notifications: %w", err)
	}

	if existingCount > 0 {
		fmt.Printf("ℹ️  Notification already exists today for fridge item %s\n", item.ID)
		return nil
	}

	// Формируем уведомление в зависимости от уровня
	var notification models.Notification
	
	// Получаем название продукта
	var ingredient models.Ingredient
	if err := db.First(&ingredient, "id = ?", item.IngredientID).Error; err != nil {
		return fmt.Errorf("failed to fetch ingredient: %w", err)
	}

	ingredientName := ingredient.Name
	if ingredient.NamePL != nil && *ingredient.NamePL != "" {
		ingredientName = *ingredient.NamePL
	}

	switch level {
	case models.NotificationLevelCritical:
		if daysLeft < 0 {
			notification = models.Notification{
				UserID:  item.UserID,
				Type:    models.NotificationTypeFridge,
				Level:   level,
				Title:   "Продукт просрочен",
				Message: fmt.Sprintf("%s просрочен. Потеря: %.2f PLN", ingredientName, item.PriceTotal),
				Meta:    &metaStr,
			}
		} else {
			notification = models.Notification{
				UserID:  item.UserID,
				Type:    models.NotificationTypeFridge,
				Level:   level,
				Title:   "Продукт истекает сегодня",
				Message: fmt.Sprintf("%s нужно использовать сегодня! Ценность: %.2f PLN", ingredientName, item.PriceTotal),
				Meta:    &metaStr,
			}
		}

	case models.NotificationLevelWarning:
		notification = models.Notification{
			UserID:  item.UserID,
			Type:    models.NotificationTypeAI,
			Level:   level,
			Title:   "Используй продукт завтра",
			Message: fmt.Sprintf("У тебя есть %s (%.1f %s). Используй завтра, иначе потеряешь %.2f PLN", 
				ingredientName, item.Quantity, item.Unit, item.PriceTotal),
			Meta:    &metaStr,
		}

	case models.NotificationLevelInfo:
		notification = models.Notification{
			UserID:  item.UserID,
			Type:    models.NotificationTypeAI,
			Level:   level,
			Title:   "Скоро истечет срок",
			Message: fmt.Sprintf("%s истечет через %d дня. Не забудь использовать!", ingredientName, daysLeft),
			Meta:    &metaStr,
		}
	}

	if err := db.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	fmt.Printf("✅ Created %s notification for %s (days left: %d)\n", level, ingredientName, daysLeft)
	return nil
}

// CheckAndNotifyExpiringItems проверяет все продукты пользователя и создает уведомления
// Используется в CRON или при GET /api/fridge/items (1 раз в день)
func CheckAndNotifyExpiringItems(db *gorm.DB, userID string) error {
	var items []models.FridgeItem
	
	// Получаем только fresh items с expiration date
	err := db.Where("user_id = ? AND status = ? AND expires_at IS NOT NULL", 
		userID, models.FridgeItemStatusFresh).
		Find(&items).Error

	if err != nil {
		return fmt.Errorf("failed to fetch fridge items: %w", err)
	}

	fmt.Printf("🔍 Checking %d fridge items for user %s\n", len(items), userID)

	for _, item := range items {
		result := EvaluateFridgeItem(&item)
		
		// Обновляем статус если просрочен
		if result.Status == models.FridgeItemStatusExpired && item.Status != models.FridgeItemStatusExpired {
			db.Model(&item).Updates(map[string]interface{}{
				"status":    result.Status,
				"days_left": result.DaysLeft,
			})
		}

		// Создаем уведомление если нужно
		if result.DaysLeft != nil {
			if err := CreateExpiryNotification(db, &item, *result.DaysLeft); err != nil {
				fmt.Printf("❌ Failed to create notification for item %s: %v\n", item.ID, err)
			}
		}
	}

	return nil
}
