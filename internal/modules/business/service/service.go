package service

import (
	"errors"
	"log"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/business/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/business/repo"
	tokenservice "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/wallet/service"
	"github.com/google/uuid"
)

type BusinessService struct {
	repo              *repo.BusinessRepository
	tokenService      *tokenservice.TokenService
}

func NewBusinessService(repository *repo.BusinessRepository) *BusinessService {
	return &BusinessService{
		repo:         repository,
		tokenService: tokenservice.NewTokenService(),
	}
}

func (s *BusinessService) GetBusinesses() ([]models.Business, error) {
	return s.repo.GetAll()
}

func (s *BusinessService) GetBusinessByID(id string) (*models.Business, error) {
	return s.repo.GetByID(id)
}

func (s *BusinessService) CreateBusiness(req *dto.CreateBusinessRequest) (*models.Business, error) {
	if req.Name == "" {
		return nil, errors.New("name is required")
	}

	business := models.Business{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		City:        req.City,
		OwnerID:     req.OwnerID,
		IsActive:    true,
	}

	if err := s.repo.Create(&business); err != nil {
		log.Printf("[BUSINESS] Error creating business: %v", err)
		return nil, err
	}

	// Create initial token
	_, err := s.tokenService.MintInitialToken(business.ID)
	if err != nil {
		log.Printf("[BUSINESS] Warning: Failed to create initial token: %v", err)
	}

	log.Printf("[BUSINESS] Created ID=%s, Name=%s", business.ID, business.Name)
	return &business, nil
}

func (s *BusinessService) UpdateBusiness(id string, req *dto.UpdateBusinessRequest) (*models.Business, error) {
	business, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("business not found")
	}

	if req.Name != nil && *req.Name != "" {
		business.Name = *req.Name
	}
	if req.Description != nil {
		business.Description = *req.Description
	}
	if req.Category != nil {
		business.Category = *req.Category
	}
	if req.City != nil {
		business.City = *req.City
	}
	if req.IsActive != nil {
		business.IsActive = *req.IsActive
	}

	if err := s.repo.Update(business); err != nil {
		return nil, err
	}

	log.Printf("[BUSINESS] Updated ID=%s, Name=%s", business.ID, business.Name)
	return business, nil
}

func (s *BusinessService) DeleteBusiness(id string) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("business not found")
	}
	return s.repo.Delete(id)
}

func (s *BusinessService) GetBusinessTokens(businessID string) (*models.BusinessToken, error) {
	token, err := s.repo.GetToken(businessID)
	if err != nil {
		return nil, errors.New("business tokens not found")
	}
	return token, nil
}
