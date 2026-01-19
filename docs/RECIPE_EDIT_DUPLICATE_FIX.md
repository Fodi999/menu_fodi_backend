# Исправление: Редактирование рецепта с тем же названием

## 🐛 Проблема

При редактировании рецепта пользователь получал ошибку **409 Conflict** даже если менял только содержимое (ингредиенты, шаги, фото), оставляя название неизменным:

```
POST /api/admin/recipes/save 409 (Conflict)
{
  "success": false,
  "code": "RECIPE_NAME_EXISTS",
  "message": "Рецепт с таким названием уже существует"
}
```

### Причина

Функция `SaveEditedRecipe` проверяла уникальность `canonicalName` во всей таблице, но **не исключала текущий редактируемый рецепт** из проверки. Результат: рецепт находил сам себя и выдавал ложное срабатывание.

```go
// ❌ БЫЛО (неправильно)
var existing models.RecipeCatalog
if err := s.db.Where("\"canonicalName\" = ?", canonicalName).First(&existing).Error; err == nil {
    return nil, fmt.Errorf("recipe with similar name already exists: %s", canonicalName)
}
// Проблема: находит сам редактируемый рецепт!
```

---

## ✅ Решение

### 1. Добавлено поле `RecipeID` в запрос

**Файл:** `internal/modules/admin/service/recipe_ai.go`

```go
type SaveEditedRecipeRequest struct {
    RecipeID    *string            `json:"recipeId,omitempty"` // ✅ НОВОЕ: UUID рецепта
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
```

- **Тип:** `*string` (указатель) - опциональное поле
- **Назначение:** Если присутствует - это редактирование, если `nil` - создание нового рецепта

### 2. Исправлена логика проверки дубликатов

```go
// ✅ СТАЛО (правильно)
var existing models.RecipeCatalog
query := s.db.Where("\"canonicalName\" = ?", canonicalName)

// Если это редактирование (есть RecipeID), исключаем текущий рецепт
if req.RecipeID != nil && *req.RecipeID != "" {
    query = query.Where("id != ?", *req.RecipeID)
}

if err := query.First(&existing).Error; err == nil {
    return nil, fmt.Errorf("recipe with similar name already exists: %s", canonicalName)
}
```

**Логика:**
1. Ищем рецепт с таким же `canonicalName`
2. **Если `RecipeID` присутствует** → исключаем его из поиска (`id != recipeID`)
3. Если найден **другой** рецепт с таким названием → ошибка
4. Если ничего не найдено или найден только сам редактируемый рецепт → OK ✅

---

## 🧪 Тестирование

### Сценарий 1: Редактирование с тем же названием ✅

**До исправления:**
```bash
# Запрос
POST /api/admin/recipes/save
{
  "title": "яйца жареные на масле",
  "ingredients": [...измененные...],
  "steps": [...измененные...]
}

# Ответ: ❌ 409 Conflict
{
  "success": false,
  "code": "RECIPE_NAME_EXISTS",
  "message": "Рецепт с таким названием уже существует"
}
```

**После исправления:**
```bash
# Запрос
POST /api/admin/recipes/save
{
  "recipeId": "4aa22783-45cc-4fc4-8800-4340a5c93ce9",  // ✅ НОВОЕ
  "title": "яйца жареные на масле",
  "ingredients": [...измененные...],
  "steps": [...измененные...]
}

# Ответ: ✅ 200 OK
{
  "success": true,
  "data": { рецепт обновлён }
}
```

### Сценарий 2: Создание нового рецепта с существующим названием ❌

```bash
# Запрос (без recipeId - создание нового)
POST /api/admin/recipes/save
{
  "title": "яйца жареные на масле",  // название уже существует
  "ingredients": [...],
  "steps": [...]
}

# Ответ: ❌ 409 Conflict (правильно!)
{
  "success": false,
  "code": "RECIPE_NAME_EXISTS",
  "message": "Рецепт с таким названием уже существует"
}
```

### Сценарий 3: Редактирование с изменением названия на уникальное ✅

```bash
# Запрос
POST /api/admin/recipes/save
{
  "recipeId": "4aa22783-45cc-4fc4-8800-4340a5c93ce9",
  "title": "яйца с зеленью на масле",  // ✅ Новое уникальное название
  "ingredients": [...],
  "steps": [...]
}

# Ответ: ✅ 200 OK
```

---

## 📋 Frontend требования

Frontend должен **всегда передавать `recipeId`** при редактировании:

```typescript
// ✅ ПРАВИЛЬНО: Редактирование
const saveRecipe = async (recipeData: RecipeEditData, recipeId: string) => {
  const response = await fetch('/api/admin/recipes/save', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      recipeId: recipeId,  // ✅ ОБЯЗАТЕЛЬНО для редактирования
      title: recipeData.title,
      ingredients: recipeData.ingredients,
      steps: recipeData.steps,
      // ...
    }),
  });
  
  return response.json();
};

// ✅ ПРАВИЛЬНО: Создание нового
const createRecipe = async (recipeData: RecipeCreateData) => {
  const response = await fetch('/api/admin/recipes/save', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      // recipeId отсутствует - создание нового
      title: recipeData.title,
      ingredients: recipeData.ingredients,
      steps: recipeData.steps,
      // ...
    }),
  });
  
  return response.json();
};
```

---

## 🎯 Итоги

### До исправления ❌
- Редактирование рецепта с тем же названием → **409 Conflict**
- Пользователь не мог сохранить изменения без смены названия

### После исправления ✅
- Редактирование рецепта с тем же названием → **200 OK**
- Проверка уникальности работает корректно (исключает текущий рецепт)
- Создание нового рецепта с существующим названием → **409 Conflict** (как и должно быть)

---

## 🔗 Связанные файлы

- **Service:** `internal/modules/admin/service/recipe_ai.go` (строки 501-555)
- **Handler:** `internal/modules/admin/transport/http/recipe_ai_handlers.go`
- **Model:** `internal/models/recipe_catalog.go`

---

## 📝 Commits

- **Main fix:** `afc8906` - "Fix recipe edit duplicate check - allow same name when editing"
- **Related:** 
  - `6324d6b` - Previous image URL fixes
  - `43a1fa2` - Added imageUrl to RecipeCatalog model

---

**Status:** ✅ Fixed and deployed to production  
**Date:** 2026-01-19  
**Priority:** High (blocking recipe editing)
