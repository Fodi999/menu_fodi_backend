# Recipe Detail with Fridge Check API

## 🎯 Problem Solved

**Проблема**: На странице `/recipes/[id]` эндпоинт `/api/recipes/{canonicalName}` НЕ проверяет холодильник пользователя. Ингредиенты не помечаются как `inFridge`.

**Решение**: Новый эндпоинт `/api/recipe-recommendations/{id}` возвращает рецепт с проверкой холодильника.

---

## 📍 Endpoint

```
GET /api/recipe-recommendations/{id}
```

### Headers
```
Authorization: Bearer <jwt_token>
```

### Path Parameters
- `id` (string, required): Recipe identifier
  - Supports **UUID**: `605c8419-2d42-4ef0-a9d2-839582e98727`
  - Supports **canonical_name**: `zharenye_yaytsa`

### Query Parameters
- `lang` (string, optional): Language code
  - Values: `pl` (default), `ru`, `en`

---

## 📦 Response

### Success (200 OK)

```json
{
  "id": "605c8419-2d42-4ef0-a9d2-839582e98727",
  "title": "Жареные яйца",
  "canonical_name": "zharenye_yaytsa",
  "image_url": "https://res.cloudinary.com/.../recipe_605c8419.webp",
  "cook_time": 7,
  "servings": 1,
  "match_percent": 66.67,
  "match_status": "almost_ready",
  "available_ingredients": [
    {
      "id": "3260aadf-52de-4038-9568-ee536495224a",
      "canonical_name": "яйца",
      "display_name": "Яйца",
      "quantity": 3,
      "unit": "pcs",
      "category": "egg"
    },
    {
      "id": "c4d477f8-9123-4175-b515-5201ee1ff61b",
      "canonical_name": "соль",
      "display_name": "Соль",
      "quantity": 2,
      "unit": "g",
      "category": "condiment"
    }
  ],
  "missing_ingredients": [
    {
      "id": "9ff773d2-a3ee-4f4b-bc45-4cfe0d7f680b",
      "canonical_name": "растительное_масло",
      "display_name": "Растительное масло",
      "quantity": 30,
      "unit": "ml",
      "category": "condiment"
    }
  ],
  "steps": [
    "Нагреть раскалённую сковороду",
    "Разбить три яйца так, чтобы желтки остались целыми",
    "Обжарить яйца до золотистой корочки",
    "Положить натарелку и посыпать солью"
  ]
}
```

### Error (404 Not Found)

```json
{
  "success": false,
  "error": "failed to get recipe",
  "message": "recipe not found: invalid_id"
}
```

---

## 🔧 Architecture

### Flow

```
Frontend Request
    ↓
Handler (HTTP Layer)
    ↓
Service (Business Logic)
    ↓
Repository (Data Access)
    ↓
PostgreSQL
```

### Components

**1. Repository** (`recipe_repository.go`)
```go
func GetRecipeByIDOrCanonical(ctx, identifier) (*RecipeCatalog, error) {
    // Try UUID first
    if uuid.Parse(identifier) == nil {
        return GetByUUID(identifier)
    }
    // Fallback to canonical_name
    return GetByCanonicalName(identifier)
}
```

**2. Service** (`recommendation_service.go`)
```go
func GetSingleRecipeWithFridge(ctx, req) (*RecipeDTO, error) {
    // 1. Get user's fridge (ingredient_id set)
    fridgeIngredientIDs := getUserFridgeIngredientIDs(ctx, req.UserID)
    
    // 2. Get recipe (UUID or canonical_name)
    recipe := recipeRepository.GetRecipeByIDOrCanonical(ctx, req.RecipeID)
    
    // 3. Build DTO with fridge check
    dto := buildRecipeDTO(recipe, fridgeIngredientIDs, req.Language)
    
    return dto
}
```

**3. Handler** (`recommendation_handler.go`)
```go
func GetSingleRecipeWithFridge(w, r) {
    userID := middleware.GetUserID(r)
    recipeID := chi.URLParam(r, "id")
    lang := r.URL.Query().Get("lang")
    
    response := service.GetSingleRecipeWithFridge(ctx, req)
    utils.RespondJSON(w, http.StatusOK, response)
}
```

---

## 🧪 Testing

### cURL Example

```bash
# By canonical_name
curl "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipe-recommendations/zharenye_yaytsa?lang=ru" \
  -H "Authorization: Bearer <TOKEN>"

# By UUID
curl "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipe-recommendations/605c8419-2d42-4ef0-a9d2-839582e98727?lang=ru" \
  -H "Authorization: Bearer <TOKEN>"
```

### TypeScript (Frontend)

