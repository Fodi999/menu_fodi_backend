# Recipe Catalog - Quick Reference

## 🎯 Цель

**Матчинг холодильника с каталогом реальных рецептов** (вместо AI-генерации каждый раз).

## 📋 Что создано

### 1. Database Schema (Migration 035)
- **Recipe** - каталог рецептов (Pierogi, Carbonara, etc.)
- **RecipeIngredient** - ингредиенты рецепта с количеством
- **Allergen** - аллергены (gluten, lactose, eggs, etc.)
- **DietTag** - диеты (vegetarian, vegan, keto, etc.)

### 2. Real Recipes (Migration 036)
Добавлено 6 настоящих рецептов:
- Pierogi Ruskie (Poland, 90 min, medium)
- Bigos (Poland, 180 min, medium)
- Spaghetti Carbonara (Italy, 25 min, easy)
- Pizza Margherita (Italy, 120 min, medium)
- Scrambled Eggs (Poland, 10 min, easy)
- Greek Salad (Greece, 15 min, easy)

### 3. Matching Algorithm
**Файл**: `internal/modules/recipes/service/match_service.go`

**Алгоритм**:
```
1. Загрузить ингредиенты из холодильника пользователя
2. Загрузить рецепты из каталога (с фильтрами)
3. Для каждого рецепта:
   - Сопоставить ингредиенты
   - Посчитать score = (matched / required) * 100
   - Bonus +2 за каждый expiring item
   - Bonus +5% за optional ingredients
4. Отсортировать по score
5. Вернуть топ-N
```

**Score calculation**:
- 100% = все ингредиенты есть в нужном количестве
- 50% = половина ингредиентов есть
- 0% = ничего нет

**Bonus**:
- +2 points за каждый истекающий продукт (приоритет waste reduction)
- +5% за optional ингредиенты

### 4. API Endpoints
**Файл**: `internal/modules/recipes/transport/http/handler.go`

```
GET /api/recipes/match?country=Poland&maxTime=60&excludeAllergens=gluten&minScore=50

Response:
{
  "success": true,
  "count": 5,
  "matches": [
    {
      "recipe": {
        "id": "uuid",
        "canonicalName": "Pierogi Ruskie",
        "localName": "Pierogi ruskie",
        "country": "Poland",
        "difficulty": "medium",
        "timeMinutes": 90,
        "servings": 4,
        "steps": [...],
        "allergens": [...],
        "dietTags": [...]
      },
      "matchScore": 85.5,
      "matchedIngredients": [
        {
          "name": "Ziemniak",
          "required": 500,
          "available": 1000,
          "unit": "g",
          "isExpiringSoon": false
        }
      ],
      "missingIngredients": [
        {
          "name": "Ser twaróg",
          "required": 250,
          "unit": "g",
          "estimatedCost": 5.50,
          "optional": false
        }
      ],
      "costToComplete": 5.50,
      "hasExpiringItems": false,
      "canMakeNow": false
    }
  ]
}
```

## 🔧 Как применить

### 1. Применить миграции
```bash
# На Neon.tech SQL Editor:
# Скопировать и выполнить:
migrations/035_create_recipe_catalog.sql
migrations/036_seed_real_recipes.sql
```

### 2. Зарегистрировать роуты
**Файл**: `cmd/server/main.go` или module router

```go
import (
    recipeService "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/recipes/service"
    recipeHttp "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/recipes/transport/http"
)

// In main() or module setup:
matchService := recipeService.NewRecipeMatchService(db)
recipeHandler := recipeHttp.NewRecipeHandler(matchService, logger)

// Register routes (protected by auth middleware):
r.Route("/api/recipes", func(r chi.Router) {
    r.Use(authMiddleware) // Требуется JWT
    r.Get("/match", recipeHandler.MatchRecipes)
    r.Get("/{id}", recipeHandler.GetRecipeByID)
    r.Get("/", recipeHandler.ListRecipes)
})
```

