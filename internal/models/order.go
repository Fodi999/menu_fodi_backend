package models

import "time"

// Order модель заказа
type Order struct {
	ID        string      `json:"id"`
	UserID    string      `json:"userId"`
	Status    string      `json:"status"` // "pending", "confirmed", "delivered", "cancelled"
	Total     float64     `json:"total"`
	Address   string      `json:"address"`
	Phone     string      `json:"phone"`
	Comment   string      `json:"comment"`
	Items     []OrderItem `json:"items,omitempty"`
	CreatedAt time.Time   `json:"createdAt"`
}

// OrderItem позиция в заказе
type OrderItem struct {
	ID        string  `json:"id"`
	OrderID   string  `json:"orderId"`
	ProductID string  `json:"productId"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}
