package service

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"gorm.io/gorm"
)

// NotificationService сервис для работы с уведомлениями
// ✅ ТОЛЬКО для событий, требующих внимания (expiry tracking)
// ❌ НЕ для activity logs (добавлен/удалён продукт)
type NotificationService interface {
	// Получение уведомлений
	GetNotificationsByLevel(userID string) (*models.NotificationGroup, error)
	GetUnreadCount(userID string) (*models.UnreadCount, error)
	
	// Действия с уведомлениями
	MarkAsRead(notificationID string, userID string) error
	MarkAllAsRead(userID string) error
	ResolveNotification(notificationID string, userID string) error
	
	// Создание уведомлений (для internal use)
	CreateExpiryNotification(userID string, level models.NotificationLevel, meta models.FridgeNotificationMeta) error
	CreateAggregatedNotification(userID string, level models.NotificationLevel, meta models.AggregatedFridgeMeta) error
	
	// Очистка устаревших
	CleanupExpiredNotifications() error
}

type notificationService struct {
	db *gorm.DB
}

func NewNotificationService(db *gorm.DB) NotificationService {
	return &notificationService{db: db}
}

// ============================================================================
// ПОЛУЧЕНИЕ УВЕДОМЛЕНИЙ
// ============================================================================

// GetNotificationsByLevel получить уведомления сгруппированные по уровням
// ✅ Возвращает только ACTIVE и UNREAD уведомления
func (s *notificationService) GetNotificationsByLevel(userID string) (*models.NotificationGroup, error) {
	var all []models.Notification

	// Получаем только активные непрочитанные уведомления типа fridge
	err := s.db.Where("user_id = ? AND type = ? AND status = ? AND read_at IS NULL",
		userID,
		models.NotificationTypeFridge,
		models.NotificationStatusActive,
	).Order("created_at DESC").
		Limit(100).
		Find(&all).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch notifications: %w", err)
	}

	// Группируем по уровню важности
	group := &models.NotificationGroup{
		Critical: []models.Notification{},
		Warning:  []models.Notification{},
		Info:     []models.Notification{},
	}

	for _, n := range all {
		switch n.Level {
		case models.NotificationLevelCritical:
			group.Critical = append(group.Critical, n)
		case models.NotificationLevelWarning:
			group.Warning = append(group.Warning, n)
		case models.NotificationLevelInfo:
			group.Info = append(group.Info, n)
		}
	}

	return group, nil
}

// GetUnreadCount получить количество непрочитанных по уровням
// ❗ INFO не считается в badge (только для display)
func (s *notificationService) GetUnreadCount(userID string) (*models.UnreadCount, error) {
	var critical, warning, info int64

	baseQuery := s.db.Model(&models.Notification{}).
		Where("user_id = ? AND type = ? AND status = ? AND read_at IS NULL",
			userID,
			models.NotificationTypeFridge,
			models.NotificationStatusActive,
		)

	// Считаем по каждому уровню
	if err := baseQuery.Where("level = ?", models.NotificationLevelCritical).Count(&critical).Error; err != nil {
		return nil, fmt.Errorf("failed to count critical: %w", err)
	}

	if err := baseQuery.Where("level = ?", models.NotificationLevelWarning).Count(&warning).Error; err != nil {
		return nil, fmt.Errorf("failed to count warning: %w", err)
	}

	if err := baseQuery.Where("level = ?", models.NotificationLevelInfo).Count(&info).Error; err != nil {
		return nil, fmt.Errorf("failed to count info: %w", err)
	}

	return &models.UnreadCount{
		Critical: int(critical),
		Warning:  int(warning),
		Info:     int(info),
		Total:    int(critical + warning), // ❗ Info НЕ считается
	}, nil
}

// ============================================================================
// ДЕЙСТВИЯ С УВЕДОМЛЕНИЯМИ
// ============================================================================

