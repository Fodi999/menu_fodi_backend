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
	Language       string                    `json:"language"`       // "pl", "en", "ru" (опционально, default "en")
	Ingredients    []RecipeIngredientInput   `json:"ingredients"`    // [{ingredientId, quantity, unit}]
	RawCookingText string                    `json:"rawCookingText"` // "Рыбу замариновать в соусе, обжарить. Рис отварить."
}

// RecipeIngredientInput - ингредиент из запроса
type RecipeIngredientInput struct {
	IngredientID string  `json:"ingredientId"` // UUID ингредиента
	Quantity     float64 `json:"quantity"`     // 150
	Unit         string  `json:"unit"`         // "g"
}

// EnrichedIngredient - обогащенный ингредиент для AI (с сохранением ID)
type EnrichedIngredient struct {
	IngredientID   string  `json:"ingredientId"`    // UUID из запроса (для возврата)
	Name           string  `json:"ingredient"`      // "salmon" или "Łosoś" (локализовано)
	Quantity       float64 `json:"quantity"`        // 150
	Unit           string  `json:"unit"`            // "g"
	NutritionGroup string  `json:"nutrition_group"` // "protein"
	Category       string  `json:"category"`        // "fish"
}

// AIRecipePromptContext - контекст для AI
type AIRecipePromptContext struct {
	Title              string                  `json:"title"`
	Language           string                  `json:"language"`        // "pl", "en", "ru"
	Ingredients        []EnrichedIngredient    `json:"ingredients"`
	OriginalIngredients []RecipeIngredientInput `json:"-"` // Для валидации (не передаем в AI)
	RawCookingText     string                  `json:"cooking_instructions"`
}

// AIRecipeResponse - строгий контракт ответа от AI (ПОЛНЫЙ)
type AIRecipeResponse struct {
	Title       string                   `json:"title"`        // Оригинальное название (НЕ МЕНЯТЬ)
	Language    string                   `json:"language"`     // Язык рецепта
	Description string                   `json:"description"`  // Краткое описание (summary)
	Servings    int                      `json:"servings"`     // Порций
	TimeMinutes int                      `json:"time_minutes"` // Общее время
	Difficulty  string                   `json:"difficulty"`   // easy/medium/hard
	Calories    int                      `json:"calories"`     // Калории на порцию
	Ingredients []AIRecipeIngredient     `json:"ingredients"`  // Ингредиенты с ID
	Steps       []RecipeStepAI           `json:"steps"`        // Шаги приготовления
}

