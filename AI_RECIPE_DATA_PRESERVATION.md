# AI Recipe System with Data Preservation

## Overview
The AI recipe creation system now **strictly preserves all user input data** and supports **multilingual recipes** (Polish, English, Russian).

## Key Design Principles

### 1. Backend as Source of Truth
- **Title**: AI must NOT modify the recipe title
- **Ingredients**: AI must NOT add, remove, or change ingredient IDs, amounts, or units
- **Language**: Descriptions and steps must be in the user-specified language
- **Validation**: System validates that AI response matches input exactly

### 2. Data Flow

```
User Input (Title + Ingredients + RawText + Language)
    ↓
Ingredient Enrichment (DB lookup with localization)
    ↓
AI System Prompt (strict rules to preserve data)
    ↓
AI Response Validation (check title, ingredient IDs match)
    ↓
Database Storage (save in correct language fields)
```

## API Contract

### Request Structure

**POST** `/api/admin/recipes/create-ai`
**POST** `/api/admin/recipes/preview-ai`

```json
{
  "title": "Łosoś z ryżem i sosem sojowym",
  "language": "pl",
  "ingredients": [
    {
      "ingredientId": "uuid",
      "quantity": 150,
      "unit": "g"
    }
  ],
  "rawCookingText": "Rybę opłukać i osuszyć..."
}
```

**Fields:**
- `title` (required, string): Recipe title in target language
- `language` (optional, string): "pl", "en", or "ru" (default: "en")
- `ingredients` (required, array): List of ingredients with IDs, amounts, units
- `rawCookingText` (required, string): Unstructured cooking instructions

### Response Structure (Preview)

```json
{
  "title": "Łosoś z ryżem i sosem sojowym",
  "language": "pl",
  "description": "Pyszny łosoś grillowany podany z ryżem...",
  "servings": 1,
  "time_minutes": 25,
  "difficulty": "easy",
  "calories": 520,
  "steps": [
    {
      "order": 1,
      "text": "Rybę opłukać i osuszyć papierowym ręcznikiem",
      "time": 2
    }
  ],
  "ingredients": [
    {
      "ingredientId": "uuid",
      "name": "Łosoś",
      "amount": 150,
      "unit": "g"
    }
  ]
}
```

### Response Structure (Create)

Returns full `RecipeCatalog` model with additional fields:
- `id`, `canonicalName`, `country`, `category`
- `descriptionPl`, `descriptionEn`, `descriptionRu` (based on language)
- `stepsPl`, `stepsEn`, `stepsRu` (based on language)
- `nutritionProfile` (JSONB with calories, protein, fat, carbs)

## Implementation Details

### 1. Language Support

**Request → Enrichment → AI → Storage**

```go
// User specifies language in request
req.Language = "pl" // or "en", "ru"

// Ingredient names are localized during enrichment
func enrichIngredientsForAI(inputs []RecipeIngredientInput, lang string) {
    name := getLocalizedName(ingredient, lang) // "Łosoś" for pl
}

// AI generates content in specified language
systemPrompt := fmt.Sprintf("Return the recipe in language: %s", lang)

// Storage uses correct field based on language
switch aiResponse.Language {
case "pl":
    recipe.DescriptionPl = &aiResponse.Description
    recipe.StepsPl = stepsJSON
case "ru":
    recipe.DescriptionRu = &aiResponse.Description
    recipe.StepsRu = stepsJSON
default: // "en"
    recipe.DescriptionEn = &aiResponse.Description
    recipe.StepsEn = stepsJSON
}
```

### 2. Data Preservation Validation

**AI Response Validation:**

```go
func validateAIResponse(response, originalTitle, originalIngredients) error {
    // 1. Title must match exactly
    if response.Title != originalTitle {
        return error("AI changed the title")
    }

    // 2. Ingredient count must match
    if len(response.Ingredients) != len(originalIngredients) {
        return error("ingredient count mismatch")
    }

    // 3. All ingredient IDs must be present
    for _, ing := range response.Ingredients {
        if !originalIDs[ing.IngredientID] {
            return error("unknown ingredient ID")
        }
    }
}
```

### 3. AI System Prompt

**Critical Rules:**
```
YOU MUST:
1. DO NOT change the recipe title provided by user
2. DO NOT invent, add, or remove any ingredients
3. Use ONLY the ingredients with EXACT amounts and units
4. Return recipe in language: {language}
5. Output ONLY valid JSON

YOU ARE STRUCTURING DATA, NOT CREATING A NEW RECIPE.
```

