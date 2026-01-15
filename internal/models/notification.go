package models

import (
	"time"
)

// NotificationType тип уведомления
type NotificationType string

const (
	NotificationTypeSystem  NotificationType = "system"
	NotificationTypeOrder   NotificationType = "order"
	NotificationTypeUser    NotificationType = "user"
	NotificationTypeFridge  NotificationType = "fridge"
	NotificationTypeAI      NotificationType = "ai"
	NotificationTypeBackup  NotificationType = "backup"
)

// NotificationLevel уровень важности
type NotificationLevel string

const (
	NotificationLevelInfo     NotificationLevel = "info"
	NotificationLevelWarning  NotificationLevel = "warning"
	NotificationLevelCritical NotificationLevel = "critical"
)

// Notification - универсальное уведомление
type Notification struct {
	ID        string            `gorm:"primaryKey;type:uuid;default:gen_random_uuid();column:id" json:"id"`
	UserID    string            `gorm:"column:user_id;not null;index" json:"userId"`
	Type      NotificationType  `gorm:"column:type;not null" json:"type"`
	Level     NotificationLevel `gorm:"column:level;not null" json:"level"`
	Title     string            `gorm:"column:title;not null" json:"title"`
	Message   string            `gorm:"column:message;not null" json:"message"`
	Meta      *string           `gorm:"column:meta;type:jsonb" json:"meta,omitempty"` // JSON для доп. данных
	ReadAt    *time.Time        `gorm:"column:read_at" json:"readAt,omitempty"`
	CreatedAt time.Time         `gorm:"column:created_at;autoCreateTime" json:"createdAt"`

	// Relations
	User *User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

func (Notification) TableName() string {
	return "notifications"
}
