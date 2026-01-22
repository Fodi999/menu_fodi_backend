# Testing New Endpoint: Recipe Detail with Fridge Check

## 🎯 New Endpoint

```
GET /api/recipe-recommendations/{id}?lang=ru
```

**Purpose**: Returns ONE recipe with fridge check (inFridge status for each ingredient)

---

## 🧪 How to Test

### Step 1: Get Fresh Token

Login to get a fresh JWT token:

```bash
curl -X POST "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "fodi85@gmail.ru",
    "password": "YOUR_PASSWORD"
  }'
```

Copy the `token` from response.

---

### Step 2: Test New Endpoint

#### By Canonical Name:

```bash
curl -X GET "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipe-recommendations/zharenye_yaytsa?lang=ru" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

#### By UUID:

```bash
curl -X GET "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipe-recommendations/605c8419-2d42-4ef0-a9d2-839582e98727?lang=ru" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

---

## ✅ Expected Response

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

---

## 🔍 Key Features to Verify

### ✅ Fridge Check Working

- `available_ingredients`: Ингредиенты, которые ЕСТЬ в холодильнике
- `missing_ingredients`: Ингредиенты, которых НЕТ в холодильнике

### ✅ Match Metrics

- `match_percent`: 66.67 (2 из 3 ингредиентов)
- `match_status`: "almost_ready" (missing ≤ 2)

### ✅ Flexible Lookup

Both work:
- `/api/recipe-recommendations/zharenye_yaytsa` (canonical_name)
- `/api/recipe-recommendations/605c8419-...` (UUID)

### ✅ Localization

- `lang=ru`: "Жареные яйца", "Яйца", "Растительное масло"
- `lang=pl`: "Smażone jajka", "Jaja", "Olej rzepakowy"
- `lang=en`: "Fried Eggs", "Eggs", "Vegetable oil"

---

## 📊 Comparison

### Old Endpoint (Broken):

```bash
GET /api/recipes/zharenye_yaytsa
```

**Response**:
```json
{
  "ingredients": [
    {
      "id": "...",
      "display_name": "Яйца",
      "inFridge": false  // ❌ Always false (not checked!)
    }
  ]
}
```

### New Endpoint (Working):

```bash
GET /api/recipe-recommendations/zharenye_yaytsa
```

**Response**:
```json
{
  "available_ingredients": [
    {"display_name": "Яйца"}  // ✅ In fridge
  ],
  "missing_ingredients": [
    {"display_name": "Растительное масло"}  // ✅ Not in fridge
  ]
}
```

---

## 🚀 Frontend Integration

### TypeScript Example:

```typescript
interface RecipeDetailWithFridge {
  id: string;
  title: string;
  canonical_name: string;
  match_percent: number;
  match_status: "ready" | "almost_ready" | "not_ready";
  available_ingredients: IngredientInfo[];
  missing_ingredients: IngredientInfo[];
  steps: string[];
}

async function getRecipeWithFridge(recipeId: string) {
  const response = await fetch(
    `/api/recipe-recommendations/${recipeId}?lang=ru`,
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }
  );
  
  const recipe: RecipeDetailWithFridge = await response.json();
  
  console.log('✅ Available:', recipe.available_ingredients);
  console.log('❌ Missing:', recipe.missing_ingredients);
  
  return recipe;
}
```

---

## 📝 Status

- ✅ **Deployed**: Commit `d6e2bd0`
- ✅ **Koyeb**: Instance healthy
- ✅ **Routes**: Registered in `module.go`
- ⏳ **Testing**: Need fresh JWT token

---

## 🔗 Related Files

- **Handler**: `recommendation_handler.go` (line 73)
- **Service**: `recommendation_service.go` (line 329)
- **Repository**: `recipe_repository.go` (line 85)
- **Module**: `module.go` (line 49)
- **Documentation**: `RECIPE_DETAIL_WITH_FRIDGE_API.md`
