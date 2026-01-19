package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ===========================
// AI Recipe Creation DTOs
// ===========================

// CreateRecipeAIRequest - минимальный человеко-ориентированный запрос
type CreateRecipeAIRequest struct {
	Title          string                  `json:"title"`          // "Лосось с рисом и соусом терияки"
	Language       string                  `json:"language"`       // "pl", "en", "ru" (опционально, default "en")
	Ingredients    []RecipeIngredientInput `json:"ingredients"`    // [{ingredientId, quantity, unit}]
	RawCookingText string                  `json:"rawCookingText"` // "Рыбу замариновать в соусе, обжарить. Рис отварить."
}

// RecipeIngredientInput - ингредиент из запроса
type RecipeIngredientInput struct {
	IngredientID string  `json:"ingredientId"` // UUID ингредиента
	Quantity     float64 `json:"quantity"`     // 150 (приоритетное поле)
	Amount       float64 `json:"amount"`       // 150 (альтернативное поле для совместимости с frontend)
	Unit         string  `json:"unit"`         // "g"
}

// GetQuantity возвращает количество (quantity или amount, что заполнено)
func (r *RecipeIngredientInput) GetQuantity() float64 {
	if r.Quantity > 0 {
		return r.Quantity
	}
	return r.Amount
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
	Title               string                  `json:"title"`
	Language            string                  `json:"language"` // "pl", "en", "ru"
	Ingredients         []EnrichedIngredient    `json:"ingredients"`
	OriginalIngredients []RecipeIngredientInput `json:"-"` // Для валидации (не передаем в AI)
	RawCookingText      string                  `json:"cooking_instructions"`
}

