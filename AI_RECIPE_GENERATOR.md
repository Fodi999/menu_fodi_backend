# 🤖 AI Recipe Generator API

AI-powered recipe generation system using Groq API (Llama 3-70B model). Generate complete recipes from just a dish name with full nutrition analytics, cost estimation, and ChefTokens rewards.

## 📡 Endpoint

**POST** `/api/ai/recipe-helper`

Generate a complete recipe from a dish name using AI.

### Request Body

```json
{
  "title": "Філадельфія рол",
  "language": "pl"
}
```

**Parameters:**
- `title` (string, required) - Name of the dish to generate recipe for
- `language` (string, optional) - Recipe language. Options: `pl` (Polish), `en` (English), `ru` (Russian), `ua` (Ukrainian). Default: `pl`

### Response

```json
{
  "status": "success",
  "data": {
    "title": "Філадельфія рол",
    "description": "Свіжий рол з лососем, крем-сиром і авокадо, завернутий в норі",
    "category": "sushi",
    "difficulty": "intermediate",
    "time": 60,
    "portions": 4,
    "grossWeight": 580,
    "netWeight": 520,
    "calories": 920,
    "protein": 42.5,
    "fats": 35.0,
    "carbs": 98.0,
    "yield": 480,
    "cost": 52.00,
    "tokensReward": 25,
    "ingredients": [
      {
        "name": "Рис для суші",
        "amount": 300,
        "unit": "г",
        "gross": 300,
        "net": 300
      },
      {
        "name": "Лосось",
        "amount": 150,
        "unit": "г",
        "gross": 180,
        "net": 150
      }
    ],
    "steps": [
      "Відварити рис для суші за інструкцією",
      "Підготувати всі інгредієнти: нарізати лосось, авокадо",
      "Покласти норі на бамбуковий килимок",
      "Розподілити рис тонким шаром",
      "Викласти начинку та згорнути рол",
      "Нарізати на 8 частин гострим ножем",
      "Подавати з соєвим соусом, васабі та маринованим імбиром"
    ]
  }
}
```

### Response Fields

#### Basic Information
- `title` (string) - Recipe name
- `description` (string) - Brief 1-2 sentence description
- `category` (string) - Recipe category: `sushi`, `ramen`, `appetizers`, `desserts`, `fusion`, `other`
- `difficulty` (string) - Difficulty level: `beginner`, `intermediate`, `advanced`
- `time` (integer) - Total cooking time in minutes
- `portions` (integer) - Number of servings

#### Nutrition Metrics
- `grossWeight` (integer) - Total gross weight of raw ingredients in grams (before processing)
- `netWeight` (integer) - Net weight after cleaning, peeling, trimming in grams
- `calories` (integer) - Total calories in kcal for entire dish
- `protein` (float) - Protein content in grams
- `fats` (float) - Fat content in grams
- `carbs` (float) - Carbohydrate content in grams
- `yield` (integer) - Final cooked dish weight in grams (usually less due to evaporation)

#### Economics
- `cost` (float) - Estimated total cost in PLN (Polish Zloty)

#### ChefTokens System
- `tokensReward` (integer) - ChefTokens reward for creating this recipe (10-50 based on complexity)
  - 10 tokens: Simple recipes (under 30 min, few ingredients)
  - 20-30 tokens: Medium complexity (30-60 min, moderate ingredients)
  - 40-50 tokens: Complex recipes (60+ min, many ingredients, advanced techniques)

#### Ingredients
Array of ingredient objects:
- `name` (string) - Ingredient name
- `amount` (number) - Quantity
- `unit` (string) - Unit of measurement (`g`, `ml`, `pcs`, `tbsp`, `tsp`, etc.)
- `gross` (integer) - Gross weight in grams (before processing)
- `net` (integer) - Net weight in grams (after cleaning/trimming)

#### Steps
- `steps` (array of strings) - Detailed cooking instructions in order

## 💡 Usage Examples

### Polish Recipe
```bash
curl -X POST http://localhost:8080/api/ai/recipe-helper \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Ramen z wieprzowiną",
    "language": "pl"
  }'
```

### English Recipe
```bash
curl -X POST http://localhost:8080/api/ai/recipe-helper \
  -H "Content-Type: application/json" \
  -d '{
    "title": "California Roll",
    "language": "en"
  }'
```

### Russian Recipe
```bash
curl -X POST http://localhost:8080/api/ai/recipe-helper \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Борщ украинский",
    "language": "ru"
  }'
```

### Ukrainian Recipe
```bash
curl -X POST http://localhost:8080/api/ai/recipe-helper \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Вареники з картоплею",
    "language": "ua"
  }'
```

## 🔄 Integration with Recipe Creation

You can directly use the AI-generated data to create a recipe:

