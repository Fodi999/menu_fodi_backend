package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai_recipe_recommendation/repository"
	"gorm.io/gorm"
)

// ============================================================================
// RECOMMENDATION SERVICE - основная бизнес-логика (один источник правды)
// ============================================================================

type RecommendationService struct {
	db               *gorm.DB
	recipeRepository *repository.RecipeRepository
}

func NewRecommendationService(db *gorm.DB) *RecommendationService {
	return &RecommendationService{
		db:               db,
		recipeRepository: repository.NewRecipeRepository(db),
	}
}

// GetRecommendations - главный метод: анализирует холодильник и подбирает рецепты
func (s *RecommendationService) GetRecommendations(
	ctx context.Context,
	req RecipeMatchRequest,
) (*RecipeRecommendationResponse, error) {
	// 1️⃣ Валидация и defaults
	if req.Limit == 0 {
		req.Limit = 10
	}
	if req.Language == "" {
		req.Language = "pl"
	}

	// 2️⃣ Получить ingredient_id из холодильника пользователя
	fridgeIngredientIDs, err := s.getUserFridgeIngredientIDs(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get fridge: %w", err)
	}

	if len(fridgeIngredientIDs) == 0 {
		return s.emptyFridgeResponse(req.Language), nil
	}

	// 3️⃣ Получить ВСЕ рецепты из каталога (через Repository)
	recipes, err := s.recipeRepository.GetAllRecipes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipes: %w", err)
	}

	if len(recipes) == 0 {
		return s.noRecipesResponse(req.Language), nil
	}

	// 4️⃣ Для каждого рецепта: matching + DTO
	recipeDTOs := make([]RecipeDTO, 0, len(recipes))
	for _, recipe := range recipes {
		dto := s.buildRecipeDTO(recipe, fridgeIngredientIDs, req.Language)
		recipeDTOs = append(recipeDTOs, dto)
	}

	// 5️⃣ Сортировка по match_percent (DESC)
	sortRecipesByMatchPercent(recipeDTOs)

	// 6️⃣ Ограничить limit
	if req.Limit > 0 && req.Limit < len(recipeDTOs) {
		recipeDTOs = recipeDTOs[:req.Limit]
	}

	// 7️⃣ Определить общее решение (decision)
	decision := analyzeDecision(recipeDTOs)
	summary := getLocalizedSummary(decision, req.Language)

	return &RecipeRecommendationResponse{
		Decision:     decision,
		Summary:      summary,
		TotalMatches: len(recipeDTOs),
		Recipes:      recipeDTOs,
	}, nil
}

// ============================================================================
// CORE LOGIC: Building Recipe DTO
// ============================================================================

// buildRecipeDTO - создает ПОЛНЫЙ DTO рецепта (единый источник правды)
func (s *RecommendationService) buildRecipeDTO(
	recipe models.RecipeCatalog,
	fridgeIngredientIDs map[string]bool,
	lang string,
) RecipeDTO {
	// 1️⃣ Разделить ингредиенты на available / missing
	var available []IngredientInfo
	var missing []IngredientInfo

	for _, recipeIng := range recipe.Ingredients {
		ingredient := recipeIng.Ingredient
		ingredientID := ingredient.ID

		info := IngredientInfo{
			ID:            ingredientID,
			CanonicalName: recipeIng.IngredientKey,
			DisplayName:   ingredient.GetName(lang),
			Quantity:      recipeIng.Quantity,
			Unit:          recipeIng.Unit,
			Category:      ingredient.Category,
		}

		// 🔥 Matching: проверяем КАК ingredient_id, ТАК И canonical_id
		// Это позволяет match "Olej roślinny" и "Olej rzepakowy" (оба = vegetable_oil)
		inFridge := false
		
		// Check 1: Direct match by ingredient_id
		if ingredientID != "" && fridgeIngredientIDs[ingredientID] {
			inFridge = true
		}
		
		// Check 2: Canonical match (e.g., vegetable_oil group)
		if !inFridge && ingredient.CanonicalID != nil && *ingredient.CanonicalID != "" {
			if fridgeIngredientIDs[*ingredient.CanonicalID] {
				inFridge = true
				fmt.Printf("🎯 [CANONICAL MATCH] Recipe needs '%s', matched via canonical_id='%s'\n",
					ingredient.GetName(lang), *ingredient.CanonicalID)
			}
		}
		
		if inFridge {
			available = append(available, info)
		} else {
			missing = append(missing, info)
		}
	}

	// 2️⃣ Рассчитать match_percent
	totalRequired := len(recipe.Ingredients)
	availableCount := len(available)
	matchPercent := 0.0
	if totalRequired > 0 {
		matchPercent = float64(availableCount) / float64(totalRequired) * 100
	}

	// 3️⃣ Определить match_status (только из missing_count)
	matchStatus := classifyMatchStatus(len(missing))

	// 4️⃣ Получить Steps (локализованные)
	steps := s.extractLocalizedSteps(recipe, lang)

	// 5️⃣ Собрать полный DTO
	return RecipeDTO{
		ID:            recipe.ID.String(),
		Title:         recipe.GetLocalizedName(lang),
		CanonicalName: recipe.CanonicalName,
		ImageURL:      &recipe.ImageUrl, // может быть пустой строкой, но не nil
		CookTime:      recipe.TimeMinutes,
		Servings:      recipe.Servings,
		MatchPercent:  matchPercent,
		MatchStatus:   matchStatus,
		AvailableIngredients: available,
		MissingIngredients:   missing,
		Steps:         steps,
		AI:            nil, // Phase 2: добавляется отдельно
	}
}