// AIRecipeResponse - строгий контракт ответа от AI (ПОЛНЫЙ)
type AIRecipeResponse struct {
	Title       string               `json:"title"`        // Оригинальное название (НЕ МЕНЯТЬ)
	Language    string               `json:"language"`     // Язык рецепта
	Description string               `json:"description"`  // Краткое описание (summary)
	Servings    int                  `json:"servings"`     // Порций
	TimeMinutes int                  `json:"time_minutes"` // Общее время
	Difficulty  string               `json:"difficulty"`   // easy/medium/hard
	Calories    int                  `json:"calories"`     // Калории на порцию
	Ingredients []AIRecipeIngredient `json:"ingredients"`  // Ингредиенты с ID
	Steps       []RecipeStepAI       `json:"steps"`        // Шаги приготовления
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
	Calories     int     `json:"calories"`     // 520
	Protein      float64 `json:"protein"`      // 38
	Fat          float64 `json:"fat"`          // 22
	Carbohydrate float64 `json:"carbohydrate"` // 42
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

	// 🔥 BACKEND НОРМАЛИЗАЦИЯ - гарантия качества (даже в preview)
	aiResponse.Title = utils.CapitalizeTitle(aiResponse.Title)
	aiResponse.Description = utils.CleanRecipeText(aiResponse.Description)
	
	// Нормализуем шаги
	for i := range aiResponse.Steps {
		aiResponse.Steps[i].Text = utils.CleanRecipeText(aiResponse.Steps[i].Text)
	}

	fmt.Printf("🔍 Recipe preview generated: %s (%d steps)\n", aiResponse.Title, len(aiResponse.Steps))
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
		// 🔍 DEBUG: Логируем входные данные от frontend
		fmt.Printf("📥 RAW INPUT: ID=%s, Quantity=%.6f, Amount=%.6f, Unit=%s\n",
			input.IngredientID, input.Quantity, input.Amount, input.Unit)

		// Загружаем ингредиент из БД
		var ingredient models.Ingredient
		if err := s.db.Where("id = ?", input.IngredientID).First(&ingredient).Error; err != nil {
			return nil, fmt.Errorf("ingredient %s not found: %w", input.IngredientID, err)
		}

		// Выбираем локализованное название на основе языка
		name := s.getLocalizedName(ingredient, lang)

		finalQuantity := input.GetQuantity()
		enriched = append(enriched, EnrichedIngredient{
			IngredientID:   input.IngredientID, // Сохраняем ID!
			Name:           name,
			Quantity:       finalQuantity, // Используем quantity или amount
			Unit:           input.Unit,
			NutritionGroup: ingredient.NutritionGroup,
			Category:       ingredient.Category,
		})

		// 🔍 DEBUG: Логируем результат обогащения
		fmt.Printf("📤 ENRICHED: Name=%s, Quantity=%.2f %s\n", name, finalQuantity, input.Unit)
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
	// SYSTEM PROMPT: Строгие правила для сохранения данных + исправление качества текста
	systemPrompt := fmt.Sprintf(`You are a professional chef, food technologist, and food editor.

CRITICAL TEXT QUALITY RULES (MANDATORY):
1. Fix ALL spelling mistakes in recipe title, description, and steps
   - Example: "яишница" → "яичница", "egs" → "eggs", "tomatoe" → "tomato"
2. Recipe title MUST start with a capital letter
3. Each step MUST start with a capital letter
4. Description MUST start with a capital letter
5. Use correct culinary terminology in %s language
6. NEVER preserve typos or incorrect casing from user input
7. DO NOT use emojis in recipe text
8. Remove extra whitespace

USER INPUT MAY CONTAIN:
- Spelling mistakes (you MUST correct them)
- Lowercase letters (you MUST capitalize appropriately)
- Incorrect casing (you MUST fix it)

CRITICAL DATA PRESERVATION RULES:
1. DO NOT change ingredient quantities or units
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
  "title": "string - CORRECTED title with proper spelling and capitalization",
  "language": "string - %s",
  "description": "string - 1-2 sentence description in %s (capitalized, no typos)",
  "servings": number,
  "time_minutes": number,
  "difficulty": "easy|medium|hard",
  "calories": number,
  "steps": [
    {"order": number, "text": "string in %s (capitalized, no typos)", "time": number}
  ],
  "ingredients": [
    {
      "ingredientId": "uuid from input",
      "name": "string - ingredient name in %s (corrected spelling)",
      "amount": number - EXACT from input,
      "unit": "string - EXACT from input"
    }
  ]
}

Remember: You are a professional editor. Fix spelling and casing, preserve quantities.`,
		context.Language, context.Language, context.Language, context.Language, context.Language, context.Language, context.Language)

	// USER PROMPT: Передаем данные для структурирования
	ingredientsJSON, _ := json.Marshal(context.Ingredients)
	userPrompt := fmt.Sprintf(`⚠️ IMPORTANT: The user input below may contain spelling mistakes and lowercase letters. You MUST correct them.

Original title (may have typos):
"%s"

Language: %s

Ingredients (preserve IDs and amounts exactly):
%s

Raw Cooking Instructions (may have typos):
%s

YOUR TASK:
1. Fix spelling mistakes in title and instructions
2. Capitalize title, description, and each step
3. Structure the data according to JSON schema
4. Return ONLY JSON (no markdown, no explanations)`,
		context.Title, context.Language, string(ingredientsJSON), context.RawCookingText)

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
	// 1. Title может отличаться (AI исправляет орфографию и капитализацию)
	// Проверяем что это не полная замена, а только коррекция
	if response.Title != originalTitle {
		// Разрешаем изменения если они в пределах разумного (например, не более 30% difference)
		originalLower := strings.ToLower(strings.ReplaceAll(originalTitle, " ", ""))
		responseLower := strings.ToLower(strings.ReplaceAll(response.Title, " ", ""))
		
		// Если после нормализации сильно отличается - это подозрительно
		if originalLower != responseLower {
			// Вычисляем простую метрику похожести
			if len(responseLower) < len(originalLower)/2 || len(responseLower) > len(originalLower)*2 {
				return fmt.Errorf("AI drastically changed the title: expected '%s', got '%s'", originalTitle, response.Title)
			}
		}
		// Если отличается только капитализацией/пробелами/орфографией - OK
		fmt.Printf("✏️ AI corrected title: '%s' → '%s'\n", originalTitle, response.Title)
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
	// 🔥 BACKEND НОРМАЛИЗАЦИЯ - гарантия качества (даже если AI ошибся)
	// Этот код - источник истины, не доверяем AI на 100%
	normalizedTitle := utils.CapitalizeTitle(aiResponse.Title)
	normalizedDescription := utils.CleanRecipeText(aiResponse.Description)
	
	// Нормализуем шаги (капитализация, очистка)
	for i := range aiResponse.Steps {
		aiResponse.Steps[i].Text = utils.CleanRecipeText(aiResponse.Steps[i].Text)
	}
	
	// Генерируем canonical name из normalized title (English slug)
	// КРИТИЧНО: canonical name БЕЗ опечаток и кириллицы
	canonicalName := utils.GenerateCanonicalName(normalizedTitle)

	// Проверка на дубликаты (using GORM field name, not SQL column)
	var existing models.RecipeCatalog
	if err := s.db.Where("\"canonicalName\" = ?", canonicalName).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("recipe with similar name already exists: %s", canonicalName)
	}

	// Создаем Source JSONB (required field)
	// Философия ChefOS: Админ утверждает → рецепт становится professional
	sourceJSON, _ := json.Marshal(map[string]interface{}{
		"type":      "professional", // Админ взял ответственность
		"generator": "groq-llama-3.3-70b",
		"authorId":  authorID,
		"timestamp": time.Now().Unix(),
	})

	// Создаем рецепт
	recipe := &models.RecipeCatalog{
		ID:            uuid.New(),
		CanonicalName: canonicalName,
		Title:         normalizedTitle,  // 🔥 Нормализованный title (Яичница, не яишница)
		Country:       "pl",              // Default, можно добавить в запрос
		Category:      "main",            // Default, можно добавить в запрос
		Difficulty:    aiResponse.Difficulty,
		TimeMinutes:   aiResponse.TimeMinutes,
		Servings:      aiResponse.Servings,
		Source:        datatypes.JSON(sourceJSON),
	}

	// Сохраняем description (нормализованный, в зависимости от языка)
	switch aiResponse.Language {
	case "pl":
		recipe.DescriptionPl = &normalizedDescription
	case "ru":
		recipe.DescriptionRu = &normalizedDescription
	default: // "en"
		recipe.DescriptionEn = &normalizedDescription
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
			Quantity:     ingInput.GetQuantity(), // Используем quantity или amount
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

// ===========================
// ЭТАП 5: Save Edited Recipe
// ===========================

// SaveEditedRecipeRequest - структура для сохранения отредактированного рецепта
type SaveEditedRecipeRequest struct {
	RecipeID    *string            `json:"recipeId,omitempty"` // UUID рецепта (если редактирование существующего)
	Title       string             `json:"title"`              // Отредактированное название
	Language    string             `json:"language"`           // Язык рецепта
	Description string             `json:"description"`        // Отредактированное описание
	Servings    int                `json:"servings"`           // Количество порций
	TimeMinutes int                `json:"time_minutes"`       // Время приготовления
	Difficulty  string             `json:"difficulty"`         // easy, medium, hard
	Calories    int                `json:"calories"`           // Калории
	Ingredients []EditedIngredient `json:"ingredients"`        // Отредактированные ингредиенты
	Steps       []EditedStep       `json:"steps"`              // Отредактированные шаги
}

// EditedIngredient - отредактированный ингредиент
type EditedIngredient struct {
	IngredientID string  `json:"ingredientId"` // UUID ингредиента из каталога
	Name         string  `json:"name"`         // Локализованное название (для отображения)
	Amount       float64 `json:"amount"`       // Количество
	Unit         string  `json:"unit"`         // Единица измерения
}

// EditedStep - отредактированный шаг приготовления
type EditedStep struct {
	Order int    `json:"order"` // Порядковый номер
	Text  string `json:"text"`  // Текст инструкции
	Time  int    `json:"time"`  // Время выполнения в минутах
}

// SaveEditedRecipe сохраняет отредактированный пользователем рецепт в БД
func (s *adminService) SaveEditedRecipe(req SaveEditedRecipeRequest, userID string) (*models.RecipeCatalog, error) {
	// Валидация
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if len(req.Ingredients) == 0 {
		return nil, fmt.Errorf("ingredients are required")
	}
	if len(req.Steps) == 0 {
		return nil, fmt.Errorf("steps are required")
	}

	fmt.Printf("💾 Saving edited recipe: '%s' (lang=%s, %d ingredients, %d steps)\n",
		req.Title, req.Language, len(req.Ingredients), len(req.Steps))

	// Генерируем canonical name (English slug)
	canonicalName := utils.GenerateCanonicalName(req.Title)

	// Проверка на дубликаты (исключая текущий рецепт при редактировании)
	var existing models.RecipeCatalog
	query := s.db.Where("\"canonicalName\" = ?", canonicalName)

	// Если это редактирование (есть RecipeID), исключаем текущий рецепт из проверки
	if req.RecipeID != nil && *req.RecipeID != "" {
		query = query.Where("id != ?", *req.RecipeID)
	}

	if err := query.First(&existing).Error; err == nil {
		return nil, fmt.Errorf("recipe with similar name already exists: %s", canonicalName)
	}

	// Начинаем транзакцию
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			fmt.Printf("🚨 Transaction rolled back due to panic: %v\n", r)
		}
	}()

	// Создаём Source JSONB
	// Философия ChefOS: Факт сохранения админом = professional
	sourceJSON, _ := json.Marshal(map[string]interface{}{
		"type":      "professional", // Админ взял ответственность
		"generator": "groq-llama-3.3-70b",
		"authorId":  userID,
		"timestamp": time.Now().Unix(),
	})

	// Определяем страну по языку
	country := "pl" // default
	if req.Language == "ru" {
		country = "ru"
	} else if req.Language == "en" {
		country = "us"
	}

	// 1. Определяем режим: создание нового или обновление существующего
	var recipe *models.RecipeCatalog
	isEditMode := req.RecipeID != nil && *req.RecipeID != ""

	if isEditMode {
		// Режим редактирования - загружаем существующий рецепт
		recipeID, err := uuid.Parse(*req.RecipeID)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("invalid recipe ID: %w", err)
		}

		recipe = &models.RecipeCatalog{}
		if err := tx.First(recipe, "id = ?", recipeID).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("recipe not found: %w", err)
		}

		fmt.Printf("📝 Editing existing recipe: ID=%s\n", recipe.ID)

		// Обновляем поля
		recipe.CanonicalName = canonicalName
		recipe.Title = req.Title
		recipe.Difficulty = req.Difficulty
		recipe.TimeMinutes = req.TimeMinutes
		recipe.Servings = req.Servings
		recipe.Country = country
		recipe.Source = datatypes.JSON(sourceJSON)

		// Удаляем старые ингредиенты (будем создавать заново)
		if err := tx.Where("\"recipeId\" = ?", recipe.ID).Delete(&models.CatalogIngredient{}).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to delete old ingredients: %w", err)
		}
	} else {
		// Режим создания - создаем новый рецепт
		recipe = &models.RecipeCatalog{
			ID:            uuid.New(),
			CanonicalName: canonicalName,
			Title:         req.Title,
			Category:      "main", // default
			Difficulty:    req.Difficulty,
			TimeMinutes:   req.TimeMinutes,
			Servings:      req.Servings,
			Country:       country,
			Source:        datatypes.JSON(sourceJSON),
		}

		fmt.Printf("✨ Creating new recipe: ID=%s\n", recipe.ID)

		// Создаем новый рецепт
		if err := tx.Create(recipe).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create recipe: %w", err)
		}
	}

	// Устанавливаем локализованное описание
	if req.Language == "ru" {
		recipe.DescriptionRu = &req.Description
	} else if req.Language == "pl" {
		recipe.DescriptionPl = &req.Description
	} else {
		recipe.DescriptionEn = &req.Description
	}

	// 2. Создаём CatalogIngredients
	for _, ing := range req.Ingredients {
		// Генерируем ingredientKey для поиска
		ingredientKey := strings.ToLower(strings.ReplaceAll(ing.Name, " ", "_"))

		recipeIng := models.CatalogIngredient{
			ID:            uuid.New(),
			RecipeID:      recipe.ID,
			IngredientID:  ing.IngredientID,
			IngredientKey: ingredientKey,
			Quantity:      ing.Amount,
			Unit:          ing.Unit,
			Optional:      false,
			SortOrder:     0,
		}

		if err := tx.Create(&recipeIng).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create recipe ingredient: %w", err)
		}
	}

	fmt.Printf("✅ Created %d recipe ingredients\n", len(req.Ingredients))

	// 3. Создаём локализованные шаги
	stepsData := make([]map[string]interface{}, 0, len(req.Steps))
	for _, step := range req.Steps {
		stepsData = append(stepsData, map[string]interface{}{
			"order": step.Order,
			"text":  step.Text,
			"time":  step.Time,
		})
	}

	stepsJSON, err := json.Marshal(stepsData)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to marshal steps: %w", err)
	}

	// Устанавливаем локализованные шаги
	if req.Language == "ru" {
		recipe.StepsRu = datatypes.JSON(stepsJSON)
	} else if req.Language == "pl" {
		recipe.StepsPl = datatypes.JSON(stepsJSON)
	} else {
		recipe.StepsEn = datatypes.JSON(stepsJSON)
	}

	fmt.Printf("✅ Created %d recipe steps\n", len(req.Steps))

	// 4. Устанавливаем nutrition profile
	nutritionJSON, _ := json.Marshal(map[string]int{
		"calories":     req.Calories,
		"protein":      0,
		"fat":          0,
		"carbohydrate": 0,
	})
	recipe.NutritionProfile = datatypes.JSON(nutritionJSON)

	// Сохраняем или обновляем рецепт
	if isEditMode {
		// Для редактирования используем Updates с явным указанием полей
		updates := map[string]interface{}{
			"canonicalName":    recipe.CanonicalName,
			"title":            recipe.Title,
			"difficulty":       recipe.Difficulty,
			"timeMinutes":      recipe.TimeMinutes,
			"servings":         recipe.Servings,
			"country":          recipe.Country,
			"source":           recipe.Source,
			"nutritionProfile": recipe.NutritionProfile,
		}

		// Добавляем локализованные поля
		if req.Language == "ru" {
			updates["description_ru"] = recipe.DescriptionRu
			updates["steps_ru"] = recipe.StepsRu
		} else if req.Language == "pl" {
			updates["description_pl"] = recipe.DescriptionPl
			updates["steps_pl"] = recipe.StepsPl
		} else {
			updates["description_en"] = recipe.DescriptionEn
			updates["steps_en"] = recipe.StepsEn
		}

		if err := tx.Model(recipe).Updates(updates).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update recipe: %w", err)
		}

		fmt.Printf("✅ Recipe updated: ID=%s\n", recipe.ID)
	} else {
		// Для нового рецепта используем Save
		if err := tx.Save(recipe).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to save recipe: %w", err)
		}

		fmt.Printf("✅ Recipe saved: ID=%s\n", recipe.ID)
	}

	// Коммит транзакции
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("transaction commit failed: %w", err)
	}

	fmt.Printf("💾 Edited recipe saved: %s [%s]\n", recipe.Title, recipe.ID)

	return recipe, nil
}

