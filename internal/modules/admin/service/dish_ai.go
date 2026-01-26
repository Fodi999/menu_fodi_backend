package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ===========================
// Dish Generation DTOs
// ===========================

// GenerateDishRequest - запрос на генерацию карточки блюда
type GenerateDishRequest struct {
	RecipeID     string  `json:"recipeId" binding:"required"`
	TargetMargin float64 `json:"targetMargin" binding:"required,min=0,max=100"` // Целевая маржа (%)
	Language     string  `json:"language"`                                      // "pl", "en", "ru" (default "en")
}

// DishAIResponse - ответ от AI для карточки блюда (многоязычный)
type DishAIResponse struct {
	// Primary language (от пользователя)
	Title       string `json:"title"`       // Привлекательное название для меню
	Description string `json:"description"` // Продающее описание (2-3 предложения)
	
	// Translations
	TitlePl       *string `json:"titlePl,omitempty"`
	TitleEn       *string `json:"titleEn,omitempty"`
	TitleRu       *string `json:"titleRu,omitempty"`
	
	DescriptionPl *string `json:"descriptionPl,omitempty"`
	DescriptionEn *string `json:"descriptionEn,omitempty"`
	DescriptionRu *string `json:"descriptionRu,omitempty"`
}

// ===========================
// Service Interface Extension
// ===========================

// Добавляем методы в AdminService interface (в service.go)
// GenerateDishWithAI(req GenerateDishRequest, adminID string) (*models.Dish, error)

// ===========================
// Implementation
// ===========================

// GenerateDishWithAI создаёт карточку блюда через AI
// Архитектура: копируем Recipe pipeline → адаптируем для Dish
func (s *adminService) GenerateDishWithAI(req GenerateDishRequest, adminID string) (*models.Dish, error) {
	ctx := context.Background()
	
	// 1️⃣ Загружаем рецепт из каталога
	recipe, err := s.loadRecipeForDish(req.RecipeID)
	if err != nil {
		return nil, fmt.Errorf("recipe not found: %w", err)
	}
	
	logger.Info("Generating dish from recipe",
		zap.String("recipe_id", req.RecipeID),
		zap.String("recipe_title", recipe.Title),
		zap.Float64("target_margin", req.TargetMargin),
	)
	
	// 2️⃣ Рассчитываем себестоимость блюда (snapshot из ингредиентов)
	cost, err := s.calculateDishCost(ctx, recipe)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate cost: %w", err)
	}
	
	logger.Info("Dish cost calculated",
		zap.String("recipe_id", req.RecipeID),
		zap.Float64("cost", cost),
	)
	
	// 3️⃣ Рассчитываем цену на основе маржи
	// Формула: Price = Cost / (1 - Margin/100)
	// Пример: Cost = 10 PLN, Margin = 60% → Price = 10 / (1 - 0.6) = 25 PLN
	price := s.calculatePrice(cost, req.TargetMargin)
	
	logger.Info("Dish price calculated",
		zap.String("recipe_id", req.RecipeID),
		zap.Float64("price", price),
		zap.Float64("margin", req.TargetMargin),
	)
	
	// 4️⃣ Генерируем AI контент (название и описание для меню)
	lang := req.Language
	if lang == "" {
		lang = "en"
	}
	
	aiContent, err := s.generateDishContentViaAI(ctx, recipe, cost, price, req.TargetMargin, lang)
	if err != nil {
		logger.Warn("AI generation failed, using fallback content",
			zap.Error(err),
		)
		// Используем fallback контент
		aiContent = s.generateFallbackDishContent(recipe, lang)
	}
	
	// 5️⃣ Сохраняем блюдо как draft (с многоязычным контентом)
	dish := &models.Dish{
		ID:          uuid.New(),
		RecipeID:    uuid.MustParse(req.RecipeID),
		Title:       aiContent.Title,
		TitlePl:     aiContent.TitlePl,
		TitleEn:     aiContent.TitleEn,
		TitleRu:     aiContent.TitleRu,
		Description: aiContent.Description,
		DescriptionPl: aiContent.DescriptionPl,
		DescriptionEn: aiContent.DescriptionEn,
		DescriptionRu: aiContent.DescriptionRu,
		ImageURL:    recipe.ImageUrl, // Используем изображение рецепта
		Cost:        cost,
		Price:       price,
		Margin:      req.TargetMargin,
		Status:      models.DishStatusDraft,
		IsAvailable: true,
		CreatedBy:   adminID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	if err := s.db.Create(dish).Error; err != nil {
		return nil, fmt.Errorf("failed to save dish: %w", err)
	}
	
	// 6️⃣ Логируем событие в history
	s.logDishEvent(adminID, dish.ID.String(), "dish_created", map[string]interface{}{
		"recipe_id":     req.RecipeID,
		"recipe_title":  recipe.Title,
		"cost":          cost,
		"price":         price,
		"margin":        req.TargetMargin,
		"ai_generated":  true,
	})
	
	logger.Info("Dish created successfully",
		zap.String("dish_id", dish.ID.String()),
		zap.String("recipe_id", req.RecipeID),
		zap.String("title", dish.Title),
		zap.Float64("cost", cost),
		zap.Float64("price", price),
		zap.Float64("margin", req.TargetMargin),
	)
	
	return dish, nil
}