// MarkAsRead пометить уведомление как прочитанное
func (s *notificationService) MarkAsRead(notificationID string, userID string) error {
	now := time.Now()
	result := s.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("read_at", now)

	if result.Error != nil {
		return fmt.Errorf("failed to mark notification as read: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

// MarkAllAsRead пометить все уведомления как прочитанные
func (s *notificationService) MarkAllAsRead(userID string) error {
	now := time.Now()
	err := s.db.Model(&models.Notification{}).
		Where("user_id = ? AND read_at IS NULL AND status = ?",
			userID,
			models.NotificationStatusActive,
		).
		Update("read_at", now).Error

	if err != nil {
		return fmt.Errorf("failed to mark all as read: %w", err)
	}

	return nil
}

// ResolveNotification пометить уведомление как решённое
// Используется когда пользователь использовал продукт или выбросил
func (s *notificationService) ResolveNotification(notificationID string, userID string) error {
	result := s.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"status":  models.NotificationStatusResolved,
			"read_at": time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to resolve notification: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

// ============================================================================
// СОЗДАНИЕ УВЕДОМЛЕНИЙ
// ============================================================================

// CreateExpiryNotification создать уведомление об истечении срока годности
// 🔒 Гарантирует уникальность: одно уведомление на продукт в день
func (s *notificationService) CreateExpiryNotification(
	userID string,
	level models.NotificationLevel,
	meta models.FridgeNotificationMeta,
) error {
	// Генерируем unique key для предотвращения дублей
	uniqueKey := generateUniqueKey(userID, level, meta.FridgeItemID, time.Now())

	// Проверяем, существует ли уже такое уведомление сегодня
	var existing models.Notification
	err := s.db.Where("user_id = ? AND unique_key = ? AND status = ?",
		userID,
		uniqueKey,
		models.NotificationStatusActive,
	).First(&existing).Error

	if err == nil {
		// Уведомление уже существует - не создаём дубликат
		return nil
	}

	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check existing notification: %w", err)
	}

	// Формируем title и message в зависимости от уровня
	title, message := buildNotificationText(level, meta)

	// Конвертируем meta в JSON
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}
	metaStr := string(metaJSON)

	// Создаём уведомление
	notification := &models.Notification{
		UserID:    userID,
		Type:      models.NotificationTypeFridge,
		Level:     level,
		Title:     title,
		Message:   message,
		Meta:      &metaStr,
		UniqueKey: &uniqueKey,
		Status:    models.NotificationStatusActive,
	}

	if err := s.db.Create(notification).Error; err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

// CreateAggregatedNotification создать агрегированное уведомление
// Используется для summary: "3 продукта скоро испортятся"
func (s *notificationService) CreateAggregatedNotification(
	userID string,
	level models.NotificationLevel,
	meta models.AggregatedFridgeMeta,
) error {
	// Для агрегированных уведомлений используем дату как часть unique key
	date := time.Now().Format("2006-01-02")
	uniqueKey := fmt.Sprintf("%x", md5.Sum([]byte(
		fmt.Sprintf("%s-aggregated-%s-%s", userID, level, date),
	)))

	// Проверяем существование
	var existing models.Notification
	err := s.db.Where("user_id = ? AND unique_key = ? AND status = ?",
		userID,
		uniqueKey,
		models.NotificationStatusActive,
	).First(&existing).Error

	if err == nil {
		// Уже есть - обновляем meta с новыми данными
		metaJSON, _ := json.Marshal(meta)
		metaStr := string(metaJSON)
		
		s.db.Model(&existing).Updates(map[string]interface{}{
			"meta":       metaStr,
			"created_at": time.Now(), // Обновляем timestamp
		})
		return nil
	}

	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check existing: %w", err)
	}

	// Создаём новое агрегированное уведомление
	title, message := buildAggregatedText(level, meta.TotalItems)

	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}
	metaStr := string(metaJSON)

	notification := &models.Notification{
		UserID:    userID,
		Type:      models.NotificationTypeFridge,
		Level:     level,
		Title:     title,
		Message:   message,
		Meta:      &metaStr,
		UniqueKey: &uniqueKey,
		Status:    models.NotificationStatusActive,
	}

	if err := s.db.Create(notification).Error; err != nil {
		return fmt.Errorf("failed to create aggregated: %w", err)
	}

	return nil
}

// ============================================================================
// ОЧИСТКА
// ============================================================================

// CleanupExpiredNotifications удалить устаревшие уведомления
// Запускается по cron: уведомления старше 7 дней → expired
func (s *notificationService) CleanupExpiredNotifications() error {
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)

	err := s.db.Model(&models.Notification{}).
		Where("created_at < ? AND status = ?",
			sevenDaysAgo,
			models.NotificationStatusActive,
		).
		Update("status", models.NotificationStatusExpired).Error

	if err != nil {
		return fmt.Errorf("failed to cleanup expired: %w", err)
	}

	return nil
}

// ============================================================================
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
// ============================================================================

// generateUniqueKey генерирует уникальный ключ для уведомления
// Формат: md5(user_id + level + fridge_item_id + date)
func generateUniqueKey(userID string, level models.NotificationLevel, fridgeItemID string, date time.Time) string {
	dateStr := date.Format("2006-01-02")
	raw := fmt.Sprintf("%s-%s-%s-%s", userID, level, fridgeItemID, dateStr)
	return fmt.Sprintf("%x", md5.Sum([]byte(raw)))
}

// buildNotificationText формирует текст уведомления в зависимости от уровня
func buildNotificationText(level models.NotificationLevel, meta models.FridgeNotificationMeta) (string, string) {
	switch level {
	case models.NotificationLevelCritical:
		if meta.DaysLeft < 0 {
			return "⛔ Срочно использовать",
				fmt.Sprintf("%s уже просрочен (%d дн.)", meta.IngredientName, meta.DaysLeft)
		}
		return "⛔ Срочно использовать",
			fmt.Sprintf("%s истекает сегодня", meta.IngredientName)

	case models.NotificationLevelWarning:
		return "⚠️ Скоро испортится",
			fmt.Sprintf("%s — остался 1 день", meta.IngredientName)

	case models.NotificationLevelInfo:
		return "ℹ️ Проверьте холодильник",
			fmt.Sprintf("%s — осталось 2 дня", meta.IngredientName)

	default:
		return "Уведомление", meta.IngredientName
	}
}

// buildAggregatedText формирует текст для агрегированного уведомления
func buildAggregatedText(level models.NotificationLevel, count int) (string, string) {
	switch level {
	case models.NotificationLevelCritical:
		return fmt.Sprintf("⛔ %d продуктов требуют внимания", count),
			"Срок годности истёк или истекает сегодня"

	case models.NotificationLevelWarning:
		return fmt.Sprintf("⚠️ %d продуктов скоро испортятся", count),
			"Остался 1 день до истечения"

	case models.NotificationLevelInfo:
		return fmt.Sprintf("ℹ️ %d продуктов в холодильнике", count),
			"Проверьте срок годности"

	default:
		return "Уведомление", fmt.Sprintf("%d продуктов", count)
	}
}
