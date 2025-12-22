# 🧪 Тестирование Fridge Matching

## Проблема
Frontend показывал "0 в холодильнике, 4 не хватает", хотя ингредиенты были в наличии.

## Решение
Backend теперь проверяет холодильник и добавляет поле `inFridge` к каждому ингредиенту.

## Архитектура

### 1️⃣ Публичный доступ (без токена)
```bash
curl https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipes/92691aae-c3af-427d-aaed-1408319f0a3c
```
Ответ: `inFridge: false` для всех ингредиентов

### 2️⃣ С авторизацией (с токеном)
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipes/92691aae-c3af-427d-aaed-1408319f0a3c
```
Ответ: `inFridge: true/false` на основе данных холодильника

## Backend Logic

```go
// EnrichRecipeWithFridgeInfo проверяет каждый ингредиент рецепта
func (s *RecipeMatchService) EnrichRecipeWithFridgeInfo(userID string, recipe *models.RecipeCatalog) error {
    // 1. Загружаем холодильник пользователя
    fridgeItems := []models.UserFridgeItem{}
    s.db.Where("user_id = ?", userID).Find(&fridgeItems)
    
    // 2. Создаем map ingredient_id -> количество в холодильнике
    fridgeMap := make(map[string]float64)
    for _, item := range fridgeItems {
        fridgeMap[item.IngredientID] = item.CurrentQuantity
    }
    
    // 3. Проверяем каждый ингредиент рецепта
    for i := range recipe.Ingredients {
        ing := &recipe.Ingredients[i]
        
        // Если ингредиент есть в холодильнике
        if qty, exists := fridgeMap[ing.IngredientID]; exists && qty > 0 {
            ing.InFridge = true
        } else {
            ing.InFridge = false
        }
    }
    
    return nil
}
```

## Frontend Integration

```typescript
// Загрузка рецепта с проверкой холодильника
const token = localStorage.getItem('authToken');
const response = await fetch(`/api/recipes/${id}`, {
  headers: token ? {
    'Authorization': `Bearer ${token}`
  } : {}
});

const recipe = await response.json();

// Подсчет доступных/недостающих ингредиентов
const stats = recipe.data.ingredients.reduce((acc, ing) => {
  if (ing.inFridge) {
    acc.available++;
  } else {
    acc.missing++;
  }
  return acc;
}, { available: 0, missing: 0 });

console.log(`✅ ${stats.available} в холодильнике`);
console.log(`❌ ${stats.missing} не хватает`);
```

## Commits
- `49ed465` - Add EnrichRecipeWithFridgeInfo method + InFridge field
- `6ada85b` - Make recipe details public with optional fridge matching

## Status
✅ **DEPLOYED** - https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app
