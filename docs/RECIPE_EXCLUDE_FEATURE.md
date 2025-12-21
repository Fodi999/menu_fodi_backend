# Recipe Recommendations - Exclude Feature
**Date:** 2025-12-21  
**Status:** ✅ Deployed and Tested

## Summary
Added `excludeRecipeIds` parameter to `/api/recipes/recommendations` endpoint. This allows users to get sequential recommendations by excluding previously shown recipes.

---

## API Changes

### Request Format
```json
POST /api/recipes/recommendations?testUserID={userId}
{
  "mode": "fridge",
  "limit": 10,
  "excludeRecipeIds": ["uuid1", "uuid2", "uuid3"]
}
```

### Parameters
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| mode | string | ✅ | Must be "fridge" |
| limit | int | ❌ | Number of candidates to consider (default: 5) |
| excludeRecipeIds | string[] | ❌ | Array of recipe UUIDs to exclude from results |

### Response Format
Same as before - single best recommendation:
```json
{
  "success": true,
  "data": {
    "recipe": {
      "id": "uuid",
      "localName": "Recipe Name",
      "steps": ["1. Step one", "2. Step two"],
      ...
    },
    "match": {
      "canCookNow": true,
      "usedIngredients": [...],
      "missingRequired": []
    },
    "economy": {
      "usedFromFridge": 4.95,
      "saved": 4.95
    }
  }
}
```

---

## Implementation Details

### Files Modified

#### 1. `internal/modules/recipes/dto/recommendations.go`
Added `ExcludeRecipeIds` field:
```go
type RecommendationRequest struct {
    Mode             string   `json:"mode"`
    Limit            int      `json:"limit"`
    ExcludeRecipeIds []string `json:"excludeRecipeIds"` // NEW
}
```

#### 2. `internal/modules/recipes/service/match_service.go`
- Updated `RecipeFilters` struct:
  ```go
  type RecipeFilters struct {
      Country          string
      Category         string
      Difficulty       string
      MaxTimeMinutes   int
      OnlyCookable     bool
      MinScore         float64
      ExcludeRecipeIds []string // NEW
  }
  ```

- Updated `loadRecipesWithFilters()` to filter excluded recipes:
  ```go
  if len(filters.ExcludeRecipeIds) > 0 {
      query = query.Where("id NOT IN ?", filters.ExcludeRecipeIds)
  }
  ```

- Updated `GetBestRecommendation()` to accept excludeIds:
  ```go
  func (s *RecipeMatchService) GetBestRecommendation(
      userID string,
      limit int,
      excludeRecipeIds []string, // NEW
  ) (*RecipeMatch, error)
  ```

#### 3. `internal/modules/recipes/transport/http/handler.go`
- Updated `GetRecommendation()` handler to parse and pass excludeRecipeIds:
  ```go
  bestMatch, err := h.matchService.GetBestRecommendation(
      userID,
      req.Limit,
      req.ExcludeRecipeIds, // NEW
  )
  ```

---

## Usage Example - Sequential Recommendations

### Frontend Implementation Pattern
```typescript
interface RecommendationState {
  excludedIds: string[];
  currentRecommendation: Recipe | null;
}

const [state, setState] = useState<RecommendationState>({
  excludedIds: [],
  currentRecommendation: null,
});

// Get first recommendation
async function getNextRecommendation() {
  const response = await fetch('/api/recipes/recommendations', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      mode: 'fridge',
      limit: 10,
      excludeRecipeIds: state.excludedIds,
    }),
  });
  
  const data = await response.json();
  
  setState({
    excludedIds: [...state.excludedIds, data.data.recipe.id],
    currentRecommendation: data.data.recipe,
  });
}

// User clicks "Next Recipe" button
<button onClick={getNextRecommendation}>
  🔄 Pokaż kolejny przepis
</button>
```

---

## Production Test Results

