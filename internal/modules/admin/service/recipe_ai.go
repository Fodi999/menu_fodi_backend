package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// ===========================
// AI Recipe Creation DTOs
// ===========================

// CreateRecipeAIRequest - минимальный человеко-ориентированный запрос
type CreateRecipeAIRequest struct {
	Title          string                    `json:"title"`          // "Лосось с рисом и соусом терияки"
	Ingredients    []RecipeIngredientInput   `json:"ingredients"`    // [{ingredientId, quantity, unit}]
	RawCookingText string                    `json:"rawCookingText"` // "Рыбу замариновать в соусе, обжарить. Рис отварить."
}

// RecipeIngredientInput - ингредиент из запроса
type RecipeIngredientInput struct {
	IngredientID string  `json:"ingredientId"` // UUID ингредиента
	Quantity     float64 `json:"quantity"`     // 150
	Unit         string  `json:"unit"`         // "g"
}

// EnrichedIngredient - обогащенный ингредиент для AI
type EnrichedIngredient struct {
	Name           string  `json:"ingredient"`      // "salmon"
	Quantity       float64 `json:"quantity"`        // 150
	Unit           string  `json:"unit"`            // "g"
	NutritionGroup string  `json:"nutrition_group"` // "protein"
	Category       string  `json:"category"`        // "fish"
}

// AIRecipePromptContext - контекст для AI
type AIRecipePromptContext struct {
	Title           string               `json:"title"`
	Ingredients     []EnrichedIngredient `json:"ingredients"`
	RawCookingText  string               `json:"cooking_instructions"`
}

// AIRecipeResponse - строгий контракт ответа от AI
type AIRecipeResponse struct {
	Summary     string          `json:"summary"`      // "Нежный лосось с ароматным соусом..."
	Servings    int             `json:"servings"`     // 1
	TimeMinutes int             `json:"time_minutes"` // 25
	Difficulty  string          `json:"difficulty"`   // "easy"
	Steps       []RecipeStepAI  `json:"steps"`        // [{order, text, time}]
	Nutrition   RecipeNutrition `json:"nutrition"`    // {calories, protein, fat, carbohydrate}
}

// RecipeStepAI - шаг приготовления от AI
type RecipeStepAI struct {
	Order int    `json:"order"` // 1
	Text  string `json:"text"`  // "Замариновать лосось..."
	Time  int    `json:"time"`  // 5 минут
}

// RecipeNutrition - БЖУ и калории
type RecipeNutrition struct {
	Calories      int     `json:"calories"`      // 520
	Protein       float64 `json:"protein"`       // 38
	Fat           float64 `json:"fat"`           // 22
	Carbohydrate  float64 `json:"carbohydrate"`  // 42
}

// ===========================
// Service Methods
// ===========================

// CreateRecipeWithAI создает рецепт через AI (ЭТАП 1-5)
func (s *adminService) CreateRecipeWithAI(req CreateRecipeAIRequest, authorID string) (*models.RecipeCatalog, error) {
	// ЭТАП 2: Обогащаем данные об ингредиентах из БД
	enrichedIngredients, err := s.enrichIngredientsForAI(req.Ingredients)
	if err != nil {
		return nil, fmt.Errorf("failed to enrich ingredients: %w", err)
	}

	// ЭТАП 3: Формируем контекст для AI
	promptContext := AIRecipePromptContext{
		Title:          req.Title,
		Ingredients:    enrichedIngredients,
		RawCookingText: req.RawCookingText,
	}

	// ЭТАП 3: Вызываем AI для структурирования рецепта
	aiResponse, err := s.generateRecipeViaAI(promptContext)
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	// ЭТАП 5: Сохраняем рецепт в БД
	recipe, err := s.saveRecipeToDB(req, aiResponse, authorID)
	if err != nil {
		return nil, fmt.Errorf("failed to save recipe: %w", err)
	}

	fmt.Printf("✅ Recipe created via AI: %s [%s]\n", recipe.Title, recipe.ID)
	return recipe, nil
}

// PreviewRecipeWithAI возвращает AI-рецепт БЕЗ сохранения (ЭТАП 6)
func (s *adminService) PreviewRecipeWithAI(req CreateRecipeAIRequest) (*AIRecipeResponse, error) {
	// ЭТАП 2: Обогащаем данные об ингредиентах
	enrichedIngredients, err := s.enrichIngredientsForAI(req.Ingredients)
	if err != nil {
		return nil, fmt.Errorf("failed to enrich ingredients: %w", err)
	}

	// ЭТАП 3: Формируем контекст и вызываем AI
	promptContext := AIRecipePromptContext{
		Title:          req.Title,
		Ingredients:    enrichedIngredients,
		RawCookingText: req.RawCookingText,
	}

	aiResponse, err := s.generateRecipeViaAI(promptContext)
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	fmt.Printf("🔍 Recipe preview generated: %s (%d steps)\n", req.Title, len(aiResponse.Steps))
	return aiResponse, nil
}

