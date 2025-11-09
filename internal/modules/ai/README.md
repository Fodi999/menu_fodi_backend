# AI Module

AI Module предоставляет AI-powered функциональность: интерактивный chef-mentor, генерация рецептов, планирование питания и рекомендации на основе холодильника.

## Архитектура

```
ai/
├── dto/                  # Data Transfer Objects
│   └── requests.go      # Request/Response DTOs
├── service/             # Business Logic Layer
│   └── service.go       # AI service with Groq client
├── transport/           # Transport Layer
│   └── http/
│       └── handlers.go  # HTTP handlers
├── module.go            # Module initialization
└── README.md           # This file
```

## Функциональность

### 1. Chef Mentor (Interactive Assistant)
Интерактивный AI-помощник, который помогает создавать рецепты шаг за шагом в разговорном формате.

### 2. Meal Plan Generator
Генерация плана питания на N дней с учетом калорий и доступных ингредиентов.

### 3. Recipe Generator
Автоматическая генерация полного рецепта по названию блюда.

### 4. Fridge Recommendations
Рекомендации рецептов на основе продуктов в холодильнике пользователя.

## API Endpoints

### Chef Mentor (Public)
```http
POST /api/ai/chef-mentor
Content-Type: application/json

{
  "message": "Хочу приготовить пасту карбонара",
  "language": "ru",
  "history": [
    {"role": "user", "content": "Привет"},
    {"role": "assistant", "content": "Здравствуйте! Чем могу помочь?"}
  ],
  "currentRecipe": {
    "title": "Паста Карбонара",
    "ingredients": [],
    "steps": []
  }
}
```

**Response:**
```json
{
  "message": "Отлично! Для пасты карбонара нам понадобятся спагетти, бекон, яйца, пармезан и черный перец. Начнем с ингредиентов?",
  "recipe": {
    "title": "Паста Карбонара",
    "ingredients": [],
    "steps": [],
    "isComplete": false
  },
  "nextQuestion": "Какие ингредиенты у вас есть?",
  "isComplete": false,
  "suggestedActions": ["Добавить ингредиенты", "Указать порции"]
}
```

### Generate Meal Plan (Protected)
```http
POST /api/ai/meal-plan
Authorization: Bearer {token}
Content-Type: application/json

{
  "language": "ru",
  "targetCalories": 2000,
  "days": 7,
  "useFridge": true
}
```

**Response:**
```json
{
  "plan": [
    {
      "day": "Day 1",
      "breakfast": "Овсянка с фруктами",
      "lunch": "Куриная грудка с овощами",
      "dinner": "Рыба на пару с рисом",
      "snack": "Йогурт с орехами",
      "totalCalories": 2000
    }
  ],
  "totalCalories": 14000,
  "avgPerDay": 2000,
  "success": true
}
```

### Generate Recipe (Public)
```http
POST /api/ai/recipe-generator
Content-Type: application/json

{
  "title": "Тирамису",
  "language": "ru"
}
```

**Response:**
```json
{
  "title": "Тирамису",
  "description": "Классический итальянский десерт",
  "category": "dessert",
  "difficulty": "intermediate",
  "time": 45,
  "portions": 6,
  "ingredients": [
    {
      "name": "Маскарпоне",
      "amount": 500,
      "unit": "г"
    },
    {
      "name": "Печенье Савоярди",
      "amount": 300,
      "unit": "г"
    }
  ],
  "steps": [
    "Взбить яичные желтки с сахаром",
    "Добавить маскарпоне",
    "Смочить печенье в кофе",
    "Выложить слоями"
  ],
  "calories": 450,
  "protein": 12.5,
  "fats": 28.0,
  "carbs": 38.0,
  "cost": 25.50,
  "tokensReward": 150
}
```

### Fridge Recommendations (Protected)
```http
POST /api/ai/fridge-recommendations
Authorization: Bearer {token}
Content-Type: application/json

{
  "dietaryPreferences": ["vegetarian"],
  "cuisine": "italian",
  "maxTime": 30
}
```

**Response:**
```json
{
  "success": true,
  "recommendations": [
    {
      "recipeName": "Капрезе с базиликом",
      "description": "Свежий салат с моцареллой и помидорами",
      "matchPercentage": 95,
      "missingItems": [],
      "prepTime": 10,
      "difficulty": "easy"
    },
    {
      "recipeName": "Паста Примавера",
      "description": "Паста с весенними овощами",
      "matchPercentage": 80,
      "missingItems": ["спаржа"],
      "prepTime": 25,
      "difficulty": "easy"
    }
  ],
  "count": 2
}
```

