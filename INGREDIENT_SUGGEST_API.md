# Ingredient Suggest API - Quick Reference

## Endpoint

```
GET /api/admin/ingredients/suggest
```

## Purpose
Fast autocomplete search for ingredients (no AI, database-only). Supports multilingual results.

## Authentication
- **Required**: Yes (JWT Bearer token)
- **Role**: Admin or Super Admin

## Request Parameters

### Query Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `q` | string | Yes | - | Search query (min 2 chars) |
| `limit` | integer | No | 5 | Max results (1-20) |

### Headers

| Header | Required | Example | Description |
|--------|----------|---------|-------------|
| `Authorization` | Yes | `Bearer eyJhbG...` | JWT token |
| `Accept-Language` | No | `pl`, `en`, `ru` | Response language (default: `en`) |

## Response Format

### Success (200 OK)

```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Локализованное название",
      "unit": "g",
      "category": "vegetable",
      "nutritionGroup": "vegetable"
    }
  ]
}
```

### Error (500 Internal Server Error)

```json
{
  "error": "Failed to fetch suggestions"
}
```

## Localization

Ingredient names are returned in the requested language:

| Language Code | Example Result |
|--------------|----------------|
| `pl` | "Pomidor" |
| `en` | "Tomato" |
| `ru` | "Помидор" |

**Fallback logic:**
1. Try requested language (`Accept-Language`)
2. Fall back to `namePl` (Polish)
3. Fall back to `name` (default)

## Examples

### Basic Search (English)

```bash
GET /api/admin/ingredients/suggest?q=tom&limit=5
Accept-Language: en
Authorization: Bearer eyJhbG...
```

**Response:**
```json
{
  "data": [
    {
      "id": "abc-123",
      "name": "Tomato",
      "unit": "g",
      "category": "vegetable",
      "nutritionGroup": "vegetable"
    },
    {
      "id": "def-456",
      "name": "Tomato paste",
      "unit": "g",
      "category": "condiment",
      "nutritionGroup": "condiment"
    }
  ]
}
```

### Search with Polish Localization

```bash
GET /api/admin/ingredients/suggest?q=пом&limit=3
Accept-Language: ru
Authorization: Bearer eyJhbG...
```

**Response:**
```json
{
  "data": [
    {
      "id": "abc-123",
      "name": "Помидор",
      "unit": "g",
      "category": "vegetable",
      "nutritionGroup": "vegetable"
    }
  ]
}
```

## Search Logic

**Search Fields:**
- `name` (default name)
- `name_pl` (Polish)
- `name_en` (English)
- `name_ru` (Russian)

**Search Method:**
- Case-insensitive
- Matches anywhere in the name (ILIKE '%query%')
- Searches across all language fields
- Orders by category, then name

**Example Query:**
```sql
SELECT * FROM "Ingredient" 
WHERE LOWER(name) LIKE '%tom%'
   OR LOWER(name_pl) LIKE '%tom%'
   OR LOWER(name_en) LIKE '%tom%'
   OR LOWER(name_ru) LIKE '%tom%'
ORDER BY category ASC, name ASC
LIMIT 5
```

## Performance

- **Cache**: None (direct DB query)
- **Index**: Recommended on `name`, `name_pl`, `name_en`, `name_ru`
- **Typical Response**: 10-50ms
- **Max Query Length**: 100 characters (auto-truncated)

## Frontend Integration

### React Hook Example

```typescript
import { apiClient } from '@/lib/api/base';

interface IngredientSuggestion {
  id: string;
  name: string;
  unit: string;
  category: string;
  nutritionGroup: string;
}

export async function suggestIngredients(
  query: string, 
  limit: number = 5,
  language: 'pl' | 'en' | 'ru' = 'en'
): Promise<IngredientSuggestion[]> {
  const response = await apiClient.get(
    `/admin/ingredients/suggest?q=${encodeURIComponent(query)}&limit=${limit}`,
    {
      headers: {
        'Accept-Language': language
      }
    }
  );
  
  return response.data;
}
```

### Usage in Component