```javascript
// 1. Generate recipe with AI
const aiResponse = await fetch('/api/ai/recipe-helper', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    title: 'California Roll',
    language: 'en'
  })
});

const { data: recipe } = await aiResponse.json();

// 2. Create recipe post with all metrics
const createResponse = await fetch('/api/recipes', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${userToken}`
  },
  body: JSON.stringify({
    title: recipe.title,
    description: recipe.description,
    imageUrl: 'https://your-cdn.com/california-roll.jpg',
    grossWeight: recipe.grossWeight,
    netWeight: recipe.netWeight,
    calories: recipe.calories,
    protein: recipe.protein,
    fats: recipe.fats,
    carbs: recipe.carbs,
    yield: recipe.yield,
    cost: recipe.cost,
    tokensReward: recipe.tokensReward
  })
});
```

## 🎯 AI Model Details

- **Model:** Llama 3-70B (via Groq API)
- **Provider:** Groq
- **Temperature:** 0.7 (balanced creativity and consistency)
- **Max Tokens:** 2000
- **Response Format:** Pure JSON (no markdown, no code blocks)

## 📊 Nutrition Data Accuracy

The AI generates realistic nutrition data based on:
- Standard ingredient nutrition databases
- Typical cooking methods and their effects
- Professional chef knowledge
- Common portion sizes
- Ingredient waste during processing (gross vs net weight)
- Water loss during cooking (net weight vs yield)

**Note:** AI-generated nutrition data is an estimate. For critical applications, verify with professional nutrition analysis.

## 🏆 ChefTokens Reward System

Recipes earn ChefTokens in two ways:

1. **Creation Reward** (`tokensReward`):
   - Automatically assigned by AI based on recipe complexity
   - Awarded once when recipe is published
   - Range: 10-50 tokens

2. **View Rewards** (`tokensEarned`):
   - 1 token per 10 views
   - Tracked via `/api/recipes/{id}/view` endpoint
   - Unlimited earning potential

## 🌍 Multi-Language Support

| Language | Code | Example Dish |
|----------|------|--------------|
| Polish   | `pl` | Ramen z wieprzowiną |
| English  | `en` | California Roll |
| Russian  | `ru` | Борщ украинский |
| Ukrainian | `ua` | Філадельфія рол |

Each language has custom prompts optimized for:
- Cultural recipe context
- Local ingredient availability
- Regional cooking terminology
- Appropriate measurement units

## ⚠️ Error Handling

### Invalid JSON from AI
If AI returns invalid JSON, fallback response:
```json
{
  "title": "User's input title",
  "description": "AI's raw response text",
  "category": "other",
  "difficulty": "intermediate",
  "time": 30,
  "portions": 4,
  "ingredients": [],
  "steps": ["Szczegóły przepisu zostaną uzupełnione."]
}
```

### API Errors
Standard error response:
```json
{
  "error": "Error message"
}
```

Common errors:
- `400 Bad Request` - Invalid input (missing title, invalid language)
- `500 Internal Server Error` - AI API failure or server error

## 🔒 Rate Limiting

**Recommendation:** Implement rate limiting on frontend:
- AI generation is compute-intensive
- Consider: 5 requests per user per minute
- Cache popular recipes to reduce API calls

## 💰 Cost Considerations

- Each AI generation costs ~$0.001-0.01 depending on recipe complexity
- Groq API offers competitive pricing for Llama 3-70B
- Consider caching common recipes (e.g., "California Roll", "Ramen")

## 📈 Future Enhancements

Planned features:
- [ ] Recipe image generation with AI (DALL-E, Midjourney)
- [ ] Dietary restrictions (vegetarian, vegan, gluten-free, halal, kosher)
- [ ] Cuisine-specific models (Japanese, Italian, Mexican, etc.)
- [ ] User feedback learning (upvote/downvote recipes)
- [ ] Recipe variations ("make it spicier", "vegetarian version")
- [ ] Ingredient substitutions
- [ ] Allergen detection and warnings
- [ ] Wine/drink pairing suggestions

## 🧪 Testing

Test the AI generator:
```bash
# Test Polish
curl -X POST http://localhost:8080/api/ai/recipe-helper \
  -H "Content-Type: application/json" \
  -d '{"title": "Pierogi ruskie", "language": "pl"}' | jq .

# Test English
curl -X POST http://localhost:8080/api/ai/recipe-helper \
  -H "Content-Type: application/json" \
  -d '{"title": "Pad Thai", "language": "en"}' | jq .

# Test Russian with nutrition focus
curl -X POST http://localhost:8080/api/ai/recipe-helper \
  -H "Content-Type: application/json" \
  -d '{"title": "Оливье", "language": "ru"}' | jq '.data | {calories, protein, fats, carbs}'
```

## 📝 License

Part of FodiFoo Menu Backend - AI-powered recipe platform.