// ===========================
// ЭТАП 6: Update Existing Recipe
// ===========================

// UpdateRecipeRequest - структура для обновления существующего рецепта
type UpdateRecipeRequest struct {
	Title       string             `json:"title"`
	Language    string             `json:"language"`
	Description string             `json:"description"`
	Servings    int                `json:"servings"`
	TimeMinutes int                `json:"time_minutes"`
	Difficulty  string             `json:"difficulty"`
	Calories    int                `json:"calories"`
	Ingredients []EditedIngredient `json:"ingredients"`
	Steps       []EditedStep       `json:"steps"`
}

// UpdateRecipe обновляет существующий рецепт
func (s *adminService) UpdateRecipe(recipeID string, req UpdateRecipeRequest) (*models.RecipeCatalog, error) {
	// Валидация
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if len(req.Ingredients) == 0 {
		return nil, fmt.Errorf("ingredients are required")
	}
	if len(req.Steps) == 0 {
		return nil, fmt.Errorf("steps are required")
	}

	fmt.Printf("🔄 Updating recipe: %s\n", recipeID)

	// Начинаем транзакцию
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Загружаем существующий рецепт
	var recipe models.RecipeCatalog
	if err := tx.Where("id = ?", recipeID).First(&recipe).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("recipe not found: %w", err)
	}

	// 2. Обновляем основные поля
	recipe.Title = req.Title
	recipe.CanonicalName = utils.GenerateCanonicalName(req.Title)
	recipe.Difficulty = req.Difficulty
	recipe.TimeMinutes = req.TimeMinutes
	recipe.Servings = req.Servings

	// Обновляем описание
	if req.Language == "ru" {
		recipe.DescriptionRu = &req.Description
	} else if req.Language == "pl" {
		recipe.DescriptionPl = &req.Description
	} else {
		recipe.DescriptionEn = &req.Description
	}

	// 3. Удаляем старые ингредиенты
	if err := tx.Where("\"recipeId\" = ?", recipeID).Delete(&models.CatalogIngredient{}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to delete old ingredients: %w", err)
	}

	// 4. Создаём новые ингредиенты
	for _, ing := range req.Ingredients {
		ingredientKey := strings.ToLower(strings.ReplaceAll(ing.Name, " ", "_"))

		recipeIng := models.CatalogIngredient{
			ID:            uuid.New(),
			RecipeID:      recipe.ID,
			IngredientID:  ing.IngredientID,
			IngredientKey: ingredientKey,
			Quantity:      ing.Amount,
			Unit:          ing.Unit,
			Optional:      false,
			SortOrder:     0,
		}

		if err := tx.Create(&recipeIng).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create recipe ingredient: %w", err)
		}
	}

	// 5. Обновляем шаги
	stepsData := make([]map[string]interface{}, 0, len(req.Steps))
	for _, step := range req.Steps {
		stepsData = append(stepsData, map[string]interface{}{
			"order": step.Order,
			"text":  step.Text,
			"time":  step.Time,
		})
	}

	stepsJSON, err := json.Marshal(stepsData)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to marshal steps: %w", err)
	}

	if req.Language == "ru" {
		recipe.StepsRu = datatypes.JSON(stepsJSON)
	} else if req.Language == "pl" {
		recipe.StepsPl = datatypes.JSON(stepsJSON)
	} else {
		recipe.StepsEn = datatypes.JSON(stepsJSON)
	}

	// 6. Обновляем nutrition
	nutritionJSON, _ := json.Marshal(map[string]int{
		"calories":     req.Calories,
		"protein":      0,
		"fat":          0,
		"carbohydrate": 0,
	})
	recipe.NutritionProfile = datatypes.JSON(nutritionJSON)

	// Сохраняем обновления
	if err := tx.Save(&recipe).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update recipe: %w", err)
	}

	// Коммит
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("transaction commit failed: %w", err)
	}

	fmt.Printf("✅ Recipe updated: %s [%s]\n", recipe.Title, recipe.ID)

	return &recipe, nil
}

