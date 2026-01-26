package dto

import (
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
)

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Name     string `json:"name" validate:"required,min=2"`
}

// LoginRequest represents user login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// VerifyTokenRequest represents token verification request
type VerifyTokenRequest struct {
	Token string `json:"token" validate:"required"`
}

// AuthResponse represents authentication response with token
type AuthResponse struct {
	Success bool     `json:"success"`
	Token   string   `json:"token"`
	User    UserInfo `json:"user"`
	Message string   `json:"message,omitempty"`
}

// UserInfo represents user information in response
type UserInfo struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

// CurrentUserResponse represents current user with profile
type CurrentUserResponse struct {
	ID            string           `json:"id"`
	Email         string           `json:"email"`
	Name          string           `json:"name"`
	Role          string           `json:"role"`
	Status        string           `json:"status"`
	CreatedAt     time.Time        `json:"createdAt"`
	WalletBalance int              `json:"walletBalance"`
	Profile       *UserProfileInfo `json:"profile,omitempty"`
}

// UserProfileInfo represents user profile information
type UserProfileInfo struct {
	Level            int    `json:"level"`
	Stars            int    `json:"stars"`
	XP               int    `json:"xp"`
	AvatarURL        string `json:"avatarUrl"`
	Language         string `json:"language"`
	CompletedCourses int    `json:"completedCourses"`
}

// VerifyTokenResponse represents token verification response
type VerifyTokenResponse struct {
	Valid  bool   `json:"valid"`
	UserID string `json:"user_id,omitempty"`
	Role   string `json:"role,omitempty"`
	Name   string `json:"name,omitempty"`
	Email  string `json:"email,omitempty"`
}

// Helper function to convert models.User to UserInfo
func ToUserInfo(user *models.User) UserInfo {
	return UserInfo{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}
}