// ===========================
// Helper Methods
// ===========================

// loadRecipeForDish загружает рецепт с ингредиентами
func (s *adminService) loadRecipeForDish(recipeID string) (*models.RecipeCatalog, error) {
	var recipe models.RecipeCatalog
	
	err := s.db.
		Preload("Ingredients").
		Preload("Ingredients.Ingredient").
		First(&recipe, "id = ?", recipeID).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("recipe not found")
		}
		return nil, fmt.Errorf("failed to load recipe: %w", err)
	}
	
	return &recipe, nil
}

// calculateDishCost рассчитывает себестоимость блюда на основе ингредиентов
// Snapshot = цены фиксируются на момент создания
func (s *adminService) calculateDishCost(ctx context.Context, recipe *models.RecipeCatalog) (float64, error) {
	totalCost := 0.0
	
	for _, ingredient := range recipe.Ingredients {
		// Пропускаем optional ингредиенты
		if ingredient.Optional {
			logger.Debug("Skipping optional ingredient",
				zap.String("ingredient_id", ingredient.IngredientID),
			)
			continue
		}
		
		// Получаем актуальную цену из каталога
		pricePerUnit := 0.0
		if ingredient.Ingredient.DefaultPricePerUnit != nil {
			pricePerUnit = *ingredient.Ingredient.DefaultPricePerUnit
		} else {
			logger.Warn("Ingredient has no price, using 0",
				zap.String("ingredient_id", ingredient.IngredientID),
				zap.String("ingredient_name", ingredient.Ingredient.Name),
			)
		}
		
		// Рассчитываем стоимость ингредиента
		ingredientCost := ingredient.Quantity * pricePerUnit
		totalCost += ingredientCost
		
		logger.Debug("Ingredient cost",
			zap.String("ingredient_id", ingredient.IngredientID),
			zap.String("ingredient_name", ingredient.Ingredient.Name),
			zap.Float64("quantity", ingredient.Quantity),
			zap.String("unit", ingredient.Unit),
			zap.Float64("price_per_unit", pricePerUnit),
			zap.Float64("ingredient_cost", ingredientCost),
		)
	}
	
	// Округляем до 2 знаков после запятой
	totalCost = math.Round(totalCost*100) / 100
	
	logger.Info("Total dish cost calculated",
		zap.String("recipe_id", recipe.ID.String()),
		zap.Float64("total_cost", totalCost),
		zap.Int("ingredients_count", len(recipe.Ingredients)),
	)
	
	return totalCost, nil
}

// calculatePrice рассчитывает цену на основе себестоимости и маржи
// Формула: Price = Cost / (1 - Margin/100)
func (s *adminService) calculatePrice(cost float64, marginPercent float64) float64 {
	if cost == 0 {
		return 0
	}
	
	// Формула обратного расчёта
	// Если Cost = 10, Margin = 60%, то Price = 10 / (1 - 0.6) = 25
	price := cost / (1 - marginPercent/100)
	
	// Округляем до 2 знаков
	price = math.Round(price*100) / 100
	
	return price
}

// generateDishContentViaAI вызывает AI для генерации привлекательного контента на ВСЕХ ЯЗЫКАХ
func (s *adminService) generateDishContentViaAI(
	ctx context.Context,
	recipe *models.RecipeCatalog,
	cost, price, margin float64,
	language string,
) (*DishAIResponse, error) {
	// Генерируем контент на основном языке
	primaryContent := s.generateFallbackDishContent(recipe, language)
	
	// Генерируем на остальных двух языках
	response := &DishAIResponse{
		Title:       primaryContent.Title,
		Description: primaryContent.Description,
	}
	
	// Устанавливаем основной язык
	switch language {
	case "pl":
		response.TitlePl = &response.Title
		response.DescriptionPl = &response.Description
		// Генерируем переводы на EN и RU
		enContent := s.generateFallbackDishContent(recipe, "en")
		response.TitleEn = &enContent.Title
		response.DescriptionEn = &enContent.Description
		ruContent := s.generateFallbackDishContent(recipe, "ru")
		response.TitleRu = &ruContent.Title
		response.DescriptionRu = &ruContent.Description
		
	case "ru":
		response.TitleRu = &response.Title
		response.DescriptionRu = &response.Description
		// Генерируем переводы на PL и EN
		plContent := s.generateFallbackDishContent(recipe, "pl")
		response.TitlePl = &plContent.Title
		response.DescriptionPl = &plContent.Description
		enContent := s.generateFallbackDishContent(recipe, "en")
		response.TitleEn = &enContent.Title
		response.DescriptionEn = &enContent.Description
		
	default: // "en"
		response.TitleEn = &response.Title
		response.DescriptionEn = &response.Description
		// Генерируем переводы на PL и RU
		plContent := s.generateFallbackDishContent(recipe, "pl")
		response.TitlePl = &plContent.Title
		response.DescriptionPl = &plContent.Description
		ruContent := s.generateFallbackDishContent(recipe, "ru")
		response.TitleRu = &ruContent.Title
		response.DescriptionRu = &ruContent.Description
	}
	
	logger.Debug("Dish content generated for all languages",
		zap.String("primary_language", language),
		zap.String("title_pl", *response.TitlePl),
		zap.String("title_en", *response.TitleEn),
		zap.String("title_ru", *response.TitleRu),
	)
	
	return response, nil
}