// ===========================
// ЭТАП 2: Enrichment
// ===========================

// enrichIngredientsForAI загружает ингредиенты из БД и обогащает данными
func (s *adminService) enrichIngredientsForAI(inputs []RecipeIngredientInput) ([]EnrichedIngredient, error) {
	enriched := make([]EnrichedIngredient, 0, len(inputs))

	for _, input := range inputs {
		// Загружаем ингредиент из БД
		var ingredient models.Ingredient
		if err := s.db.Where("id = ?", input.IngredientID).First(&ingredient).Error; err != nil {
			return nil, fmt.Errorf("ingredient %s not found: %w", input.IngredientID, err)
		}

		// Выбираем английское название (приоритет для AI)
		name := ingredient.Name
		if ingredient.NameEN != nil && *ingredient.NameEN != "" {
			name = *ingredient.NameEN
		}

		enriched = append(enriched, EnrichedIngredient{
			Name:           name,
			Quantity:       input.Quantity,
			Unit:           input.Unit,
			NutritionGroup: ingredient.NutritionGroup,
			Category:       ingredient.Category,
		})
	}

	fmt.Printf("🔧 Enriched %d ingredients for AI\n", len(enriched))
	return enriched, nil
}

// ===========================
// ЭТАП 3: AI Generation
// ===========================

// generateRecipeViaAI вызывает Groq AI для структурирования рецепта
func (s *adminService) generateRecipeViaAI(context AIRecipePromptContext) (*AIRecipeResponse, error) {
	systemPrompt := `You are a professional culinary AI assistant.

Given a recipe title, ingredients with nutritional data, and raw cooking instructions, you must:
1. Create a professional summary (1-2 sentences)
2. Break cooking text into clear, numbered steps
3. Estimate time for each step
4. Calculate total cooking time
5. Determine difficulty (easy/medium/hard)
6. Calculate nutrition (calories, protein, fat, carbs) based on ingredients
7. Determine servings (default 1 unless specified)

CRITICAL RULES:
- Respond ONLY with valid JSON (no markdown, no explanations)
- Steps must be actionable and specific
- Time estimates must be realistic
- Nutrition must be calculated from ingredient quantities
- Difficulty: easy (≤30min), medium (30-60min), hard (>60min)

JSON Format:
{
  "summary": "Brief appetizing description",
  "servings": 1,
  "time_minutes": 25,
  "difficulty": "easy",
  "steps": [
    {"order": 1, "text": "Step description", "time": 5}
  ],
  "nutrition": {
    "calories": 520,
    "protein": 38.0,
    "fat": 22.0,
    "carbohydrate": 42.0
  }
}

Nutrition calculation guidelines:
- Fish (protein group): ~150 kcal/100g, 20g protein, 5g fat
- Rice (carbohydrate group): ~130 kcal/100g, 3g protein, 30g carbs
- Vegetables: ~25 kcal/100g, 1g protein, 5g carbs
- Oils/sauces: ~900 kcal/100ml, 100g fat

Do not add explanations. Just JSON.`

	// Формируем user prompt с контекстом
	ingredientsJSON, _ := json.Marshal(context.Ingredients)
	userPrompt := fmt.Sprintf(`Create a structured recipe:

Title: %s

Ingredients:
%s

Cooking Instructions:
%s

Return JSON only.`, context.Title, string(ingredientsJSON), context.RawCookingText)

	fmt.Printf("🤖 Calling AI for recipe: %s\n", context.Title)
	fmt.Printf("📋 Ingredients count: %d\n", len(context.Ingredients))

	// Вызов Groq AI
	response, err := s.groqClient.SimpleChat(systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("groq API call failed: %w", err)
	}

	// Очистка ответа от markdown
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	fmt.Printf("📥 AI Response length: %d chars\n", len(response))

	// Парсинг JSON
	var aiResponse AIRecipeResponse
	if err := json.Unmarshal([]byte(response), &aiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse AI JSON: %w (response: %s)", err, response)
	}

	// ЭТАП 4: Валидация ответа
	if err := validateAIResponse(&aiResponse); err != nil {
		return nil, fmt.Errorf("AI response validation failed: %w", err)
	}

	fmt.Printf("✅ AI generated recipe: %d steps, %d min, %s difficulty\n",
		len(aiResponse.Steps), aiResponse.TimeMinutes, aiResponse.Difficulty)

	return &aiResponse, nil
}

