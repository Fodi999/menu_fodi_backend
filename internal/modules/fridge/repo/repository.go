package repo

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
)

var (
	ErrItemNotFound = errors.New("fridge item not found")
	ErrUnauthorized = errors.New("unauthorized to access this item")
	ErrInvalidItem  = errors.New("invalid item data")
)

// FridgeRepository handles fridge data operations
type FridgeRepository interface {
	// Item operations
	GetUserFridge(userID uuid.UUID) ([]models.UserFridge, error)
	GetItemByID(itemID, userID uuid.UUID) (*models.UserFridge, error)
	AddItem(userID uuid.UUID, product string, quantity float64, unit string) error
	UpdateItem(itemID, userID uuid.UUID, updates map[string]interface{}) error
	DeleteItem(itemID, userID uuid.UUID) error

	// Query operations
	CountItems(userID uuid.UUID) (int, error)
	GetAvailableItems(userID uuid.UUID) ([]models.UserFridge, error)
}

type fridgeRepository struct {
	db *gorm.DB
}

// NewFridgeRepository creates new fridge repository
func NewFridgeRepository(db *gorm.DB) FridgeRepository {
	return &fridgeRepository{db: db}
}

func (r *fridgeRepository) GetUserFridge(userID uuid.UUID) ([]models.UserFridge, error) {
	var items []models.UserFridge
	err := r.db.Where("user_id = ?", userID).
		Order("added_at DESC").
		Find(&items).Error
	return items, err
}

func (r *fridgeRepository) GetItemByID(itemID, userID uuid.UUID) (*models.UserFridge, error) {
	var item models.UserFridge
	err := r.db.Where("id = ? AND user_id = ?", itemID, userID).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrItemNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *fridgeRepository) AddItem(userID uuid.UUID, product string, quantity float64, unit string) error {
	item := &models.UserFridge{
		UserID:    userID,
		Product:   product,
		Quantity:  quantity,
		Unit:      unit,
		Available: true,
	}
	return r.db.Create(item).Error
}

func (r *fridgeRepository) UpdateItem(itemID, userID uuid.UUID, updates map[string]interface{}) error {
	// First check if item exists and belongs to user
	var item models.UserFridge
	err := r.db.Where("id = ? AND user_id = ?", itemID, userID).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrItemNotFound
		}
		return err
	}

	// Perform update
	return r.db.Model(&item).Updates(updates).Error
}

func (r *fridgeRepository) DeleteItem(itemID, userID uuid.UUID) error {
	result := r.db.Where("id = ? AND user_id = ?", itemID, userID).
		Delete(&models.UserFridge{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrItemNotFound
	}

	return nil
}

func (r *fridgeRepository) CountItems(userID uuid.UUID) (int, error) {
	var count int64
	err := r.db.Model(&models.UserFridge{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return int(count), err
}

func (r *fridgeRepository) GetAvailableItems(userID uuid.UUID) ([]models.UserFridge, error) {
	var items []models.UserFridge
	err := r.db.Where("user_id = ? AND available = ?", userID, true).
		Order("added_at DESC").
		Find(&items).Error
	return items, err
}
