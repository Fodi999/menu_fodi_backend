# 📊 Данные рецепта: Request → Database → Response

## 🎯 POST /api/admin/recipes - Создание draft

### 📥 Request (Frontend → Backend)

**Минимальный payload (3 обязательных поля):**
```json
{
  "localName": "Pierogi ruskie",         // ✅ REQUIRED
  "category": "main",                     // ✅ REQUIRED
  "difficulty": "easy"                    // ✅ REQUIRED
}
```

**Расширенный payload (опциональные поля):**
```json
{
  "localName": "Pierogi ruskie",         // ✅ REQUIRED
  "canonicalName": "pierogi-ruskie",     // ⚪ optional (slug)
  "category": "main",                     // ✅ REQUIRED
  "difficulty": "easy",                   // ✅ REQUIRED
  
  "description": "Традиційні польські вареники...",
  "imageUrl": "https://example.com/pierogi.jpg",
  "country": "PL",
  "region": "Małopolska",
  "timeMinutes": 45,
  "servings": 4,
  "portionWeight": 300,
  
  "grossWeight": 500,      // Брутто (г)
  "netWeight": 450,        // Нетто (г)
  "calories": 300,         // ккал
  "protein": 12.5,         // Белки (г)
  "fats": 8.2,             // Жиры (г)
  "carbs": 45.0            // Углеводы (г)
}
```

---

## 💾 Database (что сохраняется в Recipe table)

### ✅ Backend автоматически добавляет:

```sql
-- ✅ Backend-controlled fields
source       = '{"type":"manual"}'     -- Backend sets
status       = 'draft'                 -- Backend sets (КРИТИЧНО)
author_id    = '7ec8aba4-...'         -- From JWT token

-- ✅ Default values (если не указано в request)
country      = 'PL'                   -- Default
timeMinutes  = 30                     -- Default
servings     = 1                      -- Default

-- ✅ System fields
id           = '3e8f4a2c-...'         -- UUID auto-generated
createdAt    = '2026-01-05 12:30:00'  -- Auto timestamp
updatedAt    = '2026-01-05 12:30:00'  -- Auto timestamp

-- ✅ Metrics (defaults)
tokens_reward = 10                    -- Default reward
views_count   = 0                     -- Initial value
tokens_earned = 0                     -- Initial value
```

### 📊 Полная запись в БД:

```sql
SELECT 
    id,                    -- '3e8f4a2c-...'
    "localName",           -- 'Pierogi ruskie'
    "canonicalName",       -- 'pierogi-ruskie' OR NULL
    title,                 -- '' (empty for draft)
    description,           -- 'Традиційні...' OR ''
    "imageUrl",            -- 'https://...' OR ''
    
    country,               -- 'PL'
    category,              -- 'main'
    difficulty,            -- 'easy'
    "timeMinutes",         -- 45 OR 30 (default)
    servings,              -- 4 OR 1 (default)
    source,                -- '{"type":"manual"}'
    status,                -- 'draft' ✅
    
    author_id,             -- '7ec8aba4-...'
    
    gross_weight,          -- 500 OR NULL
    net_weight,            -- 450 OR NULL
    calories,              -- 300 OR NULL
    protein,               -- 12.5 OR NULL
    fats,                  -- 8.2 OR NULL
    carbs,                 -- 45.0 OR NULL
    yield,                 -- NULL
    cost,                  -- NULL
    
    tokens_reward,         -- 10
    views_count,           -- 0
    tokens_earned,         -- 0
    
    "createdAt",           -- '2026-01-05 12:30:00'
    "updatedAt"            -- '2026-01-05 12:30:00'
FROM "Recipe"
WHERE id = '3e8f4a2c-...';
```

---

## 📤 Response (Backend → Frontend)

### 🎯 CreateRecipeResponse (МИНИМАЛЬНЫЙ - только критичные поля)

```json
{
  "status": "success",
  "data": {
    "id": "3e8f4a2c-1a2b-3c4d-5e6f-7a8b9c0d1e2f",
    "localName": "Pierogi ruskie",
    "canonicalName": "pierogi-ruskie",
    "status": "draft",                    // ✅ Всегда "draft"
    "category": "main",
    "difficulty": "easy",
    "authorId": "7ec8aba4-9f8e-7d6c-5b4a-3c2d1e0f9a8b",
    "createdAt": "2026-01-05T12:30:00+01:00"
  }
}
```

### 📋 Поля в Response:

| Поле | Тип | Откуда | Описание |
|------|-----|--------|----------|
| `id` | string (UUID) | Database | Auto-generated ID |
| `localName` | string | Request OR default | Отображаемое имя |
| `canonicalName` | string? | Request OR null | URL slug (optional) |
| `status` | string | Backend | Всегда `"draft"` |
| `category` | string | Request (required) | main/soup/dessert/... |
| `difficulty` | string | Request (required) | easy/medium/hard |
| `authorId` | string (UUID) | JWT token | ID автора |
| `createdAt` | string (ISO) | Database | Timestamp создания |

---

## 🔍 Что НЕ возвращается в Response?

### ❌ НЕ включено в CreateRecipeResponse:

