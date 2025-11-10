# Fridge Chat Integration Guide

## Overview

This document describes the integration between the AI Chef Mentor chat and the Fridge module, allowing users to save recipe ingredients directly to their fridge through conversation.

## Features

### 1. Chef Mentor Chat
- **Endpoint**: `POST /api/ai/chef-mentor`
- **Description**: Interactive AI assistant that helps create recipes step-by-step
- **Authentication**: Not required
- **Response**: Includes suggested actions when recipe is complete

```json
{
  "message": "Assistant's response about the recipe",
  "recipe": {
    "title": "Pasta Carbonara",
    "ingredients": [
      {
        "name": "Pasta",
        "amount": 400,
        "unit": "г"
      }
    ]
  },
  "nextQuestion": "What ingredients do you need?",
  "isComplete": true,
  "suggestedActions": [
    "save_recipe",
    "save_ingredients_to_fridge",
    "generate_meal_plan"
  ]
}
```

### 2. Save Ingredients to Fridge
- **Endpoint**: `POST /api/ai/save-ingredients`
- **Authentication**: Required (JWT)
- **Description**: Saves all ingredients from a recipe to the user's fridge

#### Request Example
```json
{
  "ingredients": [
    {
      "name": "Pasta",
      "amount": 400,
      "unit": "г"
    },
    {
      "name": "Eggs",
      "amount": 3,
      "unit": "шт"
    },
    {
      "name": "Bacon",
      "amount": 200,
      "unit": "г"
    },
    {
      "name": "Parmesan Cheese",
      "amount": 100,
      "unit": "г"
    }
  ]
}
```

#### Response Example
```json
{
  "success": true,
  "message": "ingredients saved to fridge",
  "count": 4
}
```

## Usage Flow

### Step 1: Start Recipe Creation via Chat
```bash
curl -X POST "https://api.example.com/api/ai/chef-mentor" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "I want to make pasta carbonara",
    "language": "en",
    "currentRecipe": null,
    "history": []
  }'
```

### Step 2: Continue the Conversation
```bash
curl -X POST "https://api.example.com/api/ai/chef-mentor" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "I have eggs, bacon, and pasta. What else do I need?",
    "language": "en",
    "currentRecipe": {
      "title": "Pasta Carbonara",
      "ingredients": [
        {"name": "Pasta", "amount": 400, "unit": "г"},
        {"name": "Eggs", "amount": 3, "unit": "шт"},
        {"name": "Bacon", "amount": 200, "unit": "г"}
      ]
    },
    "history": [
      {"role": "user", "content": "I want to make pasta carbonara"},
      {"role": "assistant", "content": "Great choice! ..."}
    ]
  }'
```

### Step 3: Save Ingredients to Fridge
When the recipe is complete (Chef Mentor returns `"isComplete": true` and suggests `"save_ingredients_to_fridge"`):

```bash
curl -X POST "https://api.example.com/api/ai/save-ingredients" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "ingredients": [
      {"name": "Pasta", "amount": 400, "unit": "г"},
      {"name": "Eggs", "amount": 3, "unit": "шт"},
      {"name": "Bacon", "amount": 200, "unit": "г"},
      {"name": "Parmesan Cheese", "amount": 100, "unit": "г"}
    ]
  }'
```

### Step 4: Verify Fridge Items
```bash
curl -X GET "https://api.example.com/api/fridge/" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

Response will include all saved ingredients marked as available.

## Database Schema

### UserFridge Table
```sql
CREATE TABLE user_fridge (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  product VARCHAR(255) NOT NULL,
  quantity DECIMAL(10,2) NOT NULL,
  unit VARCHAR(20) NOT NULL,
  available BOOLEAN DEFAULT false,
  category VARCHAR(50),
  expiry_date TIMESTAMP,
  added_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);
