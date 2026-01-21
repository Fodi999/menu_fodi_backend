package models

import (
	"time"
)

// NotificationType тип уведомления
type NotificationType string

const (
	NotificationTypeSystem NotificationType = "system"
	NotificationTypeOrder  NotificationType = "order"
	NotificationTypeUser   NotificationType = "user"
	NotificationTypeFridge NotificationType = "fridge"
	NotificationTypeAI     NotificationType = "ai"
	NotificationTypeBackup NotificationType = "backup"
)

// NotificationLevel уровень важности
type NotificationLevel string

const (
	NotificationLevelInfo     NotificationLevel = "info"
	NotificationLevelWarning  NotificationLevel = "warning"
	NotificationLevelCritical NotificationLevel = "critical"
)

// NotificationStatus статус уведомления
type NotificationStatus string

const (
	NotificationStatusActive   NotificationStatus = "active"   // текущее, актуальное
	NotificationStatusResolved NotificationStatus = "resolved" // решено пользователем
	NotificationStatusExpired  NotificationStatus = "expired"  // устарело автоматически
)

// Notification - универсальное уведомление
// ✅ Используется ТОЛЬКО для событий, требующих внимания
// ❌ НЕ используется для activity logs (добавлен/удалён продукт)
type Notification struct {
	ID        string             `gorm:"primaryKey;type:uuid;default:gen_random_uuid();column:id" json:"id"`
	UserID    string             `gorm:"column:user_id;not null;index" json:"userId"`
	Type      NotificationType   `gorm:"column:type;not null" json:"type"`
	Level     NotificationLevel  `gorm:"column:level;not null" json:"level"`
	Title     string             `gorm:"column:title;not null" json:"title"`
	Message   string             `gorm:"column:message;not null" json:"message"`
	Meta      *string            `gorm:"column:meta;type:jsonb" json:"meta,omitempty"` // JSON для доп. данных
	UniqueKey *string            `gorm:"column:unique_key" json:"-"`                   // для предотвращения дублей
	Status    NotificationStatus `gorm:"column:status;default:'active'" json:"status"`
	ReadAt    *time.Time         `gorm:"column:read_at" json:"readAt,omitempty"`
	CreatedAt time.Time          `gorm:"column:created_at;autoCreateTime" json:"createdAt"`

	// Relations
	User *User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

func (Notification) TableName() string {
	return "notifications"
}

// ============================================================================
// META STRUCTURES для разных типов уведомлений
// ============================================================================

// FridgeNotificationMeta - метаданные для уведомлений о холодильнике
// Используется для expiry tracking (истекает срок годности)
type FridgeNotificationMeta struct {
	FridgeItemID   string  `json:"fridgeItemId"`            // ID записи в холодильнике
	IngredientID   string  `json:"ingredientId"`            // ID ингредиента
	IngredientName string  `json:"ingredientName"`          // Имя продукта (для быстрого доступа)
	DaysLeft       int     `json:"daysLeft"`                // Дней до истечения (-1 = просрочен)
	ExpiresAt      string  `json:"expiresAt"`               // ISO timestamp срока годности
	Quantity       float64 `json:"quantity"`                // Количество
	Unit           string  `json:"unit"`                    // Единица измерения
	CategoryKey    string  `json:"categoryKey,omitempty"`   // Категория (fish, meat, etc)
}

// AggregatedFridgeMeta - агрегированные данные (несколько продуктов)
// Используется для summary notifications ("3 продукта скоро испортятся")
type AggregatedFridgeMeta struct {
	TotalItems int                        `json:"totalItems"` // Общее количество продуктов
	Items      []FridgeNotificationMeta   `json:"items"`      // Список продуктов
}

// ============================================================================
// API RESPONSE STRUCTURES
// ============================================================================

// NotificationGroup - группировка уведомлений по уровню важности
type NotificationGroup struct {
	Critical []Notification `json:"critical"` // Срочные (daysLeft <= 0)
	Warning  []Notification `json:"warning"`  // Скоро (daysLeft = 1)
	Info     []Notification `json:"info"`     // Информационные (daysLeft = 2, summary)
}

// UnreadCount - количество непрочитанных уведомлений
type UnreadCount struct {
	Critical int `json:"critical"` // Считаются в badge
	Warning  int `json:"warning"`  // Считаются в badge
	Info     int `json:"info"`     // НЕ считаются в badge (только для display)
	Total    int `json:"total"`    // critical + warning
}
