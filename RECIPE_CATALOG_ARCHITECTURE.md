# 📚 АРХИТЕКТУРА КАТАЛОГА РЕЦЕПТОВ (PRODUCTION-READY)

## 🎯 ЦЕЛЬ
Создать масштабируемую систему каталога рецептов с правильной нормализацией данных для AI-рекомендаций.

---

## 1️⃣ МОДЕЛЬ RECIPE (Обязательная структура)

```typescript
Recipe {
  // === IDENTITY ===
  id: UUID                     // Первичный ключ
  canonicalName: string        // ❗ КРИТИЧНО: английский slug (уникальный)
  slug: string                 // URL-friendly версия (fried-salmon)
  
  // === CLASSIFICATION ===
  category: enum               // main | salad | soup | dessert | sauce | drink
  difficulty: enum             // easy | medium | hard
  timeMinutes: number          // Время приготовления
  servings: number             // Количество порций
  
  // === LOCALIZED CONTENT ===
  titles: {
    pl: string                 // "Smażony łosoś"
    ru: string                 // "Жареный лосось"
    en: string                 // "Fried Salmon"
  }
  
  description?: {
    pl?: string
    ru?: string
    en?: string
  }
  
  // === RELATIONS ===
  ingredients: RecipeIngredient[]  // Связь многие-ко-многим
  steps: Step[]                    // Шаги приготовления
  
  // === MEDIA ===
  image: {
    url: string                // S3 / Cloudflare R2 URL
    thumb: string              // Thumbnail
    dominantColor: string      // #FF5733 для UI
  }
  
  // === METADATA ===
  tags: string[]               // protein, keto, vegan, quick, spicy
  country: string              // Poland, Italy, Russia
  cuisine: string              // polish, italian, russian
  
  // === OWNERSHIP ===
  createdBy: userId
  isPublished: boolean         // draft | published
  status: enum                 // draft | published | archived
  
  // === TIMESTAMPS ===
  createdAt: DateTime
  updatedAt: DateTime
}
```

---

## 2️⃣ CANONICAL NAME (Ключ всей системы)

### ❗ ПРАВИЛА:
1. **ВСЕГДА на английском** (никогда не локализуется)
2. **Уникальный** (constraint в БД)
3. **Lowercase snake_case** (fried_salmon, не Fried-Salmon)
4. **Стабильный** (никогда не меняется после создания)

### ✅ ПРАВИЛЬНО:
```
scrambled_eggs
fried_salmon
pasta_carbonara
pierogi_ruskie
greek_salad
```

### ❌ НЕПРАВИЛЬНО:
```
яичница              // локализовано
Scrambled Eggs       // uppercase + spaces
жареный_лосось       // кириллица
fried-salmon         // kebab-case
```

---

## 3️⃣ МАППИНГ СУЩЕСТВУЮЩИХ РЕЦЕПТОВ

| Текущий canonicalName | → | Правильный canonical |
|---|---|---|
| `яичница` | → | `scrambled_eggs` |
| `жареный_лосось` | → | `fried_salmon` |
| `лосось_жареный` | → | `fried_salmon` |
| `Pierogi Ruskie` | → | `pierogi_ruskie` |
| `паста_карбонара_(авторский_рецепт)` | → | `pasta_carbonara` |
| `Greek Salad` | → | `greek_salad` |
| `Spaghetti Carbonara` | → | `pasta_carbonara` |
| `Polish Meat Dumplings` | → | `polish_meat_dumplings` |

---

## 4️⃣ ДУБЛИКАТЫ (Текущие проблемы)

### 🚨 Найденные дубликаты:

**Жареный лосось (6 вариантов):**
- `жареный_лосось`
- `лосось_жареный`
- `жареный_лосось_(микроскопический_тест)`
- `жареный_лосось_(реалистичный_тест)`
- `жареный_лосось_с_хрустящей_кожей`
- `домашний_рецепт_жареного_лосося`

**Решение:** Объединить в 1 рецепт: `fried_salmon`

**Pierogi ruskie (4 варианта):**
- `Pierogi Ruskie`
- 3× без canonical name, но `localName = "Pierogi ruskie"`

**Решение:** Объединить в 1 рецепт: `pierogi_ruskie`

**Яичница (2 варианта):**
- `яичница`
- `Scrambled Eggs`

**Решение:** Объединить в 1 рецепт: `scrambled_eggs`

---

## 5️⃣ INGREDIENTS (Правильная структура)

### Модель Ingredient:
```typescript
Ingredient {
  id: UUID
  canonicalName: string        // salmon_fillet (английский!)
  category: enum               // protein | vegetable | dairy | grain | spice
  
  names: {
    en: string                 // "Salmon Fillet"
    pl: string                 // "Filet z łososia"
    ru: string                 // "Филе лосося"
  }
  
  aliases: string[]            // ["лосось", "łosoś", "salmon"]
  
  // Nutrition (optional)
  calories?: number            // per 100g
  protein?: number
  fats?: number
  carbs?: number
}
```

### Модель RecipeIngredient (связь многие-ко-многим):
```typescript
RecipeIngredient {
  id: UUID
  recipeId: UUID
  ingredientId: UUID
  
  amount: number               // 200
  unit: string                 // g, ml, шт, ст.л.
  
  required: boolean            // true = обязательный
  optional: boolean            // false = можно без него
  
  notes?: string               // "комнатной температуры"
}
```

