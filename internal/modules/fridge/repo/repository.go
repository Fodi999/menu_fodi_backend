package repo

import (
	"errors"
	"fmt"

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
	GetUserFridge(userID uuid.UUID) ([]models.UserFridgeItem, error)
	GetItemByID(itemID, userID uuid.UUID) (*models.UserFridgeItem, error)
	AddItem(userID uuid.UUID, product string, quantity float64, unit string) error
	UpdateItem(itemID, userID uuid.UUID, updates map[string]interface{}) error
	DeleteItem(itemID, userID uuid.UUID) error

	// Query operations
	CountItems(userID uuid.UUID) (int, error)
	GetAvailableItems(userID uuid.UUID) ([]models.UserFridgeItem, error)
}

type fridgeRepository struct {
	db *gorm.DB
}

// NewFridgeRepository creates new fridge repository
func NewFridgeRepository(db *gorm.DB) FridgeRepository {
	return &fridgeRepository{db: db}
}

func (r *fridgeRepository) GetUserFridge(userID uuid.UUID) ([]models.UserFridgeItem, error) {
	var items []models.UserFridgeItem
	err := r.db.Where("user_id = ?", userID).
		Order("added_at DESC").
		Find(&items).Error
	return items, err
}

func (r *fridgeRepository) GetItemByID(itemID, userID uuid.UUID) (*models.UserFridgeItem, error) {
	var item models.UserFridgeItem
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
	// Преобразуем quantity + unit в string формат
	quantityStr := fmt.Sprintf("%.2f %s", quantity, unit)

	item := &models.UserFridgeItem{
		ID:       uuid.New().String(),
		UserID:   userID.String(),
		Name:     product,
		Quantity: quantityStr,
	}
	return r.db.Create(item).Error
}

func (r *fridgeRepository) UpdateItem(itemID, userID uuid.UUID, updates map[string]interface{}) error {
	// First check if item exists and belongs to user
	var item models.UserFridgeItem
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
		Delete(&models.UserFridgeItem{})

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
	err := r.db.Model(&models.UserFridgeItem{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return int(count), err
}

func (r *fridgeRepository) GetAvailableItems(userID uuid.UUID) ([]models.UserFridgeItem, error) {
	var items []models.UserFridgeItem
	err := r.db.Where("user_id = ? AND available = ?", userID, true).
		Order("added_at DESC").
		Find(&items).Error
	return items, err
}
