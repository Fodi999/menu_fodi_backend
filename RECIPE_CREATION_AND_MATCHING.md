# 🍳 Создание рецептов и подбор блюд по холодильнику

**Дата:** 2026-01-26  
**Статус:** ✅ Работает  

---

## 📋 Оглавление

1. [Создание рецепта](#1-создание-рецепта)
2. [Подбор блюд по холодильнику](#2-подбор-блюд-по-холодильнику)
3. [Алгоритм matching](#3-алгоритм-matching)
4. [Примеры использования](#4-примеры-использования)

---

## 1️⃣ Создание рецепта

### Вариант A: Простое создание (POST /api/recipes)

**Endpoint:** `POST /api/recipes`

**Авторизация:** ✅ Требуется (любой авторизованный пользователь)

**Request:**

```json
{
  "title": "Pasta Carbonara",
  "description": "Classic Italian pasta dish",
  "imageUrl": "https://example.com/image.jpg",
  "country": "Italy",
  "category": "Main Course",
  "difficulty": "medium",
  "timeMinutes": 30,
  "servings": 4,
  "calories": 450,
  "protein": 20.5,
  "fats": 15.2,
  "carbs": 55.0
}
```

**Response:** `201 Created`

```json
{
  "status": "success",
  "data": {
    "id": "uuid",
    "title": "Pasta Carbonara",
    "authorId": "user-uuid",
    "createdAt": "2026-01-26T19:00:00Z",
    "viewsCount": 0
  }
}
```

**Реализация:**

```54:134:/Users/dmitrijfomin/Desktop/backend/internal/modules/recipes/transport/http/handlers.go
func (h *RecipeHandlers) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title        string   `json:"title"`
		Description  string   `json:"description"`
		ImageUrl     string   `json:"imageUrl"`
		Country      string   `json:"country"`     // Required
		Category     string   `json:"category"`    // Required
		Difficulty   string   `json:"difficulty"`  // Required
		TimeMinutes  int      `json:"timeMinutes"` // Required
		Servings     int      `json:"servings"`    // Required
		GrossWeight  *int     `json:"grossWeight"`
		NetWeight    *int     `json:"netWeight"`
		Calories     *int     `json:"calories"`
		Protein      *float64 `json:"protein"`
		Fats         *float64 `json:"fats"`
		Carbs        *float64 `json:"carbs"`
		RecipeYield  *int     `json:"yield"`
		Cost         *float64 `json:"cost"`
		TokensReward *int     `json:"tokensReward"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	claims, ok := r.Context().Value(middleware.UserContextKey).(*authservice.Claims)
	if !ok || claims.Subject == "" {
		utils.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}
	authorID := claims.Subject

	if input.Title == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Title is required")
		return
	}

	db := database.GetDB()
	var author models.User
	if err := db.First(&author, "id = ?", authorID).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Author not found")
		return
	}

	recipe := models.Recipe{
		ID:           uuid.New().String(),
		LocalName:    input.Title, // Use title as localName for user recipes
		Title:        input.Title,
		Description:  input.Description,
		ImageUrl:     input.ImageUrl,
		Country:      input.Country,                               // Required field
		Category:     input.Category,                              // Required field
		Difficulty:   input.Difficulty,                            // Required field
		TimeMinutes:  input.TimeMinutes,                           // Required field
		Servings:     input.Servings,                              // Required field
		Source:       datatypes.JSON([]byte(`{"type":"manual"}`)), // User-generated recipe
		AuthorID:     authorID,
		GrossWeight:  input.GrossWeight,
		NetWeight:    input.NetWeight,
		Calories:     input.Calories,
		Protein:      input.Protein,
		Fats:         input.Fats,
		Carbs:        input.Carbs,
		RecipeYield:  input.RecipeYield,
		Cost:         input.Cost,
		TokensReward: input.TokensReward,
		ViewsCount:   0,
		TokensEarned: 0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := db.Create(&recipe).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create recipe")
		return
	}

	db.Preload("Author").First(&recipe, "id = ?", recipe.ID)
	utils.RespondWithJSON(w, http.StatusCreated, map[string]interface{}{"status": "success", "data": recipe})
}
```

---

### Вариант B: Создание через AI (POST /api/admin/recipes/create-ai)

**Endpoint:** `POST /api/admin/recipes/create-ai`

**Авторизация:** ✅ Требуется (admin/super_admin)

**Request (минимальный):**

```json
{
  "title": "Pasta Carbonara",
  "language": "en",
  "ingredients": [
    {
      "ingredientId": "uuid-pasta",
      "quantity": 400,
      "unit": "g"
    },
    {
      "ingredientId": "uuid-eggs",
      "quantity": 4,
      "unit": "шт"
    }
  ],
  "rawCookingText": "Boil pasta. Mix eggs with cheese. Combine with hot pasta."
}
```

**Response:** `200 OK`

```json
{
  "id": "recipe-uuid",
  "title": "Pasta Carbonara",
  "description": "Classic Italian pasta with eggs and cheese",
  "steps": [
    {
      "stepNumber": 1,
      "text": "Boil 3L of salted water in a large pot"
    },
    {
      "stepNumber": 2,
      "text": "Cook pasta for 8-10 minutes until al dente"
    }
  ],
  "ingredients": [...],
  "timeMinutes": 30,
  "difficulty": "medium",
  "servings": 4
}
```

**Процесс создания через AI (5 этапов):**

```106:143:/Users/dmitrijfomin/Desktop/backend/internal/modules/admin/service/recipe_ai.go
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
```

---

## 2️⃣ Подбор блюд по холодильнику

### Endpoint A: Matching с фильтрами (GET /api/recipes/match)

**Endpoint:** `GET /api/recipes/match`

**Авторизация:** ✅ Требуется (доступ к холодильнику пользователя)

**Query Parameters:**

```
?country=Poland
&category=Main Course
&maxTime=60
&minScore=50
&onlyCookable=true
&limit=10
```

**Response:**

```json
{
  "success": true,
  "data": {
    "recipes": [
      {
        "id": "recipe-uuid",
        "title": "Pasta Carbonara",
        "matchScore": 95.5,
        "canCookNow": true,
        "matchedIngredients": [
          {
            "ingredientId": "uuid-pasta",
            "name": "Pasta",
            "required": 400,
            "available": 500,
            "unit": "g",
            "isExpiringSoon": false
          }
        ],
        "missingIngredients": [],
        "costToComplete": 0,
        "usedValue": 12.50,
        "timeMinutes": 30
      }
    ],
    "count": 1
  }
}
```

**Реализация:**

```87:158:/Users/dmitrijfomin/Desktop/backend/internal/modules/recipes/transport/http/handler.go
// MatchRecipes finds recipes matching user's fridge
// GET /api/recipes/match?country=Poland&maxTime=60&excludeAllergens=gluten,lactose&minScore=50
func (h *RecipeHandler) MatchRecipes(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		// DEV MODE: Allow testing without auth by using test user ID from query
		testUserID := r.URL.Query().Get("testUserID")
		if testUserID != "" {
			h.logger.Warn("⚠️ DEV MODE: Using test userID from query parameter")
			userID = testUserID
		} else {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// Parse query parameters
	filters := service.RecipeFilters{
		Country:          r.URL.Query().Get("country"),
		Category:         r.URL.Query().Get("category"),
		Difficulty:       r.URL.Query().Get("difficulty"),
		MaxTime:          parseIntQuery(r, "maxTime", 0),
		ExcludeAllergens: parseArrayQuery(r, "excludeAllergens"),
		IncludeDietTags:  parseArrayQuery(r, "dietTags"),
		MinScore:         parseFloatQuery(r, "minScore", 0.0),
		OnlyCookable:     parseBoolQuery(r, "onlyCookable", false),
		Limit:            parseIntQuery(r, "limit", 10),
	}

	h.logger.Info("Matching recipes with fridge",
		zap.String("userId", userID),
		zap.Any("filters", filters),
	)

	// Find matching recipes
	matches, err := h.matchService.MatchRecipesWithFridge(userID, filters)
	if err != nil {
		h.logger.Error("Failed to match recipes", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(dto.RecipeMatchResponse{
			Success: false,
			Error:   "Failed to find recipes",
		})
		return
	}

	h.logger.Info("Found matching recipes",
		zap.String("userId", userID),
		zap.Int("count", len(matches)),
	)

	// Get user's preferred language
	userLang := h.getUserLanguage(r)

	// Convert to DTO format with localization
	recipeItems := make([]dto.RecipeMatchItem, len(matches))
	for i, match := range matches {
		recipeItems[i] = convertToRecipeMatchItem(match, userLang)
	}

	// Return results with standard contract
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.RecipeMatchResponse{
		Success: true,
		Data: &dto.RecipeMatchData{
			Recipes: recipeItems,
			Count:   len(recipeItems),
		},
	})
}
```

---

### Endpoint B: Категоризация рецептов (GET /api/recipes/available)

**Endpoint:** `GET /api/recipes/available`

**Авторизация:** ✅ Требуется

**Response:**

```json
{
  "success": true,
  "data": {
    "canCook": [
      {
        "id": "uuid",
        "title": "Pasta Carbonara",
        "matchScore": 100,
        "canCook": true,
        "timeMinutes": 30,
        "costToComplete": 0
      }
    ],
    "almostReady": [
      {
        "id": "uuid",
        "title": "Pizza Margherita",
        "matchScore": 85,
        "canCook": false,
        "missingCount": 1,
        "costToComplete": 3.50
      }
    ],
    "needMore": [...],
    "canCookCount": 5,
    "almostReadyCount": 12,
    "needMoreCount": 18
  }
}
```

**Категории:**
- `canCook` — можно готовить прямо сейчас (все обязательные ингредиенты есть)
- `almostReady` — почти готово (match score 50-99%)
- `needMore` — нужно докупить (match score < 50%)

---

## 3️⃣ Алгоритм matching

### Основной метод: MatchRecipesWithFridge

```64:125:/Users/dmitrijfomin/Desktop/backend/internal/modules/recipes/service/match_service.go
// MatchRecipesWithFridge finds recipes that match user's fridge contents
func (s *RecipeMatchService) MatchRecipesWithFridge(
	userID string,
	filters RecipeFilters,
) ([]RecipeMatch, error) {
	// 1. Load user's fridge with prices
	fridgeItems, err := s.loadFridgeWithPrices(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load fridge: %w", err)
	}

	if len(fridgeItems) == 0 {
		return []RecipeMatch{}, nil
	}

	// 2. Build ingredient map for fast lookup by ingredientId
	fridgeMap := make(map[string]*FridgeItem)
	for i := range fridgeItems {
		// Use ingredientId as key for precise matching
		fridgeMap[fridgeItems[i].ID] = &fridgeItems[i]

		// Also add normalized name as fallback for compatibility
		key := normalizeIngredientName(fridgeItems[i].Name)
		if _, exists := fridgeMap[key]; !exists {
			fridgeMap[key] = &fridgeItems[i]
		}
	}

	// 3. Load recipes from catalog with filters
	recipes, err := s.loadRecipesWithFilters(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to load recipes: %w", err)
	}

	// 4. Calculate match score for each recipe
	matches := make([]RecipeMatch, 0, len(recipes))
	for _, recipe := range recipes {
		match := s.calculateRecipeMatch(recipe, fridgeMap)

		// Apply minimum score threshold
		if match.MatchScore < filters.MinScore {
			continue
		}

		// Apply cookable filter if requested
		if filters.OnlyCookable && !match.CanMakeNow {
			continue
		}

		matches = append(matches, match)
	}

	// 5. Sort by: canCookNow DESC -> score DESC -> costToComplete ASC -> timeMinutes ASC
	sortRecipeMatches(matches)

	// 6. Return top N results
	if filters.Limit > 0 && len(matches) > filters.Limit {
		matches = matches[:filters.Limit]
	}

	return matches, nil
}
```

### Расчет match score

```128:275:/Users/dmitrijfomin/Desktop/backend/internal/modules/recipes/service/match_service.go
// calculateRecipeMatch calculates match score and details
func (s *RecipeMatchService) calculateRecipeMatch(
	recipe models.RecipeCatalog,
	fridgeMap map[string]*FridgeItem,
) RecipeMatch {
	match := RecipeMatch{
		Recipe:             recipe,
		MatchedIngredients: []MatchedIngredient{},
		MissingIngredients: []MissingIngredient{},
		CostToComplete:     0,
		HasExpiringItems:   false,
		ExpiringItemsCount: 0,
		UsedValue:          0,
		SavedMoney:         0,
		TotalRecipeCost:    0,
		WasteRiskSaved:     0,
	}

	requiredCount := 0
	matchedCount := 0
	optionalMatchedCount := 0

	// DEBUG: Log recipe matching attempt
	fmt.Printf("🔍 Matching recipe: %s (canonicalName=%s, ingredients=%d)\n",
		recipe.ID.String(), recipe.CanonicalName, len(recipe.Ingredients))

	// CRITICAL: Skip recipes without ingredients (invalid data)
	if len(recipe.Ingredients) == 0 {
		fmt.Printf("  ⚠️  SKIPPED: Recipe has no ingredients (invalid data)\n\n")
		match.CanMakeNow = false
		match.MatchScore = 0
		return match
	}

	for _, recipeIng := range recipe.Ingredients {
		fmt.Printf("  → Ingredient: id=%s, quantity=%.2f %s, optional=%v\n",
			recipeIng.IngredientID, recipeIng.Quantity, recipeIng.Unit, recipeIng.Optional)

		// CRITICAL: Skip ingredients with invalid quantity (data quality issue)
		if recipeIng.Quantity <= 0 {
			fmt.Printf("    ⚠️  INVALID: quantity <= 0 (skipping this ingredient)\n")
			continue
		}

		if recipeIng.Optional {
			// Optional ingredients don't affect core match score
			fridgeItem := s.findIngredientInFridge(recipeIng, fridgeMap)
			if fridgeItem != nil {
				optionalMatchedCount++
				match.MatchedIngredients = append(match.MatchedIngredients, MatchedIngredient{
					IngredientID:   recipeIng.IngredientID,
					Name:           recipeIng.Ingredient.Name,
					Required:       recipeIng.Quantity,
					Available:      fridgeItem.Quantity,
					Unit:           recipeIng.Unit,
					IsExpiringSoon: fridgeItem.IsExpiringSoon,
					ExpiresAt:      fridgeItem.ExpiresAt,
					Optional:       true,
					Ingredient:     &recipeIng.Ingredient, // Store full ingredient for localization
				})
			}
			continue
		}

		requiredCount++
		fridgeItem := s.findIngredientInFridge(recipeIng, fridgeMap)

		if fridgeItem != nil && fridgeItem.Quantity >= recipeIng.Quantity {
			// Ingredient available in sufficient quantity
			matchedCount++
			fmt.Printf("    ✅ MATCHED: found in fridge (available=%.2f %s, required=%.2f %s)\n",
				fridgeItem.Quantity, fridgeItem.Unit, recipeIng.Quantity, recipeIng.Unit)

			// Calculate value of used ingredient
			ingredientValue := recipeIng.Quantity * fridgeItem.PricePerUnit
			match.UsedValue += ingredientValue

			// Track expiring items value (waste prevention)
			if fridgeItem.IsExpiringSoon {
				match.HasExpiringItems = true
				match.ExpiringItemsCount++
				match.WasteRiskSaved += ingredientValue
			}

			matched := MatchedIngredient{
				IngredientID:   recipeIng.IngredientID,
				Name:           recipeIng.Ingredient.Name,
				Required:       recipeIng.Quantity,
				Available:      fridgeItem.Quantity,
				Unit:           recipeIng.Unit,
				IsExpiringSoon: fridgeItem.IsExpiringSoon,
				ExpiresAt:      fridgeItem.ExpiresAt,
				Optional:       false,
				Ingredient:     &recipeIng.Ingredient, // Store full ingredient for localization
			}
			match.MatchedIngredients = append(match.MatchedIngredients, matched)
		} else {
			// Ingredient missing or insufficient
			if fridgeItem != nil {
				fmt.Printf("    ❌ INSUFFICIENT: found but not enough (available=%.2f, required=%.2f)\n",
					fridgeItem.Quantity, recipeIng.Quantity)
			} else {
				fmt.Printf("    ❌ NOT FOUND: ingredient not in fridge\n")
			}

			pricePerUnit := 0.0
			if recipeIng.Ingredient.DefaultPricePerUnit != nil {
				pricePerUnit = *recipeIng.Ingredient.DefaultPricePerUnit
			}
			estimatedCost := roundToTwoDecimals(recipeIng.Quantity * pricePerUnit)
			missing := MissingIngredient{
				IngredientID:  recipeIng.IngredientID,
				Name:          recipeIng.Ingredient.Name,
				Required:      recipeIng.Quantity,
				Unit:          recipeIng.Unit,
				EstimatedCost: estimatedCost,
				Optional:      recipeIng.Optional,
				Ingredient:    &recipeIng.Ingredient, // Store full ingredient for localization
			}
			match.MissingIngredients = append(match.MissingIngredients, missing)
			match.CostToComplete += estimatedCost
		}
	}

	// Calculate base match score
	if requiredCount > 0 {
		match.MatchScore = (float64(matchedCount) / float64(requiredCount)) * 100
	}

	// Bonus for optional ingredients
	if len(recipe.Ingredients) > requiredCount && optionalMatchedCount > 0 {
		optionalBonus := (float64(optionalMatchedCount) / float64(len(recipe.Ingredients)-requiredCount)) * 5
		match.MatchScore += optionalBonus
	}

	// Bonus for using expiring items (prioritize waste reduction)
	if match.HasExpiringItems {
		expiryBonus := float64(match.ExpiringItemsCount) * 2.0
		match.MatchScore += expiryBonus
	}

	// Cap at 100
	if match.MatchScore > 100 {
		match.MatchScore = 100
	}

	// Round score to 2 decimals
	match.MatchScore = roundToTwoDecimals(match.MatchScore)
```

---

## 📊 Алгоритм matching (детально)

### Шаг 1: Загрузка холодильника

```go
// Загружаем холодильник пользователя с ценами
fridgeItems := loadFridgeWithPrices(userID)
// → [{id, name, quantity, unit, pricePerUnit, expiresAt}, ...]
```

### Шаг 2: Создание индекса

```go
fridgeMap := make(map[string]*FridgeItem)
for item := range fridgeItems {
    fridgeMap[item.ID] = &item              // По ID ингредиента
    fridgeMap[normalize(item.Name)] = &item // По нормализованному имени (fallback)
}
```

### Шаг 3: Загрузка рецептов

```go
recipes := loadRecipesWithFilters(filters)
// Фильтры: country, category, maxTime, difficulty, allergens
```

### Шаг 4: Расчет match score для каждого рецепта

```
Для каждого рецепта:
  requiredCount = 0
  matchedCount = 0
  
  Для каждого обязательного ингредиента:
    requiredCount++
    
    Если ингредиент есть в холодильнике И количество достаточное:
      matchedCount++
      ✅ Добавить в matchedIngredients[]
      💰 Рассчитать стоимость использованного ингредиента
      ⏰ Проверить, истекает ли срок годности
    Иначе:
      ❌ Добавить в missingIngredients[]
      💰 Рассчитать стоимость докупки
  
  matchScore = (matchedCount / requiredCount) * 100
  
  Бонусы:
    + 5% за каждый опциональный ингредиент (если есть)
    + 2% за каждый ингредиент с истекающим сроком
  
  canCookNow = (matchedCount == requiredCount)
```

### Шаг 5: Сортировка результатов

**Приоритет сортировки:**

1. **canCookNow** (DESC) — рецепты, которые можно готовить сейчас
2. **matchScore** (DESC) — процент совпадения
3. **costToComplete** (ASC) — стоимость докупки
4. **timeMinutes** (ASC) — время приготовления

```go
sort.Slice(matches, func(i, j int) bool {
    // 1. Recipes you can cook NOW go first
    if matches[i].CanMakeNow != matches[j].CanMakeNow {
        return matches[i].CanMakeNow
    }
    
    // 2. Higher match score
    if matches[i].MatchScore != matches[j].MatchScore {
        return matches[i].MatchScore > matches[j].MatchScore
    }
    
    // 3. Lower cost to complete
    if matches[i].CostToComplete != matches[j].CostToComplete {
        return matches[i].CostToComplete < matches[j].CostToComplete
    }
    
    // 4. Faster recipes
    return matches[i].Recipe.TimeMinutes < matches[j].Recipe.TimeMinutes
})
```

### Шаг 6: Возврат топ N результатов

```go
if filters.Limit > 0 && len(matches) > filters.Limit {
    matches = matches[:filters.Limit]
}
return matches
```

---

## 📝 Структуры данных

### RecipeMatch (результат matching)

```go
type RecipeMatch struct {
    Recipe             models.RecipeCatalog  // Полный рецепт
    MatchedIngredients []MatchedIngredient   // Ингредиенты, которые есть
    MissingIngredients []MissingIngredient   // Ингредиенты, которых нет
    MatchScore         float64               // 0-100 (процент совпадения)
    CanMakeNow         bool                  // true если все required есть
    CostToComplete     float64               // Стоимость докупки
    UsedValue          float64               // Стоимость использованных ингредиентов
    HasExpiringItems   bool                  // Есть ли продукты с истекающим сроком
    ExpiringItemsCount int                   // Количество истекающих продуктов
    WasteRiskSaved     float64               // Стоимость продуктов, спасённых от выброса
}
```

### MatchedIngredient (есть в холодильнике)

```go
type MatchedIngredient struct {
    IngredientID   string    // UUID ингредиента
    Name           string    // Название
    Required       float64   // Сколько нужно для рецепта
    Available      float64   // Сколько есть в холодильнике
    Unit           string    // Единица измерения
    IsExpiringSoon bool      // Истекает ли срок годности
    ExpiresAt      *time.Time // Дата истечения
    Optional       bool      // Опциональный ингредиент
}
```

### MissingIngredient (нужно докупить)

```go
type MissingIngredient struct {
    IngredientID  string  // UUID ингредиента
    Name          string  // Название
    Required      float64 // Сколько нужно
    Unit          string  // Единица измерения
    EstimatedCost float64 // Стоимость покупки
    Optional      bool    // Опциональный ингредиент
}
```

---

## 4️⃣ Примеры использования

### Пример 1: Создание простого рецепта

```bash
curl -X POST http://localhost:8080/api/recipes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Pasta Carbonara",
    "description": "Classic Italian pasta",
    "country": "Italy",
    "category": "Main Course",
    "difficulty": "medium",
    "timeMinutes": 30,
    "servings": 4,
    "calories": 450
  }'

# Response: 201 Created
{
  "status": "success",
  "data": {
    "id": "uuid",
    "title": "Pasta Carbonara",
    "authorId": "user-uuid",
    "createdAt": "2026-01-26T19:00:00Z"
  }
}
```

### Пример 2: Подбор рецептов по холодильнику

```bash
# Найти рецепты, которые можно приготовить прямо сейчас
curl -X GET "http://localhost:8080/api/recipes/match?onlyCookable=true&limit=5" \
  -H "Authorization: Bearer $TOKEN"

# Response
{
  "success": true,
  "data": {
    "recipes": [
      {
        "id": "uuid-1",
        "title": "Pasta Carbonara",
        "matchScore": 100,
        "canCookNow": true,
        "matchedIngredients": [
          {
            "name": "Pasta",
            "required": 400,
            "available": 500,
            "unit": "g"
          },
          {
            "name": "Eggs",
            "required": 4,
            "available": 6,
            "unit": "шт"
          }
        ],
        "missingIngredients": [],
        "costToComplete": 0,
        "timeMinutes": 30
      }
    ],
    "count": 1
  }
}
```

### Пример 3: Категоризация рецептов

```bash
# Получить все рецепты, разбитые по категориям
curl -X GET http://localhost:8080/api/recipes/available \
  -H "Authorization: Bearer $TOKEN"

# Response
{
  "success": true,
  "data": {
    "canCook": [
      {"id": "uuid-1", "title": "Pasta Carbonara", "matchScore": 100}
    ],
    "almostReady": [
      {"id": "uuid-2", "title": "Pizza", "matchScore": 85, "missingCount": 1}
    ],
    "needMore": [
      {"id": "uuid-3", "title": "Sushi", "matchScore": 30, "missingCount": 5}
    ],
    "canCookCount": 5,
    "almostReadyCount": 12,
    "needMoreCount": 18
  }
}
```

### Пример 4: Фильтрация по стране и времени

```bash
# Найти итальянские рецепты, которые готовятся <= 45 минут
curl -X GET "http://localhost:8080/api/recipes/match?country=Italy&maxTime=45&minScore=70" \
  -H "Authorization: Bearer $TOKEN"
```

---

## 🔍 Логирование и отладка

### Логи при matching

```
🔍 Matching recipe: abc-123 (canonicalName=Pasta Carbonara, ingredients=5)
  → Ingredient: id=pasta-uuid, quantity=400.00 g, optional=false
    ✅ MATCHED: found in fridge (available=500.00 g, required=400.00 g)
  → Ingredient: id=eggs-uuid, quantity=4.00 шт, optional=false
    ✅ MATCHED: found in fridge (available=6.00 шт, required=4.00 шт)
  → Ingredient: id=cheese-uuid, quantity=100.00 g, optional=false
    ❌ NOT FOUND: ingredient not in fridge

✅ Match result:
   matchScore: 66.67%
   canCookNow: false
   matched: 2/3
   costToComplete: 5.50 PLN
```

---

## 📋 Эндпоинты сводка

### Создание рецептов

| Endpoint                           | Метод | Auth     | Описание                  |
|------------------------------------|-------|----------|---------------------------|
| `/api/recipes`                     | POST  | Required | Простое создание рецепта  |
| `/api/admin/recipes/create-ai`     | POST  | Admin    | Создание через AI         |

### Подбор рецептов

| Endpoint                      | Метод | Auth     | Описание                           |
|-------------------------------|-------|----------|------------------------------------|
| `/api/recipes/match`          | GET   | Required | Matching с фильтрами               |
| `/api/recipes/available`      | GET   | Required | Категоризация (canCook/almost/need)|
| `/api/recipes/recommendations`| POST  | Required | AI рекомендации                    |
| `/api/recipes/{id}`           | GET   | Optional | Детали рецепта с флагами inFridge |

---

## ✅ Итог

### Что работает

1. **Создание рецептов:**
   - ✅ Простое создание через POST
   - ✅ Создание через AI (admin)
   - ✅ Автор сохраняется в authorID

2. **Подбор по холодильнику:**
   - ✅ Точный matching по ID ингредиентов
   - ✅ Fallback на нормализованные названия
   - ✅ Расчет match score (0-100%)
   - ✅ Определение canCookNow
   - ✅ Расчет стоимости докупки
   - ✅ Приоритизация продуктов с истекающим сроком
   - ✅ Сортировка по приоритету

3. **Фильтрация:**
   - ✅ По стране, категории, сложности
   - ✅ По времени приготовления
   - ✅ По аллергенам
   - ✅ По минимальному match score
   - ✅ Только готовые к приготовлению

### Ключевые особенности

- ✅ БД как источник истины
- ✅ Точный matching по UUID ингредиентов
- ✅ Экономическая модель (стоимость, экономия, предотвращение waste)
- ✅ Приоритизация продуктов с истекающим сроком
- ✅ Поддержка опциональных ингредиентов
- ✅ Мультиязычность

---

**Статус:** ✅ Реализовано  
**Документация:** Полная  
**Тестирование:** Готово к использованию
