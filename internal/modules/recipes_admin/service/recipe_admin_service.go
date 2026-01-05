package service

import (
	"errors"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/recipes_admin/dto"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// RecipeAdminService - Admin service для управления рецептами
type RecipeAdminService struct {
	db *gorm.DB
}

// NewRecipeAdminService - Constructor
func NewRecipeAdminService() *RecipeAdminService {
	return &RecipeAdminService{
		db: database.GetDB(),
	}
}

// CreateDraft - Создать draft рецепт (минимальная валидация)
func (s *RecipeAdminService) CreateDraft(authorID string, req *dto.CreateRecipeRequest) (*models.Recipe, error) {
	// Defaults
	country := req.Country
	if country == "" {
		country = "PL" // Default country
	}
	
	timeMinutes := req.TimeMinutes
	if timeMinutes == 0 {
		timeMinutes = 30 // Default time
	}
	
	servings := req.Servings
	if servings == 0 {
		servings = 1 // Default servings
	}

	recipe := &models.Recipe{
		ID:            uuid.New().String(),
		LocalName:     req.LocalName,                              // Required: display name
		Title:         req.LocalName,                              // Sync title with localName
		CanonicalName: req.CanonicalName,                          // Optional: slug
		Description:   req.Description,                            // Optional
		ImageUrl:      req.ImageUrl,                               // Optional
		Country:       country,                                    // Default: PL
		Category:      req.Category,                               // Required
		Difficulty:    req.Difficulty,                             // Required
		TimeMinutes:   timeMinutes,                                // Default: 30
		Servings:      servings,                                   // Default: 1
		Source:        datatypes.JSON([]byte(`{"type":"manual"}`)), // Backend controlled
		Status:        "draft",                                    // Backend controlled (КРИТИЧНО)
		AuthorID:      authorID,                                   // From JWT
		GrossWeight:   req.GrossWeight,                            // Optional
		NetWeight:     req.NetWeight,                              // Optional
		Calories:      req.Calories,                               // Optional
		Protein:       req.Protein,                                // Optional
		Fats:          req.Fats,                                   // Optional
		Carbs:         req.Carbs,                                  // Optional
		TokensReward:  intPtr(10),                                 // Default
		ViewsCount:    0,
		TokensEarned:  0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.db.Create(recipe).Error; err != nil {
		return nil, err
	}

	// Preload author
	s.db.Preload("Author").First(recipe, "id = ?", recipe.ID)
	return recipe, nil
}

// UpdateDraft - Обновить draft рецепт (только если status = draft)
func (s *RecipeAdminService) UpdateDraft(recipeID string, req *dto.UpdateRecipeRequest) (*models.Recipe, error) {
	var recipe models.Recipe
	if err := s.db.First(&recipe, "id = ?", recipeID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("recipe not found")
		}
		return nil, err
	}

	// КРИТИЧНО: можно обновлять только draft
	if recipe.Status != "draft" {
		return nil, errors.New("can only update draft recipes")
	}

	// Update только переданные поля
	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
		updates["localName"] = *req.Title // Sync localName
	}
	if req.CanonicalName != nil {
		updates["canonicalName"] = *req.CanonicalName
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.ImageUrl != nil {
		updates["imageUrl"] = *req.ImageUrl
	}
	if req.Country != nil {
		updates["country"] = *req.Country
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.Difficulty != nil {
		updates["difficulty"] = *req.Difficulty
	}
	if req.TimeMinutes != nil {
		updates["timeMinutes"] = *req.TimeMinutes
	}
	if req.Servings != nil {
		updates["servings"] = *req.Servings
	}
	if req.GrossWeight != nil {
		updates["gross_weight"] = *req.GrossWeight
	}
	if req.NetWeight != nil {
		updates["net_weight"] = *req.NetWeight
	}
	if req.Calories != nil {
		updates["calories"] = *req.Calories
	}
	if req.Protein != nil {
		updates["protein"] = *req.Protein
	}
	if req.Fats != nil {
		updates["fats"] = *req.Fats
	}
	if req.Carbs != nil {
		updates["carbs"] = *req.Carbs
	}

	updates["updatedAt"] = time.Now()

	if err := s.db.Model(&recipe).Updates(updates).Error; err != nil {
		return nil, err
	}

	// Reload
	s.db.Preload("Author").First(&recipe, "id = ?", recipeID)
	return &recipe, nil
}

// Publish - Публикация рецепта с ПОЛНОЙ валидацией
func (s *RecipeAdminService) Publish(recipeID string, req *dto.PublishRecipeRequest) (*models.Recipe, []string, error) {
	var recipe models.Recipe
	if err := s.db.First(&recipe, "id = ?", recipeID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("recipe not found")
		}
		return nil, nil, err
	}

	// Можно публиковать только draft или archived
	if recipe.Status != "draft" && recipe.Status != "archived" {
		return nil, nil, errors.New("can only publish draft or archived recipes")
	}

	// ВАЛИДАЦИЯ
	warnings := []string{}

	// 1. Проверка ingredients (REQUIRED)
	if len(req.Ingredients) == 0 {
		if !req.Force {
			return nil, nil, errors.New("at least 1 ingredient required for publishing")
		}
		warnings = append(warnings, "No ingredients specified")
	}

	// 2. Проверка steps (REQUIRED)
	if len(req.Steps) == 0 {
		if !req.Force {
			return nil, nil, errors.New("at least 1 step required for publishing")
		}
		warnings = append(warnings, "No cooking steps specified")
	}

	// 3. Проверка description (WARNING, not blocking)
	if recipe.Description == "" {
		warnings = append(warnings, "Missing description")
	}

	// 4. Проверка title length
	if len(recipe.Title) < 3 {
		if !req.Force {
			return nil, nil, errors.New("title must be at least 3 characters")
		}
		warnings = append(warnings, "Title is too short")
	}

	// 5. Проверка nutrition (warning only)
	if recipe.Calories == nil || *recipe.Calories == 0 {
		warnings = append(warnings, "Missing nutrition information")
	}

	// 6. Проверка порядка steps
	for i, step := range req.Steps {
		if step.Order != i+1 {
			if !req.Force {
				return nil, nil, errors.New("steps must be in sequential order starting from 1")
			}
			warnings = append(warnings, "Steps are not in sequential order")
			break
		}
	}

	// TODO: Save ingredients and steps to RecipeCatalog table or related tables
	// This depends on your ingredients/steps storage strategy

	// Обновляем status
	if err := s.db.Model(&recipe).Updates(map[string]interface{}{
		"status":    "published",
		"updatedAt": time.Now(),
	}).Error; err != nil {
		return nil, nil, err
	}

	// Reload
	s.db.Preload("Author").First(&recipe, "id = ?", recipeID)
	return &recipe, warnings, nil
}

// Archive - Архивировать рецепт
func (s *RecipeAdminService) Archive(recipeID string) error {
	return s.db.Model(&models.Recipe{}).Where("id = ?", recipeID).Update("status", "archived").Error
}

// GetDrafts - Получить все draft рецепты
func (s *RecipeAdminService) GetDrafts(authorID string) ([]models.Recipe, error) {
	var recipes []models.Recipe
	err := s.db.Where("author_id = ? AND status = ?", authorID, "draft").
		Preload("Author").
		Order("updated_at DESC").
		Find(&recipes).Error
	return recipes, err
}

// Helper
func intPtr(val int) *int {
	return &val
}
