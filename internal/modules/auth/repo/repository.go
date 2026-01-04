package repo

import (
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuthRepository handles authentication database operations
type AuthRepository interface {
	FindByEmail(email string) (*models.User, error)
	FindByID(id string) (*models.User, error)
	Create(user *models.User) error
	Update(user *models.User) error
	UpdateLastLogin(userID string, loginTime time.Time) error
	GetUserProfile(userID uuid.UUID) (*models.UserProfile, error)
}

// authRepository implements AuthRepository
type authRepository struct {
	db *gorm.DB
}

// NewAuthRepository creates a new auth repository
func NewAuthRepository() AuthRepository {
	return &authRepository{
		db: database.DB,
	}
}

// FindByEmail finds user by email
func (r *authRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID finds user by ID
func (r *authRepository) FindByID(id string) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create creates a new user
func (r *authRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// Update updates existing user
func (r *authRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// UpdateLastLogin updates user's last login timestamp
func (r *authRepository) UpdateLastLogin(userID string, loginTime time.Time) error {
	return r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("last_login", loginTime).Error
}

// GetUserProfile retrieves user profile
func (r *authRepository) GetUserProfile(userID uuid.UUID) (*models.UserProfile, error) {
	var profile models.UserProfile
	err := r.db.Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Profile doesn't exist
		}
		return nil, err
	}
	return &profile, nil
}
