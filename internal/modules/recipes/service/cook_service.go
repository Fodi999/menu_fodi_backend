package service

import (
	"fmt"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/recipes/dto"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RecipeCookService handles recipe cooking logic
type RecipeCookService struct {
	db *gorm.DB
}

func NewRecipeCookService(db *gorm.DB) *RecipeCookService {
	return &RecipeCookService{db: db}
}

// CookRecipe deducts ingredients from fridge and logs the cooking event
func (s *RecipeCookService) CookRecipe(
	userID string,
	recipeID string,
	servingsMultiplier float64,
	idempotencyKey *string,
) (*dto.CookRecipeData, error) {
	// Default servings multiplier
	if servingsMultiplier <= 0 {
		servingsMultiplier = 1.0
	}

	// Check for idempotency
	if idempotencyKey != nil && *idempotencyKey != "" {
		var existingLog models.RecipeCookLog
		err := s.db.Where("\"idempotencyKey\" = ?", *idempotencyKey).First(&existingLog).Error
		if err == nil {
			// Already cooked with this key, return existing result
			return s.buildCookResponse(&existingLog)
		} else if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to check idempotency: %w", err)
		}
	}

	// Parse recipe UUID
	recipeUUID, err := uuid.Parse(recipeID)
	if err != nil {
		return nil, fmt.Errorf("invalid recipe ID: %w", err)
	}

	// Load recipe with ingredients
	var recipe models.RecipeCatalog
	err = s.db.
		Preload("Ingredients.Ingredient").
		Where("id = ?", recipeUUID).
		First(&recipe).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("recipe not found")
		}
		return nil, fmt.Errorf("failed to load recipe: %w", err)
	}

	// Load user's fridge items
	var fridgeItems []models.UserFridgeItem
	err = s.db.
		Preload("Ingredient").
		Where("user_id = ?", userID).
		Find(&fridgeItems).Error
	if err != nil {
		return nil, fmt.Errorf("failed to load fridge: %w", err)
	}

	// Build fridge map for quick lookup
	fridgeMap := make(map[string]*models.UserFridgeItem)
	for i := range fridgeItems {
		fridgeMap[fridgeItems[i].IngredientID] = &fridgeItems[i]
	}

	// Validate all required ingredients are available
	var missingIngredients []string
	for _, recipeIng := range recipe.Ingredients {
		if recipeIng.Optional {
			continue // Skip optional ingredients
		}
		
		requiredQty := recipeIng.Quantity * servingsMultiplier
		fridgeItem, exists := fridgeMap[recipeIng.IngredientID]
		
		if !exists || fridgeItem.Quantity < requiredQty {
			missingIngredients = append(missingIngredients, recipeIng.Ingredient.Name)
		}
	}

	if len(missingIngredients) > 0 {
		return nil, fmt.Errorf("missing ingredients: %v", missingIngredients)
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Prepare cook log
	cookLog := models.RecipeCookLog{
		UserID:             userID,
		RecipeID:           recipeUUID,
		ServingsMultiplier: servingsMultiplier,
		CookedAt:           time.Now(),
		IdempotencyKey:     idempotencyKey,
		UsedValue:          0,
		WasteRiskSaved:     0,
		TotalRecipeCost:    0,
	}

	var cookIngredients []models.RecipeCookIngredient
	var usedIngredients []dto.CookedIngredient

	// Deduct ingredients and track usage
	for _, recipeIng := range recipe.Ingredients {
		fridgeItem, exists := fridgeMap[recipeIng.IngredientID]
		if !exists {
			continue // Skip if not in fridge (optional ingredients)
		}

		requiredQty := recipeIng.Quantity * servingsMultiplier
		
		// Skip if insufficient (should not happen after validation)
		if fridgeItem.Quantity < requiredQty {
			continue
		}

		// Calculate cost
		pricePerUnit := 0.0
		if fridgeItem.CurrentPricePerUnit != nil {
			pricePerUnit = *fridgeItem.CurrentPricePerUnit
		} else if recipeIng.Ingredient.DefaultPricePerUnit != nil {
			pricePerUnit = *recipeIng.Ingredient.DefaultPricePerUnit
		}
		
		totalCost := requiredQty * pricePerUnit
		cookLog.UsedValue += totalCost
		cookLog.TotalRecipeCost += totalCost

		// Check if expiring (within 3 days)
		isExpiringSoon := false
		if fridgeItem.ExpiresAt != nil {
			daysUntilExpiry := time.Until(*fridgeItem.ExpiresAt).Hours() / 24
			isExpiringSoon = daysUntilExpiry <= 3 && daysUntilExpiry > 0
			if isExpiringSoon {
				cookLog.WasteRiskSaved += totalCost
			}
		}

		// Deduct from fridge
		newQuantity := fridgeItem.Quantity - requiredQty
		err = tx.Model(&models.UserFridgeItem{}).
			Where("id = ?", fridgeItem.ID).
			Update("quantity", newQuantity).Error
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to deduct ingredient %s: %w", fridgeItem.Ingredient.Name, err)
		}

		// Delete if quantity becomes zero or negative
		if newQuantity <= 0 {
			err = tx.Where("id = ?", fridgeItem.ID).Delete(&models.UserFridgeItem{}).Error
			if err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to remove empty ingredient: %w", err)
			}
		}

		// Track in cook log
		pricePtr := &pricePerUnit
		costPtr := &totalCost
		cookIngredient := models.RecipeCookIngredient{
			IngredientID:    recipeIng.IngredientID,
			QuantityUsed:    requiredQty,
			Unit:            recipeIng.Unit,
			PricePerUnit:    pricePtr,
			TotalCost:       costPtr,
			WasExpiringSoon: isExpiringSoon,
		}
		cookIngredients = append(cookIngredients, cookIngredient)

		// For response
		usedIngredients = append(usedIngredients, dto.CookedIngredient{
			IngredientID:      recipeIng.IngredientID,
			Name:              recipeIng.Ingredient.Name,
			QuantityUsed:      requiredQty,
			Unit:              recipeIng.Unit,
			PricePerUnit:      pricePerUnit,
			TotalCost:         totalCost,
			WasExpiringSoon:   isExpiringSoon,
			RemainingInFridge: newQuantity,
		})
	}

	// Round economy values
	cookLog.UsedValue = roundToTwoDecimals(cookLog.UsedValue)
	cookLog.WasteRiskSaved = roundToTwoDecimals(cookLog.WasteRiskSaved)
	cookLog.TotalRecipeCost = roundToTwoDecimals(cookLog.TotalRecipeCost)

	// Save cook log
	err = tx.Create(&cookLog).Error
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to save cook log: %w", err)
	}

	// Save cook ingredients
	for i := range cookIngredients {
		cookIngredients[i].CookLogID = cookLog.ID
		err = tx.Create(&cookIngredients[i]).Error
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to save cook ingredient: %w", err)
		}
	}

	// Commit transaction
	err = tx.Commit().Error
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Count remaining fridge items
	var remainingCount int64
	s.db.Model(&models.UserFridgeItem{}).Where("user_id = ?", userID).Count(&remainingCount)

	// Build response
	return &dto.CookRecipeData{
		CookLogID:          cookLog.ID.String(),
		RecipeID:           recipe.ID.String(),
		CanonicalName:      recipe.CanonicalName,
		LocalName:          recipe.LocalName,
		ServingsMultiplier: servingsMultiplier,
		CookedAt:           cookLog.CookedAt.Format(time.RFC3339),
		UsedValue:          cookLog.UsedValue,
		WasteRiskSaved:     cookLog.WasteRiskSaved,
		TotalRecipeCost:    cookLog.TotalRecipeCost,
		IngredientsUsed:    usedIngredients,
		FridgeUpdated:      true,
		RemainingItems:     int(remainingCount),
	}, nil
}