**JSON Schema Enforcement:**
```json
{
  "title": "EXACT from input",
  "language": "EXACT from input",
  "description": "string in {language}",
  "ingredients": [
    {"ingredientId": "uuid FROM INPUT", "amount": "EXACT FROM INPUT"}
  ]
}
```

## Testing

### Test Script

```bash
./test_ai_recipe_pl.sh
```

**Test Scenarios:**
1. ✅ Preview AI recipe (Polish)
2. ✅ Validate title preservation
3. ✅ Validate ingredient IDs match
4. ✅ Validate language is "pl"
5. ✅ Create recipe in database
6. ✅ Verify recipe in catalog

### Manual Testing

**Polish Recipe:**
```bash
curl -X POST "$BASE_URL/admin/recipes/preview-ai" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Łosoś z ryżem",
    "language": "pl",
    "ingredients": [...],
    "rawCookingText": "Rybę grillować..."
  }'
```

**English Recipe:**
```bash
curl -X POST "$BASE_URL/admin/recipes/preview-ai" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Grilled Salmon with Rice",
    "language": "en",
    "ingredients": [...],
    "rawCookingText": "Grill the fish..."
  }'
```

## Error Handling

### Validation Errors

**Title Changed:**
```json
{
  "error": "AI changed the title: expected 'Łosoś z ryżem', got 'Salmon with Rice'"
}
```

**Ingredient Missing:**
```json
{
  "error": "ingredient count mismatch: expected 3, got 2"
}
```

**Unknown Ingredient ID:**
```json
{
  "error": "ingredient #2 has unknown ID: abc-123"
}
```

### AI Errors

**Invalid JSON:**
```json
{
  "error": "failed to parse AI JSON: unexpected token"
}
```

**Missing Required Fields:**
```json
{
  "error": "description is empty"
}
```

## Future Improvements

### 1. Full Nutrition Support
Currently AI only estimates calories. Future versions should calculate:
- Protein (g)
- Fat (g)
- Carbohydrates (g)

Based on ingredient nutrition groups and amounts.

### 2. Additional Languages
Add support for:
- Spanish (es)
- German (de)
- French (fr)

### 3. Country/Category Detection
Auto-detect `country` and `category` fields based on:
- Recipe title analysis
- Ingredient combinations
- Cooking techniques

### 4. Image Generation
Integration with DALL-E or Stable Diffusion for:
- Recipe cover images
- Step-by-step visual guides

### 5. Dietary Tags
Auto-tagging based on ingredients:
- Vegetarian, Vegan, Gluten-Free
- Keto, Paleo, Low-Carb
- Allergen warnings

## Migration Guide

### For Frontend Developers

**Old Request Format:**
```javascript
// ❌ Old (no language support)
{
  "title": "Recipe",
  "ingredients": [...],
  "rawCookingText": "..."
}
```

**New Request Format:**
```javascript
// ✅ New (with language)
{
  "title": "Przepis",
  "language": "pl", // Add this!
  "ingredients": [...],
  "rawCookingText": "..."
}
```

**Old Response:**
```javascript
// ❌ Old (summary + nutrition object)
{
  "summary": "Description",
  "nutrition": {
    "calories": 520,
    "protein": 38,
    "fat": 22,
    "carbohydrate": 42
  }
}
```

**New Response:**
```javascript
// ✅ New (description + ingredients with IDs)
{
  "title": "Przepis",
  "language": "pl",
  "description": "Opis",
  "calories": 520, // Flat field
  "ingredients": [ // Now included!
    {"ingredientId": "uuid", "name": "...", "amount": 150, "unit": "g"}
  ]
}
```

## Deployment Checklist

- [x] Update `CreateRecipeAIRequest` to include `language` field
- [x] Update `AIRecipeResponse` structure (title, description, ingredients)
- [x] Rewrite AI system prompt with strict preservation rules
- [x] Add `validateAIResponse()` with title/ingredient checks
- [x] Update `enrichIngredientsForAI()` to support localization
- [x] Update `saveRecipeToDB()` to store in correct language fields
- [x] Create test script `test_ai_recipe_pl.sh`
- [x] Update documentation
- [ ] Deploy to production
- [ ] Test with pl/en/ru languages
- [ ] Monitor AI response quality
- [ ] Update frontend integration

## Related Documentation

- `docs/GROQ_QUICK_REF.md` - Groq API usage patterns
- `docs/RECIPE_API_ENDPOINTS.md` - All recipe endpoints
- `MULTILINGUAL_INGREDIENTS.md` - Localization system
- `docs/AI_RETRY_QUICK_REF.md` - Error handling patterns