// DeleteRecipe - удалить рецепт из каталога
func (s *adminService) DeleteRecipe(recipeID string) error {
	fmt.Printf("🗑️  Deleting recipe: %s\n", recipeID)

	// Начинаем транзакцию
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Проверяем существование рецепта
	var recipe models.RecipeCatalog
	if err := tx.Where("id = ?", recipeID).First(&recipe).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("recipe not found")
		}
		return fmt.Errorf("failed to find recipe: %w", err)
	}

	recipeName := recipe.Title

	// 2. Удаляем связанные ингредиенты (CASCADE через GORM)
	if err := tx.Where("\"recipeId\" = ?", recipeID).Delete(&models.CatalogIngredient{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete recipe ingredients: %w", err)
	}

	// 3. Удаляем связи с аллергенами (many2many)
	if err := tx.Exec("DELETE FROM \"RecipeAllergen\" WHERE \"recipeId\" = ?", recipeID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete recipe allergens: %w", err)
	}

	// 4. Удаляем связи с диет-тегами (many2many)
	if err := tx.Exec("DELETE FROM \"RecipeDietTag\" WHERE \"recipeId\" = ?", recipeID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete recipe diet tags: %w", err)
	}

	// 5. Удаляем сам рецепт
	if err := tx.Delete(&recipe).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete recipe: %w", err)
	}

	// Коммит
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	fmt.Printf("✅ Recipe deleted: %s [%s]\n", recipeName, recipeID)
	return nil
}