// validateAIResponse проверяет корректность ответа AI
func validateAIResponse(response *AIRecipeResponse) error {
	if response.Summary == "" {
		return fmt.Errorf("summary is empty")
	}
	if response.Servings <= 0 {
		return fmt.Errorf("servings must be > 0")
	}
	if response.TimeMinutes <= 0 {
		return fmt.Errorf("time_minutes must be > 0")
	}
	if len(response.Steps) == 0 {
		return fmt.Errorf("steps array is empty")
	}
	validDifficulty := map[string]bool{"easy": true, "medium": true, "hard": true}
	if !validDifficulty[response.Difficulty] {
		return fmt.Errorf("difficulty must be easy/medium/hard")
	}
	if response.Nutrition.Calories <= 0 {
		return fmt.Errorf("calories must be > 0")
	}
	return nil
}

// ===========================
// ЭТАП 5: Database Save
// ===========================

// saveRecipeToDB сохраняет рецепт в нормализованные таблицы
func (s *adminService) saveRecipeToDB(req CreateRecipeAIRequest, aiResponse *AIRecipeResponse, authorID string) (*models.RecipeCatalog, error) {
	// Генерируем canonical name из title
	canonicalName := strings.ToLower(strings.ReplaceAll(req.Title, " ", "_"))
	
	// Проверка на дубликаты (using GORM field name, not SQL column)
	var existing models.RecipeCatalog
	if err := s.db.Where("\"canonicalName\" = ?", canonicalName).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("recipe with similar name already exists: %s", canonicalName)
	}

	// Создаем Source JSONB (required field)
	sourceJSON, _ := json.Marshal(map[string]interface{}{
		"type":      "ai",
		"generator": "groq-llama-3.3-70b",
		"authorId":  authorID,
		"timestamp": time.Now().Unix(),
	})

	// Создаем рецепт
	recipe := &models.RecipeCatalog{
		ID:            uuid.New(),
		CanonicalName: canonicalName,
		Title:         req.Title,
		Country:       "pl", // Default, можно добавить в запрос
		Category:      "main", // Default, можно добавить в запрос
		Difficulty:    aiResponse.Difficulty,
		TimeMinutes:   aiResponse.TimeMinutes,
		Servings:      aiResponse.Servings,
		Source:        datatypes.JSON(sourceJSON),
	}

	// Сохраняем summary как description
	descPl := aiResponse.Summary
	recipe.DescriptionPl = &descPl

	// Начинаем транзакцию
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Сохраняем рецепт
	if err := tx.Create(recipe).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create recipe: %w", err)
	}

	// 2. Сохраняем ингредиенты
	for i, ingInput := range req.Ingredients {
		catalogIng := models.CatalogIngredient{
			ID:           uuid.New(),
			RecipeID:     recipe.ID,
			IngredientID: ingInput.IngredientID,
			Quantity:     ingInput.Quantity,
			Unit:         ingInput.Unit,
			SortOrder:    i + 1,
		}
		if err := tx.Create(&catalogIng).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to add ingredient: %w", err)
		}
	}

	// 3. Сохраняем шаги (в JSONB пока, позже можно в отдельную таблицу)
	stepsJSON, _ := json.Marshal(aiResponse.Steps)
	recipe.StepsPl = stepsJSON

	// 4. Сохраняем nutrition (в NutritionProfile JSONB)
	nutritionJSON, _ := json.Marshal(map[string]interface{}{
		"calories":     aiResponse.Nutrition.Calories,
		"protein":      aiResponse.Nutrition.Protein,
		"fat":          aiResponse.Nutrition.Fat,
		"carbohydrate": aiResponse.Nutrition.Carbohydrate,
	})
	recipe.NutritionProfile = nutritionJSON

	// Обновляем рецепт с steps и nutrition
	if err := tx.Save(recipe).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update recipe: %w", err)
	}

	// Коммит транзакции
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("transaction commit failed: %w", err)
	}

	fmt.Printf("💾 Recipe saved: %s [%s] with %d ingredients, %d steps\n",
		recipe.Title, recipe.ID, len(req.Ingredients), len(aiResponse.Steps))

	return recipe, nil
}
