# ✅ Kitchen OS: canCookNow Validation Implemented

**Дата:** 22 января 2026  
**Commit:** `5a785b4` - "fix: correct SQL column name in canCookNow check (recipeId quoted)"  
**Статус:** ✅ DEPLOYED

---

## 🎯 Что Реализовано

### Backend = Source of Truth ✅

**Строгое правило:**
```
IF canCookNow != true
→ 409 Conflict
→ { "error": "insufficient_ingredients", "missing_ingredients": [...] }
```

**Перед добавлением рецепта в меню:**
1. ✅ Backend загружает ингредиенты рецепта из `RecipeIngredient`
2. ✅ Backend проверяет холодильник пользователя `FridgeItem`
3. ✅ Backend сопоставляет через `canonical_id` (например, "vegetable_oil")
4. ✅ Если чего-то НЕ хватает → 409 Conflict с списком недостающих ингредиентов
5. ✅ Если ВСЁ есть → рецепт добавляется в меню (status=`planned`)

---

## 🔧 Техническая Реализация

### 1. Custom Error Type

```go
// internal/modules/menu/service/menu_service.go

type InsufficientIngredientsError struct {
	RecipeID           uuid.UUID
	MissingIngredients []string
}

func (e *InsufficientIngredientsError) Error() string {
	return fmt.Sprintf("cannot add to menu: missing ingredients: %s", 
		strings.Join(e.MissingIngredients, ", "))
}
```

### 2. Validation Logic in AddToMenu

```go
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
	
	// ✅ CRITICAL: Check if user can cook this recipe NOW
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
	
	// Create menu item (only if canCook == true)
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
	
	// Return response
	// ...
}
```

### 3. checkCanCookNow Implementation

```go
func (s *MenuService) checkCanCookNow(
	ctx context.Context, 
	userID string, 
	recipeID uuid.UUID,
) (bool, []string, error) {
	// 1. Get recipe ingredients
	var recipeIngredients []models.CatalogIngredient
	err := s.db.WithContext(ctx).
		Table("RecipeIngredient").
		Preload("Ingredient").
		Where("\"recipeId\" = ?", recipeID).
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
			continue
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
```

### 4. HTTP Handler (409 Conflict)

```go
// internal/modules/menu/transport/http/menu_handler.go

func (h *MenuHandler) AddToMenu(w http.ResponseWriter, r *http.Request) {
	// ... parse request ...
	
	item, err := h.service.AddToMenu(r.Context(), userID, recipeID, servings, notes)
	if err != nil {
		// Check if it's InsufficientIngredientsError
		var insufficientErr *service.InsufficientIngredientsError
		if errors.As(err, &insufficientErr) {
			// 409 Conflict: Cannot cook - missing ingredients
			utils.RespondJSON(w, http.StatusConflict, map[string]interface{}{
				"error":               "insufficient_ingredients",
				"message":             "Cannot add to menu: missing ingredients in fridge",
				"missing_ingredients": insufficientErr.MissingIngredients,
			})
			return
		}
		
		// Other errors
		utils.RespondError(w, http.StatusInternalServerError, "failed to add to menu", err.Error())
		return
	}
	
	utils.RespondJSON(w, http.StatusCreated, item)
}
```

---

## 📊 Canonical Ingredient Matching

**Как работает сопоставление:**

### Example: Vegetable Oil (растительное масло)

**В рецепте:**
```json
{
  "ingredientId": "ing_001",
  "ingredient": {
    "id": "ing_001",
    "name": "растительное масло",
    "canonicalId": "vegetable_oil"  // ← Группа
  }
}
```

**В холодильнике:**
```json
[
  {
    "ingredientId": "ing_002",
    "ingredient": {
      "id": "ing_002",
      "name": "подсолнечное масло",
      "canonicalId": "vegetable_oil"  // ← Та же группа!
    },
    "quantity": 500
  }
]
```

**Результат:**
✅ `canCook = true` потому что `vegetable_oil` есть в холодильнике (хотя это другой конкретный ингредиент).

### Canonical Groups

| Canonical ID | Variants |
|--------------|----------|
| `vegetable_oil` | растительное масло, подсолнечное масло, оливковое масло |
| `salt` | соль, морская соль, йодированная соль |
| `sugar` | сахар, сахарный песок, тростниковый сахар |
| `eggs` | яйца куриные, яйца перепелиные |
| `milk` | молоко, молоко 3.2%, обезжиренное молоко |

---

## 🔍 SQL Queries (Logs)

**После деплоя должны появиться запросы:**

```sql
-- 1. Load recipe ingredients
SELECT * FROM "RecipeIngredient" 
WHERE "recipeId" = '605c8419-2d42-4ef0-a9d2-839582e98727';

-- 2. Preload Ingredient details
SELECT * FROM "Ingredient" 
WHERE "id" IN ('ing_001', 'ing_002', ...);

-- 3. Load user's fridge
SELECT * FROM "FridgeItem" 
WHERE user_id = '407582be-59d5-4d21-873b-1a72d31b0d42' 
  AND quantity > 0;

-- 4. Preload fridge ingredients
SELECT * FROM "Ingredient" 
WHERE "id" IN ('ing_fridge_001', 'ing_fridge_002', ...);
```