// ===========================
// ЭТАП 7: Smart Conflict Resolution
// ===========================

// GenerateAlternativeTitles генерирует альтернативные названия рецепта через AI
func (s *adminService) GenerateAlternativeTitles(originalTitle, language string) ([]string, error) {
	// Определяем язык для промпта
	langName := "Russian"
	if language == "pl" {
		langName = "Polish"
	} else if language == "en" {
		langName = "English"
	}

	systemPrompt := `You are a culinary naming assistant. Your task is to suggest alternative recipe titles that are:
- Natural and appetizing
- SEO-friendly
- Clear and descriptive
- Suitable for a cooking platform

Return ONLY a valid JSON array of 5 strings. No markdown, no explanations.`

	userPrompt := fmt.Sprintf(`Language: %s

The recipe title "%s" already exists in the database.

Generate 5 alternative titles that are unique but similar in meaning.

Examples of good alternatives:
- Add cooking method: "Pan-Fried Salmon"
- Add style: "Homestyle Fried Salmon"
- Add detail: "Fried Salmon with Crispy Skin"
- Add variation: "Garlic Butter Fried Salmon"
- Add regional touch: "Russian-Style Fried Salmon"

Return JSON array:
["Title 1", "Title 2", "Title 3", "Title 4", "Title 5"]`, langName, originalTitle)

	fmt.Printf("🤖 Generating alternative titles for '%s' (lang=%s)...\n", originalTitle, language)

	// Вызываем Groq AI
	response, err := s.groqClient.SimpleChat(systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI call failed: %w", err)
	}

	fmt.Printf("📥 AI Response: %s\n", response)

	// Очищаем ответ от markdown
	cleaned := strings.TrimSpace(response)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	// Парсим JSON array
	var suggestions []string
	if err := json.Unmarshal([]byte(cleaned), &suggestions); err != nil {
		fmt.Printf("❌ Failed to parse AI response as JSON array: %v\n", err)
		// Fallback: базовые предложения
		return []string{
			originalTitle + " (домашний рецепт)",
			originalTitle + " (авторский)",
			originalTitle + " на сковороде",
			originalTitle + " с пряностями",
			originalTitle + " (классический)",
		}, nil
	}

	fmt.Printf("✅ Generated %d alternative titles\n", len(suggestions))
	return suggestions, nil
}