// ============================================================================
// HELPER METHODS
// ============================================================================

// getUserFridgeIngredientIDs - получает ingredient_id из холодильника
// + загружает canonical_id для группировки похожих ингредиентов
func (s *RecommendationService) getUserFridgeIngredientIDs(
	ctx context.Context,
	userID string,
) (map[string]bool, error) {
	fmt.Printf("🔍 [FRIDGE CHECK] Starting for userID: %s\n", userID)
	
	// Загружаем ingredient_id + canonical_id
	type FridgeItem struct {
		IngredientID string
		CanonicalID  *string
	}
	
	var items []FridgeItem
	err := s.db.WithContext(ctx).
		Table("user_fridge_items AS ufi").
		Select("ufi.ingredient_id, i.canonical_id").
		Joins(`LEFT JOIN "Ingredient" AS i ON i.id = ufi.ingredient_id`).
		Where("ufi.user_id = ? AND ufi.quantity > 0", userID).
		Scan(&items).
		Error

	if err != nil {
		fmt.Printf("❌ [FRIDGE CHECK] Database error: %v\n", err)
		return nil, err
	}

	fmt.Printf("✅ [FRIDGE CHECK] Found %d ingredients in fridge for user %s\n", len(items), userID)

	// Создаем map для O(1) lookup
	// Добавляем КАК ingredient_id, ТАК И canonical_id
	fridgeSet := make(map[string]bool)
	for _, item := range items {
		// Direct match by ingredient_id
		if item.IngredientID != "" {
			fridgeSet[item.IngredientID] = true
		}
		
		// Canonical match (e.g., vegetable_oil matches any oil type)
		if item.CanonicalID != nil && *item.CanonicalID != "" {
			fridgeSet[*item.CanonicalID] = true
			fmt.Printf("📦 [FRIDGE CHECK] Canonical group: %s (ingredient_id: %s)\n", 
				*item.CanonicalID, item.IngredientID)
		}
	}

	fmt.Printf("📊 [FRIDGE CHECK] Total keys in fridgeSet: %d (includes canonical groups)\n", len(fridgeSet))

	return fridgeSet, nil
}

// extractLocalizedSteps - извлекает шаги на нужном языке
func (s *RecommendationService) extractLocalizedSteps(
	recipe models.RecipeCatalog,
	lang string,
) []string {
	var stepsJSON []byte

	switch lang {
	case "ru":
		stepsJSON = recipe.StepsRu
	case "en":
		stepsJSON = recipe.StepsEn
	case "pl":
		stepsJSON = recipe.StepsPl
	default:
		stepsJSON = recipe.StepsPl // default fallback
	}

	if len(stepsJSON) == 0 {
		return []string{}
	}

	// Parse JSON: [{"text":"...", "order":1}, ...]
	var stepsData []struct {
		Text  string `json:"text"`
		Order int    `json:"order"`
	}

	if err := json.Unmarshal(stepsJSON, &stepsData); err != nil {
		return []string{}
	}

	// Извлечь только text
	steps := make([]string, 0, len(stepsData))
	for _, step := range stepsData {
		steps = append(steps, step.Text)
	}

	return steps
}

// classifyMatchStatus - классифицирует статус по количеству missing
func classifyMatchStatus(missingCount int) string {
	switch {
	case missingCount == 0:
		return StatusReady // 🟢 готово
	case missingCount <= 2:
		return StatusAlmostReady // 🟡 почти готово (1-2 missing)
	default:
		return StatusNotReady // 🔴 не хватает (3+ missing)
	}
}

// analyzeDecision - анализирует общую ситуацию
func analyzeDecision(recipes []RecipeDTO) string {
	if len(recipes) == 0 {
		return DecisionNeedMore
	}

	// Проверяем лучший результат (первый, т.к. отсортировано)
	best := recipes[0]

	switch best.MatchStatus {
	case StatusReady:
		return DecisionReady // 🟢 Есть готовые рецепты
	case StatusAlmostReady:
		return DecisionAlmostReady // 🟡 Почти готовы
	default:
		return DecisionNeedMore // 🔴 Нужно больше продуктов
	}
}