## Business Logic

### Chef Mentor Flow
1. User sends message with optional conversation history
2. System maintains recipe draft state
3. AI guides user through recipe creation steps
4. Tracks completion status (title, ingredients, steps)
5. Suggests next questions based on current state

### Meal Plan Generation
1. Validates input (1-14 days, positive calories)
2. Optionally fetches user's fridge items
3. Builds AI prompt with constraints
4. Generates balanced meal plan
5. Calculates total and average calories

### Recipe Generation
1. Takes dish title and language
2. Builds structured prompt for AI
3. Requests JSON-formatted recipe
4. Parses and validates response
5. Includes nutrition and cost estimates

### Fridge Recommendations
1. Fetches available fridge items
2. Considers dietary preferences and cuisine
3. Applies time constraints
4. Calculates match percentage
5. Lists missing ingredients

## AI Integration

### Groq Client
- Model: `llama-3.1-70b-versatile` (default)
- Temperature: 0.7-0.8 (creative responses)
- Max Tokens: 1000-2000 depending on task
- Streaming support available

### Prompt Engineering
- System prompts tailored by functionality
- Context injection for recipe state
- Multi-language support (ua, en, ru, pl)
- JSON output formatting for structured data

## Error Handling

### Custom Errors
- `ErrEmptyMessage` - пустое сообщение в chef mentor
- `ErrEmptyTitle` - пустое название рецепта
- `ErrInvalidDays` - некорректное количество дней (1-14)
- `ErrInvalidCalories` - некорректные калории

### HTTP Status Codes
- **200 OK** - успешный запрос
- **400 Bad Request** - неверные данные
- **401 Unauthorized** - не авторизован (для protected endpoints)
- **500 Internal Server Error** - ошибка AI или сервера

## Dependencies

### External
- `github.com/dmitrijfomin/menu-fodifood/backend/internal/ai` - Groq AI client
- `go.uber.org/zap` - логирование
- `gorm.io/gorm` - database для fridge items

### Internal
- `internal/models` - модели данных
- `internal/middleware` - JWT middleware
- `internal/platform/httpx` - HTTP helpers
- `internal/platform/logger` - логгер

## Usage Example

```go
import (
    "gorm.io/gorm"
    "backend/internal/modules/ai"
)

// Initialize module
aiModule := ai.NewModule(db)

// Register routes
aiModule.RegisterRoutes(router, jwtMiddleware)
```

## Testing

```bash
# Chef Mentor
curl -X POST http://localhost:8080/api/ai/chef-mentor \
  -H "Content-Type: application/json" \
  -d '{"message":"Хочу приготовить борщ","language":"ru"}'

# Generate Recipe
curl -X POST http://localhost:8080/api/ai/recipe-generator \
  -H "Content-Type: application/json" \
  -d '{"title":"Пицца Маргарита","language":"ru"}'

# Meal Plan (requires auth)
curl -X POST http://localhost:8080/api/ai/meal-plan \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"language":"ru","targetCalories":2000,"days":7,"useFridge":true}'

# Fridge Recommendations (requires auth)
curl -X POST http://localhost:8080/api/ai/fridge-recommendations \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"cuisine":"italian","maxTime":30}'
```

## Security

- ✅ Chef Mentor и Recipe Generator - публичные (демо)
- ✅ Meal Plan и Fridge Recommendations требуют JWT
- ✅ Fridge data доступны только владельцу
- ✅ Валидация всех входных данных
- ✅ Rate limiting через Groq API
- ✅ Логирование всех AI запросов

## Performance

### Response Times
- Chef Mentor: ~2-3 seconds
- Recipe Generator: ~3-5 seconds
- Meal Plan: ~5-8 seconds (больше контента)
- Fridge Recommendations: ~2-3 seconds

### Optimization
- Caching для популярных рецептов (TODO)
- Streaming responses для длинных генераций
- Batch processing для meal plans
- Parallel AI requests где возможно

## Future Improvements

- [ ] WebSocket streaming для real-time responses
- [ ] Recipe image generation (DALL-E integration)
- [ ] Voice input support
- [ ] Multi-turn conversation persistence
- [ ] User feedback loop для улучшения промптов
- [ ] A/B testing разных промптов
- [ ] Локальное кэширование популярных запросов
- [ ] Nutrition API integration для точных расчетов
- [ ] Cost estimation API для реальных цен
