package database

import (
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
)

// UserRepository репозиторий для работы с пользователями
type UserRepository struct{}

// FindByEmail находит пользователя по email
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	result := DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// FindByID находит пользователя по ID
func (r *UserRepository) FindByID(id string) (*models.User, error) {
	var user models.User
	result := DB.First(&user, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// Create создает нового пользователя
func (r *UserRepository) Create(user *models.User) error {
	result := DB.Create(user)
	return result.Error
}

// Update обновляет пользователя
func (r *UserRepository) Update(user *models.User) error {
	result := DB.Save(user)
	return result.Error
}

// Delete удаляет пользователя
func (r *UserRepository) Delete(id string) error {
	result := DB.Delete(&models.User{}, "id = ?", id)
	return result.Error
}

// FindAll возвращает всех пользователей
func (r *UserRepository) FindAll() ([]models.User, error) {
	var users []models.User
	result := DB.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}