```

## API Endpoints Summary

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/api/ai/chef-mentor` | POST | No | Start/continue recipe creation chat |
| `/api/ai/save-ingredients` | POST | Yes | Save recipe ingredients to fridge |
| `/api/ai/recipe-generator` | POST | No | Generate complete recipe from title |
| `/api/ai/meal-plan` | POST | Yes | Generate meal plan |
| `/api/ai/fridge-recommendations` | POST | Yes | Get recipe recommendations based on fridge items |
| `/api/fridge/` | GET | Yes | Get all fridge items |
| `/api/fridge/` | POST | Yes | Add item to fridge |
| `/api/fridge/{id}` | PUT | Yes | Update fridge item |
| `/api/fridge/{id}` | DELETE | Yes | Remove item from fridge |
| `/api/fridge/available` | GET | Yes | Get available fridge items |

## Error Handling

### Common Errors

**400 Bad Request**
- Empty ingredients list
- Invalid ingredient format
- Missing required fields

**401 Unauthorized**
- Missing JWT token
- Invalid JWT token
- Expired JWT token

**500 Internal Server Error**
- Database connection issues
- Failed to save ingredients
- AI service errors

## Implementation Details

### Service Layer
The AI service includes:
- `ChefMentor()` - Manages recipe creation conversation
- `GenerateMealPlan()` - Creates meal plans
- `GenerateRecipe()` - Generates recipes from titles
- `GetFridgeRecommendations()` - Suggests recipes based on available ingredients

### Handler Layer
The AI handler includes:
- `ChefMentor()` - HTTP handler for chat
- `GenerateMealPlan()` - HTTP handler for meal plans
- `GenerateRecipe()` - HTTP handler for recipe generation
- `GetFridgeRecommendations()` - HTTP handler for recommendations
- `SaveRecipeIngredientsToFridge()` - HTTP handler for saving ingredients

### Database Integration
- Automatically creates `UserFridge` records
- Sets `available` flag to `true` by default
- Timestamps added automatically
- User association maintained through `user_id`

## Frontend Integration

### Suggested Frontend Flow

1. User starts chat with AI Chef Mentor
2. User and AI have conversation to build recipe
3. When recipe is complete, show button "Save to Fridge"
4. Click button sends ingredients to `/api/ai/save-ingredients`
5. Show success message and update fridge view
6. Show ingredients added to fridge

### Example React Component
```typescript
// Call chef mentor
const response = await api.post('/ai/chef-mentor', {
  message: userMessage,
  language: 'en',
  currentRecipe: recipe,
  history: conversationHistory
});

// If recipe is complete and user wants to save
if (response.isComplete) {
  const saveResponse = await api.post('/ai/save-ingredients', {
    ingredients: response.recipe.ingredients
  });
  // Update fridge items
}
```

## Testing

Run the integration tests:
```bash
go test ./tests/api/... -v -run TestFridgeChat
```

Test specific functionality:
```bash
go test ./tests/api/... -v -run TestSaveIngredientsRequest
```

## Future Enhancements

1. **AI Recipe Parsing**: Automatically extract ingredients from AI responses
2. **Batch Operations**: Save multiple recipes at once
3. **Ingredient Validation**: Check for duplicates and conflicts
4. **Quantity Conversion**: Auto-convert units (g ↔ kg, ml ↔ l)
5. **Expiry Tracking**: Add expiration dates when saving ingredients
6. **Recipe Saving**: Save complete recipes to user library
7. **Nutrition Tracking**: Calculate total nutrition from saved ingredients
8. **Shopping List**: Generate shopping list from missing ingredients

## Troubleshooting

### Issue: Ingredients not saving
- Check JWT token validity
- Verify database connection
- Check logs for database errors

### Issue: Chat not responding
- Verify AI service (Groq) is configured
- Check API key
- Review error logs

### Issue: Fridge items not showing
- Verify `added_at` column exists (not `created_at`)
- Check user_id association
- Verify available flag is set correctly

## References

- [Fridge Module Documentation](./internal/modules/fridge/README.md)
- [AI Module Documentation](./internal/modules/ai/README.md)
- [API Documentation](./ENDPOINTS_SUMMARY.txt)
