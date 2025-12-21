# Recipe Match API Contract

## 🎯 Цель

**Официальный контракт** для endpoint `/api/recipes/match` - поиск рецептов по содержимому холодильника.

Этот контракт **НЕ МОЖЕТ меняться без версионирования** - используется фронтендом и AI.

---

## 📋 Endpoint: Match Recipes

### Request

```http
GET /api/recipes/match?country=Poland&maxTime=60&minScore=50
Authorization: Bearer {JWT_TOKEN}
```

### Query Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `country` | string | No | - | Фильтр по стране ("Poland", "Italy", "France") |
| `category` | string | No | - | Категория ("main", "dessert", "soup", "salad") |
| `difficulty` | string | No | - | Сложность ("easy", "medium", "hard") |
| `maxTime` | int | No | - | Максимальное время приготовления (минуты) |
| `excludeAllergens` | string | No | - | Исключить аллергены (через запятую: "gluten,lactose") |
| `includeDietTags` | string | No | - | Только с диетами (через запятую: "vegetarian,keto") |
| `minScore` | float | No | 50.0 | Минимальный score (0-100) |
| `limit` | int | No | 10 | Максимум результатов |

### Response

```json
{
  "success": true,
  "data": {
    "recipes": [
      {
        "recipeId": "550e8400-e29b-41d4-a716-446655440000",
        "canonicalName": "Spaghetti Carbonara",
        "localName": "Spaghetti alla Carbonara",
        "country": "Italy",
        "category": "main",
        "difficulty": "easy",
        "timeMinutes": 25,
        "servings": 4,
        
        "score": 82.5,
        "coverage": 0.75,
        
        "usedIngredients": [
          {
            "ingredientId": "uuid",
            "name": "Jajko",
            "quantity": 4,
            "unit": "pcs",
            "available": 6,
            "isExpiringSoon": false
          },
          {
            "ingredientId": "uuid",
            "name": "Makaron spaghetti",
            "quantity": 400,
            "unit": "g",
            "available": 500,
            "isExpiringSoon": false
          },
          {
            "ingredientId": "uuid",
            "name": "Boczek",
            "quantity": 150,
            "unit": "g",
            "available": 200,
            "isExpiringSoon": true
          }
        ],
        
        "missingIngredients": [
          {
            "ingredientId": "uuid",
            "name": "Parmezan",
            "quantity": 100,
            "unit": "g",
            "optional": false,
            "estimatedCost": 8.50
          }
        ],
        
        "canCookNow": false,
        "costToComplete": 8.50,
        "hasExpiringItems": true,
        "expiringItemsCount": 1,
        
        "allergens": ["gluten", "eggs", "lactose"],
        "dietTags": ["high-protein"]
      }
    ],
    "count": 1
  }
}
```

### Error Response

```json
{
  "success": false,
  "error": "Failed to find recipes",
  "message": "Database connection error"
}
```

---

## 🔑 Ключевые поля

### `score` (float, 0-100)
**Что это**: Общий рейтинг совпадения рецепта с холодильником.

**Формула**:
```
base_score = (matched_ingredients / required_ingredients) * 100
bonus_optional = (matched_optional / total_optional) * 5
bonus_expiring = expiring_items_count * 2
final_score = min(base_score + bonus_optional + bonus_expiring, 100)
```

**Примеры**:
- `100` - все ингредиенты есть + есть expiring items
- `75` - 75% ингредиентов есть
- `50` - половина ингредиентов есть (минимальный порог)
- `0` - ничего нет

**Использование на фронте**:
```typescript
// Сортировка по score
recipes.sort((a, b) => b.score - a.score);

// Визуализация
<ProgressBar value={recipe.score} max={100} />
<Badge color={recipe.score > 80 ? 'green' : 'yellow'}>
  {recipe.score}%
</Badge>
```

---

### `coverage` (float, 0-1)
**Что это**: Процент покрытия REQUIRED ингредиентов (без учета optional).

**Формула**:
```
coverage = matched_ingredients / required_ingredients
```

**Примеры**:
- `1.0` - все required ингредиенты есть
- `0.75` - 75% required ингредиентов есть
- `0.0` - ни одного required ингредиента

**Использование на фронте**:
```typescript
// Визуализация покрытия
<CircularProgress value={recipe.coverage * 100} />
<Text>{Math.round(recipe.coverage * 100)}% готов</Text>
```

---

### `canCookNow` (boolean)
**Что это**: Можно ли готовить ПРЯМО СЕЙЧАС (все required ингредиенты есть).

**Логика**:
```
canCookNow = (missing_required_ingredients.length === 0)
```

**Использование на фронте**:
```typescript
{recipe.canCookNow ? (
  <Button color="green">🍳 Gotuj teraz</Button>
) : (
  <Button color="yellow">
    ➕ Dodaj {recipe.missingIngredients.length} składników
  </Button>
)}
```

---

### `usedIngredients` vs `missingIngredients`

**`usedIngredients`** - что используется из холодильника:
- `available` - сколько есть в холодильнике
- `quantity` - сколько нужно для рецепта
- `isExpiringSoon` - истекает в течение 3 дней

**`missingIngredients`** - что нужно докупить:
- `quantity` - сколько нужно
- `estimatedCost` - примерная стоимость (PLN)
- `optional` - можно ли обойтись без этого

**Использование на фронте**:
```typescript
// Кнопка "Добавить все в холодильник"
const addMissingToFridge = () => {
  recipe.missingIngredients.forEach(ing => {
    fridgeApi.addIngredient({
      ingredientId: ing.ingredientId,
      name: ing.name,
      quantity: ing.quantity,
      unit: ing.unit
    });
  });
};

// Показать expiring items с приоритетом
const expiringIngredients = recipe.usedIngredients.filter(
  ing => ing.isExpiringSoon
);
```

