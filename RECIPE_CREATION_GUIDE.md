# Recipe Creation Guide - User vs Catalog Recipes

## Overview

The system supports **two types of recipes** using a shared `Recipe` table:

1. **User-Generated Recipes** - Created by registered users via POST `/api/recipes`
2. **Catalog Recipes** - System recipes managed via admin interface

---

## Recipe Types Comparison

### User-Generated Recipes (Recipe Model)

**Endpoint:** `POST /api/recipes`  
**Auth:** Required (JWT token)  
**Model:** `models.Recipe`

**Characteristics:**
- ✅ `author_id` - Set to authenticated user ID
- ✅ `source` - Automatically set to `{"type":"manual"}` by backend
- ✅ `canonicalName` - Always NULL (not used for user recipes)
- ✅ `localName` - Set to title value
- ✅ Nutrition fields optional (grossWeight, netWeight, calories, etc.)
- ✅ ChefTokens integration (tokensReward, viewsCount, tokensEarned)

**Request Example:**
```bash
curl -X POST https://your-api.com/api/recipes \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My Amazing Pierogi",
    "description": "Family recipe passed down for generations",
    "imageUrl": "https://example.com/image.jpg",
    "country": "PL",
    "category": "main",
    "difficulty": "medium",
    "timeMinutes": 45,
    "servings": 4,
    "grossWeight": 500,
    "netWeight": 450,
    "calories": 300,
    "protein": 12.5,
    "fats": 8.2,
    "carbs": 45.0
  }'
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "id": "uuid-here",
    "title": "My Amazing Pierogi",
    "localName": "My Amazing Pierogi",
    "canonicalName": null,
    "source": {"type": "manual"},
    "authorId": "user-uuid",
    "country": "PL",
    "category": "main",
    "difficulty": "medium",
    "timeMinutes": 45,
    "servings": 4,
    "tokensReward": 10,
    "viewsCount": 0,
    "tokensEarned": 0,
    "createdAt": "2026-01-05T10:30:00Z",
    "updatedAt": "2026-01-05T10:30:00Z"
  }
}
```

---

### Catalog Recipes (RecipeCatalog Model)

**Endpoint:** Admin interface or seed data  
**Auth:** Admin only  
**Model:** `models.RecipeCatalog`

**Characteristics:**
- ✅ `canonicalName` - Required (unique English name)
- ✅ `author_id` - Always NULL (system recipes)
- ✅ `source` - Set to `{"type":"cookbook","reference":"..."}` or similar
- ✅ Multilingual support (name_pl, name_en, name_ru, etc.)
- ✅ Structured with ingredients, steps, nutrition
- ✅ Used for matching with user's fridge items

**Example (Seed Data):**
```sql
INSERT INTO "Recipe" (
  "canonicalName", "localName", "title", 
  "country", "category", "difficulty",
  "source", "author_id"
) VALUES (
  'Pierogi Ruskie',           -- Unique English name
  'Pierogi ruskie',           -- Display name
  'Pierogi ruskie',           -- Title
  'Poland', 'main', 'medium',
  '{"type":"cookbook","reference":"Traditional Polish Cookbook"}',
  NULL                        -- No author (system recipe)
);
```

---

## Field Reference

### Required Fields (Both Types)

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `title` | string | Primary recipe title | "Pierogi ruskie" |
| `country` | string | Country of origin | "Poland" or "PL" |
| `category` | string | Recipe category | "main", "soup", "dessert", "appetizer", "salad" |
| `difficulty` | string | Difficulty level | "easy", "medium", "hard" |
| `timeMinutes` | int | Preparation time | 30, 45, 60 |
| `servings` | int | Number of servings | 1, 2, 4 (default: 1) |
| `source` | jsonb | Recipe source | Auto-set by backend |

### User Recipe Specific

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `author_id` | string | ✅ | User UUID (auto-set from JWT) |
| `grossWeight` | int | ❌ | Total weight before prep (grams) |
| `netWeight` | int | ❌ | Weight after prep (grams) |
| `calories` | int | ❌ | Calories per serving |
| `protein` | decimal | ❌ | Protein (grams) |
| `fats` | decimal | ❌ | Fats (grams) |
| `carbs` | decimal | ❌ | Carbohydrates (grams) |
| `yield` | int | ❌ | Recipe yield (grams) |
| `cost` | decimal | ❌ | Estimated cost (PLN) |
| `tokensReward` | int | ❌ | ChefTokens reward (default: 10) |