```typescript
// ❌ Эти поля ЕСТЬ в БД, но НЕ возвращаются в Response:
title           // '' (empty for draft)
description     // Может быть заполнено
imageUrl        // Может быть заполнено
country         // 'PL' (default)
timeMinutes     // 30 (default)
servings        // 1 (default)
source          // {"type":"manual"}
region          // NULL
portionWeight   // NULL

grossWeight     // Nutrition fields
netWeight
calories
protein
fats
carbs
yield
cost

tokensReward    // Metrics
viewsCount
tokensEarned

updatedAt       // Timestamps
```

### 💡 Почему минимальный Response?

1. **Frontend нужен только ID** для редиректа: `/admin/recipes/{id}/edit`
2. **Status важен** для UI бейджа "Draft"
3. **Остальное подгрузится** через GET `/api/admin/recipes/{id}`

---

## 🔄 Полная схема данных

```
┌─────────────────────────────────────────────────────────────────┐
│                    Frontend Request                              │
├─────────────────────────────────────────────────────────────────┤
│ {                                                                │
│   localName: "Pierogi ruskie"     ✅ REQUIRED                   │
│   category: "main"                ✅ REQUIRED                    │
│   difficulty: "easy"              ✅ REQUIRED                    │
│   canonicalName?: "pierogi-ruskie" ⚪ optional                  │
│   description?: "..."              ⚪ optional                   │
│   ... (nutrition)                  ⚪ optional                   │
│ }                                                                │
└─────────────────────────────────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────────┐
│                  Backend Processing                              │
├─────────────────────────────────────────────────────────────────┤
│ ✅ Validate: localName, category, difficulty                    │
│ ✅ Get authorId from JWT token                                  │
│ ✅ Add backend fields:                                          │
│    - source = {"type":"manual"}                                 │
│    - status = "draft"                                           │
│ ✅ Apply defaults:                                              │
│    - country = "PL" (if not provided)                           │
│    - timeMinutes = 30 (if not provided)                         │
│    - servings = 1 (if not provided)                             │
│ ✅ Generate UUID                                                │
│ ✅ Set timestamps                                               │
└─────────────────────────────────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────────┐
│                   Database (Recipe table)                        │
├─────────────────────────────────────────────────────────────────┤
│ ALL fields saved (35+ columns)                                  │
│ - User input: localName, category, difficulty, etc.             │
│ - Backend controlled: source, status, authorId                  │
│ - Defaults: country, timeMinutes, servings                      │
│ - System: id, createdAt, updatedAt                              │
│ - Metrics: tokensReward, viewsCount, tokensEarned               │
└─────────────────────────────────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────────┐
│              Backend Response (MINIMAL)                          │
├─────────────────────────────────────────────────────────────────┤
│ {                                                                │
│   status: "success",                                             │
│   data: {                                                        │
│     id: "uuid"                    ← для редиректа              │
│     localName: "..."               ← отображение               │
│     canonicalName: "..." | null    ← slug                      │
│     status: "draft"                ← UI badge                  │
│     category: "..."                ← фильтр                    │
│     difficulty: "..."              ← фильтр                    │
│     authorId: "uuid"               ← владелец                  │
│     createdAt: "ISO-8601"          ← timestamp                 │
│   }                                                              │
│ }                                                                │
└─────────────────────────────────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────────┐
│                    Frontend Action                               │
├─────────────────────────────────────────────────────────────────┤
│ ✅ Show success toast: "Draft created!"                         │
│ ✅ Redirect to: /admin/recipes/{id}/edit                        │
│ ✅ Display badge: "Draft" (from status)                         │
│                                                                  │
│ ⏩ Next: GET /api/admin/recipes/{id} для полных данных         │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🎯 Зачем минимальный Response?

### ✅ Преимущества:

1. **Быстрее** - меньше данных передается по сети
2. **Безопаснее** - не раскрываем лишние internal поля
3. **Проще** - frontend получает только то, что нужно для UI
4. **REST-принцип** - POST возвращает ID + location header

### 📖 Полные данные:

После создания frontend делает:
```
GET /api/admin/recipes/{id}
```

И получает **ВСЕ** поля для формы редактирования:
```json
{
  "id": "...",
  "localName": "...",
  "canonicalName": "...",
  "title": "",
  "description": "...",
  "imageUrl": "...",
  "country": "PL",
  "category": "main",
  "difficulty": "easy",
  "timeMinutes": 30,
  "servings": 1,
  "source": {"type":"manual"},
  "status": "draft",
  "authorId": "...",
  "grossWeight": null,
  "netWeight": null,
  "calories": null,
  "protein": null,
  "fats": null,
  "carbs": null,
  "tokensReward": 10,
  "viewsCount": 0,
  "tokensEarned": 0,
  "createdAt": "...",
  "updatedAt": "..."
}
```

---

## 🔥 Summary

| Этап | Количество полей | Назначение |
|------|------------------|------------|
| **Request** | 3 required, 13 optional | Минимум для создания draft |
| **Database** | 35+ columns | Полное хранение + defaults |
| **Response** | 8 fields | Только для UI feedback |
| **GET detail** | 35+ fields | Полные данные для редактирования |

**Архитектурный принцип:**  
POST возвращает минимум → Frontend подгружает полные данные через GET

**Status:** ✅ Оптимизировано для UX и безопасности
