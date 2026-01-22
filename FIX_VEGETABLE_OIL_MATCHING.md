# 🔥 ПРОБЛЕМА: Растительное масло не распознается в холодильнике

## 🎯 Симптомы
- У пользователя в холодильнике: **Olej roślinny** (1000 ml)
- Рецепт требует: **Olej rzepakowy** (30 ml) 
- Система показывает: "Нужно купить" ❌
- Правильно должно быть: "Из холодильника" ✅

## 🔍 Root Cause Analysis

### 1. Проверка базы данных
```sql
SELECT id, name, name_pl, name_en, name_ru, unit 
FROM "Ingredient" 
WHERE id IN (
    '1b7cea8e-b026-4329-9d2e-c94952e3fa6c',  -- В холодильнике
    '9ff773d2-a3ee-4f4b-bc45-4cfe0d7f680b'   -- В рецепте
);
```

**Результат**:
| id | name | name_pl | name_en | name_ru | unit |
|----|------|---------|---------|---------|------|
| 1b7cea8e... | Olej roślinny | Olej roślinny | Vegetable oil | Растительное масло | ml |
| 9ff773d2... | Olej rzepakowy | Olej rzepakowy | Vegetable oil | Растительное масло | ml |

### 2. Проблема
**ДВА РАЗНЫХ ingredient_id** для одного и того же типа продукта!
- Оба переводятся как "Vegetable oil" (EN)
- Оба переводятся как "Растительное масло" (RU)
- Но это **разные записи** в таблице Ingredient

### 3. Текущая логика matching
```go
// internal/modules/ai_recipe_recommendation/service/recommendation_service.go
func (s *RecommendationService) buildRecipeDTO(
    recipe RecipeCatalog,
    fridgeIngredientIDs map[string]bool,  // ❌ Прямое сравнение ID
    language string,
) RecipeDTO {
    for _, ri := range recipe.Ingredients {
        if fridgeIngredientIDs[ri.IngredientID] {  // ❌ Точное совпадение ID
            // available
        } else {
            // missing
        }
    }
}
```

**Проблема**: Сравнивается `ri.IngredientID == fridge.ingredient_id`
- Рецепт: `9ff773d2-a3ee-4f4b-bc45-4cfe0d7f680b`
- Холодильник: `1b7cea8e-b026-4329-9d2e-c94952e3fa6c`
- Результат: **НЕ СОВПАЛО** ❌

## ✅ Правильное решение

### Вариант 1: Использовать canonical_id (RECOMMENDED)
У тебя уже есть система canonical ingredients из `CANONICAL_INGREDIENTS_GUIDE.md`!

```sql
-- Добавить canonical_id в таблицу Ingredient
ALTER TABLE "Ingredient" ADD COLUMN canonical_id VARCHAR(255);

-- Создать canonical группу для растительных масел
UPDATE "Ingredient" 
SET canonical_id = 'vegetable_oil'
WHERE id IN (
    '1b7cea8e-b026-4329-9d2e-c94952e3fa6c',  -- Olej roślinny
    '9ff773d2-a3ee-4f4b-bc45-4cfe0d7f680b',  -- Olej rzepakowy
    -- можно добавить оливковое, подсолнечное и т.д.
);
```

**Логика matching**:
```go
func (s *RecommendationService) getUserFridgeIngredientIDs(
    ctx context.Context,
    userID string,
) (map[string]bool, error) {
    type IngredientWithCanonical struct {
        IngredientID string
        CanonicalID  *string
    }
    
    var items []IngredientWithCanonical
    err := s.db.WithContext(ctx).
        Table("user_fridge_items AS ufi").
        Select("ufi.ingredient_id, i.canonical_id").
        Joins("LEFT JOIN \"Ingredient\" AS i ON i.id = ufi.ingredient_id").
        Where("ufi.user_id = ? AND ufi.quantity > 0", userID).
        Scan(&items).Error
    
    // Создаем SET: и по ingredient_id, и по canonical_id
    fridgeSet := make(map[string]bool)
    for _, item := range items {
        fridgeSet[item.IngredientID] = true  // Direct match
        if item.CanonicalID != nil {
            fridgeSet[*item.CanonicalID] = true  // Canonical match
        }
    }
    return fridgeSet, nil
}

func (s *RecommendationService) buildRecipeDTO(...) RecipeDTO {
    for _, ri := range recipe.Ingredients {
        ingredientID := ri.IngredientID
        canonicalID := ri.Ingredient.CanonicalID  // Preloaded
        
        // Match by direct ID OR canonical ID
        inFridge := fridgeIngredientIDs[ingredientID] || 
                    (canonicalID != nil && fridgeIngredientIDs[*canonicalID])
        
        if inFridge {
            availableIngredients = append(...)
        } else {
            missingIngredients = append(...)
        }
    }
}
```