// buildCookResponse builds response from existing cook log (for idempotency)
func (s *RecipeCookService) buildCookResponse(cookLog *models.RecipeCookLog) (*dto.CookRecipeData, error) {
	// Load recipe
	var recipe models.RecipeCatalog
	err := s.db.Where("id = ?", cookLog.RecipeID).First(&recipe).Error
	if err != nil {
		return nil, fmt.Errorf("failed to load recipe: %w", err)
	}

	// Load cook ingredients
	var cookIngredients []models.RecipeCookIngredient
	err = s.db.Preload("Ingredient").Where("\"cookLogId\" = ?", cookLog.ID).Find(&cookIngredients).Error
	if err != nil {
		return nil, fmt.Errorf("failed to load cook ingredients: %w", err)
	}

	// Convert to DTO
	usedIngredients := make([]dto.CookedIngredient, len(cookIngredients))
	for i, ci := range cookIngredients {
		pricePerUnit := 0.0
		if ci.PricePerUnit != nil {
			pricePerUnit = *ci.PricePerUnit
		}
		totalCost := 0.0
		if ci.TotalCost != nil {
			totalCost = *ci.TotalCost
		}
		
		usedIngredients[i] = dto.CookedIngredient{
			IngredientID:    ci.IngredientID,
			Name:            ci.Ingredient.Name,
			QuantityUsed:    ci.QuantityUsed,
			Unit:            ci.Unit,
			PricePerUnit:    pricePerUnit,
			TotalCost:       totalCost,
			WasExpiringSoon: ci.WasExpiringSoon,
		}
	}

	// Count remaining fridge items
	var remainingCount int64
	s.db.Model(&models.UserFridgeItem{}).Where("user_id = ?", cookLog.UserID).Count(&remainingCount)

	return &dto.CookRecipeData{
		CookLogID:          cookLog.ID.String(),
		RecipeID:           recipe.ID.String(),
		CanonicalName:      recipe.CanonicalName,
		LocalName:          recipe.LocalName,
		ServingsMultiplier: cookLog.ServingsMultiplier,
		CookedAt:           cookLog.CookedAt.Format(time.RFC3339),
		UsedValue:          cookLog.UsedValue,
		WasteRiskSaved:     cookLog.WasteRiskSaved,
		TotalRecipeCost:    cookLog.TotalRecipeCost,
		IngredientsUsed:    usedIngredients,
		FridgeUpdated:      true,
		RemainingItems:     int(remainingCount),
	}, nil
}
