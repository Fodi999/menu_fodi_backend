# AI Recipe Quick Reference

## Endpoints

```
POST /api/admin/recipes/preview-ai  - Preview without saving
POST /api/admin/recipes/create-ai   - Create and save to DB
```

## Request Format

```json
{
  "title": "Recipe Title",
  "language": "pl|en|ru",
  "ingredients": [
    {"ingredientId": "uuid", "quantity": 150, "unit": "g"}
  ],
  "rawCookingText": "Cooking instructions..."
}
```

## Response Format

```json
{
  "title": "Recipe Title (preserved)",
  "language": "pl",
  "description": "1-2 sentence description in target language",
  "servings": 1,
  "time_minutes": 25,
  "difficulty": "easy|medium|hard",
  "calories": 520,
  "steps": [
    {"order": 1, "text": "Step description", "time": 5}
  ],
  "ingredients": [
    {"ingredientId": "uuid", "name": "...", "amount": 150, "unit": "g"}
  ]
}
```

## Data Preservation Rules

✅ **AI MUST preserve:**
- Exact recipe title
- All ingredient IDs
- Exact amounts and units
- Ingredient count

❌ **AI CANNOT:**
- Change title
- Add/remove ingredients
- Modify amounts or units
- Return in wrong language

## Testing

```bash
# Polish
./test_ai_recipe_pl.sh

# Manual curl
curl -X POST "$BASE_URL/admin/recipes/preview-ai" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"Łosoś z ryżem","language":"pl",...}'
```

## Validation

System validates:
1. Title matches input exactly
2. Ingredient count matches
3. All ingredient IDs are original
4. Language is correct
5. All required fields present

## Language Support

| Code | Language | Fields |
|------|----------|--------|
| `pl` | Polish   | descriptionPl, stepsPl |
| `en` | English  | descriptionEn, stepsEn |
| `ru` | Russian  | descriptionRu, stepsRu |

Default: `en` if not specified

## Example Usage

```bash
# Get ingredient IDs
SALMON_ID=$(curl "$BASE_URL/ingredients/suggest?query=salmon" | jq -r '.[0].id')

# Preview recipe
curl -X POST "$BASE_URL/admin/recipes/preview-ai" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"title\": \"Łosoś z ryżem\",
    \"language\": \"pl\",
    \"ingredients\": [{\"ingredientId\": \"$SALMON_ID\", \"quantity\": 150, \"unit\": \"g\"}],
    \"rawCookingText\": \"Rybę grillować 5 minut z każdej strony.\"
  }"
```

## Common Errors

```json
// Title changed
{"error": "AI changed the title: expected '...', got '...'"}

// Ingredient mismatch
{"error": "ingredient count mismatch: expected 3, got 2"}

// Unknown ID
{"error": "ingredient #2 has unknown ID: abc-123"}
```