---

## 6️⃣ АЛГОРИТМ ПОДБОРА РЕЦЕПТОВ (БЕЗ LLM)

### Formula:
```typescript
matchScore = matchedIngredients / totalRequiredIngredients

где:
- matchedIngredients = количество ингредиентов пользователя в рецепте
- totalRequiredIngredients = COUNT(WHERE required = true)
```

### Фильтры:
```sql
SELECT r.*
FROM "Recipe" r
JOIN "RecipeIngredient" ri ON r.id = ri."recipeId"
WHERE 
  -- Только обязательные ингредиенты
  ri.required = true
  
  -- Пользователь имеет ингредиент
  AND ri."ingredientId" IN (SELECT ingredient_id FROM user_fridge WHERE user_id = ?)
  
  -- Минимальный % совпадения
  AND (matchedCount::float / totalRequiredCount::float) >= 0.6

ORDER BY matchScore DESC
LIMIT 10
```

### Категории результатов:
- **CAN_COOK_NOW** (≥ 70% match)
- **ALMOST_READY** (50-69% match)
- **NEED_MORE** (< 50% match)

---

## 7️⃣ ГДЕ ИСПОЛЬЗОВАТЬ AI (И ТОЛЬКО ТАМ)

### ❌ AI НЕ НУЖЕН ДЛЯ:
- Поиска рецептов (SQL быстрее и дешевле)
- Подсчёта matchScore (backend делает это)
- Фильтрации по аллергенам (простой WHERE)

### ✅ AI НУЖЕН ДЛЯ:
1. **Объяснения выбора:**
   - "Почему этот рецепт подходит?"
   - "Что не хватает и где купить?"
   
2. **Генерации вариаций:**
   - "Замена ингредиента X на Y"
   - "Веганская версия рецепта"
   
3. **Умных подсказок:**
   - "С чем подавать это блюдо?"
   - "Как хранить остатки?"

---

## 8️⃣ API ENDPOINTS (Минимальный набор)

### ADMIN (управление каталогом):
```
POST   /api/admin/recipes                 - Создать рецепт
PUT    /api/admin/recipes/:id             - Обновить рецепт
DELETE /api/admin/recipes/:id             - Удалить рецепт
POST   /api/admin/recipes/:id/image       - Загрузить фото
POST   /api/admin/recipes/:id/publish     - Опубликовать
POST   /api/admin/recipes/:id/archive     - Архивировать
```

### USER (публичные):
```
GET    /api/recipes                       - Список рецептов (фильтры)
GET    /api/recipes/:id                   - Детали рецепта
GET    /api/recipes/stats                 - Статистика каталога
```

### AI-POWERED:
```
GET    /api/ai-recipe/recommendation      - ✅ УЖЕ ГОТОВО!
POST   /api/recipes/suggest               - Подбор по ингредиентам (TODO)
```

---

## 9️⃣ МИГРАЦИЯ (Чеклист)

### ✅ СРОЧНО (Критично):
1. ✅ Исправить `.data.recipes[]` в JQ
2. ⏳ Выполнить SQL миграцию: `NORMALIZE_CANONICAL_NAMES.sql`
3. ⏳ Удалить дубликаты (оставить 1 на canonical)
4. ⏳ Добавить UNIQUE constraint на `canonicalName`
5. ⏳ Добавить NOT NULL constraint на `canonicalName`

### ⏳ ВАЖНО (Архитектура):
6. ⏳ Создать таблицу `Ingredient` (если нет)
7. ⏳ Создать таблицу `RecipeIngredient` (связь многие-ко-многим)
8. ⏳ Добавить поле `required` в `RecipeIngredient`
9. ⏳ Добавить `matchScore` calculation в backend

### 📝 ЖЕЛАТЕЛЬНО (Улучшения):
10. 📝 Добавить `tags[]` для рецептов
11. 📝 Добавить `image` upload endpoint
12. 📝 Создать endpoint `/api/recipes/suggest`
13. 📝 Добавить больше категорий (dessert, breakfast, appetizer)

---

## 🔟 ВАЖНЫЕ ВЫВОДЫ

### ❌ ТЕКУЩИЕ ПРОБЛЕМЫ:
- **22 рецепта**, но реально уникальных ~12-15
- **3 рецепта БЕЗ canonical name** (критично!)
- **Локализованные canonical names** (яичница, жареный_лосось)
- **Дубликаты** (~9 лишних записей)

### ✅ ПОСЛЕ ИСПРАВЛЕНИЙ:
- 🚀 **AI станет дешёвым** (SQL вместо LLM для поиска)
- ⚡ **Рекомендации — быстрыми** (индексы по canonical)
- 📈 **Масштабирование — реальным** (нормализованная БД)
- 🎯 **Точность — 100%** (нет путаницы с дубликатами)

---

## 📚 ССЫЛКИ

- SQL миграция: `NORMALIZE_CANONICAL_NAMES.sql`
- AI модуль: `internal/modules/ai_recipe_recommendation/`
- Текущий каталог: 22 рецепта (см. `/tmp/recipes_catalog.txt`)

---

**Статус:** 🚧 В РАБОТЕ  
**Приоритет:** 🔴 КРИТИЧНО  
**Дата:** 2026-01-18