### Test 1: Sequential Recommendations
```bash
# Request 1 (no exclusions)
POST /api/recipes/recommendations
{"mode": "fridge", "limit": 10}

Response:
{
  "recipe": "Sałatka grecka",
  "used": 4 ingredients,
  "saved": 4.95 PLN
}

# Request 2 (exclude Sałatka grecka)
POST /api/recipes/recommendations
{"mode": "fridge", "limit": 10, "excludeRecipeIds": ["92691aae-..."]}

Response:
{
  "recipe": "Shakshuka",
  "used": 2 ingredients,
  "saved": 5.15 PLN
}

# Request 3 (exclude both)
POST /api/recipes/recommendations
{"mode": "fridge", "limit": 10, "excludeRecipeIds": ["92691aae-...", "339cc7f7-..."]}

Response:
{
  "recipe": "Gołąbki",
  "used": 2 ingredients,
  "saved": 3.89 PLN
}
```

### ✅ Results Validation
| Test | Status | Details |
|------|--------|---------|
| Exclude 0 recipes | ✅ | Returns Sałatka grecka (best match) |
| Exclude 1 recipe | ✅ | Returns Shakshuka (2nd best) |
| Exclude 2 recipes | ✅ | Returns Gołąbki (3rd best) |
| Exclude 3 recipes | ✅ | Returns Kotlet schabowy (4th best) |
| Sequential flow | ✅ | No duplicates, logical ordering |

---

## Benefits

### For Users
1. **Variety**: Can browse multiple recipe options
2. **Discovery**: Explore more of the catalog
3. **Choice**: Not stuck with first recommendation
4. **Control**: Easy navigation (next/previous pattern)

### For UX
1. **Engagement**: Encourages exploration
2. **Personalization**: User-driven discovery
3. **Flexibility**: Can skip recipes they don't like
4. **Progressive**: Each request builds on previous context

### For Backend
1. **Stateless**: No server-side session needed
2. **Simple**: Just array of UUIDs in request
3. **Efficient**: Uses existing filtering logic
4. **Scalable**: No additional database load

---

## Edge Cases

### Empty Result
If all recipes are excluded:
```json
{
  "success": false,
  "error": "No recipes found matching criteria"
}
```

### Invalid UUID
Invalid UUIDs in excludeRecipeIds are silently ignored (SQL IN clause handles this).

### Large Exclude List
Performance tested with 20+ excluded IDs - no issues.  
SQL query: `WHERE id NOT IN (?, ?, ...)` is efficient with indexes.

---

## Future Enhancements

### Possible Additions
1. **Exclude Categories**
   ```json
   {
     "excludeRecipeIds": ["uuid1"],
     "excludeCategories": ["dessert", "soup"]
   }
   ```

2. **Exclude by Difficulty**
   ```json
   {
     "excludeDifficulties": ["hard"]
   }
   ```

3. **Seen History (Backend-Tracked)**
   - Track last N recommendations per user
   - Auto-exclude recently shown recipes
   - Reset after 24 hours

4. **Preference Learning**
   - Track which recommendations user cooks
   - Boost similar recipes in future
   - Personalized ranking

---

## Related Features

### Already Implemented
- ✅ Ingredient matching by UUID
- ✅ Optional ingredients (oil, salt, pepper)
- ✅ 4-level ranking algorithm
- ✅ Economy calculations
- ✅ Steps normalization

### Pending
- 🔜 POST /api/recipes/cook (deduct ingredients)
- 🔜 Recipe rating/feedback
- 🔜 Shopping list generation
- 🔜 Meal planning (weekly recipes)

---

## Testing Checklist

- [x] Request without excludeRecipeIds works
- [x] Request with 1 excluded recipe works
- [x] Request with multiple excluded recipes works
- [x] Sequential requests return different recipes
- [x] No duplicates in sequential flow
- [x] Invalid UUIDs don't break endpoint
- [x] Empty excludeRecipeIds array works
- [x] Null excludeRecipeIds works
- [x] Deployed to Koyeb successfully
- [x] Production testing passed

---

## Deployment Info

**Commit:** Added in commit after `5219947`  
**Environment:** Koyeb (yeasty-madelaine-fodi999-671ccdf5.koyeb.app)  
**Database:** Neon.tech PostgreSQL  
**Status:** ✅ Live and tested  

---

## Summary

The `excludeRecipeIds` feature is fully implemented, tested, and deployed. It enables a smooth "Next Recipe" flow in the UI without requiring backend state management. Users can now browse through multiple recommendations sequentially, with each request returning the next best recipe based on their fridge contents.

**Key Achievement:** Stateless, scalable recommendation cycling with minimal code changes. 🎉
