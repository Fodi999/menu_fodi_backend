package service

import (
	"fmt"
	"math"
	"strings"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/semi_finished/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/semi_finished/repo"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SemiFinishedService handles business logic for semi-finished products
type SemiFinishedService struct {
	repo *repo.SemiFinishedRepository
}

// NewSemiFinishedService creates a new service instance
func NewSemiFinishedService(repository *repo.SemiFinishedRepository) *SemiFinishedService {
	return &SemiFinishedService{repo: repository}
}

// NormalizeFloat rounds a float to specified decimal places
func (s *SemiFinishedService) NormalizeFloat(value float64, decimals int) float64 {
	mult := math.Pow(10, float64(decimals))
	return math.Round(value*mult) / mult
}

// ConvertToBaseUnit converts value to base unit (g -> kg, ml -> l)
func (s *SemiFinishedService) ConvertToBaseUnit(value float64, unit string) float64 {
	switch strings.ToLower(unit) {
	case "g":
		return value / 1000
	case "ml":
		return value / 1000
	default:
		return value
	}
}

// CalculateCostPerUnit calculates cost per unit of semi-finished product
func (s *SemiFinishedService) CalculateCostPerUnit(ingredients []dto.CreateSemiFinishedIngredientInput, outputQty float64) float64 {
	var totalCost float64
	for _, ing := range ingredients {
		qty := s.ConvertToBaseUnit(ing.Quantity, ing.Unit)
		totalCost += qty * ing.PricePerUnit
	}
	if outputQty == 0 {
		return 0
	}
	return s.NormalizeFloat(totalCost/outputQty, 2)
}

// GetAll retrieves all semi-finished products
func (s *SemiFinishedService) GetAll() ([]models.SemiFinished, error) {
	return s.repo.GetAll()
}

// GetByID retrieves a semi-finished product by ID
func (s *SemiFinishedService) GetByID(id string) (*models.SemiFinished, error) {
	return s.repo.GetByID(id)
}

// Create creates a new semi-finished product
func (s *SemiFinishedService) Create(req *dto.CreateSemiFinishedRequest) (*models.SemiFinished, error) {
	// Validate required fields
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.OutputQuantity <= 0 {
		return nil, fmt.Errorf("output quantity must be positive")
	}
	if req.OutputUnit == "" {
		return nil, fmt.Errorf("output unit is required")
	}
	if len(req.Ingredients) == 0 {
		return nil, fmt.Errorf("at least one ingredient is required")
	}

	// Validate ingredients
	for i, ing := range req.Ingredients {
		if ing.IngredientID == "" {
			return nil, fmt.Errorf("ingredient ID is required for ingredient %d", i+1)
		}
		if ing.IngredientName == "" {
			return nil, fmt.Errorf("ingredient name is required for ingredient %d", i+1)
		}
		if ing.Unit == "" {
			return nil, fmt.Errorf("unit is required for ingredient %d", i+1)
		}
		if ing.Quantity <= 0 {
			return nil, fmt.Errorf("ingredient quantity must be positive for ingredient %d", i+1)
		}

		// Check if ingredient exists
		exists, err := s.repo.CheckIngredientExists(ing.IngredientID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, fmt.Errorf("ingredient with ID '%s' does not exist", ing.IngredientID)
		}
	}

	// Check for duplicate names
	exists, err := s.repo.CheckNameExists(req.Name, "")
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("semi-finished with this name already exists")
	}

	// Normalize numeric values
	for i := range req.Ingredients {
		req.Ingredients[i].Quantity = s.NormalizeFloat(req.Ingredients[i].Quantity, 3)
		req.Ingredients[i].PricePerUnit = s.NormalizeFloat(req.Ingredients[i].PricePerUnit, 2)
		req.Ingredients[i].TotalPrice = s.NormalizeFloat(req.Ingredients[i].TotalPrice, 2)
	}
	outputQty := s.NormalizeFloat(req.OutputQuantity, 3)

	// Calculate cost per unit
	costPerUnit := s.CalculateCostPerUnit(req.Ingredients, outputQty)
	totalCost := s.NormalizeFloat(costPerUnit*outputQty, 2)

	// Create semi-finished product
	id := uuid.New().String()
	var description *string
	if req.Description != "" {
		description = &req.Description
	}

	sf := &models.SemiFinished{
		ID:             id,
		Name:           req.Name,
		Description:    description,
		OutputQuantity: outputQty,
		OutputUnit:     req.OutputUnit,
		CostPerUnit:    costPerUnit,
		TotalCost:      totalCost,
		Category:       req.Category,
		IsVisible:      true,
		IsArchived:     false,
	}

	if err := s.repo.Create(sf); err != nil {
		return nil, err
	}

	// Add ingredients in transaction
	for _, ing := range req.Ingredients {
		ingredient := &models.SemiFinishedIngredient{
			ID:             uuid.New().String(),
			SemiFinishedID: id,
			IngredientID:   ing.IngredientID,
			IngredientName: ing.IngredientName,
			Quantity:       ing.Quantity,
			Unit:           ing.Unit,
			PricePerUnit:   ing.PricePerUnit,
			TotalPrice:     ing.TotalPrice,
		}
		if err := s.repo.AddIngredient(ingredient); err != nil {
			return nil, err
		}
	}

	return sf, nil
}