// sortRecipesByMatchPercent - сортировка по match_percent (DESC)
func sortRecipesByMatchPercent(recipes []RecipeDTO) {
	n := len(recipes)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if recipes[j].MatchPercent < recipes[j+1].MatchPercent {
				recipes[j], recipes[j+1] = recipes[j+1], recipes[j]
			}
		}
	}
}

// getLocalizedSummary - возвращает локализованное резюме
func getLocalizedSummary(decision string, lang string) string {
	summaries := map[string]map[string]string{
		DecisionReady: {
			"pl": "Świetnie! Możesz przygotować kilka przepisów już teraz.",
			"en": "Great! You can cook several recipes right now.",
			"ru": "Отлично! Вы можете приготовить несколько рецептов прямо сейчас.",
		},
		DecisionAlmostReady: {
			"pl": "Prawie gotowe! Brakuje tylko kilku składników.",
			"en": "Almost ready! Just a few ingredients missing.",
			"ru": "Почти готово! Не хватает всего нескольких ингредиентов.",
		},
		DecisionNeedMore: {
			"pl": "Potrzebujesz więcej produktów, aby przygotować te przepisy.",
			"en": "You need more ingredients to cook these recipes.",
			"ru": "Вам нужно больше продуктов, чтобы приготовить эти рецепты.",
		},
	}

	if langMap, ok := summaries[decision]; ok {
		if summary, ok := langMap[lang]; ok {
			return summary
		}
	}

	return "No recommendations available."
}

// emptyFridgeResponse - ответ для пустого холодильника
func (s *RecommendationService) emptyFridgeResponse(lang string) *RecipeRecommendationResponse {
	return &RecipeRecommendationResponse{
		Decision:     DecisionNeedMore,
		Summary:      getLocalizedSummary(DecisionNeedMore, lang),
		TotalMatches: 0,
		Recipes:      []RecipeDTO{},
	}
}

// noRecipesResponse - ответ когда в каталоге нет рецептов
func (s *RecommendationService) noRecipesResponse(lang string) *RecipeRecommendationResponse {
	return &RecipeRecommendationResponse{
		Decision:     DecisionNeedMore,
		Summary:      "No recipes available in catalog.",
		TotalMatches: 0,
		Recipes:      []RecipeDTO{},
	}
}

// ============================================================================
// GET SINGLE RECIPE WITH FRIDGE CHECK
// ============================================================================

// GetSingleRecipeWithFridge - получает ОДИН рецепт с проверкой холодильника
// Используется на странице /recipes/[id] для показа inFridge статуса
func (s *RecommendationService) GetSingleRecipeWithFridge(
	ctx context.Context,
	req RecipeMatchRequest,
) (*RecipeDTO, error) {
	fmt.Printf("🎯 [GET SINGLE RECIPE] Request: userID=%s, recipeID=%s, lang=%s\n", 
		req.UserID, req.RecipeID, req.Language)
	
	// Валидация
	if req.RecipeID == "" {
		return nil, fmt.Errorf("recipe_id is required")
	}
	if req.Language == "" {
		req.Language = "pl"
	}

	// 1️⃣ Получить холодильник пользователя
	fmt.Printf("📦 [GET SINGLE RECIPE] Step 1: Getting fridge for user %s\n", req.UserID)
	fridgeIngredientIDs, err := s.getUserFridgeIngredientIDs(ctx, req.UserID)
	if err != nil {
		fmt.Printf("❌ [GET SINGLE RECIPE] Fridge error: %v\n", err)
		return nil, fmt.Errorf("failed to get fridge: %w", err)
	}
	fmt.Printf("✅ [GET SINGLE RECIPE] Fridge loaded: %d ingredients\n", len(fridgeIngredientIDs))

	// 2️⃣ Получить рецепт (try UUID first, then canonical_name)
	fmt.Printf("🍳 [GET SINGLE RECIPE] Step 2: Getting recipe %s\n", req.RecipeID)
	recipe, err := s.recipeRepository.GetRecipeByIDOrCanonical(ctx, req.RecipeID)
	if err != nil {
		fmt.Printf("❌ [GET SINGLE RECIPE] Recipe error: %v\n", err)
		return nil, fmt.Errorf("recipe not found: %w", err)
	}
	fmt.Printf("✅ [GET SINGLE RECIPE] Recipe found: %s (%d ingredients)\n", 
		recipe.CanonicalName, len(recipe.Ingredients))

	// 3️⃣ Построить DTO с проверкой холодильника
	fmt.Printf("🔨 [GET SINGLE RECIPE] Step 3: Building DTO with fridge check\n")
	dto := s.buildRecipeDTO(*recipe, fridgeIngredientIDs, req.Language)
	
	fmt.Printf("✅ [GET SINGLE RECIPE] DTO built: %d available, %d missing, %.2f%% match\n",
		len(dto.AvailableIngredients), len(dto.MissingIngredients), dto.MatchPercent)

	return &dto, nil
}
