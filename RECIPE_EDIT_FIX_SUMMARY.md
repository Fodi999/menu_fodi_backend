# Recipe Edit Fix - Quick Summary

## 🐛 Problem
Backend **always created new recipe** (INSERT) instead of updating existing one when `recipeId` was provided.

```
Frontend: POST /api/admin/recipes/save { recipeId: "123", ... }
Backend: "Creating NEW recipe with NEW UUID..."
Database: ERROR duplicate key "canonicalName" ❌
```

## ✅ Solution (2 commits)

### Commit 1: `afc8906` - Duplicate Check Fix
- Added `RecipeID *string` field to `SaveEditedRecipeRequest`
- Modified duplicate check: exclude current recipe from canonicalName search
```go
if req.RecipeID != nil && *req.RecipeID != "" {
    query = query.Where("id != ?", *req.RecipeID)
}
```

### Commit 2: `ac5d7d4` - CREATE vs UPDATE Logic
- Added `isEditMode` detection: `recipeId != nil && recipeId != ""`
- **EDIT MODE**: Load existing → update fields → `Updates(map)`
- **CREATE MODE**: New UUID → `Create()` → `Save()`

## 🔧 Technical Changes

**File:** `internal/modules/admin/service/recipe_ai.go`

### Before ❌
```go
recipe := &models.RecipeCatalog{
    ID: uuid.New(),  // Always new!
    // ...
}
tx.Create(recipe)  // Always INSERT!
```

### After ✅
```go
if isEditMode {
    // Load existing recipe
    recipe = &models.RecipeCatalog{}
    tx.First(recipe, "id = ?", recipeID)
    
    // Update fields
    recipe.Title = req.Title
    // ...
    
    // Save with Updates
    tx.Model(recipe).Updates(map[string]interface{}{
        "title": recipe.Title,
        // ...
    })
} else {
    // Create new
    recipe = &models.RecipeCatalog{
        ID: uuid.New(),
        // ...
    }
    tx.Create(recipe)
}
```

## 📋 Frontend Requirements

**Must send `recipeId` when editing:**

```typescript
// ✅ EDIT: Include recipeId
fetch('/api/admin/recipes/save', {
  body: JSON.stringify({
    recipeId: "4aa22783-45cc-4fc4-8800-4340a5c93ce9",  // ✅ Required!
    title: "яйца жареные на масле",
    ingredients: [...],
    steps: [...]
  })
});

// ✅ CREATE: Omit recipeId
fetch('/api/admin/recipes/save', {
  body: JSON.stringify({
    // no recipeId = create new
    title: "новый рецепт",
    ingredients: [...],
    steps: [...]
  })
});
```

## 🧪 Testing

### Test 1: Edit existing recipe ✅
```bash
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/admin/recipes/save \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "recipeId": "4aa22783-45cc-4fc4-8800-4340a5c93ce9",
    "title": "яйца жареные на масле",
    "language": "ru",
    "description": "Updated description",
    "servings": 2,
    "time_minutes": 10,
    "difficulty": "easy",
    "calories": 250,
    "ingredients": [...],
    "steps": [...]
  }'

# Expected: 200 OK, recipe updated
# Backend log: 📝 Editing existing recipe: ID=4aa22783...
```

### Test 2: Create new recipe ✅
```bash
# Same request but WITHOUT recipeId
# Expected: 200 OK, new recipe created
# Backend log: ✨ Creating new recipe: ID=<new-uuid>
```

### Test 3: Create duplicate ❌
```bash
# Without recipeId + existing title
# Expected: 409 Conflict "Recipe name already exists"
```

## 🔍 Debugging

### Backend Logs
```
📝 Editing existing recipe: ID=xxx     → EDIT mode
✨ Creating new recipe: ID=xxx         → CREATE mode
✅ Recipe updated: ID=xxx              → UPDATE successful
✅ Recipe saved: ID=xxx                → INSERT successful
```

### Database Queries
```sql
-- EDIT mode
SELECT * FROM "Recipe" WHERE id = '4aa22783...' LIMIT 1;  -- Load existing
UPDATE "Recipe" SET ... WHERE id = '4aa22783...';         -- Update

-- CREATE mode
INSERT INTO "Recipe" (...) VALUES (...) RETURNING id;     -- Insert new
```

## 📊 Comparison Table

| Aspect | CREATE MODE | EDIT MODE |
|--------|-------------|-----------|
| **Trigger** | `recipeId == nil` | `recipeId != nil` |
| **UUID** | `uuid.New()` | From request |
| **Database** | `tx.Create()` | `tx.First()` + `tx.Updates()` |
| **Ingredients** | `tx.Create()` | `tx.Delete()` + `tx.Create()` |
| **SQL** | INSERT | SELECT + UPDATE + DELETE + INSERT |

## 📁 Files Changed

- `internal/modules/admin/service/recipe_ai.go` (lines 501-750)
  - Added `RecipeID *string` field
  - Added `isEditMode` logic
  - Split create/update paths
- `docs/RECIPE_EDIT_DUPLICATE_FIX.md` (detailed explanation)

## ✅ Status

- **Commits:** `afc8906` + `ac5d7d4`
- **Deployed:** Production (Koyeb auto-deploy)
- **Tested:** Waiting for frontend verification
- **Priority:** Critical (editing was completely broken)

## 🔗 Related Docs

- Full explanation: `docs/RECIPE_EDIT_DUPLICATE_FIX.md`
- Image upload guide: `docs/CLOUDINARY_IMAGE_UPLOAD_GUIDE.md`
- Frontend integration: `docs/FRONTEND_RECIPE_CREATION_WITH_IMAGES.md`