### Catalog Recipe Specific

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `canonicalName` | string | ✅ | Unique English identifier |
| `name_pl` | string | ❌ | Polish name |
| `name_en` | string | ❌ | English name |
| `name_ru` | string | ❌ | Russian name |
| `description_pl` | text | ❌ | Polish description |
| `description_en` | text | ❌ | English description |
| `description_ru` | text | ❌ | Russian description |
| `ingredients` | jsonb | ✅ | Structured ingredients list |
| `steps_pl` | jsonb | ❌ | Polish cooking steps |
| `steps_en` | jsonb | ❌ | English cooking steps |
| `steps_ru` | jsonb | ❌ | Russian cooking steps |

### Discriminator Fields

| Field | User Recipes | Catalog Recipes | Purpose |
|-------|--------------|-----------------|---------|
| `canonicalName` | NULL | Required (unique) | System identifier |
| `author_id` | Required | NULL | User ownership |
| `source.type` | "manual" | "cookbook", "ai", etc. | Recipe origin |

---

## Database Schema

```sql
CREATE TABLE "Recipe" (
  id VARCHAR(255) PRIMARY KEY,
  
  -- Names (unified)
  "canonicalName" VARCHAR(255) NULL,          -- NULL for user recipes
  "localName" VARCHAR(255) NOT NULL DEFAULT '',
  title VARCHAR(255) NOT NULL,
  
  -- Multilingual (catalog only)
  name_pl VARCHAR(255),
  name_en VARCHAR(255),
  name_ru VARCHAR(255),
  
  -- Descriptions
  description TEXT,
  description_pl TEXT,
  description_en TEXT,
  description_ru TEXT,
  
  -- Metadata (shared)
  "imageUrl" TEXT,
  country VARCHAR(100) NOT NULL,
  category VARCHAR(50) NOT NULL,
  difficulty VARCHAR(20) NOT NULL,
  "timeMinutes" INT NOT NULL,
  servings INT NOT NULL DEFAULT 1,
  "portionWeightGrams" INT,
  source JSONB NOT NULL,
  
  -- User recipes only
  author_id VARCHAR(255),                     -- NULL for catalog
  gross_weight INT,
  net_weight INT,
  calories INT,
  protein DECIMAL(10,2),
  fats DECIMAL(10,2),
  carbs DECIMAL(10,2),
  yield INT,
  cost DECIMAL(10,2),
  tokens_reward INT DEFAULT 10,
  views_count INT DEFAULT 0,
  tokens_earned INT DEFAULT 0,
  
  -- Timestamps
  "createdAt" TIMESTAMP NOT NULL,
  "updatedAt" TIMESTAMP NOT NULL,
  
  CONSTRAINT fk_author FOREIGN KEY (author_id) REFERENCES "User"(id)
);

-- Indexes
CREATE INDEX idx_recipe_author ON "Recipe"(author_id);
CREATE INDEX idx_recipe_category ON "Recipe"(category);
CREATE INDEX idx_recipe_difficulty ON "Recipe"(difficulty);
CREATE INDEX idx_recipe_country ON "Recipe"(country);
CREATE UNIQUE INDEX idx_recipe_canonical ON "Recipe"("canonicalName") WHERE "canonicalName" IS NOT NULL;
```

---

## Source Field Values

The `source` field tracks recipe origin:

| Source Type | Value | Used For | Who Sets |
|-------------|-------|----------|----------|
| Manual | `{"type":"manual"}` | User-created recipes | Backend (automatic) |
| Cookbook | `{"type":"cookbook","reference":"Book Name","page":123}` | Traditional recipes | Admin/Seed |
| Website | `{"type":"website","url":"https://..."}` | Online recipes | Admin |
| AI Generated | `{"type":"ai","model":"gpt-4","prompt":"..."}` | AI recipes | Backend (future) |
| Imported | `{"type":"import","from":"service-name"}` | Third-party import | Backend (future) |

**Important:** Users **CANNOT** set the source field. It's always controlled by the backend.

---

## Validation Rules

### Country Codes
- Accept both full names ("Poland") and ISO codes ("PL")
- Currently supported: Poland (PL), Italy (IT), Greece (GR)

### Categories
- `main` - Main courses
- `soup` - Soups and broths
- `dessert` - Desserts and sweets
- `appetizer` - Starters
- `salad` - Salads

