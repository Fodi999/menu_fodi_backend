package models

import "time"

// TokenBank модель для отслеживания баланса токинов пользователя
type TokenBank struct {
	ID             string    `gorm:"primaryKey;column:id;type:uuid" json:"id"`
	UserID         string    `gorm:"uniqueIndex;column:user_id;type:uuid" json:"user_id"`
	Balance        int64     `gorm:"column:balance;default:0" json:"balance"`           // Текущий доступный баланс
	TotalAllocated int64     `gorm:"column:total_allocated;default:0" json:"total_allocated"` // Всего выдано админом
	TotalUsed      int64     `gorm:"column:total_used;default:0" json:"total_used"`     // Всего использовано
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	// Association
	User *User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

// TableName указывает имя таблицы для GORM
func (TokenBank) TableName() string {
	return "token_bank"
}

// AllocateTokensRequest запрос на выделение токинов пользователю
type AllocateTokensRequest struct {
	UserID string `json:"user_id"`
	Amount int64  `json:"amount"`
	Reason string `json:"reason,omitempty"` // Причина выделения (опционально)
}

// RevokeTokensRequest запрос на отзыв токинов у пользователя
type RevokeTokensRequest struct {
	UserID string `json:"user_id"`
	Amount int64  `json:"amount"`
	Reason string `json:"reason,omitempty"` // Причина отзыва (опционально)
}

// UpdateTokenBalanceRequest запрос на обновление баланса
type UpdateTokenBalanceRequest struct {
	UserID string `json:"user_id"`
	Amount int64  `json:"amount"`
	Action string `json:"action"` // "add" или "subtract"
}

// TokenBankStats статистика по токинам
type TokenBankStats struct {
	TotalTokensAllocated int64 `json:"total_tokens_allocated"`
	TotalTokensUsed      int64 `json:"total_tokens_used"`
	TotalUsersWithTokens int64 `json:"total_users_with_tokens"`
	AverageBalancePerUser float64 `json:"average_balance_per_user"`
}
