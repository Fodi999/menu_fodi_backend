package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ChefMentorSession - persistent session for AI Chef Mentor
type ChefMentorSession struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       *uuid.UUID `gorm:"type:uuid" json:"userId,omitempty"` // Optional - can be anonymous
	Language     string     `gorm:"type:varchar(5);not null;default:'ua'" json:"language"`
	Context      JSONB      `gorm:"type:jsonb;default:'{}'" json:"context"` // Additional context data
	Recipe       JSONB      `gorm:"type:jsonb;default:'{}'" json:"recipe"`  // Current recipe draft
	IsComplete   bool       `gorm:"default:false" json:"isComplete"`        // Recipe completed
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
	LastActivity time.Time  `gorm:"not null" json:"lastActivity"` // For cleanup
}

// TableName sets table name
func (ChefMentorSession) TableName() string {
	return "chef_mentor_sessions"
}

// ChefMentorMessage - conversation message
type ChefMentorMessage struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SessionID uuid.UUID `gorm:"type:uuid;not null;index" json:"sessionId"`
	Role      string    `gorm:"type:varchar(20);not null" json:"role"` // "user" | "assistant"
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName sets table name
func (ChefMentorMessage) TableName() string {
	return "chef_mentor_messages"
}

// JSONB type for PostgreSQL JSONB columns
type JSONB map[string]interface{}

// Value implements driver.Valuer interface for JSONB
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements sql.Scanner interface for JSONB
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONB)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	result := make(JSONB)
	err := json.Unmarshal(bytes, &result)
	if err != nil {
		return err
	}

	*j = result
	return nil
}