// GenerateMultilingualTitles генерирует альтернативные названия на всех языках (RU/EN/PL)
func (s *adminService) GenerateMultilingualTitles(originalTitle, primaryLanguage string) (map[string][]string, error) {
	fmt.Printf("🌍 Generating multilingual alternative titles for '%s' (primary=%s)...\n", originalTitle, primaryLanguage)

	// Определяем все языки для генерации
	languages := []string{"ru", "en", "pl"}

	// Результат: map языка на список предложений
	result := make(map[string][]string)

	// Оптимизация: генерируем все языки в одном AI запросе
	systemPrompt := `You are a multilingual culinary naming assistant.
Generate alternative recipe titles in 3 languages: Russian, English, and Polish.
Return ONLY valid JSON. No markdown, no explanations.`

	// Определяем примеры для каждого языка
	examples := map[string]string{
		"ru": `- Жареный Лосось с Хрустящей Кожей
- Домашний Жареный Лосось
- Лосось на Сковороде`,
		"en": `- Pan-Fried Salmon with Crispy Skin
- Homestyle Fried Salmon
- Skillet Salmon`,
		"pl": `- Smażony Łosoś z Chrupiącą Skórką
- Domowy Smażony Łosoś
- Łosoś na Patelni`,
	}

	userPrompt := fmt.Sprintf(`The recipe title "%s" already exists.

Generate 5 alternative titles in EACH language (Russian, English, Polish).
Make titles:
- Natural and appetizing
- SEO-friendly
- Unique variations (add cooking method, style, or ingredients)

Examples:
Russian: %s
English: %s
Polish: %s

Return JSON object:
{
  "ru": ["Title 1", "Title 2", "Title 3", "Title 4", "Title 5"],
  "en": ["Title 1", "Title 2", "Title 3", "Title 4", "Title 5"],
  "pl": ["Title 1", "Title 2", "Title 3", "Title 4", "Title 5"]
}`, originalTitle, examples["ru"], examples["en"], examples["pl"])

	fmt.Printf("🤖 Calling AI for multilingual suggestions...\n")

	// Вызываем Groq AI
	response, err := s.groqClient.SimpleChat(systemPrompt, userPrompt)
	if err != nil {
		fmt.Printf("❌ AI call failed: %v\n", err)
		// Fallback: генерируем простые варианты для каждого языка
		return s.generateFallbackTitles(originalTitle), nil
	}

	fmt.Printf("📥 AI Response length: %d chars\n", len(response))

	// Очищаем ответ от markdown
	cleaned := strings.TrimSpace(response)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	// Парсим JSON object с языками
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		fmt.Printf("❌ Failed to parse multilingual response: %v\n", err)
		// Fallback
		return s.generateFallbackTitles(originalTitle), nil
	}

	// Валидация: проверяем, что все языки присутствуют
	for _, lang := range languages {
		if titles, ok := result[lang]; !ok || len(titles) == 0 {
			fmt.Printf("⚠️  Missing suggestions for %s, adding fallback\n", lang)
			result[lang] = s.generateFallbackForLanguage(originalTitle, lang)
		}
	}

	fmt.Printf("✅ Generated multilingual titles: RU=%d, EN=%d, PL=%d\n",
		len(result["ru"]), len(result["en"]), len(result["pl"]))

	return result, nil
}

