# Исправление: Редактирование рецепта - CREATE vs UPDATE

## 🐛 Проблема

При редактировании рецепта пользователь получал ошибку **500 Internal Server Error** с duplicate key constraint:

```
POST /api/admin/recipes/save 500 (Internal Server Error)

Backend log:
ERROR: duplicate key value violates unique constraint "Recipe_canonicalName_unique" (SQLSTATE 23505)
INSERT INTO "Recipe" ("canonicalName", ...) VALUES ('яйца_жареные_на_масле', ...)
```

### Корневая причина

Функция `SaveEditedRecipe` **всегда создавала новый рецепт** (INSERT), даже когда frontend передавал `recipeId` для редактирования существующего:

```go
// ❌ БЫЛО (неправильно)
recipe := &models.RecipeCatalog{
    ID:            uuid.New(),  // ❌ Всегда новый UUID!
    CanonicalName: canonicalName,
    Title:         req.Title,
    // ...
}

// ❌ Всегда INSERT
if err := tx.Create(recipe).Error; err != nil {
    return nil, fmt.Errorf("failed to create recipe: %w", err)
}
```

**Последовательность ошибки:**
1. Проверка дубликатов проходила успешно ✅ (мы исключали текущий рецепт)
2. Но дальше код **игнорировал RecipeID** и создавал новый UUID
3. Попытка INSERT с существующим `canonicalName` → **UNIQUE constraint violation** ❌

---

## ✅ Решение

### Добавлена логика определения режима работы

**Файл:** `internal/modules/admin/service/recipe_ai.go` (строки 590-750)

```go
// ✅ СТАЛО (правильно)
var recipe *models.RecipeCatalog
isEditMode := req.RecipeID != nil && *req.RecipeID != ""

if isEditMode {
    // РЕЖИМ РЕДАКТИРОВАНИЯ
    recipeID, err := uuid.Parse(*req.RecipeID)
    if err != nil {
        return nil, fmt.Errorf("invalid recipe ID: %w", err)
    }

    // Загружаем существующий рецепт
    recipe = &models.RecipeCatalog{}
    if err := tx.First(recipe, "id = ?", recipeID).Error; err != nil {
        return nil, fmt.Errorf("recipe not found: %w", err)
    }

    fmt.Printf("📝 Editing existing recipe: ID=%s\n", recipe.ID)

    // Обновляем поля
    recipe.CanonicalName = canonicalName
    recipe.Title = req.Title
    recipe.Difficulty = req.Difficulty
    // ...

    // Удаляем старые ингредиенты (создадим новые)
    if err := tx.Where("recipe_id = ?", recipe.ID).Delete(&models.CatalogIngredient{}).Error; err != nil {
        return nil, fmt.Errorf("failed to delete old ingredients: %w", err)
    }

} else {
    // РЕЖИМ СОЗДАНИЯ
    recipe = &models.RecipeCatalog{
        ID:            uuid.New(),  // ✅ Новый UUID только для создания
        CanonicalName: canonicalName,
        Title:         req.Title,
        // ...
    }

    fmt.Printf("✨ Creating new recipe: ID=%s\n", recipe.ID)

    // INSERT нового рецепта
    if err := tx.Create(recipe).Error; err != nil {
        return nil, fmt.Errorf("failed to create recipe: %w", err)
    }
}
```

### Разные методы сохранения для разных режимов

```go
// Финальное сохранение
if isEditMode {
    // ✅ Для редактирования: Updates с явным указанием полей
    updates := map[string]interface{}{
        "canonicalName":     recipe.CanonicalName,
        "title":             recipe.Title,
        "difficulty":        recipe.Difficulty,
        "timeMinutes":       recipe.TimeMinutes,
        "servings":          recipe.Servings,
        "country":           recipe.Country,
        "source":            recipe.Source,
        "nutritionProfile":  recipe.NutritionProfile,
    }

    // Локализованные поля
    if req.Language == "ru" {
        updates["description_ru"] = recipe.DescriptionRu
        updates["steps_ru"] = recipe.StepsRu
    }
    // ...

    if err := tx.Model(recipe).Updates(updates).Error; err != nil {
        return nil, fmt.Errorf("failed to update recipe: %w", err)
    }

    fmt.Printf("✅ Recipe updated: ID=%s\n", recipe.ID)

} else {
    // ✅ Для создания: Save
    if err := tx.Save(recipe).Error; err != nil {
        return nil, fmt.Errorf("failed to save recipe: %w", err)
    }

    fmt.Printf("✅ Recipe saved: ID=%s\n", recipe.ID)
}
```

**Ключевые отличия:**
- **EDIT**: `First()` → изменение полей → `Updates()` с явной картой
- **CREATE**: `uuid.New()` → `Create()` → `Save()`
- **EDIT**: Удаляем старые ингредиенты перед созданием новых
- **CREATE**: Просто создаем ингредиенты

---

## 🧪 Тестирование

### Сценарий 1: Редактирование существующего рецепта ✅

