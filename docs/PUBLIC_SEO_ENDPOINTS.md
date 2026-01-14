# 🌍 SEO-Ready Public Recipe Endpoints

## Overview
Public, unauthenticated endpoints for recipe catalog access. Optimized for SEO, search engines, and organic traffic.

---

## 📋 Endpoints

### 1. Public Recipe Catalog
**Endpoint:** `GET /api/public/recipes`  
**Auth:** ❌ No authentication required  
**Cache:** ✅ 120 seconds (`public, max-age=120`)

**Query Parameters:**
- `category` - Filter by category (main, salad, soup, dessert, appetizer)
- `difficulty` - Filter by difficulty (easy, medium, hard)
- `timeLte` - Maximum cooking time in minutes
- `timeGte` - Minimum cooking time in minutes
- `limit` - Results per page (default: 20, max: 50)
- `page` - Page number (default: 1)
- `sort` - Sorting: `newest` | `time_asc` | `time_desc` | `name_asc` | `name_desc`

**Example Requests:**
```bash
# All recipes (newest first)
curl http://localhost:8080/api/public/recipes

# Main dishes, easy difficulty, max 30 minutes
curl 'http://localhost:8080/api/public/recipes?category=main&difficulty=easy&timeLte=30'

# Paginated (5 per page)
curl 'http://localhost:8080/api/public/recipes?limit=5&page=1'
```

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "canonicalName": "recipe-slug",
      "title": "Recipe Title",
      "namePl": "Polish Name",
      "nameEn": "English Name",
      "nameRu": "Russian Name",
      "descriptionPl": "...",
      "descriptionEn": "...",
      "descriptionRu": "...",
      "country": "pl",
      "region": "mazovia",
      "category": "main",
      "difficulty": "easy",
      "timeMinutes": 30,
      "servings": 4,
      "portionWeightGrams": 350,
      "stepsPl": [...],
      "stepsEn": [...],
      "stepsRu": [...],
      "nutritionProfile": {
        "calories": 450,
        "protein": 25,
        "fat": 15,
        "carbohydrate": 50
      },
      "createdAt": "2026-01-11"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 125,
    "count": 20
  }
}
```

---

### 2. Single Recipe by Slug (SEO URL)
**Endpoint:** `GET /api/public/recipes/{slug}`  
**Auth:** ❌ No authentication required  
**Cache:** ✅ 300 seconds (`public, max-age=300`)

**Parameters:**
- `slug` - Recipe canonical name (URL-friendly identifier)

**Example Requests:**
```bash
# English slug
curl http://localhost:8080/api/public/recipes/pasta-carbonara

# URL-encoded Cyrillic slug
curl 'http://localhost:8080/api/public/recipes/%D0%BF%D0%B0%D1%81%D1%82%D0%B0_%D0%BA%D0%B0%D1%80%D0%B1%D0%BE%D0%BD%D0%B0%D1%80%D0%B0'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "canonicalName": "pasta-carbonara",
    "title": "Pasta Carbonara",
    "namePl": "Makaron Carbonara",
    "nameEn": "Pasta Carbonara",
    "nameRu": "Паста Карбонара",
    "descriptionPl": "...",
    "descriptionEn": "...",
    "descriptionRu": "...",
    "country": "it",
    "region": "lazio",
    "category": "main",
    "difficulty": "medium",
    "timeMinutes": 25,
    "servings": 4,
    "portionWeightGrams": 350,
    "stepsPl": [...],
    "stepsEn": [...],
    "stepsRu": [...],
    "nutritionProfile": {...},
    "createdAt": "2026-01-10"
  }
}
```

**Error Response (404):**
```json
{
  "error": "Recipe not found"
}
```

---

## 🔒 Security & Privacy

### Public DTO (Data Transfer Object)
Internal fields are **NOT exposed** in public endpoints:
- ❌ `authorId` - Hidden (internal)
- ❌ `source` - Hidden (internal)
- ❌ `updatedAt` - Hidden (internal)

Only **safe, consumer-facing data** is returned.

---

## ⚡ Performance Optimizations

### 1. Response Caching
- **Catalog:** 120 seconds (`Cache-Control: public, max-age=120`)
- **Single Recipe:** 300 seconds (`Cache-Control: public, max-age=300`)

### 2. Database Indexes
Applied indexes for fast filtering:
```sql
-- Single indexes
idx_recipes_category
idx_recipes_difficulty
idx_recipes_time
idx_recipes_created_at