// generateFallbackTitles создаёт простые варианты на всех языках
func (s *adminService) generateFallbackTitles(originalTitle string) map[string][]string {
	return map[string][]string{
		"ru": {
			originalTitle + " (домашний рецепт)",
			originalTitle + " (авторский)",
			originalTitle + " на сковороде",
			originalTitle + " с пряностями",
			originalTitle + " (классический)",
		},
		"en": {
			originalTitle + " (Homestyle)",
			originalTitle + " (Chef's Version)",
			originalTitle + " (Pan-Fried)",
			originalTitle + " (Spiced)",
			originalTitle + " (Classic)",
		},
		"pl": {
			originalTitle + " (Domowy Przepis)",
			originalTitle + " (Autorski)",
			originalTitle + " (Na Patelni)",
			originalTitle + " (Z Przyprawami)",
			originalTitle + " (Klasyczny)",
		},
	}
}

// generateFallbackForLanguage создаёт простые варианты для одного языка
func (s *adminService) generateFallbackForLanguage(originalTitle, lang string) []string {
	suffixes := map[string][]string{
		"ru": {" (домашний)", " (авторский)", " (улучшенный)", " на сковороде", " классический"},
		"en": {" (Homestyle)", " (Chef's)", " (Improved)", " (Pan-Fried)", " (Classic)"},
		"pl": {" (Domowy)", " (Autorski)", " (Ulepszony)", " (Na Patelni)", " (Klasyczny)"},
	}

	suggestions := make([]string, 0, 5)
	for _, suffix := range suffixes[lang] {
		suggestions = append(suggestions, originalTitle+suffix)
	}
	return suggestions
}
