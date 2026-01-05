# ШАГ 1: POST /api/admin/recipes - МИНИМАЛЬНЫЙ DRAFT

## 🎯 Цель
**Нажал «Создать рецепт» → рецепт появился в БД со статусом `draft`**

---

## ✅ Что сделано

### 1. Минимальный DTO (CMS-черновик)

**File:** `internal/modules/recipes_admin/dto/create_recipe.go`

```go
type CreateRecipeRequest struct {
    // ОБЯЗАТЕЛЬНЫЕ поля (минимум для draft)
    LocalName     string  `json:"localName" binding:"required"`       // Отображаемое имя
    CanonicalName *string `json:"canonicalName,omitempty"`            // Slug (optional)
    Category      string  `json:"category" binding:"required"`        // main, soup, dessert
    Difficulty    string  `json:"difficulty" binding:"required"`      // easy, medium, hard
    
    // ОПЦИОНАЛЬНЫЕ поля (можно добавить позже через PATCH)
    Description   string  `json:"description,omitempty"`
    ImageUrl      string  `json:"imageUrl,omitempty"`
    Country       string  `json:"country,omitempty"`                  // Default: PL
    TimeMinutes   int     `json:"timeMinutes,omitempty"`              // Default: 30
    Servings      int     `json:"servings,omitempty"`                 // Default: 1
    // ... nutrition (all optional)
}
```

### 2. Backend автоматически устанавливает:

```go
Source:   {"type":"manual"}  // Backend controlled
Status:   "draft"            // Backend controlled (КРИТИЧНО)
AuthorID: from JWT token     // Backend controlled
Country:  "PL"              // Default if not provided
TimeMinutes: 30             // Default if not provided
Servings: 1                 // Default if not provided
```

### 3. НЕТ валидации:
- ❌ Ingredients
- ❌ Steps
- ❌ Translations
- ❌ Nutrition

---

## 📋 API Contract (ЗАФИКСИРОВАН)

### Request

**Endpoint:** `POST /api/admin/recipes`  
**Auth:** Required (JWT Bearer token)

**Минимальный payload (канонический):**
```json
{
  "localName": "Pierogi ruskie",
  "canonicalName": "pierogi-ruskie",
  "description": "Черновик рецепта",
  "category": "main",
  "difficulty": "easy"
}
```

**Опциональные поля:**
```json
{
  "localName": "Pierogi ruskie",
  "canonicalName": "pierogi-ruskie",
  "description": "Черновик рецепта",
  "category": "main",
  "difficulty": "easy",
  
  "imageUrl": "https://example.com/image.jpg",
  "country": "PL",
  "region": "Małopolska",
  "timeMinutes": 45,
  "servings": 4,
  "portionWeight": 300,
  
  "grossWeight": 500,
  "netWeight": 450,
  "calories": 300,
  "protein": 12.5,
  "fats": 8.2,
  "carbs": 45.0
}
```

### Response

**Success:** `201 Created`

```json
{
  "status": "success",
  "data": {
    "id": "uuid-here",
    "localName": "Pierogi ruskie",
    "canonicalName": "pierogi-ruskie",
    "status": "draft",
    "category": "main",
    "difficulty": "easy",
    "authorId": "admin-uuid",
    "createdAt": "2026-01-05T12:00:00Z"
  }
}
```

**Error:** `400 Bad Request`

```json
{
  "error": "Invalid input"
}
```

**Error:** `401 Unauthorized`

```json
{
  "error": "User not authenticated"
}
```

---

## 🧪 Тестирование

### Тест-скрипт

```bash
./TEST_STEP1_MINIMAL_DRAFT.sh
```

### Ожидаемый результат:

```
=== 🔐 Login as admin ===
✅ Token received: eyJhbGciOiJIUzI1NiIsInR5cCI6...

=== 📝 Create MINIMAL draft (канонический payload) ===
Response:
{
  "status": "success",
  "data": {
    "id": "3e8f4a2c-...",
    "localName": "Pierogi ruskie",
    "canonicalName": "pierogi-ruskie",
    "status": "draft",
    "category": "main",
    "difficulty": "easy",
    "authorId": "7ec8aba4-...",
    "createdAt": "2026-01-05T12:30:00Z"
  }
}

✅ SUCCESS! Draft created with ID: 3e8f4a2c-...
✅ Status: draft
✅ LocalName: Pierogi ruskie
✅ Category: main
✅ Difficulty: easy
```

### Проверка в БД:

```sql
SELECT 
    id, 
    "localName", 
    status, 
    source, 
    category, 
    difficulty, 
    country,
    "timeMinutes",
    servings,
    "authorId",
    "createdAt"
FROM "Recipe" 
WHERE status = 'draft'
ORDER BY "createdAt" DESC
LIMIT 5;
```

**Ожидаемый результат:**
```
id                                   | localName       | status | source              | category | difficulty | country | timeMinutes | servings | authorId
-------------------------------------|-----------------|--------|---------------------|----------|------------|---------|-------------|----------|----------
3e8f4a2c-...                         | Pierogi ruskie  | draft  | {"type":"manual"}   | main     | easy       | PL      | 30          | 1        | 7ec8aba4-...
```

---

## ✅ Критерии успеха ШАГ 1

1. ✅ **Минимальный payload работает:**
   - Только `localName`, `category`, `difficulty` required
   - Все остальное опционально

2. ✅ **Backend контролирует:**
   - `source = {"type":"manual"}`
   - `status = "draft"`
   - `authorId` из JWT

3. ✅ **Defaults работают:**
   - `country = "PL"` если не указано
   - `timeMinutes = 30` если не указано
   - `servings = 1` если не указано

4. ✅ **NO validation:**
   - Никакой проверки ingredients
   - Никакой проверки steps
   - Никакой проверки translations

5. ✅ **Response содержит:**
   - `id` (uuid)
   - `localName`
   - `canonicalName`
   - `status = "draft"`
   - `category`
   - `difficulty`
   - `authorId`
   - `createdAt`

---

## 🔥 Результат

**ШАГ 1 ЗАВЕРШЁН:**
- ✅ Один `curl` → один новый рецепт в БД
- ✅ Без 500 ошибок
- ✅ Контракт API зафиксирован
- ✅ Frontend и backend больше не спорят

**Следующий шаг:**
- ⏩ ШАГ 2: Зафиксировать контракт в документации
- ⏩ ШАГ 3: `PATCH /api/admin/recipes/{id}` для редактирования
- ⏩ ШАГ 4: `POST /api/admin/recipes/{id}/publish` с полной валидацией

---

## 📂 Измененные файлы

```
internal/modules/recipes_admin/
├── dto/
│   └── create_recipe.go          ✅ Упрощен до минимума
├── service/
│   └── recipe_admin_service.go   ✅ Добавлены defaults
└── transport/http/
    └── handlers.go               ✅ Обновлен response

migrations/
└── 067_add_recipe_status.sql     ✅ Status field

TEST_STEP1_MINIMAL_DRAFT.sh       ✅ Тест-скрипт
```

---

## 🎯 Архитектурные принципы (подтверждены)

1. ✅ **CMS-черновик:** Минимум полей для создания
2. ✅ **Backend control:** Source, status, authorId контролируются backend
3. ✅ **Progressive validation:** Lenient (draft) → Strict (publish)
4. ✅ **No breaking changes:** Существующий recipes module не тронут
5. ✅ **Industry standard:** Как Strapi/Sanity/WordPress

**Status:** 🎉 ШАГ 1 завершён, готов к продакшену