**До исправления:**
```bash
# Запрос
POST /api/admin/recipes/save
{
  "recipeId": "4aa22783-45cc-4fc4-8800-4340a5c93ce9",
  "title": "яйца жареные на масле",
  "ingredients": [...измененные...],
  "steps": [...измененные...]
}

# Ответ: ❌ 500 Internal Server Error
Backend log:
ERROR: duplicate key value violates unique constraint "Recipe_canonicalName_unique"
INSERT INTO "Recipe" ("canonicalName", ...) VALUES ('яйца_жареные_на_масле', ...)
```

**После исправления:**
```bash
# Запрос
POST /api/admin/recipes/save
{
  "recipeId": "4aa22783-45cc-4fc4-8800-4340a5c93ce9",  // ✅ Используется для UPDATE
  "title": "яйца жареные на масле",
  "ingredients": [...измененные...],
  "steps": [...измененные...]
}

# Ответ: ✅ 200 OK
Backend log:
📝 Editing existing recipe: ID=4aa22783-45cc-4fc4-8800-4340a5c93ce9
✅ Recipe updated: ID=4aa22783-45cc-4fc4-8800-4340a5c93ce9
{
  "success": true,
  "data": { ...обновленный рецепт... }
}
```

### Сценарий 2: Создание нового рецепта ✅

```bash
# Запрос (без recipeId - создание нового)
POST /api/admin/recipes/save
{
  "title": "новый рецепт блинов",
  "ingredients": [...],
  "steps": [...]
}

# Ответ: ✅ 200 OK
Backend log:
✨ Creating new recipe: ID=<новый-uuid>
✅ Recipe saved: ID=<новый-uuid>
{
  "success": true,
  "data": { ...новый рецепт... }
}
```

### Сценарий 3: Попытка создать дубликат ❌

```bash
# Запрос (без recipeId, но название уже существует)
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

### Сценарий 4: Редактирование с изменением названия ✅

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
Backend log:
📝 Editing existing recipe: ID=4aa22783-45cc-4fc4-8800-4340a5c93ce9
✅ Recipe updated: ID=4aa22783-45cc-4fc4-8800-4340a5c93ce9
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

### Проблема ❌
- Backend **всегда создавал новый рецепт** (INSERT), даже при передаче `recipeId`
- Игнорировал существующий UUID, генерировал новый
- Результат: Duplicate key constraint на `canonicalName`
- Редактирование рецепта **полностью не работало**

### Решение ✅
- Добавлена логика **определения режима**: `isEditMode = recipeId != nil && recipeId != ""`
- **EDIT MODE**: Загружаем существующий рецепт → обновляем поля → `Updates()`
- **CREATE MODE**: Новый UUID → создаем рецепт → `Create()` + `Save()`
- Проверка дубликатов теперь корректно исключает текущий рецепт при редактировании

### До исправления ❌
```
Frontend: "Edit recipe with recipeId=123"
Backend: "Okay, creating NEW recipe with NEW uuid..."
Database: "ERROR: canonicalName already exists!"
```

### После исправления ✅
```
Frontend: "Edit recipe with recipeId=123"
Backend: "Loading recipe 123, updating fields..."
Database: "UPDATE successful!"
```

---

## 🔧 Технические детали

### Два режима работы

| Аспект | CREATE MODE | EDIT MODE |
|--------|-------------|-----------|
| **Триггер** | `recipeId == nil` | `recipeId != nil` |
| **UUID** | `uuid.New()` | Из `req.RecipeID` |
| **Загрузка** | - | `tx.First(recipe, id)` |
| **Рецепт** | `tx.Create(recipe)` | Обновление полей |
| **Ингредиенты** | `tx.Create(ing)` | `tx.Delete()` → `tx.Create()` |
| **Финал** | `tx.Save(recipe)` | `tx.Model().Updates(map)` |
| **Лог** | `✨ Creating new recipe` | `📝 Editing existing recipe` |

### Почему Updates() а не Save()?

```go
// ❌ ПЛОХО: Save() может вызвать проблемы с FK
tx.Save(recipe)

// ✅ ХОРОШО: Updates() с явной картой полей
tx.Model(recipe).Updates(map[string]interface{}{
    "canonicalName": recipe.CanonicalName,
    "title":         recipe.Title,
    // ... только нужные поля
})
```

**Причина:** `Save()` обновляет все поля, включая ассоциации, что может вызвать FK constraint violations.

---

## 🔗 Связанные файлы

- **Service:** `internal/modules/admin/service/recipe_ai.go` (строки 501-555)
- **Handler:** `internal/modules/admin/transport/http/recipe_ai_handlers.go`
- **Model:** `internal/models/recipe_catalog.go`

---

## 📝 Commits

- **Main fix (CREATE/UPDATE):** `ac5d7d4` - "Fix SaveEditedRecipe - properly handle create vs update modes"
- **Duplicate check fix:** `afc8906` - "Fix recipe edit duplicate check - allow same name when editing"
- **Related (image URL):** 
  - `6324d6b` - Add imageUrl to admin RecipeResponse
  - `43a1fa2` - Add imageUrl to RecipeCatalog model

---

**Status:** ✅ Fixed and deployed to production  
**Date:** 2026-01-19  
**Priority:** Critical (recipe editing was completely broken)