---

## 🧪 Test Scenarios

### Scenario 1: All Ingredients Available ✅

**Request:**
```bash
curl -X POST "/api/menu/today" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "recipe_id": "605c8419-...",  # Recipe requires: eggs, oil, salt
    "servings": 2
  }'
```

**Fridge Contains:**
- ✅ яйца куриные (canonical: eggs) - 6 pcs
- ✅ подсолнечное масло (canonical: vegetable_oil) - 500 ml
- ✅ соль (canonical: salt) - 1000 g

**Response:**
```json
{
  "id": "menu_item_uuid",
  "status": "planned",
  "recipe": { ... },
  "servings": 2
}
```
**Status:** `201 Created` ✅

---

### Scenario 2: Missing Ingredients ❌

**Request:**
```bash
curl -X POST "/api/menu/today" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "recipe_id": "recipe_pierogi_...",  # Requires: flour, eggs, butter, mushrooms
    "servings": 4
  }'
```

**Fridge Contains:**
- ✅ мука (flour) - 1000 g
- ✅ яйца (eggs) - 3 pcs
- ❌ сливочное масло (butter) - **NOT IN FRIDGE**
- ❌ грибы (mushrooms) - **NOT IN FRIDGE**

**Response:**
```json
{
  "error": "insufficient_ingredients",
  "message": "Cannot add to menu: missing ingredients in fridge",
  "missing_ingredients": [
    "сливочное масло",
    "грибы"
  ]
}
```
**Status:** `409 Conflict` ❌

---

### Scenario 3: Canonical Matching Works ✅

**Request:**
```bash
curl -X POST "/api/menu/today" \
  -d '{
    "recipe_id": "recipe_salad_...",  # Requires: "растительное масло" (vegetable_oil)
    "servings": 2
  }'
```

**Fridge Contains:**
- ✅ оливковое масло (canonical: vegetable_oil) - 250 ml  ← **Different specific ingredient!**

**Response:**
```json
{
  "id": "menu_item_uuid",
  "status": "planned",
  ...
}
```
**Status:** `201 Created` ✅  
**Why:** Because `оливковое масло` has same `canonical_id = "vegetable_oil"` as `растительное масло` in recipe.

---

## 📝 Frontend Integration

### Display Missing Ingredients

```typescript
// app/recipes/[id]/page.tsx

async function handleAddToMenu(recipeId: string, servings: number) {
  try {
    const response = await fetch('/api/menu/today', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ recipe_id: recipeId, servings })
    });
    
    if (response.status === 409) {
      const data = await response.json();
      
      // Show missing ingredients to user
      toast.error(
        `Не хватает ингредиентов:\n${data.missing_ingredients.join(', ')}`,
        { duration: 5000 }
      );
      
      // Suggest adding to shopping list
      showShoppingListModal(data.missing_ingredients);
      return;
    }
    
    if (response.ok) {
      toast.success('Рецепт добавлен в меню!');
      router.push('/menu/today');
    }
  } catch (error) {
    toast.error('Ошибка при добавлении в меню');
  }
}
```

### Shopping List Integration

```typescript
function showShoppingListModal(missingIngredients: string[]) {
  return (
    <Dialog>
      <DialogTitle>Не хватает ингредиентов</DialogTitle>
      <DialogContent>
        <p>Для этого рецепта нужны:</p>
        <ul>
          {missingIngredients.map(ing => (
            <li key={ing}>{ing}</li>
          ))}
        </ul>
      </DialogContent>
      <DialogActions>
        <Button onClick={() => addToShoppingList(missingIngredients)}>
          Добавить в список покупок
        </Button>
        <Button onClick={closeModal}>Закрыть</Button>
      </DialogActions>
    </Dialog>
  );
}
```

---

## ✅ Success Criteria

- ✅ Backend проверяет наличие ВСЕХ ингредиентов перед добавлением в меню
- ✅ Использует `canonical_id` для группировки похожих ингредиентов
- ✅ Возвращает 409 Conflict с списком недостающих ингредиентов
- ✅ Возвращает локализованные названия (русские)
- ✅ Frontend может обработать 409 и показать пользователю ЧТО нужно купить
- ✅ Backend = Single Source of Truth (Frontend НЕ решает, можно ли готовить)

---

## 🚀 Next Steps

1. **Test with real data:**
   - Добавить ингредиенты в холодильник
   - Попытаться добавить рецепт с недостающими ингредиентами
   - Проверить логи: должны быть SQL запросы к `RecipeIngredient` и `FridgeItem`

2. **Frontend modal:**
   - Показывать диалог с недостающими ингредиентами
   - Кнопка "Добавить в список покупок"

3. **Analytics:**
   - Логировать попытки добавить рецепт без ингредиентов
   - Метрика: "Какие ингредиенты чаще всего отсутствуют?"

4. **Quantity validation (Phase 2):**
   - Сейчас проверяем только `quantity > 0`
   - Позже: проверять `fridgeQuantity >= recipeQuantity * servings`

---

**Git Commits:**
- `d2a3d83` - "feat: add canCookNow validation"
- `5a785b4` - "fix: correct SQL column name in canCookNow check (recipeId quoted)"

**Status:** ✅ DEPLOYED & READY FOR TESTING