---

### `costToComplete` (float, PLN)
**Что это**: Сколько стоит докупить недостающие ингредиенты.

**Формула**:
```
costToComplete = Σ(missing_ingredient.estimatedCost)
```

**Примеры**:
- `0.0` - все есть, ничего докупать не нужно
- `8.50` - нужно докупить на 8.50 PLN
- `45.0` - нужно докупить на 45 PLN (дорогой рецепт)

**Использование на фронте**:
```typescript
// Сортировка по цене
recipes.sort((a, b) => a.costToComplete - b.costToComplete);

// Визуализация
<Badge color={recipe.costToComplete < 10 ? 'green' : 'red'}>
  {recipe.costToComplete.toFixed(2)} PLN
</Badge>
```

---

### `hasExpiringItems` + `expiringItemsCount`
**Что это**: Есть ли продукты истекающие в течение 3 дней.

**Использование**:
- **Приоритизация** - рецепты с expiring items выше в списке
- **Waste reduction** - помогает избежать выброса продуктов
- **UX badge** - показать "🔥 Используй скоро портящиеся продукты"

**Использование на фронте**:
```typescript
{recipe.hasExpiringItems && (
  <Badge color="orange">
    🔥 {recipe.expiringItemsCount} składnik(ów) expire soon
  </Badge>
)}

// Сортировка с приоритетом expiring
recipes.sort((a, b) => {
  if (a.hasExpiringItems && !b.hasExpiringItems) return -1;
  if (!a.hasExpiringItems && b.hasExpiringItems) return 1;
  return b.score - a.score;
});
```

---

## 🎨 Frontend Examples

### 1. Recipe Card Component

```typescript
interface RecipeCardProps {
  recipe: RecipeMatchItem;
}

const RecipeCard: React.FC<RecipeCardProps> = ({ recipe }) => {
  return (
    <Card>
      <CardHeader>
        <Title>{recipe.localName}</Title>
        <Badge>{recipe.country}</Badge>
      </CardHeader>
      
      <CardBody>
        {/* Score bar */}
        <ProgressBar value={recipe.score} max={100} />
        <Text>Match: {recipe.score.toFixed(0)}%</Text>
        
        {/* Quick info */}
        <Flex gap={2}>
          <Badge>⏱️ {recipe.timeMinutes} min</Badge>
          <Badge>{recipe.difficulty}</Badge>
          <Badge>👥 {recipe.servings} osób</Badge>
        </Flex>
        
        {/* Expiring warning */}
        {recipe.hasExpiringItems && (
          <Alert color="orange">
            🔥 Użyj {recipe.expiringItemsCount} składników zanim się zepsują
          </Alert>
        )}
        
        {/* Missing ingredients */}
        {recipe.missingIngredients.length > 0 && (
          <Box>
            <Text>Brakuje:</Text>
            <List>
              {recipe.missingIngredients.map(ing => (
                <ListItem key={ing.ingredientId}>
                  {ing.name} - {ing.quantity} {ing.unit}
                  <Badge>{ing.estimatedCost.toFixed(2)} PLN</Badge>
                </ListItem>
              ))}
            </List>
            <Text>Koszt: {recipe.costToComplete.toFixed(2)} PLN</Text>
          </Box>
        )}
      </CardBody>
      
      <CardFooter>
        {recipe.canCookNow ? (
          <Button color="green" onClick={() => startCooking(recipe)}>
            🍳 Gotuj teraz
          </Button>
        ) : (
          <Button color="yellow" onClick={() => addMissing(recipe)}>
            ➕ Dodaj {recipe.missingIngredients.length} składników
          </Button>
        )}
      </CardFooter>
    </Card>
  );
};
```

### 2. Filtering Logic

```typescript
const fetchRecipes = async (filters: {
  country?: string;
  maxTime?: number;
  minScore?: number;
}) => {
  const params = new URLSearchParams();
  
  if (filters.country) params.append('country', filters.country);
  if (filters.maxTime) params.append('maxTime', filters.maxTime.toString());
  if (filters.minScore) params.append('minScore', filters.minScore.toString());
  
  const response = await fetch(`/api/recipes/match?${params}`, {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  
  const data = await response.json();
  
  if (data.success) {
    return data.data.recipes;
  } else {
    throw new Error(data.error);
  }
};
```

### 3. Add Missing to Fridge

```typescript
const addMissingToFridge = async (recipe: RecipeMatchItem) => {
  const batch = recipe.missingIngredients.map(ing => ({
    ingredientId: ing.ingredientId,
    name: ing.name,
    quantity: ing.quantity,
    unit: ing.unit,
  }));
  
  await fridgeApi.addBatch(batch);
  
  toast.success(`Dodano ${batch.length} składników do lodówki`);
  
  // Refresh matches
  const updatedMatches = await fetchRecipes();
};
```

---

## 🔄 Версионирование

**Current version**: `v1`

**Breaking changes policy**:
- Нельзя удалять поля
- Нельзя менять типы полей
- Можно добавлять новые опциональные поля
- При breaking change → создать `/api/v2/recipes/match`

**Changelog**:
- `2025-12-21` - v1 initial release

---

## ✅ Checklist для frontend integration

- [ ] Создать TypeScript types из этого контракта
- [ ] Реализовать `RecipeCard` компонент
- [ ] Реализовать фильтрацию (country, maxTime, minScore)
- [ ] Реализовать сортировку (score, costToComplete, hasExpiringItems)
- [ ] Реализовать "Add missing to fridge" кнопку
- [ ] Реализовать "Start cooking" flow
- [ ] Добавить error handling
- [ ] Добавить loading states
- [ ] Тесты для API integration