// buildDishPrompt формирует prompt для AI
func (s *adminService) buildDishPrompt(
	recipe *models.RecipeCatalog,
	cost, price, margin float64,
	language string,
) string {
	// Получаем локализованное название рецепта
	recipeName := recipe.Title
	if language == "pl" && recipe.NamePl != nil {
		recipeName = *recipe.NamePl
	} else if language == "en" && recipe.NameEn != nil {
		recipeName = *recipe.NameEn
	} else if language == "ru" && recipe.NameRu != nil {
		recipeName = *recipe.NameRu
	}
	
	// Список ингредиентов для контекста
	ingredientsList := ""
	for i, ing := range recipe.Ingredients {
		if !ing.Optional {
			if i > 0 {
				ingredientsList += ", "
			}
			ingredientsList += ing.Ingredient.Name
		}
	}
	
	langName := map[string]string{
		"pl": "Polish",
		"en": "English",
		"ru": "Russian",
	}[language]
	
	return fmt.Sprintf(`You are a professional restaurant menu writer and marketing expert.

TASK: Create a compelling dish card for a restaurant menu that will make customers want to order it.

RECIPE INFORMATION:
- Name: %s
- Category: %s
- Cooking Time: %d minutes
- Difficulty: %s
- Main Ingredients: %s

PRICING:
- Cost: %.2f PLN
- Selling Price: %.2f PLN
- Margin: %.0f%%

REQUIREMENTS:
1. Create an attractive dish title (short, appetizing, makes people hungry)
2. Write a compelling description (2-3 sentences maximum):
   - Highlight the key ingredients and their quality
   - Mention the cooking method if interesting
   - Create emotional connection (taste, aroma, texture)
   - Make it sound delicious and worth the price
3. Language: %s
4. Output ONLY valid JSON, no additional text

OUTPUT FORMAT (strict JSON):
{
  "title": "Attractive dish name",
  "description": "Compelling 2-3 sentence description that sells the dish"
}

EXAMPLES:
- Good title: "Grilled Salmon with Lemon Butter Sauce"
- Bad title: "Salmon Recipe #5"

- Good description: "Tender Atlantic salmon fillet, grilled to perfection and topped with our signature lemon butter sauce. Served with aromatic jasmine rice and fresh seasonal vegetables."
- Bad description: "Fish cooked with sauce and rice."

Remember: This is a commercial menu - your words should make customers hungry and eager to order!`,
		recipeName,
		recipe.Category,
		recipe.TimeMinutes,
		recipe.Difficulty,
		ingredientsList,
		cost,
		price,
		margin,
		langName,
	)
}

// generateFallbackDishContent генерирует контент без AI
func (s *adminService) generateFallbackDishContent(recipe *models.RecipeCatalog, language string) *DishAIResponse {
	// Используем название рецепта
	title := recipe.Title
	
	// Генерируем простое описание
	description := fmt.Sprintf(
		"Delicious %s prepared with fresh ingredients. Cooking time: %d minutes.",
		recipe.Category,
		recipe.TimeMinutes,
	)
	
	if language == "pl" {
		description = fmt.Sprintf(
			"Pyszne danie z kategorii %s przygotowane ze świeżych składników. Czas przygotowania: %d minut.",
			recipe.Category,
			recipe.TimeMinutes,
		)
	} else if language == "ru" {
		description = fmt.Sprintf(
			"Вкусное блюдо категории %s, приготовленное из свежих ингредиентов. Время приготовления: %d минут.",
			recipe.Category,
			recipe.TimeMinutes,
		)
	}
	
	return &DishAIResponse{
		Title:       title,
		Description: description,
	}
}

// logDishEvent логирует событие блюда в history
func (s *adminService) logDishEvent(adminID, dishID, eventType string, metadata map[string]interface{}) {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		logger.Error("Failed to marshal dish event metadata",
			zap.Error(err),
		)
		return
	}
	
	historyEvent := &models.HistoryEvent{
		ID:         uuid.New().String(),
		UserID:     dishID, // Используем dishID вместо userID
		EventType:  models.HistoryEventType(eventType),
		SourceType: "admin",
		SourceID:   &adminID,
		Metadata:   metadataJSON,
		CreatedAt:  time.Now(),
	}
	
	if err := s.db.Create(historyEvent).Error; err != nil {
		logger.Error("Failed to log dish event to history",
			zap.String("dish_id", dishID),
			zap.String("event_type", eventType),
			zap.Error(err),
		)
	} else {
		logger.Debug("Dish event logged to history",
			zap.String("dish_id", dishID),
			zap.String("event_type", eventType),
		)
	}
}
