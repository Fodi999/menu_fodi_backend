# 🚀 AI Chef Mentor - Quick Start for Frontend

## Endpoint
```
POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/ai/chef-mentor
```

## Simple Request
```javascript
const response = await fetch('/api/ai/chef-mentor', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    message: "Хочу зробити роли Дракон з вугром 🐉",
    language: "ua"
  })
});

const { data } = await response.json();
console.log(data.message); // AI response
console.log(data.recipe);  // Current recipe draft
console.log(data.nextQuestion); // What to ask next
```

## Full Conversation
```javascript
const [messages, setMessages] = useState([]);
const [recipe, setRecipe] = useState(null);

const chat = async (userMessage) => {
  const res = await fetch('/api/ai/chef-mentor', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      message: userMessage,
      language: 'ua',
      history: messages,
      currentRecipe: recipe
    })
  });
  
  const { data } = await res.json();
  
  setMessages([
    ...messages,
    { role: 'user', content: userMessage },
    { role: 'assistant', content: data.message }
  ]);
  
  setRecipe(data.recipe);
  
  return data;
};
```

## UI Components

### Chat Interface
```jsx
<div className="chat">
  {messages.map((msg, i) => (
    <div key={i} className={`message ${msg.role}`}>
      {msg.content}
    </div>
  ))}
</div>

<input 
  onSubmit={(e) => chat(e.target.value)}
  placeholder="Опишіть вашу страву..."
/>
```

### Recipe Preview
```jsx
{recipe && (
  <div className="recipe-preview">
    <h3>{recipe.title}</h3>
    <p>{recipe.category} • {recipe.difficulty} • {recipe.time}хв</p>
    
    {recipe.ingredients?.length > 0 && (
      <div>
        <h4>Інгредієнти ({recipe.ingredients.length})</h4>
        {recipe.ingredients.map(ing => (
          <div>{ing.name}: {ing.amount}{ing.unit}</div>
        ))}
      </div>
    )}
    
    {recipe.isComplete && (
      <button onClick={saveRecipe}>💾 Зберегти рецепт</button>
    )}
  </div>
)}
```

### Suggested Actions
```jsx
{data.suggestedActions?.map(action => (
  <button onClick={() => chat(action)}>
    {action}
  </button>
))}
```

## Example Flow

**1. Start:**
```
USER: "Привіт!"
AI: "Привіт! Я AI Chef Mentor. Яку страву ви хочете приготувати?"
```

**2. Describe Dish:**
```
USER: "Хочу зробити роли Дракон з вугром 🐉"
AI: "Чудово! Роли Дракон звучать смачно! Скільки порцій?"
RECIPE: { title: "Роли Дракон з вугром", category: "sushi" }
```

**3. Add Details:**
```
USER: "4 порції, 45 хвилин, середня складність"
AI: "Відмінно! Тепер додамо інгредієнти..."
RECIPE: { ..., portions: 4, time: 45, difficulty: "intermediate" }
```

**4. Complete:**
```
USER: "Готово!"
AI: "Рецепт завершено! Калорії: 1096, Вартість: 62 PLN, Нагорода: 25 CT"
RECIPE: { ..., isComplete: true }
```

## Response Structure
```typescript
interface MentorResponse {
  message: string;           // AI's natural response
  recipe: RecipeDraft;       // Current state
  nextQuestion: string;      // Guidance
  isComplete: boolean;       // Ready to save?
  suggestedActions: string[]; // Quick buttons
}

interface RecipeDraft {
  title?: string;
  description?: string;
  category?: string;
  difficulty?: string;
  time?: number;
  portions?: number;
  ingredients?: Ingredient[];
  steps?: string[];
  calories?: number;
  protein?: number;
  fats?: number;
  carbs?: number;
  cost?: number;
  tokensReward?: number;
  isComplete?: boolean;
}
```

## Save Recipe
```javascript
// When recipe is complete
const saveRecipe = async () => {
  const response = await fetch('/api/recipes', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({
      title: recipe.title,
      description: recipe.description,
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
  
  const { data } = await response.json();
  console.log('Recipe saved:', data.id);
};
```

## Testing
```bash
# Test with curl
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/ai/chef-mentor \
  -H "Content-Type: application/json" \
  -d '{"message":"Хочу зробити філадельфія рол","language":"ua"}'
```

## Ready to Use! 🎉
- ✅ Endpoint live on production
- ✅ Multi-language support
- ✅ Natural conversation
- ✅ Real-time recipe building
- ✅ Automatic metrics calculation
