# Система многоязычных рецептов (PL/EN/RU)

**Дата создания:** 19 января 2026 г.  
**Статус:** ✅ Реализовано  

## Обзор

Система автоматически создает рецепты на трех языках: польском (PL), английском (EN) и русском (RU). При создании рецепта администратором или через AI, все тексты автоматически переводятся на три языка.

## Структура базы данных

### Таблица `Recipe` - Поля для переводов

```sql
-- Названия рецепта на 3 языках
name_pl VARCHAR(255)        -- Польское название
name_en VARCHAR(255)        -- Английское название  
name_ru VARCHAR(255)        -- Русское название

-- Описания рецепта на 3 языках
description_pl TEXT         -- Польское описание
description_en TEXT         -- Английское описание
description_ru TEXT         -- Русское описание

-- Шаги приготовления на 3 языках (JSON массивы)
steps_pl JSONB             -- Шаги на польском
steps_en JSONB             -- Шаги на английском
steps_ru JSONB             -- Шаги на русском
```

## Go модели

### Обновленная модель `Recipe`

```go
type Recipe struct {
    // ... базовые поля ...
    
    // Multi-language support (PL/EN/RU)
    NamePL        *string        `json:"namePl,omitempty" gorm:"column:name_pl;type:varchar(255)"`
    NameEN        *string        `json:"nameEn,omitempty" gorm:"column:name_en;type:varchar(255)"`
    NameRU        *string        `json:"nameRu,omitempty" gorm:"column:name_ru;type:varchar(255)"`
    DescriptionPL *string        `json:"descriptionPl,omitempty" gorm:"column:description_pl;type:text"`
    DescriptionEN *string        `json:"descriptionEn,omitempty" gorm:"column:description_en;type:text"`
    DescriptionRU *string        `json:"descriptionRu,omitempty" gorm:"column:description_ru;type:text"`
    StepsPL       datatypes.JSON `json:"stepsPl,omitempty" gorm:"column:steps_pl;type:jsonb"`
    StepsEN       datatypes.JSON `json:"stepsEn,omitempty" gorm:"column:steps_en;type:jsonb"`
    StepsRU       datatypes.JSON `json:"stepsRu,omitempty" gorm:"column:steps_ru;type:jsonb"`
}
```

## API для создания рецептов

### 1. Создание draft рецепта (Admin)

**Endpoint:** `POST /api/admin/recipes`

**Request Body:**
```json
{
  "localName": "Яичница с беконом",
  "category": "main",
  "difficulty": "easy",
  "description": "Классическое сочетание яиц и бекона",
  "timeMinutes": 15,
  "servings": 2,
  
  // ОПЦИОНАЛЬНО - если не указано, будет автоперевод при публикации
  "namePl": "Jajecznica z boczkiem",
  "nameEn": "Scrambled eggs with bacon",
  "nameRu": "Яичница с беконом",
  "descriptionPl": "Klasyczne połączenie jajek i boczku",
  "descriptionEn": "Classic combination of eggs and bacon",
  "descriptionRu": "Классическое сочетание яиц и бекона",
  "stepsPl": ["Podgrzej patelnię", "Usmaż boczek", "Dodaj jajka"],
  "stepsEn": ["Heat the pan", "Fry the bacon", "Add eggs"],
  "stepsRu": ["Разогреть сковороду", "Обжарить бекон", "Добавить яйца"]
}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "id": "uuid",
    "localName": "Яичница с беконом",
    "status": "draft",
    "category": "main",
    "difficulty": "easy",
    "authorId": "user-id",
    "createdAt": "2026-01-19T20:00:00Z"
  }
}
```

### 2. Публикация рецепта с автопереводом

**Endpoint:** `POST /api/admin/recipes/{id}/publish`

**Логика:**
1. Проверяется наличие переводов (`name_pl`, `name_en`, `name_ru`)
2. Если переводы отсутствуют → **автоматический перевод через AI**
3. Рецепт публикуется со статусом `published`

**Request Body:**
```json
{
  "ingredients": [
    {"ingredientId": "uuid", "quantity": 100, "unit": "g"}
  ],
  "steps": [
    {"order": 1, "description": "Разогреть сковороду"},
    {"order": 2, "description": "Обжарить бекон"}
  ],
  "force": false
}
```

