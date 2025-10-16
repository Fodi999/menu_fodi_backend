package models

import (
	"time"
)

// TransactionType типы транзакций
const (
	TransactionTypeBuy      = "buy"
	TransactionTypeSell     = "sell"
	TransactionTypeBurn     = "burn"
	TransactionTypeTransfer = "transfer"
)

// Transaction представляет транзакцию с токенами
type Transaction struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	BusinessID string    `gorm:"type:uuid;not null" json:"businessId"`
	FromUser   string    `gorm:"type:text" json:"fromUser"`
	ToUser     string    `gorm:"type:text" json:"toUser"`
	Tokens     int64     `gorm:"not null" json:"tokens"`
	Amount     float64   `gorm:"type:numeric(10,2)" json:"amount"`
	TxType     string    `gorm:"type:text;not null" json:"txType"` // buy, sell, burn, transfer
	CreatedAt  time.Time `gorm:"default:now()" json:"createdAt"`

	// Связи
	Business Business `gorm:"foreignKey:BusinessID;references:ID" json:"business,omitempty"`
}

// TableName указывает имя таблицы
func (Transaction) TableName() string {
	return "Transaction"
}
