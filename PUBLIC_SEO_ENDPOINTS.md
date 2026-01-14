# 🌐 SEO-Ready Public Recipe Endpoints

## Overview

Публичные endpoints для каталога рецептов без требования авторизации.  
**Цель**: SEO-оптимизация, индексация Google/Yandex, органический трафик.

---

## 📋 Endpoints

### 1. **GET** `/api/public/recipes` - Публичный каталог рецептов

**Описание**: Полный каталог с фильтрацией, пагинацией и сортировкой (без auth).

**Query Parameters**:
- `category` - Категория (main, soup, salad, appetizer, dessert)
- `difficulty` - Сложность (easy, medium, hard)
- `timeLte` - Максимальное время готовки (минуты)
- `timeGte` - Минимальное время готовки (минуты)
- `page` - Номер страницы (default: 1)
- `limit` - Рецептов на странице (default: 20, max: 50)
- `sort` - Сортировка (newest, time_asc, time_desc, name_asc, name_desc)

**Response**:
```json
{
  "data": [
    {
      "id": "uuid",
      "canonicalName": "лосось_на_сковороде_с_травами",
      "title": "Лосось на Сковороде с Травами",
      "namePl": "",
      "nameEn": "",
      "nameRu": "",
      "descriptionRu": "...",
      "category": "main",
      "difficulty": "easy",
      "timeMinutes": 15,
      "servings": 1,
      "stepsRu": [...],
      "nutritionProfile": {...},
      "createdAt": "2026-01-11"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 25,
    "count": 20
  }
}
```

**Cache**: `Cache-Control: public, max-age=120` (2 минуты)

**Examples**:
```bash
# Все рецепты
GET /api/public/recipes

# Основные блюда, лёгкие, до 30 минут
GET /api/public/recipes?category=main&difficulty=easy&timeLte=30

# Пагинация
GET /api/public/recipes?page=2&limit=10

# Сортировка по времени
GET /api/public/recipes?sort=time_asc
```

---

### 2. **GET** `/api/public/recipes/{slug}` - Рецепт по SEO URL

**Описание**: Получение одного рецепта по `canonicalName` (SEO-friendly URL).

**URL Parameter**:
- `{slug}` - Canonical name рецепта (поддерживает UTF-8 и URL-encoded)

**Response**:
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "canonicalName": "лосось_на_сковороде_с_травами",
    "title": "Лосось на Сковороде с Травами",
    "category": "main",
    "difficulty": "easy",
    "timeMinutes": 15,
    "servings": 1,
    "stepsRu": [
      {
        "order": 1,
        "text": "Обжарить лосось на сковороде",
        "time": 10
      }
    ],
    "nutritionProfile": {
      "calories": 420
    },
    "createdAt": "2026-01-11"
  }
}
```

**Cache**: `Cache-Control: public, max-age=300` (5 минут)

**URL Encoding Support** (критично для SEO):
```bash
# ✅ URL-encoded (Google crawler)
GET /api/public/recipes/%D0%BB%D0%BE%D1%81%D0%BE%D1%81%D1%8C_%D0%BD%D0%B0_%D1%81%D0%BA%D0%BE%D0%B2%D0%BE%D1%80%D0%BE%D0%B4%D0%B5_%D1%81_%D1%82%D1%80%D0%B0%D0%B2%D0%B0%D0%BC%D0%B8

