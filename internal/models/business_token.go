package models

import (
	"time"
)

// BusinessToken представляет токены бизнеса
type BusinessToken struct {
	ID          string  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	BusinessID  string  `gorm:"type:uuid;not null" json:"businessId"`
	Symbol      string  `gorm:"type:text;not null" json:"symbol"`
	TotalSupply int64   `gorm:"default:1" json:"totalSupply"`
	Price       float64 `gorm:"type:numeric(10,2);default:19.00" json:"price"`
	CreatedAt   time.Time `gorm:"default:now()" json:"createdAt"`

	// Связи
	Business Business `gorm:"foreignKey:BusinessID;references:ID" json:"business"`
}

// TableName указывает имя таблицы
func (BusinessToken) TableName() string {
	return "BusinessToken"
}

// GetMarketCap рассчитывает капитализацию (Total Supply × Price)
func (t *BusinessToken) GetMarketCap() float64 {
	return float64(t.TotalSupply) * t.Price
}
