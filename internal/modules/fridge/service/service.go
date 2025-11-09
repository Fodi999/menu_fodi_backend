package service

import (
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/fridge/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/fridge/repo"
)

var (
	ErrInvalidQuantity = errors.New("quantity must be greater than zero")
	ErrEmptyProduct    = errors.New("product name cannot be empty")
	ErrEmptyUnit       = errors.New("unit cannot be empty")
	ErrNoUpdates       = errors.New("no updates provided")
)

// FridgeService handles fridge business logic
type FridgeService interface {
	GetUserFridge(userID uuid.UUID) (*dto.FridgeListResponse, error)
	AddItem(userID uuid.UUID, req dto.AddFridgeItemRequest) error
	UpdateItem(itemID, userID uuid.UUID, req dto.UpdateFridgeItemRequest) (*dto.FridgeItemResponse, error)
	DeleteItem(itemID, userID uuid.UUID) error
	GetAvailableItems(userID uuid.UUID) ([]dto.FridgeItemResponse, error)
}

type fridgeService struct {
	repo repo.FridgeRepository
}

// NewFridgeService creates new fridge service
func NewFridgeService(repo repo.FridgeRepository) FridgeService {
	return &fridgeService{repo: repo}
}

func (s *fridgeService) GetUserFridge(userID uuid.UUID) (*dto.FridgeListResponse, error) {
	items, err := s.repo.GetUserFridge(userID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.FridgeItemResponse, len(items))
	for i, item := range items {
		responses[i] = modelToDTO(&item)
	}

	return &dto.FridgeListResponse{
		Success: true,
		Items:   responses,
		Count:   len(responses),
	}, nil
}

func (s *fridgeService) AddItem(userID uuid.UUID, req dto.AddFridgeItemRequest) error {
	// Validate request
	if err := validateAddItemRequest(req); err != nil {
		return err
	}

	// Normalize product name
	product := strings.TrimSpace(req.Product)
	unit := strings.TrimSpace(req.Unit)

	return s.repo.AddItem(userID, product, req.Quantity, unit)
}

func (s *fridgeService) UpdateItem(itemID, userID uuid.UUID, req dto.UpdateFridgeItemRequest) (*dto.FridgeItemResponse, error) {
	// Validate that at least one field is being updated
	if req.Quantity == 0 && req.Available == nil {
		return nil, ErrNoUpdates
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Quantity > 0 {
		updates["quantity"] = req.Quantity
	}
	if req.Available != nil {
		updates["available"] = *req.Available
	}

	// Perform update
	if err := s.repo.UpdateItem(itemID, userID, updates); err != nil {
		return nil, err
	}

	// Fetch updated item
	item, err := s.repo.GetItemByID(itemID, userID)
	if err != nil {
		return nil, err
	}

	response := modelToDTO(item)
	return &response, nil
}

func (s *fridgeService) DeleteItem(itemID, userID uuid.UUID) error {
	return s.repo.DeleteItem(itemID, userID)
}

func (s *fridgeService) GetAvailableItems(userID uuid.UUID) ([]dto.FridgeItemResponse, error) {
	items, err := s.repo.GetAvailableItems(userID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.FridgeItemResponse, len(items))
	for i, item := range items {
		responses[i] = modelToDTO(&item)
	}

	return responses, nil
}

// Helper functions

func validateAddItemRequest(req dto.AddFridgeItemRequest) error {
	if strings.TrimSpace(req.Product) == "" {
		return ErrEmptyProduct
	}
	if req.Quantity <= 0 {
		return ErrInvalidQuantity
	}
	if strings.TrimSpace(req.Unit) == "" {
		return ErrEmptyUnit
	}
	return nil
}

func modelToDTO(item *models.UserFridge) dto.FridgeItemResponse {
	return dto.FridgeItemResponse{
		ID:        item.ID,
		UserID:    item.UserID,
		Product:   item.Product,
		Quantity:  item.Quantity,
		Unit:      item.Unit,
		Available: item.Available,
		CreatedAt: item.AddedAt,
		UpdatedAt: item.UpdatedAt,
	}
}
