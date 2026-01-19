package service

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/recipes_admin/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// RecipeAdminService - Admin service для управления рецептами
type RecipeAdminService struct {
	db        *gorm.DB
	aiService AITranslator // Interface for AI translation
}

// AITranslator interface for recipe translation
type AITranslator interface {
	TranslateRecipeField(fieldType, text, sourceLang string) (pl, en, ru string, err error)
}

// NewRecipeAdminService - Constructor
func NewRecipeAdminService() *RecipeAdminService {
	return &RecipeAdminService{
		db:        database.GetDB(),
		aiService: nil, // AI service will be injected later if needed
	}
}

// SetAIService sets the AI translation service
func (s *RecipeAdminService) SetAIService(ai AITranslator) {
	s.aiService = ai
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

	// Нормализация текста (исправление опечаток, капитализация)
	localName := utils.CleanRecipeText(req.LocalName)
	title := utils.CapitalizeTitle(req.LocalName)
	description := utils.CleanRecipeText(req.Description)

	// Генерация канонического имени (slug для URL)
	var canonicalName string
	if req.CanonicalName != nil && *req.CanonicalName != "" {
		canonicalName = *req.CanonicalName
	} else {
		// Автоматическая генерация: "Яичница глазунья" → "yaichnitsa_glazunya"
		canonicalName = utils.GenerateCanonicalName(localName)
	}

	// Multi-language support - copy from request if provided
	var stepsPLJSON, stepsENJSON, stepsRUJSON datatypes.JSON
	if req.StepsPL != nil && len(*req.StepsPL) > 0 {
		// Нормализуем шаги (капитализация)
		normalizedSteps := utils.CapitalizeSteps(*req.StepsPL)
		stepsPLBytes, _ := json.Marshal(normalizedSteps)
		stepsPLJSON = datatypes.JSON(stepsPLBytes)
	}
	if req.StepsEN != nil && len(*req.StepsEN) > 0 {
		normalizedSteps := utils.CapitalizeSteps(*req.StepsEN)
		stepsENBytes, _ := json.Marshal(normalizedSteps)
		stepsENJSON = datatypes.JSON(stepsENBytes)
	}
	if req.StepsRU != nil && len(*req.StepsRU) > 0 {
		normalizedSteps := utils.CapitalizeSteps(*req.StepsRU)
		stepsRUBytes, _ := json.Marshal(normalizedSteps)
		stepsRUJSON = datatypes.JSON(stepsRUBytes)
	}

	// Нормализуем переводы названий и описаний
	var namePL, nameEN, nameRU, descPL, descEN, descRU *string
	if req.NamePL != nil {
		cleaned := utils.CleanRecipeText(*req.NamePL)
		namePL = &cleaned
	}
	if req.NameEN != nil {
		cleaned := utils.CleanRecipeText(*req.NameEN)
		nameEN = &cleaned
	}
	if req.NameRU != nil {
		cleaned := utils.CleanRecipeText(*req.NameRU)
		nameRU = &cleaned
	}
	if req.DescriptionPL != nil {
		cleaned := utils.CleanRecipeText(*req.DescriptionPL)
		descPL = &cleaned
	}
	if req.DescriptionEN != nil {
		cleaned := utils.CleanRecipeText(*req.DescriptionEN)
		descEN = &cleaned
	}
	if req.DescriptionRU != nil {
		cleaned := utils.CleanRecipeText(*req.DescriptionRU)
		descRU = &cleaned
	}

	recipe := &models.Recipe{
		ID:            uuid.New().String(),
		LocalName:     localName,                                   // Нормализованное имя
		Title:         title,                                       // Капитализированный заголовок
		CanonicalName: &canonicalName,                              // Slug (yaichnitsa_glazunya)
		Description:   description,                                 // Нормализованное описание
		ImageUrl:      req.ImageUrl,                                // Optional
		Country:       country,                                     // Default: PL
		Category:      req.Category,                                // Required
		Difficulty:    req.Difficulty,                              // Required
		TimeMinutes:   timeMinutes,                                 // Default: 30
		Servings:      servings,                                    // Default: 1
		Source:        datatypes.JSON([]byte(`{"type":"manual"}`)), // Backend controlled
		Status:        "draft",                                     // Backend controlled (КРИТИЧНО)
		AuthorID:      authorID,                                    // From JWT
		GrossWeight:   req.GrossWeight,                             // Optional
		NetWeight:     req.NetWeight,                               // Optional
		Calories:      req.Calories,                                // Optional
		Protein:       req.Protein,                                 // Optional
		Fats:          req.Fats,                                    // Optional
		Carbs:         req.Carbs,                                   // Optional
		TokensReward:  intPtr(10),                                  // Default
		ViewsCount:    0,
		TokensEarned:  0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		// Multi-language fields (нормализованные)
		NamePL:        namePL,
		NameEN:        nameEN,
		NameRU:        nameRU,
		DescriptionPL: descPL,
		DescriptionEN: descEN,
		DescriptionRU: descRU,
		StepsPL:       stepsPLJSON,
		StepsEN:       stepsENJSON,
		StepsRU:       stepsRUJSON,
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

	// 7. АВТОМАТИЧЕСКИЙ ПЕРЕВОД на 3 языка (PL/EN/RU) если переводы отсутствуют
	translationWarnings := s.ensureTranslations(&recipe)
	warnings = append(warnings, translationWarnings...)

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

// ensureTranslations проверяет наличие переводов рецепта на все 3 языка
// Если переводы отсутствуют, пытается создать их через AI
// Возвращает список предупреждений
func (s *RecipeAdminService) ensureTranslations(recipe *models.Recipe) []string {
	warnings := []string{}

	// Проверяем наличие базовых переводов названия
	needsTranslation := recipe.NamePL == nil || *recipe.NamePL == "" ||
		recipe.NameEN == nil || *recipe.NameEN == "" ||
		recipe.NameRU == nil || *recipe.NameRU == ""

	if !needsTranslation {
		// Переводы уже есть
		return warnings
	}

	// Пытаемся перевести через AI
	if s.aiService == nil {
		warnings = append(warnings, "AI translation service not available - recipe published without translations")
		return warnings
	}

	// Определяем исходный язык и текст для перевода
	sourceLang := "unknown"
	sourceTitle := recipe.Title
	if sourceTitle == "" {
		sourceTitle = recipe.LocalName
	}

	// Переводим название
	if sourceTitle != "" {
		plName, enName, ruName, err := s.aiService.TranslateRecipeField("recipe name", sourceTitle, sourceLang)
		if err != nil {
			warnings = append(warnings, "Failed to auto-translate recipe name: "+err.Error())
		} else {
			recipe.NamePL = &plName
			recipe.NameEN = &enName
			recipe.NameRU = &ruName

			// Сохраняем переводы названия
			s.db.Model(recipe).Updates(map[string]interface{}{
				"name_pl": plName,
				"name_en": enName,
				"name_ru": ruName,
			})
		}
	}

	// Переводим описание, если оно есть
	if recipe.Description != "" {
		plDesc, enDesc, ruDesc, err := s.aiService.TranslateRecipeField("recipe description", recipe.Description, sourceLang)
		if err != nil {
			warnings = append(warnings, "Failed to auto-translate description: "+err.Error())
		} else {
			recipe.DescriptionPL = &plDesc
			recipe.DescriptionEN = &enDesc
			recipe.DescriptionRU = &ruDesc

			// Сохраняем переводы описания
			s.db.Model(recipe).Updates(map[string]interface{}{
				"description_pl": plDesc,
				"description_en": enDesc,
				"description_ru": ruDesc,
			})
		}
	}

	if len(warnings) == 0 {
		warnings = append(warnings, "Recipe auto-translated to PL/EN/RU")
	}

	return warnings
}
