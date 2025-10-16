package models

import (
	"time"
)

// Business представляет бизнес в системе
type Business struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OwnerID     string    `gorm:"type:text" json:"ownerId"`
	Name        string    `gorm:"type:text;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Category    string    `gorm:"type:text" json:"category"`
	City        string    `gorm:"type:text" json:"city"`
	IsActive    bool      `gorm:"default:true" json:"isActive"`
	CreatedAt   time.Time `gorm:"default:now()" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"default:now()" json:"updatedAt"`

	// Связи (без foreign key constraint для упрощения)
	Tokens        []BusinessToken        `gorm:"foreignKey:BusinessID" json:"tokens,omitempty"`
	Subscriptions []BusinessSubscription `gorm:"foreignKey:BusinessID" json:"subscriptions,omitempty"`
	Transactions  []Transaction          `gorm:"foreignKey:BusinessID" json:"transactions,omitempty"`
}

// TableName указывает имя таблицы
func (Business) TableName() string {
	return "Business"
}