// AIRecipeIngredient - ингредиент в ответе AI (с сохранением ID)
type AIRecipeIngredient struct {
	IngredientID string  `json:"ingredientId"` // UUID из запроса (НЕ МЕНЯТЬ)
	Name         string  `json:"name"`         // Локализованное название
	Amount       float64 `json:"amount"`       // Количество (НЕ МЕНЯТЬ)
	Unit         string  `json:"unit"`         // Единица измерения (НЕ МЕНЯТЬ)
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
	// Используем язык из запроса, по умолчанию "en"
	lang := req.Language
	if lang == "" {
		lang = "en"
	}

	// ЭТАП 2: Обогащаем данные об ингредиентах из БД
	enrichedIngredients, err := s.enrichIngredientsForAI(req.Ingredients, lang)
	if err != nil {
		return nil, fmt.Errorf("failed to enrich ingredients: %w", err)
	}

	// ЭТАП 3: Формируем контекст для AI
	promptContext := AIRecipePromptContext{
		Title:               req.Title,
		Language:            lang,
		Ingredients:         enrichedIngredients,
		OriginalIngredients: req.Ingredients,
		RawCookingText:      req.RawCookingText,
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
	// Используем язык из запроса, по умолчанию "en"
	lang := req.Language
	if lang == "" {
		lang = "en"
	}

	// ЭТАП 2: Обогащаем данные об ингредиентах
	enrichedIngredients, err := s.enrichIngredientsForAI(req.Ingredients, lang)
	if err != nil {
		return nil, fmt.Errorf("failed to enrich ingredients: %w", err)
	}

	// ЭТАП 3: Формируем контекст и вызываем AI
	promptContext := AIRecipePromptContext{
		Title:               req.Title,
		Language:            lang,
		Ingredients:         enrichedIngredients,
		OriginalIngredients: req.Ingredients,
		RawCookingText:      req.RawCookingText,
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
// Теперь с поддержкой локализации и сохранением ID
func (s *adminService) enrichIngredientsForAI(inputs []RecipeIngredientInput, lang string) ([]EnrichedIngredient, error) {
	enriched := make([]EnrichedIngredient, 0, len(inputs))

	for _, input := range inputs {
		// Загружаем ингредиент из БД
		var ingredient models.Ingredient
		if err := s.db.Where("id = ?", input.IngredientID).First(&ingredient).Error; err != nil {
			return nil, fmt.Errorf("ingredient %s not found: %w", input.IngredientID, err)
		}

		// Выбираем локализованное название на основе языка
		name := s.getLocalizedName(ingredient, lang)

		enriched = append(enriched, EnrichedIngredient{
			IngredientID:   input.IngredientID, // Сохраняем ID!
			Name:           name,
			Quantity:       input.Quantity,
			Unit:           input.Unit,
			NutritionGroup: ingredient.NutritionGroup,
			Category:       ingredient.Category,
		})
	}

	fmt.Printf("🔧 Enriched %d ingredients for AI (lang=%s)\n", len(enriched), lang)
	return enriched, nil
}

// ===========================
// ЭТАП 3: AI Generation
// ===========================
// ===========================

// generateRecipeViaAI вызывает Groq AI для структурирования рецепта
func (s *adminService) generateRecipeViaAI(context AIRecipePromptContext) (*AIRecipeResponse, error) {
	// SYSTEM PROMPT: Строгие правила для сохранения данных пользователя
	systemPrompt := fmt.Sprintf(`You are a professional chef and food technologist.

CRITICAL RULES - MUST FOLLOW STRICTLY:
1. DO NOT change the recipe title provided by the user
2. DO NOT invent, add, or remove any ingredients
3. Use ONLY the ingredients provided with their EXACT amounts and units
4. Return the recipe in the language specified: %s
5. Output ONLY valid JSON (no markdown, no explanations, no code blocks)

YOUR TASK:
- Create a 1-2 sentence description explaining what makes this dish special
- Break down the raw cooking text into clear, actionable steps with time estimates
- Calculate total cooking time based on step durations
- Determine difficulty: easy (≤30min), medium (30-60min), hard (>60min)
- Estimate calories per serving based on ingredient nutrition groups
- Determine servings (analyze ingredient quantities, default to 1 if unclear)

IMPORTANT:
- Each ingredient has a nutrition_group (protein/carbohydrate/vegetable/fat/other)
- Use this to estimate calories:
  * Protein (fish/meat): ~150-200 kcal/100g
  * Carbohydrate (grains/pasta): ~130-150 kcal/100g
  * Vegetables: ~25-50 kcal/100g
  * Fats/oils: ~800-900 kcal/100ml
- Steps must have realistic time estimates (in minutes)
- Description must be in %s language

STRICT JSON SCHEMA (return ONLY this structure):
{
  "title": "string - EXACT title from user input",
  "language": "string - %s",
  "description": "string - 1-2 sentence description in %s",
  "servings": number,
  "time_minutes": number,
  "difficulty": "easy|medium|hard",
  "calories": number,
  "steps": [
    {"order": number, "text": "string in %s", "time": number}
  ],
  "ingredients": [
    {
      "ingredientId": "uuid from input",
      "name": "string - ingredient name in %s",
      "amount": number - EXACT from input,
      "unit": "string - EXACT from input"
    }
  ]
}

Remember: You are structuring existing data, NOT creating a new recipe. Preserve all user input.`,
		context.Language, context.Language, context.Language, context.Language, context.Language, context.Language)

	// USER PROMPT: Передаем данные для структурирования
	ingredientsJSON, _ := json.Marshal(context.Ingredients)
	userPrompt := fmt.Sprintf(`Structure this recipe data:

Title: %s
Language: %s

Ingredients (preserve IDs and amounts exactly):
%s

Raw Cooking Instructions:
%s

Return ONLY JSON. No markdown, no explanations.`, context.Title, context.Language, string(ingredientsJSON), context.RawCookingText)

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
	if err := validateAIResponse(&aiResponse, context.Title, context.OriginalIngredients); err != nil {
		return nil, fmt.Errorf("AI response validation failed: %w", err)
	}

	fmt.Printf("✅ AI generated recipe: %d steps, %d min, %s difficulty\n",
		len(aiResponse.Steps), aiResponse.TimeMinutes, aiResponse.Difficulty)

	return &aiResponse, nil
}

// validateAIResponse проверяет корректность ответа AI
func validateAIResponse(response *AIRecipeResponse, originalTitle string, originalIngredients []RecipeIngredientInput) error {
	// 1. Title должен совпадать с оригиналом
	if response.Title != originalTitle {
		return fmt.Errorf("AI changed the title: expected '%s', got '%s'", originalTitle, response.Title)
	}

	// 2. Description не пустой
	if response.Description == "" {
		return fmt.Errorf("description is empty")
	}

	// 3. Базовые проверки
	if response.Servings <= 0 {
		return fmt.Errorf("servings must be > 0")
	}
	if response.TimeMinutes <= 0 {
		return fmt.Errorf("time_minutes must be > 0")
	}
	if len(response.Steps) == 0 {
		return fmt.Errorf("steps array is empty")
	}

	// 4. Difficulty валидация
	validDifficulty := map[string]bool{"easy": true, "medium": true, "hard": true}
	if !validDifficulty[response.Difficulty] {
		return fmt.Errorf("difficulty must be easy/medium/hard")
	}

	// 5. Calories проверка
	if response.Calories <= 0 {
		return fmt.Errorf("calories must be > 0")
	}

	// 6. Ingredients должны быть все с ID и совпадать с оригинальными
	if len(response.Ingredients) != len(originalIngredients) {
		return fmt.Errorf("ingredient count mismatch: expected %d, got %d", len(originalIngredients), len(response.Ingredients))
	}

	// Создаем map для проверки наличия всех ID
	originalIDs := make(map[string]bool)
	for _, ing := range originalIngredients {
		originalIDs[ing.IngredientID] = true
	}

	for i, ing := range response.Ingredients {
		if ing.IngredientID == "" {
			return fmt.Errorf("ingredient #%d has empty ingredientId", i+1)
		}
		if !originalIDs[ing.IngredientID] {
			return fmt.Errorf("ingredient #%d has unknown ID: %s", i+1, ing.IngredientID)
		}
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
		Title:         aiResponse.Title, // Используем title из AI (должен совпадать с req.Title)
		Country:       "pl", // Default, можно добавить в запрос
		Category:      "main", // Default, можно добавить в запрос
		Difficulty:    aiResponse.Difficulty,
		TimeMinutes:   aiResponse.TimeMinutes,
		Servings:      aiResponse.Servings,
		Source:        datatypes.JSON(sourceJSON),
	}

	// Сохраняем description (в зависимости от языка)
	switch aiResponse.Language {
	case "pl":
		recipe.DescriptionPl = &aiResponse.Description
	case "ru":
		recipe.DescriptionRu = &aiResponse.Description
	default: // "en"
		recipe.DescriptionEn = &aiResponse.Description
	}

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
	switch aiResponse.Language {
	case "pl":
		recipe.StepsPl = stepsJSON
	case "ru":
		recipe.StepsRu = stepsJSON
	default: // "en"
		recipe.StepsEn = stepsJSON
	}

	// 4. Сохраняем nutrition (в NutritionProfile JSONB)
	// TODO: В будущем AI будет возвращать полный nutrition (protein, fat, carbs)
	nutritionJSON, _ := json.Marshal(map[string]interface{}{
		"calories":     aiResponse.Calories,
		"protein":      0, // Пока AI возвращает только калории
		"fat":          0,
		"carbohydrate": 0,
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
