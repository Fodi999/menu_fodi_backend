package models

import "time"

// User модель пользователя (соответствует Prisma схеме)
type User struct {
	ID        string    `gorm:"primaryKey;column:id" json:"id"`
	Email     string    `gorm:"unique;column:email" json:"email"`
	Name      string    `gorm:"column:name" json:"name"`
	Password  string    `gorm:"column:password" json:"-"`             // не возвращается в JSON
	Role      string    `gorm:"column:role;default:user" json:"role"` // "user" или "admin"
	CreatedAt time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
}

// TableName указывает имя таблицы для GORM (Prisma использует "User")
func (User) TableName() string {
	return "User"
}

// RegisterRequest запрос на регистрацию
type RegisterRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// LoginRequest запрос на вход
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse ответ при входе
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// UpdateProfileRequest запрос на обновление профиля
type UpdateProfileRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
