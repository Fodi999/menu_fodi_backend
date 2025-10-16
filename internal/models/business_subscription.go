package models

import (
	"time"
)

// BusinessSubscription представляет инвестицию пользователя в бизнес
type BusinessSubscription struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      string    `gorm:"type:text;not null" json:"userId"`
	BusinessID  string    `gorm:"type:uuid;not null" json:"businessId"`
	TokensOwned int64     `gorm:"default:1" json:"tokensOwned"`
	Invested    float64   `gorm:"type:numeric(10,2);default:19.00" json:"invested"`
	CreatedAt   time.Time `gorm:"default:now()" json:"createdAt"`

	// Связи
	User     User     `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	Business Business `gorm:"foreignKey:BusinessID;references:ID" json:"business,omitempty"`
}

// TableName указывает имя таблицы
func (BusinessSubscription) TableName() string {
	return "BusinessSubscription"
}

// GetSharePercentage рассчитывает процент владения
func (s *BusinessSubscription) GetSharePercentage(totalSupply int64) float64 {
	if totalSupply == 0 {
		return 0
	}
	return (float64(s.TokensOwned) / float64(totalSupply)) * 100
}