### Вариант 2: Merge duplicate ingredients (NOT RECOMMENDED)
Объединить дубликаты в один ID - **НЕ РЕКОМЕНДУЕТСЯ**, потому что:
- Может сломать существующие рецепты
- Может сломать холодильники пользователей
- Тяжелая миграция данных

## 📋 Action Plan

### Step 1: Добавить canonical_id в схему ✅
```sql
ALTER TABLE "Ingredient" ADD COLUMN IF NOT EXISTS canonical_id VARCHAR(255);
CREATE INDEX IF NOT EXISTS idx_ingredient_canonical ON "Ingredient"(canonical_id);
```

### Step 2: Обновить модель Go
```go
// internal/models/ingredient.go
type Ingredient struct {
    ID          string  `gorm:"primaryKey;column:id" json:"id"`
    CanonicalID *string `gorm:"column:canonical_id" json:"canonicalId,omitempty"`  // NEW
    // ... остальные поля
}
```

### Step 3: Создать canonical группы для типичных продуктов
```sql
-- Растительные масла
UPDATE "Ingredient" 
SET canonical_id = 'vegetable_oil'
WHERE name_en IN ('Vegetable oil', 'Rapeseed oil', 'Sunflower oil', 'Olive oil');

-- Соль
UPDATE "Ingredient"
SET canonical_id = 'salt'
WHERE name_en IN ('Salt', 'Sea salt', 'Himalayan salt');

-- Яйца
UPDATE "Ingredient"
SET canonical_id = 'eggs'
WHERE name_en IN ('Eggs', 'Chicken eggs', 'Quail eggs');
```

### Step 4: Обновить логику matching в RecommendationService
- Изменить `getUserFridgeIngredientIDs` для загрузки canonical_id
- Изменить `buildRecipeDTO` для проверки canonical_id
- Добавить debug логи для отслеживания canonical matching

### Step 5: Тестирование
```bash
# Тест 1: Рецепт с "Olej rzepakowy", в холодильнике "Olej roślinny"
curl -H "Authorization: Bearer $TOKEN" \
  "https://.../api/recipe-recommendations/zharenye_yaytsa?lang=ru"

# Ожидаем:
# - available_ingredients: [Яйца, Соль, Растительное масло]  ✅
# - missing_ingredients: []
# - match_percent: 100%
```

## 📊 Impact Analysis

### Затронутые модули:
- ✅ `internal/models/ingredient.go` - добавить поле
- ✅ `internal/modules/ai_recipe_recommendation/service/recommendation_service.go` - логика
- ✅ `migrations/` - новая миграция для canonical_id

### Breaking changes:
- ❌ НЕТ (обратно совместимо, canonical_id опционален)

### Performance impact:
- Минимальный (один JOIN в SQL запросе)
- Индекс на canonical_id ускорит поиск

## 🎯 Quick Fix (временное решение)

Если нужно СРОЧНО исправить для конкретного случая:

```sql
-- Заменить ID в рецепте на ID из холодильника
UPDATE "RecipeIngredient"
SET ingredient_id = '1b7cea8e-b026-4329-9d2e-c94952e3fa6c'  -- Olej roślinny
WHERE ingredient_id = '9ff773d2-a3ee-4f4b-bc45-4cfe0d7f680b'  -- Olej rzepakowy
  AND recipe_id = (SELECT id FROM "Recipe" WHERE canonical_name = 'zharenye_yaytsa');
```

**⚠️ НО ЭТО КОСТЫЛЬ!** Правильное решение - canonical_id.

## 📝 Related Documents
- `CANONICAL_INGREDIENTS_GUIDE.md` - система canonical ingredients
- `CANONICAL_INGREDIENTS_SUMMARY.md` - summary
- `INGREDIENT_CLEANUP_SUMMARY.md` - чистка дубликатов