```typescript
interface RecipeDetailWithFridge {
  id: string;
  title: string;
  canonical_name: string;
  image_url: string;
  cook_time: number;
  servings: number;
  match_percent: number;
  match_status: "ready" | "almost_ready" | "not_ready";
  available_ingredients: IngredientInfo[];
  missing_ingredients: IngredientInfo[];
  steps: string[];
}

interface IngredientInfo {
  id: string;
  canonical_name: string;
  display_name: string;
  quantity: number;
  unit: string;
  category: string;
}

// Usage
async function getRecipeWithFridge(recipeId: string, lang: string = 'ru') {
  const response = await fetch(
    `/api/recipe-recommendations/${recipeId}?lang=${lang}`,
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }
  );
  
  if (!response.ok) {
    throw new Error('Recipe not found');
  }
  
  const recipe: RecipeDetailWithFridge = await response.json();
  
  // Now you can check inFridge status:
  console.log('Available:', recipe.available_ingredients); // ✅ In fridge
  console.log('Missing:', recipe.missing_ingredients);     // ❌ Not in fridge
  
  return recipe;
}
```

---

## 📊 Comparison with Old Endpoint

| Feature | `/api/recipes/{id}` | `/api/recipe-recommendations/{id}` |
|---------|---------------------|-----------------------------------|
| **Checks Fridge** | ❌ No | ✅ Yes |
| **Ingredients Status** | ❌ No `inFridge` | ✅ Split: `available` / `missing` |
| **Match Metrics** | ❌ No | ✅ `match_percent`, `match_status` |
| **Authentication** | Optional | Required |
| **Lookup Support** | UUID + canonical_name | UUID + canonical_name |
| **Localization** | ✅ Yes | ✅ Yes |

---

## 🎯 Use Cases

### Use Case 1: Recipe Detail Page

**Page**: `/recipes/[id]`

**Old Implementation** (broken):
```typescript
// ❌ No fridge check
const recipe = await fetch(`/api/recipes/${id}`);
// Result: ingredients have NO inFridge status
```

**New Implementation** (correct):
```typescript
// ✅ With fridge check
const recipe = await fetch(`/api/recipe-recommendations/${id}`);
// Result: ingredients split into available / missing
```

### Use Case 2: Show Matching Indicators

```typescript
function RecipeDetailPage({ recipeId }) {
  const { data: recipe } = useSWR(
    `/api/recipe-recommendations/${recipeId}?lang=${lang}`,
    fetcher
  );
  
  return (
    <div>
      <h1>{recipe.title}</h1>
      
      {/* Match indicator */}
      <MatchBadge 
        status={recipe.match_status} 
        percent={recipe.match_percent} 
      />
      
      {/* Available ingredients (green checkmark) */}
      <h2>В холодильнике ({recipe.available_ingredients.length})</h2>
      {recipe.available_ingredients.map(ing => (
        <IngredientCard key={ing.id} ingredient={ing} inFridge={true} />
      ))}
      
      {/* Missing ingredients (red X) */}
      <h2>Нужно докупить ({recipe.missing_ingredients.length})</h2>
      {recipe.missing_ingredients.map(ing => (
        <IngredientCard key={ing.id} ingredient={ing} inFridge={false} />
      ))}
    </div>
  );
}
```

---

## 🚀 Benefits

1. **Reuses Existing Logic**: Uses the same `buildRecipeDTO` from recommendation engine
2. **Clean Architecture**: Repository → Service → Handler (no duplication)
3. **Flexible Lookup**: Supports both UUID and canonical_name
4. **Single Source of Truth**: Fridge checking logic centralized in Service
5. **Type-Safe**: Full TypeScript interfaces available

---

## 📝 Migration Guide

### Step 1: Update Frontend Code

**Before**:
```typescript
// ❌ Old endpoint (no fridge check)
const recipe = await fetch(`/api/recipes/${id}`);
```

**After**:
```typescript
// ✅ New endpoint (with fridge check)
const recipe = await fetch(`/api/recipe-recommendations/${id}?lang=ru`);
```

### Step 2: Update Component

**Before**:
```typescript
<IngredientList ingredients={recipe.ingredients} />
```

**After**:
```typescript
<>
  <IngredientList 
    title="В холодильнике" 
    ingredients={recipe.available_ingredients} 
    inFridge={true}
  />
  <IngredientList 
    title="Нужно докупить" 
    ingredients={recipe.missing_ingredients} 
    inFridge={false}
  />
</>
```

---

## 🔒 Security

- **Required**: JWT authentication (Bearer token)
- **User-Specific**: Returns fridge status for authenticated user only
- **Private**: Other users cannot see your fridge status

---

## ⚡ Performance

- **Optimized**: Single query with `Preload("Ingredients").Preload("Ingredients.Ingredient")`
- **Fast**: Fridge check is O(1) lookup (map-based)
- **Cacheable**: Consider adding Redis cache for popular recipes

---

## 📚 Related Endpoints

- `GET /api/recipes` - Catalog (no fridge check)
- `GET /api/recipes/{id}` - Single recipe (no fridge check)
- `GET /api/recipe-recommendations` - Multiple recipes with matching
- `GET /api/recipe-recommendations/{id}` - **Single recipe with fridge check** (THIS)

---

## 🎉 Summary

**NEW**: `GET /api/recipe-recommendations/{id}`
- ✅ Checks user's fridge
- ✅ Returns `available_ingredients` / `missing_ingredients`
- ✅ Supports UUID and canonical_name
- ✅ Reuses existing matching logic
- ✅ Perfect for recipe detail pages

**Status**: ✅ Deployed to production (commit `d6e2bd0`)