## 🎮 Использование

### Frontend: Recipe Matching
```typescript
// pages/dashboard/recipes/match.tsx

const matchRecipes = async () => {
  const response = await fetch('/api/recipes/match?minScore=50&limit=10', {
    headers: { 
      'Authorization': `Bearer ${token}` 
    }
  });
  
  const data = await response.json();
  
  // Show top matches
  data.matches.forEach(match => {
    console.log(`${match.recipe.localName} - ${match.matchScore}%`);
    console.log(`Missing: ${match.missingIngredients.length} items`);
    console.log(`Cost to complete: ${match.costToComplete} PLN`);
  });
};
```

### Filters
```typescript
// Фильтры для поиска
const filters = {
  country: 'Poland',           // Только польские рецепты
  category: 'main',            // Только основные блюда
  difficulty: 'easy',          // Только простые
  maxTime: 30,                 // До 30 минут
  excludeAllergens: 'gluten,lactose', // Исключить аллергены
  dietTags: 'vegetarian',      // Только вегетарианские
  minScore: 70,                // Минимум 70% совпадение
  limit: 5                     // Топ-5
};

const url = `/api/recipes/match?${new URLSearchParams(filters)}`;
```

## 🧠 Когда использовать AI vs Catalog

### Use Catalog:
- ✅ Пользователь нажимает "Stwórz przepis" без специальных требований
- ✅ Есть рецепты с match score > 50%
- ✅ Быстрый результат нужен (50-150ms)

### Use AI:
- ❌ Все рецепты < 50% match (слишком мало ингредиентов)
- ❌ Пользователь явно просит "придумай что-то новое"
- ❌ Специальный запрос ("без сахара, с имбирем")

### Hybrid Approach:
```go
// 1. Try catalog first
matches, _ := matchService.MatchRecipesWithFridge(userID, filters)

if len(matches) > 0 && matches[0].MatchScore > 70 {
    // Return catalog match
    return matches
} else {
    // Fallback to AI generation
    return aiService.CreateRecipe(userID, language)
}
```

## 📊 Мониторинг

**Метрики для отслеживания**:
- Catalog hit rate (% запросов без AI)
- Average match score
- Most popular recipes
- Missing ingredients frequency (для добавления в каталог)

## 🔄 Расширение каталога

### Добавить новый рецепт:
```sql
-- Migration 037: Add Kotlet Schabowy

DO $$
DECLARE
    recipe_id UUID;
BEGIN
    INSERT INTO "Recipe" (
        "canonicalName", "localName", country, category, difficulty,
        "timeMinutes", servings, steps, "nutritionProfile", source
    ) VALUES (
        'Pork Cutlet',
        'Kotlet schabowy',
        'Poland',
        'main',
        'medium',
        30,
        4,
        '[...]'::jsonb,
        '{"type": "high-protein", "calories": 450}'::jsonb,
        '{"type": "traditional"}'::jsonb
    ) RETURNING id INTO recipe_id;

    -- Add ingredients...
END $$;
```

## ✅ TODO

- [ ] Применить migrations 035, 036 на Neon.tech
- [ ] Зарегистрировать роуты в main.go
- [ ] Протестировать GET /api/recipes/match
- [ ] Добавить больше рецептов (target: 50-100)
- [ ] Реализовать GET /api/recipes/:id (детали)
- [ ] Реализовать GET /api/recipes (список с фильтрами)
- [ ] Frontend: страница matching результатов
- [ ] Hybrid mode: catalog → AI fallback
- [ ] Analytics: track hit rate

## 🎯 Expected Impact

**До** (только AI):
- 100% запросов → AI generation
- 3-5 секунд на рецепт
- Непредсказуемый результат
- Нет reuse

**После** (catalog + AI):
- 70-80% запросов → catalog (50-150ms)
- 20-30% запросов → AI fallback
- Предсказуемый результат
- Полный reuse