### Difficulty Levels
- `easy` - Simple, beginner-friendly
- `medium` - Intermediate skills required
- `hard` - Advanced techniques

### Servings
- Minimum: 1
- Default: 1 (base portion)
- Use serving multiplier in frontend for scaling

---

## Error Handling

### Common Errors

**1. Missing Required Fields**
```json
{
  "error": "Invalid input",
  "status": 400
}
```
**Solution:** Ensure title, country, category, difficulty, timeMinutes, servings are provided.

**2. Invalid Category/Difficulty**
```json
{
  "error": "Failed to create recipe",
  "status": 500
}
```
**Solution:** Use only allowed values from validation rules.

**3. Unauthorized**
```json
{
  "error": "User not authenticated",
  "status": 401
}
```
**Solution:** Include valid JWT token in Authorization header.

**4. Author Not Found**
```json
{
  "error": "Author not found",
  "status": 404
}
```
**Solution:** Ensure user exists and token is valid.

---

## Best Practices

### For Frontend Developers

1. ✅ **Never send `source` field** - Backend sets it automatically
2. ✅ **Never send `canonicalName`** - Only for catalog recipes
3. ✅ **Always send required fields** - title, country, category, difficulty, timeMinutes, servings
4. ✅ **Use consistent country codes** - Prefer ISO codes (PL, IT, GR)
5. ✅ **Validate difficulty** - Only "easy", "medium", "hard"
6. ✅ **Optional nutrition** - Send only if available

### For Backend Developers

1. ✅ **Auto-set system fields** - source, localName, author_id, timestamps
2. ✅ **Validate enums** - category, difficulty values
3. ✅ **Use explicit GORM tags** - Prevent snake_case conversion issues
4. ✅ **Handle both recipe types** - Check author_id to distinguish
5. ✅ **Preload Author** - Always include user info in responses

---

## Testing

### Test User Recipe Creation

```bash
# 1. Login as user
TOKEN=$(curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}' \
  | jq -r '.token')

# 2. Create recipe
curl -X POST http://localhost:3000/api/recipes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Pierogi",
    "description": "Test description",
    "country": "PL",
    "category": "main",
    "difficulty": "easy",
    "timeMinutes": 30,
    "servings": 4
  }'
```

**Expected:** `201 Created` with recipe data including `id`, `authorId`, `source: {"type":"manual"}`

### Test Recipe Retrieval

```bash
# Get all user recipes
curl http://localhost:3000/api/recipes \
  -H "Authorization: Bearer $TOKEN"

# Get specific recipe
curl http://localhost:3000/api/recipes/{recipe-id} \
  -H "Authorization: Bearer $TOKEN"
```

---

## Migration History

| Migration | Purpose | Date |
|-----------|---------|------|
| 061 | Normalize recipe servings to 1 | 2026-01-05 |
| 062 | Add portionWeightGrams field | 2026-01-05 |
| 063 | Add title column (TEXT) | 2026-01-05 |
| 064 | Fix title type to VARCHAR(255) | 2026-01-05 |
| 065 | Add user recipe columns (author_id, nutrition, tokens) | 2026-01-05 |
| 066 | Make canonicalName nullable | 2026-01-05 |

---

## Troubleshooting

### Recipe Creation Returns 500

**Check:**
1. All required fields present?
2. Valid JWT token?
3. User exists in database?
4. Source field NOT sent by frontend?
5. Category/difficulty valid values?

**Debug:**
```bash
# Check backend logs
kubectl logs -f deployment/backend | grep "ERROR"

# Check database constraints
psql -c "SELECT column_name, is_nullable, data_type FROM information_schema.columns WHERE table_name = 'Recipe';"
```

### Frontend Shows "undefined" Recipe ID

**Cause:** Response not properly parsed or error during creation

**Check:**
1. Network tab shows 201 Created?
2. Response body contains `data.id`?
3. Frontend extracts ID correctly: `response.data.data.id`

---

## Summary

✅ **User Recipes**: Manual, user-owned, optional nutrition  
✅ **Catalog Recipes**: System-managed, multilingual, structured ingredients  
✅ **Shared Table**: Both use same `Recipe` table with discriminators  
✅ **Backend Control**: Source field auto-set, user cannot override  
✅ **Professional Pattern**: Clear separation via author_id and canonicalName  

**Status:** 🎉 System fully operational - both recipe types working