-- Composite indexes
idx_recipes_category_difficulty
idx_recipes_category_time
```

### 3. Query Optimization
- Efficient filtering via `RecipeFilter` DTO
- Declarative query building (no if-else hell)
- Eager loading with `Preload()` (avoid N+1 queries)
- Total count optimization (separate COUNT query)

---

## 🌍 SEO Features

### 1. **Canonical URLs**
Each recipe has a unique, SEO-friendly slug:
```
/api/public/recipes/pasta-carbonara
/api/public/recipes/chicken-tikka-masala
/api/public/recipes/пельмени_домашние
```

### 2. **Multilingual Support**
Every recipe includes:
- `namePl` - Polish name
- `nameEn` - English name
- `nameRu` - Russian name
- `descriptionPl/En/Ru` - Localized descriptions
- `stepsPl/En/Ru` - Cooking instructions

### 3. **Cache Headers**
All responses include proper `Cache-Control` headers for:
- CDN caching
- Browser caching
- Search engine crawlers

### 4. **Response Time**
- Target: <100ms (cached)
- Acceptable: <300ms (uncached)
- Slow query warnings: >300ms

---

## 🧪 Testing

### Quick Test
```bash
# 1. Start server
./bin/server

# 2. Test catalog
curl 'http://localhost:8080/api/public/recipes?limit=5'

# 3. Get first recipe slug
SLUG=$(curl -s 'http://localhost:8080/api/public/recipes?limit=1' | jq -r '.data[0].canonicalName')

# 4. Test single recipe
curl "http://localhost:8080/api/public/recipes/${SLUG}"
```

### Comprehensive Test Suite
```bash
chmod +x test_public_recipes.sh
./test_public_recipes.sh
```

**Tests include:**
1. ✅ Public catalog (no auth)
2. ✅ Category filtering
3. ✅ Pagination
4. ✅ Single recipe by slug
5. ✅ 404 handling
6. ✅ Public DTO (no internal fields)
7. ✅ Multilingual fields
8. ✅ Cache headers
9. ✅ Response time

---

## 🚀 Deployment Checklist

### Backend
- [ ] Deploy updated `bin/server` binary
- [ ] Verify database indexes applied
- [ ] Test public endpoints (no auth)
- [ ] Verify cache headers present
- [ ] Monitor response times (<300ms)

### CDN / Proxy
- [ ] Configure cache rules:
  - `/api/public/recipes` → 120s cache
  - `/api/public/recipes/*` → 300s cache
- [ ] Enable gzip compression
- [ ] Add CORS headers (if needed)

### SEO
- [ ] Submit sitemap to Google Search Console
- [ ] Add Open Graph meta tags (frontend)
- [ ] Implement structured data (JSON-LD)
- [ ] Monitor indexing status

---

## 📊 Monitoring

### Key Metrics
1. **Response Time**: Target <100ms (cached), <300ms (uncached)
2. **Cache Hit Rate**: Aim for >80%
3. **Error Rate**: <1% (404s expected for invalid slugs)
4. **Query Performance**: Log slow queries (>300ms)

### Logs to Watch
```
⏱️ Recipe catalog took 1.23s (filters: {...})
⚠️ SLOW QUERY: Recipe catalog took 692ms
```

---

## 🔄 Next Steps

### Immediate
- ✅ Public catalog endpoint
- ✅ Single recipe by slug
- ✅ Cache headers
- ✅ Public DTO

### Short-term
- [ ] Sitemap generation (`/sitemap.xml`)
- [ ] RSS feed (`/feed.xml`)
- [ ] Category pages (`/recipes/category/{category}`)
- [ ] Response caching layer (Redis)

### Long-term
- [ ] CDN integration (Cloudflare, Fastly)
- [ ] Full-text search (Elasticsearch)
- [ ] Related recipes (AI recommendations)
- [ ] Recipe ratings & reviews

---

## 🐛 Troubleshooting

### Issue: Recipe not found (404)
**Cause:** Slug mismatch or URL encoding issue  
**Solution:** Ensure slug matches `canonicalName` exactly (URL-encoded if Cyrillic)

### Issue: Slow responses
**Cause:** Missing database indexes or large result sets  
**Solution:** Apply indexes, use pagination, monitor slow query logs

### Issue: Cache not working
**Cause:** Headers set after response written  
**Solution:** Set headers **before** calling `utils.RespondWithJSON()`

---

## 📚 Related Documentation

- `RECIPE_FILTERS_QUICK_REF.md` - Filtering system
- `RECIPE_CATALOG_QUICK_REF.md` - Catalog structure
- `RECIPE_DATABASE_STRUCTURE.md` - Database schema
- `test_public_recipes.sh` - Test suite

---

**Status:** ✅ **PRODUCTION READY**  
**Version:** 1.0.0  
**Last Updated:** 2026-01-11
