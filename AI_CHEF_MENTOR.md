# 🤖 AI Chef Mentor - Interactive Recipe Assistant

## Overview
AI Chef Mentor is an interactive conversational assistant that guides users through creating professional recipes step-by-step, just like a real chef teaching a student.

**Endpoint:** `POST /api/ai/chef-mentor`

---

## Features
- ✅ Natural conversational interface
- ✅ Step-by-step recipe building
- ✅ Multi-language support (Ukrainian, English, Russian, Polish)
- ✅ Contextual questions based on what's missing
- ✅ Real-time recipe draft updates
- ✅ Suggested actions for quick workflows

---

## Request Format

```json
{
  "message": "Хочу зробити роли Дракон з вугром, авокадо та огірком 🐉",
  "language": "ua",
  "history": [
    {
      "role": "assistant",
      "content": "Привіт! Я AI Chef Mentor. Яку страву ви хочете приготувати?"
    },
    {
      "role": "user",
      "content": "Хочу зробити суші"
    }
  ],
  "currentRecipe": {
    "title": "Роли Дракон",
    "category": "sushi",
    "difficulty": "intermediate",
    "time": 45,
    "portions": 4,
    "ingredients": [],
    "steps": []
  }
}
```

### Fields:
- `message` (required) - User's current message
- `language` (optional) - "ua", "en", "ru", "pl" (default: "ua")
- `history` (optional) - Previous conversation messages
- `currentRecipe` (optional) - Current state of the recipe being built

---

## Response Format

```json
{
  "status": "success",
  "data": {
    "message": "Чудово! Роли Дракон з вугром звучать смачно! Скільки порцій ви хочете приготувати?",
    "recipe": {
      "title": "Роли Дракон з вугром, авокадо та огірком",
      "category": "sushi",
      "difficulty": "intermediate",
      "time": 45,
      "portions": 4,
      "ingredients": [
        {
          "name": "рис (суші-рис)",
          "amount": 200,
          "unit": "г",
          "gross": 200,
          "net": 200
        }
      ],
      "steps": [
        "Крок 1: Приготуйте рис..."
      ],
      "grossWeight": 805,
      "netWeight": 750,
      "calories": 1096,
      "protein": 34.0,
      "fats": 51.2,
      "carbs": 102.0,
      "yield": 480,
      "cost": 62.0,
      "tokensReward": 25,
      "isComplete": false
    },
    "nextQuestion": "Скільки порцій ви хочете приготувати?",
    "isComplete": false,
    "suggestedActions": [
      "Додати інгредієнт",
      "Додати крок приготування"
    ]
  }
}
```

### Response Fields:
- `message` - AI assistant's natural language response
- `recipe` - Current state of the recipe draft
- `nextQuestion` - Suggested next question to guide the user
- `isComplete` - Boolean indicating if recipe is complete
- `suggestedActions` - Array of quick action buttons

---

## Conversation Flow Example

### 1. Start Conversation
```bash
POST /api/ai/chef-mentor
{
  "message": "Привіт!",
  "language": "ua"
}
```

**Response:**
```json
{
  "message": "Привіт! Я AI Chef Mentor з Академії Діми Фоміна. Яку страву ви хочете приготувати сьогодні?",
  "recipe": {},
  "nextQuestion": "Яка назва вашої страви?",
  "isComplete": false,
  "suggestedActions": ["Розпочати новий рецепт", "Показати приклад"]
}
```

### 2. Describe Dish
```bash
POST /api/ai/chef-mentor
{
  "message": "Хочу зробити роли Дракон з вугром, авокадо та огірком, зверху соус унагі 🐉",
  "language": "ua",
  "history": [...]
}
```

**Response:**
```json
{
  "message": "Чудово! Роли Дракон з вугром звучать смачно! Це класична складна страва. Скільки порцій ви хочете приготувати?",
  "recipe": {
    "title": "Роли Дракон з вугром, авокадо та огірком, соус унагі",
    "category": "sushi"
  },
  "nextQuestion": "На скільки порцій розрахований рецепт?",
  "isComplete": false
}
```

### 3. Provide Details
```bash
POST /api/ai/chef-mentor
{
  "message": "4 порції, складність середня, 45 хвилин",
  "language": "ua",
  "currentRecipe": {...}
}
```

**Response:**
```json
{
  "message": "Відмінно! 4 порції за 45 хвилин. Тепер давайте додамо інгредієнти. Які основні продукти вам потрібні?",
  "recipe": {
    "title": "...",
    "category": "sushi",
    "difficulty": "intermediate",
    "time": 45,
    "portions": 4
  },
  "nextQuestion": "Які інгредієнти потрібні?",
  "isComplete": false,
  "suggestedActions": ["Додати інгредієнт"]
}
```

