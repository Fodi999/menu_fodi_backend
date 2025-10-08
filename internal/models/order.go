package models

import "time"

// Order модель заказа
type Order struct {
	ID        string      `gorm:"primaryKey;type:text;column:id" json:"id"`
	UserID    *string     `gorm:"type:text;column:user_id" json:"userId,omitempty"` // Nullable для гостевых заказов
	Name      string      `gorm:"type:varchar(100);column:name" json:"name"`
	Status    string      `gorm:"type:varchar(20);default:'pending';column:status" json:"status"` // "pending", "confirmed", "preparing", "delivered", "cancelled"
	Total     float64     `gorm:"type:decimal(10,2);not null;column:total" json:"total"`
	Address   string      `gorm:"type:text;column:address" json:"address"`
	Phone     string      `gorm:"type:varchar(20);column:phone" json:"phone"`
	Comment   string      `gorm:"type:text;column:comment" json:"comment"`
	Items     []OrderItem `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
	CreatedAt time.Time   `gorm:"autoCreateTime;column:created_at" json:"createdAt"`
	UpdatedAt time.Time   `gorm:"autoUpdateTime;column:updated_at" json:"updatedAt"`
}

// TableName указывает имя таблицы для GORM
func (Order) TableName() string {
	return "Order"
}

// OrderItem позиция в заказе
type OrderItem struct {
	ID        string  `gorm:"primaryKey;type:text;column:id" json:"id"`
	OrderID   string  `gorm:"type:text;not null;column:order_id" json:"orderId"`
	ProductID string  `gorm:"type:text;not null;column:product_id" json:"productId"`
	Quantity  int     `gorm:"type:int;not null;column:quantity" json:"quantity"`
	Price     float64 `gorm:"type:decimal(10,2);not null;column:price" json:"price"`
}

// TableName указывает имя таблицы для GORM
func (OrderItem) TableName() string {
	return "OrderItem"
}