**Response с автопереводом:**
```json
{
  "status": "success",
  "data": {
    "id": "uuid",
    "status": "published",
    "namePl": "Jajecznica z boczkiem",
    "nameEn": "Scrambled eggs with bacon",
    "nameRu": "Яичница с беконом"
  },
  "warnings": [
    "Recipe auto-translated to PL/EN/RU"
  ]
}
```

## AI Translation Service

### Новый сервис: `recipe_translator.go`

**Функции:**

#### 1. `TranslateRecipe()` - Полный перевод рецепта
```go
func (s *aiService) TranslateRecipe(
    name, description string, 
    steps []string
) (*RecipeTranslation, error)
```

**Возвращает:**
```go
type RecipeTranslation struct {
    NamePL        string   `json:"name_pl"`
    NameEN        string   `json:"name_en"`
    NameRU        string   `json:"name_ru"`
    DescriptionPL string   `json:"description_pl"`
    DescriptionEN string   `json:"description_en"`
    DescriptionRU string   `json:"description_ru"`
    StepsPL       []string `json:"steps_pl"`
    StepsEN       []string `json:"steps_en"`
    StepsRU       []string `json:"steps_ru"`
}
```

#### 2. `TranslateRecipeField()` - Перевод отдельного поля
```go
func (s *aiService) TranslateRecipeField(
    fieldType, text, sourceLang string
) (pl, en, ru string, err error)
```

**Примеры использования:**
```go
// Перевод названия
pl, en, ru, err := aiService.TranslateRecipeField(
    "recipe name", 
    "Яичница с беконом", 
    "ru"
)

// Перевод описания
pl, en, ru, err := aiService.TranslateRecipeField(
    "recipe description", 
    "Классический завтрак", 
    "ru"
)
```

## Автоматический перевод при публикации

### Функция `ensureTranslations()`

```go
func (s *RecipeAdminService) ensureTranslations(recipe *models.Recipe) []string
```

**Логика работы:**

1. **Проверка наличия переводов**
   ```go
   needsTranslation := recipe.NamePL == nil || *recipe.NamePL == "" ||
       recipe.NameEN == nil || *recipe.NameEN == "" ||
       recipe.NameRU == nil || *recipe.NameRU == ""
   ```

2. **Автоматический перевод через AI**
   - Если переводы отсутствуют → вызов AI Translation Service
   - Перевод названия рецепта
   - Перевод описания (если есть)
   - Перевод шагов (если есть)

3. **Сохранение переводов в БД**
   ```go
   s.db.Model(recipe).Updates(map[string]interface{}{
       "name_pl": plName,
       "name_en": enName,
       "name_ru": ruName,
   })
   ```

4. **Возврат предупреждений**
   ```go
   warnings = []string{"Recipe auto-translated to PL/EN/RU"}
   ```

## Создание рецепта из холодильника (AI)

### Endpoint: `POST /api/ai/fridge/create-recipe`

**Request:**
```json
{
  "language": "pl"  // или "en", "ru"
}
```

**Response:**
```json
{
  "success": true,
  "recipe": {
    "name": "Jajecznica z boczkiem",
    "description": "Pyszne śniadanie z jajek i boczku",
    "ingredientsUsed": [...],
    "steps": [
      "Podgrzej patelnię na średnim ogniu",
      "Usmaż boczek do chrupkości",
      "Dodaj jajka i smaż 3-4 minuty"
    ],
    "cookingTime": 15,
    "economy": {
      "usedFromFridge": true,
      "usedValue": 12.50,
      "savedMoney": 8.30,
      "currency": "PLN"
    }
  }
}
```

⚠️ **ВАЖНО:** Рецепт создается на языке, указанном в `language`, но **НЕ сохраняется в базу автоматически**.

### Рекомендуемый flow для сохранения AI-рецепта:

