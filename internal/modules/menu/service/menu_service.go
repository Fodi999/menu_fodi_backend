package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/menu/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================================================
// CUSTOM ERRORS
// ============================================================================

// InsufficientIngredientsError - недостаточно ингредиентов в холодильнике
type InsufficientIngredientsError struct {
	RecipeID           uuid.UUID
	MissingIngredients []string
}

func (e *InsufficientIngredientsError) Error() string {
	return fmt.Sprintf("cannot add to menu: missing ingredients: %s", strings.Join(e.MissingIngredients, ", "))
}

// ============================================================================
// SERVICE: Menu (Kitchen Pipeline)
// Business logic: what can user cook, when, and with what ingredients
// ============================================================================

type MenuService struct {
	menuRepo *repository.MenuRepository
	db       *gorm.DB // For checking fridge ingredients
}

func NewMenuService(menuRepo *repository.MenuRepository, db *gorm.DB) *MenuService {
	return &MenuService{
		menuRepo: menuRepo,
		db:       db,
	}
}

// ============================================================================
// PUBLIC API METHODS
// ============================================================================

// GetTodayMenu - получить меню на сегодня
func (s *MenuService) GetTodayMenu(ctx context.Context, userID string, lang string) ([]models.MenuItemResponse, error) {
	// Get raw menu items from repository
	items, err := s.menuRepo.GetTodayMenu(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get today menu: %w", err)
	}
	
	// Transform to response DTOs
	responses := make([]models.MenuItemResponse, 0, len(items))
	for _, item := range items {
		response := s.buildMenuItemResponse(item, lang)
		responses = append(responses, response)
	}
	
	return responses, nil
}

// AddToMenu - добавить рецепт в меню (с валидацией!)
func (s *MenuService) AddToMenu(
	ctx context.Context,
	userID string,
	recipeID uuid.UUID,
	servings int,
	notes *string,
) (*models.MenuItemResponse, error) {
	// Validate servings
	if servings < 1 {
		servings = 1
	}
	if servings > 10 {
		return nil, fmt.Errorf("servings must be between 1 and 10")
	}
	
	// ✅ CRITICAL: Check if user can cook this recipe NOW (Backend = Source of Truth)
	canCook, missingIngredients, err := s.checkCanCookNow(ctx, userID, recipeID)
	if err != nil {
		return nil, fmt.Errorf("failed to check ingredients: %w", err)
	}
	
	if !canCook {
		return nil, &InsufficientIngredientsError{
			RecipeID:           recipeID,
			MissingIngredients: missingIngredients,
		}
	}
	
	// Create menu item
	item := &models.UserMenuItem{
		UserID:     userID,
		RecipeID:   recipeID,
		Servings:   servings,
		Status:     models.MenuItemPlanned,
		PlannedFor: time.Now(),
		Notes:      notes,
	}
	
	// Save to database
	if err := s.menuRepo.AddToMenu(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to add to menu: %w", err)
	}
	
	// Reload with relations
	savedItem, err := s.menuRepo.GetMenuItem(ctx, item.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload menu item: %w", err)
	}
	
	// Build response
	response := s.buildMenuItemResponse(*savedItem, "pl") // TODO: pass actual lang
	return &response, nil
}

// StartCooking - начать готовить
func (s *MenuService) StartCooking(ctx context.Context, itemID uuid.UUID, servings *int) error {
	// Update servings if provided
	if servings != nil && *servings > 0 {
		if err := s.menuRepo.UpdateServings(ctx, itemID, *servings); err != nil {
			return fmt.Errorf("failed to update servings: %w", err)
		}
	}
	
	// Change status to cooking
	if err := s.menuRepo.StartCooking(ctx, itemID); err != nil {
		return fmt.Errorf("failed to start cooking: %w", err)
	}
	
	return nil
}

// CompleteCooking - завершить готовку (создает prepared_dish)
func (s *MenuService) CompleteCooking(ctx context.Context, itemID uuid.UUID, actualServings *int) error {
	// Get menu item
	item, err := s.menuRepo.GetMenuItem(ctx, itemID)
	if err != nil {
		return fmt.Errorf("menu item not found: %w", err)
	}
	
	// Validate status
	if item.Status != models.MenuItemCooking {
		return fmt.Errorf("menu item is not in cooking status")
	}
	
	// Use actual servings if provided
	servings := item.Servings
	if actualServings != nil && *actualServings > 0 {
		servings = *actualServings
	}
	
	// Mark as completed
	if err := s.menuRepo.CompleteCooking(ctx, itemID); err != nil {
		return fmt.Errorf("failed to complete cooking: %w", err)
	}
	
	// TODO: Create prepared_dish record
	// TODO: Deduct ingredients from fridge
	
	fmt.Printf("✅ [MENU] Completed cooking: recipe=%s, servings=%d\n", 
		item.RecipeID, servings)
	
	return nil
}

