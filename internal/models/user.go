package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// User roles constants - используй ТОЛЬКО эти константы для избежания опечаток
const (
	RoleCustomer   = "customer"    // Покупатель (обычный пользователь)
	RoleHomeChef   = "home_chef"   // Домашний повар
	RoleChefStaff  = "chef_staff"  // Младший повар / персонал
	RoleAdmin      = "admin"       // Администратор
	RoleSuperAdmin = "super_admin" // Супер администратор (владелец системы)
)

// User account status constants
const (
	UserStatusPending   = "pending"   // Зарегистрирован, но не активирован
	UserStatusActive    = "active"    // Активен - может использовать все функции
	UserStatusSuspended = "suspended" // Временно отключён
	UserStatusBlocked   = "blocked"   // Заблокирован администратором
)

// User модель пользователя (соответствует Prisma схеме)
type User struct {
	ID        string       `gorm:"primaryKey;column:id" json:"id"`
	Email     string       `gorm:"unique;column:email" json:"email"`
	Name      string       `gorm:"column:name" json:"name"`
	Password  string       `gorm:"column:password" json:"-"`                     // не возвращается в JSON
	Role      string       `gorm:"column:role;default:home_chef" json:"role"`    // home_chef | pro_chef | admin
	Status    string       `gorm:"column:status;default:active" json:"status"`   // active | blocked | pending
	Settings  UserSettings `gorm:"type:jsonb;column:settings" json:"settings"`   // User preferences
	LastLogin *time.Time   `gorm:"column:last_login" json:"lastLogin,omitempty"` // Activity tracking
	CreatedAt time.Time    `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
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

// JSONB support for UserSettings
// Scan implements sql.Scanner for GORM
func (s *UserSettings) Scan(value interface{}) error {
	if value == nil {
		*s = DefaultUserSettings()
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		*s = DefaultUserSettings()
		return nil
	}

	err := json.Unmarshal(bytes, s)
	if err != nil {
		*s = DefaultUserSettings()
		return err
	}

	return nil
}

// Value implements driver.Valuer for GORM
func (s UserSettings) Value() (driver.Value, error) {
	return json.Marshal(s)
}