### 4. Add Ingredients
```bash
POST /api/ai/chef-mentor
{
  "message": "Рис суші 200г, вугор 200г, авокадо 200г, огірок 100г, соус унагі 50мл",
  "language": "ua"
}
```

**Response:**
```json
{
  "message": "Чудовий набір інгредієнтів! Я додав їх до рецепту. Тепер опишіть кроки приготування.",
  "recipe": {
    "ingredients": [
      {"name": "рис (суші-рис)", "amount": 200, "unit": "г"},
      {"name": "вугор (кремовий)", "amount": 200, "unit": "г"},
      ...
    ]
  },
  "nextQuestion": "Опишіть кроки приготування",
  "isComplete": false,
  "suggestedActions": ["Додати крок приготування"]
}
```

### 5. Complete Recipe
```bash
POST /api/ai/chef-mentor
{
  "message": "Готово, збережи рецепт!",
  "language": "ua"
}
```

**Response:**
```json
{
  "message": "Рецепт завершено! Я розрахував калорії (1096 ккал), вартість (62 PLN) та нагороду (25 ChefTokens). Бажаєте зберегти?",
  "recipe": {
    "isComplete": true,
    "grossWeight": 805,
    "netWeight": 750,
    "calories": 1096,
    "protein": 34.0,
    "cost": 62.0,
    "tokensReward": 25
  },
  "isComplete": true,
  "suggestedActions": ["Зберегти рецепт", "Розрахувати калорії", "Оцінити вартість"]
}
```

---

## Integration with Recipe Generator

Chef Mentor can work together with the existing AI Recipe Generator:

1. **Quick Generation:** Use `/api/ai/recipe-helper` for instant full recipes
2. **Guided Creation:** Use `/api/ai/chef-mentor` for interactive step-by-step building
3. **Hybrid Approach:** Start with mentor, then use generator to fill missing details

### Example Workflow:
```javascript
// 1. User starts with mentor
POST /api/ai/chef-mentor
{ "message": "Роли Філадельфія" }

// 2. Mentor collects basic info
// User provides: category, portions, time

// 3. Use generator for detailed recipe
POST /api/ai/recipe-helper
{ "title": "Роли Філадельфія", "language": "ua" }

// 4. Merge results and let user edit
// Final save with all metrics
```

---

## Language Support

| Language   | Code | Example Greeting |
|------------|------|------------------|
| Ukrainian  | `ua` | "Привіт! Я AI Chef Mentor." |
| English    | `en` | "Hello! I'm AI Chef Mentor." |
| Russian    | `ru` | "Привет! Я AI Chef Mentor." |
| Polish     | `pl` | "Cześć! Jestem AI Chef Mentor." |

---

## Frontend Integration Example

```javascript
// React component example
const [conversation, setConversation] = useState([]);
const [currentRecipe, setCurrentRecipe] = useState(null);

const sendMessage = async (message) => {
  const response = await fetch('/api/ai/chef-mentor', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      message,
      language: 'ua',
      history: conversation,
      currentRecipe
    })
  });
  
  const { data } = await response.json();
  
  // Update conversation
  setConversation([
    ...conversation,
    { role: 'user', content: message },
    { role: 'assistant', content: data.message }
  ]);
  
  // Update recipe draft
  setCurrentRecipe(data.recipe);
  
  // Show suggested actions
  if (data.isComplete) {
    showSaveButton();
  }
};
```

---

## Recipe Completion Criteria

A recipe is marked as complete (`isComplete: true`) when it has:
- ✅ Title
- ✅ Category
- ✅ At least 1 ingredient
- ✅ At least 1 cooking step

Optional fields (auto-calculated if missing):
- Difficulty (default: "intermediate")
- Time (default: 30 minutes)
- Portions (default: 4)
- Nutrition metrics (calculated from ingredients)
- Cost (estimated from ingredients)
- Tokens reward (based on complexity)

---

## Tips for Best Results

1. **Be Specific:** "Роли Філадельфія з лососем" better than just "суші"
2. **Provide Context:** "4 порції, 30 хвилин, легка складність"
3. **Use Natural Language:** AI understands conversational style
4. **Review Suggestions:** AI provides quick actions based on context
5. **Iterate:** You can always go back and add more details

---

## Error Handling

```json
{
  "status": "error",
  "message": "AI service error"
}
```

Common errors:
- `400 Bad Request` - Invalid JSON format
- `500 Internal Server Error` - AI service unavailable
- `429 Too Many Requests` - Rate limit exceeded

---

## Production URL

**Endpoint:** `https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/ai/chef-mentor`

---

## Status: 🚀 Ready for Integration

Created: November 6, 2025  
Version: 1.0  
AI Model: Groq API (Llama 3-70B)
