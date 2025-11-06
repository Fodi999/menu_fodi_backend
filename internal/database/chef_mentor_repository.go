package database

import (
	"encoding/json"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/google/uuid"
)

// ChefMentorRepository handles database operations for Chef Mentor sessions
type ChefMentorRepository struct{}

// NewChefMentorRepository creates a new repository instance
func NewChefMentorRepository() *ChefMentorRepository {
	return &ChefMentorRepository{}
}

// CreateSession creates a new Chef Mentor session
func (r *ChefMentorRepository) CreateSession(userID *uuid.UUID, language string) (*models.ChefMentorSession, error) {
	session := &models.ChefMentorSession{
		ID:           uuid.New(),
		UserID:       userID,
		Language:     language,
		Context:      make(models.JSONB),
		Recipe:       make(models.JSONB),
		IsComplete:   false,
		LastActivity: time.Now(),
	}

	if err := DB.Create(session).Error; err != nil {
		return nil, err
	}

	return session, nil
}

// GetSession retrieves a session by ID
func (r *ChefMentorRepository) GetSession(sessionID string) (*models.ChefMentorSession, error) {
	var session models.ChefMentorSession
	
	if err := DB.Where("id = ?", sessionID).First(&session).Error; err != nil {
		return nil, err
	}

	return &session, nil
}

// UpdateSession updates session data (recipe, context, etc.)
func (r *ChefMentorRepository) UpdateSession(sessionID string, recipe interface{}, context interface{}) error {
	updates := map[string]interface{}{
		"last_activity": time.Now(),
	}

	if recipe != nil {
		recipeJSON, err := json.Marshal(recipe)
		if err != nil {
			return err
		}
		var recipeMap models.JSONB
		if err := json.Unmarshal(recipeJSON, &recipeMap); err != nil {
			return err
		}
		updates["recipe"] = recipeMap
	}

	if context != nil {
		contextJSON, err := json.Marshal(context)
		if err != nil {
			return err
		}
		var contextMap models.JSONB
		if err := json.Unmarshal(contextJSON, &contextMap); err != nil {
			return err
		}
		updates["context"] = contextMap
	}

	return DB.Model(&models.ChefMentorSession{}).
		Where("id = ?", sessionID).
		Updates(updates).Error
}

// MarkComplete marks session as complete
func (r *ChefMentorRepository) MarkComplete(sessionID string) error {
	return DB.Model(&models.ChefMentorSession{}).
		Where("id = ?", sessionID).
		Updates(map[string]interface{}{
			"is_complete":   true,
			"last_activity": time.Now(),
		}).Error
}

// DeleteSession deletes a session and all its messages
func (r *ChefMentorRepository) DeleteSession(sessionID string) error {
	// Delete messages first (cascade will handle this, but explicit is clearer)
	if err := DB.Where("session_id = ?", sessionID).Delete(&models.ChefMentorMessage{}).Error; err != nil {
		return err
	}

	// Delete session
	return DB.Where("id = ?", sessionID).Delete(&models.ChefMentorSession{}).Error
}

// SaveMessage saves a conversation message
func (r *ChefMentorRepository) SaveMessage(sessionID uuid.UUID, role string, content string) error {
	message := &models.ChefMentorMessage{
		SessionID: sessionID,
		Role:      role,
		Content:   content,
	}

	// Also update session's last_activity
	if err := DB.Create(message).Error; err != nil {
		return err
	}

	return DB.Model(&models.ChefMentorSession{}).
		Where("id = ?", sessionID).
		Update("last_activity", time.Now()).Error
}

// GetMessages retrieves all messages for a session
func (r *ChefMentorRepository) GetMessages(sessionID string) ([]models.ChefMentorMessage, error) {
	var messages []models.ChefMentorMessage
	
	if err := DB.Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		return nil, err
	}

	return messages, nil
}

// DeleteOldSessions removes sessions older than specified duration
func (r *ChefMentorRepository) DeleteOldSessions(olderThan time.Duration) (int64, error) {
	threshold := time.Now().Add(-olderThan)

	result := DB.Where("last_activity < ?", threshold).
		Delete(&models.ChefMentorSession{})

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

// GetUserSessions retrieves all sessions for a user
func (r *ChefMentorRepository) GetUserSessions(userID uuid.UUID, limit int) ([]models.ChefMentorSession, error) {
	var sessions []models.ChefMentorSession
	
	query := DB.Where("user_id = ?", userID).
		Order("last_activity DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&sessions).Error; err != nil {
		return nil, err
	}

	return sessions, nil
}

// CountMessages returns the number of messages in a session
func (r *ChefMentorRepository) CountMessages(sessionID string) (int64, error) {
	var count int64
	
	if err := DB.Model(&models.ChefMentorMessage{}).
		Where("session_id = ?", sessionID).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}
