package service

import (
	"fmt"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"gorm.io/gorm"
)

// NotificationService сервис для работы с уведомлениями
type NotificationService interface {
	GetNotifications(userID string, unreadOnly bool) ([]models.Notification, error)
	MarkAsRead(notificationID string, userID string) error
	MarkAllAsRead(userID string) error
	GetUnreadCount(userID string) (int64, error)
}

type notificationService struct {
	db *gorm.DB
}

func NewNotificationService(db *gorm.DB) NotificationService {
	return &notificationService{db: db}
}

// GetNotifications получить уведомления пользователя
func (s *notificationService) GetNotifications(userID string, unreadOnly bool) ([]models.Notification, error) {
	var notifications []models.Notification
	
	query := s.db.Where("user_id = ?", userID)
	
	if unreadOnly {
		query = query.Where("read_at IS NULL")
	}
	
	err := query.Order("created_at DESC").
		Limit(100). // Последние 100 уведомлений
		Find(&notifications).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch notifications: %w", err)
	}

	return notifications, nil
}

// MarkAsRead пометить уведомление как прочитанное
func (s *notificationService) MarkAsRead(notificationID string, userID string) error {
	result := s.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("read_at", gorm.Expr("NOW()"))

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
	err := s.db.Model(&models.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Update("read_at", gorm.Expr("NOW()")).Error

	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}

// GetUnreadCount получить количество непрочитанных уведомлений
func (s *notificationService) GetUnreadCount(userID string) (int64, error) {
	var count int64
	
	err := s.db.Model(&models.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	return count, nil
}