1. **Фронтенд получает рецепт от AI** (на выбранном языке)
2. **Пользователь нажимает "Сохранить рецепт"**
3. **Фронтенд отправляет:** `POST /api/admin/recipes`
   ```json
   {
     "localName": "{recipe.name из AI}",
     "description": "{recipe.description из AI}",
     "category": "main",
     "difficulty": "easy",
     "timeMinutes": "{recipe.cookingTime}",
     "namePl": "{если язык был pl}",
     "nameEn": "{если язык был en}",
     "nameRu": "{если язык был ru}",
     "stepsPl": "{если язык был pl}",
     "stepsEn": "{если язык был en}",
     "stepsRu": "{если язык был ru}"
   }
   ```
4. **Backend автоматически переведет** недостающие языки при публикации

## Проверка переводов в БД

### SQL запрос для проверки:
```sql
SELECT 
    id,
    title,
    name_pl,
    name_en,
    name_ru,
    status,
    CASE 
        WHEN name_pl IS NULL OR name_pl = '' THEN 'missing PL'
        WHEN name_en IS NULL OR name_en = '' THEN 'missing EN'
        WHEN name_ru IS NULL OR name_ru = '' THEN 'missing RU'
        ELSE 'OK'
    END as translation_status
FROM "Recipe"
WHERE status = 'published'
ORDER BY "createdAt" DESC;
```

## Примеры API запросов

### 1. Создать рецепт с полными переводами
```bash
curl -X POST http://localhost:8080/api/admin/recipes \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "localName": "Pasta Carbonara",
    "category": "main",
    "difficulty": "medium",
    "namePl": "Makaron Carbonara",
    "nameEn": "Pasta Carbonara",
    "nameRu": "Паста Карбонара",
    "descriptionPl": "Klasyczny włoski makaron z boczkiem",
    "descriptionEn": "Classic Italian pasta with bacon",
    "descriptionRu": "Классическая итальянская паста с беконом"
  }'
```

### 2. Создать рецепт БЕЗ переводов (автоперевод при публикации)
```bash
curl -X POST http://localhost:8080/api/admin/recipes \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "localName": "Борщ украинский",
    "category": "soup",
    "difficulty": "medium",
    "description": "Традиционный украинский суп со свеклой"
  }'
```

### 3. Опубликовать рецепт (с автопереводом)
```bash
curl -X POST http://localhost:8080/api/admin/recipes/{id}/publish \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "ingredients": [
      {"ingredientId": "uuid-1", "quantity": 500, "unit": "g"}
    ],
    "steps": [
      {"order": 1, "description": "Нарезать овощи"}
    ]
  }'
```

## Статус реализации

✅ **Выполнено:**
1. Добавлены поля переводов в модель `Recipe`
2. Обновлен DTO `CreateRecipeRequest` для приема переводов
3. Создан AI Translation Service (`recipe_translator.go`)
4. Реализована функция `ensureTranslations()` для автоперевода
5. Интеграция автоперевода в процесс публикации рецепта

⚠️ **Требуется доработка:**
1. Интеграция AI Service в RecipeAdminService (внедрение зависимости)
2. Перевод шагов приготовления (steps) через AI
3. Фронтенд: UI для отображения рецептов на выбранном языке
4. Фронтенд: Кнопка "Сохранить рецепт" для AI-сгенерированных рецептов

## Рекомендации

### Для Backend разработчиков:
1. При создании новых рецептов **всегда заполнять** хотя бы одно из полей: `namePl`, `nameEn`, `nameRu`
2. При публикации система **автоматически дополнит** недостающие переводы
3. Использовать AI Translation Service для качественного перевода

### Для Frontend разработчиков:
1. При получении рецепта, проверять наличие переводов для текущего языка
2. Если перевод отсутствует → fallback на `title` или `localName`
3. При создании рецепта предоставить UI для ручного ввода переводов (опционально)

### Для QA:
1. Проверить, что все опубликованные рецепты имеют переводы на PL/EN/RU
2. Проверить quality переводов (AI может допускать ошибки)
3. Проверить, что draft рецепты могут быть без переводов

---

**Следующие шаги:**
1. Тестирование AI Translation Service
2. Интеграция с фронтендом
3. Добавление кеширования переводов
4. Мониторинг качества AI переводов