// CancelMenuItem - отменить приготовление
func (s *MenuService) CancelMenuItem(ctx context.Context, itemID uuid.UUID) error {
	return s.menuRepo.CancelMenuItem(ctx, itemID)
}

// DeleteMenuItem - удалить из меню
func (s *MenuService) DeleteMenuItem(ctx context.Context, itemID uuid.UUID) error {
	return s.menuRepo.DeleteMenuItem(ctx, itemID)
}

// ============================================================================
// HELPER METHODS
// ============================================================================

// buildMenuItemResponse - собирает DTO для API response
func (s *MenuService) buildMenuItemResponse(item models.UserMenuItem, lang string) models.MenuItemResponse {
	response := models.MenuItemResponse{
		ID:         item.ID.String(),
		Servings:   item.Servings,
		Status:     item.Status,
		PlannedFor: item.PlannedFor.Format("2006-01-02"),
		CreatedAt:  item.CreatedAt.Format(time.RFC3339),
		Notes:      item.Notes,
	}
	
	// Add timestamps if present
	if item.StartedCookingAt != nil {
		t := item.StartedCookingAt.Format(time.RFC3339)
		response.StartedCookingAt = &t
	}
	if item.CompletedAt != nil {
		t := item.CompletedAt.Format(time.RFC3339)
		response.CompletedAt = &t
	}
	
	// Add recipe details
	if item.Recipe != nil {
		response.Recipe = models.RecipeBasicInfo{
			ID:            item.Recipe.ID.String(),
			Title:         item.Recipe.GetLocalizedName(lang),
			CanonicalName: item.Recipe.CanonicalName,
			ImageURL:      &item.Recipe.ImageUrl,
			CookTime:      item.Recipe.TimeMinutes,
			Servings:      item.Recipe.Servings,
		}
	}
	
	return response
}

// ============================================================================
// PRIVATE HELPER: Check if user can cook recipe NOW
// ============================================================================

// checkCanCookNow - проверяет, есть ли все ингредиенты в холодильнике
func (s *MenuService) checkCanCookNow(ctx context.Context, userID string, recipeID uuid.UUID) (bool, []string, error) {
	// 1. Get recipe ingredients
	var recipeIngredients []models.CatalogIngredient
	err := s.db.WithContext(ctx).
		Preload("Ingredient").
		Where("recipeId = ?", recipeID).
		Find(&recipeIngredients).Error
	if err != nil {
		return false, nil, fmt.Errorf("failed to load recipe ingredients: %w", err)
	}
	
	if len(recipeIngredients) == 0 {
		// No ingredients required = can cook
		return true, nil, nil
	}
	
	// 2. Get user's fridge ingredients (with canonical_id)
	var fridgeItems []models.FridgeItem
	err = s.db.WithContext(ctx).
		Preload("Ingredient").
		Where("user_id = ? AND quantity > 0", userID).
		Find(&fridgeItems).Error
	if err != nil {
		return false, nil, fmt.Errorf("failed to load fridge items: %w", err)
	}
	
	// Build set of available canonical IDs from fridge
	availableCanonicalIDs := make(map[string]bool)
	for _, item := range fridgeItems {
		if item.Ingredient != nil && item.Ingredient.CanonicalID != nil {
			availableCanonicalIDs[*item.Ingredient.CanonicalID] = true
		}
		// Also add ingredient ID itself (for non-canonical ingredients)
		if item.IngredientID != "" {
			availableCanonicalIDs[item.IngredientID] = true
		}
	}
	
	// 3. Check each recipe ingredient
	var missing []string
	for _, recipeIng := range recipeIngredients {
		if recipeIng.Ingredient.ID == "" {
			continue // Skip if no ingredient data
		}
		
		// Check if available: either by ingredient ID or by canonical ID
		ingredientID := recipeIng.Ingredient.ID
		canonicalID := recipeIng.Ingredient.CanonicalID
		
		hasIngredient := availableCanonicalIDs[ingredientID]
		if canonicalID != nil {
			hasIngredient = hasIngredient || availableCanonicalIDs[*canonicalID]
		}
		
		if !hasIngredient {
			// Get localized name for error message
			ingredientName := recipeIng.Ingredient.Name
			if recipeIng.Ingredient.NameRU != nil && *recipeIng.Ingredient.NameRU != "" {
				ingredientName = *recipeIng.Ingredient.NameRU
			}
			missing = append(missing, ingredientName)
		}
	}
	
	// 4. Return result
	if len(missing) > 0 {
		return false, missing, nil
	}
	
	return true, nil, nil
}
