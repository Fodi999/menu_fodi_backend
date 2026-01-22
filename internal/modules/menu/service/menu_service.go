
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/menu/repository"
	"github.com/google/uuid"
)

// ============================================================================
// SERVICE: Menu (Kitchen Pipeline)
// Business logic: what can user cook, when, and with what ingredients
// ============================================================================

type MenuService struct {
	menuRepo *repository.MenuRepository
}

func NewMenuService(menuRepo *repository.MenuRepository) *MenuService {
	return &MenuService{
		menuRepo: menuRepo,
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