```typescript
const [suggestions, setSuggestions] = useState<IngredientSuggestion[]>([]);

const handleSearch = async (query: string) => {
  if (query.length < 2) {
    setSuggestions([]);
    return;
  }
  
  const results = await suggestIngredients(query, 10, 'ru');
  setSuggestions(results);
};
```

## Validation

### Query Validation
- ✅ Minimum length: 2 characters
- ✅ Maximum length: 100 characters
- ✅ Auto-trimmed whitespace
- ✅ Auto-truncated if too long

### Limit Validation
- ✅ Minimum: 1
- ✅ Maximum: 20
- ✅ Default: 5

### Language Validation
- ✅ Supported: `pl`, `en`, `ru`
- ✅ Default: `en`
- ✅ Normalized: `pl-PL` → `pl`

## Error Handling

### Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| 401 Unauthorized | Missing/invalid token | Check JWT token |
| 500 Internal | Database error | Check server logs |
| Empty results | No matches | Try broader query |

### Debugging

**Server logs:**
```
📥 Request: GET /suggest?q=tom&limit=5 (Accept-Language: en → en)
🔍 SuggestIngredients: query='tom', limit=5, lang='en'
✅ Returning 2 suggestions (lang=en)
```

## Related Endpoints

| Endpoint | Purpose |
|----------|---------|
| `GET /api/admin/ingredients` | Full ingredient list with pagination |
| `POST /api/admin/ingredients` | Create new ingredient |
| `PUT /api/admin/ingredients/{id}` | Update ingredient |
| `DELETE /api/admin/ingredients/{id}` | Delete ingredient |

## Migration from Frontend Route

### ❌ Old (Next.js API Route)

```typescript
// app/api/admin/ingredients/suggest/route.ts
export async function GET(req: Request) {
  const query = new URL(req.url).searchParams.get('q');
  
  // Direct SQL query or fetch to Go API
  const response = await fetch(
    process.env.GO_API + '/admin/ingredients/suggest?q=' + query
  );
  
  return Response.json(await response.json());
}
```

### ✅ New (Proxy to Go)

```typescript
// app/api/admin/ingredients/suggest/route.ts
import { proxyToGo } from '@/lib/api-proxy';

export async function GET(req: Request) {
  return proxyToGo(req, '/admin/ingredients/suggest');
}
```

**Why?**
- ✅ No code duplication
- ✅ Single source of truth (Go backend)
- ✅ Consistent error handling
- ✅ Automatic header forwarding (`Accept-Language`, `Authorization`)
- ✅ Easier to maintain

## Testing

### Manual Test (curl)

```bash
curl -X GET \
  'http://localhost:8080/api/admin/ingredients/suggest?q=tom&limit=5' \
  -H 'Authorization: Bearer YOUR_TOKEN' \
  -H 'Accept-Language: en'
```

### Automated Test (Go)

```go
func TestSuggestIngredients(t *testing.T) {
    req := httptest.NewRequest("GET", "/api/admin/ingredients/suggest?q=tom&limit=5", nil)
    req.Header.Set("Accept-Language", "en")
    w := httptest.NewRecorder()
    
    handler.SuggestIngredients(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
    // Add more assertions
}
```

## Performance Tips

1. **Use debouncing** on frontend (wait 300ms before searching)
2. **Cache results** for same query
3. **Add DB indexes** on name fields for faster search
4. **Limit query length** to prevent slow queries
5. **Use connection pooling** for DB

## Roadmap

### Phase 1 (Current)
- ✅ Basic search with localization
- ✅ Accept-Language support
- ✅ Multi-field search

### Phase 2 (Planned)
- 🔜 Search by category filter
- 🔜 Fuzzy matching (Levenshtein distance)
- 🔜 Search history/popular ingredients

### Phase 3 (Future)
- 🔮 AI-powered search suggestions
- 🔮 Synonym support ("tomato" = "помидор")
- 🔮 Nutritional filter (high protein, low carb)

---

**Last Updated:** 2026-01-08  
**API Version:** 1.0  
**Status:** ✅ Production Ready
