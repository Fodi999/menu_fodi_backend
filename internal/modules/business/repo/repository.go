package repo

import (
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"gorm.io/gorm"
)

type BusinessRepository struct {
	db *gorm.DB
}

func NewBusinessRepository(db *gorm.DB) *BusinessRepository {
	return &BusinessRepository{db: db}
}

func (r *BusinessRepository) GetAll() ([]models.Business, error) {
	var businesses []models.Business
	if err := r.db.Order("created_at DESC").Find(&businesses).Error; err != nil {
		return nil, err
	}
	return businesses, nil
}

func (r *BusinessRepository) GetByID(id string) (*models.Business, error) {
	var business models.Business
	if err := r.db.First(&business, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &business, nil
}

func (r *BusinessRepository) Create(business *models.Business) error {
	return r.db.Create(business).Error
}

func (r *BusinessRepository) Update(business *models.Business) error {
	return r.db.Save(business).Error
}

func (r *BusinessRepository) Delete(id string) error {
	return r.db.Model(&models.Business{}).Where("id = ?", id).Update("is_active", false).Error
}

func (r *BusinessRepository) GetToken(businessID string) (*models.BusinessToken, error) {
	var token models.BusinessToken
	if err := r.db.First(&token, "business_id = ?", businessID).Error; err != nil {
		return nil, err
	}
	return &token, nil
}