// Update updates a semi-finished product
func (s *SemiFinishedService) Update(id string, req *dto.UpdateSemiFinishedRequest) (*models.SemiFinished, error) {
	sf, err := s.repo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("semi-finished not found")
		}
		return nil, err
	}

	// Update name if provided
	if req.Name != nil && *req.Name != "" {
		exists, err := s.repo.CheckNameExists(*req.Name, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("semi-finished with this name already exists")
		}
		sf.Name = *req.Name
	}

	// Update other fields
	if req.Description != nil {
		sf.Description = req.Description
	}
	if req.Category != nil {
		sf.Category = *req.Category
	}
	if req.OutputUnit != nil {
		sf.OutputUnit = *req.OutputUnit
	}

	// Normalize numeric values if ingredients provided
	if len(req.Ingredients) > 0 {
		for i := range req.Ingredients {
			req.Ingredients[i].Quantity = s.NormalizeFloat(req.Ingredients[i].Quantity, 3)
			req.Ingredients[i].PricePerUnit = s.NormalizeFloat(req.Ingredients[i].PricePerUnit, 2)
			req.Ingredients[i].TotalPrice = s.NormalizeFloat(req.Ingredients[i].TotalPrice, 2)
		}
	}

	// Handle output quantity and cost recalculation
	outputQty := sf.OutputQuantity
	if req.OutputQuantity != nil && *req.OutputQuantity > 0 {
		outputQty = s.NormalizeFloat(*req.OutputQuantity, 3)
		sf.OutputQuantity = outputQty
	}

	// Recalculate cost if ingredients provided
	if len(req.Ingredients) > 0 {
		// Convert to CreateSemiFinishedIngredientInput for calculation
		ingredients := make([]dto.CreateSemiFinishedIngredientInput, len(req.Ingredients))
		for i, ing := range req.Ingredients {
			ingredients[i] = dto.CreateSemiFinishedIngredientInput{
				IngredientID:  ing.IngredientID,
				IngredientName: ing.IngredientName,
				Quantity:      ing.Quantity,
				Unit:          ing.Unit,
				PricePerUnit:  ing.PricePerUnit,
				TotalPrice:    ing.TotalPrice,
			}
		}
		costPerUnit := s.CalculateCostPerUnit(ingredients, outputQty)
		sf.CostPerUnit = costPerUnit
		sf.TotalCost = s.NormalizeFloat(costPerUnit*outputQty, 2)

		// Delete old ingredients and add new ones
		if err := s.repo.DeleteIngredients(id); err != nil {
			return nil, err
		}
		for _, ing := range req.Ingredients {
			ingredient := &models.SemiFinishedIngredient{
				ID:             uuid.New().String(),
				SemiFinishedID: id,
				IngredientID:   ing.IngredientID,
				IngredientName: ing.IngredientName,
				Quantity:       ing.Quantity,
				Unit:           ing.Unit,
				PricePerUnit:   ing.PricePerUnit,
				TotalPrice:     ing.TotalPrice,
			}
			if err := s.repo.AddIngredient(ingredient); err != nil {
				return nil, err
			}
		}
	} else if req.OutputQuantity != nil {
		// Only quantity changed, recalculate total cost
		sf.TotalCost = s.NormalizeFloat(sf.CostPerUnit*outputQty, 2)
	}

	if err := s.repo.Update(sf); err != nil {
		return nil, err
	}

	return sf, nil
}

// Delete deletes a semi-finished product
func (s *SemiFinishedService) Delete(id string) error {
	// Delete ingredients first
	if err := s.repo.DeleteIngredients(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}