# ✅ Human-readable UTF-8 (copy-paste)
GET /api/public/recipes/лосось_на_сковороде_с_травами
```

**Оба варианта работают благодаря `url.PathUnescape()`!**

---

## 🔒 Public DTO (Security)

**Скрытые поля** (не экспортируются в публичный API):
- `source` - источник рецепта (internal)
- `authorId` - ID создателя (internal)
- `updatedAt` - дата обновления (internal)

**Экспортируемые поля**:
- Все мультиязычные поля (namePl, nameEn, nameRu, descriptions, steps)
- Категория, сложность, время, порции
- Нутриционная информация
- Дата создания (`createdAt`)

---

## ⚡ Performance & SEO

### Cache Headers
- **Catalog**: `public, max-age=120` (2 минуты)
- **Single Recipe**: `public, max-age=300` (5 минут)

### Response Times
- **Catalog** (без кеша): ~900ms
- **Catalog** (с кешем): <100ms (будущая оптимизация)
- **Single Recipe**: <150ms

### Database Indexes
Используются существующие индексы из filtering system:
- `idx_recipes_category`
- `idx_recipes_difficulty`
- `idx_recipes_time`
- `idx_recipes_created_at`

---

## 🌍 Multilingual Support

Все рецепты содержат поля для 3 языков:
- `namePl`, `descriptionPl`, `stepsPl` - Polish
- `nameEn`, `descriptionEn`, `stepsEn` - English
- `nameRu`, `descriptionRu`, `stepsRu` - Russian

Frontend выбирает нужный язык на основе `Accept-Language` или user preference.

---

## 🧪 Testing

**Test Script**: `test_public_recipes.sh`

**Тесты**:
1. ✅ Публичный каталог (без auth)
2. ✅ Фильтрация по категории
3. ✅ Пагинация (limit/page)
4. ✅ Рецепт по URL-encoded slug
5. ✅ Рецепт по UTF-8 slug (human-readable)
6. ✅ 404 для несуществующих рецептов
7. ✅ Public DTO (internal поля скрыты)
8. ✅ Cache headers присутствуют
9. ✅ Response time <1s

**Run Tests**:
```bash
chmod +x test_public_recipes.sh
./test_public_recipes.sh
```

---

## 🛠️ Implementation Details

### Architecture
```
internal/modules/public/
├── module.go               # Module initialization & route registration
├── recipe_handlers.go      # HTTP handlers (GetPublicRecipes, GetRecipeBySlug)
└── README.md              # This file
```

### Service Layer
Использует **AdminService** из `internal/modules/admin/service/`:
- `GetFilteredRecipes()` - с полной фильтрацией
- `GetRecipeByCanonicalName()` - по canonicalName (с Preload)

### URL Decoding
```go
// CRITICAL FOR SEO
slugParam := chi.URLParam(r, "slug")
decodedSlug, err := url.PathUnescape(slugParam)
recipe, err := h.service.GetRecipeByCanonicalName(decodedSlug)
```

Поддерживает:
- Google crawler (URL-encoded UTF-8)
- User copy-paste (human-readable UTF-8)

---

## 🚀 Future Optimizations

### 1. In-Memory Cache
```go
type CacheEntry struct {
    Data      interface{}
    ExpiresAt time.Time
}

var recipeCache sync.Map
```

### 2. CDN Integration
- Cloudflare / Fastly для статического кеша
- Edge caching для разных регионов

### 3. Sitemap Generation
```xml
GET /sitemap.xml
<url>
  <loc>https://menu-fodi.com/recipes/лосось_на_сковороде_с_травами</loc>
  <lastmod>2026-01-11</lastmod>
  <priority>0.8</priority>
</url>
```

### 4. Meta Tags Endpoint
```json
GET /api/public/recipes/{slug}/meta
{
  "title": "Лосось на Сковороде с Травами | Menu Fodi",
  "description": "Простой рецепт жареного лосося...",
  "image": "https://cdn.menu-fodi.com/recipes/...",
  "canonical": "https://menu-fodi.com/recipes/лосось_на_сковороде_с_травами"
}
```

---

## 📊 Metrics & Monitoring

### Key Metrics (TODO)
- Request rate по endpoint
- Cache hit/miss ratio
- Average response time
- Popular recipes (топ по views)

### Logging
```bash
# Example logs
[INFO] GET /api/public/recipes?category=main&limit=10 - 200 (142ms)
[INFO] GET /api/public/recipes/лосось_на_сковороде_с_травами - 200 (87ms)
[WARN] Slow query: GET /api/public/recipes - 892ms
```

---

## ✅ Success Criteria

- ✅ Работают без авторизации
- ✅ Cache headers для SEO
- ✅ URL encoding support (UTF-8 + URL-encoded)
- ✅ Скрыты internal поля (source, authorId)
- ✅ Response time <1s (без кеша)
- ✅ Pagination работает корректно
- ✅ Filtering использует indexes
- ⏳ TODO: In-memory cache для <100ms response
- ⏳ TODO: Sitemap generation
- ⏳ TODO: Meta tags для Open Graph

---

## 🔗 Related Endpoints

- **Admin Catalog**: `GET /api/admin/recipes` (требует auth + admin role)
- **Admin Recipe CRUD**: `POST/PUT/DELETE /api/admin/recipes` (admin only)
- **Filter Metadata**: `GET /api/admin/recipes/filters/meta` (admin only)

Public endpoints - это **read-only** версия admin catalog без internal полей.

---

**Status**: ✅ Production Ready  
**Last Updated**: 2026-01-11  
**Author**: Backend Team
